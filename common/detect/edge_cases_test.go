package detect

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
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
			_, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64, ConfigurationId: "cfg"}, DeidentifyFileOptions{})
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
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{Value: "v", DataSource: SkyReidentifySourceTypeBase64}, ReidentifyFileOptions{})
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

	// An empty run ID can't be polled or looked up either.
	f.responses[pathDeidentifyFile] = []string{`{"runId":""}`}
	if err := allOperations(t.TempDir())[methodDeidentifyFile](c); err == nil ||
		!strings.Contains(err.GetMessage(), fmt.Sprintf(errEmptyResponse, methodDeidentifyFile)) {
		t.Errorf("empty run ID: err = %v", err)
	}
}

func TestApiErrorWithoutApiDetails(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	err := c.apiError(context.Background(), core.NewAPIError(http.StatusBadGateway, nil, errors.New("bad gateway")))
	if err == nil || err.GetRequestId() != "" {
		t.Fatalf("API error without a header: err = %v", err)
	}
}

// A call that got no usable response is TRANSPORT_ERROR ("0", no status), unless the caller's
// context ended: then it is REQUEST_CANCELLED ("499"). The error doesn't decide: net/http's own timeout error matches context.DeadlineExceeded.
func TestApiErrorWithoutAResponse(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	reset := errors.New("connection reset")
	err := c.apiError(context.Background(), reset)
	if err.GetHttpCode() != string(skyflowError.TRANSPORT_ERROR) || err.GetHttpStatusCode() != "" ||
		!errors.Is(err, reset) || !strings.Contains(err.GetMessage(), "connection reset") {
		t.Fatalf("network failure: code = %q, status = %q, err = %v", err.GetHttpCode(), err.GetHttpStatusCode(), err)
	}
	clientTimeout := fmt.Errorf("Post: %w", context.DeadlineExceeded) // as http.Client.Timeout reports
	if err = c.apiError(context.Background(), clientTimeout); err.GetHttpCode() != string(skyflowError.TRANSPORT_ERROR) {
		t.Fatalf("client timeout: code = %q, want %q", err.GetHttpCode(), skyflowError.TRANSPORT_ERROR)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err = c.apiError(ctx, context.Canceled); err.GetHttpCode() != string(skyflowError.REQUEST_CANCELLED) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled: code = %q, err = %v", err.GetHttpCode(), err)
	}
}

// An error response has the response's HTTP status as its code, and a message even when the body isn't a Skyflow error object.
func TestApiErrorFromAnErrorResponse(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	header := http.Header{}
	header.Set("x-request-id", "req-1")
	long := strings.Repeat("x", 600)
	for name, tc := range map[string]struct {
		status  int
		body    string
		message string
	}{
		"skyflow error, body http_code ignored": {http.StatusBadRequest, `{"error":{"http_code":500,"message":"bad config","grpc_code":3,"http_status":"INVALID_ARGUMENT"}}`, "bad config"},
		"error string":                          {http.StatusForbidden, `{"error":"denied"}`, "denied"},
		"top-level message":                     {http.StatusNotFound, `{"message":"no such run"}`, "no such run"},
		"non-JSON body":                         {http.StatusBadGateway, `<html>bad gateway</html>`, "<html>bad gateway</html>"},
		"long non-JSON body":                    {http.StatusBadGateway, long, long[:512] + "... (truncated)"},
		"empty body":                            {http.StatusServiceUnavailable, ``, "Error with status code 503"},
	} {
		err := c.apiError(context.Background(), core.NewAPIError(tc.status, header, errors.New(tc.body)))
		if err.GetHttpCode() != fmt.Sprint(tc.status) || err.GetMessage() != "Message: "+tc.message || err.GetRequestId() != "req-1" {
			t.Errorf("%s: code = %q, message = %q, request ID = %q", name, err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
		}
	}

	err := c.apiError(context.Background(), core.NewAPIError(http.StatusBadRequest, header, errors.New(`{"error":{"http_code":500,"message":"bad config","grpc_code":3,"http_status":"INVALID_ARGUMENT"}}`)))
	if err.GetGrpcCode() != "3" || err.GetHttpStatusCode() != "INVALID_ARGUMENT" {
		t.Errorf("grpc code = %q, status = %q", err.GetGrpcCode(), err.GetHttpStatusCode())
	}
	// GetResponseBody is the server's body as sent, not the normalized one.
	if code := err.GetResponseBody()["error"].(map[string]interface{})["http_code"]; code != float64(500) {
		t.Errorf("response body http_code = %v, want the server's 500", code)
	}
	// A multi-byte character at the cut is dropped whole, not split.
	multiByte := strings.Repeat("x", 511) + "é" + "tail"
	if got := truncateBody(multiByte); got != strings.Repeat("x", 511)+"... (truncated)" {
		t.Errorf("truncated multi-byte body = %q", got)
	}
	if body := c.apiError(context.Background(), core.NewAPIError(http.StatusBadGateway, header, errors.New("<html/>"))).GetResponseBody(); body != nil {
		t.Errorf("non-JSON response body = %v, want nil", body)
	}

	// grpcCode and httpStatus are read when the snake_case keys are missing, and an
	// error-from-client header is added to the details.
	fromClient := http.Header{}
	fromClient.Set("error-from-client", "true")
	err = c.apiError(context.Background(), core.NewAPIError(http.StatusBadRequest, fromClient,
		errors.New(`{"error":{"message":"m","grpcCode":5,"httpStatus":"NOT_FOUND","details":[{"a":1}]}}`)))
	if err.GetGrpcCode() != "5" || err.GetHttpStatusCode() != "NOT_FOUND" {
		t.Errorf("camelCase keys: grpc code = %q, status = %q", err.GetGrpcCode(), err.GetHttpStatusCode())
	}
	want := []interface{}{map[string]interface{}{"a": float64(1)}, map[string]interface{}{"errorFromClient": "true"}}
	if !reflect.DeepEqual(err.GetDetails(), want) {
		t.Errorf("details = %v, want %v", err.GetDetails(), want)
	}
	// The snake_case key wins when both are sent, and no header means no extra detail.
	err = c.apiError(context.Background(), core.NewAPIError(http.StatusBadRequest, header,
		errors.New(`{"error":{"message":"m","grpc_code":3,"grpcCode":5}}`)))
	if err.GetGrpcCode() != "3" || len(err.GetDetails()) != 0 {
		t.Errorf("grpc code = %q, details = %v", err.GetGrpcCode(), err.GetDetails())
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
		{"reidentify string, blank text", errInvalidTextReidentifyString, func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyString(ctx, ReidentifyStringRequest{Text: " "}, ReidentifyStringOptions{})
			return err
		}},
		{"guardrails, empty text", errInvalidTextCheckGuardrails, func() *skyflowError.SkyflowError {
			_, err := c.CheckGuardrails(ctx, CheckGuardrailsRequest{Text: ""}, CheckGuardrailsOptions{})
			return err
		}},
		{"reidentify file, no data source", errInvalidReidentifyDataSource, func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{Value: "v"}, ReidentifyFileOptions{})
			return err
		}},
		{"reidentify file, blank value", errInvalidReidentifyValue, func() *skyflowError.SkyflowError {
			_, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{DataSource: SkyReidentifySourceTypeBase64, Value: " "}, ReidentifyFileOptions{})
			return err
		}},
		{"deidentify file, value without data source", errInvalidDeidentifyDataSource, func() *skyflowError.SkyflowError {
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

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64, ConfigurationId: "cfg"},
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

	resp, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64, ConfigurationId: "cfg"},
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
		if err := c.writeOutputs(nil, SkyDeidentifySourceTypeBase64, "a.pdf", dir); err != nil {
			t.Fatal(err)
		}
		if err := c.writeOutputs([]FileOutput{pdf}, SkyDeidentifySourceTypePresignedURL, "a.pdf", dir); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Fatalf("directory created for unwritten outputs: %v", err)
		}
	})

	t.Run("a blank directory is the current one", func(t *testing.T) {
		chdir(t, t.TempDir())
		if err := c.writeOutputs([]FileOutput{pdf}, SkyDeidentifySourceTypeBase64, "b.pdf", "  "); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat("processed-b.pdf"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("current directory, default name, empty outputs skipped", func(t *testing.T) {
		chdir(t, t.TempDir())
		if err := c.writeOutputs([]FileOutput{{}, pdf}, SkyDeidentifySourceTypeBase64, "", ""); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile("processed-output.pdf")
		if err != nil || string(got) != "data" {
			t.Fatalf("output = %q, err = %v", got, err)
		}
		// A dotfile keeps its whole name as the stem.
		if err := c.writeOutputs([]FileOutput{pdf}, SkyDeidentifySourceTypeBase64, ".env", ""); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat("processed-.env.pdf"); err != nil {
			t.Fatalf("dotfile output: %v", err)
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
		err := c.writeOutputs([]FileOutput{bad}, SkyDeidentifySourceTypeBase64, "a.txt", dir)
		if err == nil || !strings.Contains(err.GetMessage(), outputNotWritten(target)) {
			t.Fatalf("err = %v", err)
		}
		// A write failure is local, so it has no HTTP status.
		if err.GetHttpCode() != string(skyflowError.TRANSPORT_ERROR) || err.GetHttpStatusCode() != "" || err.Unwrap() == nil {
			t.Fatalf("http code = %q, status = %q, cause = %v", err.GetHttpCode(), err.GetHttpStatusCode(), err.Unwrap())
		}
	})

	t.Run("target is a directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "processed-a.pdf"), outputDirMode); err != nil {
			t.Fatal(err)
		}
		if err := c.writeOutputs([]FileOutput{pdf}, SkyDeidentifySourceTypeBase64, "a.pdf", dir); err == nil || !errors.Is(err, errNotRegularFile) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("existing file is replaced with an owner-only file", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "processed-a.pdf")
		if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := c.writeOutputs([]FileOutput{pdf}, SkyDeidentifySourceTypeBase64, "a.pdf", dir); err != nil {
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
		if err := c.writeOutputs([]FileOutput{pdf}, SkyDeidentifySourceTypeBase64, "a.pdf", dir); err == nil || !errors.Is(err, errNotRegularFile) {
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
				if err := c.writeOutputs([]FileOutput{pdf}, SkyDeidentifySourceTypeBase64, "a.pdf", dir); err == nil || !strings.Contains(err.GetMessage(), outputNotWritten(target)) {
					t.Fatalf("err = %v", err)
				}
			})
		}
	})
}

func TestReadLocalFileRefusesNonRegularFiles(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	dir := t.TempDir()

	if _, err := c.readLocalFile(dir, nil); err == nil || !strings.Contains(err.GetMessage(), errFileNotFound) {
		t.Fatalf("file path: err = %v", err)
	}

	opened, err := os.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	if _, err := c.readLocalFile("", opened); err == nil || !strings.Contains(err.GetMessage(), errFileNotFound) {
		t.Fatalf("file object: err = %v", err)
	}

	closed, err := os.Create(filepath.Join(dir, "in.txt"))
	if err != nil {
		t.Fatal(err)
	}
	_ = closed.Close()
	if _, err := c.readLocalFile("", closed); err == nil || !strings.Contains(err.GetMessage(), errFailedToEncodeFile) {
		t.Fatalf("closed file object: err = %v", err)
	}
}

// outputNotWritten is the write-failure message for path. It has no cause; the cause is the error's Unwrap.
func outputNotWritten(path string) string {
	return fmt.Sprintf(errOutputNotWritten, path)
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
	second := outputFileName("doc", FileOutput{ProcessedFileExtension: "unknown"}, used)
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
	if err != nil || len(resp.Entities) != 1 || resp.Entities[0].Token != "T" || resp.Entities[0].EntityScores == nil {
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

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64, ConfigurationId: "cfg"},
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

	resp, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{Value: "v", DataSource: SkyDeidentifySourceTypeBase64, ConfigurationId: "cfg"},
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

	// A failure that is not the context's, such as a refused connection, is TRANSPORT_ERROR:
	// code "0" and no status.
	f.server.Close()
	_, err = c.GetRun(context.Background(), GetRunRequest{RunId: "run-1"}, GetRunOptions{})
	if err == nil || err.GetHttpCode() != string(skyflowError.TRANSPORT_ERROR) || err.GetHttpStatusCode() != "" || err.Unwrap() == nil {
		t.Fatalf("err = %v, want code TRANSPORT_ERROR with no status and its cause", err)
	}
}

// The source rules: exactly one of FilePath, FileObject and Value, each
// counted when set; Value then needs a DataSource and must not be blank. The DataSource is
// sent as given, so the server decides whether its vault accepts it.
func TestDeidentifyFileSourceRules(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathDeidentifyFile] = []string{`{"runId":"run-1"}`}
	c, _ := newTestController(f)
	c = NewController(c.base, FlowSources)
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "in.txt")
	if err := os.WriteFile(file, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	opened, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = opened.Close() }()
	var nilFlow *FlowDeidentifySourceType
	var nilSky *SkyDeidentifySourceType
	cases := []struct {
		name    string
		request DeidentifyFileRequest
		want    string
	}{
		{"none", DeidentifyFileRequest{DataSource: FlowDeidentifySourceTypeBase64}, errMissingFileSource},
		{"path and value", DeidentifyFileRequest{FilePath: file, Value: "v", DataSource: FlowDeidentifySourceTypeBase64}, errMultipleFileSources},
		{"path and object", DeidentifyFileRequest{FilePath: file, FileObject: opened}, errMultipleFileSources},
		{"object and blank value", DeidentifyFileRequest{FileObject: opened, Value: " "}, errMultipleFileSources},
		{"value without data source", DeidentifyFileRequest{Value: "v"}, errInvalidDeidentifyDataSource},
		{"value with an empty data source", DeidentifyFileRequest{Value: "v", DataSource: FlowDeidentifySourceType("")}, errInvalidDeidentifyDataSource},
		{"value with a nil data source pointer", DeidentifyFileRequest{Value: "v", DataSource: nilFlow}, errInvalidDeidentifyDataSource},
		{"value with a nil SkyVault data source pointer", DeidentifyFileRequest{Value: "v", DataSource: nilSky}, errInvalidDeidentifyDataSource},
		{"blank value", DeidentifyFileRequest{Value: "  ", DataSource: FlowDeidentifySourceTypeBase64}, errInvalidDeidentifyValue},
		{"missing file", DeidentifyFileRequest{FilePath: filepath.Join(t.TempDir(), "missing.txt")}, errFileNotFound},
	}
	for _, tc := range cases {
		_, err := c.DeidentifyFile(ctx, tc.request, DeidentifyFileOptions{})
		if err == nil || err.GetHttpCode() != string(skyflowError.INVALID_INPUT_CODE) || !strings.Contains(err.GetMessage(), tc.want) {
			t.Errorf("%s: err = %v, want %q", tc.name, err, tc.want)
		}
	}
	if len(f.requests) != 0 {
		t.Fatalf("requests sent: %v", f.requests)
	}

	// A DataSource the vault type doesn't list is sent as given.
	if _, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "sid", DataSource: FlowDeidentifySourceType("SKYFLOW_ID")}, DeidentifyFileOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := f.last().body["dataSource"]; got != "SKYFLOW_ID" {
		t.Fatalf("dataSource = %v", got)
	}
	// A pointer to a source works too.
	presigned := FlowDeidentifySourceTypePresignedURL
	if _, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "https://x", DataSource: &presigned}, DeidentifyFileOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := f.last().body["dataSource"]; got != "PRESIGNED_URL" {
		t.Fatalf("dataSource = %v", got)
	}
	// A blank OutputDirectory counts as unset, so it needs no Poll.
	if _, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{Value: "v", DataSource: FlowDeidentifySourceTypeBase64}, DeidentifyFileOptions{OutputDirectory: " "}); err != nil {
		t.Fatal(err)
	}
	// A local file is sent as the vault type's BASE64, ignoring DataSource.
	if _, err := c.DeidentifyFile(ctx, DeidentifyFileRequest{FilePath: file, DataSource: FlowDeidentifySourceTypePresignedURL}, DeidentifyFileOptions{}); err != nil {
		t.Fatal(err)
	}
	if got := f.last().body["dataSource"]; got != "BASE64" {
		t.Fatalf("local file dataSource = %v", got)
	}
}

// A ReidentifyFile DataSource is sent as given, whatever its value or
// form; a missing one, including a nil pointer or an empty value, is rejected.
func TestReidentifyFileDataSource(t *testing.T) {
	f := newFakeDetect(t)
	f.responses[pathReidentifyFile] = []string{`{"status":"SUCCESS","outputType":"BASE64"}`}
	c, _ := newTestController(f)
	ctx := context.Background()
	presigned := SkyReidentifySourceTypePresignedURL
	for source, want := range map[ReidentifyFileSource]string{
		SkyReidentifySourceType("SKYFLOW_ID"): "SKYFLOW_ID",
		FlowReidentifySourceTypeBase64:        "BASE64",
		&presigned:                            "PRESIGNED_URL",
	} {
		if _, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{DataSource: source, Value: "v"}, ReidentifyFileOptions{}); err != nil {
			t.Fatal(err)
		}
		if got := f.last().body["dataSource"]; got != want {
			t.Fatalf("dataSource = %v, want %s", got, want)
		}
	}
	var nilFlow *FlowReidentifySourceType
	var nilSky *SkyReidentifySourceType
	for _, source := range []ReidentifyFileSource{nil, nilFlow, nilSky, FlowReidentifySourceType("")} {
		if _, err := c.ReidentifyFile(ctx, ReidentifyFileRequest{DataSource: source, Value: "v"}, ReidentifyFileOptions{}); err == nil || !strings.Contains(err.GetMessage(), errInvalidReidentifyDataSource) {
			t.Fatalf("%#v: err = %v", source, err)
		}
	}
}

// A missing local file is reported before the poll options are
// checked, and a file that exists but can't be read fails to encode.
func TestDeidentifyFileLocalFileChecks(t *testing.T) {
	c, _ := newTestController(newFakeDetect(t))
	badWait := time.Duration(0)
	missing := filepath.Join(t.TempDir(), "missing.txt")
	_, err := c.DeidentifyFile(context.Background(), DeidentifyFileRequest{FilePath: missing}, DeidentifyFileOptions{Poll: &PollOptions{WaitTime: &badWait}})
	if err == nil || !strings.Contains(err.GetMessage(), errFileNotFound) {
		t.Fatalf("missing file with a bad WaitTime: err = %v, want the file-not-found error", err)
	}
	if _, err = c.DeidentifyFile(context.Background(), DeidentifyFileRequest{FilePath: t.TempDir()}, DeidentifyFileOptions{}); err == nil || !strings.Contains(err.GetMessage(), errFileNotFound) {
		t.Fatalf("directory: err = %v, want the file-not-found error", err)
	}

	if os.Geteuid() == 0 {
		t.Skip("root can read any file")
	}
	locked := filepath.Join(t.TempDir(), "locked.txt")
	if err := os.WriteFile(locked, []byte("x"), 0o000); err != nil {
		t.Fatal(err)
	}
	if _, err = c.DeidentifyFile(context.Background(), DeidentifyFileRequest{FilePath: locked}, DeidentifyFileOptions{}); err == nil || !strings.Contains(err.GetMessage(), errFailedToEncodeFile) {
		t.Fatalf("unreadable file: err = %v, want the failed-to-encode error", err)
	}
}
