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
	"net/url"
	"reflect"
	"strings"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type GetApi struct {
	Configuration common.Configuration
	Records       map[string]interface{}
	Options       common.GetOptions
	Token         string
}

var getTag = "Get"

func (g *GetApi) GetRecords(ctx context.Context) (map[string]interface{}, *errors.SkyflowError) {
	err := g.doValidations()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(g.Records)
	var getRecord common.GetInput
	if err := json.Unmarshal(jsonRecord, &getRecord); err != nil {
		logger.Error(fmt.Sprintf(messages.INVALID_RECORDS, getTag))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_RECORDS, getTag))
	}
	res, err := g.doRequest(ctx, getRecord, g.Options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (g *GetApi) doValidations() *errors.SkyflowError {
	var err = isValidVaultDetails(g.Configuration)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf(messages.VALIDATE_GET_INPUT, getTag))

	var totalRecords = g.Records["records"]
	if totalRecords == nil {
		logger.Error(fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, getTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, getTag))
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		logger.Error(fmt.Sprintf(messages.EMPTY_RECORDS, getTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, getTag))
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var table = singleRecord["table"]
		var ids = singleRecord["ids"]
		var redaction = singleRecord["redaction"]
		var columnName = singleRecord["columnName"]
		var columnValues = singleRecord["columnValues"]
		if table == nil {
			logger.Error(fmt.Sprintf(messages.MISSING_TABLE, getTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TABLE, getTag))
		} else if table == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_TABLE_NAME, getTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_NAME, getTag))
		} else if reflect.TypeOf(table).Kind() != reflect.String {
			logger.Error(fmt.Sprintf(messages.INVALID_TABLE_NAME_TYPE, getTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_TABLE_NAME_TYPE, getTag))
		}
		if ids != nil {
			if ids == "" {
				logger.Error(fmt.Sprintf(messages.EMPTY_RECORD_IDS, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getTag))
			}
			if reflect.TypeOf(ids).Kind() != reflect.Slice {
				logger.Error(fmt.Sprintf(messages.INVALID_IDS_TYPE, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_IDS_TYPE, getTag))
			}
			idArray := (ids).([]interface{})
			if len(idArray) == 0 {
				logger.Error(fmt.Sprintf(messages.EMPTY_RECORD_IDS, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getTag))
			}
			for index := range idArray {
				if idArray[index] == "" {
					logger.Error(fmt.Sprintf(messages.EMPTY_TOKEN_ID, getTag))
					return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKEN_ID, getTag))
				}
			}
		}

		if g.Options.Tokens == false {
			if redaction == nil {
				logger.Error(fmt.Sprintf(messages.MISSING_REDACTION, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_REDACTION, getTag))
			} else if redaction != common.PLAIN_TEXT && redaction != common.DEFAULT && redaction != common.REDACTED && redaction != common.MASKED {
				logger.Error(fmt.Sprintf(messages.INVALID_REDACTION_TYPE, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_REDACTION_TYPE, getTag))
			}
		} else if g.Options.Tokens == true {
			if columnName != nil || columnValues != nil {
				logger.Error(fmt.Sprintf(messages.TOKENS_GET_COLUMN_NOT_SUPPORTED, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.TOKENS_GET_COLUMN_NOT_SUPPORTED, getTag))
			} else {
				if redaction != nil && (redaction == common.PLAIN_TEXT || redaction == common.DEFAULT || redaction == common.REDACTED || redaction == common.MASKED) {
					logger.Error(fmt.Sprintf(messages.REDACTION_WITH_TOKEN_NOT_SUPPORTED, getTag))
					return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.REDACTION_WITH_TOKEN_NOT_SUPPORTED, getTag))
				}
			}
		}
		if columnName == nil {
			if ids == nil && columnValues == nil {
				logger.Error(fmt.Sprintf(messages.MISSING_IDS_OR_COLUMN_VALUES_IN_GET, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_IDS_OR_COLUMN_VALUES_IN_GET, getTag))

			}
		} else if columnName != nil && columnValues == nil {
			logger.Error(fmt.Sprintf(messages.MISSING_RECORD_COLUMN_VALUE, getTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_RECORD_COLUMN_VALUE, getTag))
		} else if ids != nil && columnName != nil {
			logger.Error(fmt.Sprintf(messages.SKYFLOW_IDS_AND_COLUMN_NAME_BOTH_SPECIFIED, getTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.SKYFLOW_IDS_AND_COLUMN_NAME_BOTH_SPECIFIED, getTag))

		}
		if columnValues != nil {
			if columnValues == "" {
				logger.Error(fmt.Sprintf(messages.EMPTY_RECORD_COLUMN_VALUES, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_COLUMN_VALUES, getTag))
			}
			if reflect.TypeOf(columnValues).Kind() != reflect.Slice {
				logger.Error(fmt.Sprintf(messages.INVALID_COLUMN_VALUES_IN_GET, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_COLUMN_VALUES_IN_GET, getTag))
			}
			columnValuesArray := (columnValues).([]interface{})
			if len(columnValuesArray) == 0 {
				logger.Error(fmt.Sprintf(messages.EMPTY_RECORD_COLUMN_VALUES, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_COLUMN_VALUES, getTag))
			}
			for index := range columnValuesArray {
				if columnValuesArray[index] == "" {
					logger.Error(fmt.Sprintf(messages.EMPTY_COLUMN_VALUE, getTag))
					return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_COLUMN_VALUE, getTag))
				}
			}
			if columnName == nil {
				logger.Error(fmt.Sprintf(messages.MISSING_COLUMN_NAME, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_COLUMN_NAME, getTag))

			} else if columnName == "" {
				logger.Error(fmt.Sprintf(messages.EMPTY_COLUMN_NAME, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_COLUMN_NAME, getTag))
			} else if reflect.TypeOf(columnName).Kind() != reflect.String {
				logger.Error(fmt.Sprintf(messages.INVALID_COLUMN_NAME, getTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_COLUMN_NAME, getTag))
			}
		}
	}
	return nil
}

func (g *GetApi) doRequest(ctx context.Context, records common.GetInput, options common.GetOptions) (map[string]interface{}, *errors.SkyflowError) {

	var finalSuccess []interface{}
	var finalError []map[string]interface{}
	responseChannel := make(chan map[string]interface{})
	for i := 0; i < len(records.Records); i++ {
		logger.Info(fmt.Sprintf(messages.GETTING_RECORDS_BY_ID, getTag, records.Records[i].Table))
		go func(i int, responseChannel chan map[string]interface{}) {
			singleRecord := records.Records[i]
			requestUrl := fmt.Sprintf("%s/v1/vaults/%s/%s", g.Configuration.VaultURL, g.Configuration.VaultID, singleRecord.Table)
			url1, err := url.Parse(requestUrl)
			v := url.Values{}

			if singleRecord.ColumnName != "" {
				for j := 0; j < len(singleRecord.ColumnValues); j++ {
					v.Add("column_values", singleRecord.ColumnValues[j])
				}
				v.Add("column_name", singleRecord.ColumnName)
			} else {
				for j := 0; j < len(singleRecord.Ids); j++ {
					v.Add("skyflow_ids", singleRecord.Ids[j])
				}
			}
			if !options.Tokens {
				v.Add("redaction", string(singleRecord.Redaction))
			} else {
				v.Add("tokenization", fmt.Sprintf("%v", options.Tokens))

			}
			url1.RawQuery = v.Encode()
			if err == nil {
				var request *http.Request
				if ctx != nil {
					request, _ = http.NewRequestWithContext(
						ctx,
						"GET",
						url1.String(),
						strings.NewReader(""),
					)
				} else {
					request, _ = http.NewRequest(
						"GET",
						url1.String(),
						strings.NewReader(""),
					)
				}
				bearerToken := fmt.Sprintf("Bearer %s", g.Token)
				request.Header.Add("Authorization", bearerToken)
				skyMetadata := common.CreateJsonMetadata()
				request.Header.Add("sky-metadata", skyMetadata)
				res, err := Client.Do(request)

				var requestId = ""
				if res != nil {
					requestId = res.Header.Get("x-request-id")
				}
				if err != nil {
					logger.Error(fmt.Sprintf(messages.GET_RECORDS_FAILED, getTag, common.AppendRequestId(singleRecord.Table, requestId)))
					var error = make(map[string]interface{})
					var errorObj = make(map[string]interface{})
					errorObj["code"] = "500"
					errorObj["description"] = common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, getTag, err), requestId)
					error["error"] = errorObj
					if singleRecord.Ids != nil {
						error["ids"] = singleRecord.Ids
					} else {
						error["columnValues"] = singleRecord.ColumnValues
						error["columnName"] = singleRecord.ColumnName

					}
					responseChannel <- error
					return
				}
				data, _ := ioutil.ReadAll(res.Body)
				defer res.Body.Close()
				var result map[string]interface{}
				err = json.Unmarshal(data, &result)

				if err != nil {
					logger.Error(fmt.Sprintf(messages.GET_RECORDS_FAILED, getTag, common.AppendRequestId(singleRecord.Table, requestId)))
					var error = make(map[string]interface{})
					var errorObj = make(map[string]interface{})
					errorObj["code"] = "500"
					errorObj["description"] = fmt.Sprintf(messages.UNKNOWN_ERROR, getTag, common.AppendRequestId(string(data), requestId))
					error["error"] = errorObj
					if singleRecord.Ids != nil {
						error["ids"] = singleRecord.Ids
					} else {
						error["columnValues"] = singleRecord.ColumnValues
						error["columnName"] = singleRecord.ColumnName

					}
					responseChannel <- error
				} else {
					errorResult := result["error"]
					if errorResult != nil {
						logger.Error(fmt.Sprintf(messages.GET_RECORDS_FAILED, getTag, common.AppendRequestId(singleRecord.Table, requestId)))
						var generatedError = (errorResult).(map[string]interface{})
						var error = make(map[string]interface{})
						var errorObj = make(map[string]interface{})
						errorObj["code"] = fmt.Sprintf("%v", (errorResult.(map[string]interface{}))["http_code"])
						errorObj["description"] = common.AppendRequestId(generatedError["message"].(string), requestId)
						error["error"] = errorObj
						if singleRecord.Ids != nil {
							error["ids"] = singleRecord.Ids
						} else {
							error["columnValues"] = singleRecord.ColumnValues
							error["columnName"] = singleRecord.ColumnName

						}
						responseChannel <- error
					} else {
						logger.Info(fmt.Sprintf(messages.GET_RECORDS_SUCCESS, getTag, singleRecord.Table))
						responseObj := make(map[string]interface{})
						var responseArr []interface{}

						records := (result["records"]).([]interface{})
						for k := 0; k < len(records); k++ {
							new := make(map[string]interface{})
							single := (records[k]).(map[string]interface{})
							fields := (single["fields"]).(map[string]interface{})
							fields["id"] = fields["skyflow_id"]
							delete(fields, "skyflow_id")
							new["fields"] = fields
							new["table"] = singleRecord.Table
							responseArr = append(responseArr, new)
						}
						responseObj["records"] = responseArr
						responseChannel <- responseObj
					}

				}
			}
		}(i, responseChannel)
	}

	for i := 0; i < len(records.Records); i++ {
		response := <-responseChannel
		if _, found := response["error"]; found {
			finalError = append(finalError, response)
		} else {
			finalSuccess = append(finalSuccess, response["records"].([]interface{})...)
		}
	}

	var finalRecord = make(map[string]interface{})
	if finalSuccess != nil {
		finalRecord["records"] = finalSuccess
	}
	if finalError != nil {
		finalRecord["errors"] = finalError
	}
	return finalRecord, nil
}
