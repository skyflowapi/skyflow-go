package internal

// internal private constants
const (
	SDK_METRICS_HEADER_KEY = "sky-metadata"
	SECURE_PROTOCOL        = "https://"
	DEV_DOMAIN             = ".vault.skyflowapis.dev"
	STAGE_DOMAIN           = ".vault.skyflowapis.tech"
	SANDBOX_DOMAIN         = ".vault.skyflowapis-preview.com"
	PROD_DOMAIN            = ".vault.skyflowapis.com"
	GRANT_TYPE             = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	SDK_LOG_PREFIX         = "[ " + SDK_PREFIX + " ] "
	SDK_NAME               = "Skyflow Go SDK "
	SDK_VERSION            = "v2.0.4"
	SDK_PREFIX             = SDK_NAME + SDK_VERSION
	ERROR_FROM_CLIENT      = "error-from-client"
	REQUEST_KEY            = "X-Request-Id"
	SKYFLOW_ID             = "skyflow_id"

	// File extensions
	FILE_EXTENSION_TXT    = "txt"
	FILE_EXTENSION_PDF    = "pdf"
	FILE_EXTENSION_JSON   = "json"
	FILE_EXTENSION_XML    = "xml"
	FILE_EXTENSION_MP3    = "mp3"
	FILE_EXTENSION_WAV    = "wav"
	FILE_EXTENSION_JPG    = "jpg"
	FILE_EXTENSION_JPEG   = "jpeg"
	FILE_EXTENSION_PNG    = "png"
	FILE_EXTENSION_BMP    = "bmp"
	FILE_EXTENSION_TIF    = "tif"
	FILE_EXTENSION_TIFF   = "tiff"
	FILE_EXTENSION_PPT    = "ppt"
	FILE_EXTENSION_PPTX   = "pptx"
	FILE_EXTENSION_CSV    = "csv"
	FILE_EXTENSION_XLS    = "xls"
	FILE_EXTENSION_XLSX   = "xlsx"
	FILE_EXTENSION_DOC    = "doc"
	FILE_EXTENSION_DOCX   = "docx"

	// Encoding types
	ENCODING_BASE64  = "base64"
	ENCODING_UTF8    = "utf-8"
	ENCODING_BINARY  = "binary"

	// File type identifiers
	FILE_TYPE_TEXT            = "text"
	FILE_TYPE_IMAGE           = "image"
	FILE_TYPE_PDF             = "pdf"
	FILE_TYPE_PPT             = "ppt"
	FILE_TYPE_SPREAD          = "spread"
	FILE_TYPE_AUDIO           = "audio"
	FILE_TYPE_DOCUMENT        = "document"
	FILE_TYPE_STRUCTURED      = "structured"
	FILE_TYPE_GENERIC         = "generic"

	// Detect status
	DETECT_STATUS_IN_PROGRESS = "IN_PROGRESS"
	DETECT_STATUS_SUCCESS     = "SUCCESS"
	DETECT_STATUS_FAILED      = "FAILED"

	// HTTP schemes and protocols
	HTTPS_PROTOCOL = "https"
	HTTP_PROTOCOL  = "http"

	// PEM key type
	PRIVATE_KEY_PEM_TYPE = "PRIVATE KEY"

	// Entity types
	ENTITY_TYPE_REDACTED          = "redacted"
	ENTITY_TYPE_MASKED            = "masked"
	ENTITY_TYPE_PLAIN_TEXT        = "plain_text"
	ENTITY_TYPE_TEXT              = "text"
	ENTITY_TYPE_ENTITY_ONLY       = "entity_only"
	ENTITY_TYPE_VAULT_TOKEN       = "vault_token"
	ENTITY_TYPE_ENTITY_UNIQUE_CTR = "entity_unique_counter"
	ENTITY_TYPE_ENTITIES          = "entities"

	// Request/API names
	REQUEST_DEIDENTIFY_FILE = "DeidentifyFileRequest"
	REQUEST_INSERT          = "Insert"
	REQUEST_INSERT_LOWER    = "insert"
	REQUEST_DETOKENIZE      = "DetokenizeRequest"
	REQUEST_GET             = "Get"
	REQUEST_DELETE          = "delete"
	REQUEST_UPDATE          = "update"
	REQUEST_UPLOAD_FILE     = "UploadFile"
	REQUEST_INVOKE_CONN     = "Invoke Connection"

	// HTTP headers
	HEADER_CONTENT_TYPE = "content-type"
	HEADER_CONTENT_TYPE_CAPITAL = "Content-Type"

	// File type mapping for Detect (removed - use FILE_TYPE_* constants instead)

	// Redaction types for Detect
	DETECT_REDACTION_TYPE_REDACTED = "redacted"
	DETECT_REDACTION_TYPE_MASKED   = "masked"
	DETECT_REDACTION_TYPE_PLAINTEXT = "plaintext"
	
	// File output types for Detect
	FILE_OUTPUT_TYPE_REDACTED_FILE = "redacted_file"
	DEIDENTIFIED_FILE_PREFIX = "deidentified."

	// File processing
	PROCESSED_PREFIX = "processed-"
	PERMISSION_CHECK_FILE = ".permission_check"

	// Error and status constants
	UNKNOWN_STATUS = "UNKNOWN"
	UNKNOWN_ERROR = "Unknown error"
	HTTP_STATUS_BAD_REQUEST = "Bad Request"
	ERROR_DETAIL_KEY_FROM_CLIENT = "errorFromClient"
	
	// Environment variables
	SKYFLOW_CREDENTIALS_ENV = "SKYFLOW_CREDENTIALS"
	
	// HTTP headers and content types
	HEADER_AUTHORIZATION = "x-skyflow-authorization"
	CONTENT_TYPE_JSON = "application/json"
	CONTENT_TYPE_TEXT_PLAIN = "text/plain"
	CONTENT_TYPE_TEXT_CHARSET = "text/plain; charset=utf-8"
	RESPONSE_HEADER_REQUEST_ID = "x-request-id"
	
	// JSON error response keys
	ERROR_KEY_ERROR = "error"
	ERROR_KEY_MESSAGE = "message"
	ERROR_KEY_HTTP_CODE = "http_code"
	ERROR_KEY_GRPC_CODE = "grpc_code"
	ERROR_KEY_HTTP_STATUS = "http_status"
	ERROR_KEY_DETAILS = "details"
	
	// JSON response keys
	REQUEST_ID_KEY = "request_id"
	RESPONSE_KEY_REQUEST_ID = "RequestId"
	RESPONSE_KEY_HTTP_CODE = "HttpCode"
	RESPONSE_KEY_SKYFLOW_ID = "skyflowId"
	
	// Other constants
	ERROR_FAILED_TO_READ = "Failed to read error"
	
	// Credentials and JWT keys
	CRED_KEY_PRIVATE_KEY = "privateKey"
	CRED_KEY_CLIENT_ID = "clientID"
	CRED_KEY_TOKEN_URI = "tokenURI"
	CRED_KEY_KEY_ID = "keyID"
	API_KEY_PREFIX = "sky-"
	
	// JWT claim keys
	JWT_CLAIM_EXP = "exp"
	JWT_CLAIM_CTX = "ctx"
	JWT_CLAIM_ISS = "iss"
	JWT_CLAIM_AUD = "aud"
	JWT_CLAIM_KEY = "key"
	JWT_CLAIM_IAT = "iat"
	JWT_CLAIM_SUB = "sub"
	JWT_CLAIM_TOK = "tok"
		
	// Request validation
	REQUEST_INVOKE_CONNECTION = "InvokeConnectionRequest"
	REQUEST_ENTITY_ONLY = "entity_only"
	REQUEST_DEIDENTIFY_TEXT = "DeidentifyTextRequest"
	REQUEST_REIDENTIFY_TEXT = "ReidentifyTextRequest"
	REQUEST_GET_DETECT_RUN = "GetDetectRunRequest"
	REQUEST_TOKENIZE = "Tokenize"
	
	// JSON keys for request/response bodies
	JSON_KEY_BODY = "Body"
	JSON_KEY_RECORDS = "records"
	JSON_KEY_TOKENS = "tokens"
	JSON_KEY_REQUEST_INDEX = "request_index"
	JSON_KEY_TOKENIZED_DATA = "tokenized_data"
	
	// SDK and token generation
	SDK_ISSUER = "sdk"
	SIGNED_TOKEN_PREFIX = "signed_token_"
	
	// SDK metadata keys for CreateJsonMetadata
	SDK_METADATA_KEY_NAME_VERSION = "sdk_name_version"
	SDK_METADATA_KEY_DEVICE_MODEL = "sdk_client_device_model"
	SDK_METADATA_KEY_OS_DETAILS = "sdk_client_os_details"
	SDK_METADATA_KEY_RUNTIME_DETAILS = "sdk_runtime_details"
	
	// Magic numbers
	MAX_REQUEST_SIZE = 42
)
