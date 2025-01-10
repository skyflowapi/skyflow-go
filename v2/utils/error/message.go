package errors

import . "github.com/skyflowapi/skyflow-go/v2/internal/constants"

// TO DO
const (
	// config
	VAULT_ID_ALREADY_IN_CONFIG_LIST        string = SDK_PREFIX + " Validation error. VaultId is present in an existing config. Specify a new vaultId in config."
	VAULT_ID_NOT_IN_CONFIG_LIST            string = SDK_PREFIX + " Validation error. VaultId is missing from the config. Specify the vaultIds from configs."
	CONNECTION_ID_NOT_IN_CONFIG_LIST       string = SDK_PREFIX + " Validation error. ConnectionId is missing from the config. Specify the connectionIds from configs."
	EMPTY_CREDENTIALS                      string = SDK_PREFIX + " Validation error. Invalid credentials. Specify a valid credentials."
	EMPTY_VAULT_CONFIG                     string = SDK_PREFIX + " Validation error. No vault configurations available"
	INVALID_VAULT_ID                       string = SDK_PREFIX + " Initialization failed. Invalid vault ID. Specify a valid vault ID."
	INVALID_CLUSTER_ID                     string = SDK_PREFIX + " Initialization failed. Invalid cluster ID. Specify cluster ID."
	EMPTY_CONNECTION_CONFIG                string = SDK_PREFIX + " Validation error. No connection configurations available"
	EMPTY_CONNECTION_ID                    string = SDK_PREFIX + " Initialization failed. Invalid connection ID. Connection ID must not be empty."
	INVALID_CONNECTION_URL                 string = SDK_PREFIX + " Initialization failed. Invalid connection URL. Specify a valid connection URL."
	EMPTY_CONNECTION_URL                   string = SDK_PREFIX + " Initialization failed. Invalid connection URL. Connection URL must not be empty."
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

	FILE_NOT_FOUND                                    string = SDK_PREFIX + " Initialization failed. Credential file not found at %s. Verify the file path."
	CREDENTIALS_STRING_INVALID_JSON                   string = SDK_PREFIX + " Initialization failed. Credentials string is not in valid JSON format. Verify the credentials string contents."
	INVALID_CREDENTIALS                               string = SDK_PREFIX + " Initialization failed. Invalid credentials provided. Specify valid credentials."
	MISSING_PRIVATE_KEY                               string = SDK_PREFIX + " Initialization failed. Unable to read private key in credentials. Verify your private key."
	MISSING_CLIENT_ID                                 string = SDK_PREFIX + " Initialization failed. Unable to read client ID in credentials. Verify your client ID."
	MISSING_KEY_ID                                    string = SDK_PREFIX + " Initialization failed. Unable to read key ID in credentials. Verify your key ID."
	MISSING_TOKEN_URI                                 string = SDK_PREFIX + " Initialization failed. Unable to read token URI in credentials. Verify your token URI."
	INVALID_TOKEN_URI                                 string = SDK_PREFIX + " Initialization failed. Token URI in not a valid URL in credentials. Verify your token URI."
	JWT_INVALID_FORMAT                                string = SDK_PREFIX + " Initialization failed. Invalid private key format. Verify your credentials."
	INVALID_ALGORITHM                                 string = SDK_PREFIX + " Initialization failed. Invalid algorithm to parse private key. Specify valid algorithm."
	INVALID_KEY_SPEC                                  string = SDK_PREFIX + " Initialization failed. Unable to parse RSA private key. Verify your credentials."
	TABLE_KEY_ERROR                                   string = SDK_PREFIX + " Validation error. 'table' key is missing from the payload. Specify a 'table' key."
	EMPTY_TABLE                                       string = SDK_PREFIX + " Validation error. 'table' can't be empty. Specify a table."
	EMPTY_VALUES                                      string = SDK_PREFIX + " Validation error. 'values' can't be empty. Specify values."
	EMPTY_KEY_IN_VALUES                               string = SDK_PREFIX + " Validation error. Invalid key in values. Specify a valid key."
	EMPTY_VALUE_IN_VALUES                             string = SDK_PREFIX + " Validation error. Invalid value in values. Specify a valid value."
	EMPTY_TOKENS                                      string = SDK_PREFIX + " Validation error. The 'tokens' field is empty. Specify tokens for one or more fields."
	EMPTY_KEY_IN_TOKENS                               string = SDK_PREFIX + " Validation error. Invalid key tokens. Specify a valid key."
	EMPTY_VALUE_IN_TOKENS                             string = SDK_PREFIX + " Validation error. Invalid value in tokens. Specify a valid value."
	HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT              string = SDK_PREFIX + " Validation error. 'homogenous' is not supported with 'upsert'. Specify either 'homogenous' or 'upsert'."
	TOKENS_PASSED_FOR_BYOT_DISABLE                    string = SDK_PREFIX + " Validation error. 'TokenMode' wasn't specified. Set 'TokenMode' to 'ENABLE' to insert tokens."
	NO_TOKENS_WITH_BYOT                               string = SDK_PREFIX + " Validation error. Tokens weren't specified for records while 'TokenMode' was %s. Specify tokens."
	MISMATCH_OF_FIELDS_AND_TOKENS                     string = SDK_PREFIX + " Validation error. 'fields' and 'tokens' have different columns names. Verify that 'fields' and 'tokens' columns match."
	INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT string = SDK_PREFIX + " Validation error. 'token_mode' is set to 'ENABLE_STRICT', but some fields are missing tokens. Specify tokens for all fields."
	TOKENS_NOT_PASSED                                 string = SDK_PREFIX + " Insert failed. 'byot' was set to 'ENABLE' or 'ENABLE_STRICT', but no tokens were specified. Specify valid tokens."
	EMPTY_PARAMETERS                                  string = SDK_PREFIX + " Validation error. Parameters must not be empty. Specify valid parameters."
	EMPTY_PARAMETER_NAME                              string = SDK_PREFIX + " Validation error. Parameter name is missing in parameters. Specify a valid name."
	EMPTY_PARAMETER_VALUE                             string = SDK_PREFIX + " Validation error. Parameter value is missing for name. Specify a valid value."
	INVALID_RESPONSE                                  string = SDK_PREFIX + " Error while parsing server response. Please verify the response."
	EMPTY_REQUEST_HEADER                              string = SDK_PREFIX + " Validation error. Request headers are empty. Specify valid request headers."
	INVALID_REQUEST_HEADERS                           string = SDK_PREFIX + "Validation error. Request headers aren't valid. Specify valid request headers."
	INVALID_QUERY_PARAM                               string = SDK_PREFIX + " Validation error. Query parameters aren't valid. Specify valid query parameters."
	EMPTY_QUERY_PARAM                                 string = SDK_PREFIX + " Validation error. Query parameters are empty. Specify valid query parameters."
	EMPTY_REQUEST_BODY                                string = SDK_PREFIX + " Validation error. Request body can't be empty. Specify a valid request body."
	EMPTY_TOKENS_DETOKENIZE                           string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify at least one data token."
	INVALID_DATA_TOKENS                               string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify valid data tokens."
	EMPTY_TOKEN_IN_DATA_TOKEN                         string = SDK_PREFIX + " Validation error. Invalid data tokens. Specify a valid data token."
	EMPTY_IDS                                         string = SDK_PREFIX + " Validation error. 'ids' can't be empty. Specify at least one id."
	EMPTY_ID_IN_IDS                                   string = SDK_PREFIX + " Validation error. Invalid id in 'ids'. Specify a valid id."
	EMPTY_FIELDS                                      string = SDK_PREFIX + " Validation error. Fields are empty in get payload. Specify at least one field."
	EMPTY_FIELD_IN_FIELDS                             string = SDK_PREFIX + " Validation error. Invalid field in 'fields'. Specify a valid field."
	REDACTION_WITH_TOKENS_NOT_SUPPORTED               string = SDK_PREFIX + " Validation error. 'redaction' can't be used when 'returnTokens' is specified. Remove 'redaction' from payload if 'returnTokens' is specified."
	TOKENS_GET_COLUMN_NOT_SUPPORTED                   string = SDK_PREFIX + " Validation error. Column name and/or column values can't be used when 'returnTokens' is specified. Remove unique column values or 'returnTokens' from the payload."
	UNIQUE_COLUMN_OR_IDS_KEY_ERROR                    string = SDK_PREFIX + " Validation error. 'ids' or 'columnName' key is missing from the payload. Specify the ids or unique 'columnName' in payload."
	BOTH_IDS_AND_COLUMN_DETAILS_SPECIFIED             string = SDK_PREFIX + " Validation error. Both Skyflow IDs and column details can't be specified. Either specify Skyflow IDs or unique column details."
	COLUMN_NAME_KEY_ERROR                             string = SDK_PREFIX + " Validation error. 'columnName' isn't specified whereas 'columnValues' are specified. Either add 'columnName' or remove 'columnValues'."
	EMPTY_COLUMN_VALUES                               string = SDK_PREFIX + " Validation error. 'columnValues' can't be empty. Specify at least one column value"
	EMPTY_VALUE_IN_COLUMN_VALUES                      string = SDK_PREFIX + " Validation error. Invalid value in column values. Specify a valid column value."
	EMPTY_QUERY                                       string = SDK_PREFIX + " Validation error. 'query' can't be empty."

	EMPTY_ID_IN_UPDATE string = SDK_PREFIX + " Validation error. 'id' can't be empty. Specify an id."
	// Error messages
	VAULT_ID_EXITS_IN_CONFIG_LIST      string = SDK_PREFIX + " Validation error. %s1 already exists in the config list. Specify a new vaultId."
	CONNECTION_ID_EXITS_IN_CONFIG_LIST string = SDK_PREFIX + " Validation error. %s1 already exists in the config list. Specify a new vaultId."
	INVALID_METHOD_NAME                string = SDK_PREFIX + " Validation error. Invalid method name. Specify a valid method name as a string."
	ERROR_OCCURRED                     string = SDK_PREFIX + " API error. Error occurred."
	UNKNOWN_ERROR                      string = SDK_PREFIX + " Error occurred. %s"
)
