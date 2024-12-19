package logs

import . "github.com/skyflowapi/skyflow-go/v2/internal/constants"

const (
	CLIENT_ID_NOT_FOUND   = SDK_LOG_PREFIX + "Invalid credentials. Client ID cannot be empty."
	TOKEN_URI_NOT_FOUND   = SDK_LOG_PREFIX + "Invalid credentials. Token URI cannot be empty."
	KEY_ID_NOT_FOUND      = SDK_LOG_PREFIX + "Invalid credentials. Key ID cannot be empty."
	PRIVATE_KEY_NOT_FOUND = SDK_LOG_PREFIX + "Invalid credentials. Private key cannot be empty."
	NOT_A_VALID_JSON      = SDK_LOG_PREFIX + "Credentials is not in valid JSON format. Verify the credentials."
	FILE_NOT_FOUND        = SDK_LOG_PREFIX + "Credential file not found at %s. Verify the file path."
	INVALID_INPUT_FILE    = SDK_LOG_PREFIX + "Unable to read credentials - file %s."

	DETOKENIZE_REQUEST_REJECTED        = SDK_LOG_PREFIX + "Detokenize request resulted in failure."
	TOKENIZE_REQUEST_REJECTED          = SDK_LOG_PREFIX + "Tokenize request resulted in failure."
	INVOKE_CONNECTION_REQUEST_REJECTED = SDK_LOG_PREFIX + "Invoke connection request resulted in failure."
	QUERY_REQUEST_REJECTED             = SDK_LOG_PREFIX + "Query request resulted in failure."
	INSERT_REQUEST_REJECTED            = SDK_LOG_PREFIX + "Insert request resulted in failure."
	GET_REQUEST_REJECTED               = SDK_LOG_PREFIX + "Get request resulted in failure."
	UPDATE_REQUEST_REJECTED            = SDK_LOG_PREFIX + "Update request resulted in failure."
	DELETE_REQUEST_REJECTED            = SDK_LOG_PREFIX + "Delete request resulted in failure."

	EMPTY_QUERY                         = SDK_LOG_PREFIX + "Invalid query request. Query can not be empty."
	EMPTY_COLUMN_GROUP_IN_COLUMN_VALUES = SDK_LOG_PREFIX + "Invalid tokenize request. Column group can not be null or empty in column values at index %v."
	EMPTY_TABLE                         = SDK_LOG_PREFIX + "Invalid %s request. Table name can not be empty."
	EMPTY_IDS                           = SDK_LOG_PREFIX + "Invalid %s request. Ids can not be empty."
	INVALID_ID                          = SDK_LOG_PREFIX + "Invalid %s request. Id can not be null or empty in ids at index %v."

	EMPTY_TOKENS_IN_DETOKENIZE = SDK_LOG_PREFIX + "Invalid detokenize request. Tokens can not be empty."

	EMPTY_COLUMN_NAME_IN_GET_COLUMN = SDK_LOG_PREFIX + "Invalid get column request. Column name can not be empty."

	EMPTY_COLUMN_VALUES_IN_GET_COLUMN = SDK_LOG_PREFIX + "Invalid get column request. Column values can not be empty."

	INVALID_SKYFLOW_ID_IN_UPDATE = SDK_LOG_PREFIX + "Invalid update request. Skyflow Id is required."

	// Client initialization
	VAULT_CONFIG_EXISTS         = SDK_LOG_PREFIX + "Vault config with vault ID %s already exists."
	VAULT_CONFIG_DOES_NOT_EXIST = SDK_LOG_PREFIX + "Vault config with vault ID %s doesn't exist."
	VAULT_ID_IS_REQUIRED        = SDK_LOG_PREFIX + "Invalid vault config. Vault ID is required."

	CLUSTER_ID_IS_REQUIRED = SDK_LOG_PREFIX + "Invalid vault config. Cluster ID is required."

	CONNECTION_CONFIG_EXISTS         = SDK_LOG_PREFIX + "Connection config with connection ID %s already exists."
	CONNECTION_CONFIG_DOES_NOT_EXIST = SDK_LOG_PREFIX + "Connection config with connection ID %s doesn't exist."
	CONNECTION_ID_IS_REQUIRED        = SDK_LOG_PREFIX + "Invalid connection config. Connection ID is required."

	CONNECTION_URL_IS_REQUIRED = SDK_LOG_PREFIX + "Invalid connection config. Connection URL is required."

	INVALID_CONNECTION_URL                 = SDK_LOG_PREFIX + "Invalid connection config. Connection URL is not a valid URL."
	MULTIPLE_TOKEN_GENERATION_MEANS_PASSED = SDK_LOG_PREFIX + "Invalid credentials. Only one of 'path', 'credentialsString', 'token' or 'apiKey' is allowed."
	NO_TOKEN_GENERATION_MEANS_PASSED       = SDK_LOG_PREFIX + "Invalid credentials. Any one of 'path', 'credentialsString', 'token' or 'apiKey' is required."
	EMPTY_CREDENTIALS_PATH                 = SDK_LOG_PREFIX + "Invalid credentials. Credentials path can not be empty."

	INVALID_API_KEY             = SDK_LOG_PREFIX + "Invalid credentials. Api key is invalid."
	EMPTY_ROLES                 = SDK_LOG_PREFIX + "Invalid credentials. Roles can not be empty."
	EMPTY_OR_NULL_ROLE_IN_ROLES = SDK_LOG_PREFIX + "Invalid credentials. Role can not be null or empty in roles at index %v."

	INVALID_CREDENTIALS_FILE_FORMAT = SDK_LOG_PREFIX + "Credentials file is not in a valid JSON format."

	INVALID_CREDENTIALS_STRING_FORMAT = SDK_LOG_PREFIX + "Credentials string is not in a valid JSON string format."

	INVALID_TOKEN_URI                                        = SDK_LOG_PREFIX + "Invalid value for token URI in credentials."
	JWT_INVALID_FORMAT                                       = SDK_LOG_PREFIX + "Private key is not in a valid format."
	INVALID_ALGORITHM                                        = SDK_LOG_PREFIX + "Algorithm for parsing private key is invalid or does not exist."
	INVALID_KEY_SPEC                                         = SDK_LOG_PREFIX + "Unable to parse RSA private key."
	BEARER_TOKEN_REJECTED                                    = SDK_LOG_PREFIX + "Bearer token request resulted in failure."
	PRIVATE_KEY_TYPE                                         = SDK_LOG_PREFIX + "RSA private key is of the wrong type Pem Type: %s"
	PARSE_JWT_PAYLOAD                                        = SDK_LOG_PREFIX + "Unable to parse jwt payload"
	EMPTY_REQUEST_HEADERS                                    = SDK_LOG_PREFIX + "Invalid %s request. Request headers can not be empty."
	INVALID_REQUEST_HEADERS                                  = SDK_LOG_PREFIX + "Invalid %s request. Request header can not be nil or empty in request headers."
	EMPTY_PATH_PARAMS                                        = SDK_LOG_PREFIX + "Invalid %s request. Path params can not be empty."
	INVALID_PATH_PARAM                                       = SDK_LOG_PREFIX + "Invalid %s request. Path parameter can not be null or empty in path params."
	EMPTY_QUERY_PARAMS                                       = SDK_LOG_PREFIX + "Invalid %s request. Query params can not be empty."
	INVALID_QUERY_PARAM                                      = SDK_LOG_PREFIX + "Invalid %s request. Query parameter can not be null or empty in query params."
	EMPTY_REQUEST_BODY                                       = SDK_LOG_PREFIX + "Invalid %s request. Request body can not be empty."
	TOKENS_REQUIRED                                          = SDK_LOG_PREFIX + "Invalid %s request. Tokens are required."
	EMPTY_OR_NULL_TOKEN_IN_TOKENS                            = SDK_LOG_PREFIX + "Invalid %s request. Token can not be nil or empty in tokens at index %v."
	VALUES_IS_REQUIRED                                       = SDK_LOG_PREFIX + "Invalid %s request. Values are required."
	EMPTY_VALUES                                             = SDK_LOG_PREFIX + "Invalid %s request. Values can not be empty."
	HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT                     = SDK_LOG_PREFIX + "Invalid %s request. Homogenous is not supported when upsert is passed."
	EMPTY_OR_NULL_KEY_IN_VALUES                              = SDK_LOG_PREFIX + "Invalid %s request. Key can not be null or empty in values"
	EMPTY_OR_NULL_VALUE_IN_VALUES                            = SDK_LOG_PREFIX + "Invalid %s request. Value can not be null or empty in values for key %s"
	TOKENS_NOT_ALLOWED_WITH_BYOT_DISABLE                     = SDK_LOG_PREFIX + "Invalid %s request. Tokens are not allowed when tokenStrict is DISABLE."
	TOKENS_REQUIRED_WITH_BYOT                                = SDK_LOG_PREFIX + "Invalid %s request. Tokens are required when tokenMode is %s."
	INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE               = SDK_LOG_PREFIX + "Invalid %s request. For tokenStrict as ENABLE, tokens should be passed for all fields object."
	INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT        = SDK_LOG_PREFIX + "Invalid %s request. For tokenStrict as ENABLE_STRICT, tokens should be passed for all fields."
	EMPTY_TOKENS                                             = SDK_LOG_PREFIX + "Invalid %s request. Tokens can not be empty."
	MISMATCH_OF_FIELDS_AND_TOKENS                            = SDK_LOG_PREFIX + "Invalid %s request. Keys for values and tokens are not matching."
	EMPTY_OR_NULL_VALUE_IN_TOKENS                            = SDK_LOG_PREFIX + "Invalid %s request. Value can not be null or empty in tokens for key %s."
	EMPTY_OR_NULL_KEY_IN_TOKENS                              = SDK_LOG_PREFIX + "Invalid %s request. Key can not be null or empty in tokens"
	EMPTY_OR_NULL_ID_IN_IDS                                  = SDK_LOG_PREFIX + "Invalid %s request. Id can not be null or empty in ids at index %v."
	EMPTY_FIELDS                                             = SDK_LOG_PREFIX + "Invalid %s request. Fields can not be empty."
	EMPTY_OR_NULL_FIELD_IN_FIELDS                            = SDK_LOG_PREFIX + "Invalid %s request. Field can not be null or empty in fields at index %v."
	TOKENIZATION_NOT_SUPPORTED_WITH_REDACTION                = SDK_LOG_PREFIX + "Invalid %s request. Return tokens is not supported when redaction is applied."
	NEITHER_IDS_NOR_COLUMN_NAME_PASSED                       = SDK_LOG_PREFIX + "Invalid %s request. Neither ids nor column name and values are passed."
	BOTH_IDS_AND_COLUMN_NAME_PASSED                          = SDK_LOG_PREFIX + "Invalid %s request. Both ids and column name and values are passed."
	EMPTY_OR_NULL_COLUMN_VALUE_IN_COLUMN_VALUES              = SDK_LOG_PREFIX + "Invalid %s request. Column value can not by null or empty in column values at index %v."
	INVALID_TOKENIZE_REQUEST                                 = SDK_LOG_PREFIX + "Invalid tokenize request. Specify a tokenize request."
	EMPTY_VAULT_ARRAY                                 string = SDK_LOG_PREFIX + "Validation error. No vault configurations passed"
	EMPTY_CONNECTION_ARRAY                            string = SDK_LOG_PREFIX + "Validation error. No Connection configurations passed"
)
