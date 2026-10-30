package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

var _ = Describe("URL, metadata and shared services", func() {
	DescribeTable("builds the vault URL for each environment",
		func(env fvcommon.Env, expected string) {
			Expect(GetURLWithEnv(env, "cluster")).To(Equal(expected))
		},
		Entry("dev", fvcommon.DEV, "https://cluster.skyvault.skyflowapis.dev"),
		Entry("stage", fvcommon.STAGE, "https://cluster.skyvault.skyflowapis.tech"),
		Entry("sandbox", fvcommon.SANDBOX, "https://cluster.skyvault.skyflowapis-preview.com"),
		Entry("prod", fvcommon.PROD, "https://cluster.skyvault.skyflowapis.com"),
	)

	It("describes the SDK and runtime in the metadata header", func() {
		var metadata map[string]string
		Expect(json.Unmarshal([]byte(CreateSDKMetadata()), &metadata)).To(Succeed())
		Expect(metadata["sdk_name_version"]).To(HavePrefix("skyflow-flowvault-go@"))
		Expect(metadata["device_model"]).To(Equal(runtime.GOOS))
		Expect(metadata["os_details"]).To(Equal(runtime.GOOS + " " + runtime.GOARCH))
		Expect(metadata["runtime_details"]).To(Equal(runtime.Version()))
	})

	It("shares one authenticator and one config validator", func() {
		Expect(Auth()).ToNot(BeNil())
		Expect(Auth()).To(BeIdenticalTo(Auth()))
		Expect(Auth().IsExpired("")).To(BeTrue())

		Expect(ConfigValidator().ValidateUpdateVaultConfig(fvcommon.VaultConfig{})).ToNot(BeNil())
		Expect(ConfigValidator().ValidateUpdateVaultConfig(fvcommon.VaultConfig{VaultId: "v"})).To(BeNil())
	})
})

var _ = Describe("HTTP config and retries, edge cases", func() {
	It("falls back to the client's retries and returns zero when no duration is set", func() {
		clientRetries := 4
		resolved := ResolveHttpConfig(fvcommon.HttpConfig{}, fvcommon.HttpConfig{MaxRetries: &clientRetries})
		Expect(resolved.MaxRetries).To(Equal(4))
		Expect(firstDuration()).To(BeZero())
		Expect(firstDuration(0, -time.Second)).To(BeZero())
	})

	It("caps the delay at max and skips jitter when it rounds to zero", func() {
		delay := backoff(1, 10*time.Second, time.Second)
		Expect(delay).To(BeNumerically(">=", 800*time.Millisecond))
		Expect(delay).To(BeNumerically("<=", 1200*time.Millisecond))
		Expect(backoff(3, 0, 0)).To(BeZero())
		Expect(backoff(1, time.Nanosecond, time.Nanosecond)).To(Equal(time.Nanosecond))
	})

	newRetryServer := func(calls *int32) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(calls, 1)
			_, _ = io.Copy(io.Discard, r.Body)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
	}

	It("doesn't retry a body it can't send again", func() {
		var calls int32
		server := newRetryServer(&calls)
		defer server.Close()
		transport := &retryTransport{base: http.DefaultTransport, config: ResolvedHttpConfig{MaxRetries: 3}, sleep: func(time.Duration) {}}

		// A body that isn't a bytes/strings reader gets no GetBody.
		req, _ := http.NewRequest(http.MethodPost, server.URL, io.NopCloser(strings.NewReader("payload")))
		Expect(req.GetBody).To(BeNil())
		resp, err := transport.RoundTrip(req)
		Expect(err).To(BeNil())
		_ = resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
		Expect(atomic.LoadInt32(&calls)).To(Equal(int32(1)))

		// GetBody that fails stops the retries too.
		req, _ = http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
		req.GetBody = func() (io.ReadCloser, error) { return nil, errors.New("no body") }
		resp, err = transport.RoundTrip(req)
		Expect(err).To(BeNil())
		_ = resp.Body.Close()
		Expect(atomic.LoadInt32(&calls)).To(Equal(int32(2)))
	})

	It("waits with a real timer and stops waiting when the request is cancelled", func() {
		var calls int32
		server := newRetryServer(&calls)
		defer server.Close()

		transport := &retryTransport{base: http.DefaultTransport, config: ResolvedHttpConfig{MaxRetries: 1, InitialRetryDelay: time.Millisecond, MaxRetryDelay: time.Millisecond}}
		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		resp, err := transport.RoundTrip(req)
		Expect(err).To(BeNil())
		_ = resp.Body.Close()
		Expect(atomic.LoadInt32(&calls)).To(Equal(int32(2)))

		ctx, cancel := context.WithCancel(context.Background())
		transport.config = ResolvedHttpConfig{MaxRetries: 1, InitialRetryDelay: time.Hour, MaxRetryDelay: time.Hour}
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		done := make(chan error, 1)
		go func() {
			_, err := transport.RoundTrip(req)
			done <- err
		}()
		Eventually(func() int32 { return atomic.LoadInt32(&calls) }).Should(Equal(int32(3)))
		cancel()
		Eventually(done).Should(Receive(MatchError(context.Canceled)))
	})
})

var _ = Describe("Files, edge cases", func() {
	It("fails a PUT when the path disappeared or the URL is invalid", func() {
		status, msg := PutFile(context.Background(), http.DefaultClient, "http://127.0.0.1", ResolvedFile{Path: filepath.Join(GinkgoT().TempDir(), "gone.txt")})
		Expect(status).To(Equal(fvcommon.UploadStatusFailed))
		Expect(msg).To(HavePrefix("PUT failed: open "))

		status, msg = PutFile(context.Background(), http.DefaultClient, "://bad-url?signature=secret", ResolvedFile{Content: []byte("a")})
		Expect(status).To(Equal(fvcommon.UploadStatusFailed))
		Expect(msg).To(Equal("PUT failed: missing protocol scheme"))
		Expect(msg).NotTo(ContainSubstring("signature"))
	})

	It("streams a pipe with an unknown length", func() {
		var length int64
		var body string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			length, body = r.ContentLength, string(raw)
			w.WriteHeader(http.StatusCreated)
		}))
		defer server.Close()

		reader, writer, err := os.Pipe()
		Expect(err).To(BeNil())
		defer reader.Close()
		Expect(remaining(reader)).To(Equal(int64(-1)))
		go func() {
			_, _ = writer.WriteString("piped")
			_ = writer.Close()
		}()
		status, msg := PutFile(context.Background(), http.DefaultClient, server.URL, ResolvedFile{Object: reader})
		Expect(msg).To(BeEmpty())
		Expect(status).To(Equal(fvcommon.UploadStatusUploaded))
		Expect(length).To(Equal(int64(-1)))
		Expect(body).To(Equal("piped"))
	})

	It("returns errors that carry no URL as they are", func() {
		plain := errors.New("plain")
		Expect(withoutURL(plain)).To(BeIdenticalTo(plain))
	})
})

var _ = Describe("Parsing, edge cases", func() {
	It("skips nil entries and stringifies non-string values", func() {
		tokens := ParseTokens(map[string]interface{}{
			"a": []interface{}{nil, 42},
			"b": nil,
			"c": map[string]interface{}{"token": 7, "path": nil},
		})
		Expect(tokens).To(Equal(map[string][]fvcommon.Token{
			"a": {{Token: "42"}},
			"c": {{Token: "7"}},
		}))

		Expect(ParseHashedData(nil)).To(BeNil())
		hashes := ParseHashedData(map[string]interface{}{
			"a": []interface{}{"h1", nil, map[string]interface{}{"data": 3.5}},
			"b": nil,
		})
		Expect(hashes).To(Equal(map[string][]fvcommon.HashedValue{
			"a": {{Data: "h1"}, {Data: "3.5"}},
		}))
	})

	It("leaves the status empty for an unexpected delete-files value", func() {
		Expect(ParseDeleteFilesColumns(map[string]interface{}{"c": 1})).
			To(Equal([]fvcommon.DeleteFilesColumnResult{{Column: "c"}}))
		Expect(RequestIdFor(nil, "boom")).To(BeEmpty())
	})
})
