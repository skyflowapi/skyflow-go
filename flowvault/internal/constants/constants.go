package internal

// internal private constants. FlowVault (v3) vaults are served from the skyvault subdomain,
// not v2's vault subdomain.
const (
	SDK_METRICS_HEADER_KEY = "sky-metadata"
	SECURE_PROTOCOL        = "https://"
	DEV_DOMAIN             = ".skyvault.skyflowapis.dev"
	STAGE_DOMAIN           = ".skyvault.skyflowapis.tech"
	SANDBOX_DOMAIN         = ".skyvault.skyflowapis-preview.com"
	PROD_DOMAIN            = ".skyvault.skyflowapis.com"
	GRANT_TYPE             = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	SDK_NAME               = "Skyflow FlowVault Go SDK "
	SDK_VERSION            = "v0.1.0" // flowvault module version; bump with each flowvault/vX.Y.Z tag
	METRICS_SDK_NAME       = "skyflow-flowvault-go"
	SDK_PREFIX             = SDK_NAME + SDK_VERSION
	SDK_LOG_PREFIX         = "[ " + SDK_PREFIX + " ] "
	ERROR_FROM_CLIENT      = "error-from-client"
	REQUEST_KEY            = "X-Request-Id"

	// HTTP
	HTTP_STATUS_BAD_REQUEST     = "400"
	HEADER_CONTENT_TYPE_CAPITAL = "Content-Type"
	CONTENT_TYPE_JSON           = "application/json"
	CONTENT_TYPE_TEXT_PLAIN     = "text/plain"
	CONTENT_TYPE_TEXT_CHARSET   = "text/plain; charset=utf-8"
	UNKNOWN_ERROR               = "unknown error"
	ERROR_FAILED_TO_READ        = "failed to read response body"

	// Error response keys
	ERROR_KEY_ERROR       = "error"
	ERROR_KEY_HTTP_CODE   = "http_code"
	ERROR_KEY_MESSAGE     = "message"
	ERROR_KEY_GRPC_CODE   = "grpc_code"
	ERROR_KEY_HTTP_STATUS = "http_status"
	ERROR_KEY_DETAILS     = "details"
	ERROR_KEY_FROM_CLIENT = "from_client"
)

// Method names used in logs and error messages, and operation names passed to interceptors.
const (
	METHOD_INSERT       = "Insert"
	METHOD_UPDATE       = "Update"
	METHOD_GET          = "Get"
	METHOD_DELETE       = "Delete"
	METHOD_DETOKENIZE   = "Detokenize"
	METHOD_GET_TOKENS   = "GetTokens"
	METHOD_QUERY        = "Query"
	METHOD_UPLOAD_FILES = "UploadFiles"
	METHOD_DELETE_FILES = "DeleteFiles"

	OPERATION_INSERT       = "INSERT"
	OPERATION_UPDATE       = "UPDATE"
	OPERATION_GET          = "GET"
	OPERATION_DELETE       = "DELETE"
	OPERATION_DETOKENIZE   = "DETOKENIZE"
	OPERATION_GET_TOKENS   = "GET_TOKENS"
	OPERATION_QUERY        = "QUERY"
	OPERATION_UPLOAD_FILES = "UPLOAD_FILES"
	OPERATION_DELETE_FILES = "DELETE_FILES"

	MAX_DELETE_FILES_RECORDS   = 20000
	MAX_CONCURRENT_UPLOADS     = 5
	RESPONSE_HEADER_REQUEST_ID = "x-request-id"
	SKYFLOW_CREDENTIALS_ENV    = "SKYFLOW_CREDENTIALS"
)

// HTTP defaults, used when neither the vault nor the client sets a value.
const (
	DEFAULT_TIMEOUT_SECONDS            = 60
	DEFAULT_MAX_RETRIES                = 0
	DEFAULT_INITIAL_RETRY_DELAY_MILLIS = 500
	DEFAULT_MAX_RETRY_DELAY_MILLIS     = 2000
	RETRY_JITTER_FACTOR                = 0.2

	// Signed-URL uploads: bounds for connecting and for the storage service's response,
	// matching the Java and Python defaults. There is no overall upload timeout.
	UPLOAD_CONNECT_TIMEOUT_SECONDS         = 10
	UPLOAD_RESPONSE_HEADER_TIMEOUT_SECONDS = 10
	UPLOAD_KEEP_ALIVE_SECONDS              = 30
)
