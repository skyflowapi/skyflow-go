package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"net/http"

	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/vaultapi"
	"github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

type VaultController struct {
	Config    common.VaultConfig
	Loglevel  *logger.LogLevel
	Token     string
	ApiKey    string
	ApiClient vaultapi.APIClient
}

var CreateRequestClientFunc = CreateRequestClient

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

// GenerateToken generates a bearer token using the provided credentials.
func GenerateToken(credentials common.Credentials) (*string, *skyflowError.SkyflowError) {
	var bearerToken string
	var options = common.BearerTokenOptions{}
	if credentials.Roles != nil {
		options.RoleIDs = credentials.Roles
	}
	if credentials.Context != "" {
		options.Ctx = credentials.Context
	}
	switch {
	case credentials.Path != "":
		token, err := serviceaccount.GenerateBearerToken(credentials.Path, options)
		if err != nil {
			return nil, err
		}
		bearerToken = token.AccessToken

	case credentials.CredentialsString != "":
		token, err := serviceaccount.GenerateBearerTokenFromCreds(credentials.CredentialsString, options)
		if err != nil {
			return nil, err
		}
		bearerToken = token.AccessToken

	default:
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CREDENTIALS)
	}
	return &bearerToken, nil
}

// SetBearerTokenForVaultController checks and updates the token if necessary.
func SetBearerTokenForVaultController(v *VaultController) *skyflowError.SkyflowError {
	// Validate token or generate a new one if expired or not set.
	if v.Token == "" || serviceaccount.IsExpired(v.Token) {
		logger.Info(logs.GENERATE_BEARER_TOKEN_TRIGGERED)
		token, err := GenerateToken(v.Config.Credentials)
		if err != nil {
			return err
		}
		v.Token = *token
	} else {
		logger.Info(logs.REUSE_BEARER_TOKEN)
	}
	return nil
}

// CreateRequestClient initializes the API client with the appropriate authorization header.
func CreateRequestClient(v *VaultController) *skyflowError.SkyflowError {
	token := ""
	if v.Config.Credentials.ApiKey != "" {
		v.ApiKey = v.Config.Credentials.ApiKey
		logger.Info(logs.USING_API_KEY)
	} else if v.Config.Credentials.Token != "" {
		if serviceaccount.IsExpired(v.Config.Credentials.Token) {
			logger.Error(logs.BEARER_TOKEN_EXPIRED)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKEN_EXPIRED)
		} else {
			logger.Info(logs.USING_BEARER_TOKEN)
			v.Token = v.Config.Credentials.Token
		}
	} else {
		err := SetBearerTokenForVaultController(v)
		if err != nil {
			return err
		}
	}
	if v.ApiKey != "" {
		token = v.ApiKey
	} else if v.Token != "" {
		token = v.Token
	}
	configuration := &vaultapi.Configuration{
		DefaultHeader: map[string]string{
			"Authorization":                  "Bearer " + token,
			constants.SDK_METRICS_HEADER_KEY: helpers.CreateJsonMetadata(),
		},
		Servers: []vaultapi.ServerConfiguration{{
			URL: GetURLWithEnv(v.Config.Env, v.Config.ClusterId),
		}},
	}
	apiClient := vaultapi.NewAPIClient(configuration)
	v.ApiClient = *apiClient
	return nil
}

// CreateInsertBulkBodyRequest createInsertBodyRequest generates the request body for bulk inserts.
func CreateInsertBulkBodyRequest(request *common.InsertRequest, options *common.InsertOptions) *vaultapi.RecordServiceInsertRecordBody {
	var records []vaultapi.V1FieldRecords
	for index, record := range request.Values {
		bulkRecord := vaultapi.V1FieldRecords{
			Fields: record,
		}
		if options.Tokens != nil {
			bulkRecord.SetTokens(options.Tokens[index])
		}
		records = append(records, bulkRecord)
	}
	body := vaultapi.NewRecordServiceInsertRecordBody()
	body.SetTokenization(options.ReturnTokens)
	body.SetUpsert(options.Upsert)
	body.SetRecords(records)
	switch options.TokenMode {
	case common.ENABLE_STRICT:
		body.SetByot(vaultapi.V1BYOT_ENABLE_STRICT)
	case common.ENABLE:
		body.SetByot(vaultapi.V1BYOT_ENABLE)
	case common.DISABLE:
		body.SetByot(vaultapi.V1BYOT_DISABLE)
	default:
		body.SetByot(vaultapi.V1BYOT_DISABLE)
	}
	return body
}

// CreateInsertBatchBodyRequest generates the request body for batch inserts.
func CreateInsertBatchBodyRequest(request *common.InsertRequest, options *common.InsertOptions) *vaultapi.RecordServiceBatchOperationBody {
	records := make([]vaultapi.V1BatchRecord, len(request.Values))
	for index, record := range request.Values {
		batchRecord := vaultapi.V1BatchRecord{}
		batchRecord.SetTableName(request.Table)
		batchRecord.SetUpsert(options.Upsert)
		batchRecord.SetTokenization(options.ReturnTokens)
		batchRecord.SetFields(record)
		batchRecord.SetMethod(vaultapi.BATCHRECORDMETHOD_POST)
		if options.Tokens != nil {
			batchRecord.SetTokens(options.Tokens[index])
		}
		records[index] = batchRecord
	}

	body := vaultapi.NewRecordServiceBatchOperationBody()
	body.Records = records
	body.ContinueOnError = &options.ContinueOnError

	SetTokenMode(options.TokenMode, body)
	return body
}

// SetTokenMode sets the tokenization mode in the request body.
func SetTokenMode(tokenMode common.BYOT, body *vaultapi.RecordServiceBatchOperationBody) {
	switch tokenMode {
	case common.ENABLE_STRICT:
		body.SetByot(vaultapi.V1BYOT_ENABLE_STRICT)
	case common.ENABLE:
		body.SetByot(vaultapi.V1BYOT_ENABLE)
	case common.DISABLE:
		body.SetByot(vaultapi.V1BYOT_DISABLE)
	default:
		body.SetByot(vaultapi.V1BYOT_DISABLE)
	}
}
func GetFormattedGetRecord(record vaultapi.V1FieldRecords) map[string]interface{} {
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
func GetDetokenizePayload(request common.DetokenizeRequest, options common.DetokenizeOptions) vaultapi.V1DetokenizePayload {
	payload := vaultapi.V1DetokenizePayload{}
	payload.SetContinueOnError(options.ContinueOnError)
	var reqArray []vaultapi.V1DetokenizeRecordRequest

	for index := range request.Tokens {
		req := vaultapi.V1DetokenizeRecordRequest{}
		req.SetToken(request.Tokens[index])
		req.SetRedaction(vaultapi.RedactionEnumREDACTION(request.RedactionType))
		reqArray = append(reqArray, req)
	}
	if len(reqArray) > 0 {
		payload.SetDetokenizationParameters(reqArray)
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
func GetFormattedBulkInsertRecord(record vaultapi.V1RecordMetaProperties) map[string]interface{} {
	insertRecord := make(map[string]interface{})
	insertRecord["skyflow_id"] = record.GetSkyflowId()

	tokensMap := record.GetTokens()
	if len(tokensMap) > 0 {
		for key, value := range tokensMap {
			insertRecord[key] = value
		}
	}
	return insertRecord
}
func GetFormattedQueryRecord(record vaultapi.V1FieldRecords) map[string]interface{} {
	queryRecord := make(map[string]interface{})
	if record.Fields != nil {
		for key, value := range record.Fields {
			queryRecord[key] = value
		}
	}
	return queryRecord
}
func GetFormattedUpdateRecord(record vaultapi.V1UpdateRecordResponse) map[string]interface{} {
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
func getTokenizePayload(request []common.TokenizeRequest) vaultapi.V1TokenizePayload {
	payload := vaultapi.V1TokenizePayload{}
	var records []vaultapi.V1TokenizeRecordRequest
	for _, tokenizeRequest := range request {
		record := vaultapi.V1TokenizeRecordRequest{
			Value:       &tokenizeRequest.Value,
			ColumnGroup: &tokenizeRequest.ColumnGroup,
		}
		records = append(records, record)
	}
	payload.SetTokenizationParameters(records)
	return payload
}
func ParseTokenizeResponse(apiResponse vaultapi.V1TokenizeResponse) *common.TokenizeResponse {
	var tokens []string
	for _, record := range apiResponse.GetRecords() {
		tokens = append(tokens, record.GetToken())
	}
	return &common.TokenizeResponse{
		Tokens: tokens,
	}
}

func (v *VaultController) callBulkInsertAPI(ctx context.Context, body vaultapi.RecordServiceInsertRecordBody, table string) (*vaultapi.V1InsertRecordResponse, *http.Response, error) {
	bulkResp, httpsRes, err := v.ApiClient.RecordsAPI.RecordServiceInsertRecord(ctx, v.Config.VaultId, table).Body(body).Execute()
	if err != nil {
		return nil, httpsRes, err
	}
	return bulkResp, httpsRes, nil
}

func (v *VaultController) callBatchInsertAPI(ctx context.Context, body vaultapi.RecordServiceBatchOperationBody) (*vaultapi.V1BatchOperationResponse, *http.Response, error) {
	batchResp, httpRes, err := v.ApiClient.RecordsAPI.RecordServiceBatchOperation(ctx, v.Config.VaultId).Body(body).Execute()
	if err != nil {
		return nil, httpRes, err
	}
	return batchResp, httpRes, nil
}

func (v *VaultController) Insert(ctx context.Context, request common.InsertRequest, options common.InsertOptions) (*common.InsertResponse, *skyflowError.SkyflowError) {
	// validate insert
	logger.Info(logs.INSERT_TRIGGERED)
	logger.Info(logs.VALIDATE_INSERT_INPUT)
	errs := validation.ValidateInsertRequest(request, options)
	if errs != nil {
		return nil, errs
	}
	// Initialize the response structure
	var resp common.InsertResponse
	var insertedFields, errors []map[string]interface{}

	// Create the API client
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	if options.ContinueOnError {
		// Batch insert handling
		body := CreateInsertBatchBodyRequest(&request, &options)
		batchResp, httpsRes, err1 := v.callBatchInsertAPI(ctx, *body)
		logger.Info(logs.INSERT_BATCH_REQUEST_RESOLVED)
		if err1 != nil && httpsRes != nil {
			logger.Error(logs.INSERT_REQUEST_REJECTED)
			return nil, skyflowError.SkyflowApiError(*httpsRes)
		} else if err1 != nil {
			logger.Error(logs.INSERT_REQUEST_REJECTED)
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", err1.Error()))
		}

		for index, record := range batchResp.GetResponses() {
			formattedRecord, parseErr := GetFormattedBatchInsertRecord(record, index)
			if parseErr != nil {
				return nil, parseErr
			}
			if formattedRecord["skyflow_id"] != nil {
				insertedFields = append(insertedFields, formattedRecord)
			} else {
				errors = append(errors, formattedRecord)
			}
		}
		resp = common.InsertResponse{
			InsertedFields: insertedFields,
			ErrorFields:    errors,
		}
	} else {
		// Bulk insert handling
		body := CreateInsertBulkBodyRequest(&request, &options)
		bulkResp, httpRes, bulkErr := v.callBulkInsertAPI(ctx, *body, request.Table)
		logger.Info(logs.INSERT_BATCH_REQUEST_RESOLVED)
		if bulkErr != nil && httpRes != nil {
			logger.Error(logs.INSERT_REQUEST_REJECTED)
			return nil, skyflowError.SkyflowApiError(*httpRes)
		} else if bulkErr != nil {
			logger.Error(logs.INSERT_REQUEST_REJECTED)
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+" %v ", bulkErr.Error()))
		}

		for _, record := range bulkResp.GetRecords() {
			formattedRes := GetFormattedBulkInsertRecord(record)
			insertedFields = append(insertedFields, formattedRes)
		}
		resp = common.InsertResponse{InsertedFields: insertedFields}
	}
	logger.Info(logs.INSERT_DATA_SUCCESS)
	return &resp, nil
}

func (v *VaultController) Detokenize(ctx context.Context, request common.DetokenizeRequest, options common.DetokenizeOptions) (*common.DetokenizeResponse, *skyflowError.SkyflowError) {
	//validate detokenize request body & options
	logger.Info(logs.DETOKENIZE_TRIGGERED)
	var detokenizedFields []map[string]interface{}
	var errorFields []map[string]interface{}
	logger.Info(logs.VALIDATE_DETOKENIZE_INPUT)
	er := validation.ValidateDetokenizeRequest(request)
	if er != nil {
		return nil, er
	}
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}

	payload := GetDetokenizePayload(request, options)
	result, httpsRes, detokenizeErr := v.ApiClient.TokensAPI.RecordServiceDetokenize(ctx, v.Config.VaultId).DetokenizePayload(payload).Execute()
	if detokenizeErr != nil && httpsRes != nil {
		logger.Error(logs.DETOKENIZE_REQUEST_REJECTED)
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if detokenizeErr != nil {
		logger.Error(logs.DETOKENIZE_REQUEST_REJECTED)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", detokenizeErr))
	}
	logger.Info(logs.DETOKENIZE_REQUEST_RESOLVED)
	if result != nil {
		records := result.GetRecords()
		for _, record := range records {
			if record.HasError() {
				er1 := map[string]interface{}{
					"ValueType": string(record.GetValueType()),
					"Token":     record.GetToken(),
					"Value":     record.GetValue(),
					"Error":     record.GetError(),
				}
				errorFields = append(errorFields, er1)
			} else {
				var rec map[string]interface{}
				rec = map[string]interface{}{
					"ValueType": string(record.GetValueType()),
					"Token":     record.GetToken(),
					"Value":     record.GetValue(),
					"Error":     record.Error,
				}
				detokenizedFields = append(detokenizedFields, rec)
			}
		}
	}
	return &common.DetokenizeResponse{
		DetokenizedFields: detokenizedFields,
		ErrorRecords:      errorFields,
	}, nil
}

func (v *VaultController) Get(ctx context.Context, request common.GetRequest, options common.GetOptions) (*common.GetResponse, *skyflowError.SkyflowError) {
	// Get validate logic here
	logger.Info(logs.GET_TRIGGERED)
	logger.Info(logs.VALIDATE_GET_INPUT)
	errs := validation.ValidateGetRequest(request, options)
	if errs != nil {
		return nil, errs
	}
	var data []map[string]interface{}
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	query := v.ApiClient.RecordsAPI.RecordServiceBulkGetRecord(ctx, v.Config.VaultId, request.Table).SkyflowIds(request.Ids)

	// Add conditional chaining for optional parameters
	if options.RedactionType != "" {
		query = query.Redaction(string(options.RedactionType))
	}
	if options.Offset != "" {
		query = query.Offset(options.Offset)
	}
	if options.Limit != "" {
		query = query.Limit(options.Limit)
	}
	if options.ColumnName != "" {
		query = query.ColumnName(options.ColumnName)
	}
	if options.ColumnValues != nil {
		query = query.ColumnValues(options.ColumnValues)
	}
	if options.OrderBy != "" {
		query = query.OrderBy(string(options.OrderBy))
	}
	query = query.Tokenization(options.ReturnTokens)
	query.DownloadURL(options.DownloadURL)
	// Execute the query
	result, httpsRes, err1 := query.Execute()
	if err1 != nil && httpsRes != nil {
		logger.Error(logs.GET_REQUEST_REJECTED)
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if err1 != nil {
		logger.Error(logs.GET_REQUEST_REJECTED)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", err1.Error()))
	}
	logger.Info(logs.GET_REQUEST_RESOLVED)
	records := result.GetRecords()
	if records != nil {
		for _, record := range records {
			data = append(data, GetFormattedGetRecord(record))
		}
	}
	logger.Info(logs.GET_SUCCESS)
	return &common.GetResponse{Data: data}, nil
}

func (v *VaultController) Delete(ctx context.Context, request common.DeleteRequest) (*common.DeleteResponse, *skyflowError.SkyflowError) {
	// Delete validate logic here
	logger.Info(logs.DELETE_TRIGGERED)
	logger.Info(logs.VALIDATE_DELETE_INPUT)
	errs := validation.ValidateDeleteRequest(request)
	if errs != nil {
		return nil, errs
	}

	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	reqBody := vaultapi.RecordServiceBulkDeleteRecordBody{}
	reqBody.SetSkyflowIds(request.Ids)
	res, httpsRes, err1 := v.ApiClient.RecordsAPI.RecordServiceBulkDeleteRecord(ctx, v.Config.VaultId, request.Table).Body(reqBody).Execute()

	if err1 != nil && httpsRes != nil {
		logger.Error(logs.DELETE_REQUEST_REJECTED)
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if err1 != nil {
		logger.Error(logs.DELETE_REQUEST_REJECTED)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", err1.Error()))
	}
	logger.Info(logs.DELETE_REQUEST_RESOLVED)
	logger.Info(logs.DELETE_SUCCESS)
	return &common.DeleteResponse{
		DeletedIds: res.GetRecordIDResponse(),
	}, nil
}

func (v *VaultController) Query(ctx context.Context, queryRequest common.QueryRequest) (*common.QueryResponse, *skyflowError.SkyflowError) {
	// validate the query request
	logger.Info(logs.QUERY_TRIGGERED)
	logger.Info(logs.VALIDATE_QUERY_INPUT)
	errs := validation.ValidateQueryRequest(queryRequest)
	if errs != nil {
		return nil, errs
	}
	var fields []map[string]interface{}
	var tokenizedData []map[string]interface{}

	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	body := vaultapi.QueryServiceExecuteQueryBody{}
	body.SetQuery(queryRequest.Query)
	result, httpsRes, errr := v.ApiClient.QueryAPI.QueryServiceExecuteQuery(ctx, v.Config.VaultId).Body(body).Execute()
	if errr != nil && httpsRes != nil {
		logger.Error(logs.QUERY_REQUEST_REJECTED)
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if errr != nil {
		logger.Error(logs.QUERY_REQUEST_REJECTED)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", errr.Error()))
	}
	logger.Info(logs.QUERY_REQUEST_RESOLVED)
	if result.GetRecords() != nil {
		for _, record := range result.GetRecords() {
			fields = append(fields, GetFormattedQueryRecord(record))
			tokenizedData = append(tokenizedData, record.Tokens)
		}
	}
	logger.Info(logs.QUERY_SUCCESS)
	return &common.QueryResponse{
		Fields:        fields,
		TokenizedData: tokenizedData,
	}, nil
}

func (v *VaultController) Update(ctx context.Context, request common.UpdateRequest, options common.UpdateOptions) (*common.UpdateResponse, *skyflowError.SkyflowError) {
	// Update validate logic here
	logger.Info(logs.UPDATE_TRIGGERED)
	logger.Info(logs.VALIDATE_UPDATE_INPUT)
	errs := validation.ValidateUpdateRequest(request, options)
	if errs != nil {
		return nil, errs
	}

	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	payload := vaultapi.RecordServiceUpdateRecordBody{}
	switch options.TokenMode {
	case common.ENABLE_STRICT:
		payload.SetByot(vaultapi.V1BYOT_ENABLE_STRICT)
	case common.ENABLE:
		payload.SetByot(vaultapi.V1BYOT_ENABLE)
	default:
		payload.SetByot(vaultapi.V1BYOT_DISABLE)
	}
	payload.SetTokenization(options.ReturnTokens)
	record := vaultapi.V1FieldRecords{}
	record.SetFields(request.Values)
	if request.Tokens != nil {
		record.SetTokens(request.Tokens)
	}
	payload.SetRecord(record)
	result, httpsRes, errr := v.ApiClient.RecordsAPI.RecordServiceUpdateRecord(ctx, v.Config.VaultId, request.Table, request.Id).Body(payload).Execute()

	if errr != nil && httpsRes != nil {
		logger.Error(logs.UPDATE_REQUEST_REJECTED)
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if errr != nil {
		logger.Error(logs.UPDATE_REQUEST_REJECTED)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", errr.Error()))
	}
	logger.Info(logs.UPDATE_REQUEST_RESOLVED)
	id := result.GetSkyflowId()
	res := GetFormattedUpdateRecord(*result)
	logger.Info(logs.UPDATE_SUCCESS)
	return &common.UpdateResponse{
		Tokens:    res,
		SkyflowId: id,
	}, nil
}

func (v *VaultController) Tokenize(ctx context.Context, request []common.TokenizeRequest) (*common.TokenizeResponse, *skyflowError.SkyflowError) {
	// Update validate logic here
	logger.Info(logs.TOKENIZE_TRIGGERED)
	logger.Info(logs.VALIDATE_TOKENIZE_INPUT)
	err := validation.ValidateTokenizeRequest(request)
	if err != nil {
		return nil, err
	}
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	payload := getTokenizePayload(request)
	result, httpsRes, tokenizeErr := v.ApiClient.TokensAPI.RecordServiceTokenize(ctx, v.Config.VaultId).TokenizePayload(payload).Execute()
	if tokenizeErr != nil && httpsRes != nil {
		logger.Error(logs.TOKENIZE_REQUEST_REJECTED)
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if tokenizeErr != nil {
		logger.Error(logs.TOKENIZE_REQUEST_REJECTED)
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", tokenizeErr.Error()))
	}
	logger.Info(logs.TOKENIZE_SUCCESS)
	return ParseTokenizeResponse(*result), nil
}
