package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type GetByIdApi struct {
	Configuration common.Configuration
	Records       map[string]interface{}
	Token         string
}

func (g *GetByIdApi) Get() (map[string]interface{}, *errors.SkyflowError) {

	err := g.doValidations()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(g.Records)
	var getByIdRecord common.GetByIdInput
	if err := json.Unmarshal(jsonRecord, &getByIdRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.INVALID_RECORDS)
	}
	res, err := g.doRequest(getByIdRecord)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (g *GetByIdApi) doValidations() *errors.SkyflowError {
	var err = isValidVaultDetails(g.Configuration)
	if err != nil {
		return err
	}

	logger.Info(messages.VALIDATE_GET_BY_ID_INPUT)

	var totalRecords = g.Records["records"]
	if totalRecords == nil {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.RECORDS_KEY_NOT_FOUND)
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.EMPTY_RECORDS)
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var table = singleRecord["table"]
		var ids = singleRecord["ids"]
		var redaction = singleRecord["redaction"]
		//var redactionInRecord = (redaction).(string)
		if table == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.MISSING_TABLE)
		} else if table == "" {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.EMPTY_TABLE_NAME)
		} else if ids == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.MISSING_KEY_IDS)
		} else if ids == "" {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.EMPTY_RECORD_IDS)
		} else if redaction == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.MISSING_REDACTION)
		}
		// else if redactionInRecord != RedactionType.PLAIN_TEXT || redactionInRecord != DEFAULT || redactionInRecord != REDACTED || redactionInRecord != MASKED {
		// 	return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.Default), errors.INVALID_REDACTION_TYPE)
		// }
		idArray := (ids).([]interface{})
		if len(idArray) == 0 {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.EMPTY_RECORD_IDS)
		}
		for index := range idArray {
			if idArray[index] == "" {
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.EMPTY_TOKEN_ID)
			}
		}
	}
	return nil
}

func (g *GetByIdApi) doRequest(records common.GetByIdInput) (map[string]interface{}, *errors.SkyflowError) {

	var finalSuccess []interface{}
	var finalError []map[string]interface{}

	responseChannel := make(chan map[string]interface{})

	for i := 0; i < len(records.Records); i++ {
		logger.Info(fmt.Sprintf(messages.GETTING_RECORDS_BY_ID, records.Records[i].Table))
		go func(i int, responseChannel chan map[string]interface{}) {
			singleRecord := records.Records[i]
			requestUrl := fmt.Sprintf("%s/v1/vaults/%s/%s", g.Configuration.VaultURL, g.Configuration.VaultID, singleRecord.Table)
			url1, err := url.Parse(requestUrl)
			v := url.Values{}
			for j := 0; j < len(singleRecord.Ids); j++ {
				v.Add("skyflow_ids", singleRecord.Ids[j])
			}
			v.Add("redaction", string(singleRecord.Redaction))
			url1.RawQuery = v.Encode()
			if err == nil {
				request, _ := http.NewRequest(
					"GET",
					url1.String(),
					strings.NewReader(""),
				)
				bearerToken := fmt.Sprintf("Bearer %s", g.Token)
				request.Header.Add("Authorization", bearerToken)

				res, err := Client.Do(request)

				if err != nil {
					logger.Error(fmt.Sprintf(messages.GET_RECORDS_BY_ID_FAILED, singleRecord.Table))
					var error = make(map[string]interface{})
					error["error"] = fmt.Sprintf(messages.SERVER_ERROR, err)
					error["ids"] = singleRecord.Ids
					responseChannel <- error
					//continue
					return
				}
				data, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()
				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				if err != nil {
					logger.Error(fmt.Sprintf(messages.GET_RECORDS_BY_ID_FAILED, singleRecord.Table))
					var error = make(map[string]interface{})
					error["error"] = fmt.Sprintf(messages.UNKNOWN_ERROR, string(data))
					error["ids"] = singleRecord.Ids
					responseChannel <- error
				} else {
					errorResult := result["error"]
					if errorResult != nil {
						logger.Error(fmt.Sprintf(messages.GET_RECORDS_BY_ID_FAILED, singleRecord.Table))
						var generatedError = (errorResult).(map[string]interface{})
						var error = make(map[string]interface{})
						error["error"] = generatedError["message"]
						error["ids"] = singleRecord.Ids
						responseChannel <- error

					} else {
						logger.Info(fmt.Sprintf(messages.GET_RECORDS_BY_ID_SUCCESS, singleRecord.Table))
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
