package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cdetect "github.com/skyflowapi/skyflow-go/common/detect"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

var _ = Describe("Detect", func() {
	const vaultId = "detect-vault"
	const apiKey = "sky-abcde-1234567890abcdef1234567890abcdef"
	var (
		server *httptest.Server
		path   string
		header http.Header
		body   map[string]interface{}
	)

	BeforeEach(func() {
		path = ""
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path, header = r.URL.Path, r.Header.Clone()
			body = nil
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"processedText":"My name is John Doe."}`))
		}))
	})
	AfterEach(func() { server.Close() })

	newClient := func(opts ...Option) *Skyflow {
		options := append([]Option{
			WithVaults(common.VaultConfig{VaultId: vaultId, ClusterId: "cluster", BaseVaultUrl: server.URL}),
			WithCredentials(common.Credentials{ApiKey: apiKey}),
		}, opts...)
		client, err := NewSkyflow(options...)
		Expect(err).To(BeNil())
		return client
	}

	It("calls the Detect endpoint with client credentials, custom headers and the vault ID", func() {
		client := newClient(WithCustomHeaders(map[common.CustomHeaderKey]string{common.SkyflowAccountId: "acc"}))
		service, err := client.Detect() // first configured vault
		Expect(err).To(BeNil())

		resp, err := service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{
			Text:           "My name is [NAME_1].",
			RedactionLevel: []common.RedactionRule{common.RedactionLevel{EntityName: common.EntityTypeName, RedactionType: common.PLAIN_TEXT}},
		}, common.ReidentifyStringOptions{})
		Expect(err).To(BeNil())
		Expect(resp.ProcessedText).To(Equal("My name is John Doe."))
		Expect(path).To(Equal("/v2/detect/reidentify/string"))
		Expect(header.Get("Authorization")).To(Equal("Bearer " + apiKey))
		Expect(header.Get("x-skyflow-account-id")).To(Equal("acc"))
		Expect(body["vaultId"]).To(Equal(vaultId))
		level := body["redactionLevel"].([]interface{})[0].(map[string]interface{})
		Expect(level).To(Equal(map[string]interface{}{"entityName": "NAME", "redactionType": "PLAIN_TEXT"}))
	})

	It("sends a call's custom headers over the client's and validates them with v2's rules", func() {
		client := newClient(WithCustomHeaders(map[common.CustomHeaderKey]string{
			common.SkyflowAccountId: "acc", common.RequestIdHeader: "client-req",
		}))
		service, err := client.Detect(vaultId)
		Expect(err).To(BeNil())

		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"},
			common.ReidentifyStringOptions{CustomHeaders: map[common.CustomHeaderKey]string{common.RequestIdHeader: "call-req"}})
		Expect(err).To(BeNil())
		Expect(header.Get("x-skyflow-account-id")).To(Equal("acc"))
		Expect(header.Get("x-request-id")).To(Equal("call-req"))

		// An empty map is rejected, as for the v2 vault operations.
		path = ""
		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"},
			common.ReidentifyStringOptions{CustomHeaders: map[common.CustomHeaderKey]string{}})
		Expect(err).NotTo(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring(skyflowError.EMPTY_REQUEST_HEADER))
		Expect(path).To(BeEmpty())

		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"},
			common.ReidentifyStringOptions{CustomHeaders: map[common.CustomHeaderKey]string{"x-other": "v"}})
		Expect(err).NotTo(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring("x-other"))
		Expect(err.GetMessage()).To(ContainSubstring(cdetect.OperationReidentifyString))
	})

	It("uses credentials set after it was obtained", func() {
		client := newClient()
		service, err := client.Detect(vaultId)
		Expect(err).To(BeNil())

		otherKey := "sky-zzzzz-1234567890abcdef1234567890abcdef"
		Expect(client.UpdateSkyflowCredentials(common.Credentials{ApiKey: otherKey})).To(BeNil())
		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"}, common.ReidentifyStringOptions{})
		Expect(err).To(BeNil())
		Expect(header.Get("Authorization")).To(Equal("Bearer " + otherKey))

		vaultKey := "sky-vvvvv-1234567890abcdef1234567890abcdef"
		Expect(client.UpdateVaultConfig(common.VaultConfig{VaultId: vaultId, Credentials: common.Credentials{ApiKey: vaultKey}})).To(BeNil())
		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"}, common.ReidentifyStringOptions{})
		Expect(err).To(BeNil())
		Expect(header.Get("Authorization")).To(Equal("Bearer " + vaultKey))
		Expect(path).To(Equal("/v2/detect/reidentify/string")) // BaseVaultUrl kept by the update
	})

	// The vault config has no HTTP settings: Detect retries with the SDK's defaults (2 retries).
	It("uses the base vault URL of UpdateVaultConfig and the default retries", func() {
		var calls int
		updated := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls == 1 {
				w.WriteHeader(http.StatusServiceUnavailable) // retried by default
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"processedText":"updated"}`))
		}))
		defer updated.Close()
		client := newClient()
		service, err := client.Detect(vaultId)
		Expect(err).To(BeNil())

		Expect(client.UpdateVaultConfig(common.VaultConfig{VaultId: vaultId, BaseVaultUrl: updated.URL})).To(BeNil())
		resp, err := service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"}, common.ReidentifyStringOptions{})
		Expect(err).To(BeNil())
		Expect(resp.ProcessedText).To(Equal("updated"))
		Expect(calls).To(Equal(2))
		Expect(path).To(BeEmpty()) // the original server was not called

		// An update that sets neither a URL nor a cluster ID keeps the URL.
		Expect(client.UpdateVaultConfig(common.VaultConfig{VaultId: vaultId, Env: common.SANDBOX})).To(BeNil())
		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"}, common.ReidentifyStringOptions{})
		Expect(err).To(BeNil())
		Expect(calls).To(Equal(3))
		Expect(path).To(BeEmpty())
	})

	It("merges every field UpdateVaultConfig sets into the Detect config", func() {
		client := newClient()
		service := client.detectServices[vaultId]
		client.updateDetect(service, common.VaultConfig{
			VaultId:      vaultId,
			ClusterId:    "new-cluster",
			BaseVaultURL: "https://legacy.example.com",
			Env:          common.SANDBOX,
		})
		got := service.base.Config()
		Expect(got.ClusterId).To(Equal("new-cluster"))
		Expect(got.BaseVaultUrl).To(BeEmpty()) // replaced together with BaseVaultURL
		Expect(got.BaseVaultURL).To(Equal("https://legacy.example.com"))
		Expect(got.Env).To(Equal(common.SANDBOX))
		// No HTTP settings reach the Detect config, so the SDK's defaults apply.
		Expect(got.Timeout).To(BeZero())
		Expect(got.MaxRetries).To(BeNil())

		// Unset fields keep their values; Env is always applied.
		client.updateDetect(service, common.VaultConfig{VaultId: vaultId})
		kept := service.base.Config()
		Expect(kept.ClusterId).To(Equal("new-cluster"))
		Expect(kept.BaseVaultURL).To(Equal("https://legacy.example.com"))
		Expect(kept.Env).To(Equal(common.PROD))
	})

	It("sends a SkyVault source type, SKYFLOW_ID included", func() {
		service, err := newClient().Detect(vaultId)
		Expect(err).To(BeNil())
		_, _ = service.DeidentifyFile(context.Background(), common.DeidentifyFileRequest{
			DataSource: common.DeidentifySourceTypeSkyflowID, Value: "sid",
		}, common.DeidentifyFileOptions{})
		Expect(path).To(Equal("/v2/detect/deidentify/file"))
		Expect(body["dataSource"]).To(Equal("SKYFLOW_ID"))
		// With neither configuration source, a configuration holding only the vault ID is sent.
		Expect(body["configuration"]).To(Equal(map[string]interface{}{"vaultId": vaultId}))
	})

	It("reports validation errors with the v2 SDK prefix", func() {
		service, err := newClient().Detect(vaultId)
		Expect(err).To(BeNil())
		_, err = service.DeidentifyString(context.Background(), common.DeidentifyStringRequest{Text: " ", ConfigurationId: "c"}, common.DeidentifyStringOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Skyflow Go SDK v2"))
		Expect(err.GetMessage()).To(ContainSubstring("Invalid text in deidentify string request."))
	})

	It("rejects invalid client headers and expired tokens with v2's messages", func() {
		client := newClient(WithCustomHeaders(map[common.CustomHeaderKey]string{"x-unknown": "v"}))
		service, err := client.Detect(vaultId)
		Expect(err).To(BeNil())
		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"}, common.ReidentifyStringOptions{})
		Expect(err).NotTo(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring("Client headers in"))
		Expect(err.GetMessage()).To(ContainSubstring("x-unknown"))
		Expect(path).To(BeEmpty())

		// A JWT whose exp claim is in 2001.
		expired := "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjEwMDAwMDAwMDB9.c2ln"
		service, err = newClient(WithCredentials(common.Credentials{Token: expired})).Detect(vaultId)
		Expect(err).To(BeNil())
		_, err = service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{Text: "x"}, common.ReidentifyStringOptions{})
		Expect(err).NotTo(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring("Token provided is either invalid or has expired"))
	})

	It("validates DeidentifyFile, GetRun, ReidentifyFile and CheckGuardrails before calling the API", func() {
		service, err := newClient().Detect(vaultId)
		Expect(err).To(BeNil())
		ctx := context.Background()

		_, err = service.DeidentifyFile(ctx, common.DeidentifyFileRequest{}, common.DeidentifyFileOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Specify one of FilePath, FileObject or Value in deidentify file request."))
		_, err = service.GetRun(ctx, common.GetRunRequest{}, common.GetRunOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Invalid RunId in get run request."))
		_, err = service.ReidentifyFile(ctx, common.ReidentifyFileRequest{}, common.ReidentifyFileOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Invalid DataSource in reidentify file request."))
		_, err = service.CheckGuardrails(ctx, common.CheckGuardrailsRequest{}, common.CheckGuardrailsOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Invalid text in check guardrails request."))
		Expect(path).To(BeEmpty())
	})

	It("fails for an unknown or removed vault", func() {
		client := newClient()
		_, err := client.Detect("unknown")
		Expect(err).NotTo(BeNil())
		Expect(client.RemoveVaultConfig(vaultId)).To(BeNil())
		_, err = client.Detect(vaultId)
		Expect(err).NotTo(BeNil())
	})
})
