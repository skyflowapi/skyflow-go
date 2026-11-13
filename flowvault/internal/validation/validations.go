package validation

import (
	"fmt"
	"regexp"
	"strings"

	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
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
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_RECORDS)
	}
	tableAtRequest := !blank(req.TableName)
	for _, record := range req.Records {
		if err := validateData(record.Data, record.Tokens); err != nil {
			return err
		}
		tableAtRecord := !blank(record.TableName)
		if tableAtRequest && tableAtRecord {
			return invalid(fverrors.TABLE_IN_BOTH_PLACES)
		}
		if !tableAtRequest && !tableAtRecord {
			return invalid(fverrors.TABLE_MISSING)
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
	return validateUpdateType(upsert.UpdateType)
}

// validateUpdateType accepts UPDATE or REPLACE in any letter case.
func validateUpdateType(updateType fvcommon.UpdateType) *fverrors.SkyflowError {
	switch fvcommon.UpdateType(strings.ToUpper(string(updateType))) {
	case "", fvcommon.UpdateTypeUpdate, fvcommon.UpdateTypeReplace:
		return nil
	}
	return invalid(fverrors.INVALID_UPDATE_TYPE)
}

// validateData checks a record's data and tokens. No column name or
// value may be empty (nil or blank), and Tokens, when set, must not be empty.
func validateData(data, tokens map[string]interface{}) *fverrors.SkyflowError {
	if len(data) == 0 {
		return invalid(fverrors.EMPTY_DATA)
	}
	for key, value := range data {
		if blank(key) {
			return invalid(fverrors.EMPTY_KEY_IN_DATA)
		}
		if emptyValue(value) {
			return invalid(fverrors.EMPTY_VALUE_IN_DATA)
		}
	}
	if tokens == nil {
		return nil
	}
	if len(tokens) == 0 {
		return invalid(fverrors.EMPTY_RECORD_TOKENS)
	}
	for key, value := range tokens {
		if blank(key) {
			return invalid(fverrors.EMPTY_KEY_IN_TOKENS)
		}
		if emptyValue(value) {
			return invalid(fverrors.EMPTY_VALUE_IN_TOKENS)
		}
	}
	return nil
}

// emptyValue reports whether a data or token value is nil or a blank string.
func emptyValue(value interface{}) bool {
	if value == nil {
		return true
	}
	text, isString := value.(string)
	return isString && blank(text)
}

func ValidateUpdate(req fvcommon.UpdateRequest) *fverrors.SkyflowError {
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_RECORDS)
	}
	if err := validateUpdateType(req.UpdateType); err != nil {
		return err
	}
	tableAtRequest := !blank(req.TableName)
	for _, record := range req.Records {
		if blank(record.SkyflowId) {
			return invalid(fverrors.MISSING_SKYFLOW_ID)
		}
		if err := validateData(record.Data, record.Tokens); err != nil {
			return err
		}
		tableAtRecord := !blank(record.TableName)
		if tableAtRequest && tableAtRecord {
			return invalid(fverrors.TABLE_IN_BOTH_PLACES)
		}
		if !tableAtRequest && !tableAtRecord {
			return invalid(fverrors.TABLE_MISSING)
		}
	}
	return nil
}

func ValidateGet(req fvcommon.GetRequest) *fverrors.SkyflowError {
	if (req.Limit != nil && *req.Limit < 0) || (req.Offset != nil && *req.Offset < 0) {
		return invalid(fverrors.NEGATIVE_LIMIT_OFFSET)
	}
	// An empty Records slice counts as unset: the request is then a
	// single-table get. Limit and Offset aren't sent with Records, so setting them is refused
	// rather than silently ignored.
	if len(req.Records) > 0 {
		if req.TableName != "" || len(req.SkyflowIds) > 0 || len(req.UniqueValues) > 0 || len(req.Columns) > 0 ||
			len(req.ColumnRedactions) > 0 || req.Limit != nil || req.Offset != nil {
			return invalid(fverrors.GET_MODE_CONFLICT)
		}
		for _, record := range req.Records {
			if err := validateGetTarget(record.TableName, record.SkyflowIds, record.UniqueValues, record.Columns, record.ColumnRedactions); err != nil {
				return err
			}
		}
		return nil
	}
	return validateGetTarget(req.TableName, req.SkyflowIds, req.UniqueValues, req.Columns, req.ColumnRedactions)
}

// validateGetTarget checks one table's lookup: the table and IDs or unique values, then the
// columns to return and the column redactions.
func validateGetTarget(table string, ids []string, uniqueValues []map[string]interface{}, columns []string,
	redactions []fvcommon.ColumnRedactions) *fverrors.SkyflowError {
	if err := validateTableTarget(table, ids, uniqueValues); err != nil {
		return err
	}
	// A blank column name is refused. An empty Columns slice is the same
	// as an unset one: every column is returned.
	for _, column := range columns {
		if blank(column) {
			return invalid(fverrors.EMPTY_COLUMN_IN_COLUMNS)
		}
	}
	return validateColumnRedactions(redactions)
}

func ValidateDelete(req fvcommon.DeleteRequest) *fverrors.SkyflowError {
	return validateTableTarget(req.TableName, req.SkyflowIds, req.UniqueValues)
}

// validateTableTarget checks a table name plus exactly one of IDs / unique values.
func validateTableTarget(table string, ids []string, uniqueValues []map[string]interface{}) *fverrors.SkyflowError {
	if blank(table) {
		return invalid(fverrors.EMPTY_TABLE_NAME)
	}
	switch hasIds, hasUniqueValues := len(ids) > 0, len(uniqueValues) > 0; {
	case hasIds && hasUniqueValues:
		return invalid(fverrors.BOTH_IDS_AND_UNIQUE_VALUES)
	case !hasIds && !hasUniqueValues:
		return invalid(fverrors.IDS_OR_UNIQUE_VALUES)
	}
	for _, id := range ids {
		if blank(id) {
			return invalid(fverrors.EMPTY_ID)
		}
	}
	return validateUniqueValues(uniqueValues)
}

func validateUniqueValues(uniqueValues []map[string]interface{}) *fverrors.SkyflowError {
	for _, value := range uniqueValues {
		if len(value) == 0 {
			return invalid(fverrors.EMPTY_UNIQUE_VALUE)
		}
	}
	return nil
}

func validateColumnRedactions(redactions []fvcommon.ColumnRedactions) *fverrors.SkyflowError {
	for _, redaction := range redactions {
		if blank(redaction.ColumnName) {
			return invalid(fverrors.EMPTY_COLUMN_NAME_IN_COLUMN_REDACTION)
		}
		if blank(redaction.Redaction) {
			return invalid(fverrors.EMPTY_REDACTION_IN_COLUMN_REDACTION)
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
		if blank(redaction.TokenGroupName) {
			return invalid(fverrors.EMPTY_TOKEN_GROUP_NAME_IN_TOKEN_GROUP)
		}
		if blank(redaction.Redaction) {
			return invalid(fverrors.EMPTY_REDACTION_IN_TOKEN_GROUP)
		}
	}
	return nil
}

func ValidateGetTokens(req fvcommon.GetTokensRequest) *fverrors.SkyflowError {
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_GET_TOKENS_RECORDS)
	}
	for _, record := range req.Records {
		if text, isString := record.Value.(string); record.Value == nil || isString && blank(text) {
			return invalid(fverrors.MISSING_VALUE)
		}
		if blank(record.TokenGroupName) {
			return invalid(fverrors.MISSING_TOKEN_GROUP_NAME)
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

// presentButBlank reports a value that was set but holds only whitespace.
// Such a SkyflowId is an error rather than being taken as "not set".
func presentButBlank(value string) bool {
	return value != "" && blank(value)
}

func ValidateUploadFiles(req fvcommon.UploadFilesRequest) *fverrors.SkyflowError {
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_UPLOAD_FILES_RECORDS)
	}
	for _, record := range req.Records {
		if blank(record.TableName) {
			return invalid(fverrors.EMPTY_UPLOAD_FILES_TABLE_NAME)
		}
		if presentButBlank(record.SkyflowId) {
			return invalid(fverrors.EMPTY_UPLOAD_FILES_SKYFLOW_ID)
		}
		if len(record.Columns) == 0 {
			return invalid(fverrors.EMPTY_UPLOAD_FILES_COLUMNS)
		}
		seen := make(map[string]bool, len(record.Columns))
		for _, column := range record.Columns {
			if blank(column.Column) {
				return invalid(fverrors.EMPTY_UPLOAD_FILES_COLUMN_NAME)
			}
			if seen[column.Column] {
				return invalid(fverrors.DUPLICATE_UPLOAD_FILES_COLUMN)
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
			switch {
			case sources == 0:
				return invalid(fverrors.MISSING_FILE_SOURCE)
			case sources > 1:
				return invalid(fverrors.MULTIPLE_FILE_SOURCES)
			}
			if column.Base64 != "" && blank(column.FileName) {
				return invalid(fverrors.BASE64_NEEDS_FILE_NAME)
			}
		}
	}
	return nil
}

func ValidateDeleteFiles(req fvcommon.DeleteFilesRequest) *fverrors.SkyflowError {
	if len(req.Records) == 0 {
		return invalid(fverrors.EMPTY_DELETE_FILES_RECORDS)
	}
	for _, record := range req.Records {
		if blank(record.TableName) {
			return invalid(fverrors.EMPTY_DELETE_FILES_TABLE_NAME)
		}
		if len(record.Columns) == 0 {
			return invalid(fverrors.EMPTY_DELETE_FILES_COLUMNS)
		}
		seen := make(map[string]bool, len(record.Columns))
		for _, column := range record.Columns {
			if blank(column) {
				return invalid(fverrors.EMPTY_DELETE_FILES_COLUMN_NAME)
			}
			if seen[column] {
				return invalid(fverrors.DUPLICATE_DELETE_FILES_COLUMN)
			}
			seen[column] = true
		}
		if presentButBlank(record.SkyflowId) {
			return invalid(fverrors.EMPTY_DELETE_FILES_SKYFLOW_ID)
		}
		if (record.SkyflowId != "") == (len(record.UniqueValues) > 0) {
			return invalid(fverrors.DELETE_FILES_TARGET)
		}
		if err := validateUniqueValues(record.UniqueValues); err != nil {
			return err
		}
	}
	return nil
}

// ValidateCustomHeaders checks that only supported header keys are set, each with a value.
func ValidateCustomHeaders(headers map[fvcommon.CustomHeaderKey]string, tag string) *fverrors.SkyflowError {
	return fvhelpers.Profile().ValidateCustomHeaders(headers, tag)
}

var (
	// apiKeyPattern is the API key format.
	apiKeyPattern = regexp.MustCompile(`^sky-[a-zA-Z0-9]{5}-[a-fA-F0-9]{32}$`)
	// contextKeyPattern is the rule for a key in a context map.
	contextKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// ValidateCredentials checks credentials when they are added: exactly
// one source, which must not be blank; an API key in the sky-xxxxx-<32 hex> format; roles; and
// a Context that is a non-blank string or a non-empty map with valid keys. The checks shared
// with the v2 SDK (one source, roles) come from the common validator; the stricter ones are
// flowvault's, so v2's released behaviour is unchanged.
func ValidateCredentials(credentials fvcommon.Credentials) *fverrors.SkyflowError {
	sources := 0
	for _, value := range []string{credentials.Path, credentials.CredentialsString, credentials.Token, credentials.ApiKey} {
		if value != "" {
			sources++
		}
	}
	if sources == 1 {
		switch {
		case credentials.Path != "" && blank(credentials.Path):
			return invalid(fverrors.EMPTY_CREDENTIAL_FILE_PATH)
		case credentials.CredentialsString != "" && blank(credentials.CredentialsString):
			return invalid(fverrors.EMPTY_CREDENTIALS_STRING)
		case credentials.Token != "" && blank(credentials.Token):
			return invalid(fverrors.EMPTY_TOKEN_CREDENTIAL)
		case credentials.ApiKey != "" && blank(credentials.ApiKey):
			return invalid(fverrors.EMPTY_API_KEY)
		case credentials.ApiKey != "" && !apiKeyPattern.MatchString(credentials.ApiKey):
			return invalid(fverrors.INVALID_API_KEY)
		}
	}
	if err := fvhelpers.ConfigValidator().ValidateCredentials(credentials); err != nil {
		return err
	}
	return validateContext(credentials.Context)
}

// validateContext checks a credentials Context: unset, a non-blank string, or a non-empty
// map[string]interface{} whose keys are letters, digits and underscores.
func validateContext(context interface{}) *fverrors.SkyflowError {
	switch value := context.(type) {
	case nil:
		return nil
	case string:
		if blank(value) {
			return invalid(fverrors.EMPTY_CONTEXT)
		}
		return nil
	case map[string]interface{}:
		if len(value) == 0 {
			return invalid(fverrors.EMPTY_CONTEXT)
		}
		for key := range value {
			if !contextKeyPattern.MatchString(key) {
				return invalid(fverrors.INVALID_CONTEXT_MAP_KEY, key)
			}
		}
		return nil
	default:
		return invalid(fverrors.INVALID_CONTEXT_TYPE)
	}
}
