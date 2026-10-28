package validation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/skyflowapi/skyflow-go/flowvault/internal/validation"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Validation Suite")
}

func expectError(err *fverrors.SkyflowError, message string) {
	if message == "" {
		Expect(err).To(BeNil())
		return
	}
	Expect(err).ToNot(BeNil())
	Expect(err.GetMessage()).To(ContainSubstring(message))
}

var data = map[string]interface{}{"name": "john"}

var _ = Describe("ValidateInsert", func() {
	DescribeTable("table and upsert placement",
		func(req fvcommon.InsertRequest, message string) { expectError(ValidateInsert(req), message) },
		Entry("request-level table", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data}}}, ""),
		Entry("record-level table and upsert", fvcommon.InsertRequest{Records: []fvcommon.InsertRequestRecord{{Data: data, TableName: "t", Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"name"}}}}}, ""),
		Entry("no records", fvcommon.InsertRequest{TableName: "t"}, "Records must not be empty"),
		Entry("empty data", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{}}}, "Data must not be empty"),
		Entry("empty column name", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{" ": 1}}}}, "empty column name"),
		Entry("table in both places", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, TableName: "t"}}}, "not both"),
		Entry("table missing on a record", fvcommon.InsertRequest{Records: []fvcommon.InsertRequestRecord{{Data: data, TableName: "t"}, {Data: data}}}, "Table name is missing"),
		Entry("record upsert with request table", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}}}}}, "Set upsert on the request"),
		Entry("request upsert with record tables", fvcommon.InsertRequest{Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}}, Records: []fvcommon.InsertRequestRecord{{Data: data, TableName: "t"}}}, "Set upsert on each record"),
		Entry("upsert without unique columns", fvcommon.InsertRequest{TableName: "t", Upsert: &fvcommon.UpsertOptions{}, Records: []fvcommon.InsertRequestRecord{{Data: data}}}, "unique columns"),
		Entry("invalid upsert update type", fvcommon.InsertRequest{TableName: "t", Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}, UpdateType: "MERGE"}, Records: []fvcommon.InsertRequestRecord{{Data: data}}}, "Invalid update type"),
	)
})

var _ = Describe("ValidateUpdate", func() {
	DescribeTable("requests",
		func(req fvcommon.UpdateRequest, message string) { expectError(ValidateUpdate(req), message) },
		Entry("valid", fvcommon.UpdateRequest{TableName: "t", UpdateType: fvcommon.UpdateTypeReplace, Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, ""),
		Entry("missing skyflow ID", fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{Data: data}}}, "Skyflow ID is missing"),
		Entry("table in both places", fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data, TableName: "t"}}}, "not both"),
		Entry("table missing", fvcommon.UpdateRequest{Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, "Table name is missing"),
		Entry("invalid update type", fvcommon.UpdateRequest{TableName: "t", UpdateType: "MERGE", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, "Invalid update type"),
	)
})

var _ = Describe("ValidateGet", func() {
	DescribeTable("requests",
		func(req fvcommon.GetRequest, message string) { expectError(ValidateGet(req), message) },
		Entry("ids", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}}, ""),
		Entry("multi-table", fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{{TableName: "t", UniqueValues: []map[string]interface{}{data}}}}, ""),
		Entry("both modes", fvcommon.GetRequest{TableName: "t", Records: []fvcommon.GetRequestRecord{{TableName: "t", SkyflowIds: []string{"id"}}}}, "not both"),
		Entry("empty multi-table records", fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{}}, "Records must not be empty"),
		Entry("ids and unique values", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, UniqueValues: []map[string]interface{}{data}}, "exactly one"),
		Entry("neither ids nor unique values", fvcommon.GetRequest{TableName: "t"}, "exactly one"),
		Entry("empty unique value", fvcommon.GetRequest{TableName: "t", UniqueValues: []map[string]interface{}{{}}}, "empty entries"),
		Entry("negative limit", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, Limit: -1}, "must not be negative"),
		Entry("bad redaction", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, ColumnRedactions: []fvcommon.ColumnRedaction{{ColumnName: "a"}}}, "Column redactions"),
	)
})

var _ = Describe("Other validators", func() {
	It("validates Delete", func() {
		expectError(ValidateDelete(fvcommon.DeleteRequest{SkyflowIds: []string{"id"}}), "Table name must not be empty")
		expectError(ValidateDelete(fvcommon.DeleteRequest{TableName: "t", SkyflowIds: []string{""}}), "empty values")
		expectError(ValidateDelete(fvcommon.DeleteRequest{TableName: "t", UniqueValues: []map[string]interface{}{data}}), "")
	})

	It("validates Detokenize", func() {
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{}), "Tokens must not be empty")
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{Tokens: []string{" "}}), "empty values")
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{Tokens: []string{"tok"}, TokenGroupRedactions: []fvcommon.TokenGroupRedaction{{TokenGroupName: "g"}}}), "Token group redactions")
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{Tokens: []string{"tok"}}), "")
	})

	It("validates GetTokens and Query", func() {
		expectError(ValidateGetTokens(fvcommon.GetTokensRequest{}), "Records must not be empty")
		expectError(ValidateGetTokens(fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{TokenGroupName: "g"}}}), "Value is missing")
		expectError(ValidateGetTokens(fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "v"}}}), "Token group name is missing")
		expectError(ValidateQuery(fvcommon.QueryRequest{Query: " "}), "Query must not be empty")
		expectError(ValidateQuery(fvcommon.QueryRequest{Query: "SELECT 1"}), "")
	})

	It("validates UploadFiles", func() {
		column := func(c fvcommon.UploadFilesRequestColumn) fvcommon.UploadFilesRequest {
			return fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{c}}}}
		}
		expectError(ValidateUploadFiles(fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t"}}}), "Columns must not be empty")
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{FilePath: "a"})), "Column name must not be empty")
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c"})), "exactly one of FilePath")
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: "a", Base64: "b"})), "exactly one of FilePath")
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "b"})), "FileName is required")
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "b", FileName: "a.txt"})), "")
		file := fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "b", FileName: "a.txt"}
		expectError(ValidateUploadFiles(fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{file, file}}}}),
			`Column "c" is listed more than once for record at index 0`)
	})

	It("validates DeleteFiles", func() {
		record := func(r fvcommon.DeleteFilesRequestRecord) fvcommon.DeleteFilesRequest {
			return fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{r}}
		}
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}})), "exactly one of SkyflowId")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}, SkyflowId: "id", UniqueValues: []map[string]interface{}{data}})), "exactly one of SkyflowId")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{""}, SkyflowId: "id"})), "Column name must not be empty")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c", "d", "c"}, SkyflowId: "id"})), `Column "c" is listed more than once for record at index 0`)
		expectError(ValidateDeleteFiles(fvcommon.DeleteFilesRequest{Records: make([]fvcommon.DeleteFilesRequestRecord, 20001)}), "At most 20000")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}, SkyflowId: "id"})), "")
	})

	It("validates custom headers", func() {
		expectError(ValidateCustomHeaders(map[fvcommon.CustomHeaderKey]string{"x-other": "v"}, "Insert"), "is not supported")
		expectError(ValidateCustomHeaders(map[fvcommon.CustomHeaderKey]string{fvcommon.RequestIdHeader: " "}, "Insert"), "must have a value")
		expectError(ValidateCustomHeaders(map[fvcommon.CustomHeaderKey]string{fvcommon.SkyflowAccountId: "acc"}, "Insert"), "")
	})
})
