package controller

import (
	"context"
	"fmt"
	"net/http"

	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
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
	ApiClient client.Client
}

var CreateRequestClientFunc = CreateRequestClient

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

	header := http.Header{}
	header.Set(constants.SDK_METRICS_HEADER_KEY, helpers.CreateJsonMetadata())

	client := client.NewClient(option.WithBaseURL(helpers.GetURLWithEnv(v.Config.Env, v.Config.ClusterId)),
		option.WithToken(token),
		option.WithMaxAttempts(1),
		option.WithHTTPHeader(header),
	)

	v.ApiClient = *client
	return nil
}

func (v *VaultController) callBulkInsertAPI(ctx context.Context, body vaultapis.RecordServiceInsertRecordBody, table string) (*vaultapis.V1InsertRecordResponse, error) {
	bulkResp, err := v.ApiClient.Records.WithRawResponse.RecordServiceInsertRecord(ctx, v.Config.VaultId, table, &body)
	if err != nil {
		return nil, err
	}
	insertRecordRes := &vaultapis.V1InsertRecordResponse{}
	if bulkResp.Body != nil {
		insertRecordRes = bulkResp.Body
	}
	return insertRecordRes, nil
}

func (v *VaultController) callBatchInsertAPI(ctx context.Context, body vaultapis.RecordServiceBatchOperationBody) (*core.Response[*vaultapis.V1BatchOperationResponse], error) {
	batchResp, err := v.ApiClient.Records.WithRawResponse.RecordServiceBatchOperation(ctx, v.Config.VaultId, &body)
	if err != nil {
		return nil, err
	}
	return batchResp, nil
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
		body, bodyErr := helpers.CreateInsertBatchBodyRequest(&request, &options)
		if bodyErr != nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf("%v", bodyErr))
		}
		batchResp, apiErr := v.callBatchInsertAPI(ctx, *body)
		logger.Info(logs.INSERT_BATCH_REQUEST_RESOLVED)

		var header http.Header
		if apiErr != nil {
			header, _ = helpers.GetHeader(apiErr)
			logger.Error(logs.INSERT_REQUEST_REJECTED)
			return nil, skyflowError.SkyflowErrorApi(apiErr, header)
		}
		if batchResp != nil {
			if batchResp.Header != nil {
				header = batchResp.Header
			}
		}
		for index, record := range batchResp.Body.GetResponses() {
			formattedRecord, parseErr := helpers.GetFormattedBatchInsertRecord(record, index)
			if parseErr != nil {
				return nil, parseErr
			}
			if formattedRecord["skyflow_id"] != nil {
				insertedFields = append(insertedFields, formattedRecord)
			} else {
				formattedRecord["RequestId"] = header.Get(constants.REQUEST_KEY)
				formattedRecord["HttpCode"] = skyflowError.INVALID_INPUT_CODE
				errors = append(errors, formattedRecord)
			}
		}
		resp = common.InsertResponse{
			InsertedFields: insertedFields,
			Errors:         errors,
		}
	} else {
		// Bulk insert handling
		body, bodyErr := helpers.CreateInsertBulkBodyRequest(&request, &options)
		if bodyErr != nil {
			return nil, bodyErr
		}
		bulkResp, bulkErr := v.callBulkInsertAPI(ctx, *body, request.Table)
		logger.Info(logs.INSERT_BATCH_REQUEST_RESOLVED)
		if bulkErr != nil {
			header, _ := helpers.GetHeader(bulkErr)
			logger.Error(logs.INSERT_REQUEST_REJECTED)
			return nil, skyflowError.SkyflowErrorApi(bulkErr, header)
		}

		for _, record := range bulkResp.GetRecords() {
			formattedRes := helpers.GetFormattedBulkInsertRecord(*record)
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
	var detokenizedFields []common.DetokenizeRecordResponse
	var errorFields []common.DetokenizeRecordResponse
	logger.Info(logs.VALIDATE_DETOKENIZE_INPUT)
	er := validation.ValidateDetokenizeRequest(request)
	if er != nil {
		return nil, er
	}
	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}

	payload := helpers.GetDetokenizePayload(request, options)
	detokenizeApiRes, apiErr := v.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(ctx, v.Config.VaultId, &payload)
	var header http.Header

	if apiErr != nil {
		logger.Error(logs.DETOKENIZE_REQUEST_REJECTED)
		header, _ = helpers.GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}
	if detokenizeApiRes != nil {
		if detokenizeApiRes.Header != nil {
			header = detokenizeApiRes.Header
		}
	}
	logger.Info(logs.DETOKENIZE_REQUEST_RESOLVED)
	if detokenizeApiRes != nil && detokenizeApiRes.Body != nil {
		records := detokenizeApiRes.Body.Records
		for _, record := range records {
			if record.Error != nil {
				fieldErr := common.DetokenizeRecordResponse{
					Token:     *record.GetToken(),
					Error:     *record.GetError(),
					RequestId: header.Get(constants.REQUEST_KEY),
				}
				errorFields = append(errorFields, fieldErr)
			} else {
				rec := common.DetokenizeRecordResponse{
					Type:  string(*record.ValueType),
					Token: *record.GetToken(),
					Value: *record.GetValue(),
				}
				detokenizedFields = append(detokenizedFields, rec)
			}
		}
	}
	return &common.DetokenizeResponse{
		DetokenizedFields: detokenizedFields,
		Errors:            errorFields,
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
	req := vaultapis.RecordServiceBulkGetRecordRequest{}
	var ids []*string
	for _, s := range request.Ids {
		str := s
		ids = append(ids, &str)
	}
	req.SkyflowIds = ids

	if options.RedactionType != "" {
		redaction, _ := vaultapis.NewRecordServiceBulkGetRecordRequestRedactionFromString(string(options.RedactionType))
		req.Redaction = &redaction
	}

	if options.Offset != "" {
		req.Offset = &options.Offset
	}
	if options.Limit != "" {
		req.Limit = &options.Limit
	}
	if options.ColumnName != "" {
		req.ColumnName = &options.ColumnName
	}
	if options.ColumnValues != nil {
		var values []*string
		for _, s := range options.ColumnValues {
			str := s
			values = append(values, &str)
		}
		req.ColumnValues = values
	}
	if options.OrderBy != "" {
		orderBy, _ := vaultapis.NewRecordServiceBulkGetRecordRequestOrderByFromString(string(options.OrderBy))
		req.OrderBy = &orderBy
	}
	if options.DownloadURL {
		req.DownloadUrl = &options.DownloadURL
	}
	if options.ReturnTokens {
		req.Tokenization = &options.ReturnTokens
	} else {
		tokens := false
		req.Tokenization = &tokens
	}
	if options.Fields != nil {
		var fields []*string
		for _, s := range options.Fields {
			str := s
			fields = append(fields, &str)
		}
		req.Fields = fields
	}

	// Execute the query
	getApiRes, apiErr := v.ApiClient.Records.WithRawResponse.RecordServiceBulkGetRecord(ctx, v.Config.VaultId, request.Table, &req)
	if apiErr != nil {
		logger.Error(logs.GET_REQUEST_REJECTED)
		header, _ := helpers.GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}
	logger.Info(logs.GET_REQUEST_RESOLVED)
	if getApiRes != nil && getApiRes.Body != nil {
		records := getApiRes.Body.GetRecords()
		if len(records) > 0 {
			for _, record := range records {
				data = append(data, helpers.GetFormattedGetRecord(*record))
			}
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
	reqBody := vaultapis.RecordServiceBulkDeleteRecordBody{}
	reqBody.SkyflowIds = request.Ids
	deleteApiRes, apiErr := v.ApiClient.Records.WithRawResponse.RecordServiceBulkDeleteRecord(ctx, v.Config.VaultId, request.Table, &reqBody)

	if apiErr != nil {
		logger.Error(logs.DELETE_REQUEST_REJECTED)
		header, _ := helpers.GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}
	logger.Info(logs.DELETE_REQUEST_RESOLVED)
	logger.Info(logs.DELETE_SUCCESS)
	deleteRes := &common.DeleteResponse{}
	if deleteApiRes != nil && deleteApiRes.Body != nil {
		deleteRes.DeletedIds = deleteApiRes.Body.GetRecordIdResponse()
	}
	return deleteRes, nil
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
	body := vaultapis.QueryServiceExecuteQueryBody{}
	body.Query = &queryRequest.Query
	queryApiRes, apiErr := v.ApiClient.Query.WithRawResponse.QueryServiceExecuteQuery(ctx, v.Config.VaultId, &body)
	if apiErr != nil {
		logger.Error(logs.QUERY_REQUEST_REJECTED)
		header, _ := helpers.GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}
	queryRes := &common.QueryResponse{}
	if queryApiRes.Body != nil && queryApiRes.Body.GetRecords() != nil {
		for _, record := range queryApiRes.Body.GetRecords() {
			fields = append(fields, helpers.GetFormattedQueryRecord(*record))
			tokenizedData = append(tokenizedData, record.Tokens)
		}
		queryRes.Fields = fields
		queryRes.TokenizedData = tokenizedData
	}
	logger.Info(logs.QUERY_REQUEST_RESOLVED)
	logger.Info(logs.QUERY_SUCCESS)
	return queryRes, nil
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
	payload := vaultapis.RecordServiceUpdateRecordBody{}
	tokenMode, tokenError := helpers.SetTokenMode(options.TokenMode)
	if tokenError != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_BYOT)
	}
	payload.Byot = tokenMode
	payload.Tokenization = &options.ReturnTokens
	record := vaultapis.V1FieldRecords{}
	skyflowId, _ := helpers.GetSkyflowID(request.Data)
	delete(request.Data, constants.SKYFLOW_ID)
	record.Fields = request.Data
	if request.Tokens != nil {
		record.Tokens = request.Tokens
	}
	payload.Record = &record
	updateApiRes, apiErr := v.ApiClient.Records.WithRawResponse.RecordServiceUpdateRecord(ctx, v.Config.VaultId, request.Table, skyflowId, &payload)

	if apiErr != nil {
		logger.Error(logs.UPDATE_REQUEST_REJECTED)
		header, _ := helpers.GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}
	logger.Info(logs.UPDATE_REQUEST_RESOLVED)
	updateRes := &vaultapis.V1UpdateRecordResponse{}
	var id *string
	if updateApiRes.Body != nil && updateApiRes.Body.GetSkyflowId() != nil {
		updateRes = updateApiRes.Body
		id = updateApiRes.Body.GetSkyflowId()
	}
	res := helpers.GetFormattedUpdateRecord(*updateRes)
	logger.Info(logs.UPDATE_SUCCESS)
	return &common.UpdateResponse{
		Tokens:    res,
		SkyflowId: *id,
		Errors:    nil,
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
	payload := helpers.GetTokenizePayload(request)
	tokenizeApiRes, apiErr := v.ApiClient.Tokens.WithRawResponse.RecordServiceTokenize(ctx, v.Config.VaultId, &payload)

	if apiErr != nil {
		logger.Error(logs.TOKENIZE_REQUEST_REJECTED)
		header, _ := helpers.GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}
	tokenizeRes := &vaultapis.V1TokenizeResponse{}
	if tokenizeApiRes.Body != nil {
		tokenizeRes = tokenizeApiRes.Body
	}
	logger.Info(logs.TOKENIZE_SUCCESS)
	return helpers.ParseTokenizeResponse(*tokenizeRes), nil
}

func (v *VaultController) UploadFile(ctx context.Context, request common.FileUploadRequest) (*common.FileUploadResponse, *skyflowError.SkyflowError) {
	logger.Info(logs.UPLOAD_FILE_TRIGGERED)
	logger.Info(logs.VALIDATE_UPDATE_INPUT)
	// validate the request
	errs := validation.ValidateFileUploadRequest(request)
	if errs != nil {
		return nil, errs
	}

	if err := CreateRequestClientFunc(v); err != nil {
		return nil, err
	}
	file, fileObjError := helpers.GetFileForFileUpload(request)
	if fileObjError != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fileObjError.Error())
	}
	// create payload
	payload := vaultapis.UploadFileV2Request{}
	payload.File = file
	payload.SkyflowId = &request.SkyflowId
	payload.ColumnName = request.ColumnName
	payload.TableName = request.Table

	fileResp, fileErr := v.ApiClient.Records.WithRawResponse.UploadFileV2(ctx, v.Config.VaultId, &payload)

	if fileErr != nil {
		logger.Error(logs.UPLOAD_FILE_REQUEST_REJECTED)
		header, _ := helpers.GetHeader(fileErr)
		return nil, skyflowError.SkyflowErrorApi(fileErr, header)
	}
	logger.Info(logs.UPLOAD_FILE_REQUEST_RESOLVED)
	return &common.FileUploadResponse{
		SkyflowId: *fileResp.Body.GetSkyflowId(),
	}, nil
}
