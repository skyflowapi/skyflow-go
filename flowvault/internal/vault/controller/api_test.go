package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

// fakeFlowService records each request and answers with the response set for its path.
type fakeFlowService struct {
	mu        sync.Mutex
	server    *httptest.Server
	requests  map[string]map[string]interface{}
	headers   map[string]http.Header
	responses map[string]fakeResponse
	puts      map[string]string
	putCalls  map[string]int
}

type fakeResponse struct {
	status int
	body   string
}

func newFakeFlowService() *fakeFlowService {
	f := &fakeFlowService{
		requests:  map[string]map[string]interface{}{},
		headers:   map[string]http.Header{},
		responses: map[string]fakeResponse{},
		puts:      map[string]string{},
		putCalls:  map[string]int{},
	}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Method == http.MethodPut {
			f.puts[r.URL.Path] = string(raw)
			f.putCalls[r.URL.Path]++
			f.headers[r.URL.Path] = r.Header.Clone()
			if strings.HasPrefix(r.URL.Path, "/signed/unavailable") {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		var body map[string]interface{}
		_ = json.Unmarshal(raw, &body)
		f.requests[r.URL.Path] = body
		f.headers[r.URL.Path] = r.Header.Clone()
		response, ok := f.responses[r.URL.Path]
		if !ok {
			response = fakeResponse{status: http.StatusOK, body: `{}`}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", "req-123")
		w.WriteHeader(response.status)
		_, _ = io.WriteString(w, response.body)
	}))
	return f
}

func (f *fakeFlowService) respond(path string, status int, body string) {
	f.responses[path] = fakeResponse{status: status, body: body}
}

var _ = Describe("VaultController API calls", func() {
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

	It("inserts with request-level table and upsert, and parses tokens and per-record errors", func() {
		fake.respond("/v2/records/insert", http.StatusOK, `{"records":[
			{"skyflowID":"id1","tableName":"users","httpCode":200,"tokens":{"email":[{"token":"tok1","tokenGroupName":"det"}]},"hashedData":{"email":[{"data":"h","hashName":"sha"}]}},
			{"httpCode":400,"error":"bad record"}]}`)
		resp, err := v.Insert(ctx, fvcommon.InsertRequest{
			TableName: "users",
			Upsert:    &fvcommon.UpsertOptions{UniqueColumns: []string{"email"}, UpdateType: fvcommon.UpdateTypeReplace},
			Records:   []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{"email": "a@b.c"}}, {Data: map[string]interface{}{"email": "x"}}},
		}, fvcommon.InsertOptions{})
		Expect(err).To(BeNil())

		sent := fake.requests["/v2/records/insert"]
		Expect(sent["vaultID"]).To(Equal("vault1"))
		Expect(sent["tableName"]).To(Equal("users"))
		Expect(sent["upsert"]).To(Equal(map[string]interface{}{"uniqueColumns": []interface{}{"email"}, "updateType": "REPLACE"}))
		Expect(sent["records"]).To(HaveLen(2))
		Expect(fake.headers["/v2/records/insert"].Get("Authorization")).To(Equal("Bearer " + testApiKey))
		Expect(fake.headers["/v2/records/insert"].Get("sky-metadata")).ToNot(BeEmpty())

		Expect(resp.Records).To(HaveLen(2))
		Expect(resp.Records[0].SkyflowId).To(Equal("id1"))
		Expect(resp.Records[0].Tokens["email"]).To(Equal([]fvcommon.Token{{Token: "tok1", TokenGroupName: "det"}}))
		Expect(resp.Records[0].HashedData["email"]).To(Equal([]fvcommon.HashedValue{{Data: "h", HashName: "sha"}}))
		Expect(resp.Records[0].RequestId).To(BeEmpty())
		Expect(resp.Records[1].Error).To(Equal("bad record"))
		Expect(resp.Records[1].HttpCode).To(Equal(400))
		Expect(resp.Records[1].RequestId).To(Equal("req-123"))
	})

	It("sends record-level tables and upserts", func() {
		_, err := v.Insert(ctx, fvcommon.InsertRequest{Records: []fvcommon.InsertRequestRecord{
			{Data: map[string]interface{}{"a": 1}, TableName: "t1", Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}}},
		}}, fvcommon.InsertOptions{})
		Expect(err).To(BeNil())
		record := fake.requests["/v2/records/insert"]["records"].([]interface{})[0].(map[string]interface{})
		Expect(record["tableName"]).To(Equal("t1"))
		Expect(record["upsert"]).To(Equal(map[string]interface{}{"uniqueColumns": []interface{}{"a"}}))
	})

	It("returns per-record results from a non-2xx response that carries records", func() {
		fake.respond("/v2/records/update", http.StatusBadRequest, `{"records":[{"skyflowID":"id1","httpCode":400,"error":"column missing"}]}`)
		resp, err := v.Update(ctx, fvcommon.UpdateRequest{
			TableName: "users", UpdateType: fvcommon.UpdateTypeUpdate,
			Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id1", Data: map[string]interface{}{"a": 1}}},
		}, fvcommon.UpdateOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/records/update"]["updateType"]).To(Equal("UPDATE"))
		Expect(resp.Records[0].Error).To(Equal("column missing"))
		Expect(resp.Records[0].RequestId).To(Equal("req-123"))
	})

	It("maps a whole-call API error to a SkyflowError", func() {
		fake.respond("/v2/query", http.StatusBadRequest, `{"error":{"grpc_code":3,"http_code":400,"http_status":"Bad Request","message":"Invalid query"}}`)
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring("Invalid query"))
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetRequestId()).To(Equal("req-123"))
	})

	It("maps an error with an unexpected status code", func() {
		fake.respond("/v2/records/get", http.StatusBadGateway, `{"error":{"http_code":502,"message":"upstream down"}}`)
		_, err := v.Get(ctx, fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}}, fvcommon.GetOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring("upstream down"))
	})

	It("keeps the status code and request ID when the error body isn't a Skyflow error", func() {
		fake.respond("/v2/query", http.StatusForbidden, `<html>denied</html>`)
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err.GetHttpCode()).To(Equal("403"))
		Expect(err.GetMessage()).To(ContainSubstring("<html>denied</html>"))
		Expect(err.GetRequestId()).To(Equal("req-123"))

		fake.respond("/v2/query", http.StatusBadRequest, `{"error":{"message":"no code"}}`)
		_, err = v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("no code"))
	})

	It("returns a final retryable response without waiting", func() {
		noRetries := 0
		v = NewVaultController(fvcommon.VaultConfig{VaultId: "vault1", BaseVaultUrl: fake.server.URL, MaxRetries: &noRetries},
			ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}})
		fake.respond("/v2/query", http.StatusServiceUnavailable, `{"error":{"http_code":503,"message":"busy"}}`)
		start := time.Now()
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(time.Since(start)).To(BeNumerically("<", 300*time.Millisecond))
		Expect(err.GetHttpCode()).To(Equal("503"))
		Expect(err.GetMessage()).To(ContainSubstring("busy"))
		Expect(err.GetRequestId()).To(Equal("req-123"))
	})

	It("gets from one table with limit, offset and redactions", func() {
		fake.respond("/v2/records/get", http.StatusOK, `{"records":[{"skyflowID":"id1","data":{"name":"j"},"httpCode":200}]}`)
		resp, err := v.Get(ctx, fvcommon.GetRequest{
			TableName: "t", SkyflowIds: []string{"id1"}, Columns: []string{"name"}, Limit: 10, Offset: 5,
			ColumnRedactions: []fvcommon.ColumnRedaction{{ColumnName: "name", Redaction: "MASKED"}},
		}, fvcommon.GetOptions{})
		Expect(err).To(BeNil())
		sent := fake.requests["/v2/records/get"]
		Expect(sent["tableName"]).To(Equal("t"))
		Expect(sent["limit"]).To(BeEquivalentTo(10))
		Expect(sent["offset"]).To(BeEquivalentTo(5))
		Expect(sent["columnRedactions"]).To(Equal([]interface{}{map[string]interface{}{"columnName": "name", "redaction": "MASKED"}}))
		Expect(resp.Records[0].Data).To(Equal(map[string]interface{}{"name": "j"}))
	})

	It("gets from several tables", func() {
		_, err := v.Get(ctx, fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{
			{TableName: "t1", SkyflowIds: []string{"a"}},
			{TableName: "t2", UniqueValues: []map[string]interface{}{{"email": "x"}}},
		}}, fvcommon.GetOptions{})
		Expect(err).To(BeNil())
		sent := fake.requests["/v2/records/get"]
		Expect(sent).ToNot(HaveKey("tableName"))
		records := sent["records"].([]interface{})
		Expect(records[1].(map[string]interface{})["uniqueValues"]).To(Equal([]interface{}{map[string]interface{}{"data": map[string]interface{}{"email": "x"}}}))
	})

	It("deletes records", func() {
		fake.respond("/v2/records/delete", http.StatusOK, `{"records":[{"skyflowID":"id1","httpCode":200}]}`)
		resp, err := v.Delete(ctx, fvcommon.DeleteRequest{TableName: "t", SkyflowIds: []string{"id1"}}, fvcommon.DeleteOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/records/delete"]["skyflowIDs"]).To(Equal([]interface{}{"id1"}))
		Expect(resp.Records).To(Equal([]fvcommon.DeleteResponseRecord{{SkyflowId: "id1", HttpCode: 200}}))
	})

	It("detokenizes with redactions and metadata", func() {
		fake.respond("/v2/tokens/detokenize", http.StatusOK, `{"response":[
			{"token":"tok1","value":"john","tokenGroupName":"g","httpCode":200,"metadata":{"skyflowID":"id1","tableName":"users"}},
			{"token":"tok2","httpCode":404,"error":"not found"}]}`)
		resp, err := v.Detokenize(ctx, fvcommon.DetokenizeRequest{
			Tokens:               []string{"tok1", "tok2"},
			TokenGroupRedactions: []fvcommon.TokenGroupRedaction{{TokenGroupName: "g", Redaction: "PLAIN_TEXT"}},
		}, fvcommon.DetokenizeOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/tokens/detokenize"]["tokenGroupRedactions"]).To(Equal([]interface{}{map[string]interface{}{"tokenGroupName": "g", "redaction": "PLAIN_TEXT"}}))
		Expect(resp.Records[0].Value).To(Equal("john"))
		Expect(resp.Records[0].Metadata).To(Equal(&fvcommon.DetokenizeRecordMetadata{SkyflowId: "id1", TableName: "users"}))
		Expect(resp.Records[1].HttpCode).To(Equal(404))
		Expect(resp.Records[1].RequestId).To(Equal("req-123"))
	})

	It("gets tokens for values", func() {
		fake.respond("/v2/records/getTokens", http.StatusOK, `{"records":[{"token":"tok1","value":"a@b.c","tokenGroupName":"det","httpCode":200}]}`)
		resp, err := v.GetTokens(ctx, fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "a@b.c", TokenGroupName: "det"}}}, fvcommon.GetTokensOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/records/getTokens"]["records"]).To(Equal([]interface{}{map[string]interface{}{"value": "a@b.c", "tokenGroupName": "det"}}))
		Expect(resp.Records).To(Equal([]fvcommon.GetTokensResponseRecord{{Value: "a@b.c", TokenGroupName: "det", Token: "tok1", HttpCode: 200}}))
	})

	It("runs a query and returns rows, columns and the request ID", func() {
		fake.respond("/v2/query", http.StatusOK, `{"records":[{"data":{"name":"j"}}],"metadata":{"columns":["name"]}}`)
		resp, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT name FROM users"}, fvcommon.QueryOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/query"]["query"]).To(Equal("SELECT name FROM users"))
		Expect(resp.Fields).To(Equal([]map[string]interface{}{{"name": "j"}}))
		Expect(resp.Metadata.Columns).To(Equal([]string{"name"}))
		Expect(resp.RequestId).To(Equal("req-123"))
	})

	It("uploads files to their signed URLs and reports each column", func() {
		fake.respond("/v2/files/upload", http.StatusOK, `{"records":[
			{"skyflowID":"new1","tableName":"docs","httpCode":200,"data":{"resume":"`+fake.server.URL+`/signed/resume"}},
			{"skyflowID":"id2","tableName":"docs","httpCode":404,"error":"record not found"}]}`)
		resp, err := v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{
				{Column: "resume", Base64: "cGRm", FileName: "resume.pdf"},
				{Column: "photo", Base64: "aW1n", FileName: "photo.png"},
			}},
			{TableName: "docs", SkyflowId: "id2", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "resume", Base64: "cGRm", FileName: "b.pdf"}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(err).To(BeNil())

		sent := fake.requests["/v2/files/upload"]["records"].([]interface{})
		Expect(sent[0].(map[string]interface{})).ToNot(HaveKey("skyflowID"))
		Expect(sent[0].(map[string]interface{})["columns"]).To(Equal([]interface{}{
			map[string]interface{}{"column": "resume", "fileName": "resume.pdf"},
			map[string]interface{}{"column": "photo", "fileName": "photo.png"},
		}))
		Expect(sent[1].(map[string]interface{})["skyflowID"]).To(Equal("id2"))

		Expect(fake.puts["/signed/resume"]).To(Equal("pdf"))
		Expect(fake.headers["/signed/resume"].Get("Authorization")).To(BeEmpty())
		Expect(resp.Records[0].SkyflowId).To(Equal("new1"))
		Expect(resp.Records[0].Columns).To(Equal([]fvcommon.UploadFilesColumnResult{
			{Column: "resume", FileName: "resume.pdf", UploadStatus: fvcommon.UploadStatusUploaded},
			{Column: "photo", FileName: "photo.png", UploadStatus: fvcommon.UploadStatusSkipped},
		}))
		Expect(resp.Records[1].Error).To(Equal("record not found"))
		Expect(resp.Records[1].RequestId).To(Equal("req-123"))
		Expect(resp.Records[1].Columns[0].UploadStatus).To(Equal(fvcommon.UploadStatusSkipped))
		Expect(resp.Records[1].Columns[0].Error).To(Equal("record not found"))
	})

	It("doesn't retry a failed file upload, even with retries on", func() {
		retries := 3
		v.SetConfig(fvcommon.VaultConfig{VaultId: "vault1", BaseVaultUrl: fake.server.URL, MaxRetries: &retries, InitialRetryDelay: time.Millisecond}, false)
		fake.respond("/v2/files/upload", http.StatusOK, `{"records":[
			{"skyflowID":"id1","tableName":"docs","httpCode":200,"data":{"resume":"`+fake.server.URL+`/signed/unavailable"}}]}`)
		resp, err := v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "resume", Base64: "cGRm", FileName: "resume.pdf"}}},
		}}, fvcommon.UploadFilesOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { rc.AddHeader(fvcommon.RequestIdHeader, "mine") },
		})
		Expect(err).To(BeNil())
		Expect(fake.putCalls["/signed/unavailable"]).To(Equal(1))
		Expect(fake.headers["/signed/unavailable"].Get("x-request-id")).To(BeEmpty())
		Expect(fake.headers["/signed/unavailable"].Get("sky-metadata")).To(BeEmpty())
		Expect(resp.Records[0].Columns[0]).To(Equal(fvcommon.UploadFilesColumnResult{
			Column: "resume", FileName: "resume.pdf", UploadStatus: fvcommon.UploadStatusFailed, Error: "PUT failed: 503",
		}))
	})

	It("deletes files", func() {
		fake.respond("/v2/files/delete", http.StatusOK, `{"records":[{"skyflowID":"id1","tableName":"docs","httpCode":200,"data":{"resume":{"status":"DELETED"}}}]}`)
		resp, err := v.DeleteFiles(ctx, fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{
			{TableName: "docs", Columns: []string{"resume"}, UniqueValues: []map[string]interface{}{{"email": "x"}}},
		}}, fvcommon.DeleteFilesOptions{})
		Expect(err).To(BeNil())
		record := fake.requests["/v2/files/delete"]["records"].([]interface{})[0].(map[string]interface{})
		Expect(record).ToNot(HaveKey("skyflowID"))
		Expect(record["columns"]).To(Equal([]interface{}{"resume"}))
		Expect(resp.Records[0].Columns).To(Equal([]fvcommon.DeleteFilesColumnResult{{Column: "resume", Status: "DELETED"}}))
	})

	It("sends interceptor headers", func() {
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { rc.AddHeader(fvcommon.RequestIdHeader, "mine") },
		})
		Expect(err).To(BeNil())
		Expect(fake.headers["/v2/query"].Get("x-request-id")).To(Equal("mine"))
	})

	It("reports a network failure as a SkyflowError", func() {
		fake.server.Close()
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetCode()).To(ContainSubstring("500"))
	})
})
