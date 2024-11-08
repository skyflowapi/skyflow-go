package serviceaccount

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt"

	constants "skyflow-go/internal/constants"
	internalAuthApi "skyflow-go/internal/generated/auth"
	skyflowError "skyflow-go/utils/error"
	"skyflow-go/utils/logger"
)

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

type BearerTokenOptions struct {
	Ctx      string
	RoleIDs  []string
	LogLevel logger.LogLevel
}

type SignedDataTokensOptions struct {
	DataTokens []string
	TimeToLive int
	Ctx        string
	LogLevel   logger.LogLevel
}

type SignedDataTokensResponse struct {
	Token       string
	SignedToken string
}

var generateBearerTokenFunc = generateBearerToken

// Helper function to read and parse credentials from file
func parseCredentialsFile(credentialsFilePath string) (map[string]interface{}, *skyflowError.SkyflowError) {
	file, err := os.Open(credentialsFilePath)
	if err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to open credentials file")
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to read credential file")
	}
	var credKeys map[string]interface{}
	if err := json.Unmarshal(bytes, &credKeys); err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to parse credential file")
	}
	return credKeys, nil
}

// GenerateBearerToken Generate Bearer Token
func GenerateBearerToken(credentialsFilePath string, options BearerTokenOptions) (*TokenResponse, *skyflowError.SkyflowError) {
	credKeys, err := parseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to parse credential file")
	}

	token, err1 := generateBearerTokenFunc(credKeys, options)
	if err1 != nil {
		return nil, err1
	}

	return &TokenResponse{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
	}, nil
}

// GenerateBearerTokenFromCreds Generate Bearer Token from Credentials String
func GenerateBearerTokenFromCreds(credentials string, options BearerTokenOptions) (*TokenResponse, *skyflowError.SkyflowError) {
	var credKeys map[string]interface{}
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to parse credential string, wrong format given")
	}

	token, err1 := generateBearerTokenFunc(credKeys, options)
	if err1 != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.InvalidInput, err1.GetMessage())
	}

	return &TokenResponse{
		AccessToken: token.GetAccessToken(),
		TokenType:   token.GetTokenType(),
	}, nil
}

// GenerateSignedDataTokens Generate Signed Data Tokens
func GenerateSignedDataTokens(credentialsFilePath string, options SignedDataTokensOptions) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
	// validate data
	if credentialsFilePath == "" {
		return nil, skyflowError.NewSkyflowError("400", "credential path not provided")
	}
	if len(options.DataTokens) == 0 {
		return nil, skyflowError.NewSkyflowError("400", "No data tokens provided")
	}
	credKeys, err := parseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to parse credential file")
	}

	return getSignedDataTokens(credKeys, options)
}

func GenerateSignedDataTokensFromCreds(credentials string, options SignedDataTokensOptions) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
	var credKeys map[string]interface{}
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to parse credential file")
	}

	return getSignedDataTokens(credKeys, options)
}

// Generate and Sign Tokens
func getSignedDataTokens(credKeys map[string]interface{}, options SignedDataTokensOptions) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
	pvtKey, err := getPrivateKey(credKeys)
	if err != nil {
		return nil, err
	}

	clientID, tokenURI, keyID, err := getCredentialParams(credKeys)
	if err != nil {
		return nil, err
	}

	return generateSignedDataTokens(clientID, keyID, pvtKey, options, tokenURI)
}

// Helper for extracting credentials
func getCredentialParams(credKeys map[string]interface{}) (string, string, string, *skyflowError.SkyflowError) {
	clientID, ok := credKeys["clientID"].(string)
	tokenURI, ok2 := credKeys["tokenURI"].(string)
	keyID, ok3 := credKeys["keyID"].(string)
	if !ok || !ok2 || !ok3 {
		return "", "", "", skyflowError.NewSkyflowError("400", "Invalid credential parameters")
	}
	return clientID, tokenURI, keyID, nil
}

// Generate signed tokens
func generateSignedDataTokens(clientID, keyID string, pvtKey *rsa.PrivateKey, options SignedDataTokensOptions, tokenURI string) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
	var responseArray []SignedDataTokensResponse
	for _, token := range options.DataTokens {
		claims := jwt.MapClaims{
			"iss": "sdk",
			"key": keyID,
			"aud": tokenURI,
			"iat": time.Now().Unix(),
			"sub": clientID,
			"tok": token,
		}
		if options.TimeToLive > 0 {
			claims["exp"] = time.Now().Add(time.Duration(options.TimeToLive) * time.Second).Unix()
		} else {
			claims["exp"] = time.Now().Add(time.Duration(60) * time.Second).Unix()
		}
		if options.Ctx != "" {
			claims["ctx"] = options.Ctx
		}

		tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(pvtKey)
		if err != nil {
			return nil, skyflowError.NewSkyflowError("400", "Token signing failed")
		}
		responseArray = append(responseArray, SignedDataTokensResponse{Token: token, SignedToken: "signed_token_" + tokenString})
	}
	return responseArray, nil
}

func getPrivateKey(credKeys map[string]interface{}) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privateKeyStr, ok := credKeys["privateKey"].(string)
	if !ok {
		return nil, skyflowError.NewSkyflowError("400", "Missing private key")
	}
	return parsePrivateKey(privateKeyStr)
}

// Parse RSA Private Key from PEM
func parsePrivateKey(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privPem, _ := pem.Decode([]byte(pemKey))
	if privPem == nil {
		return nil, skyflowError.NewSkyflowError("400", "Invalid private key format")
	}
	if privPem.Type != "PRIVATE KEY" {
		return nil, skyflowError.NewSkyflowError("400", "Invalid private key type")
	}

	key, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err == nil {
		return key, nil
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(privPem.Bytes)
	if err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Invalid private key")
	}

	if privateKey, ok := parsedKey.(*rsa.PrivateKey); ok {
		return privateKey, nil
	}
	return nil, skyflowError.NewSkyflowError("400", "Invalid private key type")
}

// generate bearer token helper functions
func generateBearerToken(credKeys map[string]interface{}, options BearerTokenOptions) (*internalAuthApi.V1GetAuthTokenResponse, *skyflowError.SkyflowError) {
	privateKey := credKeys["privateKey"]
	if privateKey == nil {
		fmt.Println("privateKey is nil")
		return nil, skyflowError.NewSkyflowError("400", "privateKey is nil")
	}
	pvtKey, err1 := getPrivateKeyFromPem(privateKey.(string))
	if err1 != nil {
		return nil, err1
	}
	clientID, ok := credKeys["clientID"].(string)
	if !ok {
		fmt.Println("clientID is nil")
		return nil, skyflowError.NewSkyflowError("400", "clientID is nil")
	}
	tokenURI, ok1 := credKeys["tokenURI"].(string)
	if !ok1 {
		fmt.Println("tokenURI is nil")
		return nil, skyflowError.NewSkyflowError("400", "tokenURI is nil")
	}
	keyID, ok2 := credKeys["keyID"].(string)
	if !ok2 {
		return nil, skyflowError.NewSkyflowError("400", "keyID is nil")
	}

	signedUserJWT, e := getSignedBearerUserToken(clientID, keyID, tokenURI, pvtKey, options)
	if e != nil {
		return nil, e
	}
	// 1. config
	config := internalAuthApi.NewConfiguration()
	var err *skyflowError.SkyflowError
	config.Servers[0].URL, err = getBaseURL(tokenURI)
	if err != nil {
		return nil, skyflowError.NewSkyflowError("400", "Failed to get token URL")
	}
	// 2. client
	client := internalAuthApi.NewAPIClient(config)
	//3. auth api
	authApi := client.AuthenticationAPI.AuthenticationServiceGetAuthToken(context.TODO())
	// 4. request
	body := internalAuthApi.V1GetAuthTokenRequest{}
	body.SetGrantType(constants.GRANT_TYPE)
	body.SetAssertion(signedUserJWT)
	if len(options.RoleIDs) > 0 {
		body.SetScope(getScopeUsingRoles(options.RoleIDs))
	}
	// 5. send request
	res, r, err2 := authApi.Body(body).Execute()
	if err2 != nil {
		fmt.Println("error", err2, "https response=>", r)
		return nil, skyflowError.NewSkyflowError("400", "Error occurred")
	}
	fmt.Println("res->", *res.AccessToken)
	return res, nil
}
func getScopeUsingRoles(roles []string) string {
	scope := ""
	if roles != nil {
		for _, role := range roles {
			scope += " role:" + role
		}
	}
	fmt.Println("scoped", scope)
	return scope
}
func getBaseURL(urlStr string) (string, *skyflowError.SkyflowError) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil || parsedUrl.Scheme != "https" || parsedUrl.Host == "" {
		fmt.Println("error", err)
		return "", skyflowError.NewSkyflowError("400", "Invalid url") // return error if URL parsing fails
	}
	// Construct the base URL using the scheme and host
	baseURL := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	return baseURL, nil
}
func getSignedBearerUserToken(clientID, keyID, tokenURI string, pvtKey *rsa.PrivateKey, options BearerTokenOptions) (string, *skyflowError.SkyflowError) {

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": clientID,
		"key": keyID,
		"aud": tokenURI,
		"sub": clientID,
		"exp": time.Now().Add(60 * time.Minute).Unix(),
	})
	if options.Ctx != "" {
		token.Claims.(jwt.MapClaims)["ctx"] = options.Ctx
	}
	var err error
	signedToken, err := token.SignedString(pvtKey)
	if err != nil {
		return "", skyflowError.NewSkyflowError("500", "Failed to sign token")
	}
	return signedToken, nil
}
func getPrivateKeyFromPem(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	var err error
	privPem, _ := pem.Decode([]byte(pemKey))
	if privPem == nil {
		return nil, skyflowError.NewSkyflowError("400", "Unable to decode the RSA private PEM")
	}
	if privPem.Type != "PRIVATE KEY" {
		return nil, skyflowError.NewSkyflowErrorf(skyflowError.InvalidInput, "RSA private key is of the wrong type Pem Type: %s", privPem.Type)
	}
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPem.Bytes); err != nil {
			return nil, skyflowError.NewSkyflowError("400", err.Error())
		}
	}
	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, skyflowError.NewSkyflowError(skyflowError.InvalidInput, "Unable to retrieve RSA private key")
	}

	return privateKey, nil
}
func IsExpired(tokenString string) bool {
	if tokenString == "" {
		//logger.Info(fmt.Sprintf(messages.EMPTY_BEARER_TOKEN, "ServiceAccountUtil"))
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
		fmt.Println("EXPIRE_BEARER_TOKEN")
		//logger.Info(fmt.Sprintf(messages.EXPIRE_BEARER_TOKEN, "ServiceAccountUtil"))
	}
	return expiryTime.Before(currentTime)
}
