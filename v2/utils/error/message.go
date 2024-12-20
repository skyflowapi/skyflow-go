package errors

import . "github.com/skyflowapi/skyflow-go/v2/internal/constants"

// TO DO
const (
	// config
	VAULT_ID_ALREADY_IN_CONFIG_LIST        string = SDK_PREFIX + " Validation error. VaultId is present in an existing config. Specify a new vaultId in config."
	VAULT_ID_NOT_IN_CONFIG_LIST            string = SDK_PREFIX + " Validation error. VaultId is missing from the config. Specify the vaultIds from configs."
	CONNECTION_ID_ALREADY_IN_CONFIG_LIST   string = SDK_PREFIX + " Validation error. ConnectionId is present in an existing config. Specify a connectionId in config."
	CONNECTION_ID_NOT_IN_CONFIG_LIST       string = SDK_PREFIX + " Validation error. ConnectionId is missing from the config. Specify the connectionIds from configs."
	EMPTY_CREDENTIALS                      string = SDK_PREFIX + " Validation error. Invalid credentials. Specify a valid credentials."
	EMPTY_VAULT_CONFIG                     string = SDK_PREFIX + " Validation error. No vault configurations available"
	INVALID_VAULT_ID                       string = SDK_PREFIX + " Initialization failed. Invalid vault ID. Specify a valid vault ID."
	EMPTY_VAULT_ID                         string = SDK_PREFIX + " Initialization failed. Invalid vault ID. Vault ID must not be empty."
	INVALID_CLUSTER_ID                     string = SDK_PREFIX + " Initialization failed. Invalid cluster ID. Specify cluster ID."
	EMPTY_CLUSTER_ID                       string = SDK_PREFIX + " Initialization failed. Invalid cluster ID. Specify a valid cluster ID."
	EMPTY_CONNECTION_CONFIG                string = SDK_PREFIX + " Validation error. No connection configurations available"
	EMPTY_CONNECTION_ID                    string = SDK_PREFIX + " Initialization failed. Invalid connection ID. Connection ID must not be empty."
	INVALID_CONNECTION_URL                 string = SDK_PREFIX + " Initialization failed. Invalid connection URL. Specify a valid connection URL."
	EMPTY_CONNECTION_URL                   string = SDK_PREFIX + " Initialization failed. Invalid connection URL. Connection URL must not be empty."
	INVALID_CONNECTION_URL_FORMAT          string = SDK_PREFIX + " Initialization failed. Connection URL is not a valid URL. Specify a valid connection URL."
	TOKEN_EXPIRED                          string = SDK_PREFIX + " Validation error. Token provided is either invalid or has expired. Specify a valid token."
	MULTIPLE_TOKEN_GENERATION_MEANS_PASSED string = SDK_PREFIX + " Initialization failed. Invalid credentials. Specify only one from 'path', 'credentialsString', 'token' or 'apiKey'."
	NO_TOKEN_GENERATION_MEANS_PASSED       string = SDK_PREFIX + " Initialization failed. Invalid credentials. Specify any one from 'path', 'credentialsString', 'token' or 'apiKey'."
	EMPTY_CREDENTIAL_FILE_PATH             string = SDK_PREFIX + " Initialization failed. Invalid credentials. Credentials file path must not be empty."
	EMPTY_CREDENTIALS_STRING               string = SDK_PREFIX + " Initialization failed. Invalid credentials. Credentials string must not be empty."
	EMPTY_TOKEN                            string = SDK_PREFIX + " Initialization failed. Invalid credentials. Token must not be empty."
	EMPTY_API_KEY                          string = SDK_PREFIX + " Initialization failed. Invalid credentials. Api key must not be empty."
	INVALID_API_KEY                        string = SDK_PREFIX + " Initialization failed. Invalid credentials. Specify valid api key."
	EMPTY_ROLES                            string = SDK_PREFIX + " Initialization failed. Invalid roles. Specify at least one role."
	EMPTY_ROLE_IN_ROLES                    string = SDK_PREFIX + " Initialization failed. Invalid role. Specify a valid role."
	EMPTY_CONTEXT                          string = SDK_PREFIX + " Initialization failed. Invalid context. Specify a valid context."

	FILE_NOT_FOUND                  string = SDK_PREFIX + " Initialization failed. Credential file not found at %s. Verify the file path."
	FILE_INVALID_JSON               string = SDK_PREFIX + " Initialization failed. File at %s is not in valid JSON format. Verify the file contents."
	CREDENTIALS_STRING_INVALID_JSON string = SDK_PREFIX + " Initialization failed. Credentials string is not in valid JSON format. Verify the credentials string contents."
	INVALID_CREDENTIALS             string = SDK_PREFIX + " Initialization failed. Invalid credentials provided. Specify valid credentials."
	MISSING_PRIVATE_KEY             string = SDK_PREFIX + " Initialization failed. Unable to read private key in credentials. Verify your private key."
	MISSING_CLIENT_ID               string = SDK_PREFIX + " Initialization failed. Unable to read client ID in credentials. Verify your client ID."
	MISSING_KEY_ID                  string = SDK_PREFIX + " Initialization failed. Unable to read key ID in credentials. Verify your key ID."
	MISSING_TOKEN_URI               string = SDK_PREFIX + " Initialization failed. Unable to read token URI in credentials. Verify your token URI."
	INVALID_TOKEN_URI               string = SDK_PREFIX + " Initialization failed. Token URI in not a valid URL in credentials. Verify your token URI."
	JWT_INVALID_FORMAT              string = SDK_PREFIX + " Initialization failed. Invalid private key format. Verify your credentials."
	INVALID_ALGORITHM               string = SDK_PREFIX + " Initialization failed. Invalid algorithm to parse private key. Specify valid algorithm."
	INVALID_KEY_SPEC                string = SDK_PREFIX + " Initialization failed. Unable to parse RSA private key. Verify your credentials."
	JWT_DECODE_ERROR                string = SDK_PREFIX + " Validation error. Invalid access token. Verify your credentials."
	MISSING_ACCESS_TOKEN            string = SDK_PREFIX + " Validation error. Access token not present in the response from bearer token generation. Verify your credentials."
	MISSING_TOKEN_TYPE              string = SDK_PREFIX + " Validation error. Token type not present in the response from bearer token generation. Verify your credentials."

	TABLE_KEY_ERROR                                   string = SDK_PREFIX + " Validation error. 'table' key is missing from the payload. Specify a 'table' key."
	EMPTY_TABLE                                       string = SDK_PREFIX + " Validation error. 'table' can't be empty. Specify a table."
	VALUES_KEY_ERROR                                  string = SDK_PREFIX + " Validation error. 'values' key is missing from the payload. Specify a 'values' key."
	EMPTY_VALUES                                      string = SDK_PREFIX + " Validation error. 'values' can't be empty. Specify values."
	EMPTY_KEY_IN_VALUES                               string = SDK_PREFIX + " Validation error. Invalid key in values. Specify a valid key."
	EMPTY_VALUE_IN_VALUES                             string = SDK_PREFIX + " Validation error. Invalid value in values. Specify a valid value."
	TOKENS_KEY_ERROR                                  string = SDK_PREFIX + " Validation error. 'tokens' key is missing from the payload. Specify a 'tokens' key."
	EMPTY_TOKENS                                      string = SDK_PREFIX + " Validation error. The 'tokens' field is empty. Specify tokens for one or more fields."
	EMPTY_KEY_IN_TOKENS                               string = SDK_PREFIX + " Validation error. Invalid key tokens. Specify a valid key."
	EMPTY_VALUE_IN_TOKENS                             string = SDK_PREFIX + " Validation error. Invalid value in tokens. Specify a valid value."
	EMPTY_UPSERT                                      string = SDK_PREFIX + " Validation error. 'upsert' key can't be empty. Specify an upsert column."
	HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT              string = SDK_PREFIX + " Validation error. 'homogenous' is not supported with 'upsert'. Specify either 'homogenous' or 'upsert'."
	TOKENS_PASSED_FOR_BYOT_DISABLE                    string = SDK_PREFIX + " Validation error. 'TokenMode' wasn't specified. Set 'TokenMode' to 'ENABLE' to insert tokens."
	NO_TOKENS_WITH_BYOT                               string = SDK_PREFIX + " Validation error. Tokens weren't specified for records while 'TokenMode' was %s. Specify tokens."
	MISMATCH_OF_FIELDS_AND_TOKENS                     string = SDK_PREFIX + " Validation error. 'fields' and 'tokens' have different columns names. Verify that 'fields' and 'tokens' columns match."
	INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT string = SDK_PREFIX + " Validation error. 'TokenMode' is set to 'ENABLE_STRICT' or 'ENABLE', but some fields are missing tokens. Specify tokens for all fields."
	BATCH_INSERT_PARTIAL_SUCCESS                      string = SDK_PREFIX + " Insert operation completed with partial success."
	BATCH_INSERT_FAILURE                              string = SDK_PREFIX + " Insert operation failed."
	MISMATCH_OF_FIELDS_AND_VALUES                     string = SDK_PREFIX + " Validation error. 'fields' and 'values' have different columns names. Verify that 'fields' and 'values' columns match."
	MISMATCH_OF_FIELDS_AND_OPTIONS                    string = SDK_PREFIX + " Validation error. 'fields' and 'options' have different columns names. Verify that 'fields' and 'options' columns match."
	INVALID_OPTION                                    string = SDK_PREFIX + " Validation error. Invalid options were specified in the payload. Specify valid options."

	EMPTY_PARAMETERS      string = SDK_PREFIX + " Validation error. Parameters must not be empty. Specify valid parameters."
	EMPTY_PARAMETER_NAME  string = SDK_PREFIX + " Validation error. Parameter name is missing in parameters. Specify a valid name."
	EMPTY_PARAMETER_VALUE string = SDK_PREFIX + " Validation error. Parameter value is missing for name. Specify a valid value."

	INVALID_TABLE     string = SDK_PREFIX + " Validation error. Invalid table. Specify a valid table."
	EMPTY_COLUMNS     string = SDK_PREFIX + " Validation error. Columns can't be empty. Specify valid columns."
	EMPTY_COLUMN_NAME string = SDK_PREFIX + " Validation error. Invalid column name. Specify valid column names."

	EMPTY_ROWS        string = SDK_PREFIX + " Validation error. Rows can't be empty. Specify valid rows."
	EMPTY_ROW_KEY     string = SDK_PREFIX + " Validation error. Rows have invalid keys. Specify valid keys."
	EMPTY_ROW_VALUE   string = SDK_PREFIX + " Validation error. Rows have invalid values. Specify valid values."
	INVALID_DATA_TYPE string = SDK_PREFIX + " Validation error. Invalid data type specified for column '%s'. Specify a valid data type."

	MULTIPLE_ERRORS          string = SDK_PREFIX + " Multiple validation errors found. Refer to individual errors for more details."
	GENERAL_VALIDATION_ERROR string = SDK_PREFIX + " 	Validation error. Please check your input and try again."

	REQUEST_ID_MISSING string = SDK_PREFIX + " Request failed. Request ID is missing in the response."
	SERVER_ERROR       string = SDK_PREFIX + " Server error. Please try again later."
	INVALID_RESPONSE   string = SDK_PREFIX + " Error while parsing server response. Please verify the response."
	TIMEOUT_ERROR      string = SDK_PREFIX + " Request timed out. Please try again later."

	OPERATION_NOT_SUPPORTED       string = SDK_PREFIX + " Operation failed. The operation is not supported for the specified parameters."
	INVALID_AUTHENTICATION_METHOD string = SDK_PREFIX + " Authentication error. Invalid authentication method specified. Use a supported method."

	UNAUTHORIZED_ACCESS     string = SDK_PREFIX + " Authorization error. You do not have permission to perform this operation."
	FORBIDDEN_OPERATION     string = SDK_PREFIX + " Authorization error. The operation is forbidden for the current credentials."
	RESOURCE_NOT_FOUND      string = SDK_PREFIX + " Resource error. The requested resource could not be found. Please verify the resource."
	RESOURCE_CONFLICT       string = SDK_PREFIX + " Resource error. Conflict detected while accessing the resource. Please resolve the conflict and try again."
	RESOURCE_LIMIT_EXCEEDED string = SDK_PREFIX + " Resource error. The resource limit has been exceeded. Please contact support."

	INTERNAL_SDK_ERROR                    string = SDK_PREFIX + " Internal SDK error. Please contact support with the error details."
	FEATURE_NOT_IMPLEMENTED               string = SDK_PREFIX + " Feature error. This feature is not implemented in the current version. Please upgrade to a supported version."
	EMPTY_REQUEST_HEADER                  string = SDK_PREFIX + " Validation error. Request headers are empty. Specify valid request headers."
	INVALID_REQUEST_HEADERS               string = SDK_PREFIX + "Validation error. Request headers aren't valid. Specify valid request headers."
	INVALID_QUERY_PARAM                   string = SDK_PREFIX + " Validation error. Query parameters aren't valid. Specify valid query parameters."
	EMPTY_QUERY_PARAM                     string = SDK_PREFIX + " Validation error. Query parameters are empty. Specify valid query parameters."
	INVALID_REQUEST_BODY                  string = SDK_PREFIX + " Validation error. Invalid request body. Specify the request body as an object."
	EMPTY_REQUEST_BODY                    string = SDK_PREFIX + " Validation error. Request body can't be empty. Specify a valid request body."
	TOKENS_REQUIRED                       string = SDK_PREFIX + " Invalid %s request. Tokens are required."
	EMPTY_TOKENS_DETOKENIZE               string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify at least one data token."
	INVALID_DATA_TOKENS                   string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify valid data tokens."
	EMPTY_TOKEN_IN_DATA_TOKEN             string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify a valid data token."
	IDS_KEY_ERROR                         string = SDK_PREFIX + " Validation error. 'ids' key is missing from the payload. Specify an 'ids' key."
	EMPTY_IDS                             string = SDK_PREFIX + " Validation error. 'ids' can't be empty. Specify at least one id."
	EMPTY_ID_IN_IDS                       string = SDK_PREFIX + " Validation error. Invalid id in 'ids'. Specify a valid id."
	EMPTY_FIELDS                          string = SDK_PREFIX + " Validation error. Fields are empty in get payload. Specify at least one field."
	EMPTY_FIELD_IN_FIELDS                 string = SDK_PREFIX + " Validation error. Invalid field in 'fields'. Specify a valid field."
	REDACTION_KEY_ERROR                   string = SDK_PREFIX + " Validation error. 'redaction' key is missing from the payload. Specify a 'redaction' key."
	REDACTION_WITH_TOKENS_NOT_SUPPORTED   string = SDK_PREFIX + " Validation error. 'redaction' can't be used when 'returnTokens' is specified. Remove 'redaction' from payload if 'returnTokens' is specified."
	TOKENS_GET_COLUMN_NOT_SUPPORTED       string = SDK_PREFIX + " Validation error. Column name and/or column values can't be used when 'returnTokens' is specified. Remove unique column values or 'returnTokens' from the payload."
	EMPTY_OFFSET                          string = SDK_PREFIX + " Validation error. 'offset' can't be empty. Specify an offset."
	EMPTY_LIMIT                           string = SDK_PREFIX + " Validation error. 'limit' can't be empty. Specify a limit."
	UNIQUE_COLUMN_OR_IDS_KEY_ERROR        string = SDK_PREFIX + " Validation error. 'ids' or 'columnName' key is missing from the payload. Specify the ids or unique 'columnName' in payload."
	BOTH_IDS_AND_COLUMN_DETAILS_SPECIFIED string = SDK_PREFIX + " Validation error. Both Skyflow IDs and column details can't be specified. Either specify Skyflow IDs or unique column details."
	COLUMN_NAME_KEY_ERROR                 string = SDK_PREFIX + " Validation error. 'columnName' isn't specified whereas 'columnValues' are specified. Either add 'columnName' or remove 'columnValues'."
	EMPTY_COLUMN_NAME_GET                 string = SDK_PREFIX + " Validation error. 'columnName' can't be empty. Specify a column name."
	COLUMN_VALUES_KEY_ERROR_GET           string = SDK_PREFIX + " Validation error. 'columnValues' aren't specified whereas 'columnName' is specified. Either add 'columnValues' or remove 'columnName'."
	EMPTY_COLUMN_VALUES                   string = SDK_PREFIX + " Validation error. 'columnValues' can't be empty. Specify at least one column value"
	EMPTY_VALUE_IN_COLUMN_VALUES          string = SDK_PREFIX + " Validation error. Invalid value in column values. Specify a valid column value."
	TOKEN_KEY_ERROR                       string = SDK_PREFIX + " Validation error. 'token' key is missing from the payload. Specify a 'token' key."
	PARTIAL_SUCCESS                       string = SDK_PREFIX + " Validation error. Check 'SkyflowError.data' for details."
	QUERY_KEY_ERROR                       string = SDK_PREFIX + " Validation error. 'query' key is missing from the payload. Specify a 'query' key."
	EMPTY_QUERY                           string = SDK_PREFIX + " Validation error. 'query' can't be empty."
	COLUMN_VALUES_KEY_ERROR_TOKENIZE      string = SDK_PREFIX + " Validation error. 'columnValues' key is missing from the payload. Specify a 'columnValues' key."
	VALUES_IS_REQUIRED_TOKENIZE           string = SDK_PREFIX + " Validation error. Invalid tokenize request. Values are required."
	EMPTY_COLUMN_GROUP_IN_COLUMN_VALUES   string = SDK_PREFIX + " Validation error. Invalid tokenize request. Column group can not be null or empty in column values at index %s."

	// tokenize
	MISSING_VALUES_IN_TOKENIZE       string = SDK_PREFIX + " Validation error. Values cannot be empty in tokenize request. Specify valid values."
	INVALID_VALUES_TYPE_IN_TOKENIZE  string = SDK_PREFIX + " Validation error. Invalid values type in tokenize request. Specify valid values of type array."
	EMPTY_VALUES_IN_TOKENIZE         string = SDK_PREFIX + " Validation error. Values array cannot be empty. Specify value's in tokenize request."
	EMPTY_DATA_IN_TOKENIZE           string = SDK_PREFIX + " Validation error. Data cannot be empty in tokenize request. Specify a valid data at index %s."
	INVALID_DATA_IN_TOKENIZE         string = SDK_PREFIX + " Validation error. Invalid Data. Specify a valid data at index %s."
	INVALID_COLUMN_GROUP_IN_TOKENIZE string = SDK_PREFIX + " Validation error. Invalid type of column group passed in tokenize request. Column group must be of type string at index %s."
	INVALID_VALUE_IN_TOKENIZE        string = SDK_PREFIX + " Validation error. Invalid type of value passed in tokenize request. Value must be of type string at index %s."
	INVALID_TOKENIZE_REQUEST         string = SDK_PREFIX + " Validation error. Invalid tokenize request. Specify a valid tokenize request."
	EMPTY_COLUMN_GROUP_IN_TOKENIZE   string = SDK_PREFIX + " Validation error. Column group cannot be empty in tokenize request. Specify a valid column group at index %s."
	EMPTY_VALUE_IN_TOKENIZE          string = SDK_PREFIX + " Validation error. Value cannot be empty in tokenize request. Specify a valid value at index %s."

	// update
	EMPTY_ID_IN_UPDATE string = SDK_PREFIX + " Validation error. 'id' can't be empty. Specify an id."
	// Error messages
	CONFIG_MISSING                            string = SDK_PREFIX + " Initialization failed. Skyflow config cannot be empty. Specify a valid skyflow config."
	INVALID_SKYFLOW_CONFIG                    string = SDK_PREFIX + " Initialization failed. Invalid skyflow config. Vaults configs key missing in skyflow config."
	INVALID_TYPE_FOR_CONFIG                   string = SDK_PREFIX + " Initialization failed. Invalid %s config. Specify a valid %s config."
	EMPTY_VAULT_ID_VALIDATION                 string = SDK_PREFIX + " Validation error. Invalid vault ID. Specify a valid vault Id."
	INVALID_TOKEN                             string = SDK_PREFIX + " Validation error. Invalid token. Specify a valid token."
	INVALID_ENV                               string = SDK_PREFIX + " Initialization failed. Invalid env. Specify a valid env for vault with vaultId %s."
	INVALID_LOG_LEVEL                         string = SDK_PREFIX + " Initialization failed. Invalid log level. Specify a valid log level."
	INVALID_CREDENTIAL_FILE_PATH              string = SDK_PREFIX + " Initialization failed. Invalid credentials. Expected file path to be a string."
	INVALID_FILE_PATH                         string = SDK_PREFIX + " Initialization failed. Invalid skyflow credentials. Expected file path to exist."
	INVALID_PARSED_CREDENTIALS_STRING         string = SDK_PREFIX + " Initialization failed. Invalid skyflow credentials. Specify a valid credentials string."
	INVALID_BEARER_TOKEN                      string = SDK_PREFIX + " Initialization failed. Invalid skyflow credentials. Specify a valid token."
	INVALID_FILE_PATH_WITH_ID                 string = SDK_PREFIX + " Initialization failed. Invalid credentials. Expected file path to exist for %s with %s %s."
	INVALID_API_KEY_WITH_ID                   string = SDK_PREFIX + " Initialization failed. Invalid credentials. Specify a valid api key for %s with %s %s."
	INVALID_PARSED_CREDENTIALS_STRING_WITH_ID string = SDK_PREFIX + " Initialization failed. Invalid credentials. Specify a valid credentials string for %s with %s %s."
	INVALID_BEARER_TOKEN_WITH_ID              string = SDK_PREFIX + " Initialization failed. Invalid credentials. Specify a valid token for %s with %s %s."
	// Validation errors
	EMPTY_CONNECTION_ID_VALIDATION string = SDK_PREFIX + " Validation error. Invalid connection ID. Specify a valid connection Id."
	INVALID_CONNECTION_ID          string = SDK_PREFIX + " Initialization failed. Invalid connection ID. Specify connection Id as a string."

	VAULT_ID_EXITS_IN_CONFIG_LIST      string = SDK_PREFIX + " Validation error. %s1 already exists in the config list. Specify a new vaultId."
	CONNECTION_ID_EXITS_IN_CONFIG_LIST string = SDK_PREFIX + " Validation error. %s1 already exists in the config list. Specify a new vaultId."

	CREDENTIALS_WITH_NO_VALID_KEY       string = SDK_PREFIX + " Validation error. Invalid credentials. Credentials must include one of the following: { apiKey, token, credentials, path }."
	MULTIPLE_CREDENTIALS_PASSED         string = SDK_PREFIX + " Validation error. Multiple credentials provided. Specify only one of the following: { apiKey, token, credentials, path }."
	INVALID_CREDENTIALS_WITH_ID         string = SDK_PREFIX + " Validation error. Invalid credentials. Credentials must include one of the following: { apiKey, token, credentials, path } for %s1 with %s2 %s3."
	MULTIPLE_CREDENTIALS_PASSED_WITH_ID string = SDK_PREFIX + " Validation error. Invalid credentials. Specify only one of the following: { apiKey, token, credentials, path } for %s1 with %s2 %s3."

	INVALID_JSON_FILE string = SDK_PREFIX + " Validation error. File at %s1 is not in valid JSON format. Verify the file contents."

	INVALID_CREDENTIALS_STRING string = SDK_PREFIX + " Validation error. Invalid credentials. Specify credentials as a string."
	INVALID_ROLES_KEY_TYPE     string = SDK_PREFIX + " Validation error. Invalid roles. Specify roles as an array."

	INVALID_JSON_FORMAT string = SDK_PREFIX + " Validation error. Credentials are not in valid JSON format. Verify the credentials."

	EMPTY_DATA_TOKENS     string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify valid data tokens."
	DATA_TOKEN_KEY_TYPE   string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify data token as a string array."
	TIME_TO_LIVE_KEY_TYPE string = SDK_PREFIX + " Validation error. Invalid time to live. Specify time to live parameter as a string."

	INVALID_DELETE_IDS_INPUT string = SDK_PREFIX + " Validation error. Invalid delete IDs type in delete request. Specify delete IDs as a string array."
	EMPTY_DELETE_IDS         string = SDK_PREFIX + " Validation error. Delete IDs array cannot be empty. Specify IDs in delete request."
	INVALID_ID_IN_DELETE     string = SDK_PREFIX + " Validation error. Invalid type of ID passed in delete request. ID must be of type string at index %s1."
	INVALID_DELETE_REQUEST   string = SDK_PREFIX + " Validation error. Invalid delete request. Specify a valid delete request."
	EMPTY_ID_IN_DELETE       string = SDK_PREFIX + " Validation error. ID cannot be empty in delete request. Specify a valid ID."

	MISSING_REDACTION_TYPE_IN_DETOKENIZE string = SDK_PREFIX + " Validation error. Redaction type cannot be empty in detokenize request. Specify the redaction type."
	INVALID_REDACTION_TYPE_IN_DETOKENIZE string = SDK_PREFIX + " Validation error. Invalid redaction type in detokenize request. Specify a redaction type."
	INVALID_TOKENS_TYPE_IN_DETOKENIZE    string = SDK_PREFIX + " Validation error. Invalid tokens type in detokenize request. Specify tokens as a string array."
	EMPTY_TOKENS_IN_DETOKENIZE           string = SDK_PREFIX + " Validation error. Tokens array cannot be empty. Specify tokens in detokenize request."
	EMPTY_TOKEN_IN_DETOKENIZE            string = SDK_PREFIX + " Validation error. Token cannot be empty in detokenize request. Specify a valid token at index %s1."
	INVALID_TOKEN_IN_DETOKENIZE          string = SDK_PREFIX + " Validation error. Invalid type of token passed in detokenize request. Token must be of type string at index %s1."
	INVALID_DETOKENIZE_REQUEST           string = SDK_PREFIX + " Validation error. Invalid detokenize request. Specify a valid detokenize request."

	// Insert errors
	INVALID_RECORD_IN_INSERT       string = SDK_PREFIX + " Validation error. Invalid data in insert request. data must be of type object at index %s1."
	INVALID_RECORD_IN_UPDATE       string = SDK_PREFIX + " Validation error. Invalid data in update request. data must be of type object."
	EMPTY_RECORD_IN_INSERT         string = SDK_PREFIX + " Validation error. Data cannot be empty in insert request. Specify valid data at index %s1."
	INVALID_INSERT_REQUEST         string = SDK_PREFIX + " Validation error. Invalid insert request. Specify a valid insert request."
	INVALID_TYPE_OF_DATA_IN_INSERT string = SDK_PREFIX + " Validation error. Invalid type of data in insert request. Specify data as a object array."
	EMPTY_DATA_IN_INSERT           string = SDK_PREFIX + " Validation error. Data array cannot be empty. Specify data in insert request."

	// Query errors
	INVALID_QUERY_REQUEST string = SDK_PREFIX + " Validation error. Invalid query request. Specify a valid query request."
	INVALID_QUERY         string = SDK_PREFIX + " Validation error. Invalid query in query request. Specify a valid query."

	// File upload errors
	INVALID_FILE_UPLOAD_REQUEST        string = SDK_PREFIX + " Validation error. Invalid file upload request. Specify a valid file upload request."
	MISSING_TABLE_IN_UPLOAD_FILE       string = SDK_PREFIX + " Validation error. Table name cannot be empty in file upload request. Specify table name as a string."
	INVALID_TABLE_IN_UPLOAD_FILE       string = SDK_PREFIX + " Validation error. Invalid table name in file upload request. Specify a valid table name."
	MISSING_SKYFLOW_ID_IN_UPLOAD_FILE  string = SDK_PREFIX + " Validation error. Skyflow id cannot be empty in file upload request. Specify a valid skyflow Id as string."
	INVALID_SKYFLOW_ID_IN_UPLOAD_FILE  string = SDK_PREFIX + " Validation error. Invalid skyflow Id in file upload request. Specify a valid skyflow Id."
	MISSING_COLUMN_NAME_IN_UPLOAD_FILE string = SDK_PREFIX + " Validation error. Column name cannot be empty in file upload request. Specify a valid column name as string."
	INVALID_COLUMN_NAME_IN_UPLOAD_FILE string = SDK_PREFIX + " Validation error. Invalid column name in file upload request. Specify a valid column name."
	MISSING_FILE_PATH_IN_UPLOAD_FILE   string = SDK_PREFIX + " Validation error. File path cannot be empty in file upload request. Specify a valid file path as string."
	INVALID_FILE_PATH_IN_UPLOAD_FILE   string = SDK_PREFIX + " Validation error. Invalid file path in file upload request. Specify a valid file path."

	// Update errors
	MISSING_SKYFLOW_ID_IN_UPDATE string = SDK_PREFIX + " Validation error. Skyflow id name cannot be empty in update request. Specify a skyflow Id name as string."
	INVALID_SKYFLOW_ID_IN_UPDATE string = SDK_PREFIX + " Validation error. Invalid skyflow Id in update request. Specify a valid skyflow Id."
	INVALID_TYPE_OF_UPDATE_DATA  string = SDK_PREFIX + " Validation error. Invalid update data in update request. Specify a valid update data as array of object."
	EMPTY_UPDATE_DATA            string = SDK_PREFIX + " Validation error. Update data cannot be empty in update request. Specify a valid update data."
	INVALID_UPDATE_REQUEST       string = SDK_PREFIX + " Validation error. Invalid update request. Specify a valid update request."
	INVALID_DATA_IN_UPDATE       string = SDK_PREFIX + " Validation error. Invalid data in update request. data must be of type object at index %s1."
	EMPTY_DATA_IN_UPDATE         string = SDK_PREFIX + " Validation error. Data cannot be empty in update request. Specify a valid data at index %s1."
	INVALID_UPDATE_TOKENS        string = SDK_PREFIX + " Validation error. Invalid tokens. Specify valid tokens as object."
	INVALID_TOKEN_IN_UPDATE      string = SDK_PREFIX + " Validation error. Invalid tokens. Specify valid tokens as key value pairs."

	// General errors
	EMPTY_TABLE_NAME       string = SDK_PREFIX + " Validation error. Table name cannot be empty. Specify a valid table name."
	INVALID_TABLE_NAME     string = SDK_PREFIX + " Validation error. Invalid table name. Specify a valid table name as string."
	EMPTY_REDACTION_TYPE   string = SDK_PREFIX + " Validation error. Redaction type cannot be empty. Specify a valid redaction type."
	INVALID_REDACTION_TYPE string = SDK_PREFIX + " Validation error. Invalid redaction type. Specify a valid redaction type."

	// Get errors
	INVALID_TYPE_OF_IDS string = SDK_PREFIX + " Validation error. Invalid ids passed in get request. Specify valid ids as array of string."
	EMPTY_IDS_IN_GET    string = SDK_PREFIX + " Validation error. Ids cannot be empty in get request. Specify valid ids."
	EMPTY_ID_IN_GET     string = SDK_PREFIX + " Validation error. Id cannot be empty. Specify a valid Id at index %s1."
	INVALID_ID_IN_GET   string = SDK_PREFIX + " Validation error. Invalid Id. Specify a valid Id at index %s1 as string."
	INVALID_GET_REQUEST string = SDK_PREFIX + " Validation error. Invalid get request. Specify a valid get request."

	INVALID_COLUMN_NAME               string = SDK_PREFIX + " Validation error. Invalid column name. Specify a valid column name as string."
	INVALID_COLUMN_VALUES             string = SDK_PREFIX + " Validation error. Invalid column values. Specify valid column values as string array."
	EMPTY_COLUMN_VALUE                string = SDK_PREFIX + " Validation error. Column value cannot be empty. Specify a valid column value at index %s."
	INVALID_COLUMN_VALUE              string = SDK_PREFIX + " Validation error. Invalid column value. Specify a valid column value at index %s as string."
	INVALID_GET_COLUMN_REQUEST        string = SDK_PREFIX + " Validation error. Invalid get column request. Specify a valid get column request."
	EMPTY_URL                         string = SDK_PREFIX + " Validation error. Url cannot be empty. Specify a valid url."
	INVALID_URL                       string = SDK_PREFIX + " Validation error. Invalid url. Specify a valid url as a string."
	EMPTY_METHOD_NAME                 string = SDK_PREFIX + " Validation error. Method name cannot be empty. Specify a valid method name."
	INVALID_METHOD_NAME               string = SDK_PREFIX + " Validation error. Invalid method name. Specify a valid method name as a string."
	EMPTY_PATH_PARAMS                 string = SDK_PREFIX + " Validation error. Path params cannot be empty. Specify valid path params."
	INVALID_PATH_PARAMS               string = SDK_PREFIX + " Validation error. Invalid path params. Specify valid path params."
	EMPTY_QUERY_PARAMS                string = SDK_PREFIX + " Validation error. Query params cannot be empty. Specify valid query params."
	INVALID_QUERY_PARAMS              string = SDK_PREFIX + " Validation error. Invalid query params. Specify valid query params."
	EMPTY_BODY                        string = SDK_PREFIX + " Validation error. Body cannot be empty. Specify a valid body."
	INVALID_BODY                      string = SDK_PREFIX + " Validation error. Invalid body. Specify a valid body."
	EMPTY_HEADERS                     string = SDK_PREFIX + " Validation error. Headers cannot be empty. Specify valid headers."
	INVALID_HEADERS                   string = SDK_PREFIX + " Validation error. Invalid headers. Specify valid headers."
	INVALID_INVOKE_CONNECTION_REQUEST string = SDK_PREFIX + " Validation error. Invalid invoke connection request. Specify a valid get invoke connection request."
	ERROR_OCCURRED                    string = SDK_PREFIX + " API error. Error occurred."
	EMPTY_INSERT_TOKEN                string = SDK_PREFIX + " Validation error. Tokens object cannot be empty. Specify a valid tokens object at index %s."
	INVALID_INSERT_TOKEN              string = SDK_PREFIX + " Validation error. Invalid tokens object. Specify a valid tokens object at index %s."
	INVALID_INSERT_TOKENS             string = SDK_PREFIX + " Validation error. Invalid tokens. Specify valid tokens as object array."
	INVALID_TOKEN_MODE                string = SDK_PREFIX + " Validation error. The token mode key has a value of type %s. Specify token as type of TokenMode."
	INVALID_HOMOGENEOUS               string = SDK_PREFIX + " Validation error. The homogeneous key has a value of type %s. Specify homogeneous as boolean."
	INVALID_TOKEN_STRICT              string = SDK_PREFIX + " Validation error. The tokenStrict key has a value of type %s. Specify tokenStrict as boolean."
	INVALID_CONTINUE_ON_ERROR         string = SDK_PREFIX + " Validation error. The continueOnError key has a value of type %s. Specify continueOnError as boolean."
	INVALID_UPSERT                    string = SDK_PREFIX + " Validation error. The upsert key has a value of type %s. Specify upsert as string."
	INVALID_RETURN_TOKEN              string = SDK_PREFIX + " Validation error. The returnToken key has a value of type %s. Specify returnToken as boolean."
	INVALID_DOWNLOAD_URL              string = SDK_PREFIX + " Validation error. The downloadURL key has a value of type %s. Specify downloadURL as string."
	EMPTY_FIELD                       string = SDK_PREFIX + " Validation error. Field value cannot be empty. Specify a valid field value at index %s."
	INVALID_FIELD                     string = SDK_PREFIX + " Validation error. Invalid field value. Specify a valid field value at index %s as string."
	INVALID_OFFSET                    string = SDK_PREFIX + " Validation error. The offset key has a value of type %s. Specify offset as string."
	INVALID_LIMIT                     string = SDK_PREFIX + " Validation error. The limit key has a value of type %s. Specify limit as string."
	INVALID_ORDER_BY                  string = SDK_PREFIX + " Validation error. The orderBy key has a value of type %s. Specify orderBy as string."
	INVALID_FIELDS                    string = SDK_PREFIX + " Validation error. The fields key has a value of type %s. Specify fields as array of strings."
	INVALID_JSON_RESPONSE             string = SDK_PREFIX + " Validation error. The invalid json response. Please reach out to skyflow using requestId - %s."
	EMPTY_VAULT_CLIENTS               string = SDK_PREFIX + " Validation error. No vault config found. Please add a vault config."
	EMPTY_CONNECTION_CLIENTS          string = SDK_PREFIX + " Validation error. No connection config found. Please add a connection config."
	UNKNOWN_ERROR                     string = SDK_PREFIX + " Error occurred. %s"
)
