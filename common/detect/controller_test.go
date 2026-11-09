package detect

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	c := NewController(base, SkySources)
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
	if m := resp.Metrics; m.WordCount == nil || *m.WordCount != 4 || m.Pages != nil {
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

// A retryable status is final once the HTTP client's own retries are done, so the call
// returns the API error at once instead of the generated client sleeping before giving up.
func TestRetryableStatusIsReturnedWithoutGeneratedRetryDelay(t *testing.T) {
	f := newFakeDetect(t)
	f.status = http.StatusServiceUnavailable
	c, _ := newTestController(f)

	start := time.Now()
	_, err := c.GetRun(context.Background(), GetRunRequest{RunId: "run-1"}, GetRunOptions{})
	if elapsed := time.Since(start); elapsed >= 400*time.Millisecond {
		t.Fatalf("call took %v; the generated retrier slept", elapsed)
	}
	if err == nil || err.GetRequestId() != "req-123" || !strings.Contains(err.GetMessage(), "Run not found") {
		t.Fatalf("err = %v", err)
	}
}

func TestRetryableResponseWithUnreadableBodyReportsTheReadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, buf, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Errorf("hijack: %v", err)
			return
		}
		defer conn.Close()
		_, _ = buf.WriteString("HTTP/1.1 503 Service Unavailable\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\nshort")
		_ = buf.Flush()
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := finalResponseClient{http.DefaultClient}.Do(req)
	if resp != nil || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("resp = %v, err = %v", resp, err)
	}
}

func TestDeidentifyStringValidation(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	cases := map[string]DeidentifyStringRequest{
		"Invalid text in deidentify string request": {Text: " ", ConfigurationId: "cfg"},
	}
	for want, request := range cases {
		_, err := c.DeidentifyString(context.Background(), request, DeidentifyStringOptions{})
		if err == nil || !strings.Contains(err.GetMessage(), want) || !strings.Contains(err.GetMessage(), "Test SDK") {
			t.Errorf("%q: err = %v", want, err)
		}
	}
}

// Configuration sources match the Java SDK: both are optional and sent as given.
func TestConfigurationSourceIsSent(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyString] = []string{`{"processedText":"x"}`, `{"processedText":"x"}`}
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	c, _ := newTestController(f)
	ctx := context.Background()

	// A stored configuration is sent by ID, without a configuration.
	if _, err := c.DeidentifyString(ctx, DeidentifyStringRequest{Text: "x", ConfigurationId: "cfg"}, DeidentifyStringOptions{}); err != nil {
		t.Fatal(err)
	}
	if body := f.last().body; body["configuration"] != nil || body["configurationId"] != "cfg" {
		t.Errorf("configuration ID: body = %v", body)
	}
	// An inline configuration carries the vault ID, without a configuration ID.
	if _, err := c.DeidentifyString(ctx, DeidentifyStringRequest{Text: "x", Configuration: &DetectConfiguration{}}, DeidentifyStringOptions{}); err != nil {
		t.Fatal(err)
	}
	if body := f.last().body; body["configurationId"] != nil {
		t.Errorf("inline configuration: body = %v", body)
	} else if cfg, ok := body["configuration"].(map[string]interface{}); !ok || cfg["vaultId"] != "vault1" {
		t.Errorf("inline configuration: body = %v", body)
	}
	if _, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64, Configuration: &DetectConfiguration{}}, DeidentifyFileOptions{}); err != nil {
		t.Fatal(err)
	}
	if cfg, ok := f.last().body["configuration"].(map[string]interface{}); !ok || cfg["vaultId"] != "vault1" {
		t.Fatalf("deidentify file with an inline configuration: body = %v", f.last().body)
	}
}

// As in the Java SDK, neither configuration source is required: with neither, a configuration
// holding only the vault ID is sent; with both, both are sent.
func TestConfigurationSourceIsOptional(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyString] = []string{`{"processedText":"x"}`}
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	c, _ := newTestController(f)
	ctx := context.Background()

	if _, err := c.DeidentifyString(ctx, DeidentifyStringRequest{Text: "x"}, DeidentifyStringOptions{}); err != nil {
		t.Fatal(err)
	}
	if cfg, ok := f.last().body["configuration"].(map[string]interface{}); !ok || len(cfg) != 1 || cfg["vaultId"] != "vault1" || f.last().body["configurationId"] != nil {
		t.Errorf("deidentify string, neither: body = %v", f.last().body)
	}
	if _, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64}, DeidentifyFileOptions{}); err != nil {
		t.Fatal(err)
	}
	if cfg, ok := f.last().body["configuration"].(map[string]interface{}); !ok || len(cfg) != 1 || cfg["vaultId"] != "vault1" {
		t.Errorf("deidentify file, neither: body = %v", f.last().body)
	}
	if _, err := c.DeidentifyString(ctx, DeidentifyStringRequest{Text: "x", ConfigurationId: "cfg", Configuration: &DetectConfiguration{}}, DeidentifyStringOptions{}); err != nil {
		t.Fatal(err)
	}
	if body := f.last().body; body["configurationId"] != "cfg" || body["configuration"] == nil {
		t.Errorf("both: body = %v", body)
	}
}

func TestDeidentifyFileSubmitOnly(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	c, _ := newTestController(f)

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{
		DataSource: SkyDeidentifySourceTypeSkyflowID, Value: "sid", DataFormat: FileDataFormatPDF, ConfigurationId: "cfg",
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
	if resp.RunId != "run-9" || resp.Status != DetectRunStatusSuccess || resp.OutputType != SkyDeidentifySourceTypeBase64 ||
		len(resp.Output) != 2 || *resp.Metrics.Pages != 3 {
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

	run, err := c.poll(context.Background(), "r", PollOptions{WaitTime: ptr(10 * time.Second)}, fetch)
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
	if _, err := c.poll(context.Background(), "r", PollOptions{MaxAttempts: ptr(2)}, fetch); err != nil || calls != 2 {
		t.Fatalf("MaxAttempts not honoured: calls = %d", calls)
	}
}

func TestDeidentifyFileValidation(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	base64Request := DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64, ConfigurationId: "c"}
	waitTime := func(d time.Duration) DeidentifyFileOptions {
		return DeidentifyFileOptions{Poll: &PollOptions{WaitTime: ptr(d)}}
	}
	maxAttempts := func(n int) DeidentifyFileOptions {
		return DeidentifyFileOptions{Poll: &PollOptions{MaxAttempts: ptr(n)}}
	}
	invalidWait := fmt.Sprintf(errInvalidWaitTime, minWaitTimeSeconds, maxWaitTimeSeconds)
	invalidAttempts := fmt.Sprintf(errInvalidMaxAttempts, minMaxAttempts, maxMaxAttempts)
	cases := []struct {
		want    string
		request DeidentifyFileRequest
		options DeidentifyFileOptions
	}{
		{errMissingFileSource, DeidentifyFileRequest{ConfigurationId: "c"}, DeidentifyFileOptions{}},
		{errInvalidDeidentifyDataSource, DeidentifyFileRequest{Value: "v", ConfigurationId: "c"}, DeidentifyFileOptions{}},
		{errOutputDirectoryNeedPoll, base64Request, DeidentifyFileOptions{OutputDirectory: "out"}},
		{invalidWait, base64Request, waitTime(301 * time.Second)},
		// A set zero is out of range, as in the Java SDK; only a nil field takes the default.
		{invalidWait, base64Request, waitTime(0)},
		{invalidWait, base64Request, waitTime(500 * time.Millisecond)},
		{invalidWait, base64Request, waitTime(-time.Second)},
		{invalidAttempts, base64Request, maxAttempts(24)},
		{invalidAttempts, base64Request, maxAttempts(0)},
		{invalidAttempts, base64Request, maxAttempts(-1)},
		{errFileNotFound, DeidentifyFileRequest{FilePath: "/does/not/exist.pdf", ConfigurationId: "c"}, DeidentifyFileOptions{}},
	}
	for _, tc := range cases {
		_, err := c.DeidentifyFile(context.Background(), tc.request, tc.options)
		if err == nil || !strings.Contains(err.GetMessage(), tc.want) {
			t.Errorf("%q: err = %v", tc.want, err)
		}
	}
	if !strings.Contains(invalidWait, "between 1 and 300 seconds") || !strings.Contains(invalidAttempts, "between 1 and 23.") {
		t.Errorf("messages = %q, %q", invalidWait, invalidAttempts)
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

	if _, err := c.GetRun(context.Background(), GetRunRequest{}, GetRunOptions{}); err == nil || !strings.Contains(err.GetMessage(), errEmptyRunId) {
		t.Fatalf("err = %v", err)
	}
}

func TestReidentifyStringSendsVaultIdAndRedactionLevels(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathReidentifyString] = []string{`{"processedText":"My name is J**n D*e"}`}
	c, _ := newTestController(f)

	resp, err := c.ReidentifyString(context.Background(), ReidentifyStringRequest{
		Text:           "My name is [NAME_1].",
		RedactionLevel: []RedactionRule{FlowRedactionLevel{TokenGroupName: "persons", RedactionPattern: "mask"}},
	}, ReidentifyStringOptions{})
	if err != nil || resp.ProcessedText != "My name is J**n D*e" {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	body := f.last().body
	level := body["redactionLevel"].([]interface{})[0].(map[string]interface{})
	if body["vaultId"] != "vault1" || level["tokenGroupName"] != "persons" || level["redactionPattern"] != "mask" || level["entityName"] != nil {
		t.Fatalf("body = %v", body)
	}

	// Rules are sent as given, as in the other SDKs: partial or empty rules are not rejected.
	for _, levels := range [][]RedactionRule{
		{FlowRedactionLevel{RedactionPattern: "p"}},
		{FlowRedactionLevel{}},
		{SkyRedactionLevel{RedactionType: RedactionTypeMasked}},
	} {
		f.responses[pathReidentifyString] = []string{`{"processedText":"x"}`}
		if _, err := c.ReidentifyString(context.Background(), ReidentifyStringRequest{Text: "x", RedactionLevel: levels}, ReidentifyStringOptions{}); err != nil {
			t.Errorf("levels %+v: err = %v", levels, err)
		}
	}

	// Nil rules, including nil pointers to the rule types, are skipped as in the other SDKs.
	var nilFlow *FlowRedactionLevel
	var nilSky *SkyRedactionLevel
	f.responses[pathReidentifyString] = []string{`{"processedText":"x"}`}
	f.responses[pathReidentifyFile] = []string{`{"status":"SUCCESS"}`}
	rules := []RedactionRule{nil, nilFlow, nilSky, &FlowRedactionLevel{TokenGroupName: "g"}}
	if _, err := c.ReidentifyString(context.Background(), ReidentifyStringRequest{Text: "x", RedactionLevel: rules}, ReidentifyStringOptions{}); err != nil {
		t.Fatalf("nil rules: err = %v", err)
	}
	sent := f.last().body["redactionLevel"].([]interface{})
	if len(sent) != 1 || sent[0].(map[string]interface{})["tokenGroupName"] != "g" {
		t.Fatalf("redactionLevel = %v", sent)
	}
	if _, err := c.ReidentifyFile(context.Background(), ReidentifyFileRequest{DataSource: SkyReidentifySourceTypeBase64, Value: "aGk=", RedactionLevel: []RedactionRule{nil}}, ReidentifyFileOptions{}); err != nil {
		t.Fatalf("nil rule on file: err = %v", err)
	}
	if _, sent := f.last().body["redactionLevel"]; sent {
		t.Fatal("only nil rules: redactionLevel should be omitted")
	}
}

func TestReidentifySendsSkyVaultRedactionLevels(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathReidentifyString] = []string{`{"processedText":"x"}`}
	c, _ := newTestController(f)

	if _, err := c.ReidentifyString(context.Background(), ReidentifyStringRequest{
		Text:           "[NAME_1]",
		RedactionLevel: []RedactionRule{SkyRedactionLevel{EntityName: EntityTypeName, RedactionType: RedactionTypePlainText}},
	}, ReidentifyStringOptions{}); err != nil {
		t.Fatal(err)
	}
	level := f.last().body["redactionLevel"].([]interface{})[0].(map[string]interface{})
	if level["entityName"] != "NAME" || level["redactionType"] != "PLAIN_TEXT" || level["tokenGroupName"] != nil || level["redactionPattern"] != nil {
		t.Fatalf("level = %v", level)
	}
}

func TestReidentifyFileReturnsTheOutputInTheResponse(t *testing.T) {
	f := newFakeDetect(t)
	plain := base64.StdEncoding.EncodeToString([]byte("plain text"))
	f.responses[pathReidentifyFile] = []string{`{"status":"SUCCESS","outputType":"BASE64","output":[{"processedFile":"` + plain + `","processedFileType":"REIDENTIFIED_FILE","processedFileExtension":"txt"}]}`}
	c, _ := newTestController(f)

	redacted := base64.StdEncoding.EncodeToString([]byte("[NAME_1]"))
	resp, err := c.ReidentifyFile(context.Background(), ReidentifyFileRequest{DataSource: SkyReidentifySourceTypeBase64, Value: redacted, DataFormat: FileDataFormatTXT},
		ReidentifyFileOptions{})
	if err != nil || resp.Status != DetectRunStatusSuccess || resp.OutputType != SkyReidentifySourceTypeBase64 ||
		len(resp.Output) != 1 || resp.Output[0].ProcessedFile != plain {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if body := f.last().body; body["vaultId"] != "vault1" || body["value"] != redacted || body["dataFormat"] != "txt" || body["dataSource"] != "BASE64" {
		t.Fatalf("body = %v", body)
	}
}

// A local file is always sent as base64: DataSource is ignored next to it.
func TestDeidentifyFileIgnoresDataSourceForALocalFile(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	c, _ := newTestController(f)
	input := filepath.Join(t.TempDir(), "in.txt")
	if err := os.WriteFile(input, []byte("data"), outputFileMode); err != nil {
		t.Fatal(err)
	}

	if _, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{FilePath: input, DataSource: SkyDeidentifySourceTypePresignedURL, ConfigurationId: "cfg"},
		DeidentifyFileOptions{}); err != nil {
		t.Fatal(err)
	}
	if body := f.last().body; body["dataSource"] != "BASE64" || body["value"] != base64.StdEncoding.EncodeToString([]byte("data")) {
		t.Fatalf("body = %v", body)
	}
}

func TestCheckGuardrails(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathGuardrails] = []string{`{"text":"hi","validation":"FAILED","toxic":true,"deniedTopic":false}`}
	c, _ := newTestController(f)

	resp, err := c.CheckGuardrails(context.Background(), CheckGuardrailsRequest{Text: "hi", CheckToxicity: ptr(true), DenyTopics: []string{"politics"}}, CheckGuardrailsOptions{})
	if err != nil || resp.Validation != GuardrailsValidationFailed || resp.Toxic == nil || !*resp.Toxic || resp.DeniedTopic == nil || *resp.DeniedTopic {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if body := f.last().body; body["vaultId"] != "vault1" || body["checkToxicity"] != true {
		t.Fatalf("body = %v", body)
	}
	// false is sent, so the server's default (true) can be turned off; nil is omitted.
	f.responses[pathGuardrails] = []string{`{"validation":"PASSED"}`, `{"validation":"PASSED"}`}
	if _, err := c.CheckGuardrails(context.Background(), CheckGuardrailsRequest{Text: "hi", CheckToxicity: ptr(false)}, CheckGuardrailsOptions{}); err != nil {
		t.Fatal(err)
	}
	if got, sent := f.last().body["checkToxicity"]; !sent || got != false {
		t.Fatalf("checkToxicity = %v (sent %v), want false", got, sent)
	}
	if _, err := c.CheckGuardrails(context.Background(), CheckGuardrailsRequest{Text: "hi"}, CheckGuardrailsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, sent := f.last().body["checkToxicity"]; sent {
		t.Fatal("an unset checkToxicity should be omitted")
	}
	// DenyTopics are sent as given, as in the other SDKs.
	f.responses[pathGuardrails] = []string{`{"validation":"PASSED"}`}
	if _, err := c.CheckGuardrails(context.Background(), CheckGuardrailsRequest{Text: "hi", DenyTopics: []string{" "}}, CheckGuardrailsOptions{}); err != nil {
		t.Fatalf("blank deny topic: err = %v", err)
	}
}

// A missing or unknown validation is UNKNOWN, and missing flags stay nil.
func TestCheckGuardrailsUnknownValidation(t *testing.T) {
	for _, body := range []string{`{"text":"hi"}`, `{"text":"hi","validation":"MAYBE"}`} {
		f := newFakeDetect(t)
		f.responses[pathGuardrails] = []string{body}
		c, _ := newTestController(f)
		resp, err := c.CheckGuardrails(context.Background(), CheckGuardrailsRequest{Text: "hi"}, CheckGuardrailsOptions{})
		if err != nil || resp.Validation != GuardrailsValidationUnknown || resp.Toxic != nil || resp.DeniedTopic != nil {
			t.Fatalf("%s: resp = %+v, err = %v", body, resp, err)
		}
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

func TestOutputFileNamesStayUniqueWhenOutputsRepeat(t *testing.T) {
	used := map[string]bool{}
	redacted := FileOutput{ProcessedFileType: ProcessedFileTypeRedactedFile, ProcessedFileExtension: FileDataFormatPDF}
	unknown := FileOutput{ProcessedFileType: "NEW_TYPE", ProcessedFileExtension: FileDataFormatPDF}
	var names []string
	for _, output := range []FileOutput{redacted, redacted, redacted, unknown, unknown} {
		names = append(names, outputFileName("doc", output, used))
	}
	want := []string{"processed-doc.pdf", "processed-doc.redacted_file.pdf", "processed-doc.redacted_file-2.pdf", "processed-doc.3.pdf", "processed-doc.4.pdf"}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("names = %v, want %v", names, want)
		}
	}
}

func TestWriteOutputsKeepsEveryOutputOfTheSameType(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	var outputs []FileOutput
	for _, content := range []string{"one", "two", "three"} {
		outputs = append(outputs, FileOutput{ProcessedFile: base64.StdEncoding.EncodeToString([]byte(content)),
			ProcessedFileType: ProcessedFileTypeRedactedFile, ProcessedFileExtension: FileDataFormatTXT})
	}
	dir := t.TempDir()
	if err := c.writeOutputs(outputs, SkyDeidentifySourceTypeBase64, "in.txt", dir); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"processed-in.txt":                 "one",
		"processed-in.redacted_file.txt":   "two",
		"processed-in.redacted_file-2.txt": "three",
	} {
		if got, err := os.ReadFile(filepath.Join(dir, name)); err != nil || string(got) != want {
			t.Errorf("%s = %q, %v; want %q", name, got, err, want)
		}
	}
}

func TestFileObjectIsSentInFullWhateverItsOffset(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`, `{"runId":"run-2"}`}
	c, _ := newTestController(f)

	input, err := os.Create(filepath.Join(t.TempDir(), "in.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = input.Close() }()
	_, _ = input.WriteString("My name is John.") // the offset is now at the end
	want := base64.StdEncoding.EncodeToString([]byte("My name is John."))

	// Sent twice: the first read must not consume the file either.
	for i := 0; i < 2; i++ {
		if _, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{FileObject: input, ConfigurationId: "cfg"}, DeidentifyFileOptions{}); err != nil {
			t.Fatal(err)
		}
		if got := f.last().body["value"]; got != want {
			t.Fatalf("call %d sent value %q, want %q", i+1, got, want)
		}
	}
}

// As in the other SDKs, a status this SDK does not know is UNKNOWN, which ends polling, and
// the run ID is trimmed before the lookup.
func TestUnknownRunStatusIsTerminal(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathRuns+"run-3"] = []string{`{"status":"PAUSED"}`}
	f.responses[pathReidentifyFile] = []string{`{"outputType":"BASE64","output":[]}`}
	c, pauses := newTestController(f)

	run, err := c.GetRun(context.Background(), GetRunRequest{RunId: " run-3 "}, GetRunOptions{})
	if err != nil || run.Status != DetectRunStatusUnknown || run.RunId != "run-3" || f.last().path != pathRuns+"run-3" {
		t.Fatalf("run = %+v, err = %v, path = %s", run, err, f.last().path)
	}
	if !isTerminal(runStatus("PAUSED")) || runStatus("QUEUED") != DetectRunStatusQueued {
		t.Fatal("unknown statuses must be terminal; known ones kept")
	}

	resp, err := c.ReidentifyFile(context.Background(), ReidentifyFileRequest{DataSource: SkyReidentifySourceTypeBase64, Value: "aGk="}, ReidentifyFileOptions{})
	if err != nil || resp.Status != DetectRunStatusUnknown {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if len(*pauses) != 0 {
		t.Fatalf("pauses = %v", *pauses)
	}
}

// As in the other SDKs, an output type or file extension this SDK doesn't know is empty.
// A file extension is matched case-insensitively, as Java does.
func TestUnknownOutputTypeAndExtensionAreEmpty(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathRuns+"run-4"] = []string{`{"status":"SUCCESS","outputType":"SOMETHING","output":[{"processedFileExtension":"PDF"},{"processedFileExtension":"xyz"}]}`}
	f.responses[pathReidentifyFile] = []string{`{"status":"SUCCESS","outputType":"SKYFLOW_ID"}`}
	c, _ := newTestController(f)

	run, err := c.GetRun(context.Background(), GetRunRequest{RunId: "run-4"}, GetRunOptions{})
	if err != nil || run.OutputType != nil || run.Output[0].ProcessedFileExtension != FileDataFormatPDF || run.Output[1].ProcessedFileExtension != "" {
		t.Fatalf("run = %+v, err = %v", run, err)
	}
	resp, err := c.ReidentifyFile(context.Background(), ReidentifyFileRequest{DataSource: SkyReidentifySourceTypeBase64, Value: "aGk="}, ReidentifyFileOptions{})
	if err != nil || resp.OutputType != nil {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}

	// The output type is the vault type's own source: a FlowVault controller reports SKYFLOW_ID
	// as nil and the others as FlowDeidentifySourceType values.
	flow := NewController(c.base, FlowSources)
	flow.sleep = c.sleep
	f.responses[pathRuns+"run-5"] = []string{`{"status":"SUCCESS","outputType":"SKYFLOW_ID"}`, `{"status":"SUCCESS","outputType":"PRESIGNED_URL"}`}
	if run, err := flow.GetRun(context.Background(), GetRunRequest{RunId: "run-5"}, GetRunOptions{}); err != nil || run.OutputType != nil {
		t.Fatalf("run = %+v, err = %v", run, err)
	}
	run, err = flow.GetRun(context.Background(), GetRunRequest{RunId: "run-5"}, GetRunOptions{})
	if err != nil || run.OutputType != FlowDeidentifySourceTypePresignedURL || run.OutputType == SkyDeidentifySourceTypePresignedURL {
		t.Fatalf("run = %+v, err = %v", run, err)
	}
	// Re-identify output types resolve the same way: each vault type has its own.
	f.responses[pathReidentifyFile] = []string{`{"status":"SUCCESS","outputType":"PRESIGNED_URL"}`}
	flowResp, err := flow.ReidentifyFile(context.Background(), ReidentifyFileRequest{DataSource: FlowReidentifySourceTypeBase64, Value: "aGk="}, ReidentifyFileOptions{})
	if err != nil || flowResp.OutputType != FlowReidentifySourceTypePresignedURL || flowResp.OutputType == SkyReidentifySourceTypePresignedURL {
		t.Fatalf("flow reidentify = %+v, err = %v", flowResp, err)
	}
	skyResp, err := c.ReidentifyFile(context.Background(), ReidentifyFileRequest{DataSource: SkyReidentifySourceTypeBase64, Value: "aGk="}, ReidentifyFileOptions{})
	if err != nil || skyResp.OutputType != SkyReidentifySourceTypePresignedURL {
		t.Fatalf("sky reidentify = %+v, err = %v", skyResp, err)
	}
	f.responses[pathRuns+"run-6"] = []string{`{"status":"SUCCESS","outputType":"SKYFLOW_ID"}`}
	if run, err := c.GetRun(context.Background(), GetRunRequest{RunId: "run-6"}, GetRunOptions{}); err != nil || run.OutputType != SkyDeidentifySourceTypeSkyflowID {
		t.Fatalf("sky run = %+v, err = %v", run, err)
	}
}

// Every metric is optional in the API: a count the server leaves out is nil, unlike a
// count of zero, as in the Java SDK.
func TestMissingMetricsAreNil(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyString] = []string{`{"processedText":"x","metrics":{"wordCount":0}}`}
	c, _ := newTestController(f)
	resp, err := c.DeidentifyString(context.Background(), DeidentifyStringRequest{Text: "x"}, DeidentifyStringOptions{})
	if err != nil {
		t.Fatal(err)
	}
	m := resp.Metrics
	if m == nil || m.WordCount == nil || *m.WordCount != 0 || m.CharacterCount != nil || m.Size != nil {
		t.Fatalf("metrics = %+v", m)
	}
}

func TestPollOptionsDefaultsApplyOnlyToNilFields(t *testing.T) {
	if wait, attempts := (PollOptions{}).effective(); wait != defaultWaitTimeSeconds*time.Second || attempts != defaultMaxAttempts {
		t.Errorf("PollOptions{}: wait = %v, attempts = %d", wait, attempts)
	}
	if wait, attempts := (PollOptions{WaitTime: ptr(5 * time.Second), MaxAttempts: ptr(3)}).effective(); wait != 5*time.Second || attempts != 3 {
		t.Errorf("set fields: wait = %v, attempts = %d", wait, attempts)
	}
}

// FileDataFormat values are upper case, as in the Java SDK, while the API's wire
// values are lower case: a request's DataFormat is sent lower case whatever its case, and a
// response's ProcessedFileExtension comes back as the upper-case constant.
func TestFileDataFormatIsUpperCaseAndWireIsLowerCase(t *testing.T) {
	t.Run("constants are upper case", func(t *testing.T) {
		if FileDataFormatTXT != "TXT" || FileDataFormatJSONL != "JSONL" {
			t.Fatalf("constants = %q, %q", FileDataFormatTXT, FileDataFormatJSONL)
		}
	})

	t.Run("deidentify sends lower case whatever the case", func(t *testing.T) {
		f := newFakeDetect(t)
		f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
		c, _ := newTestController(f)
		for _, format := range []FileDataFormat{FileDataFormatCSV, "csv", "Csv"} {
			if _, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{
				DataSource: SkyDeidentifySourceTypeBase64, Value: "aGk=", DataFormat: format, ConfigurationId: "cfg",
			}, DeidentifyFileOptions{}); err != nil {
				t.Fatal(err)
			}
			if got := f.last().body["dataFormat"]; got != "csv" {
				t.Errorf("DataFormat %q sent as %v", format, got)
			}
		}
	})

	t.Run("reidentify sends lower case and returns upper case", func(t *testing.T) {
		f := newFakeDetect(t)
		f.responses[pathReidentifyFile] = []string{`{"status":"SUCCESS","outputType":"BASE64","output":[{"processedFile":"aGk=","processedFileType":"REIDENTIFIED_FILE","processedFileExtension":"docx"}]}`}
		c, _ := newTestController(f)
		resp, err := c.ReidentifyFile(context.Background(), ReidentifyFileRequest{
			DataSource: SkyReidentifySourceTypeBase64, Value: "aGk=", DataFormat: FileDataFormatDOCX,
		}, ReidentifyFileOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got := f.last().body["dataFormat"]; got != "docx" {
			t.Errorf("reidentify DataFormat sent as %v", got)
		}
		if got := resp.Output[0].ProcessedFileExtension; got != FileDataFormatDOCX {
			t.Errorf("ProcessedFileExtension = %q, want %q", got, FileDataFormatDOCX)
		}
	})

	t.Run("output file names keep the lower-case extension", func(t *testing.T) {
		if got := outputFileName("notes", FileOutput{ProcessedFileType: ProcessedFileTypeRedactedFile, ProcessedFileExtension: FileDataFormatPDF}, map[string]bool{}); got != "processed-notes.pdf" {
			t.Errorf("output file name = %q", got)
		}
	})

	t.Run("format from name ignores case", func(t *testing.T) {
		if got := formatFromName("Scan.PNG"); got != FileDataFormatPNG {
			t.Errorf("formatFromName = %q", got)
		}
	})
}
