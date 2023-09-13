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
		if tokens, ok := singleRecord["tokens"]; !ok {
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
		}
	}
	return nil
}

func (insertApi *InsertApi) Post(ctx context.Context, token string) (common.ResponseBody, *errors.SkyflowError) {
	err := insertApi.doValidations()
	if err != nil {
		return nil, err
	}
	if insertApi.Options.ContinueOnError {
		jsonRecord, _ := json.Marshal(insertApi.Records)
		var insertRecord common.InsertRecords
		if err := json.Unmarshal(jsonRecord, &insertRecord); err != nil {
			logger.Error(fmt.Sprintf(messages.INVALID_RECORDS, insertTag))
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_RECORDS, insertTag))
		}
		record, err := insertApi.constructBatchRequestBody(insertRecord, insertApi.Options)
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
		records := insertApi.arrangeRecords(insertApi.Records["records"].([]interface{}))
		recordssArray := records["RECORDS"].(map[string]interface{})
		var finalSuccess []interface{}
		var finalError []map[string]interface{}
		responseChannels := make([]chan map[string]interface{}, len(recordssArray))

		logger.Info(fmt.Sprintf(messages.INSERTING_RECORDS, insertTag, insertApi.Configuration.VaultID))
		i := 0
		for index := range recordssArray {
			responseChannel := make(chan map[string]interface{})
			responseChannels[i] = responseChannel

			tableName := index
			requestUrl := fmt.Sprintf("%s/v1/vaults/%s/%s", insertApi.Configuration.VaultURL, insertApi.Configuration.VaultID, index)
			var UniqueColumn = getUniqueColumn(index, insertApi.Options.Upsert)
			insertRecord := recordssArray[index].([]map[string]interface{})
			go func(i int, responseChannel chan<- map[string]interface{}) {
				record, err := insertApi.constructBulkRequestBody(insertRecord, insertApi.Options)
				if err == nil {
					record["upsert"] = UniqueColumn
					requestBody, err1 := json.Marshal(record)
					if err1 != nil {
						logger.Error(fmt.Sprintf(messages.EMPTY_RECORDS, insertTag))
						return
					}
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
					res, err := Client.Do(request)
					var requestId = ""
					if res != nil {
						requestId = res.Header.Get("x-request-id")
					}
					if err != nil {
						logger.Error(fmt.Sprintf(messages.SERVER_ERROR, insertTag, common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, insertTag, err), requestId)))
						var error = make(map[string]interface{})
						var errorObj = make(map[string]interface{})
						errorObj["code"] = "500"
						errorObj["description"] = common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, insertTag, err), requestId)
						error["error"] = errorObj
						var errorObject = make(map[string]interface{})
						errorObject = insertApi.addIndexInErrorObject(error, insertRecord)
						responseChannel <- errorObject
						return
					}
					data, _ := ioutil.ReadAll(res.Body)

					defer res.Body.Close()
					var result map[string]interface{}
					err = json.Unmarshal(data, &result)

					if err != nil {
						logger.Error(fmt.Sprintf(messages.SERVER_ERROR, insertTag, common.AppendRequestId(string(data), requestId)))
						var error = make(map[string]interface{})
						var errorObj = make(map[string]interface{})
						errorObj["code"] = "500"
						errorObj["description"] = fmt.Sprintf(messages.UNKNOWN_ERROR, insertTag, common.AppendRequestId(string(data), requestId))
						error["error"] = errorObj
						var errorObject = make(map[string]interface{})
						errorObject = insertApi.addIndexInErrorObject(error, insertRecord)
						responseChannel <- errorObject
					} else {
						errorResult := result["error"]
						if errorResult != nil {
							var generatedError = (errorResult).(map[string]interface{})
							var error = make(map[string]interface{})
							var errorObj = make(map[string]interface{})
							errorObj["code"] = fmt.Sprintf("%v", (errorResult.(map[string]interface{}))["http_code"])
							errorObj["description"] = common.AppendRequestId((generatedError["message"]).(string), requestId)
							error["error"] = errorObj
							var errorObject = make(map[string]interface{})
							errorObject = insertApi.addIndexInErrorObject(error, insertRecord)
							responseChannel <- errorObject
						} else {
							var record = make(map[string]interface{})
							record = insertApi.buildResponseWithoutContinueOnErr(result, tableName, insertRecord)
							delete(record, "valueType")
							responseChannel <- record
						}
					}
				}
			}(i, responseChannel)
			i++
		}
		for _, responseChan := range responseChannels {
			response := <-responseChan
			if _, found := response["errors"]; found {
				finalErrorsArray := response["errors"].([]interface{})
				for i := range finalErrorsArray {
					finalError = append(finalError, finalErrorsArray[i].(map[string]interface{}))
				}
			} else {
				finalArray := response["records"].([]interface{})
				for i := range finalArray {
					finalSuccess = append(finalSuccess, response["records"].([]interface{})[i])
				}
			}
		}

		var finalRecord = make(map[string]interface{})
		var Partial bool
		if len(finalSuccess) != 0 && (len(finalError) != 0) {
			Partial = true
		}
		if Partial {
			logger.Error(fmt.Sprintf(messages.PARTIAL_SUCCESS, insertTag))
		} else if len(finalSuccess) == 0 {
			logger.Error(fmt.Sprintf(messages.BATCH_INSERT_FAILURE, insertTag))
		} else {
			logger.Info(fmt.Sprintf(messages.INSERTING_RECORDS_SUCCESS, insertTag, insertApi.Configuration.VaultID))
		}

		if finalSuccess != nil {
			finalRecord["records"] = finalSuccess
		}
		if finalError != nil {
			finalRecord["errors"] = finalError
		}
		return finalRecord, nil
	}
}
func (InsertApi *InsertApi) constructBatchRequestBody(record common.InsertRecords, options common.InsertOptions) (map[string]interface{}, *errors.SkyflowError) {
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
	return body, nil
}
func (InsertApi *InsertApi) constructBulkRequestBody(record []map[string]interface{}, options common.InsertOptions) (map[string]interface{}, *errors.SkyflowError) {
	body := make(map[string]interface{})
	body["quorum"] = true
	body["records"] = record
	body["tokenization"] = options.Tokens
	return body, nil
}

func (InsertApi *InsertApi) arrangeRecords(recordsArray []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	recordGroups := make(map[string]interface{}) // Group by table

	for index, record := range recordsArray {
		rec := record.(map[string]interface{})
		table := rec["table"].(string)

		fieldsInterface, fieldsExists := rec["fields"]
		tokensInterface, tokensExists := rec["tokens"]
		var fields map[string]interface{}
		if fieldsExists && fieldsInterface != nil {
			fields = fieldsInterface.(map[string]interface{})
		} else {
			fields = make(map[string]interface{})
		}
		var tokens map[string]interface{}
		if tokensExists && tokensInterface != nil {
			tokens = tokensInterface.(map[string]interface{})
		} else {
			tokens = make(map[string]interface{})
		}
		group, exists := recordGroups[table]
		if !exists {
			group = make([]map[string]interface{}, 0)
		}

		// Combine fields and tokens maps
		combinedMap := map[string]interface{}{
			"fields": fields,
		}
		if len(tokens) != 0 {
			combinedMap["tokens"] = tokens
		}
		combinedMap["request_index"] = index

		group = append(group.([]map[string]interface{}), combinedMap)
		recordGroups[table] = group
	}

	result["RECORDS"] = recordGroups
	return result
}

func (InsertApi *InsertApi) addIndexInErrorObject(error map[string]interface{}, insertRecord []map[string]interface{}) common.ResponseBody {
	var errorArray = []interface{}{}
	var errorObject = make(map[string]interface{})

	for i := 0; i < len(insertRecord); i++ {
		var singleError = make(map[string]interface{})
		var errorObj = make(map[string]interface{})
		errorObj["code"] = error["error"].(map[string]interface{})["code"]
		errorObj["description"] = error["error"].(map[string]interface{})["description"]
		errorObj["request_index"] = insertRecord[i]["request_index"]
		singleError["error"] = errorObj
		errorArray = append(errorArray, singleError)
	}
	errorObject["errors"] = errorArray
	return errorObject
}
func (insertApi *InsertApi) buildResponseWithoutContinueOnErr(responseJson map[string]interface{}, tableName string, insertRecord []map[string]interface{}) common.ResponseBody {
	var recordsArray = []interface{}{}
	var records = responseJson["records"].([]interface{})
	var responseObject = make(map[string]interface{})
	if insertApi.Options.Tokens {
		for i := 0; i < len(records); i++ {
			index := insertRecord[i]["request_index"]
			result := make(map[string]interface{})
			id := records[i].(map[string]interface{})["skyflow_id"]
			tokens := records[i].(map[string]interface{})["tokens"]

			var fields = tokens.(map[string]interface{})
			fields["skyflow_id"] = id
			result["request_index"] = index
			result["fields"] = fields
			result["table"] = tableName
			recordsArray = append(recordsArray, result)
		}
	} else {
		for i := 0; i < len(records); i++ {
			index := insertRecord[i]["request_index"]
			var record = records[i]
			var newRecord = make(map[string]interface{})
			newRecord["request_index"] = index
			newRecord["table"] = tableName
			newRecord["fields"] = record
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
