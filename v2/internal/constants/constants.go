package internal

// internal private constants
const (
	SECURE_PROTOCOL          = "https://"
	DEV_DOMAIN               = ".vaultapi.skyflowapis.dev"
	STAGE_DOMAIN             = ".vaultapi.skyflowapis.tech"
	SANDBOX_DOMAIN           = ".vaultapi.skyflowapis-preview.com"
	PROD_DOMAIN              = ".vaultapi.skyflowapis.com"
	PKCS8_PRIVATE_HEADER     = "-----BEGIN PRIVATE KEY-----"
	PKCS8_PRIVATE_FOOTER     = "-----END PRIVATE KEY-----"
	GRANT_TYPE               = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	SIGNED_DATA_TOKEN_PREFIX = "signed_token_"
)
