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
	SDK_NAME               = "Skyflow FlowVault Go SDK "
	SDK_VERSION            = "1.0.0" // flowvault module version without the "v"; bump with each flowvault/vX.Y.Z tag
	METRICS_SDK_NAME       = "skyflow-flowvault-go"
	SDK_PREFIX             = SDK_NAME + SDK_VERSION
	SDK_LOG_PREFIX         = "[" + SDK_PREFIX + "] " // e.g. "[Skyflow FlowVault Go SDK 1.0.0] "

	// sky-metadata header keys, the same in every Skyflow SDK.
	SDK_METADATA_KEY_NAME_VERSION    = "sdk_name_version"
	SDK_METADATA_KEY_DEVICE_MODEL    = "sdk_client_device_model"
	SDK_METADATA_KEY_OS_DETAILS      = "sdk_client_os_details"
	SDK_METADATA_KEY_RUNTIME_DETAILS = "sdk_runtime_details"

	// Error response keys
	ERROR_KEY_ERROR       = "error"
	ERROR_KEY_HTTP_CODE   = "http_code"
	ERROR_KEY_MESSAGE     = "message"
	ERROR_KEY_GRPC_CODE   = "grpc_code"
	ERROR_KEY_HTTP_STATUS = "http_status"
	ERROR_KEY_DETAILS     = "details"
	ERROR_KEY_FROM_CLIENT = "from_client"
	ERROR_KEY_RECORDS     = "records"  // records-shaped error body: {"records":[{"error":"..."}]}
	ERROR_KEY_RESPONSE    = "response" // the records key of a detokenize body

	// RECORD_ERRORS_SEPARATOR joins the distinct record errors of a failed uploadFiles call.
	RECORD_ERRORS_SEPARATOR = "; "

	// A non-JSON error body becomes the error message. A gateway or proxy can echo the
	// request, which carries sensitive data, so only its first MAX_RAW_ERROR_BODY bytes are
	// kept (as in the common module).
	MAX_RAW_ERROR_BODY = 512
	TRUNCATED_SUFFIX   = "... (truncated)"

	// API_ERROR_FORMAT is "<status>: <JSON error body>", the error text SkyflowErrorApi parses.
	API_ERROR_FORMAT = "%d: %s"
	// EMPTY_JSON_OBJECT replaces a blank 2xx body so it parses as an empty response.
	EMPTY_JSON_OBJECT = "{}"

	// ERROR_WITH_STATUS_CODE is the message for an error response with an empty body.
	ERROR_WITH_STATUS_CODE = "Error with status code %d"
	// UNKNOWN_RECORD_ERROR is the error on a record of a failed call that reports neither an
	// error nor a 2xx status of its own.
	UNKNOWN_RECORD_ERROR = "Unknown error"
)

// Per-column and per-record upload errors.
const (
	NO_SIGNED_URL_ERROR            = "No signed URL was returned for this column; file was not uploaded."
	UPLOAD_COLUMNS_FAILED_ERROR    = "%d of %d files were not uploaded. See each column for details."
	UPLOAD_RESPONSE_MISMATCH_ERROR = "Upload response did not match the request; file was not uploaded."
)

// Method names used in logs and error messages, and operation names used in header validation errors.
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

	QUERY_RECORD_LIMIT         = 25 // the most rows the vault returns in one query call
	MAX_CONCURRENT_UPLOADS     = 5
	RESPONSE_HEADER_REQUEST_ID = "x-request-id"
	SKYFLOW_CREDENTIALS_ENV    = "SKYFLOW_CREDENTIALS"
)

// HTTP defaults, used when neither the vault nor the client sets a value.
const (
	DEFAULT_TIMEOUT_SECONDS            = 60
	DEFAULT_CONNECT_TIMEOUT_SECONDS    = 10 // per attempt
	DEFAULT_READ_TIMEOUT_SECONDS       = 10 // per attempt
	DEFAULT_WRITE_TIMEOUT_SECONDS      = 10 // per attempt
	DEFAULT_MAX_RETRIES                = 0
	DEFAULT_INITIAL_RETRY_DELAY_MILLIS = 500
	DEFAULT_MAX_RETRY_DELAY_MILLIS     = 2000
	RETRY_JITTER_FACTOR                = 0.2
)

// DEFAULT_UPLOAD_CONTENT_TYPE is the Content-Type of an uploaded file whose type isn't known.
const DEFAULT_UPLOAD_CONTENT_TYPE = "application/octet-stream"

// UploadContentTypes maps a lowercase file extension to the Content-Type of an uploaded file.
// It is the SDK's own list, so a file gets the same type on every OS (Go's
// mime table depends on the system's). Other extensions fall back to Go's table, then to
// DEFAULT_UPLOAD_CONTENT_TYPE.
var UploadContentTypes = map[string]string{
	"pdf": "application/pdf",
	"jpg": "image/jpeg", "jpeg": "image/jpeg", "png": "image/png", "gif": "image/gif",
	"bmp": "image/bmp", "tif": "image/tiff", "tiff": "image/tiff", "webp": "image/webp",
	"heic": "image/heic", "heif": "image/heif", "svg": "image/svg+xml",
	"txt": "text/plain", "csv": "text/csv", "tsv": "text/tab-separated-values",
	"json": "application/json", "xml": "application/xml", "html": "text/html",
	"htm": "text/html", "rtf": "application/rtf",
	"doc":  "application/msword",
	"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"xls":  "application/vnd.ms-excel",
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"ppt":  "application/vnd.ms-powerpoint",
	"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"odt":  "application/vnd.oasis.opendocument.text",
	"ods":  "application/vnd.oasis.opendocument.spreadsheet",
	"odp":  "application/vnd.oasis.opendocument.presentation",
	"zip":  "application/zip", "gz": "application/gzip", "tar": "application/x-tar",
	"7z":  "application/x-7z-compressed",
	"mp3": "audio/mpeg", "wav": "audio/wav", "m4a": "audio/mp4", "ogg": "audio/ogg",
	"flac": "audio/flac",
	"mp4":  "video/mp4", "mov": "video/quicktime", "avi": "video/x-msvideo",
	"webm": "video/webm", "mkv": "video/x-matroska",
}
