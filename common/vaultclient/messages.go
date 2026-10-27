package vaultclient

// CredentialsEnv names the environment variable used when neither the vault nor the client
// has credentials.
const CredentialsEnv = "SKYFLOW_CREDENTIALS"

// Message bodies without the SDK prefix; Profile prepends the owning SDK's prefix (same
// pattern as auth and config).

// Error messages: prefixed with Profile.SDKPrefix.
const (
	errEmptyCredentials   = " Validation error. Invalid credentials. Specify a valid credentials."
	errBearerTokenExpired = " Validation error. Bearer token is expired. Specify a valid bearer token."
	errInvalidHeaderKey   = " Validation error. %s: header key %q is not supported. Use x-skyflow-account-id, x-skyflow-account-name or x-request-id."
	errEmptyHeaderValue   = " Validation error. %s: header %q must have a value."
)

// Log messages: prefixed with Profile.LogPrefix.
const (
	logBearerTokenExpired           = "Bearer token is expired."
	logGenerateBearerTokenTriggered = "Generating a new bearer token."
	logReuseBearerToken             = "Reusing the cached bearer token."
	logEmptyCredentials             = "No credentials found in the vault config, client config or SKYFLOW_CREDENTIALS."
)
