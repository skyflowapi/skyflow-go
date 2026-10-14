package auth

// Message bodies without the SDK prefix. The prefix ("Skyflow Go SDK v2.1.0",
// "[ Skyflow Go SDK v2.1.0 ] ", ...) is supplied by the SDK that owns the
// Service, so v2 and flowvault keep their own version string in every error
// and log line while sharing one implementation.

// Error messages: prefixed with Config.SDKPrefix.
const (
	errEmptyCredentialFilePath      = " Initialization failed. Invalid credentials. Credentials file path must not be empty."
	errEmptyCredentialsString       = " Initialization failed. Invalid credentials. Credentials string must not be empty."
	errFileNotFound                 = " Initialization failed. Credential file not found at %s. Verify the file path."
	errCredentialsStringInvalidJSON = " Initialization failed. Credentials string is not in valid JSON format. Verify the credentials string contents."
	errInvalidCredentials           = " Initialization failed. Invalid credentials provided. Specify valid credentials."
	errMissingPrivateKey            = " Initialization failed. Unable to read private key in credentials. Verify your private key."
	errMissingClientId              = " Initialization failed. Unable to read client ID in credentials. Verify your client ID."
	errMissingKeyId                 = " Initialization failed. Unable to read key ID in credentials. Verify your key ID."
	errMissingTokenUri              = " Initialization failed. Unable to read token URI in credentials. Verify your token URI."
	errInvalidTokenUri              = " Initialization failed. Token URI in not a valid URL in credentials. Verify your token URI."
	errJwtInvalidFormat             = " Initialization failed. Invalid private key format. Verify your credentials."
	errInvalidAlgorithm             = " Initialization failed. Invalid algorithm to parse private key. Specify valid algorithm."
	errInvalidKeySpec               = " Initialization failed. Unable to parse RSA private key. Verify your credentials."
	errEmptyTokensDetokenize        = " Validation error. Invalid data tokens. Specify at least one data token."
	errErrorOccurred                = " API error. Error occurred."
	errUnknownError                 = " Error occurred. %s"
	errInvalidCtxType               = " Initialization failed. Invalid ctx type. Specify ctx as a string, number, boolean, or a map[string]interface{}."
	errInvalidCtxMapKey             = " Initialization failed. Invalid key '%s' in ctx map. Keys must contain only alphanumeric characters and underscores."
)

// Log messages: prefixed with Config.LogPrefix.
const (
	logEmptyBearerToken                  = "BearerToken is Empty"
	logBearerTokenExpired                = "Bearer Token provided is either invalid or has expired."
	logGenerateBearerTokenTriggered      = "GenerateBearerToken is triggered"
	logGenerateBearerTokenSuccess        = "BearerToken is generated"
	logGenerateSignedDataTokenSuccess    = "Signed Data tokens are generated"
	logGenerateSignedDataTokensTriggered = "generateSignedDataTokens is triggered"
	logDeprecatedFieldRoleIds            = "Deprecated: BearerTokenOptions.RoleIDs is deprecated and will be removed in a future version. Use RoleIds instead."
	logClientIdNotFound                  = "Invalid credentials. Client ID cannot be empty."
	logTokenUriNotFound                  = "Invalid credentials. Token URI cannot be empty."
	logKeyIdNotFound                     = "Invalid credentials. Key ID cannot be empty."
	logPrivateKeyNotFound                = "Invalid credentials. Private key cannot be empty."
	logNotAValidJSON                     = "Credentials is not in valid JSON format. Verify the credentials."
	logFileNotFound                      = "Credential file not found at %s. Verify the file path."
	logInvalidInputFile                  = "Unable to read credentials - file %s."
	logEmptyCredentialsPath              = "Invalid credentials. Credentials path can not be empty."
	logInvalidCredentialsStringFormat    = "Credentials string is not in a valid JSON string format."
	logInvalidTokenUri                   = "Invalid value for token URI in credentials."
	logJwtInvalidFormat                  = "Private key is not in a valid format."
	logInvalidAlgorithm                  = "Algorithm for parsing private key is invalid or does not exist."
	logInvalidKeySpec                    = "Unable to parse RSA private key."
	logBearerTokenRejected               = "Bearer token request resulted in failure."
	logPrivateKeyType                    = "RSA private key is of the wrong type Pem Type: %s"
	logParseJwtPayload                   = "Unable to parse jwt payload"
)
