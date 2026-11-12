package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
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
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.INVALID_FILE_PATH))
		Expect(fake.requests).To(BeEmpty())
	})

	It("uploads nothing when the response has a different number of records", func() {
		fake.respond("/v2/files/upload", http.StatusOK, `{"records":[null,{"skyflowID":"extra","tableName":"t","httpCode":200}]}`)
		resp, err := v.UploadFiles(ctx, validUpload, fvcommon.UploadFilesOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records).To(HaveLen(len(validUpload.Records)))
		for _, record := range resp.Records {
			Expect(record.Error).To(Equal(fvconsts.UPLOAD_RESPONSE_MISMATCH_ERROR))
			Expect(record.HttpCode).To(Equal(http.StatusInternalServerError))
			Expect(record.RequestId).To(Equal("req-123"))
			for _, column := range record.Columns {
				Expect(column.UploadStatus).To(Equal(fvcommon.UploadStatusSkipped))
				Expect(column.Error).To(Equal(fvconsts.UPLOAD_RESPONSE_MISMATCH_ERROR))
			}
		}
	})

	It("doesn't upload into a null record or one for another table", func() {
		twoRecords := fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "YQ==", FileName: "a.txt"}}},
			{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "YQ==", FileName: "b.txt"}}},
		}}
		fake.respond("/v2/files/upload", http.StatusOK, `{"records":[null,{"skyflowID":"x","tableName":"other","httpCode":200,"data":{"c":"`+fake.server.URL+`/signed/c"}}]}`)
		resp, err := v.UploadFiles(ctx, twoRecords, fvcommon.UploadFilesOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records[0].TableName).To(Equal("t"))
		Expect(resp.Records[0].Error).To(Equal(fvconsts.UPLOAD_RESPONSE_MISMATCH_ERROR))
		Expect(resp.Records[0].Columns[0].FileName).To(Equal("a.txt"))
		Expect(resp.Records[1].Error).To(Equal(fvconsts.UPLOAD_RESPONSE_MISMATCH_ERROR))
		Expect(resp.Records[1].HttpCode).To(Equal(http.StatusOK))
		Expect(fake.putCalls).To(BeEmpty())
	})
})

var _ = Describe("Response conversion", func() {
	It("returns an empty body and header for a nil response", func() {
		body, header, err := result[api.InsertResponse](context.Background(), nil, nil, nil)
		Expect(err).To(BeNil())
		Expect(body).To(Equal(&api.InsertResponse{}))
		Expect(header).To(Equal(http.Header{}))
	})

	It("uses an empty header when an API error has none", func() {
		hasRecords := func(b *api.InsertResponse) bool { return len(b.Records) > 0 }

		body, header, err := result[api.InsertResponse](context.Background(), nil, core.NewAPIError(http.StatusBadRequest, nil, errors.New(`{"records":[{"httpCode":400}]}`)), hasRecords)
		Expect(err).To(BeNil())
		Expect(body.Records).To(HaveLen(1))
		Expect(header).To(Equal(http.Header{}))

		_, _, err = result[api.InsertResponse](context.Background(), nil, core.NewAPIError(http.StatusBadRequest, nil, errors.New(`{"records":[]}`)), hasRecords)
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetRequestId()).To(BeEmpty())
	})

	It("uses an error string, a top-level message, the raw body or the status code as the message", func() {
		err := apiError(core.NewAPIError(http.StatusConflict, http.Header{}, errors.New(`{"error":"already exists"}`)), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("already exists"))
		Expect(err.GetHttpCode()).To(Equal("409"))

		err = apiError(core.NewAPIError(http.StatusBadRequest, http.Header{}, errors.New(`{"message":"bad table","code":3}`)), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("bad table"))
		Expect(err.GetMessage()).ToNot(ContainSubstring("code"))

		err = apiError(core.NewAPIError(http.StatusBadGateway, http.Header{}, errors.New("upstream down")), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("upstream down"))

		err = apiError(core.NewAPIError(http.StatusNotFound, http.Header{}, errors.New("  ")), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("Error with status code 404"))

		long := `{"error":"` + strings.Repeat("y", 1000) + `"}`
		err = apiError(core.NewAPIError(http.StatusConflict, http.Header{}, errors.New(long)), http.Header{})
		Expect(err.GetMessage()).To(HaveSuffix("... (truncated)"))

		err = apiError(core.NewAPIError(http.StatusBadGateway, http.Header{}, nil), http.Header{})
		Expect(err.GetMessage()).To(ContainSubstring("Error with status code 502"))
		Expect(err.GetHttpCode()).To(Equal("502"))
	})

	It("joins a failed uploadFiles call's record errors and keeps other errors as they are", func() {
		original := fverrors.NewSkyflowError(fverrors.TRANSPORT_ERROR, "dial failed")
		Expect(uploadFilesCallError(errors.New("dial failed"), original)).To(BeIdenticalTo(original))
		Expect(uploadFilesCallError(core.NewAPIError(http.StatusBadGateway, nil, nil), original)).To(BeIdenticalTo(original))

		err := uploadFilesCallError(core.NewAPIError(http.StatusBadRequest, nil, errors.New(`{"records":[{"error":"a"},{"error":"b"}]}`)), original)
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("a; b"))

		many := make([]string, 200)
		for i := range many {
			many[i] = fmt.Sprintf(`{"error":"record %d failed"}`, i)
		}
		err = uploadFilesCallError(core.NewAPIError(http.StatusBadRequest, nil, errors.New(`{"records":[`+strings.Join(many, ",")+`]}`)), original)
		Expect(err.GetMessage()).To(ContainSubstring("... (truncated)"))
	})

	It("treats a record of a failed call as failed unless it reports a 2xx itself", func() {
		header := http.Header{}
		header.Set("x-request-id", "req-9")
		ok, bad := 200, 404
		raw := []map[string]interface{}{
			{"token": "t1", "httpCode": 200.0},
			{"token": "t2"},
			{"token": "t3", "httpCode": 404.0},
			{"token": "t4", "http_code": "404", "message": "not found"},
			{"token": "t5", "statusCode": 201.0},
		}
		tokens := toGetTokensResults([]*api.TokenizeResponseObject{
			{Token: "t1", HTTPCode: &ok},
			{Token: "t2"},
			{Token: "t3", HTTPCode: &bad},
			{Token: "t4"},
			{Token: "t5"},
		}, header, http.StatusBadRequest, raw)
		Expect(tokens[0].Error).To(BeEmpty())
		Expect(tokens[0].RequestId).To(BeEmpty())
		Expect(tokens[1].Error).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
		Expect(tokens[1].HttpCode).To(Equal(http.StatusBadRequest))
		Expect(tokens[1].RequestId).To(Equal("req-9"))
		Expect(tokens[2].Error).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
		Expect(tokens[2].HttpCode).To(Equal(404))
		// As in the Java SDK, "message" and the http_code / statusCode spellings are read too.
		Expect(tokens[3].Error).To(Equal("not found"))
		Expect(tokens[3].HttpCode).To(Equal(404))
		Expect(tokens[4].Error).To(BeEmpty())
		Expect(tokens[4].HttpCode).To(Equal(201))

		files := toDeleteFilesResults([]*api.FileDeleteResponseObject{{TableName: "t", HTTPCode: 500}}, header, http.StatusMultiStatus,
			[]map[string]interface{}{{"tableName": "t", "httpCode": 500.0}})
		Expect(files[0].Error).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
		Expect(files[0].HttpCode).To(Equal(500))
		Expect(files[0].RequestId).To(Equal("req-9"))

		// In a successful call a record's own values are kept.
		files = toDeleteFilesResults([]*api.FileDeleteResponseObject{{TableName: "t", HTTPCode: 500}}, header, 0, nil)
		Expect(files[0].Error).To(BeEmpty())
		Expect(files[0].HttpCode).To(Equal(500))
	})

	It("fills in a record's missing httpCode in a successful call, as in Java: 500 with an error, else 200", func() {
		header := http.Header{}
		failed := "boom"
		records := toRecordResults([]*api.RecordResponseObject{{}, {Error: &failed}, {HTTPCode: 207}}, header)
		Expect([]int{records[0].HttpCode, records[1].HttpCode, records[2].HttpCode}).To(Equal([]int{200, 500, 207}))
		deleted := toDeleteResults([]*api.DeleteResponseObject{{}, {Error: &failed}}, header)
		Expect([]int{deleted[0].HttpCode, deleted[1].HttpCode}).To(Equal([]int{200, 500}))
		detokenized := toDetokenizeResults([]*api.DetokenizeResponseObject{{}, {Error: &failed}}, header)
		Expect([]int{detokenized[0].HttpCode, detokenized[1].HttpCode}).To(Equal([]int{200, 500}))
		tokens := toGetTokensResults([]*api.TokenizeResponseObject{{}, {Error: &failed}}, header, 0, nil)
		Expect([]int{tokens[0].HttpCode, tokens[1].HttpCode}).To(Equal([]int{200, 500}))
		files := toDeleteFilesResults([]*api.FileDeleteResponseObject{{}, {Error: &failed}}, header, 0, nil)
		Expect([]int{files[0].HttpCode, files[1].HttpCode}).To(Equal([]int{200, 500}))
	})

	It("reads a failed call's records as raw maps, and only from an API error with a matching body", func() {
		apiErr := func(body string) error { return core.NewAPIError(http.StatusBadRequest, nil, errors.New(body)) }
		Expect(failedRecordMaps(errors.New("network"), fvconsts.ERROR_KEY_RECORDS)).To(BeNil())
		Expect(failedRecordMaps(core.NewAPIError(http.StatusBadRequest, nil, nil), fvconsts.ERROR_KEY_RECORDS)).To(BeNil())
		Expect(failedRecordMaps(apiErr(`not json`), fvconsts.ERROR_KEY_RECORDS)).To(BeNil())
		Expect(failedRecordMaps(apiErr(`{"records":"nope"}`), fvconsts.ERROR_KEY_RECORDS)).To(BeNil())
		Expect(failedRecordMaps(apiErr(`{"records":[{"a":1},7,null]}`), fvconsts.ERROR_KEY_RECORDS)).
			To(Equal([]map[string]interface{}{{"a": 1.0}, nil, nil}))
		Expect(rawAt(nil, 0)).To(BeNil())

		Expect(readString(map[string]interface{}{"n": 42.0, "b": true}, "n")).To(Equal("42"))
		Expect(readString(map[string]interface{}{"b": true}, "b")).To(Equal("true"))
		Expect(readString(nil, "missing")).To(BeEmpty())
		Expect(failedRecordHTTPCode(map[string]interface{}{"httpCode": "x", "statusCode": 418.0}, 400)).To(Equal(418))
		Expect(failedRecordErrorOrSuccess(map[string]interface{}{"error": "boom", "httpCode": 200.0})).To(Equal("boom"))
		Expect(failedRecordErrorOrSuccess(nil)).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
	})

	It("skips null records", func() {
		header := http.Header{}
		Expect(toRecordResults([]*api.RecordResponseObject{nil}, header)).To(BeEmpty())
		Expect(toDeleteResults([]*api.DeleteResponseObject{nil}, header)).To(BeEmpty())
		Expect(toDetokenizeResults([]*api.DetokenizeResponseObject{nil}, header)).To(BeEmpty())
		Expect(toGetTokensResults([]*api.TokenizeResponseObject{nil}, header, 0, nil)).To(BeEmpty())
		Expect(toDeleteFilesResults([]*api.FileDeleteResponseObject{nil}, header, 0, nil)).To(BeEmpty())
	})
})

var _ = Describe("rawBodyMessage", func() {
	It("keeps a short body and cuts a long one without splitting a character", func() {
		Expect(rawBodyMessage("short")).To(Equal("short"))
		body := strings.Repeat("a", 511) + "é" + strings.Repeat("b", 10) // é spans bytes 511-512
		Expect(rawBodyMessage(body)).To(Equal(strings.Repeat("a", 511) + "... (truncated)"))
	})
})
