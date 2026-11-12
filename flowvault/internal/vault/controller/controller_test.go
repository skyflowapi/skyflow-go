package controller

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/skyflowapi/skyflow-go/flowvault/internal/testutil"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Controller Suite")
}

var (
	testApiKey  = testutil.Resource("client_api_key.txt")
	vaultApiKey = testutil.Resource("vault_api_key.txt")
)

var _ = Describe("SetupRequest", func() {
	It("uses the client API key and the interceptor's headers", func() {
		config := fvcommon.VaultConfig{VaultId: "v", ClusterId: "cluster", Env: fvcommon.DEV, Timeout: 5 * time.Second}
		v := NewVaultController(config, ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}})

		var operation string
		var batchIndex, totalBatches int
		setup, err := v.SetupRequest(context.Background(), "QUERY", func(rc *fvcommon.RequestContext) {
			operation = rc.Operation
			batchIndex, totalBatches = rc.BatchIndex, rc.TotalBatches
			rc.AddHeader(fvcommon.RequestIdHeader, "custom-id")
			rc.AddHeader(fvcommon.SkyflowAccountId, "acc")
		})
		Expect(err).To(BeNil())
		Expect(operation).To(Equal("QUERY"))
		// A single, unbatched call, as the Java and Python SDKs report it.
		Expect([]int{batchIndex, totalBatches}).To(Equal([]int{fvcommon.NotBatched, fvcommon.NotBatched}))
		Expect(fvcommon.NotBatched).To(Equal(-1))
		Expect(setup.AuthToken).To(Equal(testApiKey))
		Expect(setup.BaseURL).To(Equal("https://cluster.skyvault.skyflowapis.dev"))
		Expect(setup.Headers.Get("x-skyflow-account-id")).To(Equal("acc"))
		Expect(setup.Headers.Get("x-request-id")).To(Equal("custom-id"))
		Expect(setup.Headers.Get("sky-metadata")).To(ContainSubstring(`"sdk_name_version":"skyflow-flowvault-go@`))
		Expect(setup.HTTPClient.Timeout).To(Equal(5 * time.Second))
	})

	It("prefers vault credentials and BaseVaultUrl", func() {
		config := fvcommon.VaultConfig{
			VaultId: "v", BaseVaultUrl: "https://example.com",
			Credentials: fvcommon.Credentials{ApiKey: vaultApiKey},
		}
		setup, err := NewVaultController(config, ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}}).SetupRequest(context.Background(), "GET", nil)
		Expect(err).To(BeNil())
		Expect(setup.AuthToken).To(Equal(vaultApiKey))
		Expect(setup.BaseURL).To(Equal("https://example.com"))
	})

	It("uses PROD when Env isn't set", func() {
		setup, err := NewVaultController(fvcommon.VaultConfig{VaultId: "v", ClusterId: "cluster"}, ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}}).SetupRequest(context.Background(), "GET", nil)
		Expect(err).To(BeNil())
		Expect(setup.BaseURL).To(Equal("https://cluster.skyvault.skyflowapis.com"))
	})

	It("fails when no credentials are available", func() {
		Expect(os.Unsetenv("SKYFLOW_CREDENTIALS")).To(Succeed())
		_, err := NewVaultController(fvcommon.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{}).SetupRequest(context.Background(), "GET", nil)
		Expect(err.GetMessage()).To(ContainSubstring("Invalid credentials"))
	})

	It("rejects an expired bearer token", func() {
		config := fvcommon.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: fvcommon.Credentials{Token: "not-a-jwt"}}
		_, err := NewVaultController(config, ClientSettings{}).SetupRequest(context.Background(), "GET", nil)
		Expect(err.GetMessage()).To(ContainSubstring("expired"))
	})
})
