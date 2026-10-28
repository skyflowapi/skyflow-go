package validation

import (
	"fmt"
	"strings"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

func invalid(format string, args ...interface{}) *fverrors.SkyflowError {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, message)
}

func blank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func ValidateInsert(req fvcommon.InsertRequest) *fverrors.SkyflowError {
	name := fvconsts.METHOD_INSERT
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_RECORDS, name)
	}
	tableAtRequest := !blank(req.TableName)
	for i, record := range req.Records {
		if err := validateData(name, i, record.Data); err != nil {
			return err
		}
		tableAtRecord := !blank(record.TableName)
		if tableAtRequest && tableAtRecord {
			return invalid(fverrors.TABLE_IN_BOTH_PLACES, name)
		}
		if !tableAtRequest && !tableAtRecord {
			return invalid(fverrors.TABLE_MISSING, name)
		}
		if tableAtRequest && record.Upsert != nil {
			return invalid(fverrors.RECORD_UPSERT_NOT_ALLOWED)
		}
		if err := validateUpsert(record.Upsert); err != nil {
			return err
		}
	}
	if !tableAtRequest && req.Upsert != nil {
		return invalid(fverrors.REQUEST_UPSERT_NOT_ALLOWED)
	}
	return validateUpsert(req.Upsert)
}

func validateUpsert(upsert *fvcommon.UpsertOptions) *fverrors.SkyflowError {
	if upsert == nil {
		return nil
	}
	if len(upsert.UniqueColumns) == 0 {
		return invalid(fverrors.EMPTY_UNIQUE_COLUMNS)
	}
	for _, column := range upsert.UniqueColumns {
		if blank(column) {
			return invalid(fverrors.EMPTY_UNIQUE_COLUMNS)
		}
	}
	return validateUpdateType(fvconsts.METHOD_INSERT, upsert.UpdateType)
}

func validateUpdateType(name string, updateType fvcommon.UpdateType) *fverrors.SkyflowError {
	switch updateType {
	case "", fvcommon.UpdateTypeUpdate, fvcommon.UpdateTypeReplace:
		return nil
	}
	return invalid(fverrors.INVALID_UPDATE_TYPE, name, string(updateType))
}

func validateData(name string, index int, data map[string]interface{}) *fverrors.SkyflowError {
	if len(data) == 0 {
		return invalid(fverrors.EMPTY_DATA, name, index)
	}
	for key := range data {
		if blank(key) {
			return invalid(fverrors.EMPTY_KEY_IN_DATA, name, index)
		}
	}
	return nil
}

func ValidateUpdate(req fvcommon.UpdateRequest) *fverrors.SkyflowError {
	name := fvconsts.METHOD_UPDATE
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_RECORDS, name)
	}
	if err := validateUpdateType(name, req.UpdateType); err != nil {
		return err
	}
	tableAtRequest := !blank(req.TableName)
	for i, record := range req.Records {
		if blank(record.SkyflowId) {
			return invalid(fverrors.MISSING_SKYFLOW_ID, i)
		}
		if err := validateData(name, i, record.Data); err != nil {
			return err
		}
		tableAtRecord := !blank(record.TableName)
		if tableAtRequest && tableAtRecord {
			return invalid(fverrors.TABLE_IN_BOTH_PLACES, name)
		}
		if !tableAtRequest && !tableAtRecord {
			return invalid(fverrors.TABLE_MISSING, name)
		}
	}
	return nil
}

func ValidateGet(req fvcommon.GetRequest) *fverrors.SkyflowError {
	name := fvconsts.METHOD_GET
	if req.Limit < 0 || req.Offset < 0 {
		return invalid(fverrors.NEGATIVE_LIMIT_OFFSET)
	}
	if req.Records != nil {
		if req.TableName != "" || req.SkyflowIds != nil || req.UniqueValues != nil || req.Columns != nil ||
			req.ColumnRedactions != nil || req.Limit != 0 || req.Offset != 0 {
			return invalid(fverrors.GET_MODE_CONFLICT)
		}
		if len(req.Records) == 0 {
			return invalid(fverrors.EMPTY_RECORDS, name)
		}
		for _, record := range req.Records {
			if err := validateTableTarget(name, record.TableName, record.SkyflowIds, record.UniqueValues); err != nil {
				return err
			}
			if err := validateColumnRedactions(name, record.ColumnRedactions); err != nil {
				return err
			}
		}
		return nil
	}
	if err := validateTableTarget(name, req.TableName, req.SkyflowIds, req.UniqueValues); err != nil {
		return err
	}
	return validateColumnRedactions(name, req.ColumnRedactions)
}

func ValidateDelete(req fvcommon.DeleteRequest) *fverrors.SkyflowError {
	return validateTableTarget(fvconsts.METHOD_DELETE, req.TableName, req.SkyflowIds, req.UniqueValues)
}

// validateTableTarget checks a table name plus exactly one of IDs / unique values.
func validateTableTarget(name, table string, ids []string, uniqueValues []map[string]interface{}) *fverrors.SkyflowError {
	if blank(table) {
		return invalid(fverrors.EMPTY_TABLE_NAME, name)
	}
	if (len(ids) > 0) == (len(uniqueValues) > 0) {
		return invalid(fverrors.IDS_OR_UNIQUE_VALUES, name)
	}
	for _, id := range ids {
		if blank(id) {
			return invalid(fverrors.EMPTY_ID, name)
		}
	}
	return validateUniqueValues(name, uniqueValues)
}

func validateUniqueValues(name string, uniqueValues []map[string]interface{}) *fverrors.SkyflowError {
	for _, value := range uniqueValues {
		if len(value) == 0 {
			return invalid(fverrors.EMPTY_UNIQUE_VALUE, name)
		}
	}
	return nil
}

func validateColumnRedactions(name string, redactions []fvcommon.ColumnRedaction) *fverrors.SkyflowError {
	for _, redaction := range redactions {
		if blank(redaction.ColumnName) || blank(redaction.Redaction) {
			return invalid(fverrors.INVALID_COLUMN_REDACTION, name)
		}
	}
	return nil
}

func ValidateDetokenize(req fvcommon.DetokenizeRequest) *fverrors.SkyflowError {
	if len(req.Tokens) == 0 {
		return invalid(fverrors.EMPTY_TOKENS)
	}
	for _, token := range req.Tokens {
		if blank(token) {
			return invalid(fverrors.EMPTY_TOKEN)
		}
	}
	for _, redaction := range req.TokenGroupRedactions {
		if blank(redaction.TokenGroupName) || blank(redaction.Redaction) {
			return invalid(fverrors.INVALID_TOKEN_GROUP_REDACTION)
		}
	}
	return nil
}

func ValidateGetTokens(req fvcommon.GetTokensRequest) *fverrors.SkyflowError {
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_RECORDS, fvconsts.METHOD_GET_TOKENS)
	}
	for i, record := range req.Records {
		if record.Value == nil {
			return invalid(fverrors.MISSING_VALUE, i)
		}
		if blank(record.TokenGroupName) {
			return invalid(fverrors.MISSING_TOKEN_GROUP_NAME, i)
		}
	}
	return nil
}

func ValidateQuery(req fvcommon.QueryRequest) *fverrors.SkyflowError {
	if blank(req.Query) {
		return invalid(fverrors.EMPTY_QUERY)
	}
	return nil
}

func ValidateUploadFiles(req fvcommon.UploadFilesRequest) *fverrors.SkyflowError {
	name := fvconsts.METHOD_UPLOAD_FILES
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_RECORDS, name)
	}
	for i, record := range req.Records {
		if blank(record.TableName) {
			return invalid(fverrors.EMPTY_TABLE_NAME, name)
		}
		if len(record.Columns) == 0 {
			return invalid(fverrors.EMPTY_COLUMNS, name, i)
		}
		seen := make(map[string]bool, len(record.Columns))
		for _, column := range record.Columns {
			if blank(column.Column) {
				return invalid(fverrors.EMPTY_COLUMN_NAME, name, i)
			}
			if seen[column.Column] {
				return invalid(fverrors.DUPLICATE_COLUMN, name, column.Column, i)
			}
			seen[column.Column] = true
			sources := 0
			if column.FilePath != "" {
				sources++
			}
			if column.Base64 != "" {
				sources++
			}
			if column.FileObject != nil {
				sources++
			}
			if sources != 1 {
				return invalid(fverrors.FILE_SOURCE_COUNT, column.Column)
			}
			if column.Base64 != "" && blank(column.FileName) {
				return invalid(fverrors.BASE64_NEEDS_FILE_NAME, column.Column)
			}
		}
	}
	return nil
}

func ValidateDeleteFiles(req fvcommon.DeleteFilesRequest) *fverrors.SkyflowError {
	name := fvconsts.METHOD_DELETE_FILES
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_RECORDS, name)
	}
	if len(req.Records) > fvconsts.MAX_DELETE_FILES_RECORDS {
		return invalid(fverrors.DELETE_FILES_LIMIT, fvconsts.MAX_DELETE_FILES_RECORDS)
	}
	for i, record := range req.Records {
		if blank(record.TableName) {
			return invalid(fverrors.EMPTY_TABLE_NAME, name)
		}
		if len(record.Columns) == 0 {
			return invalid(fverrors.EMPTY_COLUMNS, name, i)
		}
		seen := make(map[string]bool, len(record.Columns))
		for _, column := range record.Columns {
			if blank(column) {
				return invalid(fverrors.EMPTY_COLUMN_NAME, name, i)
			}
			if seen[column] {
				return invalid(fverrors.DUPLICATE_COLUMN, name, column, i)
			}
			seen[column] = true
		}
		if blank(record.SkyflowId) == (len(record.UniqueValues) == 0) {
			return invalid(fverrors.DELETE_FILES_TARGET, i)
		}
		if err := validateUniqueValues(name, record.UniqueValues); err != nil {
			return err
		}
	}
	return nil
}

// ValidateCustomHeaders checks that only supported header keys are set, each with a value.
func ValidateCustomHeaders(headers map[fvcommon.CustomHeaderKey]string, tag string) *fverrors.SkyflowError {
	for key, value := range headers {
		switch key {
		case fvcommon.SkyflowAccountId, fvcommon.SkyflowAccountName, fvcommon.RequestIdHeader:
		default:
			return invalid(fverrors.INVALID_HEADER_KEY, tag, string(key))
		}
		if blank(value) {
			return invalid(fverrors.EMPTY_HEADER_VALUE, tag, string(key))
		}
	}
	return nil
}
