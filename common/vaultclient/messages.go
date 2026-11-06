package vaultclient

// CredentialsEnv names the environment variable used when neither the vault nor the client
// has credentials.
const CredentialsEnv = "SKYFLOW_CREDENTIALS"

// Message bodies without the SDK prefix; Profile prepends the owning SDK's prefix (same
// pattern as auth and config).

// Error messages: prefixed with Profile.SDKPrefix.
const (
	errEmptyCredentials     = " Validation error. Invalid credentials. Specify a valid credentials."
	errBearerTokenExpired   = " Validation error. Bearer token is expired. Specify a valid bearer token."
	errBearerTokenCancelled = " Request cancelled. The context ended before a bearer token was available."
	errInvalidHeaderKey     = " Validation error. %s: header key %q is not supported. Use x-skyflow-account-id, x-skyflow-account-name or x-request-id."
	errEmptyHeaderValue     = " Validation error. %s: header %q must have a value."
)

// Log messages: prefixed with Profile.LogPrefix.
const (
	logBearerTokenExpired           = "Bearer token is expired."
	logGenerateBearerTokenTriggered = "Generating a new bearer token."
	logReuseBearerToken             = "Reusing the cached bearer token."
	logEmptyCredentials             = "No credentials found in the vault config, client config or SKYFLOW_CREDENTIALS (environment or .env file)."
	logBearerTokenCancelled         = "The context ended before a bearer token was available."

	// The 401 retry; the same lines as the Java SDK.
	logUnauthorizedRegenerating       = "Bearer token rejected with 401 and is past expiry; regenerating once and resending the request."
	logUnauthorizedReusing            = "Bearer token rejected with 401; reusing the token another request already regenerated."
	logUnauthorizedNotDue             = "Bearer token rejected with 401 but is not near expiry; not retrying."
	logUnauthorizedAlreadyFailed      = "Bearer token rejected with 401; regeneration for this token already failed, not retrying."
	logUnauthorizedRegenerationFailed = "Bearer token regeneration after 401 failed; returning the 401 response."
)
