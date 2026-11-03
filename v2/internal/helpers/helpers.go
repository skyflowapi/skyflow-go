package helpers

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/skyflowapi/skyflow-go/v2/internal/generated/core"

	"github.com/skyflowapi/skyflow-go/common/auth"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	internal "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	common "github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

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
	if tokenMode != "" && tokenMode != common.DISABLE && tokenMode != common.ENABLE_STRICT && tokenMode != common.ENABLE {
		logger.Error(fmt.Sprintf(logs.INVALID_BYOT_MODE, tokenMode))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_BYOT_MODE, tokenMode))
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
			if key == constants.API_SKYFLOW_ID {
				getRecord[constants.SKYFLOW_ID] = value
				getRecord[constants.API_SKYFLOW_ID] = value // backward compat
			} else {
				getRecord[key] = value
			}
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
	if options.DownloadUrl {
		t := true
		payload.DownloadUrl = &t
	} else if options.DownloadURL {
		logger.Warn(logs.DEPRECATED_FIELD_DOWNLOAD_URL)
		t := true
		payload.DownloadUrl = &t
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

	// Extract relevant data from Body
	body, bodyExists := bodyObject[constants.JSON_KEY_BODY].(map[string]interface{})
	if !bodyExists {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_RESPONSE)
	}

	// Handle extracted data
	if records, ok := body[constants.JSON_KEY_RECORDS].([]interface{}); ok {
		for _, rec := range records {
			recordObject, isMap := rec.(map[string]interface{})
			if !isMap {
				continue
			}
			if skyflowID, exists := recordObject[constants.API_SKYFLOW_ID].(string); exists {
				insertRecord[constants.SKYFLOW_ID] = skyflowID
				insertRecord[constants.API_SKYFLOW_ID] = skyflowID // backward compat
			}
			if tokens, exists := recordObject[constants.JSON_KEY_TOKENS].(map[string]interface{}); exists {
				for key, value := range tokens {
					insertRecord[key] = value
				}
			}
		}
	}

	if errorField, exists := body[constants.ERROR_KEY_ERROR].(string); exists {
		insertRecord[constants.ERROR_KEY_ERROR] = errorField
	}

	insertRecord["RequestIndex"] = requestIndex
	insertRecord["request_index"] = requestIndex
	return insertRecord, nil
}
func GetFormattedBulkInsertRecord(record vaultapis.V1RecordMetaProperties) map[string]interface{} {
	insertRecord := make(map[string]interface{})
	if id := record.GetSkyflowId(); id != nil {
		insertRecord[constants.SKYFLOW_ID] = *id
		insertRecord[constants.API_SKYFLOW_ID] = *id // backward compat
	}

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
			if key == constants.API_SKYFLOW_ID {
				queryRecord[constants.SKYFLOW_ID] = value
				queryRecord[constants.API_SKYFLOW_ID] = value // backward compat
			} else {
				queryRecord[key] = value
			}
		}
		if record.Tokens != nil && len(record.Tokens) > 0 {
			tokens := make(map[string]interface{})
			for key, value := range record.Tokens {
				tokens[key] = value
			}
			queryRecord[constants.TOKENIZED_DATA] = tokens
			queryRecord[constants.API_TOKENIZED_DATA] = tokens // backward compat
		} else {
			queryRecord[constants.TOKENIZED_DATA] = nil
			queryRecord[constants.API_TOKENIZED_DATA] = nil // backward compat
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
		if t := record.GetToken(); t != nil {
			tokens = append(tokens, *t)
		}
	}
	return &common.TokenizeResponse{
		Tokens: tokens,
	}
}

type namedReader struct {
	*bytes.Reader
	name string
}

func (r *namedReader) Name() string {
	return r.name
}

func (r *namedReader) Close() error {
	return nil
}

func GetFileForFileUpload(request common.FileUploadRequest) (io.ReadCloser, error) {
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
			return nil, fmt.Errorf(logs.FAILED_TO_DECODE_BASE64, err)
		}
		return &namedReader{
			Reader: bytes.NewReader(data),
			name:   request.FileName,
		}, nil
	}
	if request.FileObject != (os.File{}) {
		return &request.FileObject, nil
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Service-account authentication is implemented once, in the shared
// skyflow-go/common/auth package, and used by every SDK in this repository.
// The functions below keep their historical names and signatures and delegate
// to that implementation configured with v2's SDK prefix.
// ---------------------------------------------------------------------------

var authSvc = auth.New(auth.Config{
	SDKPrefix: constants.SDK_PREFIX,
	LogPrefix: constants.SDK_LOG_PREFIX,
})

// GetBaseURLHelper resolves the token endpoint base URL. It is a variable so
// tests can point it at a local server.
var GetBaseURLHelper = authSvc.GetBaseURL

func init() {
	// Read GetBaseURLHelper at call time so test overrides take effect.
	authSvc.SetBaseURLResolver(func(tokenUri string) (string, *skyflowError.SkyflowError) {
		return GetBaseURLHelper(tokenUri)
	})
}

// Auth returns the shared service-account authenticator configured for v2.
func Auth() *auth.Service { return authSvc }

// Helper function to read and parse credentials from file
func ParseCredentialsFile(credentialsFilePath string) (map[string]interface{}, *skyflowError.SkyflowError) {
	return authSvc.ParseCredentialsFile(credentialsFilePath)
}

// Generate and Sign Tokens
func GetSignedDataTokens(credKeys map[string]interface{}, options common.SignedDataTokensOptions) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	return authSvc.GetSignedDataTokens(credKeys, options)
}

func GetCredentialParams(credKeys map[string]interface{}) (string, string, string, *skyflowError.SkyflowError) {
	return authSvc.GetCredentialParams(credKeys)
}

func GenerateSignedDataTokensHelper(clientId, keyId string, pvtKey *rsa.PrivateKey, options common.SignedDataTokensOptions, tokenUri string) ([]common.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	return authSvc.GenerateSignedDataTokensHelper(clientId, keyId, pvtKey, options, tokenUri)
}

func GetPrivateKey(credKeys map[string]interface{}) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	return authSvc.GetPrivateKey(credKeys)
}

func ParsePrivateKey(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	return authSvc.ParsePrivateKey(pemKey)
}

// GenerateBearerTokenHelper  helper functions
func GenerateBearerTokenHelper(credKeys map[string]interface{}, options common.BearerTokenOptions) (*internal.V1GetAuthTokenResponse, *skyflowError.SkyflowError) {
	res, err := authSvc.GenerateBearerTokenHelper(credKeys, options)
	if err != nil {
		return nil, err
	}
	return &internal.V1GetAuthTokenResponse{AccessToken: res.AccessToken, TokenType: res.TokenType}, nil
}

func GetScopeUsingRoles(roles []*string) string {
	ids := make([]string, 0, len(roles))
	for _, role := range roles {
		if role != nil {
			ids = append(ids, *role)
		}
	}
	return auth.GetScopeUsingRoles(ids)
}

func GetBaseURL(urlStr string) (string, *skyflowError.SkyflowError) {
	return authSvc.GetBaseURL(urlStr)
}

func ValidateAndResolveCtx(ctx interface{}) (interface{}, *skyflowError.SkyflowError) {
	return authSvc.ValidateAndResolveCtx(ctx)
}

func GetSignedBearerUserToken(clientId, keyId, tokenUri string, pvtKey *rsa.PrivateKey, options common.BearerTokenOptions) (string, *skyflowError.SkyflowError) {
	return authSvc.GetSignedBearerUserToken(clientId, keyId, tokenUri, pvtKey, options)
}

func GetPrivateKeyFromPem(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	return authSvc.ParsePrivateKey(pemKey)
}

func CreateJsonMetadata() string {
	// Create a map to hold the key-value pairs
	data := map[string]string{
		constants.SDK_METADATA_KEY_NAME_VERSION:    fmt.Sprintf("%s@%s", constants.METRICS_SDK_NAME, constants.SDK_VERSION),
		constants.SDK_METADATA_KEY_DEVICE_MODEL:    string(runtime.GOOS),
		constants.SDK_METADATA_KEY_OS_DETAILS:      fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		constants.SDK_METADATA_KEY_RUNTIME_DETAILS: runtime.Version(),
	}

	// Marshal the map into JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Debug(logs.FAILED_TO_MARSHALL_JSON_METADATA)
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
	if id, ok := data[constants.SKYFLOW_ID].(string); ok {
		return id, true
	}
	// backward compat: accept old key from main branch
	if id, ok := data[constants.API_SKYFLOW_ID].(string); ok {
		return id, true
	}
	return "", false
}
