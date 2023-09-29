/*
Copyright (c) 2022 Skyflow, Inc.
*/
package vaultapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type InsertApi struct {
	Configuration common.Configuration
	Records       map[string]interface{}
	Options       common.InsertOptions
}

var insertTag = "Insert"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client HTTPClient
)

func init() {
	Client = &http.Client{}
}

func (insertApi *InsertApi) doValidations() *errors.SkyflowError {

	var err = isValidVaultDetails(insertApi.Configuration)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf(messages.VALIDATE_RECORDS, insertTag))

	var totalRecords = insertApi.Records["records"]
	if totalRecords == nil {
		logger.Error(fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, insertTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, insertTag))
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		logger.Error(fmt.Sprintf(messages.EMPTY_RECORDS, insertTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, insertTag))
	}

	for _, upsertOption := range insertApi.Options.Upsert {
		var table = upsertOption.Table
		var column = upsertOption.Column

		if table == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_TABLE_IN_UPSERT_OPTIONS, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_IN_UPSERT_OPTIONS, insertTag))
		}
		if column == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_COLUMN_IN_UPSERT_OPTIONS, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_COLUMN_IN_UPSERT_OPTIONS, insertTag))
		}
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var table = singleRecord["table"]
		var fields = singleRecord["fields"]
		if table == nil {
			logger.Error(fmt.Sprintf(messages.MISSING_TABLE, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TABLE, insertTag))
		} else if table == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_TABLE_NAME, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_NAME, insertTag))
		} else if fields == nil {
			logger.Error(fmt.Sprintf(messages.FIELDS_KEY_ERROR, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.FIELDS_KEY_ERROR, insertTag))
		} else if fields == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_FIELDS, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_FIELDS, insertTag))
		}
		field := (singleRecord["fields"]).(map[string]interface{})
		if len(field) == 0 {
			logger.Error(fmt.Sprintf(messages.EMPTY_FIELDS, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_FIELDS, insertTag))
		}
		for index := range field {
			if index == "" {
				logger.Error(fmt.Sprintf(messages.EMPTY_COLUMN_NAME, insertTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_COLUMN_NAME, insertTag))
			}
		}
		if insertApi.Options.Byot != common.ENABLE && insertApi.Options.Byot != common.DISABLE && insertApi.Options.Byot != common.ENABLE_STRICT {
			logger.Error(fmt.Sprintf(messages.INVALID_BYOT_TYPE, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_BYOT_TYPE, insertTag))
		}
		if tokens, ok := singleRecord["tokens"]; !ok {
			if insertApi.Options.Byot == common.ENABLE || insertApi.Options.Byot == common.ENABLE_STRICT {
				logger.Error(fmt.Sprintf(messages.NO_TOKENS_IN_INSERT, insertTag, insertApi.Options.Byot))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.NO_TOKENS_IN_INSERT, insertTag, insertApi.Options.Byot))
			}
		} else if tokens == nil {
			logger.Error(fmt.Sprintf(messages.EMPTY_TOKENS_IN_INSERT, insertTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKENS_IN_INSERT, insertTag))
		} else if _, isString := tokens.(string); isString {
			logger.Error(fmt.Sprintf(messages.INVALID_TOKENS_IN_INSERT_RECORD, insertTag, reflect.TypeOf(tokens)))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_TOKENS_IN_INSERT_RECORD, insertTag, reflect.TypeOf(tokens)))
		} else if _, isMap := tokens.(map[string]interface{}); !isMap {
			logger.Error(fmt.Sprintf(messages.INVALID_TOKENS_IN_INSERT_RECORD, insertTag, reflect.TypeOf(tokens)))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_TOKENS_IN_INSERT_RECORD, insertTag, reflect.TypeOf(tokens)))
		} else {
			tokensMap, _ := tokens.(map[string]interface{})
			fieldsMap, _ := fields.(map[string]interface{})
			if len(tokensMap) == 0 {
				logger.Error(fmt.Sprintf(messages.EMPTY_TOKENS_IN_INSERT, insertTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKENS_IN_INSERT, insertTag))
			}
			for tokenKey := range tokensMap {
				if _, exists := fieldsMap[tokenKey]; !exists {
					logger.Error(fmt.Sprintf(messages.MISMATCH_OF_FIELDS_AND_TOKENS, insertTag))
					return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISMATCH_OF_FIELDS_AND_TOKENS, insertTag))
				}
			}
			if insertApi.Options.Byot == common.DISABLE {
				logger.Error(fmt.Sprintf(messages.TOKENS_PASSED_FOR_BYOT_DISABLE, insertTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.TOKENS_PASSED_FOR_BYOT_DISABLE, insertTag))
			}
			if len(fieldsMap) != len(tokensMap) && insertApi.Options.Byot == common.ENABLE_STRICT {
				logger.Error(fmt.Sprintf(messages.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT, insertTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT, insertTag))
			}
		}

	}
	return nil
}

func (insertApi *InsertApi) Post(ctx context.Context, token string) (common.ResponseBody, *errors.SkyflowError) {
	err := insertApi.doValidations()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(insertApi.Records)
	var insertRecord common.InsertRecords
	if err := json.Unmarshal(jsonRecord, &insertRecord); err != nil {
		logger.Error(fmt.Sprintf(messages.INVALID_RECORDS, insertTag))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_RECORDS, insertTag))
	}

	record, err := insertApi.constructRequestBody(insertRecord, insertApi.Options)
	if err != nil {
		return nil, err
	}
	requestBody, err1 := json.Marshal(record)
	if err1 != nil {
		logger.Error(fmt.Sprintf(messages.EMPTY_RECORDS, insertTag))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, insertTag, err1))
	}
	requestUrl := fmt.Sprintf("%s/v1/vaults/%s", insertApi.Configuration.VaultURL, insertApi.Configuration.VaultID)
	var request *http.Request
	if ctx != nil {
		request, _ = http.NewRequestWithContext(
			ctx,
			"POST",
			requestUrl,
			strings.NewReader(string(requestBody)),
		)
	} else {
		request, _ = http.NewRequest(
			"POST",
			requestUrl,
			strings.NewReader(string(requestBody)),
		)
	}
	bearerToken := fmt.Sprintf("Bearer %s", token)
	request.Header.Add("Authorization", bearerToken)
	skyMetadata := common.CreateJsonMetadata()
	request.Header.Add("sky-metadata", skyMetadata)
	logger.Info(fmt.Sprintf(messages.INSERTING_RECORDS, insertTag, insertApi.Configuration.VaultID))
	res, err2 := Client.Do(request)
	var requestId = ""
	var code = "500"
	if res != nil {
		requestId = res.Header.Get("x-request-id")
		code = strconv.Itoa(res.StatusCode)
	}
	if err2 != nil {
		logger.Error(fmt.Sprintf(messages.SERVER_ERROR, insertTag, common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, insertTag, err2), requestId)))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(code), common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, insertTag, err2), requestId))
	}

	data, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	var result map[string]interface{}
	err2 = json.Unmarshal(data, &result)
	if insertApi.Options.ContinueOnError {
		if err2 != nil {
			logger.Error(fmt.Sprintf(messages.SERVER_ERROR, insertTag, common.AppendRequestId(string(data), requestId)))
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, insertTag, common.AppendRequestId(string(data), requestId)))
		}
		response, Partial := insertApi.buildResponseWithContinueOnErr((result["responses"]).([]interface{}), insertRecord, requestId)
		if Partial {
			logger.Error(fmt.Sprintf(messages.PARTIAL_SUCCESS, insertTag))
		} else if len(response["records"].([]interface{})) == 0 {
			logger.Error(fmt.Sprintf(messages.BATCH_INSERT_FAILURE, insertTag))
		} else {
			logger.Info(fmt.Sprintf(messages.INSERTING_RECORDS_SUCCESS, insertTag, insertApi.Configuration.VaultID))
		}
		return response, nil
	} else {
		if err2 != nil {
			logger.Error(fmt.Sprintf(messages.SERVER_ERROR, insertTag, common.AppendRequestId(string(data), requestId)))
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, insertTag, common.AppendRequestId(string(data), requestId)))
		} else if result["error"] != nil {
			var generatedError = (result["error"]).(map[string]interface{})
			logger.Error(fmt.Sprintf(messages.SERVER_ERROR, insertTag, common.AppendRequestId(generatedError["message"].(string), requestId)))
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(fmt.Sprintf("%v", generatedError["http_code"])), fmt.Sprintf(messages.SERVER_ERROR, insertTag, common.AppendRequestId(generatedError["message"].(string), requestId)))
		}
		logger.Info(fmt.Sprintf(messages.INSERTING_RECORDS_SUCCESS, insertTag, insertApi.Configuration.VaultID))
		return insertApi.buildResponseWithoutContinueOnErr((result["responses"]).([]interface{}), insertRecord), nil
	}
}

func (InsertApi *InsertApi) constructRequestBody(record common.InsertRecords, options common.InsertOptions) (map[string]interface{}, *errors.SkyflowError) {
	postPayload := []interface{}{}
	records := record.Records
	for _, value := range records {
		singleRecord := value
		table := singleRecord.Table
		fields := singleRecord.Fields
		tokens := singleRecord.Tokens
		var UniqueColumn = getUniqueColumn(singleRecord.Table, options.Upsert)
		var finalRecord = make(map[string]interface{})
		finalRecord["tableName"] = table
		finalRecord["fields"] = fields
		finalRecord["tokens"] = tokens
		finalRecord["method"] = "POST"
		finalRecord["quorum"] = true
		if options.Upsert != nil {
			finalRecord["upsert"] = UniqueColumn
		}

		finalRecord["tokenization"] = options.Tokens
		postPayload = append(postPayload, finalRecord)
	}
	body := make(map[string]interface{})
	body["records"] = postPayload
	if options.ContinueOnError {
		body["continueOnError"] = options.ContinueOnError
	}
	body["byot"] = options.Byot
	return body, nil
}

func (insertApi *InsertApi) buildResponseWithoutContinueOnErr(responseJson []interface{}, requestRecords common.InsertRecords) common.ResponseBody {

	var inputRecords = requestRecords.Records
	var recordsArray = []interface{}{}
	var responseObject = make(map[string]interface{})
	if insertApi.Options.Tokens {
		for i := 0; i < len(responseJson); i = i + 1 {
			var mainRecord = responseJson[i].(map[string]interface{})
			var record = mainRecord["records"].([]interface{})[0]
			id := record.(map[string]interface{})["skyflow_id"]
			tokens := record.(map[string]interface{})["tokens"]

			var inputRecord = inputRecords[i]
			records := map[string]interface{}{}
			var fields = tokens.(map[string]interface{})
			fields["skyflow_id"] = id
			records["request_index"] = i
			records["fields"] = fields
			records["table"] = inputRecord.Table
			recordsArray = append(recordsArray, records)
		}
	} else {
		for i := 0; i < len(responseJson); i++ {
			var inputRecord = inputRecords[i]
			var record = ((responseJson[i]).(map[string]interface{})["records"]).([]interface{})
			var newRecord = make(map[string]interface{})
			newRecord["request_index"] = i
			newRecord["table"] = inputRecord.Table
			newRecord["fields"] = record[0]
			recordsArray = append(recordsArray, newRecord)

		}
	}
	responseObject["records"] = recordsArray

	return responseObject
}
func (insertApi *InsertApi) buildResponseWithContinueOnErr(responseJson []interface{}, requestRecords common.InsertRecords, requestId string) (common.ResponseBody, bool) {
	var inputRecords = requestRecords.Records
	var Partial = false
	var recordsArray = []interface{}{}
	var errorsArray = []interface{}{}
	var responseObject = make(map[string]interface{})
	for i := 0; i < len(responseJson); i = i + 1 {
		var mainRecord = responseJson[i].(map[string]interface{})
		var getBody = mainRecord["Body"].(map[string]interface{})
		if _, ok := getBody["records"]; ok {
			var record = getBody["records"].([]interface{})[0]
			id := record.(map[string]interface{})["skyflow_id"]

			var inputRecord = inputRecords[i]
			records := map[string]interface{}{}
			if insertApi.Options.Tokens {
				tokens := record.(map[string]interface{})["tokens"]
				var fields = tokens.(map[string]interface{})
				fields["skyflow_id"] = id
				records["request_index"] = i
				records["fields"] = fields
				records["table"] = inputRecord.Table
				recordsArray = append(recordsArray, records)
			} else {
				var fields = make(map[string]interface{})
				fields["skyflow_id"] = id
				records["request_index"] = i
				records["fields"] = fields
				records["table"] = inputRecord.Table
				recordsArray = append(recordsArray, records)
			}
		} else if _, ok := getBody["error"]; ok {
			var StatusCode = mainRecord["Status"].(float64)
			var error = getBody["error"]
			errorsObj := map[string]interface{}{}
			var errorObj = make(map[string]interface{})
			errorObj["request_index"] = i
			errorObj["description"] = common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, insertTag, error), requestId)
			errorObj["code"] = strconv.FormatFloat(StatusCode, 'f', -1, 64)
			errorsObj["error"] = errorObj
			errorsArray = append(errorsArray, errorsObj)
		}
	}
	if errorsArray != nil {
		responseObject["errors"] = errorsArray
	}
	if recordsArray != nil {
		responseObject["records"] = recordsArray
	}
	if len(recordsArray) != 0 && (len(errorsArray) != 0) {
		Partial = true
	}
	return responseObject, Partial
}
func getUniqueColumn(table string, upsertArray []common.UpsertOptions) string {
	var UniqueColumn string
	for _, eachOption := range upsertArray {
		if eachOption.Table == table {
			UniqueColumn = eachOption.Column
		}
	}
	return UniqueColumn
}
