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
		Entry("no records", fvcommon.InsertRequest{TableName: "t"}, fverrors.EMPTY_RECORDS),
		Entry("empty data", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{}}}, fverrors.EMPTY_DATA),
		Entry("empty column name", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{" ": 1}}}}, fverrors.EMPTY_KEY_IN_DATA),
		// No nil or blank value, and Tokens, when set, non-empty with
		// non-blank names and values. Non-string values such as 0 and false are fine.
		Entry("nil value", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{"name": nil}}}}, fverrors.EMPTY_VALUE_IN_DATA),
		Entry("blank value", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{"name": "  "}}}}, fverrors.EMPTY_VALUE_IN_DATA),
		Entry("zero and false values", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: map[string]interface{}{"n": 0, "b": false}}}}, ""),
		Entry("empty tokens", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, Tokens: map[string]interface{}{}}}}, fverrors.EMPTY_RECORD_TOKENS),
		Entry("blank token column", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, Tokens: map[string]interface{}{"": "tok"}}}}, fverrors.EMPTY_KEY_IN_TOKENS),
		Entry("blank token value", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, Tokens: map[string]interface{}{"name": ""}}}}, fverrors.EMPTY_VALUE_IN_TOKENS),
		Entry("valid tokens", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, Tokens: map[string]interface{}{"name": "tok"}}}}, ""),
		Entry("table in both places", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, TableName: "t"}}}, fverrors.TABLE_IN_BOTH_PLACES),
		Entry("table missing on a record", fvcommon.InsertRequest{Records: []fvcommon.InsertRequestRecord{{Data: data, TableName: "t"}, {Data: data}}}, fverrors.TABLE_MISSING),
		Entry("record upsert with request table", fvcommon.InsertRequest{TableName: "t", Records: []fvcommon.InsertRequestRecord{{Data: data, Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}}}}}, fverrors.RECORD_UPSERT_NOT_ALLOWED),
		Entry("request upsert with record tables", fvcommon.InsertRequest{Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}}, Records: []fvcommon.InsertRequestRecord{{Data: data, TableName: "t"}}}, fverrors.REQUEST_UPSERT_NOT_ALLOWED),
		Entry("upsert without unique columns", fvcommon.InsertRequest{TableName: "t", Upsert: &fvcommon.UpsertOptions{}, Records: []fvcommon.InsertRequestRecord{{Data: data}}}, fverrors.EMPTY_UNIQUE_COLUMNS),
		Entry("upsert update type in any case", fvcommon.InsertRequest{TableName: "t", Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}, UpdateType: "Update"}, Records: []fvcommon.InsertRequestRecord{{Data: data}}}, ""),
		Entry("invalid upsert update type", fvcommon.InsertRequest{TableName: "t", Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a"}, UpdateType: "MERGE"}, Records: []fvcommon.InsertRequestRecord{{Data: data}}}, fverrors.INVALID_UPDATE_TYPE),
	)
})

var _ = Describe("ValidateUpdate", func() {
	DescribeTable("requests",
		func(req fvcommon.UpdateRequest, message string) { expectError(ValidateUpdate(req), message) },
		Entry("valid", fvcommon.UpdateRequest{TableName: "t", UpdateType: fvcommon.UpdateTypeReplace, Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, ""),
		Entry("missing skyflow ID", fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{Data: data}}}, fverrors.MISSING_SKYFLOW_ID),
		Entry("table in both places", fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data, TableName: "t"}}}, fverrors.TABLE_IN_BOTH_PLACES),
		Entry("table missing", fvcommon.UpdateRequest{Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, fverrors.TABLE_MISSING),
		Entry("update type in any case", fvcommon.UpdateRequest{TableName: "t", UpdateType: "replace", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, ""),
		Entry("invalid update type", fvcommon.UpdateRequest{TableName: "t", UpdateType: "MERGE", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: data}}}, fverrors.INVALID_UPDATE_TYPE),
	)
})

var _ = Describe("ValidateGet", func() {
	DescribeTable("requests",
		func(req fvcommon.GetRequest, message string) { expectError(ValidateGet(req), message) },
		Entry("ids", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}}, ""),
		Entry("multi-table", fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{{TableName: "t", UniqueValues: []map[string]interface{}{data}}}}, ""),
		Entry("both modes", fvcommon.GetRequest{TableName: "t", Records: []fvcommon.GetRequestRecord{{TableName: "t", SkyflowIds: []string{"id"}}}}, fverrors.GET_MODE_CONFLICT),
		Entry("zero offset with records", fvcommon.GetRequest{Offset: intPtr(0), Records: []fvcommon.GetRequestRecord{{TableName: "t", SkyflowIds: []string{"id"}}}}, fverrors.GET_MODE_CONFLICT),
		Entry("empty Records counts as unset", fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{}}, fverrors.EMPTY_TABLE_NAME),
		Entry("empty Records with a single-table get", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, Records: []fvcommon.GetRequestRecord{}}, ""),
		Entry("empty single-table fields with records", fvcommon.GetRequest{SkyflowIds: []string{}, Columns: []string{}, Records: []fvcommon.GetRequestRecord{{TableName: "t", SkyflowIds: []string{"id"}}}}, ""),
		Entry("ids and unique values", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, UniqueValues: []map[string]interface{}{data}}, fverrors.BOTH_IDS_AND_UNIQUE_VALUES),
		Entry("neither ids nor unique values", fvcommon.GetRequest{TableName: "t"}, fverrors.IDS_OR_UNIQUE_VALUES),
		Entry("empty unique value", fvcommon.GetRequest{TableName: "t", UniqueValues: []map[string]interface{}{{}}}, fverrors.EMPTY_UNIQUE_VALUE),
		Entry("negative limit", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, Limit: intPtr(-1)}, fverrors.NEGATIVE_LIMIT_OFFSET),
		Entry("blank column", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, Columns: []string{"name", " "}}, fverrors.EMPTY_COLUMN_IN_COLUMNS),
		Entry("blank column on a record", fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{{TableName: "t", SkyflowIds: []string{"id"}, Columns: []string{""}}}}, fverrors.EMPTY_COLUMN_IN_COLUMNS),
		Entry("empty columns are unset", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, Columns: []string{}}, ""),
		Entry("bad redaction", fvcommon.GetRequest{TableName: "t", SkyflowIds: []string{"id"}, ColumnRedactions: []fvcommon.ColumnRedactions{{ColumnName: "a"}}}, fverrors.EMPTY_REDACTION_IN_COLUMN_REDACTION),
	)
})

var _ = Describe("Other validators", func() {
	It("validates Delete", func() {
		expectError(ValidateDelete(fvcommon.DeleteRequest{SkyflowIds: []string{"id"}}), fverrors.EMPTY_TABLE_NAME)
		expectError(ValidateDelete(fvcommon.DeleteRequest{TableName: "t", SkyflowIds: []string{""}}), fverrors.EMPTY_ID)
		expectError(ValidateDelete(fvcommon.DeleteRequest{TableName: "t", UniqueValues: []map[string]interface{}{data}}), "")
	})

	It("validates Detokenize", func() {
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{}), fverrors.EMPTY_TOKENS)
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{Tokens: []string{" "}}), fverrors.EMPTY_TOKEN)
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{Tokens: []string{"tok"}, TokenGroupRedactions: []fvcommon.TokenGroupRedactions{{TokenGroupName: "g"}}}), fverrors.EMPTY_REDACTION_IN_TOKEN_GROUP)
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{Tokens: []string{"tok"}, TokenGroupRedactions: []fvcommon.TokenGroupRedactions{{Redaction: "MASKED"}}}), fverrors.EMPTY_TOKEN_GROUP_NAME_IN_TOKEN_GROUP)
		expectError(ValidateDetokenize(fvcommon.DetokenizeRequest{Tokens: []string{"tok"}}), "")
	})

	It("validates GetTokens and Query", func() {
		expectError(ValidateGetTokens(fvcommon.GetTokensRequest{}), fverrors.EMPTY_GET_TOKENS_RECORDS)
		expectError(ValidateGetTokens(fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{TokenGroupName: "g"}}}), fverrors.MISSING_VALUE)
		expectError(ValidateGetTokens(fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "  ", TokenGroupName: "g"}}}), fverrors.MISSING_VALUE)
		Expect(ValidateGetTokens(fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: 0, TokenGroupName: "g"}}})).To(BeNil())
		expectError(ValidateGetTokens(fvcommon.GetTokensRequest{Records: []fvcommon.GetTokensRequestRecord{{Value: "v"}}}), fverrors.MISSING_TOKEN_GROUP_NAME)
		expectError(ValidateQuery(fvcommon.QueryRequest{Query: " "}), fverrors.EMPTY_QUERY)
		expectError(ValidateQuery(fvcommon.QueryRequest{Query: "SELECT 1"}), "")
	})

	It("validates UploadFiles", func() {
		column := func(c fvcommon.UploadFilesRequestColumn) fvcommon.UploadFilesRequest {
			return fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{c}}}}
		}
		expectError(ValidateUploadFiles(fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t"}}}), fverrors.EMPTY_UPLOAD_FILES_COLUMNS)
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{FilePath: "a"})), fverrors.EMPTY_UPLOAD_FILES_COLUMN_NAME)
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c"})), fverrors.MISSING_FILE_SOURCE)
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", FilePath: "a", Base64: "b"})), fverrors.MULTIPLE_FILE_SOURCES)
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "b"})), fverrors.BASE64_NEEDS_FILE_NAME)
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "b", FileName: "a.txt"})), "")
		// A SkyflowId that is set but blank is an error, not a new record.
		expectError(ValidateUploadFiles(fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t", SkyflowId: " ",
			Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "b", FileName: "a.txt"}}}}}), fverrors.EMPTY_UPLOAD_FILES_SKYFLOW_ID)
		file := fvcommon.UploadFilesRequestColumn{Column: "c", Base64: "b", FileName: "a.txt"}
		expectError(ValidateUploadFiles(fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{file, file}}}}),
			fverrors.DUPLICATE_UPLOAD_FILES_COLUMN)
	})

	It("validates DeleteFiles", func() {
		record := func(r fvcommon.DeleteFilesRequestRecord) fvcommon.DeleteFilesRequest {
			return fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{r}}
		}
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}})), fverrors.DELETE_FILES_TARGET)
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}, SkyflowId: "id", UniqueValues: []map[string]interface{}{data}})), fverrors.DELETE_FILES_TARGET)
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{""}, SkyflowId: "id"})), fverrors.EMPTY_DELETE_FILES_COLUMN_NAME)
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c", "d", "c"}, SkyflowId: "id"})), fverrors.DUPLICATE_DELETE_FILES_COLUMN)
		// No record limit: the vault enforces its own.
		many := make([]fvcommon.DeleteFilesRequestRecord, 20001)
		for i := range many {
			many[i] = fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}, SkyflowId: "id"}
		}
		expectError(ValidateDeleteFiles(fvcommon.DeleteFilesRequest{Records: many}), "")
		// A SkyflowId that is set but blank is an error, not "unset".
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}, SkyflowId: "  "})), fverrors.EMPTY_DELETE_FILES_SKYFLOW_ID)
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}, SkyflowId: " ",
			UniqueValues: []map[string]interface{}{data}})), fverrors.EMPTY_DELETE_FILES_SKYFLOW_ID)
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"}, SkyflowId: "id"})), "")
	})

	It("validates custom headers", func() {
		expectError(ValidateCustomHeaders(map[fvcommon.CustomHeaderKey]string{"x-other": "v"}, "Insert"), "is not supported")
		expectError(ValidateCustomHeaders(map[fvcommon.CustomHeaderKey]string{fvcommon.RequestIdHeader: " "}, "Insert"), "must have a value")
		expectError(ValidateCustomHeaders(map[fvcommon.CustomHeaderKey]string{fvcommon.SkyflowAccountId: "acc"}, "Insert"), "")
	})
})

var _ = Describe("Update data checks", func() {
	It("checks update data values and tokens", func() {
		err := ValidateUpdate(fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: map[string]interface{}{"a": nil}}}})
		Expect(err.GetMessage()).To(Equal(fverrors.EMPTY_VALUE_IN_DATA))
		err = ValidateUpdate(fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id", Data: map[string]interface{}{"a": 1}, Tokens: map[string]interface{}{}}}})
		Expect(err.GetMessage()).To(Equal(fverrors.EMPTY_RECORD_TOKENS))
	})
})

func intPtr(v int) *int { return &v }

var _ = Describe("ValidateCredentials", func() {
	const apiKey = "sky-ab1cd-0123456789abcdef0123456789abcdef"

	DescribeTable("checks the source, API key, roles and context",
		func(credentials fvcommon.Credentials, message string) {
			expectError(ValidateCredentials(credentials), message)
		},
		Entry("api key", fvcommon.Credentials{ApiKey: apiKey}, ""),
		Entry("token with roles and a string context", fvcommon.Credentials{Token: "tok", Roles: []string{"r1"}, Context: "user_1"}, ""),
		Entry("path with a map context", fvcommon.Credentials{Path: "creds.json", Context: map[string]interface{}{"user_id": "u1"}}, ""),
		Entry("no source", fvcommon.Credentials{}, "Specify any one from"),
		Entry("two sources", fvcommon.Credentials{Token: "tok", ApiKey: apiKey}, "Specify only one from"),
		Entry("blank path", fvcommon.Credentials{Path: "  "}, fverrors.EMPTY_CREDENTIAL_FILE_PATH),
		Entry("blank credentials string", fvcommon.Credentials{CredentialsString: " "}, fverrors.EMPTY_CREDENTIALS_STRING),
		Entry("blank token", fvcommon.Credentials{Token: "\t"}, fverrors.EMPTY_TOKEN_CREDENTIAL),
		Entry("blank api key", fvcommon.Credentials{ApiKey: " "}, fverrors.EMPTY_API_KEY),
		Entry("api key of the right length but the wrong format", fvcommon.Credentials{ApiKey: "xx-sky-1-0123456789abcdef0123456789abcdef"}, fverrors.INVALID_API_KEY),
		Entry("api key with non-hex characters", fvcommon.Credentials{ApiKey: "sky-ab1cd-0123456789abcdef0123456789abcdeg"}, fverrors.INVALID_API_KEY),
		Entry("empty roles", fvcommon.Credentials{Token: "tok", Roles: []string{}}, "Specify at least one role"),
		Entry("blank context", fvcommon.Credentials{Token: "tok", Context: " "}, fverrors.EMPTY_CONTEXT),
		Entry("empty context map", fvcommon.Credentials{Token: "tok", Context: map[string]interface{}{}}, fverrors.EMPTY_CONTEXT),
		Entry("bad context map key", fvcommon.Credentials{Token: "tok", Context: map[string]interface{}{"user-id": 1}}, "Invalid key 'user-id' in context map"),
		Entry("bool context", fvcommon.Credentials{Token: "tok", Context: true}, fverrors.INVALID_CONTEXT_TYPE),
		Entry("number context", fvcommon.Credentials{Token: "tok", Context: 42}, fverrors.INVALID_CONTEXT_TYPE),
		Entry("map[string]string context", fvcommon.Credentials{Token: "tok", Context: map[string]string{"a": "b"}}, fverrors.INVALID_CONTEXT_TYPE),
	)
})
