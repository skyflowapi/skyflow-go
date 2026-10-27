package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/skyflowapi/skyflow-go/v2/utils/common"
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

	It("calls the V2 endpoint with client credentials, custom headers and the vault ID", func() {
		client := newClient(WithCustomHeaders(map[common.CustomHeaderKey]string{common.SkyflowAccountId: "acc"}))
		service, err := client.Detect() // first configured vault
		Expect(err).To(BeNil())

		resp, err := service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{
			Text:           "My name is [NAME_1].",
			RedactionLevel: []common.RedactionLevel{{EntityName: common.EntityTypeName, RedactionType: common.PLAIN_TEXT}},
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

	It("reports validation errors with the v2 SDK prefix", func() {
		service, err := newClient().Detect(vaultId)
		Expect(err).To(BeNil())
		_, err = service.DeidentifyString(context.Background(), common.DeidentifyStringRequest{Text: " ", ConfigurationId: "c"}, common.DeidentifyStringOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Skyflow Go SDK v2"))
		Expect(err.GetMessage()).To(ContainSubstring("Text must not be empty"))
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

	It("fails for an unknown or removed vault", func() {
		client := newClient()
		_, err := client.Detect("unknown")
		Expect(err).NotTo(BeNil())
		Expect(client.RemoveVaultConfig(vaultId)).To(BeNil())
		_, err = client.Detect(vaultId)
		Expect(err).NotTo(BeNil())
	})
})
