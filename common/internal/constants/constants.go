package constants

// Constants used by the shared SkyflowError type. Private to the common module.
const (
	REQUEST_KEY                 = "X-Request-Id"
	ERROR_FROM_CLIENT           = "error-from-client"
	HEADER_CONTENT_TYPE_CAPITAL = "Content-Type"
	CONTENT_TYPE_JSON           = "application/json"
	CONTENT_TYPE_TEXT_PLAIN     = "text/plain"
	CONTENT_TYPE_TEXT_CHARSET   = "text/plain; charset=utf-8"
	UNKNOWN_ERROR               = "Unknown error"
	HTTP_STATUS_BAD_REQUEST     = "Bad Request"
	ERROR_FAILED_TO_READ        = "Failed to read error"

	ERROR_KEY_ERROR       = "error"
	ERROR_KEY_MESSAGE     = "message"
	ERROR_KEY_HTTP_CODE   = "http_code"
	ERROR_KEY_GRPC_CODE   = "grpc_code"
	ERROR_KEY_HTTP_STATUS = "http_status"
	ERROR_KEY_DETAILS     = "details"
	ERROR_KEY_FROM_CLIENT = "errorFromClient"
)
