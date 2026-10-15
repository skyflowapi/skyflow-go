package logs

import . "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"

const (
	CLIENT_INITIALIZED           = SDK_LOG_PREFIX + "Initialized skyflow client successfully."
	VALIDATING_VAULT_CONFIG      = SDK_LOG_PREFIX + "Validating vault config."
	VALIDATING_CRED              = SDK_LOG_PREFIX + "Validating credentials."
	VAULT_CONTROLLER_INITIALIZED = SDK_LOG_PREFIX + "Initialized vault controller with vault ID %s."
	CURRENT_LOG_LEVEL            = SDK_LOG_PREFIX + "Current log level is %v."
	EMPTY_VAULT_ARRAY            = SDK_LOG_PREFIX + "Vault config array is empty."
	VAULT_CONFIG_EXISTS          = SDK_LOG_PREFIX + "Vault config with vault ID %s already exists."
	VAULT_CONFIG_DOES_NOT_EXIST  = SDK_LOG_PREFIX + "Vault config with vault ID %s doesn't exist."
)
