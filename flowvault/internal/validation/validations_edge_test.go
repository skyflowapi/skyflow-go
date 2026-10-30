package validation_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/skyflowapi/skyflow-go/flowvault/internal/validation"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

var _ = Describe("Validation edge cases", func() {
	It("rejects an invalid record-level upsert and a blank unique column", func() {
		expectError(ValidateInsert(fvcommon.InsertRequest{Records: []fvcommon.InsertRequestRecord{
			{Data: data, TableName: "t", Upsert: &fvcommon.UpsertOptions{}},
		}}), "unique columns")
		expectError(ValidateInsert(fvcommon.InsertRequest{TableName: "t", Upsert: &fvcommon.UpsertOptions{UniqueColumns: []string{"a", " "}},
			Records: []fvcommon.InsertRequestRecord{{Data: data}}}), "unique columns")
	})

	It("rejects an update with no records or empty data", func() {
		expectError(ValidateUpdate(fvcommon.UpdateRequest{TableName: "t"}), "Records must not be empty")
		expectError(ValidateUpdate(fvcommon.UpdateRequest{TableName: "t", Records: []fvcommon.UpdateRequestRecord{{SkyflowId: "id"}}}), "Data must not be empty")
	})

	It("checks every record of a multi-table get", func() {
		expectError(ValidateGet(fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{{SkyflowIds: []string{"id"}}}}), "Table name must not be empty")
		expectError(ValidateGet(fvcommon.GetRequest{Records: []fvcommon.GetRequestRecord{{
			TableName: "t", SkyflowIds: []string{"id"}, ColumnRedactions: []fvcommon.ColumnRedaction{{Redaction: "MASKED"}},
		}}}), "Column redactions")
	})

	It("rejects uploads with no records, no table or two sources including a file object", func() {
		expectError(ValidateUploadFiles(fvcommon.UploadFilesRequest{}), "Records must not be empty")
		expectError(ValidateUploadFiles(fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{
			Columns: []fvcommon.UploadFilesRequestColumn{{Column: "c", Base64: "b", FileName: "a.txt"}},
		}}}), "Table name must not be empty")

		file, err := os.CreateTemp(GinkgoT().TempDir(), "upload")
		Expect(err).To(BeNil())
		defer file.Close()
		column := func(c fvcommon.UploadFilesRequestColumn) fvcommon.UploadFilesRequest {
			return fvcommon.UploadFilesRequest{Records: []fvcommon.UploadFilesRequestRecord{{TableName: "t", Columns: []fvcommon.UploadFilesRequestColumn{c}}}}
		}
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", FileObject: file})), "")
		expectError(ValidateUploadFiles(column(fvcommon.UploadFilesRequestColumn{Column: "c", FileObject: file, FilePath: "a"})), "exactly one of FilePath")
	})

	It("rejects file deletes with no records, no table, no columns or an empty unique value", func() {
		record := func(r fvcommon.DeleteFilesRequestRecord) fvcommon.DeleteFilesRequest {
			return fvcommon.DeleteFilesRequest{Records: []fvcommon.DeleteFilesRequestRecord{r}}
		}
		expectError(ValidateDeleteFiles(fvcommon.DeleteFilesRequest{}), "Records must not be empty")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{Columns: []string{"c"}, SkyflowId: "id"})), "Table name must not be empty")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", SkyflowId: "id"})), "Columns must not be empty")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"},
			UniqueValues: []map[string]interface{}{{}}})), "empty entries")
		expectError(ValidateDeleteFiles(record(fvcommon.DeleteFilesRequestRecord{TableName: "t", Columns: []string{"c"},
			UniqueValues: []map[string]interface{}{data}})), "")
	})
})
