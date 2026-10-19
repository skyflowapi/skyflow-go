package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/skyflowapi/skyflow-go/flowvault/internal/testutil"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Client Suite")
}

const testVaultID = "vault123"

var testApiKey = testutil.Resource("client_api_key.txt")

func newTestClient(opts ...Option) *Skyflow {
	options := append([]Option{
		WithVaults(fvcommon.VaultConfig{VaultId: testVaultID, ClusterId: "cluster", Env: fvcommon.DEV}),
		WithCredentials(fvcommon.Credentials{ApiKey: testApiKey}),
	}, opts...)
	client, err := NewSkyflow(options...)
	Expect(err).To(BeNil())
	return client
}

var _ = Describe("Skyflow client", func() {
	It("rejects an empty vault list and duplicate vault IDs", func() {
		_, err := NewSkyflow(WithVaults())
		Expect(err.GetMessage()).To(ContainSubstring("No vault configurations available"))

		config := fvcommon.VaultConfig{VaultId: testVaultID, ClusterId: "cluster"}
		_, err = NewSkyflow(WithVaults(config, config))
		Expect(err.GetMessage()).To(ContainSubstring("already exists"))
	})

	It("rejects negative HTTP settings on the client and on a vault", func() {
		_, err := NewSkyflow(WithHttpConfig(fvcommon.HttpConfig{Timeout: -time.Second}))
		Expect(err.GetMessage()).To(ContainSubstring("must not be negative"))

		retries := -1
		_, err = NewSkyflow(WithVaults(fvcommon.VaultConfig{VaultId: testVaultID, ClusterId: "cluster", MaxRetries: &retries}))
		Expect(err.GetMessage()).To(ContainSubstring("must not be negative"))
	})

	It("passes client settings to vaults registered before or after them", func() {
		retries := 2
		client := newTestClient(
			WithHttpConfig(fvcommon.HttpConfig{MaxRetries: &retries}),
		)
		vault, err := client.Vault(testVaultID)
		Expect(err).To(BeNil())
		settings := vault.controller.ClientSettings()
		Expect(settings.Credentials.ApiKey).To(Equal(testApiKey))
		Expect(*settings.HttpConfig.MaxRetries).To(Equal(2))

		Expect(client.AddVaultConfig(fvcommon.VaultConfig{VaultId: "other", ClusterId: "cluster"})).To(BeNil())
		other, _ := client.Vault("other")
		Expect(*other.controller.ClientSettings().HttpConfig.MaxRetries).To(Equal(2))
	})

	It("keeps the vault's Env when an update doesn't set one", func() {
		client := newTestClient()
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, ClusterId: "other"})).To(BeNil())
		config, _ := client.GetVaultConfig(testVaultID)
		Expect(config.Env).To(Equal(fvcommon.DEV))
		Expect(config.ClusterId).To(Equal("other"))

		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, Env: fvcommon.STAGE})).To(BeNil())
		config, _ = client.GetVaultConfig(testVaultID)
		Expect(config.Env).To(Equal(fvcommon.STAGE))
	})

	It("returns copies, so changing them doesn't change the client", func() {
		client := newTestClient()
		config, _ := client.GetVaultConfig(testVaultID)
		config.ClusterId = "changed"
		again, _ := client.GetVaultConfig(testVaultID)
		Expect(again.ClusterId).To(Equal("cluster"))

		creds := client.GetSkyflowCredentials()
		creds.ApiKey = "changed"
		Expect(client.GetSkyflowCredentials().ApiKey).To(Equal(testApiKey))
	})

	It("can be changed while calls are running", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()
		client := newTestClient(WithVaults(fvcommon.VaultConfig{VaultId: "local", BaseVaultUrl: server.URL}))
		vault, err := client.Vault("local")
		Expect(err).To(BeNil())

		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()
				for j := 0; j < 20; j++ {
					_, err := vault.Query(context.Background(), fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
					Expect(err).To(BeNil())
				}
			}()
			go func(i int) {
				defer wg.Done()
				defer GinkgoRecover()
				id := fmt.Sprintf("extra-%d", i)
				for j := 0; j < 20; j++ {
					Expect(client.AddVaultConfig(fvcommon.VaultConfig{VaultId: id, ClusterId: "c"})).To(BeNil())
					Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: "local", Timeout: time.Duration(j+1) * time.Second})).To(BeNil())
					Expect(client.UpdateSkyflowCredentials(fvcommon.Credentials{ApiKey: testApiKey})).To(BeNil())
					_, _ = client.GetVaultConfig("local")
					Expect(client.RemoveVaultConfig(id)).To(BeNil())
				}
			}(i)
		}
		wg.Wait()
	})

	It("updates a vault's HTTP settings in place", func() {
		client := newTestClient()
		retries := 3
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, Timeout: 5 * time.Second, MaxRetries: &retries})).To(BeNil())
		config, _ := client.GetVaultConfig(testVaultID)
		Expect(config.Timeout).To(Equal(5 * time.Second))
		Expect(*config.MaxRetries).To(Equal(3))
	})
})

var _ = Describe("VaultService", func() {
	var (
		vault *VaultService
		ctx   = context.Background()
	)
	BeforeEach(func() {
		var err error
		vault, err = newTestClient().Vault(testVaultID)
		Expect(err).To(BeNil())
	})

	It("validates before anything else", func() {
		_, err := vault.Insert(ctx, fvcommon.InsertRequest{}, fvcommon.InsertOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Records must not be empty"))
		_, err = vault.Query(ctx, fvcommon.QueryRequest{}, fvcommon.QueryOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Query must not be empty"))
	})

	It("rejects interceptor headers that aren't supported", func() {
		_, err := vault.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { rc.AddHeader("x-other", "v") },
		})
		Expect(err.GetMessage()).To(ContainSubstring("is not supported"))
	})

	It("sends valid requests to the vault", func() {
		var paths []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			paths = append(paths, r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()
		client := newTestClient(WithVaults(fvcommon.VaultConfig{VaultId: "local", BaseVaultUrl: server.URL}))
		vault, err := client.Vault("local")
		Expect(err).To(BeNil())

		data := map[string]interface{}{"a": 1}
		_, err = vault.Insert(ctx, fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data}}}, fvcommon.InsertOptions{})
		Expect(err).To(BeNil())
		_, err = vault.Update(ctx, fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, fvcommon.UpdateOptions{})
		Expect(err).To(BeNil())
		_, err = vault.Get(ctx, fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}}, fvcommon.GetOptions{})
		Expect(err).To(BeNil())
		_, err = vault.Delete(ctx, fvcommon.DeleteRequest{TableName: "t", SkyflowIds: []string{"id"}}, fvcommon.DeleteOptions{})
		Expect(err).To(BeNil())
		_, err = vault.Detokenize(ctx, fvcommon.DetokenizeRequest{Tokens: []string{"tok"}}, fvcommon.DetokenizeOptions{})
		Expect(err).To(BeNil())
		_, err = vault.GetTokens(ctx, fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "v", TokenGroupName: "g"}}}, fvcommon.GetTokensOptions{})
		Expect(err).To(BeNil())
		_, err = vault.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).To(BeNil())
		_, err = vault.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "YQ==", FileName: "a.txt"}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(err).To(BeNil())
		_, err = vault.DeleteFiles(ctx, fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{
			{TableName: "t", Columns: []string{"c"}, SkyflowId: "id"},
		}}, fvcommon.DeleteFilesOptions{})
		Expect(err).To(BeNil())

		Expect(paths).To(Equal([]string{
			"/v2/records/insert", "/v2/records/update", "/v2/records/get", "/v2/records/delete",
			"/v2/tokens/detokenize", "/v2/records/getTokens", "/v2/query", "/v2/files/upload", "/v2/files/delete",
		}))
	})
})
