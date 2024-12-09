package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	constants "skyflow-go/v2/internal/constants"
	"skyflow-go/v2/internal/generated/vaultapi"
	"skyflow-go/v2/internal/validation"
	"skyflow-go/v2/serviceaccount"
	. "skyflow-go/v2/utils/common"
	skyflowError "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
)

type VaultController struct {
	Config    VaultConfig
	Loglevel  *logger.LogLevel
	Token     string
	ApiKey    string
	ApiClient vaultapi.APIClient
}

var CreateRequestClientFunc = CreateRequestClient

// GetURLWithEnv constructs the URL for the given environment and clusterId.
func GetURLWithEnv(env Env, clusterId string) string {
	var url = constants.SECURE_PROTOCOL + clusterId
	switch env {
	case DEV:
		url = url + constants.DEV_DOMAIN
	case PROD:
		url = url + constants.PROD_DOMAIN
	case STAGE:
		url = url + constants.STAGE_DOMAIN
	case SANDBOX:
		url = url + constants.SANDBOX_DOMAIN
	default:
		url = url + constants.PROD_DOMAIN
	}
	return url
}

// GenerateToken generates a bearer token using the provided credentials.
func GenerateToken(credentials Credentials) (*string, *skyflowError.SkyflowError) {
	var bearerToken string
	switch {
	case credentials.Path != "":
		token, err := serviceaccount.GenerateBearerToken(credentials.Path, BearerTokenOptions{})
		if err != nil {
			return nil, err
		}
		bearerToken = token.AccessToken

	case credentials.CredentialsString != "":
		token, err := serviceaccount.GenerateBearerTokenFromCreds(credentials.CredentialsString, BearerTokenOptions{})
		if err != nil {
			return nil, err
		}
		bearerToken = token.AccessToken

	case credentials.Token != "":
		bearerToken = credentials.Token

	default:
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CREDENTIALS)
	}
	return &bearerToken, nil
}

// SetBearerTokenForVaultController checks and updates the token if necessary.
func SetBearerTokenForVaultController(v *VaultController) *skyflowError.SkyflowError {
	// Validate token or generate a new one if expired or not set.
	if v.Token == "" || serviceaccount.IsExpired(v.Token) {
		token, err := GenerateToken(v.Config.Credentials)
		if err != nil {
			return err
		}
		v.Token = *token
	}
	if serviceaccount.IsExpired(v.Token) {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKEN_EXPIRED)
	}
	return nil
}

// CreateRequestClient initializes the API client with the appropriate authorization header.
func CreateRequestClient(v *VaultController) *skyflowError.SkyflowError {
	configuration := vaultapi.NewConfiguration()
	if v.Config.Credentials.ApiKey != "" {
		v.ApiKey = v.Config.Credentials.ApiKey
		configuration.AddDefaultHeader("Authorization", "Bearer "+v.ApiKey)
	} else {
		err := SetBearerTokenForVaultController(v)
		if err != nil {
			return err
		}
		configuration.AddDefaultHeader("Authorization", "Bearer "+v.Token)
	}
	configuration.Servers[0].URL = GetURLWithEnv(v.Config.Env, v.Config.ClusterId)
	apiClient := vaultapi.NewAPIClient(configuration)
	v.ApiClient = *apiClient
	return nil
}

// CreateInsertBulkBodyRequest createInsertBodyRequest generates the request body for bulk inserts.
func CreateInsertBulkBodyRequest(request *InsertRequest, options *InsertOptions) *vaultapi.RecordServiceInsertRecordBody {
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
	case ENABLE_STRICT:
		body.SetByot(vaultapi.V1BYOT_ENABLE_STRICT)
	case ENABLE:
		body.SetByot(vaultapi.V1BYOT_ENABLE)
	case DISABLE:
		body.SetByot(vaultapi.V1BYOT_DISABLE)
	default:
		body.SetByot(vaultapi.V1BYOT_DISABLE)
	}
	return body
}

// CreateInsertBatchBodyRequest generates the request body for batch inserts.
func CreateInsertBatchBodyRequest(request *InsertRequest, options *InsertOptions) *vaultapi.RecordServiceBatchOperationBody {
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
func SetTokenMode(tokenMode BYOT, body *vaultapi.RecordServiceBatchOperationBody) {
	switch tokenMode {
	case ENABLE_STRICT:
		body.SetByot(vaultapi.V1BYOT_ENABLE_STRICT)
	case ENABLE:
		body.SetByot(vaultapi.V1BYOT_ENABLE)
	case DISABLE:
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
func GetDetokenizePayload(request DetokenizeRequest, options DetokenizeOptions) vaultapi.V1DetokenizePayload {
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
func getTokenizePayload(request []TokenizeRequest) vaultapi.V1TokenizePayload {
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
func ParseTokenizeResponse(apiResponse vaultapi.V1TokenizeResponse) *TokenizeResponse {
	var tokens []string
	for _, record := range apiResponse.GetRecords() {
		tokens = append(tokens, record.GetToken())
	}
	return &TokenizeResponse{
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

func (v *VaultController) Insert(ctx context.Context, request InsertRequest, options InsertOptions) (*InsertResponse, *skyflowError.SkyflowError) {
	// validate insert
	errs := validation.ValidateInsertRequest(request, options)
	if errs != nil {
		return nil, errs
	}
	// Initialize the response structure
	var resp InsertResponse
	var insertedFields, errors []map[string]interface{}

	// Create the API client
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	if options.ContinueOnError {
		// Batch insert handling
		body := CreateInsertBatchBodyRequest(&request, &options)
		batchResp, httpsRes, err1 := v.callBatchInsertAPI(ctx, *body)
		if err1 != nil && httpsRes != nil {
			return nil, skyflowError.SkyflowApiError(*httpsRes)
		} else if err1 != nil {
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
		resp = InsertResponse{
			InsertedFields: insertedFields,
			ErrorFields:    errors,
		}
	} else {
		// Bulk insert handling
		body := CreateInsertBulkBodyRequest(&request, &options)
		bulkResp, httpRes, bulkErr := v.callBulkInsertAPI(ctx, *body, request.Table)
		if bulkErr != nil && httpRes != nil {
			return nil, skyflowError.SkyflowApiError(*httpRes)
		} else if bulkErr != nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+" %v ", bulkErr.Error()))
		}

		for _, record := range bulkResp.GetRecords() {
			formattedRes := GetFormattedBulkInsertRecord(record)
			insertedFields = append(insertedFields, formattedRes)
		}

		resp = InsertResponse{InsertedFields: insertedFields}
	}

	return &resp, nil
}

func (v *VaultController) Detokenize(ctx context.Context, request DetokenizeRequest, options DetokenizeOptions) (*DetokenizeResponse, *skyflowError.SkyflowError) {
	//validate detokenize request body & options
	var detokenizedFields []map[string]interface{}
	var errorFields []map[string]interface{}
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
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if detokenizeErr != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", detokenizeErr.Error()))
	}
	if result != nil {
		records := result.GetRecords()
		for _, record := range records {
			if record.HasError() {
				er, _ := record.ToMap()
				errorFields = append(errorFields, er)
			} else {
				var rec map[string]interface{}
				rec = map[string]interface{}{
					"ValueType": string(record.GetValueType()),
					"Token":     record.GetToken(),
					"Value":     record.GetValue(),
					"Error":     record.GetError(),
				}
				detokenizedFields = append(detokenizedFields, rec)
			}
		}
	}
	return &DetokenizeResponse{
		DetokenizedFields: detokenizedFields,
		ErrorRecords:      errorFields,
	}, nil
}

func (v *VaultController) Get(ctx context.Context, request GetRequest, options GetOptions) (*GetResponse, *skyflowError.SkyflowError) {
	// Get validate logic here
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
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if err1 != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", err1.Error()))
	}

	records := result.GetRecords()
	if records != nil {
		for _, record := range records {
			data = append(data, GetFormattedGetRecord(record))
		}
	}
	return &GetResponse{Data: data}, nil
}

func (v *VaultController) Delete(ctx context.Context, request DeleteRequest) (*DeleteResponse, *skyflowError.SkyflowError) {
	// Delete validate logic here
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
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if err1 != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", err1.Error()))
	}
	return &DeleteResponse{
		DeletedIds: res.GetRecordIDResponse(),
	}, nil
}

func (v *VaultController) Query(ctx context.Context, queryRequest QueryRequest) (*QueryResponse, *skyflowError.SkyflowError) {
	// validate the query request
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
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if errr != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", errr.Error()))
	}
	if result.GetRecords() != nil {
		for _, record := range result.GetRecords() {
			fields = append(fields, GetFormattedQueryRecord(record))
			tokenizedData = append(tokenizedData, record.Tokens)
		}
	}
	return &QueryResponse{
		Fields:        fields,
		TokenizedData: tokenizedData,
	}, nil
}

func (v *VaultController) Update(ctx context.Context, request UpdateRequest, options UpdateOptions) (*UpdateResponse, *skyflowError.SkyflowError) {
	// Update validate logic here
	errs := validation.ValidateUpdateRequest(request, options)
	if errs != nil {
		return nil, errs
	}

	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	payload := vaultapi.RecordServiceUpdateRecordBody{}
	switch options.TokenMode {
	case ENABLE_STRICT:
		payload.SetByot(vaultapi.V1BYOT_ENABLE_STRICT)
	case ENABLE:
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
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if errr != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", errr.Error()))
	}

	id := result.GetSkyflowId()
	res := GetFormattedUpdateRecord(*result)

	return &UpdateResponse{
		Tokens:    res,
		SkyflowId: id,
	}, nil
}

func (v *VaultController) Tokenize(ctx context.Context, request []TokenizeRequest) (*TokenizeResponse, *skyflowError.SkyflowError) {
	// Update validate logic here
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
		return nil, skyflowError.SkyflowApiError(*httpsRes)
	} else if tokenizeErr != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.ERROR_OCCURRED+"%v", tokenizeErr.Error()))
	}

	return ParseTokenizeResponse(*result), nil
}
