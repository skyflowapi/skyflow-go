package serviceaccount

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	"github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

// GenerateBearerToken Generate Bearer Token
func GenerateBearerToken(credentialsFilePath string, options common.BearerTokenOptions) (*common.TokenResponse, *skyflowError.SkyflowError) {
	logger.Info(logs.GENERATE_BEARER_TOKEN_TRIGGERED)
	if credentialsFilePath == "" {
		logger.Error(logs.EMPTY_CREDENTIALS_PATH)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CREDENTIAL_FILE_PATH)
	}
	credKeys, err := helpers.ParseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	token, err1 := helpers.GenerateBearerTokenHelper(credKeys, options)
	if err1 != nil {
		return nil, err1
	}

	logger.Info(logs.GENERATE_BEARER_TOKEN_SUCCESS)
	return &common.TokenResponse{
		AccessToken: *token.GetAccessToken(),
		TokenType:   *token.GetTokenType(),
	}, nil
}

// GenerateBearerTokenFromCreds Generate Bearer Token from Credentials String
func GenerateBearerTokenFromCreds(credentials string, options common.BearerTokenOptions) (*common.TokenResponse, *skyflowError.SkyflowError) {
	var credKeys map[string]interface{}
	logger.Info(logs.GENERATE_BEARER_TOKEN_TRIGGERED)
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		logger.Error(logs.INVALID_CREDENTIALS_STRING_FORMAT)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CREDENTIALS)
	}

	token, err1 := helpers.GenerateBearerTokenHelper(credKeys, options)
	if err1 != nil {
		return nil, err1
	}
	logger.Info(logs.GENERATE_BEARER_TOKEN_SUCCESS)
	return &common.TokenResponse{
		AccessToken: *token.GetAccessToken(),
		TokenType:   *token.GetTokenType(),
	}, nil
}

// GenerateSignedDataTokens Generate Signed Data Tokens
func GenerateSignedDataTokens(credentialsFilePath string, options common.SignedDataTokensOptions) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	// validate data
	logger.Info(logs.GENERATE_SIGNED_DATA_TOKENS_TRIGGERED)
	if credentialsFilePath == "" {
		logger.Error(logs.EMPTY_CREDENTIALS_PATH)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CREDENTIAL_FILE_PATH)
	}
	if len(options.DataTokens) == 0 {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS_DETOKENIZE)
	}
	credKeys, err := helpers.ParseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	return helpers.GetSignedDataTokens(credKeys, options)
}

func GenerateSignedDataTokensFromCreds(credentials string, options common.SignedDataTokensOptions) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	if credentials == "" {
		logger.Error(logs.INVALID_CREDENTIALS_STRING_FORMAT)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CREDENTIALS_STRING)
	}
	var credKeys map[string]interface{}
	logger.Info(logs.GENERATE_SIGNED_DATA_TOKENS_TRIGGERED)
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		logger.Error(logs.INVALID_CREDENTIALS_STRING_FORMAT)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CREDENTIALS)
	}

	return helpers.GetSignedDataTokens(credKeys, options)
}

func IsExpired(tokenString string) bool {
	if tokenString == "" {
		logger.Info(fmt.Sprintf(logs.EMPTY_BEARER_TOKEN))
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
		logger.Info(logs.BEARER_TOKEN_EXPIRED)
	}
	return expiryTime.Before(currentTime)
}
