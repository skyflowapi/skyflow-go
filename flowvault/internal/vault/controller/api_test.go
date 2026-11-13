package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
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
	calls     map[string]int
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
		calls:     map[string]int{},
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
		f.calls[r.URL.Path]++
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
		Expect(resp.Records[0].HashedData["email"]).To(Equal([]interface{}{map[string]interface{}{"data": "h", "hashName": "sha"}}))
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

	// Response filling for a non-2xx response that still lists records: the status
	// comes from the record (any of its spellings) or the call, the error from "error",
	// "message" or "Unknown error", every record gets the request ID, and data is dropped.
	It("fills in failed records from a non-2xx response", func() {
		fake.respond("/v2/records/insert", http.StatusBadRequest, `{"records":[
			{"skyflowID":"id1","tableName":"users","http_code":409,"error":"duplicate","data":{"a":1}},
			{"statusCode":"422","message":"bad column"},
			{"httpCode":"abc","error":{"code":7}},
			{"error":""}]}`)
		resp, err := v.Insert(ctx, fvcommon.InsertRequest{TableName: "users", Records: []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{"a": 1}}}}, fvcommon.InsertOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records).To(HaveLen(4))
		Expect(resp.Records[0]).To(Equal(fvcommon.InsertResponseRecord{
			SkyflowId: "id1", TableName: "users", HttpCode: 409, Error: "duplicate", RequestId: "req-123",
		}))
		Expect(resp.Records[1].HttpCode).To(Equal(422))
		Expect(resp.Records[1].Error).To(Equal("bad column"))
		Expect(resp.Records[2].HttpCode).To(Equal(400)) // the call's status
		Expect(resp.Records[2].Error).To(Equal(`{"code":7}`))
		Expect(resp.Records[3].Error).To(Equal("Unknown error"))
		for _, record := range resp.Records {
			Expect(record.RequestId).To(Equal("req-123"))
		}

		fake.respond("/v2/tokens/detokenize", http.StatusNotFound, `{"response":[{"token":"t1","tokenGroupName":"g","value":"secret"}]}`)
		detok, err := v.Detokenize(ctx, fvcommon.DetokenizeRequest{Tokens: []string{"t1"}}, fvcommon.DetokenizeOptions{})
		Expect(err).To(BeNil())
		Expect(detok.Records[0].Token).To(Equal("t1"))
		Expect(detok.Records[0].TokenGroupName).To(Equal("g"))
		Expect(detok.Records[0].Value).To(BeNil()) // values are not carried over
		Expect(detok.Records[0].HttpCode).To(Equal(404))
		Expect(detok.Records[0].Error).To(Equal("Unknown error"))
	})

	// Response filling on a 2xx response: a record's missing httpCode (0) is 500 when
	// the record has an error and 200 when it doesn't.
	It("fills in a missing record httpCode on a 2xx response", func() {
		fake.respond("/v2/records/delete", http.StatusOK, `{"records":[{"skyflowID":"id1"},{"skyflowID":"id2","error":"not found"},{"skyflowID":"id3","httpCode":207}]}`)
		resp, err := v.Delete(ctx, fvcommon.DeleteRequest{TableName: "t", SkyflowIds: []string{"id1", "id2", "id3"}}, fvcommon.DeleteOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records[0].HttpCode).To(Equal(200))
		Expect(resp.Records[0].RequestId).To(BeEmpty())
		Expect(resp.Records[1].HttpCode).To(Equal(500))
		Expect(resp.Records[1].RequestId).To(Equal("req-123"))
		Expect(resp.Records[2].HttpCode).To(Equal(207))
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
			TableName: "t", SkyflowIds: []string{"id1"}, Columns: []string{"name"}, Limit: intPtr(10), Offset: intPtr(5),
			ColumnRedactions: []fvcommon.ColumnRedactions{{ColumnName: "name", Redaction: "MASKED"}},
		}, fvcommon.GetOptions{})
		Expect(err).To(BeNil())
		sent := fake.requests["/v2/records/get"]
		Expect(sent["tableName"]).To(Equal("t"))
		Expect(sent["limit"]).To(BeEquivalentTo(10))
		Expect(sent["offset"]).To(BeEquivalentTo(5))
		Expect(sent["columnRedactions"]).To(Equal([]interface{}{map[string]interface{}{"columnName": "name", "redaction": "MASKED"}}))
		Expect(resp.Records[0].Data).To(Equal(map[string]interface{}{"name": "j"}))
	})

	It("sends an explicit zero offset and omits an unset limit", func() {
		fake.respond("/v2/records/get", http.StatusOK, `{"records":[]}`)
		_, err := v.Get(ctx, fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id1"}, Offset: intPtr(0)}, fvcommon.GetOptions{})
		Expect(err).To(BeNil())
		sent := fake.requests["/v2/records/get"]
		Expect(sent).To(HaveKeyWithValue("offset", BeEquivalentTo(0)))
		Expect(sent).ToNot(HaveKey("limit"))
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
			TokenGroupRedactions: []fvcommon.TokenGroupRedactions{{TokenGroupName: "g", Redaction: "PLAIN_TEXT"}},
		}, fvcommon.DetokenizeOptions{})
		Expect(err).To(BeNil())
		Expect(fake.requests["/v2/tokens/detokenize"]["tokenGroupRedactions"]).To(Equal([]interface{}{map[string]interface{}{"tokenGroupName": "g", "redaction": "PLAIN_TEXT"}}))
		Expect(resp.Records[0].Value).To(Equal("john"))
		Expect(resp.Records[0].Metadata).To(Equal(&fvcommon.DetokenizeResponseRecordMetadata{SkyflowId: "id1", TableName: "users"}))
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
		Expect(resp.Records).To(Equal([]map[string]interface{}{{"name": "j"}}))
		Expect(resp.Fields).To(Equal(resp.Records)) //nolint:staticcheck // the deprecated alias still works
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
			{Column: "photo", FileName: "photo.png", UploadStatus: fvcommon.UploadStatusSkipped, Error: fvconsts.NO_SIGNED_URL_ERROR},
		}))
		// The vault accepted the record, but one of its files wasn't uploaded.
		Expect(resp.Records[0].Error).To(Equal("1 of 2 files were not uploaded. See each column for details."))
		Expect(resp.Records[0].HttpCode).To(Equal(200))
		Expect(resp.Records[0].RequestId).To(Equal("req-123"))
		Expect(resp.Records[1].Error).To(Equal("record not found"))
		Expect(resp.Records[1].RequestId).To(Equal("req-123"))
		Expect(resp.Records[1].Columns[0].UploadStatus).To(Equal(fvcommon.UploadStatusSkipped))
		Expect(resp.Records[1].Columns[0].Error).To(Equal("record not found"))
	})

	It("returns an error when the upload call fails, even when the body carries records", func() {
		fake.respond("/v2/files/upload", http.StatusBadRequest, `{"records":[
			{"skyflowID":"id1","httpCode":400,"error":"column not found"},
			{"skyflowID":"id2","httpCode":400,"error":"column not found"},
			{"skyflowID":"id3","httpCode":400,"error":"table not found"}]}`)
		resp, err := v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "resume", Base64: "cGRm", FileName: "resume.pdf"}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(resp).To(BeNil())
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("column not found; table not found"))
		Expect(err.GetRequestId()).To(Equal("req-123"))

		fake.respond("/v2/files/upload", http.StatusBadRequest, `{"records":[{"skyflowID":"id1","httpCode":400}]}`)
		_, err = v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "resume", Base64: "cGRm", FileName: "resume.pdf"}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring(`"records"`))
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
		Expect(resp.Records[0].Error).To(Equal("1 of 1 files were not uploaded. See each column for details."))
	})

	It("uploads one *os.File used by two columns in full to both", func() {
		path := filepath.Join(GinkgoT().TempDir(), "shared.txt")
		Expect(os.WriteFile(path, []byte(strings.Repeat("0123456789", 10000)), 0o600)).To(Succeed())
		handle, err := os.Open(path)
		Expect(err).To(BeNil())
		DeferCleanup(handle.Close)
		fake.respond("/v2/files/upload", http.StatusOK, `{"records":[{"skyflowID":"id1","tableName":"docs","httpCode":200,"data":{
			"a":"`+fake.server.URL+`/signed/a","b":"`+fake.server.URL+`/signed/b"}}]}`)
		resp, callErr := v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "a", FileObject: handle}, {Column: "b", FileObject: handle}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(callErr).To(BeNil())
		Expect(resp.Records[0].Error).To(BeEmpty())
		Expect(fake.puts["/signed/a"]).To(HaveLen(100000))
		Expect(fake.puts["/signed/b"]).To(Equal(fake.puts["/signed/a"]))
	})

	It("fails an uploadFiles call with an error status even when the body has records", func() {
		request := fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "a", Base64: "cGRm", FileName: "a.pdf"}}},
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "a", Base64: "cGRm", FileName: "a.pdf"}}},
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "a", Base64: "cGRm", FileName: "a.pdf"}}},
		}}
		fake.respond("/v2/files/upload", http.StatusBadRequest, `{"records":[
			{"tableName":"docs","httpCode":400,"error":"bad column"},
			{"skyflowID":"id2","tableName":"docs","data":{"a":"`+fake.server.URL+`/signed/a"}},
			{"tableName":"docs","httpCode":400,"error":"bad column"},
			null,
			{"tableName":"docs","error":"quota exceeded"}]}`)
		resp, err := v.UploadFiles(ctx, request, fvcommon.UploadFilesOptions{})
		Expect(resp).To(BeNil())
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("bad column; quota exceeded"))
		Expect(err.GetRequestId()).To(Equal("req-123"))
		Expect(fake.puts).To(BeEmpty())

		fake.respond("/v2/files/upload", http.StatusBadRequest, `{"records":[{"skyflowID":"id2","tableName":"docs","data":{"a":"`+fake.server.URL+`/signed/a"}}]}`)
		_, err = v.UploadFiles(ctx, request, fvcommon.UploadFilesOptions{})
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring(`"skyflowID":"id2"`))
		Expect(fake.puts).To(BeEmpty())

		fake.respond("/v2/files/upload", http.StatusBadRequest, `{"records":{"error":"not a list"}}`)
		_, err = v.UploadFiles(ctx, request, fvcommon.UploadFilesOptions{})
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("not a list"))
	})

	It("doesn't retry uploadFiles on a 5xx, but retries other calls and 429", func() {
		retries := 2
		v.SetConfig(fvcommon.VaultConfig{VaultId: "vault1", BaseVaultUrl: fake.server.URL, MaxRetries: &retries, InitialRetryDelay: time.Millisecond}, false)
		fake.respond("/v2/files/upload", http.StatusInternalServerError, `{"error":{"message":"boom"}}`)
		fake.respond("/v2/files/delete", http.StatusServiceUnavailable, `{"error":{"message":"boom"}}`)
		fake.respond("/v2/records/getTokens", http.StatusServiceUnavailable, `{"error":{"message":"boom"}}`)

		_, err := v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "resume", Base64: "cGRm", FileName: "resume.pdf"}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(err.GetHttpCode()).To(Equal("500"))
		_, err = v.DeleteFiles(ctx, fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{
			{TableName: "docs", Columns: []string{"resume"}, SkyflowId: "id1"},
		}}, fvcommon.DeleteFilesOptions{})
		Expect(err.GetHttpCode()).To(Equal("503"))
		_, err = v.GetTokens(ctx, fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "v", TokenGroupName: "g"}}}, fvcommon.GetTokensOptions{})
		Expect(err).ToNot(BeNil())
		Expect(fake.calls["/v2/files/upload"]).To(Equal(1))
		Expect(fake.calls["/v2/files/delete"]).To(Equal(3))
		Expect(fake.calls["/v2/records/getTokens"]).To(Equal(3))

		fake.respond("/v2/files/upload", http.StatusTooManyRequests, `{"error":{"message":"slow down"}}`)
		_, _ = v.UploadFiles(ctx, fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{
			{TableName: "docs", Columns: []fvcommon.UploadFilesRequestColumn{{Column: "resume", Base64: "cGRm", FileName: "resume.pdf"}}},
		}}, fvcommon.UploadFilesOptions{})
		Expect(fake.calls["/v2/files/upload"]).To(Equal(4))
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
		Expect(resp.Records[0].Columns).To(Equal([]fvcommon.DeleteFilesColumnResult{
			{Column: "resume", Status: fvcommon.DeleteFilesStatusDeleted, RawStatus: "DELETED"},
		}))
	})

	It("sends interceptor headers", func() {
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { rc.AddHeader(fvcommon.RequestIdHeader, "mine") },
		})
		Expect(err).To(BeNil())
		Expect(fake.headers["/v2/query"].Get("x-request-id")).To(Equal("mine"))
	})

	It("reports a network failure with HTTP code 0 and no status", func() {
		fake.server.Close()
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal(string(fverrors.TRANSPORT_ERROR)))
		Expect(err.GetHttpStatusCode()).To(BeEmpty())
		Expect(errors.Unwrap(err)).ToNot(BeNil())
	})

	It("reports the client's Timeout as HTTP code 0 and keeps the deadline error", func() {
		slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(500 * time.Millisecond)
		}))
		DeferCleanup(slow.Close)
		timed := NewVaultController(fvcommon.VaultConfig{VaultId: "vault1", BaseVaultUrl: slow.URL, Timeout: 50 * time.Millisecond},
			ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}})
		_, err := timed.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal(string(fverrors.TRANSPORT_ERROR)))
		Expect(errors.Is(err, context.DeadlineExceeded)).To(BeTrue())
	})

	It("logs a warning when a query returns a full page of rows", func() {
		rows := make([]string, fvconsts.QUERY_RECORD_LIMIT)
		for i := range rows {
			rows[i] = `{"data":{"n":1}}`
		}
		fake.respond("/v2/query", http.StatusOK, `{"records":[`+strings.Join(rows, ",")+`]}`)
		resp, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT n FROM t"}, fvcommon.QueryOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records).To(HaveLen(fvconsts.QUERY_RECORD_LIMIT))
	})

	It("reports getTokens records of a failed call as failed unless they report a 2xx", func() {
		fake.respond("/v2/records/getTokens", http.StatusBadRequest, `{"records":[
			{"token":"t1","httpCode":200},{"token":"","httpCode":400},{"token":""}]}`)
		resp, err := v.GetTokens(ctx, fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{
			{Value: "a", TokenGroupName: "g"}, {Value: "b", TokenGroupName: "g"}, {Value: "c", TokenGroupName: "g"},
		}}, fvcommon.GetTokensOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records[0].Error).To(BeEmpty())
		Expect(resp.Records[1].Error).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
		Expect(resp.Records[1].RequestId).To(Equal("req-123"))
		Expect(resp.Records[2].HttpCode).To(Equal(http.StatusBadRequest))
		Expect(resp.Records[2].Error).To(Equal(fvconsts.UNKNOWN_RECORD_ERROR))
	})

	It("treats a blank 2xx body as an empty response", func() {
		fake.respond("/v2/query", http.StatusOK, "")
		resp, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).To(BeNil())
		Expect(resp.Records).To(BeEmpty())
		Expect(resp.RequestId).To(Equal("req-123"))

		fake.respond("/v2/records/getTokens", http.StatusOK, " \n")
		tokens, err := v.GetTokens(ctx, fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "v", TokenGroupName: "g"}}}, fvcommon.GetTokensOptions{})
		Expect(err).To(BeNil())
		Expect(tokens.Records).To(BeEmpty())
	})

	It("reports a 2xx body that isn't a JSON object with its real status and body", func() {
		for _, body := range []string{"<html>gateway</html>", "[1,2]", "null"} {
			fake.respond("/v2/files/delete", http.StatusOK, body)
			_, err := v.DeleteFiles(ctx, fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{
				{TableName: "docs", Columns: []string{"resume"}, SkyflowId: "id1"},
			}}, fvcommon.DeleteFilesOptions{})
			Expect(err).ToNot(BeNil(), body)
			Expect(err.GetHttpCode()).To(Equal("200"))
			Expect(err.GetMessage()).To(ContainSubstring("isn't valid JSON. Response body: " + body))
			Expect(err.GetRequestId()).To(Equal("req-123"))
		}
		fake.respond("/v2/query", http.StatusOK, strings.Repeat("x", 2000))
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err.GetMessage()).To(HaveSuffix("... (truncated)"))
	})

	It("fails the call with a SkyflowError when the interceptor panics", func() {
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { panic(errors.New("no tenant")) },
		})
		Expect(err).ToNot(BeNil())
		// There is no HTTP code ("0", empty status) and the message is the interceptor's own
		// when it has one.
		Expect(err.GetHttpCode()).To(Equal(string(fverrors.TRANSPORT_ERROR)))
		Expect(err.GetHttpStatusCode()).To(BeEmpty())
		Expect(err.GetMessage()).To(ContainSubstring("no tenant"))
		Expect(errors.Unwrap(err)).To(MatchError("no tenant"))
		Expect(fake.calls).To(BeEmpty())

		_, err = v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { panic("missing tenant") },
		})
		Expect(err.GetMessage()).To(ContainSubstring("missing tenant"))
		for _, recovered := range []interface{}{42, errors.New(""), ""} {
			_, err = v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
				Interceptor: func(rc *fvcommon.RequestContext) { panic(recovered) },
			})
			Expect(err.GetMessage()).To(ContainSubstring(fverrors.REQUEST_INTERCEPTOR_FAILED))
		}
		Expect(fake.calls).To(BeEmpty())

		own := fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "mine")
		_, err = v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { panic(own) },
		})
		Expect(err).To(BeIdenticalTo(own))

		_, err = v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{
			Interceptor: func(rc *fvcommon.RequestContext) { panic((*fverrors.SkyflowError)(nil)) },
		})
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal(string(fverrors.TRANSPORT_ERROR)))
		Expect(fake.calls).To(BeEmpty())
	})

	It("reports a context cancelled during the call as REQUEST_CANCELLED", func() {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		_, err := v.Query(cancelled, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal(string(fverrors.REQUEST_CANCELLED)))
		Expect(errors.Is(err, context.Canceled)).To(BeTrue())
	})

	It("truncates a long non-JSON error body", func() {
		fake.respond("/v2/query", http.StatusBadGateway, strings.Repeat("x", 2000))
		_, err := v.Query(ctx, fvcommon.QueryRequest{Query: "SELECT 1"}, fvcommon.QueryOptions{})
		Expect(err).ToNot(BeNil())
		Expect(err.GetHttpCode()).To(Equal("502"))
		Expect(err.GetMessage()).To(HaveSuffix("... (truncated)"))
		Expect(len(err.GetMessage())).To(BeNumerically("<", 600))
		Expect(err.GetRequestId()).To(Equal("req-123"))
	})
})

func intPtr(v int) *int { return &v }
