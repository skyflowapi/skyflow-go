package serviceaccount

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	. "skyflow-go/v2/internal/helpers"
	. "skyflow-go/v2/utils/common"
	skyflowError "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
	. "skyflow-go/v2/utils/messages"
	"time"
)

// GenerateBearerToken Generate Bearer Token
func GenerateBearerToken(credentialsFilePath string, options BearerTokenOptions) (*TokenResponse, *skyflowError.SkyflowError) {
	logger.Info(GENERATE_BEARER_TOKEN_TRIGGERED)
	credKeys, err := ParseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	token, err1 := GenerateBearerTokenHelper(credKeys, options)
	if err1 != nil {
		return nil, err1
	}

	logger.Info(GENERATE_BEARER_TOKEN_SUCCESS)
	return &TokenResponse{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
	}, nil
}

// GenerateBearerTokenFromCreds Generate Bearer Token from Credentials String
func GenerateBearerTokenFromCreds(credentials string, options BearerTokenOptions) (*TokenResponse, *skyflowError.SkyflowError) {
	var credKeys map[string]interface{}
	logger.Info(GENERATE_BEARER_TOKEN_TRIGGERED)
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		logger.Error(INVALID_CREDENTIALS_STRING_FORMAT)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CREDENTIALS)
	}

	token, err1 := GenerateBearerTokenHelper(credKeys, options)
	if err1 != nil {
		return nil, err1
	}
	logger.Info(GENERATE_BEARER_TOKEN_SUCCESS)
	return &TokenResponse{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
	}, nil
}

// GenerateSignedDataTokens Generate Signed Data Tokens
func GenerateSignedDataTokens(credentialsFilePath string, options SignedDataTokensOptions) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
	// validate data
	logger.Error(GENERATE_SIGNED_DATA_TOKENS_TRIGGERED)
	if credentialsFilePath == "" {
		logger.Error(EMPTY_CREDENTIALS_PATH)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CREDENTIAL_FILE_PATH)
	}
	if len(options.DataTokens) == 0 {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS_DETOKENIZE)
	}
	credKeys, err := ParseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	return GetSignedDataTokens(credKeys, options)
}

func GenerateSignedDataTokensFromCreds(credentials string, options SignedDataTokensOptions) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
	var credKeys map[string]interface{}
	logger.Error(GENERATE_SIGNED_DATA_TOKENS_TRIGGERED)
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		logger.Error(INVALID_CREDENTIALS_STRING_FORMAT)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CREDENTIALS)
	}

	return GetSignedDataTokens(credKeys, options)
}

func IsExpired(tokenString string) bool {
	if tokenString == "" {
		logger.Info(fmt.Sprintf(EMPTY_BEARER_TOKEN))
		return true
	}
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return true
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return true
	}
	var expiryTime time.Time
	switch exp := claims["exp"].(type) {
	case float64:
		expiryTime = time.Unix(int64(exp), 0)
	case json.Number:
		v, _ := exp.Int64()
		expiryTime = time.Unix(v, 0)
	}
	currentTime := time.Now()
	if expiryTime.Before(currentTime) {
		logger.Info(BEARER_TOKEN_EXPIRED)
	}
	return expiryTime.Before(currentTime)
}
