package helpers

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/skyflowapi/skyflow-go/v2/internal/generated/core"

	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	"github.com/golang-jwt/jwt"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	internal "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	internalAuthApi "github.com/skyflowapi/skyflow-go/v2/internal/generated/authentication"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	common "github.com/skyflowapi/skyflow-go/v2/utils/common"
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

// SetTokenMode sets the tokenization mode in the request body.
func SetTokenMode(tokenMode common.BYOT) (*vaultapis.V1Byot, error) {
	var tokenModes vaultapis.V1Byot
	var tokenError error
	switch tokenMode {
	case common.ENABLE_STRICT:
		tokenModes, tokenError = vaultapis.NewV1ByotFromString(string(common.ENABLE_STRICT))
	case common.ENABLE:
		tokenModes, tokenError = vaultapis.NewV1ByotFromString(string(common.ENABLE))
	default:
		tokenModes, tokenError = vaultapis.NewV1ByotFromString(string(common.DISABLE))
	}
	if tokenError != nil {
		return nil, tokenError
	}
	return &tokenModes, nil
}
func GetFormattedGetRecord(record vaultapis.V1FieldRecords) map[string]interface{} {
	getRecord := make(map[string]interface{})
	var sourceMap map[string]interface{}

	// Decide whether to use Tokens or Fields
	if record.Tokens != nil {
		sourceMap = record.Tokens
	} else {
		sourceMap = record.Fields
	}

	// Copy elements from sourceMap to getRecord
	if sourceMap != nil {
		for key, value := range sourceMap {
			getRecord[key] = value
		}
	}

	return getRecord
}
func GetDetokenizePayload(request common.DetokenizeRequest, options common.DetokenizeOptions) vaultapis.V1DetokenizePayload {
	payload := vaultapis.V1DetokenizePayload{}
	payload.ContinueOnError = &options.ContinueOnError
	var reqArray []*vaultapis.V1DetokenizeRecordRequest

	for index := range request.DetokenizeData {
		req := vaultapis.V1DetokenizeRecordRequest{}
		req.Token = &request.DetokenizeData[index].Token
		if request.DetokenizeData[index].RedactionType != "" {
			redaction, _ := vaultapis.NewRedactionEnumRedactionFromString(string(request.DetokenizeData[index].RedactionType))
			req.Redaction = &redaction
		} else {
			redaction, _ := vaultapis.NewRedactionEnumRedactionFromString(string(common.DEFAULT))
			req.Redaction = &redaction
		}
		reqArray = append(reqArray, &req)
	}
	if len(reqArray) > 0 {
		payload.DetokenizationParameters = reqArray
	}
	return payload
}
func GetFormattedBatchInsertRecord(record interface{}, requestIndex int) (map[string]interface{}, *skyflowError.SkyflowError) {
	insertRecord := make(map[string]interface{})
	// Convert the record to JSON and unmarshal it
	jsonData, err := json.Marshal(record)
	if err != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_RESPONSE)
	}

	var bodyObject map[string]interface{}
	if err := json.Unmarshal(jsonData, &bodyObject); err != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_RESPONSE)
	}

	// Extract relevant data from "Body"
	body, bodyExists := bodyObject["Body"].(map[string]interface{})
	if !bodyExists {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_RESPONSE)
	}

	// Handle extracted data
	if records, ok := body["records"].([]interface{}); ok {
		for _, rec := range records {
			recordObject, isMap := rec.(map[string]interface{})
			if !isMap {
				continue
			}
			if skyflowID, exists := recordObject["skyflow_id"].(string); exists {
				insertRecord["skyflow_id"] = skyflowID
			}
			if tokens, exists := recordObject["tokens"].(map[string]interface{}); exists {
				for key, value := range tokens {
					insertRecord[key] = value
				}
			}
		}
	}

	if errorField, exists := body["error"].(string); exists {
		insertRecord["error"] = errorField
	}

	insertRecord["request_index"] = requestIndex
	return insertRecord, nil
}
func GetFormattedBulkInsertRecord(record vaultapis.V1RecordMetaProperties) map[string]interface{} {
	insertRecord := make(map[string]interface{})
	insertRecord["skyflow_id"] = *record.GetSkyflowId()

	tokensMap := record.GetTokens()
	if len(tokensMap) > 0 {
		for key, value := range tokensMap {
			insertRecord[key] = value
		}
	}
	return insertRecord
}
func GetFormattedQueryRecord(record vaultapis.V1FieldRecords) map[string]interface{} {
	queryRecord := make(map[string]interface{})
	if record.Fields != nil {
		for key, value := range record.Fields {
			queryRecord[key] = value
		}
	}
	return queryRecord
}
func GetFormattedUpdateRecord(record vaultapis.V1UpdateRecordResponse) map[string]interface{} {
	updateTokens := make(map[string]interface{})

	// Check if tokens are not nil
	if record.Tokens != nil {
		// Iterate through the map and populate updateTokens
		for key, value := range record.Tokens {
			updateTokens[key] = value
		}
	}

	return updateTokens
}

// CreateInsertBulkBodyRequest createInsertBodyRequest generates the request body for bulk inserts.
func CreateInsertBulkBodyRequest(request *common.InsertRequest, options *common.InsertOptions) (*vaultapis.RecordServiceInsertRecordBody, *skyflowError.SkyflowError) {
	var records []*vaultapis.V1FieldRecords
	for i, value := range request.Values {
		field := vaultapis.V1FieldRecords{}
		field.Fields = value
		// Ensure options.Tokens are not nil and the index i exists
		if options.Tokens != nil && i < len(options.Tokens) {
			field.Tokens = options.Tokens[i]
		}
		records = append(records, &field)
	}
	tokenMode, tokenError := SetTokenMode(options.TokenMode)
	if tokenError != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_BYOT)
	}
	insertBody := vaultapis.RecordServiceInsertRecordBody{}
	insertBody.Records = records
	insertBody.Upsert = &options.Upsert
	insertBody.Tokenization = &options.ReturnTokens
	insertBody.Byot = tokenMode
	return &insertBody, nil
}

// CreateInsertBatchBodyRequest generates the request body for batch inserts.
func CreateInsertBatchBodyRequest(request *common.InsertRequest, options *common.InsertOptions) (*vaultapis.RecordServiceBatchOperationBody, error) {
	records := make([]*vaultapis.V1BatchRecord, len(request.Values))
	for index, record := range request.Values {
		batchRecord := vaultapis.V1BatchRecord{}
		batchRecord.TableName = &request.Table
		batchRecord.Upsert = &options.Upsert
		batchRecord.Tokenization = &options.ReturnTokens
		batchRecord.Fields = record
		batchRecord.Method = vaultapis.BatchRecordMethodPost.Ptr()
		if options.Tokens != nil && index < len(options.Tokens) {
			batchRecord.Tokens = options.Tokens[index]
		}
		records[index] = &batchRecord
	}

	body := vaultapis.RecordServiceBatchOperationBody{}
	body.Records = records
	body.ContinueOnError = &options.ContinueOnError

	tokenMode, tokenError := SetTokenMode(options.TokenMode)
	if tokenError != nil {
		return nil, tokenError
	}
	body.Byot = tokenMode
	return &body, nil
}

func GetTokenizePayload(request []common.TokenizeRequest) vaultapis.V1TokenizePayload {
	payload := vaultapis.V1TokenizePayload{}
	var records []*vaultapis.V1TokenizeRecordRequest
	for _, tokenizeRequest := range request {
		record := vaultapis.V1TokenizeRecordRequest{
			Value:       &tokenizeRequest.Value,
			ColumnGroup: &tokenizeRequest.ColumnGroup,
		}
		records = append(records, &record)
	}
	payload.TokenizationParameters = records
	return payload
}

// GetURLWithEnv constructs the URL for the given environment and clusterId.
func GetURLWithEnv(env common.Env, clusterId string) string {
	var url = constants.SECURE_PROTOCOL + clusterId
	switch env {
	case common.DEV:
		url = url + constants.DEV_DOMAIN
	case common.PROD:
		url = url + constants.PROD_DOMAIN
	case common.STAGE:
		url = url + constants.STAGE_DOMAIN
	case common.SANDBOX:
		url = url + constants.SANDBOX_DOMAIN
	default:
		url = url + constants.PROD_DOMAIN
	}
	return url
}

func ParseTokenizeResponse(apiResponse vaultapis.V1TokenizeResponse) *common.TokenizeResponse {
	var tokens []string
	for _, record := range apiResponse.GetRecords() {
		tokens = append(tokens, *record.GetToken())
	}
	return &common.TokenizeResponse{
		Tokens: tokens,
	}
}
func GetFileForFileUpload(request common.FileUploadRequest) (*os.File, error) {
	if request.FilePath != "" {
		file, err := os.Open(request.FilePath)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	if request.Base64 != "" {
		data, err := base64.StdEncoding.DecodeString(request.Base64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
		file, err := os.Create(request.FileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		// Write data
		_, err = file.Write(data)
		if err != nil {
			file.Close()
			return nil, err
		}
		// Reset pointer
		file.Seek(0, io.SeekStart)

		// Remove from disk but keep open
		os.Remove(request.FileName)
		return file, nil
	}

	if request.FileObject != (os.File{}) {
		// make *os.File act as ReadCloser
		return &request.FileObject, nil
	}
	return nil, nil
}

// Generate and Sign Tokens
func GetSignedDataTokens(credKeys map[string]interface{}, options common.SignedDataTokensOptions) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
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
func GenerateSignedDataTokensHelper(clientID, keyID string, pvtKey *rsa.PrivateKey, options common.SignedDataTokensOptions, tokenURI string) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	var responseArray []common.SignedDataTokensResponse
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
		responseArray = append(responseArray, common.SignedDataTokensResponse{Token: token, SignedToken: "signed_token_" + tokenString})
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
func GenerateBearerTokenHelper(credKeys map[string]interface{}, options common.BearerTokenOptions) (*internal.V1GetAuthTokenResponse, *skyflowError.SkyflowError) {
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
func GetSignedBearerUserToken(clientID, keyID, tokenURI string, pvtKey *rsa.PrivateKey, options common.BearerTokenOptions) (string, *skyflowError.SkyflowError) {

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