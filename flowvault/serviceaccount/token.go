package serviceaccount

import (
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// The service-account flows are implemented once in the shared
// skyflow-go/common/auth package. These functions forward to it, configured
// with flowvault's SDK prefix, so errors and logs carry flowvault's version.

// GenerateBearerToken generates a bearer token using the credentials file.
func GenerateBearerToken(credentialsFilePath string, options fvcommon.BearerTokenOptions) (*fvcommon.TokenResponse, *fverrors.SkyflowError) {
	return fvhelpers.Auth().GenerateBearerToken(credentialsFilePath, options)
}

// GenerateBearerTokenFromCreds generates a bearer token from a credentials string.
func GenerateBearerTokenFromCreds(credentialsString string, options fvcommon.BearerTokenOptions) (*fvcommon.TokenResponse, *fverrors.SkyflowError) {
	return fvhelpers.Auth().GenerateBearerTokenFromCreds(credentialsString, options)
}

// GenerateSignedDataTokens generates signed data tokens.
func GenerateSignedDataTokens(credentialsFilePath string, options fvcommon.SignedDataTokensOptions) ([]fvcommon.SignedDataTokensResponse, *fverrors.SkyflowError) {
	return fvhelpers.Auth().GenerateSignedDataTokens(credentialsFilePath, options)
}

// GenerateSignedDataTokensFromCreds generates signed data tokens from a credentials string.
func GenerateSignedDataTokensFromCreds(credentialsString string, options fvcommon.SignedDataTokensOptions) ([]fvcommon.SignedDataTokensResponse, *fverrors.SkyflowError) {
	return fvhelpers.Auth().GenerateSignedDataTokensFromCreds(credentialsString, options)
}

// IsExpired reports whether the bearer token's exp claim is in the past.
func IsExpired(tokenString string) bool {
	return fvhelpers.Auth().IsExpired(tokenString)
}
