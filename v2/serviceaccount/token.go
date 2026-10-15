package serviceaccount

import (
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

// The service-account flows are implemented once in the shared
// skyflow-go/common/auth package. These functions keep v2's public API and
// forward to that implementation, configured with v2's SDK prefix.

// GenerateBearerToken Generate Bearer Token
func GenerateBearerToken(credentialsFilePath string, options common.BearerTokenOptions) (*common.TokenResponse, *skyflowError.SkyflowError) {
	return helpers.Auth().GenerateBearerToken(credentialsFilePath, options)
}

// GenerateBearerTokenFromCreds Generate Bearer Token from Credentials String
func GenerateBearerTokenFromCreds(credentials string, options common.BearerTokenOptions) (*common.TokenResponse, *skyflowError.SkyflowError) {
	return helpers.Auth().GenerateBearerTokenFromCreds(credentials, options)
}

// GenerateSignedDataTokens Generate Signed Data Tokens
func GenerateSignedDataTokens(credentialsFilePath string, options common.SignedDataTokensOptions) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	return helpers.Auth().GenerateSignedDataTokens(credentialsFilePath, options)
}

// GenerateSignedDataTokensFromCreds Generate Signed Data Tokens from Credentials String
func GenerateSignedDataTokensFromCreds(credentials string, options common.SignedDataTokensOptions) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	return helpers.Auth().GenerateSignedDataTokensFromCreds(credentials, options)
}

// IsExpired reports whether the bearer token's exp claim is in the past.
func IsExpired(tokenString string) bool {
	return helpers.Auth().IsExpired(tokenString)
}
