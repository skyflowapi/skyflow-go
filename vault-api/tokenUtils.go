package vaultapi

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
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
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})

	if err != nil {
		log.Fatal(err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Fatalf("Can't convert token's claims to standard claims")
	}

	var expiryTime time.Time
	switch exp := claims["exp"].(type) {
	case float64:
		fmt.Println("float64..")
		expiryTime = time.Unix(int64(exp), 0)
	case json.Number:
		fmt.Println("json..")
		v, _ := exp.Int64()
		expiryTime = time.Unix(v, 0)
	}

	currentTime := time.Now()
	fmt.Println(currentTime)
	fmt.Println(expiryTime)

	return expiryTime.Before(currentTime)
}
