package logs

import . "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"

const (
	CLIENT_INITIALIZED              = SDK_LOG_PREFIX + "Initialized skyflow client successfully."
	VALIDATING_VAULT_CONFIG         = SDK_LOG_PREFIX + "Validating vault config."
	VALIDATING_CRED                 = SDK_LOG_PREFIX + "Validating credentials."
	VAULT_CONTROLLER_INITIALIZED    = SDK_LOG_PREFIX + "Initialized vault controller with vault ID %s."
	CURRENT_LOG_LEVEL               = SDK_LOG_PREFIX + "Current log level is %v."
	EMPTY_VAULT_ARRAY               = SDK_LOG_PREFIX + "Vault config array is empty."
	VAULT_CONFIG_EXISTS             = SDK_LOG_PREFIX + "Vault config with vault ID %s already exists."
	VAULT_CONFIG_DOES_NOT_EXIST     = SDK_LOG_PREFIX + "Vault config with vault ID %s doesn't exist."
	BEARER_TOKEN_EXPIRED            = SDK_LOG_PREFIX + "Bearer token is expired."
	GENERATE_BEARER_TOKEN_TRIGGERED = SDK_LOG_PREFIX + "Generating a new bearer token."
	REUSE_BEARER_TOKEN              = SDK_LOG_PREFIX + "Reusing the cached bearer token."
	EMPTY_CREDENTIALS               = SDK_LOG_PREFIX + "No credentials found in the vault config, client config or SKYFLOW_CREDENTIALS."
	METHOD_TRIGGERED                = SDK_LOG_PREFIX + "%s method triggered."
	VALIDATING_REQUEST              = SDK_LOG_PREFIX + "Validating %s request."
	REQUEST_RESOLVED                = SDK_LOG_PREFIX + "%s request resolved."
	REQUEST_REJECTED                = SDK_LOG_PREFIX + "%s request resulted in failure."
)
