package internal

// internal private constants — flowvault copy (cannot import v2's internal)
const (
	SDK_METRICS_HEADER_KEY = "sky-metadata"
	SECURE_PROTOCOL        = "https://"
	DEV_DOMAIN             = ".vault.skyflowapis.dev"
	STAGE_DOMAIN           = ".vault.skyflowapis.tech"
	SANDBOX_DOMAIN         = ".vault.skyflowapis-preview.com"
	PROD_DOMAIN            = ".vault.skyflowapis.com"
	GRANT_TYPE             = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	SDK_NAME               = "Skyflow Go SDK "
	SDK_VERSION            = "v1.0.0"
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
