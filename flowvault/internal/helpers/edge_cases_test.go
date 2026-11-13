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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
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

	It("describes the SDK and runtime in the metadata header, with the shared Skyflow SDK keys", func() {
		var metadata map[string]string
		Expect(json.Unmarshal([]byte(CreateSDKMetadata()), &metadata)).To(Succeed())
		Expect(metadata).To(HaveLen(4))
		Expect(metadata["sdk_name_version"]).To(HavePrefix("skyflow-flowvault-go@"))
		Expect(metadata["sdk_client_device_model"]).To(Equal(runtime.GOOS))
		Expect(metadata["sdk_client_os_details"]).To(Equal(runtime.GOOS + " " + runtime.GOARCH))
		Expect(metadata["sdk_runtime_details"]).To(Equal(runtime.Version()))
	})

	It("shares one authenticator and one config validator", func() {
		Expect(Auth()).ToNot(BeNil())
		Expect(Auth()).To(BeIdenticalTo(Auth()))
		Expect(Profile().Auth).To(BeIdenticalTo(Auth())) // the vault plumbing reuses the authenticator
		Expect(Auth().IsExpired("")).To(BeTrue())

		// The messages are raw (no "Message: " prefix), like every flowvault error.
		Expect(ConfigValidator().ValidateUpdateVaultConfig(fvcommon.VaultConfig{}).GetMessage()).
			To(Equal(fvconsts.SDK_PREFIX + " Initialization failed. Invalid vault ID. Vault ID must not be empty."))
		Expect(ConfigValidator().ValidateUpdateVaultConfig(fvcommon.VaultConfig{VaultId: "v"})).To(BeNil())
		Expect(ConfigValidator().ValidateVaultConfig(fvcommon.VaultConfig{VaultId: "v"}).GetMessage()).
			To(Equal(fvconsts.SDK_PREFIX + " Initialization failed. Specify either 'clusterId' or 'vaultURL'."))
		// Plain http is rejected for hosts other than localhost.
		for _, url := range []string{"http://vault.example.com", "ftp://vault.example.com"} {
			err := ConfigValidator().ValidateUpdateVaultConfig(fvcommon.VaultConfig{VaultId: "v", VaultUrl: url})
			Expect(err.GetMessage()).To(Equal(fvconsts.SDK_PREFIX + " Initialization failed. Vault URL must start with 'https://'."))
		}
		_, err := Auth().GenerateBearerToken("", fvcommon.BearerTokenOptions{})
		Expect(err.GetMessage()).To(HavePrefix(fvconsts.SDK_PREFIX))
	})

	It("names the SDK as Skyflow FlowVault Go SDK with the version without its v", func() {
		Expect(fvconsts.SDK_PREFIX).To(Equal("Skyflow FlowVault Go SDK 1.0.0"))
		Expect(fvconsts.SDK_LOG_PREFIX).To(Equal("[Skyflow FlowVault Go SDK 1.0.0] "))
		var metadata map[string]string
		Expect(json.Unmarshal([]byte(CreateSDKMetadata()), &metadata)).To(Succeed())
		Expect(metadata["sdk_name_version"]).To(Equal("skyflow-flowvault-go@1.0.0"))
	})
})

// The retry transport itself is tested in common/vaultclient.
var _ = Describe("HTTP config, edge cases", func() {
	It("falls back to the client's retries", func() {
		clientRetries := 4
		resolved := ResolveHttpConfig(fvcommon.HttpConfig{}, fvcommon.HttpConfig{MaxRetries: &clientRetries})
		Expect(resolved.MaxRetries).To(Equal(4))
	})
})

var _ = Describe("Files, edge cases", func() {
	It("fails a PUT when the path disappeared or the URL is invalid", func() {
		status, msg := PutFile(context.Background(), http.DefaultClient, "http://127.0.0.1", ResolvedFile{Path: filepath.Join(GinkgoT().TempDir(), "gone.txt")})
		Expect(status).To(Equal(fvcommon.UploadStatusFailed))
		Expect(msg).To(HavePrefix("PUT failed: open "))

		status, msg = PutFile(context.Background(), http.DefaultClient, "://bad-url?signature=secret", ResolvedFile{Base64: "YQ==", DecodedSize: 1})
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
		Expect(fileSize(reader)).To(Equal(int64(-1)))
		// A seekable file that isn't a regular file has no known length either.
		devNull, err := os.Open(os.DevNull)
		Expect(err).To(BeNil())
		defer devNull.Close()
		Expect(fileSize(devNull)).To(Equal(int64(-1)))
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

	})

	It("keeps a column whose token list is empty", func() {
		tokens := ParseTokens(map[string]interface{}{"email": []interface{}{}, "ssn": []interface{}{nil}, "gone": nil})
		Expect(tokens).To(HaveLen(2))
		Expect(tokens).To(HaveKeyWithValue("email", []fvcommon.Token{}))
		Expect(tokens["email"]).ToNot(BeNil())
		Expect(tokens).To(HaveKeyWithValue("ssn", []fvcommon.Token{}))
		Expect(tokens).ToNot(HaveKey("gone"))
	})

	It("reports an unexpected delete-files value as UNKNOWN with no raw status", func() {
		Expect(ParseDeleteFilesColumns(map[string]interface{}{"c": 1})).
			To(Equal([]fvcommon.DeleteFilesColumnResult{{Column: "c", Status: fvcommon.DeleteFilesStatusUnknown}}))
		Expect(RequestIdFor(nil, "boom")).To(BeEmpty())
	})
})
