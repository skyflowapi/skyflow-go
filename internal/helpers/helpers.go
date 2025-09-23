package helpers

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/skyflowapi/skyflow-go/v2/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"

	"github.com/golang-jwt/jwt"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	internal "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	internalAuthApi "github.com/skyflowapi/skyflow-go/v2/internal/generated/authentication"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

// Helper function to read and parse credentials from file
func ParseCredentialsFile(credentialsFilePath string) (map[string]interface{}, *skyflowError.SkyflowError) {
	file, err := os.Open(credentialsFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf(logs.FILE_NOT_FOUND, credentialsFilePath))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.FILE_NOT_FOUND, credentialsFilePath))
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Error(fmt.Sprintf(logs.INVALID_INPUT_FILE, credentialsFilePath))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(logs.INVALID_INPUT_FILE, credentialsFilePath))
	}
	var credKeys map[string]interface{}
	if err := json.Unmarshal(bytes, &credKeys); err != nil {
		logger.Error(logs.NOT_A_VALID_JSON)
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
		logger.Error(logs.INVALID_CREDENTIALS_FILE_FORMAT)
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
			logger.Error(logs.PARSE_JWT_PAYLOAD)
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", err))
		}
		responseArray = append(responseArray, SignedDataTokensResponse{Token: token, SignedToken: "signed_token_" + tokenString})
	}
	logger.Info(logs.GENERATE_SIGNED_DATA_TOKEN_SUCCESS)
	return responseArray, nil
}

func GetPrivateKey(credKeys map[string]interface{}) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privateKeyStr, ok := credKeys["privateKey"].(string)
	if !ok {
		logger.Error(logs.PRIVATE_KEY_NOT_FOUND)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_PRIVATE_KEY)
	}
	return ParsePrivateKey(privateKeyStr)
}

// ParsePrivateKey Parse RSA Private Key from PEM
func ParsePrivateKey(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privPem, _ := pem.Decode([]byte(pemKey))
	if privPem == nil {
		logger.Error(logs.JWT_INVALID_FORMAT)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}
	if privPem.Type != "PRIVATE KEY" {
		logger.Error(fmt.Sprintf(logs.PRIVATE_KEY_TYPE, privPem.Type))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}

	key, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
	if err == nil {
		return key, nil
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(privPem.Bytes)
	if err != nil {
		logger.Error(logs.INVALID_ALGORITHM)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_ALGORITHM)
	}

	if privateKey, ok := parsedKey.(*rsa.PrivateKey); ok {
		return privateKey, nil
	}
	logger.Error(logs.INVALID_KEY_SPEC)
	return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_KEY_SPEC)
}

var GetBaseURLHelper = GetBaseURL

// GenerateBearerTokenHelper  helper functions
func GenerateBearerTokenHelper(credKeys map[string]interface{}, options BearerTokenOptions) (*internal.V1GetAuthTokenResponse, *skyflowError.SkyflowError) {
	privateKey := credKeys["privateKey"]
	if privateKey == nil {
		logger.Error(fmt.Sprintf(logs.PRIVATE_KEY_NOT_FOUND))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_PRIVATE_KEY)
	}
	pvtKey, err1 := GetPrivateKeyFromPem(privateKey.(string))
	if err1 != nil {
		return nil, err1
	}
	clientID, ok := credKeys["clientID"].(string)
	if !ok {
		logger.Error(fmt.Sprintf(logs.CLIENT_ID_NOT_FOUND))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_CLIENT_ID)
	}
	tokenURI, ok1 := credKeys["tokenURI"].(string)
	if !ok1 {
		logger.Error(fmt.Sprintf(logs.TOKEN_URI_NOT_FOUND))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_TOKEN_URI)
	}
	keyID, ok2 := credKeys["keyID"].(string)
	if !ok2 {
		logger.Error(fmt.Sprintf(logs.KEY_ID_NOT_FOUND))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_KEY_ID)
	}

	signedUserJWT, e := GetSignedBearerUserToken(clientID, keyID, tokenURI, pvtKey, options)
	if e != nil {
		return nil, e
	}
	// 1. config
	//config := internal.V1GetAuthTokenRequest{}
	var err *skyflowError.SkyflowError
	var url string
	url, err = GetBaseURLHelper(tokenURI)
	if err != nil {
		return nil, err
	}
	// 2. client
	//client := internalAuthApi.NewAPIClient(config)
	client := internalAuthApi.NewClient(option.WithBaseURL(url))
	//3. request
	body := internal.V1GetAuthTokenRequest{}
	body.GrantType = constants.GRANT_TYPE
	body.Assertion = signedUserJWT
	if len(options.RoleIDs) > 0 {
		var roles []*string
		for _, roleID := range options.RoleIDs {
			roles = append(roles, &roleID)
		}
		roleString := GetScopeUsingRoles(roles)
		body.Scope = &roleString
	}
	// 5. send request
	authApi, apiErr := client.WithRawResponse.AuthenticationServiceGetAuthToken(context.Background(), &body)
	if apiErr != nil {
		header, _ := GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}
	return authApi.Body, nil
}
func GetScopeUsingRoles(roles []*string) string {
	scope := ""
	if roles != nil {
		for _, role := range roles {
			scope += " role:" + *role
		}
	}
	return scope
}
func GetBaseURL(urlStr string) (string, *skyflowError.SkyflowError) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil || parsedUrl.Scheme != "https" || parsedUrl.Host == "" {
		logger.Error(logs.INVALID_TOKEN_URI)
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
		logger.Error(fmt.Sprintf("%s", "unable to parse jwt payload"))
		return "", skyflowError.NewSkyflowError(skyflowError.SERVER, fmt.Sprintf(skyflowError.UNKNOWN_ERROR, err))
	}
	return signedToken, nil
}
func GetPrivateKeyFromPem(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	var err error
	privPem, _ := pem.Decode([]byte(pemKey))
	if privPem == nil {
		logger.Error(fmt.Sprintf(logs.JWT_INVALID_FORMAT))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}
	if privPem.Type != "PRIVATE KEY" {
		logger.Error(logs.JWT_INVALID_FORMAT)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.JWT_INVALID_FORMAT)
	}
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPem.Bytes); err != nil {
			logger.Error(logs.INVALID_KEY_SPEC)
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_ALGORITHM)
		}
	}
	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		logger.Error(logs.INVALID_KEY_SPEC)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_KEY_SPEC)
	}

	return privateKey, nil
}

func CreateJsonMetadata() string {
	// Create a map to hold the key-value pairs
	data := map[string]string{
		"sdk_name_version":        fmt.Sprintf("%s@%s", constants.SDK_NAME, constants.SDK_VERSION),
		"sdk_client_device_model": string(runtime.GOOS),
		"sdk_client_os_details":   fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		"sdk_runtime_details":     runtime.Version(),
	}

	// Marshal the map into JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Debug("failed for marshalling json data in createJSONMetadata()")
		return ""
	}
	return string(jsonData)
}

func Float64Ptr(v float64) *float64 {
	return &v
}

func GetHeader(err error) (http.Header, bool) {
	if err == nil {
		return http.Header{}, false
	}
	var apiError *core.APIError
	if errors.As(err, &apiError) {
		return apiError.Header, true
	}
	return http.Header{}, false
}

func GetSkyflowID(data map[string]interface{}) (string, bool) {
	if id, ok := data["skyflow_id"].(string); ok {
		return id, true
	}
	return "", false
}