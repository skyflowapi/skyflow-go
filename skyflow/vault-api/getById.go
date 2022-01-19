package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/skyflowapi/skyflow-go/errors"
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
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.INVALID_RECORDS)
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
	var totalRecords = g.Records["records"]
	if totalRecords == nil {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.RECORDS_KEY_NOT_FOUND)
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORDS)
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var table = singleRecord["table"]
		var ids = singleRecord["ids"]
		var redaction = singleRecord["redaction"]
		//var redactionInRecord = (redaction).(string)
		if table == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_TABLE)
		} else if table == "" {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TABLE_NAME)
		} else if ids == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_KEY_IDS)
		} else if ids == "" {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORD_IDS)
		} else if redaction == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_REDACTION)
		}
		// else if redactionInRecord != RedactionType.PLAIN_TEXT || redactionInRecord != DEFAULT || redactionInRecord != REDACTED || redactionInRecord != MASKED {
		// 	return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.Default), errors.INVALID_REDACTION_TYPE)
		// }
		idArray := (ids).([]interface{})
		if len(idArray) == 0 {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_FIELDS)
		}
		for index := range idArray {
			if idArray[index] == "" {
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TOKEN_ID)
			}
		}
	}
	return nil
}

func (g *GetByIdApi) doRequest(records common.GetByIdInput) (map[string]interface{}, *errors.SkyflowError) {

	var wg = sync.WaitGroup{}
	var finalSuccess []interface{}
	var finalError []map[string]interface{}
	for i := 0; i < len(records.Records); i++ {
		wg.Add(1)
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
			res, err := http.DefaultClient.Do(request)
			if err != nil {
				var error = make(map[string]interface{})
				error["error"] = fmt.Sprintf(errors.SERVER_ERROR, err)
				error["ids"] = singleRecord.Ids
				finalError = append(finalError, error)
				continue
			}
			data, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				var error = make(map[string]interface{})
				error["error"] = fmt.Sprintf(errors.UNKNOWN_ERROR, string(data))
				error["ids"] = singleRecord.Ids
				finalError = append(finalError, error)
			} else {
				errorResult := result["error"]
				if errorResult != nil {
					var generatedError = (errorResult).(map[string]interface{})
					var error = make(map[string]interface{})
					error["error"] = generatedError["message"]
					error["ids"] = singleRecord.Ids
					finalError = append(finalError, error)

				} else {
					records := (result["records"]).([]interface{})
					for k := 0; k < len(records); k++ {
						new := make(map[string]interface{})
						single := (records[k]).(map[string]interface{})
						fields := (single["fields"]).(map[string]interface{})
						fields["id"] = fields["skyflow_id"]
						delete(fields, "skyflow_id")
						new["fields"] = fields
						new["table"] = singleRecord.Table
						finalSuccess = append(finalSuccess, new)
					}
				}

			}
		}
		wg.Done()
	}

	wg.Wait()
	var finalRecord = make(map[string]interface{})
	if finalSuccess != nil {
		finalRecord["success"] = finalSuccess
	}
	if finalError != nil {
		finalRecord["errors"] = finalError
	}
	return finalRecord, nil
}