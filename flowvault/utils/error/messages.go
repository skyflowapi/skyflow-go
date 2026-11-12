package errors

import internal "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"

const (
	VAULT_ID_ALREADY_IN_CONFIG_LIST string = internal.SDK_PREFIX + " Validation error. VaultId is present in an existing config. Specify a new vaultId in config."
	VAULT_ID_NOT_IN_CONFIG_LIST     string = internal.SDK_PREFIX + " Validation error. VaultId is missing from the config. Specify the vaultIds from configs."
	EMPTY_VAULT_CONFIG              string = internal.SDK_PREFIX + " Validation error. No vault configurations available"
	EMPTY_VAULT_ID                  string = internal.SDK_PREFIX + " Initialization failed. Invalid vault ID. Vault ID must not be empty."
)

// Request validation messages, worded as in the Java SDK (flowvault ErrorMessage and common
// BaseErrorMessage). Messages containing %s / %d are formatted with fmt.Sprintf.
const (
	// Insert / update
	EMPTY_RECORDS              string = internal.SDK_PREFIX + " Validation error. 'records' can't be empty. Specify records."
	TABLE_IN_BOTH_PLACES       string = internal.SDK_PREFIX + " Validation error. Table name cannot be specified at both the request and record levels. Please specify the table name in only one place."
	TABLE_MISSING              string = internal.SDK_PREFIX + " Validation error. Table name is missing. Table name should be specified at one place either at the request level or record level. Please specify the table name at one place."
	RECORD_UPSERT_NOT_ALLOWED  string = internal.SDK_PREFIX + " Validation error. Table name should be present at each record level when upsert is present at record level."
	REQUEST_UPSERT_NOT_ALLOWED string = internal.SDK_PREFIX + " Validation error. Upsert should be present at each record level when table name is present at record level."
	EMPTY_UNIQUE_COLUMNS       string = internal.SDK_PREFIX + " Validation error. Upsert column values can't be empty. Specify at least one upsert column."
	INVALID_UPDATE_TYPE        string = internal.SDK_PREFIX + " Validation error. Invalid upsert updateType. Specify either 'UPDATE' or 'REPLACE'."
	EMPTY_DATA                 string = internal.SDK_PREFIX + " Validation error. 'data' can't be empty. Specify data."
	EMPTY_KEY_IN_DATA          string = internal.SDK_PREFIX + " Validation error. Invalid key in data in records. Specify a valid key."
	MISSING_SKYFLOW_ID         string = internal.SDK_PREFIX + " Validation error. 'skyflowId' can't be empty in the record. Specify a valid skyflow ID."
	// As in the Java SDK, a data value can't be nil or blank, and Tokens, when set, can't be
	// empty or hold a blank column name or value.
	EMPTY_VALUE_IN_DATA   string = internal.SDK_PREFIX + " Validation error. Invalid value in values. Specify a valid value."
	EMPTY_RECORD_TOKENS   string = internal.SDK_PREFIX + " Validation error. The 'tokens' field is empty. Specify tokens for one or more fields."
	EMPTY_KEY_IN_TOKENS   string = internal.SDK_PREFIX + " Validation error. Invalid key tokens. Specify a valid key."
	EMPTY_VALUE_IN_TOKENS string = internal.SDK_PREFIX + " Validation error. Invalid value in tokens. Specify a valid value."

	// Get / delete
	GET_MODE_CONFLICT                     string = internal.SDK_PREFIX + " Validation error. Both single-table lookup fields ('table', 'ids', 'fields', 'uniqueValues', 'columnRedactions') and 'records' can't be specified. Either specify single-table fields or 'records'."
	NEGATIVE_LIMIT_OFFSET                 string = internal.SDK_PREFIX + " Validation error. 'limit' and 'offset' can't be negative. Specify a non-negative limit and offset."
	EMPTY_TABLE_NAME                      string = internal.SDK_PREFIX + " Validation error. 'table' can't be empty. Specify a table."
	BOTH_IDS_AND_UNIQUE_VALUES            string = internal.SDK_PREFIX + " Validation error. Both Skyflow IDs and unique values can't be specified. Either specify Skyflow IDs or unique values."
	IDS_OR_UNIQUE_VALUES                  string = internal.SDK_PREFIX + " Validation error. 'ids' or 'uniqueValues' key is missing from the payload. Specify ids or uniqueValues in payload."
	EMPTY_ID                              string = internal.SDK_PREFIX + " Validation error. Invalid id in 'ids'. Specify a valid id."
	EMPTY_UNIQUE_VALUE                    string = internal.SDK_PREFIX + " Validation error. Invalid unique value in 'uniqueValues'. Specify a valid unique value."
	EMPTY_COLUMN_IN_COLUMNS               string = internal.SDK_PREFIX + " Validation error. Invalid field in 'fields'. Specify a valid field."
	EMPTY_COLUMN_NAME_IN_COLUMN_REDACTION string = internal.SDK_PREFIX + " Validation error. Column name can not be null or empty in column redaction. Specify a valid column name."
	EMPTY_REDACTION_IN_COLUMN_REDACTION   string = internal.SDK_PREFIX + " Validation error. Redaction can not be null or empty in column redaction. Specify a valid redaction."

	// Detokenize
	EMPTY_TOKENS                          string = internal.SDK_PREFIX + " Validation error. Invalid data tokens. Specify at least one data token."
	EMPTY_TOKEN                           string = internal.SDK_PREFIX + " Validation error. Invalid data tokens. Specify a valid data token."
	EMPTY_TOKEN_GROUP_NAME_IN_TOKEN_GROUP string = internal.SDK_PREFIX + " Validation error. TokenGroupName in TokenGroupRedactions is null or empty. Specify a valid tokenGroupName."
	EMPTY_REDACTION_IN_TOKEN_GROUP        string = internal.SDK_PREFIX + " Validation error. Redaction in TokenGroupRedactions is null or empty. Specify a valid redaction."

	// Get tokens
	EMPTY_GET_TOKENS_RECORDS string = internal.SDK_PREFIX + " Validation error. 'records' can't be empty. Specify at least one GetTokensRequestRecord."
	MISSING_VALUE            string = internal.SDK_PREFIX + " Validation error. 'value' in GetTokensRequestRecord is null or empty. Specify a valid value."
	MISSING_TOKEN_GROUP_NAME string = internal.SDK_PREFIX + " Validation error. 'tokenGroupName' in GetTokensRequestRecord is null or empty. Specify a valid token group name."

	// Query
	EMPTY_QUERY string = internal.SDK_PREFIX + " Validation error. 'query' can't be empty. Specify a query"

	// Upload files
	EMPTY_UPLOAD_FILES_RECORDS     string = internal.SDK_PREFIX + " Validation error. 'records' can't be empty. Specify at least one UploadFilesRequestRecord."
	EMPTY_UPLOAD_FILES_TABLE_NAME  string = internal.SDK_PREFIX + " Validation error. 'tableName' in UploadFilesRequestRecord is null or empty. Specify a valid table name."
	EMPTY_UPLOAD_FILES_SKYFLOW_ID  string = internal.SDK_PREFIX + " Validation error. 'skyflowId' in UploadFilesRequestRecord is empty. Specify a valid skyflow ID, or omit it to create a new record."
	EMPTY_UPLOAD_FILES_COLUMNS     string = internal.SDK_PREFIX + " Validation error. 'columns' in UploadFilesRequestRecord can't be empty. Specify at least one UploadFilesRequestColumn."
	EMPTY_UPLOAD_FILES_COLUMN_NAME string = internal.SDK_PREFIX + " Validation error. 'column' in UploadFilesRequestColumn is null or empty. Specify a valid column name."
	DUPLICATE_UPLOAD_FILES_COLUMN  string = internal.SDK_PREFIX + " Validation error. 'columns' in UploadFilesRequestRecord contains a duplicate column name. Specify each column only once."
	MISSING_FILE_SOURCE            string = internal.SDK_PREFIX + " Validation error. UploadFilesRequestColumn has no file. Provide one of filePath, base64, or fileObject."
	MULTIPLE_FILE_SOURCES          string = internal.SDK_PREFIX + " Validation error. UploadFilesRequestColumn has more than one file. Provide exactly one of filePath, base64, or fileObject."
	INVALID_FILE_PATH              string = internal.SDK_PREFIX + " Validation error. No readable file at 'filePath' in UploadFilesRequestColumn. Specify the path of an existing file."
	INVALID_BASE64                 string = internal.SDK_PREFIX + " Validation error. 'base64' in UploadFilesRequestColumn is not a valid base64 string. Specify valid base64-encoded content."
	BASE64_NEEDS_FILE_NAME         string = internal.SDK_PREFIX + " Validation error. 'fileName' is required in UploadFilesRequestColumn when 'base64' is used. Specify a file name."
	INVALID_FILE_OBJECT            string = internal.SDK_PREFIX + " Validation error. 'fileObject' in UploadFilesRequestColumn is not an existing file. Specify a valid file object."

	// Delete files
	EMPTY_DELETE_FILES_RECORDS     string = internal.SDK_PREFIX + " Validation error. 'records' can't be empty. Specify at least one DeleteFilesRequestRecord."
	EMPTY_DELETE_FILES_TABLE_NAME  string = internal.SDK_PREFIX + " Validation error. 'tableName' in DeleteFilesRequestRecord is null or empty. Specify a valid table name."
	EMPTY_DELETE_FILES_COLUMNS     string = internal.SDK_PREFIX + " Validation error. 'columns' in DeleteFilesRequestRecord can't be empty. Specify at least one file column."
	EMPTY_DELETE_FILES_COLUMN_NAME string = internal.SDK_PREFIX + " Validation error. 'columns' in DeleteFilesRequestRecord contains a null or empty column name. Specify valid column names."
	DUPLICATE_DELETE_FILES_COLUMN  string = internal.SDK_PREFIX + " Validation error. 'columns' in DeleteFilesRequestRecord contains a duplicate column name. Specify each column only once."
	EMPTY_DELETE_FILES_SKYFLOW_ID  string = internal.SDK_PREFIX + " Validation error. 'skyflowId' in DeleteFilesRequestRecord is empty. Specify a valid skyflow ID, or use 'uniqueValues' instead."
	DELETE_FILES_TARGET            string = internal.SDK_PREFIX + " Validation error. DeleteFilesRequestRecord must set exactly one of 'skyflowId' or 'uniqueValues'. Specify one of them."

	// Credentials, as in the Java SDK (BaseErrorMessage). The type names in INVALID_CONTEXT_TYPE
	// are Go's.
	EMPTY_CREDENTIAL_FILE_PATH string = internal.SDK_PREFIX + " Initialization failed. Invalid credentials. Credentials file path must not be empty."
	EMPTY_CREDENTIALS_STRING   string = internal.SDK_PREFIX + " Initialization failed. Invalid credentials. Credentials string must not be empty."
	EMPTY_TOKEN_CREDENTIAL     string = internal.SDK_PREFIX + " Initialization failed. Invalid credentials. Token must not be empty."
	EMPTY_API_KEY              string = internal.SDK_PREFIX + " Initialization failed. Invalid credentials. Api key must not be empty."
	INVALID_API_KEY            string = internal.SDK_PREFIX + " Initialization failed. Invalid credentials. Specify valid api key."
	EMPTY_CONTEXT              string = internal.SDK_PREFIX + " Initialization failed. Invalid context. Specify a valid context."
	INVALID_CONTEXT_TYPE       string = internal.SDK_PREFIX + " Initialization failed. Invalid context type. Specify context as a string or a map[string]interface{}."
	INVALID_CONTEXT_MAP_KEY    string = internal.SDK_PREFIX + " Initialization failed. Invalid key '%s' in context map. Keys must contain only alphanumeric characters and underscores."

	// Credentials and calls
	EMPTY_CREDENTIALS            string = internal.SDK_PREFIX + " Validation error. Invalid credentials. Credentials must not be empty."
	BEARER_TOKEN_EXPIRED         string = internal.SDK_PREFIX + " Validation error. Bearer token is invalid or expired. Please provide a valid bearer token."
	NIL_CONTEXT                  string = internal.SDK_PREFIX + " Validation error. Context is nil. Pass a context, e.g. context.Background()."
	BEARER_TOKEN_CANCELLED       string = internal.SDK_PREFIX + " Request cancelled. The context ended before a bearer token was available."
	MULTIPLE_VAULT_CONFIGS       string = internal.SDK_PREFIX + " Validation error. Multiple vault configurations found. Specify a vaultId."
	INVALID_HEADER_KEY           string = internal.SDK_PREFIX + " Validation error. %s: header key %q is not supported. Use x-skyflow-account-id, x-skyflow-account-name or x-request-id."
	EMPTY_HEADER_VALUE           string = internal.SDK_PREFIX + " Validation error. %s: header %q must have a value."
	REQUEST_INTERCEPTOR_FAILED   string = internal.SDK_PREFIX + " Request interceptor threw an exception. The request was not sent."
	UNPARSEABLE_SUCCESS_RESPONSE string = internal.SDK_PREFIX + " The vault returned a response that isn't valid JSON. Response body: %s"
)
