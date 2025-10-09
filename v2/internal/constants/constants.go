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
	SDK_VERSION            = "v2.0.0-beta.1"
	SDK_PREFIX             = SDK_NAME + SDK_VERSION
	ERROR_FROM_CLIENT      = "error-from-client"
	REQUEST_KEY            = "X-Request-Id"
	SKYFLOW_ID             = "skyflow_id"
)
