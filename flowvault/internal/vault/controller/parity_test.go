package controller

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// Behaviour aligned with the released Java (flowvault/v1.4.0) and Python (flowvault/v1.1.0) SDKs.
var _ = Describe("VaultController, Java and Python parity", func() {
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

	// The generated request types used to send "tableName": "" and "skyflowIDs": null. These
	// tests guard the omitempty tags added to internal/generated/records.go, so regenerating
	// that code without them fails here.
	Describe("request bodies", func() {
		It("leaves out the request-level tableName when the table is set on each record", func() {
			_, err := v.Insert(ctx, fvcommon.InsertRequest{Records: []fvcommon.InsertRequestRecord{
				{TableName: "users", Data: map[string]interface{}{"name": "a"}},
			}}, fvcommon.InsertOptions{})
			Expect(err).To(BeNil())
			sent := fake.requests["/v2/records/insert"]
			Expect(sent).ToNot(HaveKey("tableName"))
			Expect(sent["records"].([]interface{})[0]).To(HaveKeyWithValue("tableName", "users"))

			_, err = v.Update(ctx, fvcommon.UpdateRequest{Records: []fvcommon.UpdateRequestRecord{
				{TableName: "users", SkyflowId: "id1", Data: map[string]interface{}{"name": "b"}},
			}}, fvcommon.UpdateOptions{})
			Expect(err).To(BeNil())
			sent = fake.requests["/v2/records/update"]
			Expect(sent).ToNot(HaveKey("tableName"))
			Expect(sent["records"].([]interface{})[0]).To(HaveKeyWithValue("tableName", "users"))
		})

		It("still sends a request-level tableName", func() {
			_, err := v.Insert(ctx, fvcommon.InsertRequest{TableName: "users", Records: []fvcommon.InsertRequestRecord{
				{Data: map[string]interface{}{"name": "a"}},
			}}, fvcommon.InsertOptions{})
			Expect(err).To(BeNil())
			Expect(fake.requests["/v2/records/insert"]).To(HaveKeyWithValue("tableName", "users"))
		})

		It("leaves out skyflowIDs on a multi-table get record that uses unique values", func() {
			_, err := v.Get(ctx, fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{
				{TableName: "users", UniqueValues: []map[string]interface{}{{"email": "a@b.c"}}},
			}}, fvcommon.GetOptions{})
			Expect(err).To(BeNil())
			record := fake.requests["/v2/records/get"]["records"].([]interface{})[0]
			Expect(record).ToNot(HaveKey("skyflowIDs"))
			Expect(record).To(HaveKey("uniqueValues"))
		})
	})

	Describe("records of a failed call", func() {
		const failedRecords = `{"records":[
			{"skyflowID":"a","tableName":"t","tokens":{"email":[{"token":"tok"}]},"data":{"email":"x"}},
			{"skyflowID":"b","message":"duplicate","http_code":409},
			{"skyflowID":"c","error":"bad record","httpCode":422},
			{"skyflowID":"d","error":"","statusCode":"404"}]}`

		expectFailed := func(records []fvcommon.InsertResponseRecord) {
			Expect(records).To(HaveLen(4))
			for _, record := range records {
				Expect(record.Error).ToNot(BeEmpty())
				Expect(record.RequestId).To(Equal("req-123"))
				Expect(record.Tokens).To(BeNil())
				Expect(record.Data).To(BeNil())
			}
			Expect(records[0].SkyflowId).To(Equal("a"))
			Expect(records[0].TableName).To(Equal("t"))
			Expect(records[0].HttpCode).To(Equal(http.StatusBadRequest))
			Expect(records[0].Error).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
			Expect(records[1].HttpCode).To(Equal(http.StatusConflict))
			Expect(records[1].Error).To(Equal("duplicate"))
			Expect(records[2].HttpCode).To(Equal(http.StatusUnprocessableEntity))
			Expect(records[2].Error).To(Equal("bad record"))
			Expect(records[3].HttpCode).To(Equal(http.StatusNotFound))
			Expect(records[3].Error).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
		}

		It("reports every record of a failed insert, update or get as failed, as the Java SDK does", func() {
			fake.respond("/v2/records/insert", http.StatusBadRequest, failedRecords)
			inserted, err := v.Insert(ctx, fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{
				{Data: map[string]interface{}{"email": "x"}},
			}}, fvcommon.InsertOptions{})
			Expect(err).To(BeNil())
			expectFailed(inserted.Records)

			fake.respond("/v2/records/update", http.StatusBadRequest, failedRecords)
			updated, err := v.Update(ctx, fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{
				{SkyflowId: "a", Data: map[string]interface{}{"email": "x"}},
			}}, fvcommon.UpdateOptions{})
			Expect(err).To(BeNil())
			expectFailed(updated.Records)

			fake.respond("/v2/records/get", http.StatusBadRequest, failedRecords)
			got, err := v.Get(ctx, fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"a"}}, fvcommon.GetOptions{})
			Expect(err).To(BeNil())
			expectFailed(got.Records)
		})

		It("reports every record of a failed delete as failed", func() {
			fake.respond("/v2/records/delete", http.StatusInternalServerError, `{"records":[{"skyflowID":"a"},{"skyflowID":"b","error":"locked","httpCode":423}]}`)
			resp, err := v.Delete(ctx, fvcommon.DeleteRequest{TableName: "t", SkyflowIds: []string{"a", "b"}}, fvcommon.DeleteOptions{})
			Expect(err).To(BeNil())
			Expect(resp.Records).To(Equal([]fvcommon.DeleteResponseRecord{
				{SkyflowId: "a", HttpCode: http.StatusInternalServerError, Error: fvconsts.UNKNOWN_RECORD_ERROR, RequestId: "req-123"},
				{SkyflowId: "b", HttpCode: http.StatusLocked, Error: "locked", RequestId: "req-123"},
			}))
		})

		It("reports every token of a failed detokenize as failed, without its value or metadata", func() {
			fake.respond("/v2/tokens/detokenize", http.StatusBadRequest, `{"response":[
				{"token":"t1","value":"secret","tokenGroupName":"g","metadata":{"skyflowID":"a","tableName":"t"}},
				{"token":"t2","message":"no access"}]}`)
			resp, err := v.Detokenize(ctx, fvcommon.DetokenizeRequest{Tokens: []string{"t1", "t2"}}, fvcommon.DetokenizeOptions{})
			Expect(err).To(BeNil())
			Expect(resp.Records).To(Equal([]fvcommon.DetokenizeResponseRecord{
				{Token: "t1", TokenGroupName: "g", HttpCode: http.StatusBadRequest, Error: fvconsts.UNKNOWN_RECORD_ERROR, RequestId: "req-123"},
				{Token: "t2", HttpCode: http.StatusBadRequest, Error: "no access", RequestId: "req-123"},
			}))
		})
	})

	It("sends UpdateType in upper case, whatever case it was given in, as the Java SDK does", func() {
		_, err := v.Insert(ctx, fvcommon.InsertRequest{TableName: "t",
			Upsert:  &fvcommon.UpsertOptions{UniqueColumns: []string{"email"}, UpdateType: "replace"},
			Records: []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{"email": "a"}}},
		}, fvcommon.InsertOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/records/insert"]["upsert"]).To(HaveKeyWithValue("updateType", "REPLACE"))

		_, err = v.Update(ctx, fvcommon.UpdateRequest{TableName: "t", UpdateType: "Update",
			Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id1", Data: map[string]interface{}{"email": "b"}}},
		}, fvcommon.UpdateOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/records/update"]).To(HaveKeyWithValue("updateType", "UPDATE"))
	})

	It("takes an API error's httpCode from the HTTP status, not from the body", func() {
		fake.respond("/v2/query", http.StatusBadRequest, `{"error":{"http_code":500,"grpc_code":3,"http_status":"Bad Request","message":"bad query"}}`)
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("bad query"))
	})

	It("returns API errors as the Java SDK does: raw message, int codes and the server's own body", func() {
		fake.respond("/v2/query", http.StatusForbidden, `{"error":{"http_code":403,"grpc_code":7,"http_status":"Forbidden","message":"denied"},"trace":"t-1"}`)
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err.GetMessage()).To(Equal("denied"))
		Expect(err.Error()).To(Equal("denied"))
		Expect(err.GetHttpCodeInt()).To(Equal(http.StatusForbidden))
		Expect(err.GetGrpcCodeInt()).To(Equal(7))
		Expect(err.GetResponseBody()).To(Equal(map[string]interface{}{
			"error": map[string]interface{}{"http_code": 403.0, "grpc_code": 7.0, "http_status": "Forbidden", "message": "denied"},
			"trace": "t-1",
		}))

		fake.respond("/v2/query", http.StatusBadGateway, `<html>bad gateway</html>`)
		_, err = v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err.GetMessage()).To(Equal("<html>bad gateway</html>"))
		Expect(err.GetResponseBody()).To(BeNil())

		fake.respond("/v2/files/upload", http.StatusBadRequest, `{"records":[{"error":"a"},{"error":"b"}]}`)
		_, err = v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{
			TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "cGRm", FileName: "a.pdf"}},
		}}}, fvcommon.UploadFilesOptions{})
		Expect(err.GetMessage()).To(Equal("a; b"))
		Expect(err.GetResponseBody()).To(HaveKey("records"))
	})

	It("returns validation errors with the raw message", func() {
		_, err := v.Query(ctx, fvcommon.QueryRequest{}, fvcommon.QueryOptions{})
		Expect(err.GetMessage()).To(Equal(fverrors.EMPTY_QUERY))
		Expect(err.GetHttpCodeInt()).To(Equal(http.StatusBadRequest))
	})

	It("rejects an upload or file delete whose SkyflowId is set but blank, before calling the API", func() {
		_, err := v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{
			TableName: "t", SkyflowId: "  ", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "cGRm", FileName: "a.pdf"}},
		}}}, fvcommon.UploadFilesOptions{})
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.EMPTY_UPLOAD_FILES_SKYFLOW_ID))
		_, err = v.DeleteFiles(ctx, fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{{
			TableName: "t", SkyflowId: " ", Columns: []string{"c"},
		}}}, fvcommon.DeleteFilesOptions{})
		Expect(err.GetMessage()).To(ContainSubstring(fverrors.EMPTY_DELETE_FILES_SKYFLOW_ID))
		Expect(fake.calls).To(BeEmpty())
	})
})
