package controller

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/skyflowapi/skyflow-go/flowvault/internal/generated"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/core"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

var _ = Describe("VaultController error paths", func() {
	var (
		fake *fakeFlowService
		v    *VaultController
		ctx  = context.Background()
	)
	BeforeEach(func() {
		fake = newFakeFlowService()
		DeferCleanup(fake.server.Close)
		v = NewVaultController(fvcommon.VaultConfig{VaultId: "vault1", BaseVaultUrl: fake.server.URL}, ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}})
	})

	data := map[string]interface{}{"a": 1}
	validUpload := fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
		{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "YQ==", FileName: "a.txt"}}},
	}}
	validDeleteFiles := fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{{TableName: "t", Columns: []string{"c"}, SkyflowId: "id"}}}

	// calls runs every operation with a valid request and returns its error.
	calls := map[string]func(*VaultController) *fverrors.SkyflowError{
		"/v2/records/insert": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.Insert(ctx, fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data}}}, fvcommon.InsertOptions{})
			return err
		},
		"/v2/records/update": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.Update(ctx, fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, fvcommon.UpdateOptions{})
			return err
		},
		"/v2/records/get": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.Get(ctx, fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}}, fvcommon.GetOptions{})
			return err
		},
		"/v2/records/delete": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.Delete(ctx, fvcommon.DeleteRequest{TableName: "t", SkyflowIds: []string{"id"}}, fvcommon.DeleteOptions{})
			return err
		},
		"/v2/tokens/detokenize": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.Detokenize(ctx, fvcommon.DetokenizeRequest{Tokens: []string{"tok"}}, fvcommon.DetokenizeOptions{})
			return err
		},
		"/v2/records/getTokens": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.GetTokens(ctx, fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "v", TokenGroupName: "g"}}}, fvcommon.GetTokensOptions{})
			return err
		},
		"/v2/query": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
			return err
		},
		"/v2/files/upload": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.UploadFiles(ctx, validUpload, fvcommon.UploadFilesOptions{})
			return err
		},
		"/v2/files/delete": func(v *VaultController) *fverrors.SkyflowError {
			_, err := v.DeleteFiles(ctx, validDeleteFiles, fvcommon.DeleteFilesOptions{})
			return err
		},
	}

	It("returns a whole-call API error from every operation", func() {
		for path, call := range calls {
			fake.respond(path, http.StatusBadRequest, `{"error":{"http_code":400,"message":"rejected `+path+`"}}`)
			err := call(v)
			Expect(err).ToNot(BeNil(), path)
			Expect(err.GetMessage()).To(ContainSubstring("rejected "+path), path)
			Expect(err.GetRequestId()).To(Equal("req-123"), path)
		}
	})

	It("fails every operation before calling the API when the call can't be set up", func() {
		noCredentials := NewVaultController(fvcommon.VaultConfig{VaultId: "vault1", BaseVaultUrl: fake.server.URL}, ClientSettings{})
		GinkgoT().Setenv("SKYFLOW_CREDENTIALS", "")
		for path, call := range calls {
			err := call(noCredentials)
			Expect(err).ToNot(BeNil(), path)
			Expect(err.GetMessage()).To(ContainSubstring("Invalid credentials"), path)
		}
		Expect(fake.requests).To(BeEmpty())
	})

	It("validates every operation before calling the API", func() {
		_, err := v.Insert(ctx, fvcommon.InsertRequest{}, fvcommon.InsertOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.Update(ctx, fvcommon.UpdateRequest{}, fvcommon.UpdateOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.Get(ctx, fvcommon.GetRequest{}, fvcommon.GetOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.Delete(ctx, fvcommon.DeleteRequest{}, fvcommon.DeleteOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.Detokenize(ctx, fvcommon.DetokenizeRequest{}, fvcommon.DetokenizeOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.GetTokens(ctx, fvcommon.GetTokensRequest{}, fvcommon.GetTokensOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.Query(ctx, fvcommon.QueryRequest{}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.UploadFiles(ctx, fvcommon.UploadFilesRequest{}, fvcommon.UploadFilesOptions{})
		Expect(err).ToNot(BeNil())
		_, err = v.DeleteFiles(ctx, fvcommon.DeleteFilesRequest{}, fvcommon.DeleteFilesOptions{})
		Expect(err).ToNot(BeNil())
		Expect(fake.requests).To(BeEmpty())
	})

	It("rejects an upload whose file can't be read before calling the API", func() {
		_, err := v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", FilePath: filepath.Join(GinkgoT().TempDir(), "missing.pdf")}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(err.GetMessage()).To(ContainSubstring("Could not read file"))
		Expect(fake.requests).To(BeEmpty())
	})

	It("skips null upload records and records beyond the request", func() {
		fake.respond("/v2/files/upload", http.StatusOK, `{"records":[null,{"skyflowID":"extra","tableName":"t","httpCode":200}]}`)
		resp, err := v.UploadFiles(ctx, validUpload, fvcommon.UploadFilesOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records).To(HaveLen(2))
		Expect(resp.Records[0]).To(Equal(fvcommon.UploadFilesResponseRecord{}))
		Expect(resp.Records[1].SkyflowId).To(Equal("extra"))
		Expect(resp.Records[1].Columns).To(BeNil())
	})
})

var _ = Describe("Response conversion", func() {
	It("returns an empty body and header for a nil response", func() {
		body, header, err := result[api.InsertResponse](nil, nil, nil)
		Expect(err).To(BeNil())
		Expect(body).To(Equal(&api.InsertResponse{}))
		Expect(header).To(Equal(http.Header{}))
	})

	It("uses an empty header when an API error has none", func() {
		hasRecords := func(b *api.InsertResponse) bool { return len(b.Records) > 0 }

		body, header, err := result[api.InsertResponse](nil, core.NewAPIError(http.StatusBadRequest, nil, errors.New(`{"records":[{"httpCode":400}]}`)), hasRecords)
		Expect(err).To(BeNil())
		Expect(body.Records).To(HaveLen(1))
		Expect(header).To(Equal(http.Header{}))

		_, _, err = result[api.InsertResponse](nil, core.NewAPIError(http.StatusBadRequest, nil, errors.New(`{"records":[]}`)), hasRecords)
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetRequestId()).To(BeEmpty())
	})

	It("uses a string error or the status text as the message", func() {
		err := apiError(core.NewAPIError(http.StatusConflict, http.Header{}, errors.New(`{"error":"already exists"}`)), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("already exists"))
		Expect(err.GetHttpCode()).To(Equal("409"))

		err = apiError(core.NewAPIError(http.StatusNotFound, http.Header{}, errors.New("  ")), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("Not Found"))

		err = apiError(core.NewAPIError(http.StatusBadGateway, http.Header{}, nil), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("Bad Gateway"))
		Expect(err.GetHttpCode()).To(Equal("502"))
	})

	It("skips null records", func() {
		header := http.Header{}
		Expect(toRecordResults([]*api.RecordResponseObject{nil}, header)).To(BeEmpty())
		Expect(toDeleteResults([]*api.DeleteResponseObject{nil}, header)).To(BeEmpty())
		Expect(toDetokenizeResults([]*api.DetokenizeResponseObject{nil}, header)).To(BeEmpty())
		Expect(toGetTokensResults([]*api.TokenizeResponseObject{nil}, header)).To(BeEmpty())
		Expect(toDeleteFilesResults([]*api.FileDeleteResponseObject{nil}, header)).To(BeEmpty())
	})
})
