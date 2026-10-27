package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

var _ = Describe("Detect service", func() {
	It("returns an error for an unknown vault", func() {
		_, err := newTestClient().Detect("unknown")
		Expect(err.GetMessage()).To(ContainSubstring("VaultId is missing from the config"))
	})

	It("calls the vault URL with the vault's credentials and follows config updates", func() {
		var path, auth string
		var body map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path, auth = r.URL.Path, r.Header.Get("Authorization")
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"text":"hi","validation":"PASSED"}`))
		}))
		defer server.Close()

		client := newTestClient()
		service, err := client.Detect(testVaultID)
		Expect(err).To(BeNil())
		// The service obtained before the update sees the new URL on its next call.
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, BaseVaultUrl: server.URL})).To(BeNil())

		resp, err := service.CheckGuardrails(context.Background(), fvcommon.CheckGuardrailsRequest{Text: "hi"}, fvcommon.CheckGuardrailsOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Validation).To(Equal(fvcommon.GuardrailsValidationPassed))
		Expect(path).To(Equal("/v2/detect/guardrails"))
		Expect(auth).To(Equal("Bearer " + testApiKey))
		Expect(body["vaultId"]).To(Equal(testVaultID))
	})

	It("reports validation errors with the FlowVault SDK prefix", func() {
		service, err := newTestClient().Detect(testVaultID)
		Expect(err).To(BeNil())
		_, err = service.DeidentifyString(context.Background(), fvcommon.DeidentifyStringRequest{Text: "x"}, fvcommon.DeidentifyStringOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Skyflow FlowVault Go SDK"))
		Expect(err.GetMessage()).To(ContainSubstring("Set exactly one of ConfigurationId or Configuration"))
	})

	It("is removed with its vault", func() {
		client := newTestClient()
		Expect(client.RemoveVaultConfig(testVaultID)).To(BeNil())
		_, err := client.Detect(testVaultID)
		Expect(err).NotTo(BeNil())
	})
})
