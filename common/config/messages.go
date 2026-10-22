package config

// Message bodies without the SDK prefix; the Validator prepends the owning
// SDK's prefix so each SDK reports its own name/version (same pattern as auth).

const (
	errInvalidVaultId                     = " Initialization failed. Invalid vault ID. Specify a valid vault ID."
	errInvalidClusterId                   = " Initialization failed. Invalid cluster ID. Specify cluster ID."
	errInvalidVaultUrl                    = " Initialization failed. Invalid vault URL. Specify a valid vault URL."
	errMultipleTokenGenerationMeansPassed = " Initialization failed. Invalid credentials. Specify only one from 'path', 'credentialsString', 'token' or 'apiKey'."
	errNoTokenGenerationMeansPassed       = " Initialization failed. Invalid credentials. Specify any one from 'path', 'credentialsString', 'token' or 'apiKey'."
	errInvalidApiKey                      = " Initialization failed. Invalid credentials. Specify valid api key."
	errEmptyRoles                         = " Initialization failed. Invalid roles. Specify at least one role."
	errEmptyRoleInRoles                   = " Initialization failed. Invalid role. Specify a valid role."
	errNegativeHttpConfig                 = " Initialization failed. Invalid HTTP config. Timeout, retry count and retry delays must not be negative."
)

const (
	logVaultIdIsRequired                  = "Invalid vault config. Vault ID is required."
	logVaultUrlIsInvalid                  = "Invalid vault config. Vault URL is not a valid URL."
	logClusterIdIsRequired                = "Invalid vault config. Cluster ID is required."
	logDeprecatedFieldBaseVaultUrl        = "Deprecated: VaultConfig.BaseVaultURL is deprecated and will be removed in a future version. Use BaseVaultUrl instead."
	logMultipleTokenGenerationMeansPassed = "Invalid credentials. Only one of 'path', 'credentialsString', 'token' or 'apiKey' is allowed."
	logNoTokenGenerationMeansPassed       = "Invalid credentials. Any one of 'path', 'credentialsString', 'token' or 'apiKey' is required."
	logInvalidApiKey                      = "Invalid credentials. Api key is invalid."
	logEmptyRoles                         = "Invalid credentials. Roles can not be empty."
	logEmptyOrNullRoleInRoles             = "Invalid credentials. Role can not be null or empty in roles at index %v."
	logNegativeHttpConfig                 = "Invalid HTTP config. Timeout, retry count and retry delays must not be negative."
)
