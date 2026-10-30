package helpers

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

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
			MaxRetries:        2,
			InitialRetryDelay: 500 * time.Millisecond,
			MaxRetryDelay:     time.Second,
		}))

		defaults := ResolveHttpConfig(fvcommon.HttpConfig{}, fvcommon.HttpConfig{})
		Expect(defaults.Timeout).To(Equal(60 * time.Second))
		Expect(defaults.MaxRetries).To(Equal(0))
		Expect(defaults.MaxRetryDelay).To(Equal(2000 * time.Millisecond))
	})

	It("keeps backoff within the jittered bounds", func() {
		for attempt := 1; attempt <= 5; attempt++ {
			delay := backoff(attempt, 500*time.Millisecond, 2*time.Second)
			Expect(delay).To(BeNumerically("<=", 2400*time.Millisecond))
			Expect(delay).To(BeNumerically(">=", 400*time.Millisecond))
		}
	})

	It("retries retryable statuses up to MaxRetries and resends the body", func() {
		var calls int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			Expect(string(body)).To(Equal("payload"))
			if atomic.AddInt32(&calls, 1) < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		transport := &retryTransport{
			base:   http.DefaultTransport,
			config: ResolvedHttpConfig{MaxRetries: 3, InitialRetryDelay: time.Millisecond, MaxRetryDelay: time.Millisecond},
			sleep:  func(time.Duration) {},
		}
		req, _ := http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
		resp, err := transport.RoundTrip(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(atomic.LoadInt32(&calls)).To(Equal(int32(3)))
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
	It("resolves files from a path, base64 and an *os.File", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "resume.pdf")
		Expect(os.WriteFile(path, []byte("pdf"), 0o600)).To(Succeed())

		file, err := ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: path})
		Expect(err).To(BeNil())
		Expect(file).To(Equal(ResolvedFile{Name: "resume.pdf", ContentType: "application/pdf", Path: path}))

		file, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: base64.StdEncoding.EncodeToString([]byte("txt")), FileName: "notes.txt"})
		Expect(err).To(BeNil())
		Expect(file.Content).To(Equal([]byte("txt")))
		Expect(file.ContentType).To(HavePrefix("text/plain"))

		handle, _ := os.Open(path)
		defer handle.Close()
		file, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FileObject: handle})
		Expect(err).To(BeNil())
		Expect(file.Name).To(Equal("resume.pdf"))
		Expect(file.Object).To(BeIdenticalTo(handle))
		Expect(file.Content).To(BeNil())
	})

	It("rejects an unreadable path, a directory, a closed file and invalid base64", func() {
		_, err := ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: "/does/not/exist"})
		Expect(err.GetMessage()).To(ContainSubstring("Could not read file"))
		_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: GinkgoT().TempDir()})
		Expect(err.GetMessage()).To(ContainSubstring("not a regular file"))
		closed, _ := os.CreateTemp(GinkgoT().TempDir(), "closed")
		_ = closed.Close()
		_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", FileObject: closed})
		Expect(err.GetMessage()).To(ContainSubstring("Could not read file"))
		_, err = ResolveFile(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "%%%", FileName: "a.txt"})
		Expect(err.GetMessage()).To(ContainSubstring("invalid base64"))
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
		file := ResolvedFile{Name: "a.pdf", ContentType: "application/pdf", Content: []byte("pdf")}

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

		handle, _ := os.Open(path)
		defer handle.Close()
		_, _ = handle.Seek(7, io.SeekStart)
		status, _ = PutFile(context.Background(), http.DefaultClient, server.URL, ResolvedFile{Object: handle})
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(length).To(Equal(int64(7)))
		Expect(body).To(Equal("content"))
		_, err := handle.Stat()
		Expect(err).To(BeNil())

		empty := filepath.Join(GinkgoT().TempDir(), "empty.txt")
		Expect(os.WriteFile(empty, nil, 0o600)).To(Succeed())
		status, _ = PutFile(context.Background(), http.DefaultClient, server.URL, ResolvedFile{Path: empty})
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(length).To(Equal(int64(0)))
	})

	It("uploads with no retries and no overall timeout, but bounded connect and response waits", func() {
		client := UploadHTTPClient()
		Expect(client).To(BeIdenticalTo(UploadHTTPClient()))
		Expect(client.Timeout).To(BeZero())
		transport, ok := client.Transport.(*http.Transport)
		Expect(ok).To(BeTrue())
		Expect(transport.TLSHandshakeTimeout).To(Equal(10 * time.Second))
		Expect(transport.ResponseHeaderTimeout).To(Equal(10 * time.Second))
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

		hashes := ParseHashedData(map[string]interface{}{"email": map[string]interface{}{"data": "abc", "hashName": "h1"}})
		Expect(hashes["email"]).To(Equal([]fvcommon.HashedValue{{Data: "abc", HashName: "h1"}}))
	})

	It("accepts table or tableName in detokenize metadata", func() {
		Expect(ParseDetokenizeMetadata(map[string]interface{}{"skyflowID": "id", "table": "t1"})).
			To(Equal(&fvcommon.DetokenizeRecordMetadata{SkyflowId: "id", TableName: "t1"}))
		Expect(ParseDetokenizeMetadata(map[string]interface{}{"skyflowId": "id", "tableName": "t2"})).
			To(Equal(&fvcommon.DetokenizeRecordMetadata{SkyflowId: "id", TableName: "t2"}))
		Expect(ParseDetokenizeMetadata(nil)).To(BeNil())
	})

	It("parses delete-files columns sorted by name", func() {
		Expect(ParseDeleteFilesColumns(map[string]interface{}{
			"resumePDF": map[string]interface{}{"status": "DELETED"},
			"photoID":   "DELETED",
		})).To(Equal([]fvcommon.DeleteFilesColumnResult{{Column: "photoID", Status: "DELETED"}, {Column: "resumePDF", Status: "DELETED"}}))
		Expect(ParseDeleteFilesColumns(nil)).To(BeNil())
	})

	It("sets the request ID only for failed records", func() {
		header := http.Header{}
		header.Set("x-request-id", "req-1")
		Expect(RequestIdFor(header, "boom")).To(Equal("req-1"))
		Expect(RequestIdFor(header, "")).To(BeEmpty())
	})
})
