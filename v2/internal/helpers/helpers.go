package helpers

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
	constants "skyflow-go/v2/internal/constants"
	internalAuthApi "skyflow-go/v2/internal/generated/auth"
	. "skyflow-go/v2/utils/common"
	skyflowError "skyflow-go/v2/utils/error"
	"time"

	"github.com/golang-jwt/jwt"
)

// Helper function to read and parse credentials from file
func ParseCredentialsFile(credentialsFilePath string) (map[string]interface{}, *skyflowError.SkyflowError) {
	file, err := os.Open(credentialsFilePath)
	if err != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.FILE_NOT_FOUND)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.UNKNOWN_ERROR, err))
	}
	var credKeys map[string]interface{}
	if err := json.Unmarshal(bytes, &credKeys); err != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.CREDENTIALS_STRING_INVALID_JSON)
	}
	return credKeys, nil
}

// Generate and Sign Tokens
func GetSignedDataTokens(credKeys map[string]interface{}, options SignedDataTokensOptions) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
	pvtKey, err := GetPrivateKey(credKeys)
	if err != nil {
		return nil, err
	}

	clientID, tokenURI, keyID, err := GetCredentialParams(credKeys)
	if err != nil {
		return nil, err
	}

	return GenerateSignedDataTokensHelper(clientID, keyID, pvtKey, options, tokenURI)
}

// Helper for extracting credentials
func GetCredentialParams(credKeys map[string]interface{}) (string, string, string, *skyflowError.SkyflowError) {
	clientID, ok := credKeys["clientID"].(string)
	tokenURI, ok2 := credKeys["tokenURI"].(string)
	keyID, ok3 := credKeys["keyID"].(string)
	if !ok || !ok2 || !ok3 {
		return "", "", "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CREDENTIALS)
	}
	return clientID, tokenURI, keyID, nil
}

// Generate signed tokens
func GenerateSignedDataTokensHelper(clientID, keyID string, pvtKey *rsa.PrivateKey, options SignedDataTokensOptions, tokenURI string) ([]SignedDataTokensResponse, *skyflowError.SkyflowError) {
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
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", err))
		}
		responseArray = append(responseArray, SignedDataTokensResponse{Token: token, SignedToken: "signed_token_" + tokenString})
	}
	return responseArray, nil
}

func GetPrivateKey(credKeys map[string]interface{}) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privateKeyStr, ok := credKeys["privateKey"].(string)
	if !ok {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_PRIVATE_KEY)
	}
	return ParsePrivateKey(privateKeyStr)
}

// ParsePrivateKey Parse RSA Private Key from PEM
func ParsePrivateKey(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privPem, _ := pem.Decode([]byte(pemKey))
	if privPem == nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}
	if privPem.Type != "PRIVATE KEY" {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}

	key, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err == nil {
		return key, nil
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(privPem.Bytes)
	if err != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_ALGORITHM)
	}

	if privateKey, ok := parsedKey.(*rsa.PrivateKey); ok {
		return privateKey, nil
	}
	return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_KEY_SPEC)
}

var GetBaseURLHelper = GetBaseURL

// GenerateBearerTokenHelper  helper functions
func GenerateBearerTokenHelper(credKeys map[string]interface{}, options BearerTokenOptions) (*internalAuthApi.V1GetAuthTokenResponse, *skyflowError.SkyflowError) {
	privateKey := credKeys["privateKey"]
	if privateKey == nil {
		fmt.Println("privateKey is nil")
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_PRIVATE_KEY)
	}
	pvtKey, err1 := GetPrivateKeyFromPem(privateKey.(string))
	if err1 != nil {
		return nil, err1
	}
	clientID, ok := credKeys["clientID"].(string)
	if !ok {
		fmt.Println("clientID is nil")
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_CLIENT_ID)
	}
	tokenURI, ok1 := credKeys["tokenURI"].(string)
	if !ok1 {
		fmt.Println("tokenURI is nil")
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_TOKEN_URI)
	}
	keyID, ok2 := credKeys["keyID"].(string)
	if !ok2 {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_KEY_ID)
	}

	signedUserJWT, e := GetSignedBearerUserToken(clientID, keyID, tokenURI, pvtKey, options)
	if e != nil {
		return nil, e
	}
	// 1. config
	config := internalAuthApi.NewConfiguration()
	var err *skyflowError.SkyflowError
	config.Servers[0].URL, err = GetBaseURLHelper(tokenURI)
	if err != nil {
		return nil, err
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
		body.SetScope(GetScopeUsingRoles(options.RoleIDs))
	}
	// 5. send request
	res, r, err2 := authApi.Body(body).Execute()
	if err2 != nil && r != nil {
		return nil, skyflowError.SkyflowApiError(*r)
	} else if err2 != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.UNKNOWN_ERROR, err2))
	}
	return res, nil
}
func GetScopeUsingRoles(roles []string) string {
	scope := ""
	if roles != nil {
		for _, role := range roles {
			scope += " role:" + role
		}
	}
	return scope
}
func GetBaseURL(urlStr string) (string, *skyflowError.SkyflowError) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil || parsedUrl.Scheme != "https" || parsedUrl.Host == "" {
		return "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TOKEN_URI) // return error if URL parsing fails
	}
	// Construct the base URL using the scheme and host
	baseURL := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	return baseURL, nil
}
func GetSignedBearerUserToken(clientID, keyID, tokenURI string, pvtKey *rsa.PrivateKey, options BearerTokenOptions) (string, *skyflowError.SkyflowError) {

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
		return "", skyflowError.NewSkyflowError(skyflowError.SERVER, fmt.Sprintf(skyflowError.UNKNOWN_ERROR, err))
	}
	return signedToken, nil
}
func GetPrivateKeyFromPem(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	var err error
	privPem, _ := pem.Decode([]byte(pemKey))
	if privPem == nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}
	if privPem.Type != "PRIVATE KEY" {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPem.Bytes); err != nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_ALGORITHM)
		}
	}
	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_KEY_SPEC)
	}

	return privateKey, nil
}
