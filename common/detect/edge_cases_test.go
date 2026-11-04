package detect

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/skyflowapi/skyflow-go/common/config"
	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
)

// detectCall runs one detect operation against the fake and returns only its error.
type detectCall func(c *Controller) *skyflowError.SkyflowError

func allOperations(dir string) map[string]detectCall {
	ctx := context.Background()
	return map[string]detectCall{
		"DeidentifyString": func(c *Controller) *skyflowError.SkyflowError {
			_, err := c.DeidentifyString(ctx, DeidentifyStringRequest{Text: "x", ConfigurationId: "cfg"}, DeidentifyStringOptions{})
			return err
		},
		"DeidentifyFile": func(c *Controller) *skyflowError.SkyflowError {
			_, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64, ConfigurationId: "cfg"}, DeidentifyFileOptions{})
			return err
		},
		"GetRun": func(c *Controller) *skyflowError.SkyflowError {
			_, err := c.GetRun(ctx, GetRunRequest{RunId: "run-1"}, GetRunOptions{})
			return err
		},
		"ReidentifyString": func(c *Controller) *skyflowError.SkyflowError {
			_, err := c.ReidentifyString(ctx, ReidentifyStringRequest{Text: "x"}, ReidentifyStringOptions{})
			return err
		},
		"ReidentifyFile": func(c *Controller) *skyflowError.SkyflowError {
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{Value: "v", DataSource: ReidentifySourceTypeBase64}, ReidentifyFileOptions{})
			return err
		},
		"CheckGuardrails": func(c *Controller) *skyflowError.SkyflowError {
			_, err := c.CheckGuardrails(ctx, CheckGuardrailsRequest{Text: "x"}, CheckGuardrailsOptions{})
			return err
		},
	}
}

func TestOperationsMapApiErrors(t *testing.T) {
	f := newFakeDetect(t)
	f.status = http.StatusNotFound
	c, _ := newTestController(f)

	for name, call := range allOperations(t.TempDir()) {
		err := call(c)
		if err == nil || err.GetHttpCode() != "404" || err.GetRequestId() != "req-123" {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestOperationsRejectEmptyResponses(t *testing.T) {
	f := newFakeDetect(t)
	for _, path := range []string{pathDeidentifyString, pathRuns + "run-1", pathReidentifyString, pathReidentifyFile, pathGuardrails} {
		f.responses[path] = []string{`null`}
	}
	// A submit response without a run ID is as unusable as an empty one.
	f.responses[pathDeidentifyFile] = []string{`{}`}
	c, _ := newTestController(f)

	for name, call := range allOperations(t.TempDir()) {
		err := call(c)
		if err == nil || !strings.Contains(err.GetMessage(), fmt.Sprintf(errEmptyResponse, name)) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestApiErrorWithoutApiDetails(t *testing.T) {
	err := apiError(errors.New("connection reset"))
	if err == nil || !strings.Contains(err.GetMessage(), "connection reset") {
		t.Fatalf("err = %v", err)
	}
	err = apiError(core.NewAPIError(http.StatusBadGateway, nil, errors.New("bad gateway")))
	if err == nil || err.GetRequestId() != "" {
		t.Fatalf("API error without a header: err = %v", err)
	}
}

func TestValidationFailuresPerOperation(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	ctx := context.Background()
	cases := []struct {
		name string
		want string
		call func() *skyflowError.SkyflowError
	}{
		{"reidentify string, blank text", fmt.Sprintf(errEmptyText, methodReidentifyString), func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyString(ctx, ReidentifyStringRequest{Text: " "}, ReidentifyStringOptions{})
			return err
		}},
		{"guardrails, empty text", fmt.Sprintf(errEmptyText, methodCheckGuardrails), func() *skyflowError.SkyflowError {
			_, err := c.CheckGuardrails(ctx, CheckGuardrailsRequest{Text: ""}, CheckGuardrailsOptions{})
			return err
		}},
		{"reidentify file, no data source", fmt.Sprintf(errMissingReidentifySource, methodReidentifyFile, "BASE64 or PRESIGNED_URL"), func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{Value: "v"}, ReidentifyFileOptions{})
			return err
		}},
		{"reidentify file, unsupported data source", fmt.Sprintf(errUnsupportedReidentifySource, methodReidentifyFile, "SKYFLOW_ID", "BASE64 or PRESIGNED_URL"), func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{DataSource: "SKYFLOW_ID", Value: "v"}, ReidentifyFileOptions{})
			return err
		}},
		{"reidentify file, blank value", fmt.Sprintf(errEmptyValue, methodReidentifyFile), func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{DataSource: ReidentifySourceTypeBase64, Value: " "}, ReidentifyFileOptions{})
			return err
		}},
		{"reidentify file, redaction level without a source", fmt.Sprintf(errRedactionSource, methodReidentifyFile, 0), func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{Value: "v", DataSource: ReidentifySourceTypeBase64, RedactionLevel: []RedactionLevel{{RedactionType: RedactionTypeMasked}}}, ReidentifyFileOptions{})
			return err
		}},
		{"deidentify file, value without data source", fmt.Sprintf(errMissingDeidentifySource, methodDeidentifyFile, "BASE64, SKYFLOW_ID or PRESIGNED_URL"), func() *skyflowError.SkyflowError {
			_, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", ConfigurationId: "cfg"}, DeidentifyFileOptions{})
			return err
		}},
	}
	for _, tc := range cases {
		err := tc.call()
		if err == nil || err.GetHttpCode() != string(skyflowError.INVALID_INPUT_CODE) || !strings.Contains(err.GetMessage(), tc.want) {
			t.Errorf("%s: err = %v", tc.name, err)
		}
	}
}

func TestDeidentifyFileReturnsPollErrors(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	// No response registered for the run lookup: the fake answers 500.
	c, _ := newTestController(f)

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64, ConfigurationId: "cfg"},
		DeidentifyFileOptions{Poll: &PollOptions{}})
	// The run exists on the server, so its ID comes back with the error.
	if err == nil || resp == nil || resp.RunId != "run-1" || resp.Status != "" || err.GetHttpCode() != "500" || err.GetRequestId() != "req-123" {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
}

func TestDeidentifyFileKeepsRunIdWhenCancelledDuringALookup(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	f.responses[pathRuns+"run-1"] = []string{`{"status":"IN_PROGRESS"}`, `{"status":"IN_PROGRESS"}`}
	c, _ := newTestController(f)
	ctx, cancel := context.WithCancel(context.Background())
	// The pause completes but cancels ctx, so the next lookup's request fails.
	c.sleep = func(context.Context, time.Duration) bool { cancel(); return true }

	resp, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64},
		DeidentifyFileOptions{Poll: &PollOptions{}})
	if err == nil || resp == nil || resp.RunId != "run-1" {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
}

func TestFileOperationsReturnResponseWhenOutputsCannotBeWritten(t *testing.T) {
	f := newFakeDetect(t)
	output := `{"status":"SUCCESS","outputType":"BASE64","output":[{"processedFile":"` +
		base64.StdEncoding.EncodeToString([]byte("data")) + `","processedFileType":"REDACTED_FILE","processedFileExtension":"txt"}]}`
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	f.responses[pathRuns+"run-1"] = []string{output}
	c, _ := newTestController(f)

	dir := t.TempDir()
	input := filepath.Join(dir, "in.txt")
	blocker := filepath.Join(dir, "not-a-dir")
	for _, path := range []string{input, blocker} {
		if err := os.WriteFile(path, []byte("x"), outputFileMode); err != nil {
			t.Fatal(err)
		}
	}
	outDir := filepath.Join(blocker, "out") // MkdirAll fails: a parent is a regular file

	deidentified, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{FilePath: input, ConfigurationId: "cfg"},
		DeidentifyFileOptions{Poll: &PollOptions{}, OutputDirectory: outDir})
	if err == nil || deidentified == nil || deidentified.Status != DetectRunStatusSuccess || !strings.Contains(err.GetMessage(), outputNotWritten(outDir)) {
		t.Fatalf("deidentify: resp = %+v, err = %v", deidentified, err)
	}
}

func TestWriteOutputs(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	encoded := base64.StdEncoding.EncodeToString([]byte("data"))
	pdf := FileOutput{ProcessedFile: encoded, ProcessedFileType: ProcessedFileTypeRedactedFile, ProcessedFileExtension: FileDataFormatPDF}

	t.Run("nothing to write", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "never-created")
		if err := c.writeOutputs(nil, DeidentifySourceTypeBase64, "a.pdf", dir); err != nil {
			t.Fatal(err)
		}
		if err := c.writeOutputs([]FileOutput{pdf}, DeidentifySourceTypePresignedURL, "a.pdf", dir); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatalf("directory created for unwritten outputs: %v", err)
		}
	})

	t.Run("current directory, default name, empty outputs skipped", func(t *testing.T) {
		chdir(t, t.TempDir())
		if err := c.writeOutputs([]FileOutput{{}, pdf}, DeidentifySourceTypeBase64, ".pdf", ""); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile("processed-output.pdf")
		if err != nil || string(got) != "data" {
			t.Fatalf("output = %q, err = %v", got, err)
		}
		info, err := os.Stat("processed-output.pdf")
		if err != nil || info.Mode().Perm() != outputFileMode {
			t.Fatalf("info = %v, err = %v", info, err)
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		bad := FileOutput{ProcessedFile: "%%%", ProcessedFileExtension: FileDataFormatTXT}
		dir := t.TempDir()
		target := filepath.Join(dir, "processed-a.txt")
		if err := c.writeOutputs([]FileOutput{bad}, DeidentifySourceTypeBase64, "a.txt", dir); err == nil || !strings.Contains(err.GetMessage(), outputNotWritten(target)) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("target is a directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "processed-a.pdf"), outputDirMode); err != nil {
			t.Fatal(err)
		}
		if err := c.writeOutputs([]FileOutput{pdf}, DeidentifySourceTypeBase64, "a.pdf", dir); err == nil || !strings.Contains(err.GetMessage(), notRegularFile) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("existing file is replaced with an owner-only file", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "processed-a.pdf")
		if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := c.writeOutputs([]FileOutput{pdf}, DeidentifySourceTypeBase64, "a.pdf", dir); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(target)
		if err != nil || string(got) != "data" {
			t.Fatalf("output = %q, err = %v", got, err)
		}
		info, err := os.Stat(target)
		if err != nil || info.Mode().Perm() != outputFileMode {
			t.Fatalf("info = %v, err = %v", info, err)
		}
	})

	t.Run("symlink at target is not followed", func(t *testing.T) {
		dir := t.TempDir()
		elsewhere := filepath.Join(t.TempDir(), "victim")
		if err := os.WriteFile(elsewhere, []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(elsewhere, filepath.Join(dir, "processed-a.pdf")); err != nil {
			t.Skipf("symlinks not supported: %v", err)
		}
		if err := c.writeOutputs([]FileOutput{pdf}, DeidentifySourceTypeBase64, "a.pdf", dir); err == nil || !strings.Contains(err.GetMessage(), notRegularFile) {
			t.Fatalf("err = %v", err)
		}
		if got, err := os.ReadFile(elsewhere); err != nil || string(got) != "keep" {
			t.Fatalf("link target = %q, err = %v", got, err)
		}
	})

	t.Run("read-only directory", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root ignores directory permissions")
		}
		for name, existing := range map[string]bool{"new file": false, "existing file": true} {
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				target := filepath.Join(dir, "processed-a.pdf")
				if existing {
					if err := os.WriteFile(target, []byte("old"), outputFileMode); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Chmod(dir, 0o500); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(dir, outputDirMode) })
				if err := c.writeOutputs([]FileOutput{pdf}, DeidentifySourceTypeBase64, "a.pdf", dir); err == nil || !strings.Contains(err.GetMessage(), outputNotWritten(target)) {
					t.Fatalf("err = %v", err)
				}
			})
		}
	})
}

func TestReadLocalFileRefusesNonRegularFiles(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	dir := t.TempDir()

	if _, err := c.readLocalFile(methodDeidentifyFile, dir, nil); err == nil || !strings.Contains(err.GetMessage(), notRegularFile) {
		t.Fatalf("file path: err = %v", err)
	}

	opened, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	if _, err := c.readLocalFile(methodDeidentifyFile, "", opened); err == nil || !strings.Contains(err.GetMessage(), notRegularFile) {
		t.Fatalf("file object: err = %v", err)
	}

	closed, err := os.Create(filepath.Join(dir, "in.txt"))
	if err != nil {
		t.Fatal(err)
	}
	_ = closed.Close()
	if _, err := c.readLocalFile(methodDeidentifyFile, "", closed); err == nil || !strings.Contains(err.GetMessage(), os.ErrClosed.Error()) {
		t.Fatalf("closed file object: err = %v", err)
	}
}

// outputNotWritten is the start of the write-failure message for path, before the OS error.
func outputNotWritten(path string) string {
	return fmt.Sprintf(errOutputNotWritten, path, "")
}

// chdir stands in for t.Chdir, which needs Go 1.24; this module targets 1.23. It changes the
// working directory of the whole process, so tests that use it must not run in parallel.
func chdir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previous) })
}

func TestOutputFileNameWithoutTypeOrExtension(t *testing.T) {
	used := map[string]bool{}
	first := outputFileName("doc", FileOutput{}, used)
	second := outputFileName("doc", FileOutput{ProcessedFileExtension: FileDataFormatUnknown}, used)
	if first != "processed-doc.bin" || second != "processed-doc.1.bin" {
		t.Fatalf("names = %q, %q", first, second)
	}
}

func TestOutputFileNameIgnoresUnknownServerValues(t *testing.T) {
	used := map[string]bool{}
	escaping := FileOutput{ProcessedFileType: "../../ENTITIES", ProcessedFileExtension: "/../../../etc/x"}
	first := outputFileName("doc", escaping, used)
	second := outputFileName("doc", escaping, used)
	if first != "processed-doc.bin" || second != "processed-doc.1.bin" {
		t.Fatalf("names = %q, %q", first, second)
	}
	for _, name := range []string{first, second} {
		if !filepath.IsLocal(name) || strings.ContainsRune(name, filepath.Separator) {
			t.Fatalf("name %q can leave the output directory", name)
		}
	}
	// Known values are matched without regard to case.
	if got := outputFileName("doc", FileOutput{ProcessedFileType: "object_entities", ProcessedFileExtension: "JSON"}, used); got != "processed-object-entities-doc.json" {
		t.Fatalf("name = %q", got)
	}
}

func TestPollReturnsLastRunWhenContextEnds(t *testing.T) {
	c, pauses := newTestController(newFakeDetect(t))
	c.sleep = func(context.Context, time.Duration) bool { return false }
	fetch := func(context.Context) (*GetRunResponse, *skyflowError.SkyflowError) {
		return &GetRunResponse{RunId: "r", Status: DetectRunStatusInProgress}, nil
	}

	run, err := c.poll(context.Background(), "r", PollOptions{}, fetch)
	if err != nil || run.RunId != "r" || run.Status != DetectRunStatusInProgress || len(*pauses) != 0 {
		t.Fatalf("run = %+v, err = %v", run, err)
	}
}

func TestSleepContext(t *testing.T) {
	if !sleepContext(context.Background(), time.Millisecond) {
		t.Error("an elapsed timer should report true")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if sleepContext(ctx, time.Hour) {
		t.Error("a cancelled context should report false")
	}
}

func TestResponsesToleratePartialData(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyString] = []string{`{"processedText":"x","entities":[null,{"token":"T"}]}`}
	f.responses[pathRuns+"run-1"] = []string{`{"output":[null,{"processedFileType":"ENTITIES"}]}`}
	c, _ := newTestController(f)

	resp, err := c.DeidentifyString(context.Background(), DeidentifyStringRequest{Text: "x", Configuration: &DetectConfiguration{}}, DeidentifyStringOptions{})
	if err != nil || len(resp.Entities) != 1 || resp.Entities[0].Token != "T" {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if detect := f.last().body["configuration"].(map[string]interface{})["detect"]; detect != nil {
		t.Fatalf("detect = %v, want omitted", detect)
	}

	run, err := c.GetRun(context.Background(), GetRunRequest{RunId: "run-1"}, GetRunOptions{})
	if err != nil || run.Status != DetectRunStatusUnknown || len(run.Output) != 1 {
		t.Fatalf("run = %+v, err = %v", run, err)
	}
}

func TestWriteOutputFileKeepsTheOldOutputWhenTheWriteFails(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "processed-in.txt")
	if err := os.WriteFile(target, []byte("old"), outputFileMode); err != nil {
		t.Fatal(err)
	}
	failRename := func(string, string) error { return errors.New("disk full") }
	if err := writeOutputFile(target, []byte("new"), failRename); err == nil || err.Error() != "disk full" {
		t.Fatalf("err = %v", err)
	}
	if got, _ := os.ReadFile(target); string(got) != "old" {
		t.Fatalf("old output = %q, want it kept", got)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Fatalf("the temporary file should be removed, dir has %d entries", len(entries))
	}
}

func TestDeidentifyFileRefreshesAuthForEachLookup(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	f.responses[pathRuns+"run-1"] = []string{`{"status":"IN_PROGRESS"}`, `{"status":"SUCCESS"}`}
	c, _ := newTestController(f)
	newKey := "sky-zzzzz-1234567890abcdef1234567890abcdef"
	c.sleep = func(context.Context, time.Duration) bool {
		c.base.SetClientSettings(vaultclient.ClientSettings{Credentials: &config.Credentials{ApiKey: newKey}})
		return true
	}

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64},
		DeidentifyFileOptions{Poll: &PollOptions{}})
	if err != nil || resp.Status != DetectRunStatusSuccess {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
	if got := f.last().header.Get("Authorization"); got != "Bearer "+newKey {
		t.Fatalf("last lookup sent %q, want the credentials set during the wait", got)
	}
}

func TestDeidentifyFileKeepsRunIdWhenAuthFailsDuringAWait(t *testing.T) {
	t.Setenv(vaultclient.CredentialsEnv, "")
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	f.responses[pathRuns+"run-1"] = []string{`{"status":"IN_PROGRESS"}`}
	c, _ := newTestController(f)
	c.sleep = func(context.Context, time.Duration) bool {
		c.base.SetClientSettings(vaultclient.ClientSettings{}) // no credentials left
		return true
	}

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeBase64},
		DeidentifyFileOptions{Poll: &PollOptions{}})
	if err == nil || resp == nil || resp.RunId != "run-1" {
		t.Fatalf("resp = %+v, err = %v", resp, err)
	}
}

func TestTransportErrorsKeepTheirCause(t *testing.T) {
	f := newFakeDetect(t)
	c, _ := newTestController(f)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.GetRun(ctx, GetRunRequest{RunId: "run-1"}, GetRunOptions{})
	if err == nil || !errors.Is(err, context.Canceled) || err.GetHttpCode() != string(skyflowError.REQUEST_CANCELLED) {
		t.Fatalf("err = %v, want it to wrap context.Canceled with code REQUEST_CANCELLED", err)
	}

	// A failure that is not the context's keeps the SERVER code.
	f.server.Close()
	_, err = c.GetRun(context.Background(), GetRunRequest{RunId: "run-1"}, GetRunOptions{})
	if err == nil || err.GetHttpCode() != string(skyflowError.SERVER) {
		t.Fatalf("err = %v, want code SERVER", err)
	}
}

func TestDeidentifyFileAcceptsOnlyTheSDKsSources(t *testing.T) {
	f := newFakeDetect(t)
	c, _ := newTestController(f)
	c = NewController(c.base, DeidentifySourceTypeBase64, DeidentifySourceTypePresignedURL)
	ctx := context.Background()

	_, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", ConfigurationId: "cfg"}, DeidentifyFileOptions{})
	if want := fmt.Sprintf(errMissingDeidentifySource, methodDeidentifyFile, "BASE64 or PRESIGNED_URL"); err == nil || !strings.Contains(err.GetMessage(), want) {
		t.Fatalf("missing source: err = %v", err)
	}
	_, err = c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: DeidentifySourceTypeSkyflowID, ConfigurationId: "cfg"}, DeidentifyFileOptions{})
	if want := fmt.Sprintf(errUnsupportedDeidentifySource, methodDeidentifyFile, DeidentifySourceTypeSkyflowID, "BASE64 or PRESIGNED_URL"); err == nil || !strings.Contains(err.GetMessage(), want) {
		t.Fatalf("unsupported source: err = %v", err)
	}
	if len(f.requests) != 0 {
		t.Fatalf("requests sent: %v", f.requests)
	}

	single := NewController(c.base, DeidentifySourceTypeBase64)
	_, err = single.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", ConfigurationId: "cfg"}, DeidentifyFileOptions{})
	if want := fmt.Sprintf(errMissingDeidentifySource, methodDeidentifyFile, "BASE64"); err == nil || !strings.Contains(err.GetMessage(), want) {
		t.Fatalf("single source: err = %v", err)
	}
}
