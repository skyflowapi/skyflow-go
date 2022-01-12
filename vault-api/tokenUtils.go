package vaultapi

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v3"
)

type TokenUtils struct {
	Token string
}

func (tokenUtils *TokenUtils) getBearerToken(tokenProvider TokenProvider) string {

	if tokenUtils.Token != "" && !isTokenExpired(tokenUtils.Token) {
		return tokenUtils.Token
	}
	tokenUtils.Token = tokenProvider()
	return tokenUtils.Token
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
