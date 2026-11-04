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

	It("routes every Detect operation to its V2 endpoint", func() {
		var paths []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			paths = append(paths, r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/v2/detect/deidentify/file":
				_, _ = w.Write([]byte(`{"runId":"run-1"}`))
			case "/v2/detect/runs/run-1":
				_, _ = w.Write([]byte(`{"status":"SUCCESS","outputType":"BASE64"}`))
			default:
				_, _ = w.Write([]byte(`{"processedText":"done","outputType":"BASE64"}`))
			}
		}))
		defer server.Close()
		client := newTestClient()
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, BaseVaultUrl: server.URL})).To(BeNil())
		service, err := client.Detect(testVaultID)
		Expect(err).To(BeNil())
		ctx := context.Background()
		levels := []fvcommon.RedactionLevel{{TokenGroupName: "group", RedactionPattern: "pattern"}}

		run, err := service.DeidentifyFile(ctx, fvcommon.DeidentifyFileRequest{
			DataSource: fvcommon.DeidentifySourceTypeBase64, Value: "aGk=", DataFormat: fvcommon.FileDataFormatTXT, ConfigurationId: "cfg",
		}, fvcommon.DeidentifyFileOptions{})
		Expect(err).To(BeNil())
		Expect(run.RunId).To(Equal("run-1"))

		got, err := service.GetRun(ctx, fvcommon.GetRunRequest{RunId: "run-1"}, fvcommon.GetRunOptions{})
		Expect(err).To(BeNil())
		Expect(got.Status).To(Equal(fvcommon.DetectRunStatusSuccess))

		text, err := service.ReidentifyString(ctx, fvcommon.ReidentifyStringRequest{Text: "[T]", RedactionLevel: levels}, fvcommon.ReidentifyStringOptions{})
		Expect(err).To(BeNil())
		Expect(text.ProcessedText).To(Equal("done"))

		_, err = service.ReidentifyFile(ctx, fvcommon.ReidentifyFileRequest{
			DataSource: fvcommon.ReidentifySourceTypeBase64, Value: "aGk=", DataFormat: fvcommon.FileDataFormatTXT, RedactionLevel: levels,
		}, fvcommon.ReidentifyFileOptions{})
		Expect(err).To(BeNil())
		Expect(paths).To(HaveLen(4))
	})

	It("does not accept a Skyflow ID as a file source", func() {
		service, err := newTestClient().Detect(testVaultID)
		Expect(err).To(BeNil())
		_, err = service.DeidentifyFile(context.Background(), fvcommon.DeidentifyFileRequest{
			DataSource: "SKYFLOW_ID", Value: "id", ConfigurationId: "cfg",
		}, fvcommon.DeidentifyFileOptions{})
		Expect(err.GetMessage()).To(ContainSubstring(`DataSource "SKYFLOW_ID" is not supported. Specify BASE64 or PRESIGNED_URL.`))

		_, err = service.DeidentifyFile(context.Background(), fvcommon.DeidentifyFileRequest{Value: "id", ConfigurationId: "cfg"}, fvcommon.DeidentifyFileOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Specify BASE64 or PRESIGNED_URL."))
		Expect(err.GetMessage()).NotTo(ContainSubstring("SKYFLOW_ID"))
	})

	It("reports validation errors with the FlowVault SDK prefix", func() {
		service, err := newTestClient().Detect(testVaultID)
		Expect(err).To(BeNil())
		_, err = service.DeidentifyString(context.Background(), fvcommon.DeidentifyStringRequest{Text: " "}, fvcommon.DeidentifyStringOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Skyflow FlowVault Go SDK"))
		Expect(err.GetMessage()).To(ContainSubstring("Text must not be empty"))
	})

	It("is removed with its vault", func() {
		client := newTestClient()
		Expect(client.RemoveVaultConfig(testVaultID)).To(BeNil())
		_, err := client.Detect(testVaultID)
		Expect(err).NotTo(BeNil())
	})
})
