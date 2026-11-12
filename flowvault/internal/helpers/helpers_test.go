package helpers

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// zeroReader is an endless source of zero bytes.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Helpers Suite")
}

var _ = Describe("HTTP config", func() {
	It("prefers vault, then client, then defaults", func() {
		vaultRetries, clientRetries := 2, 5
		resolved := ResolveHttpConfig(
			fvcommon.HttpConfig{Timeout: 5 * time.Second, MaxRetries: &vaultRetries},
			fvcommon.HttpConfig{Timeout: 9 * time.Second, MaxRetries: &clientRetries, MaxRetryDelay: time.Second},
		)
		Expect(resolved).To(Equal(ResolvedHttpConfig{
			Timeout:           5 * time.Second,
			ConnectTimeout:    10 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			MaxRetries:        2,
			InitialRetryDelay: 500 * time.Millisecond,
			MaxRetryDelay:     time.Second,
		}))

		defaults := ResolveHttpConfig(fvcommon.HttpConfig{}, fvcommon.HttpConfig{})
		Expect(defaults.Timeout).To(Equal(60 * time.Second))
		// Per-attempt timeouts default to 10s, as in Java (OkHttp) and Python.
		Expect(defaults.ConnectTimeout).To(Equal(10 * time.Second))
		Expect(defaults.ReadTimeout).To(Equal(10 * time.Second))
		Expect(defaults.WriteTimeout).To(Equal(10 * time.Second))
		Expect(defaults.MaxRetries).To(Equal(0))
		Expect(defaults.MaxRetryDelay).To(Equal(2000 * time.Millisecond))
	})

	It("does not retry when retries are off", func() {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()
		resp, err := NewHTTPClient(ResolveHttpConfig(fvcommon.HttpConfig{}, fvcommon.HttpConfig{})).Get(server.URL)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusTooManyRequests))
		Expect(atomic.LoadInt32(&calls)).To(Equal(int32(1)))
	})
})

var _ = Describe("Files", func() {
	It("picks the upload Content-Type as the Java SDK does", func() {
		Expect(UploadContentType("Report.DOCX")).To(Equal("application/vnd.openxmlformats-officedocument.wordprocessingml.document"))
		Expect(UploadContentType("page.html")).To(Equal("text/html"))
		Expect(UploadContentType("site.css")).To(HavePrefix("text/css"))
		Expect(UploadContentType("data.unknownext")).To(Equal("application/octet-stream"))
		Expect(UploadContentType("noextension")).To(Equal("application/octet-stream"))
		Expect(UploadContentType("")).To(Equal("application/octet-stream"))
	})

	It("resolves files from a path, base64 and an *os.File", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "resume.pdf")
		Expect(os.WriteFile(path, []byte("pdf"), 0o600)).To(Succeed())

		file, err := ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: path})
		Expect(err).To(BeNil())
		Expect(file).To(Equal(ResolvedFile{Name: "resume.pdf", ContentType: "application/pdf", Path: path}))

		file, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: base64.StdEncoding.EncodeToString([]byte("txt")), FileName: "notes.txt"})
		Expect(err).To(BeNil())
		Expect(file.Base64).To(Equal(base64.StdEncoding.EncodeToString([]byte("txt"))))
		Expect(file.DecodedSize).To(Equal(int64(3)))
		Expect(file.ContentType).To(Equal("text/plain"))

		handle, _ := os.Open(path)
		defer handle.Close()
		file, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FileObject: handle})
		Expect(err).To(BeNil())
		Expect(file.Name).To(Equal("resume.pdf"))
		Expect(file.Object).To(BeIdenticalTo(handle))
		Expect(file.Base64).To(BeEmpty())
	})

	It("rejects an unreadable path, a directory, a closed file and invalid base64", func() {
		// The messages are the Java SDK's; the reason is kept as the error's cause.
		_, err := ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: "/does/not/exist"})
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.INVALID_FILE_PATH))
		Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
		_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: GinkgoT().TempDir()})
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.INVALID_FILE_PATH))
		Expect(err.Error()).To(ContainSubstring("not a regular file"))
		closed, _ := os.CreateTemp(GinkgoT().TempDir(), "closed")
		_ = closed.Close()
		_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FileObject: closed})
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.INVALID_FILE_OBJECT))
		_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "%%%", FileName: "a.txt"})
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.INVALID_BASE64))
		// Strict base64, as in the Java and Python SDKs: Go's decoder would skip line breaks.
		for _, content := range []string{"cGRm\ncGRm", "cGRm\r\n", "cGRm cGRm"} {
			_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: content, FileName: "a.txt"})
			Expect(err).ToNot(BeNil(), content)
			Expect(err.GetMessage()).To(ContainSubstring(fverrors.INVALID_BASE64))
		}
		_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "cGRmcGRm", FileName: "a.txt"})
		Expect(err).To(BeNil())
	})

	It("rejects a regular file that can't be opened", func() {
		if os.Geteuid() == 0 {
			Skip("root can open a file without read permission")
		}
		locked := filepath.Join(GinkgoT().TempDir(), "locked.txt")
		Expect(os.WriteFile(locked, []byte("x"), 0o000)).To(Succeed())
		_, err := ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: locked})
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.INVALID_FILE_PATH))
		Expect(errors.Is(err, os.ErrPermission)).To(BeTrue())
	})

	It("PUTs a file without an Authorization header and reports failures", func() {
		var got *http.Request
		var body string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			got, body = r, string(raw)
			if r.URL.Path == "/forbidden" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		file := ResolvedFile{Name: "a.pdf", ContentType: "application/pdf", Base64: "cGRm", DecodedSize: 3}

		status, msg := PutFile(context.Background(), http.DefaultClient, server.URL+"/ok", file)
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(msg).To(BeEmpty())
		Expect(got.Method).To(Equal(http.MethodPut))
		Expect(got.Header.Get("Content-Type")).To(Equal("application/pdf"))
		Expect(got.Header.Get("Authorization")).To(BeEmpty())
		Expect(body).To(Equal("pdf"))

		status, msg = PutFile(context.Background(), http.DefaultClient, server.URL+"/forbidden", file)
		Expect(status).To(Equal(fvcommon.UploadStatusFailed))
		Expect(msg).To(Equal("PUT failed: 403"))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		status, msg = PutFile(ctx, http.DefaultClient, server.URL+"/ok?signature=secret", file)
		Expect(status).To(Equal(fvcommon.UploadStatusFailed))
		Expect(msg).To(Equal("PUT failed: context canceled"))
		Expect(msg).NotTo(ContainSubstring("signature"))
	})

	It("streams a path or file object with its length and leaves the file object open", func() {
		var length int64
		var body string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			length, body = r.ContentLength, string(raw)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		path := filepath.Join(GinkgoT().TempDir(), "notes.txt")
		Expect(os.WriteFile(path, []byte("header:content"), 0o600)).To(Succeed())

		status, _ := PutFile(context.Background(), http.DefaultClient, server.URL, ResolvedFile{Path: path})
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(length).To(Equal(int64(14)))
		Expect(body).To(Equal("header:content"))

		// A FileObject is uploaded whole, from its start, whatever its offset (as in the Java
		// and Python SDKs), and its offset is left where it was.
		handle, _ := os.Open(path)
		defer handle.Close()
		_, _ = handle.Seek(7, io.SeekStart)
		resolved, rerr := ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "doc", FileObject: handle})
		Expect(rerr).To(BeNil())
		status, _ = PutFile(context.Background(), http.DefaultClient, server.URL, resolved)
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(length).To(Equal(int64(14)))
		Expect(body).To(Equal("header:content"))
		status, _ = PutFile(context.Background(), http.DefaultClient, server.URL, ResolvedFile{Object: handle})
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(body).To(Equal("header:content"))
		offset, err := handle.Seek(0, io.SeekCurrent)
		Expect(err).To(BeNil())
		Expect(offset).To(Equal(int64(7)))

		empty := filepath.Join(GinkgoT().TempDir(), "empty.txt")
		Expect(os.WriteFile(empty, nil, 0o600)).To(Succeed())
		status, _ = PutFile(context.Background(), http.DefaultClient, server.URL, ResolvedFile{Path: empty})
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(length).To(Equal(int64(0)))
	})

	It("uploads with the vault's per-attempt timeouts, but no retries and no overall timeout", func() {
		api := NewHTTPClient(ResolveHttpConfig(fvcommon.HttpConfig{ReadTimeout: 7 * time.Second}, fvcommon.HttpConfig{}))
		client := vaultclient.SignedURLClient(api)
		Expect(client.Timeout).To(BeZero())
		transport, ok := client.Transport.(*http.Transport)
		Expect(ok).To(BeTrue()) // the bare transport: no retry wrapper
		Expect(transport.ResponseHeaderTimeout).To(Equal(7 * time.Second))
	})

	It("runs exactly limit tasks at once", func() {
		const limit = 5
		var running, peak, timedOut int32
		// The first limit tasks hold their slot until all of them are running, so the
		// barrier only opens if RunConcurrently really runs limit tasks at the same time.
		barrier := make(chan struct{})
		tasks := make([]func(), 20)
		for i := range tasks {
			tasks[i] = func() {
				now := atomic.AddInt32(&running, 1)
				for {
					old := atomic.LoadInt32(&peak)
					if now <= old || atomic.CompareAndSwapInt32(&peak, old, now) {
						break
					}
				}
				if now == limit {
					select {
					case <-barrier:
					default:
						close(barrier)
					}
				}
				select {
				case <-barrier:
				case <-time.After(5 * time.Second):
					atomic.StoreInt32(&timedOut, 1)
				}
				atomic.AddInt32(&running, -1)
			}
		}
		RunConcurrently(tasks, limit)
		Expect(atomic.LoadInt32(&timedOut)).To(BeZero(), "fewer than limit tasks ran at once")
		Expect(atomic.LoadInt32(&peak)).To(Equal(int32(limit)))
		Expect(atomic.LoadInt32(&running)).To(Equal(int32(0)))
	})

	It("runs tasks one at a time when the limit is below 1", func() {
		for _, limit := range []int{0, -3} {
			var done int32
			tasks := []func(){
				func() { atomic.AddInt32(&done, 1) },
				func() { atomic.AddInt32(&done, 1) },
			}
			RunConcurrently(tasks, limit)
			Expect(atomic.LoadInt32(&done)).To(Equal(int32(2)))
		}
	})
})

var _ = Describe("Parsing", func() {
	It("parses tokens and hashed data in object, list and string form", func() {
		tokens := ParseTokens(map[string]interface{}{
			"email": []interface{}{map[string]interface{}{"token": "tok", "tokenGroupName": "det", "path": "$.a"}},
			"ssn":   "plain-token",
		})
		Expect(tokens["email"]).To(Equal([]fvcommon.Token{{Token: "tok", TokenGroupName: "det", Path: "$.a"}}))
		Expect(tokens["ssn"]).To(Equal([]fvcommon.Token{{Token: "plain-token"}}))
		Expect(ParseTokens(nil)).To(BeNil())
	})

	It("accepts table or tableName in detokenize metadata", func() {
		Expect(ParseDetokenizeMetadata(map[string]interface{}{"skyflowID": "id", "table": "t1"})).
			To(Equal(&fvcommon.DetokenizeResponseRecordMetadata{SkyflowId: "id", TableName: "t1"}))
		Expect(ParseDetokenizeMetadata(map[string]interface{}{"skyflowId": "id", "tableName": "t2"})).
			To(Equal(&fvcommon.DetokenizeResponseRecordMetadata{SkyflowId: "id", TableName: "t2"}))
		Expect(ParseDetokenizeMetadata(nil)).To(BeNil())
	})

	It("parses delete-files columns sorted by name", func() {
		Expect(ParseDeleteFilesColumns(map[string]interface{}{
			"resumePDF": map[string]interface{}{"status": "DELETED"},
			"photoID":   "DELETED",
			"other":     map[string]interface{}{"status": "archived"},
			"lower":     "deleted",
		})).To(Equal([]fvcommon.DeleteFilesColumnResult{
			{Column: "lower", Status: fvcommon.DeleteFilesStatusDeleted, RawStatus: "deleted"},
			{Column: "other", Status: fvcommon.DeleteFilesStatusUnknown, RawStatus: "archived"},
			{Column: "photoID", Status: fvcommon.DeleteFilesStatusDeleted, RawStatus: "DELETED"},
			{Column: "resumePDF", Status: fvcommon.DeleteFilesStatusDeleted, RawStatus: "DELETED"},
		}))
		Expect(ParseDeleteFilesColumns(nil)).To(BeNil())
	})

	It("sets the request ID only for failed records", func() {
		header := http.Header{}
		header.Set("x-request-id", "req-1")
		Expect(RequestIdFor(header, "boom")).To(Equal("req-1"))
		Expect(RequestIdFor(header, "")).To(BeEmpty())
	})
})
