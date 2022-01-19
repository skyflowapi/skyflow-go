package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/skyflowapi/skyflow-go/errors"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type DetokenizeApi struct {
	Configuration common.Configuration
	Records       map[string]interface{}
	Token         string
}

func (detokenize *DetokenizeApi) Get() (map[string]interface{}, *errors.SkyflowError) {

	err := detokenize.doValidations()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(detokenize.Records)
	var detokenizeRecord common.DetokenizeInput
	if err := json.Unmarshal(jsonRecord, &detokenizeRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.INVALID_RECORDS)
	}
	res, err := detokenize.sendRequest(detokenizeRecord)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (detokenizeApi *DetokenizeApi) doValidations() *errors.SkyflowError {
	var err = isValidVaultDetails(detokenizeApi.Configuration)
	if err != nil {
		return err
	}
	var totalRecords = detokenizeApi.Records["records"]
	if totalRecords == nil {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.RECORDS_KEY_NOT_FOUND)
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORDS)
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var token = singleRecord["token"]
		if token == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_TOKEN)
		} else if token == "" {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TOKEN_ID)
		}
	}
	return nil

}

func (detokenize *DetokenizeApi) sendRequest(records common.DetokenizeInput) (map[string]interface{}, *errors.SkyflowError) {

	var wg = sync.WaitGroup{}
	var finalSuccess []map[string]interface{}
	var finalError []map[string]interface{}
	for i := 0; i < len(records.Records); i++ {
		wg.Add(1)
		singleRecord := records.Records[i]
		requestUrl := fmt.Sprintf("%s/v1/vaults/%s/detokenize", detokenize.Configuration.VaultURL, detokenize.Configuration.VaultID)
		var detokenizeParameter = make(map[string]interface{})
		detokenizeParameter["token"] = singleRecord.Token
		var body = make(map[string]interface{})
		var params []map[string]interface{}
		params = append(params, detokenizeParameter)
		body["detokenizationParameters"] = params
		requestBody, err := json.Marshal(body)
		if err == nil {
			request, _ := http.NewRequest(
				"POST",
				requestUrl,
				strings.NewReader(string(requestBody)),
			)
			bearerToken := fmt.Sprintf("Bearer %s", detokenize.Token)
			request.Header.Add("Authorization", bearerToken)
			res, err := http.DefaultClient.Do(request)
			if err != nil {
				fmt.Println("server error")
				var error = make(map[string]interface{})
				error["error"] = fmt.Sprintf(errors.SERVER_ERROR, err)
				error["token"] = singleRecord.Token
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
				error["token"] = singleRecord.Token
				finalError = append(finalError, error)
			} else {
				errorResult := result["error"]
				if errorResult != nil {
					var generatedError = (errorResult).(map[string]interface{})
					var error = make(map[string]interface{})
					error["error"] = generatedError["message"]
					error["token"] = singleRecord.Token
					finalError = append(finalError, error)

				} else {
					var generatedResult = (result["records"]).([]interface{})
					var record = (generatedResult[0]).(map[string]interface{})
					delete(record, "valueType")
					finalSuccess = append(finalSuccess, record)
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
