package logs

import . "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"

const (
	CLIENT_INITIALIZED               = SDK_LOG_PREFIX + "Initialized skyflow client."
	VALIDATING_VAULT_CONFIG          = SDK_LOG_PREFIX + "Validating vault config."
	VALIDATING_CRED                  = SDK_LOG_PREFIX + "Validating credentials."
	VAULT_CONTROLLER_INITIALIZED     = SDK_LOG_PREFIX + "Initialized vault controller with vault ID %s."
	CURRENT_LOG_LEVEL                = SDK_LOG_PREFIX + "Current log level is %v."
	EMPTY_VAULT_ARRAY                = SDK_LOG_PREFIX + "Vault config array is empty."
	VAULT_CONFIG_EXISTS              = SDK_LOG_PREFIX + "Vault config with vault ID %s already exists."
	VAULT_CONFIG_DOES_NOT_EXIST      = SDK_LOG_PREFIX + "Vault config with vault ID %s doesn't exist."
	DEFAULT_VAULT                    = "not specified (default vault)"
	BEARER_TOKEN_EXPIRED             = SDK_LOG_PREFIX + "Bearer token is expired."
	GENERATE_BEARER_TOKEN_TRIGGERED  = SDK_LOG_PREFIX + "Generating a new bearer token."
	REUSE_BEARER_TOKEN               = SDK_LOG_PREFIX + "Reusing bearer token."
	NIL_CONTEXT                      = SDK_LOG_PREFIX + "Context is nil."
	BEARER_TOKEN_CANCELLED           = SDK_LOG_PREFIX + "The context ended before a bearer token was available."
	EMPTY_CREDENTIALS                = SDK_LOG_PREFIX + "No credentials found in the vault config, client config or SKYFLOW_CREDENTIALS."
	METHOD_TRIGGERED                 = SDK_LOG_PREFIX + "%s method triggered."
	VALIDATING_REQUEST               = SDK_LOG_PREFIX + "Validating %s request."
	REQUEST_RESOLVED                 = SDK_LOG_PREFIX + "%s request resolved."
	REQUEST_REJECTED                 = SDK_LOG_PREFIX + "%s request resulted in failure."
	REQUEST_INTERCEPTOR_FAILED       = SDK_LOG_PREFIX + "Request interceptor threw an exception. The request was not sent."
	QUERY_RECORD_LIMIT_REACHED       = SDK_LOG_PREFIX + "Query returned %d rows, the most the vault returns in one call; more rows may match. Use LIMIT and OFFSET in the query to page through them."
	INCOMPLETE_UPLOAD_FILES_RESPONSE = SDK_LOG_PREFIX + "The upload response has a different number of records than the request, so no file was uploaded."
	UPLOAD_FILES_RECORD_MISMATCH     = SDK_LOG_PREFIX + "Upload response record %d does not match its request record, so its files were not uploaded."
)
