package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	constants "skyflow-go/internal/constants"
	vaultmethods "skyflow-go/internal/generated/vaultapi"
	"skyflow-go/serviceaccount"
	"skyflow-go/utils/common"
	skyflowError "skyflow-go/utils/error"
	"skyflow-go/utils/logger"
)

type VaultController struct {
	Config    common.VaultConfig
	Loglevel  *logger.LogLevel
	Token     string
	ApiKey    string
	ApiClient vaultmethods.APIClient
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
	switch {
	case credentials.Path != "":
		token, err := serviceaccount.GenerateBearerToken(credentials.Path, serviceaccount.BearerTokenOptions{})
		if err != nil {
			return nil, err
		}
		bearerToken = token.AccessToken

	case credentials.CredentialsString != "":
		token, err := serviceaccount.GenerateBearerTokenFromCreds(credentials.CredentialsString, serviceaccount.BearerTokenOptions{})
		if err != nil {
			return nil, err
		}
		bearerToken = token.AccessToken

	case credentials.Token != "":
		bearerToken = credentials.Token

	default:
		return nil, skyflowError.NewSkyflowError("400", "creds not valid")
	}
	return &bearerToken, nil
}

// SetBearerToken checks and updates the token if necessary.
func SetBearerToken(v *VaultController) *skyflowError.SkyflowError {
	// Validate token or generate a new one if expired or not set.
	if v.Token == "" || serviceaccount.IsExpired(v.Token) {
		token, err := GenerateToken(v.Config.Credentials)
		if err != nil {
			return err
		}
		v.Token = *token
	}
	return nil
}

// CreateRequestClient initializes the API client with the appropriate authorization header.
func CreateRequestClient(v *VaultController) *skyflowError.SkyflowError {
	configuration := vaultmethods.NewConfiguration()
	if v.Config.Credentials.ApiKey != "" {
		v.ApiKey = v.Config.Credentials.ApiKey
		configuration.AddDefaultHeader("Authorization", "Bearer "+v.ApiKey)
	} else {
		err := SetBearerToken(v)
		if err != nil {
			return skyflowError.NewSkyflowError("400", "error occurred in token generation")
		}
		configuration.AddDefaultHeader("Authorization", "Bearer "+v.Token)
	}
	configuration.Servers[0].URL = GetURLWithEnv(v.Config.Env, v.Config.ClusterId)
	apiClient := vaultmethods.NewAPIClient(configuration)
	v.ApiClient = *apiClient
	return nil
}

// CreateInsertBulkBodyRequest createInsertBodyRequest generates the request body for bulk inserts.
func CreateInsertBulkBodyRequest(request *common.InsertRequest, options *common.InsertOptions) *vaultmethods.RecordServiceInsertRecordBody {
	var records []vaultmethods.V1FieldRecords
	for index, record := range request.Values {
		bulkRecord := vaultmethods.V1FieldRecords{
			Fields: record,
		}
		if options.Tokens != nil {
			bulkRecord.SetTokens(options.Tokens[index])
		}
		records = append(records, bulkRecord)
	}
	body := vaultmethods.NewRecordServiceInsertRecordBody()
	body.SetTokenization(options.ReturnTokens)
	body.SetUpsert(options.Upsert)
	body.SetRecords(records)
	switch options.TokenMode {
	case common.ENABLE_STRICT:
		body.SetByot(vaultmethods.V1BYOT_ENABLE_STRICT)
	case common.ENABLE:
		body.SetByot(vaultmethods.V1BYOT_ENABLE)
	case common.DISABLE:
		body.SetByot(vaultmethods.V1BYOT_DISABLE)
	default:
		body.SetByot(vaultmethods.V1BYOT_DISABLE)
	}
	return body
}

// CreateInsertBatchBodyRequest generates the request body for batch inserts.
func CreateInsertBatchBodyRequest(request *common.InsertRequest, options *common.InsertOptions) *vaultmethods.RecordServiceBatchOperationBody {
	records := make([]vaultmethods.V1BatchRecord, len(request.Values))
	for index, record := range request.Values {
		batchRecord := vaultmethods.V1BatchRecord{}
		batchRecord.SetTableName(request.Table)
		batchRecord.SetUpsert(options.Upsert)
		batchRecord.SetTokenization(options.ReturnTokens)
		batchRecord.SetFields(record)
		batchRecord.SetMethod(vaultmethods.BATCHRECORDMETHOD_POST)
		if options.Tokens != nil {
			batchRecord.SetTokens(options.Tokens[index])
		}
		records[index] = batchRecord
	}

	body := vaultmethods.NewRecordServiceBatchOperationBody()
	body.Records = records
	body.ContinueOnError = &options.ContinueOnError

	SetTokenMode(options.TokenMode, body)
	return body
}

// SetTokenMode sets the tokenization mode in the request body.
func SetTokenMode(tokenMode common.BYOT, body *vaultmethods.RecordServiceBatchOperationBody) {
	switch tokenMode {
	case common.ENABLE_STRICT:
		body.SetByot(vaultmethods.V1BYOT_ENABLE_STRICT)
	case common.ENABLE:
		body.SetByot(vaultmethods.V1BYOT_ENABLE)
	case common.DISABLE:
		body.SetByot(vaultmethods.V1BYOT_DISABLE)
	default:
		body.SetByot(vaultmethods.V1BYOT_DISABLE)
	}
}

// GetFormattedBatchInsertRecord formats the response from batch insert into a map.
func GetFormattedBatchInsertRecord(record interface{}, requestIndex int) (map[string]interface{}, error) {
	insertRecord := make(map[string]interface{})
	// Convert the record to JSON and unmarshal it
	jsonData, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %v", err)
	}

	var bodyObject map[string]interface{}
	if err := json.Unmarshal(jsonData, &bodyObject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Extract relevant data from "Body"
	body, bodyExists := bodyObject["Body"].(map[string]interface{})
	if !bodyExists {
		return nil, fmt.Errorf("Body field not found in JSON")
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
	fmt.Println("--------->", insertRecord)
	return insertRecord, nil
}
func GetFormattedBulkInsertRecord(record vaultmethods.V1RecordMetaProperties) map[string]interface{} {
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

func (v *VaultController) callBulkInsertAPI(ctx context.Context, body vaultmethods.RecordServiceInsertRecordBody, table string) (*vaultmethods.V1InsertRecordResponse, *http.Response, error) {
	bulkResp, httpsRes, err := v.ApiClient.RecordsAPI.RecordServiceInsertRecord(ctx, v.Config.VaultId, table).Body(body).Execute()
	if err != nil {
		return nil, httpsRes, err
	}
	return bulkResp, httpsRes, nil
}

// CallBatchInsertAPI handles the batch insert API call.
func (v *VaultController) CallBatchInsertAPI(ctx context.Context, body vaultmethods.RecordServiceBatchOperationBody) (*vaultmethods.V1BatchOperationResponse, *http.Response, error) {
	batchResp, httpRes, err := v.ApiClient.RecordsAPI.RecordServiceBatchOperation(ctx, v.Config.VaultId).Body(body).Execute()
	if err != nil {
		fmt.Println("error making batch insert API call", err.Error(), "=>", httpRes.Body, "=>", batchResp.HasResponses())
		return nil, httpRes, fmt.Errorf("CallBatchInsertAPI error making batch insert API call: %v", err)
	}
	fmt.Println("called the original", batchResp, httpRes)
	return batchResp, httpRes, nil
}

// Insert performs the insert operation based on provided options.
func (v *VaultController) Insert(ctx *context.Context, request *common.InsertRequest, options *common.InsertOptions) (*common.InsertResponse, *skyflowError.SkyflowError) {
	// Initialize the response structure
	var resp common.InsertResponse
	var insertedFields, errors []map[string]interface{}

	// Create the API client
	if err := CreateRequestClientFunc(v); err != nil {
		fmt.Println(err)
		return nil, skyflowError.NewSkyflowError("400", "some issue with client")
	}

	if options.ContinueOnError {
		// Batch insert handling
		body := CreateInsertBatchBodyRequest(request, options)
		batchResp, httpsRes, err1 := v.CallBatchInsertAPI(*ctx, *body)
		//v.ApiClient.RecordsAPI.RecordServiceBatchOperation(ctx, v.Config.VaultId).Body(*body).Execute()
		fmt.Println("here===>1", err1, batchResp.GetResponses(), httpsRes.Body)

		if err1 != nil {
			fmt.Println("here===>2", err1, batchResp, httpsRes)
			return nil, skyflowError.NewSkyflowError("400", "insert call failed with client")
		}

		for index, record := range batchResp.GetResponses() {
			formattedRecord, _ := GetFormattedBatchInsertRecord(record, index)
			if formattedRecord["skyflow_id"] != nil {
				insertedFields = append(insertedFields, formattedRecord)
			} else {
				errors = append(errors, formattedRecord)
			}
		}
		fmt.Println("here is", insertedFields)
		resp = common.InsertResponse{
			InsertedFields: insertedFields,
			ErrorFields:    errors,
		}
	} else {
		// Bulk insert handling
		body := CreateInsertBulkBodyRequest(request, options)
		bulkResp, httpRes, bulkErr := v.callBulkInsertAPI(*ctx, *body, request.Table)
		fmt.Println("here is1", body, httpRes)

		if bulkErr != nil {
			fmt.Println(bulkErr)
			return nil, skyflowError.NewSkyflowError("400", "insert failed")
		}

		for _, record := range bulkResp.GetRecords() {
			formattedRes := GetFormattedBulkInsertRecord(record)
			insertedFields = append(insertedFields, formattedRes)
		}

		resp = common.InsertResponse{InsertedFields: insertedFields}
	}

	return &resp, nil
}

func GetDetokenizePayload(request common.DetokenizeRequest, options common.DetokenizeOptions) vaultmethods.V1DetokenizePayload {
	payload := vaultmethods.V1DetokenizePayload{}
	payload.SetContinueOnError(options.ContinueOnError)
	var reqArray []vaultmethods.V1DetokenizeRecordRequest

	for index := range request.Tokens {
		req := vaultmethods.V1DetokenizeRecordRequest{}
		req.SetToken(request.Tokens[index])
		req.SetRedaction(vaultmethods.RedactionEnumREDACTION(request.RedactionType))
		reqArray = append(reqArray, req)
	}
	if len(reqArray) > 0 {
		payload.SetDetokenizationParameters(reqArray)
	}
	return payload
}
func (v *VaultController) Detokenize(ctx *context.Context, request *common.DetokenizeRequest, options *common.DetokenizeOptions) (*common.DetokenizeResponse, *skyflowError.SkyflowError) {
	//validate detokenize request body & options
	var detokenizedFields []map[string]interface{}
	var errorFields []map[string]interface{}
	err := SetBearerToken(v)
	if err != nil {
		return nil, err
	}
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, skyflowError.NewSkyflowError("400", "some issue with client")
	}

	payload := GetDetokenizePayload(*request, *options)
	result, httpsRes, detokenizeErr := v.ApiClient.TokensAPI.RecordServiceDetokenize(*ctx, v.Config.VaultId).DetokenizePayload(payload).Execute()
	if detokenizeErr != nil {
		fmt.Println("===>", httpsRes, detokenizeErr)
		return nil, skyflowError.NewSkyflowError("400", "some issue with client")
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
	return &common.DetokenizeResponse{
		DetokenizedFields: detokenizedFields,
		ErrorRecords:      errorFields,
	}, nil
}

func GetFormattedGetRecord(record vaultmethods.V1FieldRecords) map[string]interface{} {
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

func (v *VaultController) Get(ctx *context.Context, request *common.GetRequest, options *common.GetOptions) (*common.GetResponse, *skyflowError.SkyflowError) {
	// Get validate logic here
	var data []map[string]interface{}
	err := SetBearerToken(v)
	if err != nil {
		return nil, err
	}
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, skyflowError.NewSkyflowError("400", "some issue with client")
	}
	query := v.ApiClient.RecordsAPI.RecordServiceBulkGetRecord(*ctx, v.Config.VaultId, request.Table).SkyflowIds(request.Ids)

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
	result, httpResp, err1 := query.Execute()
	if err1 != nil {
		fmt.Println("===>", httpResp, err1)
		return nil, skyflowError.NewSkyflowError("400", "some issue with request")
	}

	records := result.GetRecords()
	if records != nil {
		for _, record := range records {
			data = append(data, GetFormattedGetRecord(record))
		}
	}
	return &common.GetResponse{Data: data}, nil
}

//func (v *VaultController) Delete(ctx context.Context, request common.DeleteRequest) (*common.DeleteResponse, *skyflowError.SkyflowError) {
//	// Delete validate logic here
//	err := SetBearerToken(v)
//	if err != nil {
//		return nil, err
//	}
//	reqBody := vaultmethods.RecordServiceBulkDeleteRecordBody{}
//	reqBody.SetSkyflowIds(request.Ids)
//	res, httpRes, err1 := v.ApiClient.RecordsAPI.RecordServiceBulkDeleteRecord(ctx, v.Config.VaultId, request.Table).Body(reqBody).Execute()
//
//	if err1 != nil {
//		fmt.Println("===>", httpRes, err1)
//		return nil, skyflowError.NewSkyflowError("400", "some issue with request")
//	}
//	return &common.DeleteResponse{
//		DeletedIds: res.GetRecordIDResponse(),
//	}, nil
//}

//func GetFormattedUpdateRecord(record vaultmethods.V1UpdateRecordResponse) map[string]interface{} {
//	updateTokens := make(map[string]interface{})
//
//	// Check if tokens are not nil
//	if record.Tokens != nil {
//		// Iterate through the map and populate updateTokens
//		for key, value := range record.Tokens {
//			updateTokens[key] = value
//		}
//	}
//
//	return updateTokens
//}

//func GetFormattedQueryRecord(record vaultmethods.V1FieldRecords) map[string]interface{} {
//	queryRecord := make(map[string]interface{})
//	if record.Fields != nil {
//		for key, value := range record.Fields {
//			queryRecord[key] = value
//		}
//	}
//	return queryRecord
//}

//func (v *VaultController) Update(ctx context.Context, request common.UpdateRequest, options common.UpdateOptions) (*common.UpdateResponse, *skyflowError.SkyflowError) {
//	// Update logic here
//	err := SetBearerToken(v)
//	if err != nil {
//		return nil, err
//	}
//	payload := vaultmethods.RecordServiceUpdateRecordBody{}
//	switch options.TokenMode {
//	case common.ENABLE_STRICT:
//		payload.SetByot(vaultmethods.V1BYOT_ENABLE_STRICT)
//	case common.ENABLE:
//		payload.SetByot(vaultmethods.V1BYOT_ENABLE)
//	default:
//		payload.SetByot(vaultmethods.V1BYOT_DISABLE)
//	}
//	payload.SetTokenization(options.ReturnTokens)
//	record := vaultmethods.V1FieldRecords{}
//	record.SetFields(request.Values)
//	if request.Tokens != nil {
//		record.SetTokens(request.Tokens)
//	}
//	result, httpRes, err1 := v.ApiClient.RecordsAPI.RecordServiceUpdateRecord(ctx, v.Config.VaultId, request.Table, request.Id).Body(payload).Execute()
//
//	if err1 != nil {
//		fmt.Println("===>", httpRes, err)
//		return nil, skyflowError.NewSkyflowError("400", "some issue with request")
//	}
//	id := result.GetSkyflowId()
//	res := GetFormattedUpdateRecord(*result)
//
//	return &common.UpdateResponse{
//		Tokens:    res,
//		SkyflowId: id,
//	}, nil
//}

//func (v *VaultController) UploadFile(ctx context.Context, request common.UploadFileRequest) (*common.UploadFileResponse, *skyflowError.SkyflowError) {
//	// UploadFile logic here
//
//	return &common.UploadFileResponse{}, nil
//}

//func (v *VaultController) Query(ctx context.Context, queryRequest common.QueryRequest) (*common.QueryResponse, *skyflowError.SkyflowError) {
//	// validate the query request
//	var fields []map[string]interface{}
//	var tokenizedData []map[string]interface{}
//
//	err := SetBearerToken(v)
//	if err != nil {
//		return nil, err
//	}
//	body := vaultmethods.QueryServiceExecuteQueryBody{}
//	body.SetQuery(queryRequest.Query)
//	result, httpRes, errr := v.ApiClient.QueryAPI.QueryServiceExecuteQuery(ctx, v.Config.VaultId).Body(body).Execute()
//	if errr != nil {
//		fmt.Println("===>", httpRes, errr)
//		return nil, skyflowError.NewSkyflowError("400", "some issue with request")
//	}
//	if result.GetRecords() != nil {
//		for _, record := range result.GetRecords() {
//			fields = append(fields, GetFormattedQueryRecord(record))
//			tokenizedData = append(tokenizedData, record.Tokens)
//		}
//	}
//	return &common.QueryResponse{
//		Fields:        fields,
//		TokenizedData: tokenizedData,
//	}, nil
//}

//func (v *VaultController) Tokenize(tokenizeRequest) {

//}
