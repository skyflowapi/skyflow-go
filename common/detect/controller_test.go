package detect

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/skyflowapi/skyflow-go/common/auth"
	"github.com/skyflowapi/skyflow-go/common/config"
	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
)

const testApiKey = "sky-abcde-1234567890abcdef1234567890abcdef"

// fakeDetect records each request and answers with the response registered for its path.
type fakeDetect struct {
	mu        sync.Mutex
	server    *httptest.Server
	requests  []recorded
	responses map[string][]string // path -> bodies, served in order; the last repeats
	status    int
}

type recorded struct {
	method, path, query string
	header              http.Header
	body                map[string]interface{}
}

func newFakeDetect(t *testing.T) *fakeDetect {
	f := &fakeDetect{responses: map[string][]string{}, status: http.StatusOK}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		rec := recorded{method: r.Method, path: r.URL.Path, query: r.URL.RawQuery, header: r.Header.Clone()}
		_ = json.NewDecoder(r.Body).Decode(&rec.body)
		f.requests = append(f.requests, rec)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req-123")
		if f.status != http.StatusOK {
			w.WriteHeader(f.status)
			_, _ = w.Write([]byte(`{"error":{"http_code":404,"message":"Run not found.","http_status":"Not Found"}}`))
			return
		}
		bodies := f.responses[r.URL.Path]
		if len(bodies) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(bodies[0]))
		if len(bodies) > 1 {
			f.responses[r.URL.Path] = bodies[1:]
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeDetect) last() recorded {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests[len(f.requests)-1]
}

func newTestController(f *fakeDetect) (*Controller, *[]time.Duration) {
	profile := vaultclient.Profile{
		SDKPrefix:   "Test SDK",
		LogPrefix:   "[ Test SDK ] ",
		Auth:        auth.New(auth.Config{}),
		ClusterURL:  func(config.Env, string) string { return "" },
		SDKMetadata: func() string { return "meta" },
	}
	base := vaultclient.New(profile, config.VaultConfig{VaultId: "vault1", BaseVaultUrl: f.server.URL},
		vaultclient.ClientSettings{Credentials: &config.Credentials{ApiKey: testApiKey}})
	c := NewController(base)
	var pauses []time.Duration
	c.sleep = func(_ context.Context, d time.Duration) bool {
		pauses = append(pauses, d)
		return true
	}
	return c, &pauses
}

func TestDeidentifyStringSendsInlineConfigurationWithVaultId(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyString] = []string{`{
		"processedText": "My name is [NAME_1].",
		"entities": [{"token":"NAME_1","value":"","location":{"startIndex":11,"endIndex":19,"startIndexProcessed":11,"endIndexProcessed":19},"entityType":"NAME","entityScores":{"NAME":0.91}}],
		"metrics": {"size":0.05,"wordCount":4,"characterCount":19}
	}`}
	c, _ := newTestController(f)

	resp, err := c.DeidentifyString(context.Background(), DeidentifyStringRequest{
		Text: "My name is John Doe.",
		Configuration: &DetectConfiguration{Detect: &Detect{Entities: []Entity{{
			EntityType: EntityTypeName, DeidentificationType: DeidentificationTypeVaultToken, Destination: "persons",
		}}}},
	}, DeidentifyStringOptions{Interceptor: func(rc *vaultclient.RequestContext) {
		rc.AddHeader(config.RequestIdHeader, "caller-req")
	}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ProcessedText != "My name is [NAME_1]." || len(resp.Entities) != 1 || resp.Entities[0].EntityScores["NAME"] != 0.91 {
		t.Fatalf("unexpected response %+v", resp)
	}
	if resp.Metrics.WordCount != 4 || resp.Metrics.Pages != nil {
		t.Fatalf("metrics = %+v", resp.Metrics)
	}

	req := f.last()
	cfg := req.body["configuration"].(map[string]interface{})
	if cfg["vaultId"] != "vault1" || req.body["vaultId"] != nil {
		t.Fatalf("vault ID must go inside configuration only: %v", req.body)
	}
	entity := cfg["detect"].(map[string]interface{})["entities"].([]interface{})[0].(map[string]interface{})
	if entity["entityType"] != "NAME" || entity["deidentificationType"] != "VAULT_TOKEN" {
		t.Fatalf("entity = %v", entity)
	}
	if req.header.Get("Authorization") != "Bearer "+testApiKey || req.header.Get("X-Request-Id") != "caller-req" || req.header.Get("sky-metadata") != "meta" {
		t.Fatalf("headers = %v", req.header)
	}
}

func TestDeidentifyStringValidation(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	cases := map[string]DeidentifyStringRequest{
		"Text must not be empty": {Text: " ", ConfigurationId: "cfg"},
	}
	for want, request := range cases {
		_, err := c.DeidentifyString(context.Background(), request, DeidentifyStringOptions{})
		if err == nil || !strings.Contains(err.GetMessage(), want) || !strings.Contains(err.GetMessage(), "Test SDK") {
			t.Errorf("%q: err = %v", want, err)
		}
	}
}

// Configuration sources match the Java SDK: both are optional and sent as given.
func TestConfigurationSourcesAreSentAsGiven(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyString] = []string{`{"processedText":"x"}`, `{"processedText":"x"}`, `{"processedText":"x"}`}
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	c, _ := newTestController(f)
	ctx := context.Background()

	cases := []struct {
		name          string
		request       DeidentifyStringRequest
		wantId        interface{}
		wantVaultOnly bool
	}{
		{"neither: vault defaults", DeidentifyStringRequest{Text: "x"}, nil, true},
		{"configuration ID only", DeidentifyStringRequest{Text: "x", ConfigurationId: "cfg"}, nil, false},
		{"both: passed through for the API to reject", DeidentifyStringRequest{Text: "x", ConfigurationId: "cfg", Configuration: &DetectConfiguration{}}, "cfg", false},
	}
	for _, tc := range cases {
		if _, err := c.DeidentifyString(ctx, tc.request, DeidentifyStringOptions{}); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		body := f.last().body
		cfg, hasCfg := body["configuration"].(map[string]interface{})
		switch {
		case tc.wantVaultOnly:
			if !hasCfg || len(cfg) != 1 || cfg["vaultId"] != "vault1" || body["configurationId"] != nil {
				t.Errorf("%s: body = %v", tc.name, body)
			}
		case tc.request.Configuration == nil:
			if hasCfg || body["configurationId"] != "cfg" {
				t.Errorf("%s: body = %v", tc.name, body)
			}
		default:
			if !hasCfg || cfg["vaultId"] != "vault1" || body["configurationId"] != tc.wantId {
				t.Errorf("%s: body = %v", tc.name, body)
			}
		}
	}

	if _, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64}, DeidentifyFileOptions{}); err != nil {
		t.Fatal(err)
	}
	if cfg, ok := f.last().body["configuration"].(map[string]interface{}); !ok || cfg["vaultId"] != "vault1" {
		t.Fatalf("deidentify file without configuration: body = %v", f.last().body)
	}
}

func TestDeidentifyFileSubmitOnly(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	c, _ := newTestController(f)

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{
		DataSource: DeidentifySourceTypeSkyflowID, Value: "sid", DataFormat: FileDataFormatPDF, ConfigurationId: "cfg",
	}, DeidentifyFileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RunId != "run-1" || resp.Status != "" || len(f.requests) != 1 {
		t.Fatalf("resp = %+v, requests = %d", resp, len(f.requests))
	}
	body := f.last().body
	if body["dataSource"] != "SKYFLOW_ID" || body["value"] != "sid" || body["dataFormat"] != "pdf" || body["configurationId"] != "cfg" {
		t.Fatalf("body = %v", body)
	}
}

func TestDeidentifyFilePollsAndWritesOutputs(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-9"}`}
	redacted := base64.StdEncoding.EncodeToString([]byte("redacted pdf"))
	entities := base64.StdEncoding.EncodeToString([]byte(`[{"token":"NAME_1"}]`))
	f.responses[pathRuns+"run-9"] = []string{
		`{"status":"QUEUED"}`,
		`{"status":"IN_PROGRESS"}`,
		`{"status":"SUCCESS","outputType":"BASE64","output":[
			{"processedFile":"` + redacted + `","processedFileType":"REDACTED_FILE","processedFileExtension":"pdf"},
			{"processedFile":"` + entities + `","processedFileType":"ENTITIES","processedFileExtension":"json"}
		],"metrics":{"pages":3}}`,
	}
	c, pauses := newTestController(f)

	dir := t.TempDir()
	input := filepath.Join(dir, "notes.pdf")
	if err := os.WriteFile(input, []byte("original pdf"), 0o600); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "out")

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{FilePath: input, ConfigurationId: "cfg"},
		DeidentifyFileOptions{Poll: &PollOptions{}, OutputDirectory: outDir})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RunId != "run-9" || resp.Status != DetectRunStatusSuccess || len(resp.Output) != 2 || *resp.Metrics.Pages != 3 {
		t.Fatalf("resp = %+v", resp)
	}
	if got := *pauses; len(got) != 2 || got[0] != time.Second || got[1] != 2*time.Second {
		t.Fatalf("pauses = %v", got)
	}
	submit := f.requests[0].body
	if submit["dataSource"] != "BASE64" || submit["dataFormat"] != "pdf" || submit["value"] != base64.StdEncoding.EncodeToString([]byte("original pdf")) {
		t.Fatalf("submit body = %v", submit)
	}
	if run := f.last(); run.method != http.MethodGet || run.query != "vaultId=vault1" {
		t.Fatalf("run lookup = %+v", run)
	}
	if got, _ := os.ReadFile(filepath.Join(outDir, "processed-notes.pdf")); string(got) != "redacted pdf" {
		t.Fatalf("processed file = %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(outDir, "processed-entities-notes.json")); !strings.Contains(string(got), "NAME_1") {
		t.Fatalf("entities file = %q", got)
	}
}

func TestPollStopsWhenBudgetRunsOut(t *testing.T) {
	f := newFakeDetect(t)
	c, pauses := newTestController(f)
	calls := 0
	fetch := func(context.Context) (*GetRunResponse, *skyflowError.SkyflowError) {
		calls++
		return &GetRunResponse{RunId: "r", Status: DetectRunStatusInProgress}, nil
	}

	run, err := c.poll(context.Background(), "r", PollOptions{WaitTime: 10 * time.Second}, fetch)
	if err != nil || run.Status != DetectRunStatusInProgress {
		t.Fatalf("run = %+v, err = %v", run, err)
	}
	// 1 + 2 + 4 + 3 (trimmed) = 10s, then the budget is spent after the 5th lookup.
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 3 * time.Second}
	if len(*pauses) != len(want) || calls != 5 {
		t.Fatalf("pauses = %v, calls = %d", *pauses, calls)
	}
	for i := range want {
		if (*pauses)[i] != want[i] {
			t.Fatalf("pauses = %v", *pauses)
		}
	}

	calls, *pauses = 0, nil
	if _, err := c.poll(context.Background(), "r", PollOptions{MaxAttempts: 2}, fetch); err != nil || calls != 2 {
		t.Fatalf("MaxAttempts not honoured: calls = %d", calls)
	}
}

func TestDeidentifyFileValidation(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	cases := []struct {
		want    string
		request DeidentifyFileRequest
		options DeidentifyFileOptions
	}{
		{"exactly one of Value, FilePath or FileObject", DeidentifyFileRequest{ConfigurationId: "c"}, DeidentifyFileOptions{}},
		{"DataSource is required", DeidentifyFileRequest{Value: "v", ConfigurationId: "c"}, DeidentifyFileOptions{}},
		{"DataSource must be empty or BASE64", DeidentifyFileRequest{FilePath: "a.pdf", DataSource: DeidentifySourceTypeSkyflowID, ConfigurationId: "c"}, DeidentifyFileOptions{}},
		{"OutputDirectory requires Poll", DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64, ConfigurationId: "c"}, DeidentifyFileOptions{OutputDirectory: "out"}},
		{"WaitTime must be between", DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64, ConfigurationId: "c"}, DeidentifyFileOptions{Poll: &PollOptions{WaitTime: 301 * time.Second}}},
		{"MaxAttempts must be between", DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64, ConfigurationId: "c"}, DeidentifyFileOptions{Poll: &PollOptions{MaxAttempts: 24}}},
		{"Could not read file", DeidentifyFileRequest{FilePath: "/does/not/exist.pdf", ConfigurationId: "c"}, DeidentifyFileOptions{}},
	}
	for _, tc := range cases {
		_, err := c.DeidentifyFile(context.Background(), tc.request, tc.options)
		if err == nil || !strings.Contains(err.GetMessage(), tc.want) {
			t.Errorf("%q: err = %v", tc.want, err)
		}
	}
}

func TestGetRunEchoesRunIdAndMapsApiErrors(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathRuns+"run-2"] = []string{`{"status":"FAILED","message":"bad file"}`}
	c, _ := newTestController(f)

	run, err := c.GetRun(context.Background(), GetRunRequest{RunId: "run-2"}, GetRunOptions{})
	if err != nil || run.RunId != "run-2" || run.Status != DetectRunStatusFailed || run.Message != "bad file" || run.Output == nil {
		t.Fatalf("run = %+v, err = %v", run, err)
	}

	f.status = http.StatusNotFound
	_, err = c.GetRun(context.Background(), GetRunRequest{RunId: "missing"}, GetRunOptions{})
	if err == nil || err.GetHttpCode() != "404" || err.GetRequestId() != "req-123" || !strings.Contains(err.GetMessage(), "Run not found") {
		t.Fatalf("err = %v", err)
	}

	if _, err := c.GetRun(context.Background(), GetRunRequest{}, GetRunOptions{}); err == nil || !strings.Contains(err.GetMessage(), "RunId must not be empty") {
		t.Fatalf("err = %v", err)
	}
}

func TestReidentifyStringSendsVaultIdAndRedactionLevels(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathReidentifyString] = []string{`{"processedText":"My name is J**n D*e"}`}
	c, _ := newTestController(f)

	resp, err := c.ReidentifyString(context.Background(), ReidentifyStringRequest{
		Text:           "My name is [NAME_1].",
		RedactionLevel: []RedactionLevel{{TokenGroupName: "persons", RedactionPattern: "mask"}},
	}, ReidentifyStringOptions{})
	if err != nil || resp.ProcessedText != "My name is J**n D*e" {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	body := f.last().body
	level := body["redactionLevel"].([]interface{})[0].(map[string]interface{})
	if body["vaultId"] != "vault1" || level["tokenGroupName"] != "persons" || level["redactionPattern"] != "mask" || level["entityName"] != nil {
		t.Fatalf("body = %v", body)
	}

	for _, levels := range [][]RedactionLevel{
		{{RedactionType: RedactionTypeMasked}},
		{{TokenGroupName: "g", EntityName: EntityTypeName, RedactionType: RedactionTypeMasked}},
		{{EntityName: EntityTypeName}},
		{{EntityName: EntityTypeName, RedactionType: RedactionTypeMasked, RedactionPattern: "p"}},
	} {
		if _, err := c.ReidentifyString(context.Background(), ReidentifyStringRequest{Text: "x", RedactionLevel: levels}, ReidentifyStringOptions{}); err == nil {
			t.Errorf("levels %+v should fail", levels)
		}
	}
}

func TestReidentifyFileWritesOutputsForLocalFile(t *testing.T) {
	f := newFakeDetect(t)
	plain := base64.StdEncoding.EncodeToString([]byte("plain text"))
	f.responses[pathReidentifyFile] = []string{`{"status":"SUCCESS","outputType":"BASE64","output":[{"processedFile":"` + plain + `","processedFileType":"REIDENTIFIED_FILE","processedFileExtension":"txt"}]}`}
	c, _ := newTestController(f)

	dir := t.TempDir()
	input, _ := os.Create(filepath.Join(dir, "redacted.txt"))
	_, _ = input.WriteString("[NAME_1]")
	_, _ = input.Seek(0, 0)
	defer func() { _ = input.Close() }()

	resp, err := c.ReidentifyFile(context.Background(), ReidentifyFileRequest{FileObject: input}, ReidentifyFileOptions{OutputDirectory: dir})
	if err != nil || resp.Status != DetectRunStatusSuccess {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if body := f.last().body; body["vaultId"] != "vault1" || body["dataFormat"] != "txt" || body["dataSource"] != "BASE64" {
		t.Fatalf("body = %v", body)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "processed-redacted.txt")); string(got) != "plain text" {
		t.Fatalf("output = %q", got)
	}
}

func TestCheckGuardrails(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathGuardrails] = []string{`{"text":"hi","validation":"FAILED","toxic":true,"deniedTopic":false}`}
	c, _ := newTestController(f)

	resp, err := c.CheckGuardrails(context.Background(), CheckGuardrailsRequest{Text: "hi", CheckToxicity: true, DenyTopics: []string{"politics"}}, CheckGuardrailsOptions{})
	if err != nil || resp.Validation != GuardrailsValidationFailed || !resp.Toxic {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if body := f.last().body; body["vaultId"] != "vault1" || body["checkToxicity"] != true {
		t.Fatalf("body = %v", body)
	}
	if _, err := c.CheckGuardrails(context.Background(), CheckGuardrailsRequest{Text: "hi", DenyTopics: []string{" "}}, CheckGuardrailsOptions{}); err == nil {
		t.Fatal("blank deny topic should fail")
	}
}

func TestOutputFileNames(t *testing.T) {
	used := map[string]bool{}
	names := []string{
		outputFileName("doc", FileOutput{ProcessedFileType: ProcessedFileTypeRedactedFile, ProcessedFileExtension: FileDataFormatPDF}, used),
		outputFileName("doc", FileOutput{ProcessedFileType: ProcessedFileTypeObjectEntities}, used),
		outputFileName("doc", FileOutput{ProcessedFileType: ProcessedFileTypeRedactedText, ProcessedFileExtension: FileDataFormatPDF}, used),
	}
	want := []string{"processed-doc.pdf", "processed-object-entities-doc.json", "processed-doc.redacted_text.pdf"}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
}
