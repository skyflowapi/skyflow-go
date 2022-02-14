package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type TokenUtils struct {
	Token string
}

func (tokenUtils *TokenUtils) getBearerToken(tokenProvider common.TokenProvider) (string, *errors.SkyflowError) {

	if tokenUtils.Token != "" && !isTokenExpired(tokenUtils.Token) {
		return tokenUtils.Token, nil
	}
	token, err := tokenProvider()
	tokenUtils.Token = token
	if err != nil {
		return "", errors.NewSkyflowErrorWrap(errors.ErrorCodesEnum(errors.SdkErrorCode), err, messages.INVALID_BEARER_TOKEN)
	}
	return tokenUtils.Token, nil
}

func isTokenExpired(tokenString string) bool {

	newToken, errParse := jwt.ParseString(tokenString)

	if errParse != nil {
		return true
	}
	var claims jwt.StandardClaims
	errClaims := json.Unmarshal(newToken.RawClaims(), &claims)
	if errClaims != nil {
		return true
	}
	var expiryTime = claims.ExpiresAt
	currentTime := time.Now()
	fmt.Println(currentTime)
	fmt.Println(expiryTime)

	return expiryTime.Before(currentTime)
}
