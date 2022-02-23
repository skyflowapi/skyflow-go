package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type DetokenizeApi struct {
	Configuration common.Configuration
	Records       map[string]interface{}
	Token         string
}

var detokenizeTag = "Detokenize"

func (detokenize *DetokenizeApi) Get() (map[string]interface{}, *errors.SkyflowError) {

	err := detokenize.doValidations()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(detokenize.Records)
	var detokenizeRecord common.DetokenizeInput
	if err := json.Unmarshal(jsonRecord, &detokenizeRecord); err != nil {
		logger.Error(fmt.Sprintf(messages.INVALID_RECORDS, detokenizeTag))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_RECORDS, detokenizeTag))
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

	logger.Info(fmt.Sprintf(messages.VALIDATE_DETOKENIZE_INPUT, detokenizeTag))

	var totalRecords = detokenizeApi.Records["records"]
	if totalRecords == nil {
		logger.Error(fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, detokenizeTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, detokenizeTag))
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		logger.Error(fmt.Sprintf(messages.EMPTY_RECORDS, detokenizeTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, detokenizeTag))
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var token = singleRecord["token"]
		if token == nil {
			logger.Error(fmt.Sprintf(messages.MISSING_TOKEN, detokenizeTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TOKEN, detokenizeTag))
		} else if token == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_TOKEN_ID, detokenizeTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKEN_ID, detokenizeTag))
		}
	}
	return nil

}

func (detokenize *DetokenizeApi) sendRequest(records common.DetokenizeInput) (map[string]interface{}, *errors.SkyflowError) {

	var finalSuccess []map[string]interface{}
	var finalError []map[string]interface{}

	responseChannel := make(chan map[string]interface{})

	for i := 0; i < len(records.Records); i++ {
		logger.Info(fmt.Sprintf(messages.DETOKENIZING_RECORDS, detokenizeTag, records.Records[i].Token))
		go func(i int, responseChannel chan map[string]interface{}) {
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

				res, err := Client.Do(request)
				var requestId = ""
				if res != nil {
					requestId = res.Header.Get("x-request-id")
				}
				if err != nil {
					logger.Error(fmt.Sprintf(messages.DETOKENIZING_FAILED, detokenizeTag, common.AppendRequestId(singleRecord.Token, requestId)))
					var error = make(map[string]interface{})
					var errorObj = make(map[string]interface{})
					errorObj["code"] = "500"
					errorObj["description"] = common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, detokenizeTag, err), requestId)
					error["error"] = errorObj
					error["token"] = singleRecord.Token
					responseChannel <- error
					return
				}
				data, _ := ioutil.ReadAll(res.Body)
				defer res.Body.Close()
				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				if err != nil {
					logger.Error(fmt.Sprintf(messages.DETOKENIZING_FAILED, detokenizeTag, common.AppendRequestId(singleRecord.Token, requestId)))
					var error = make(map[string]interface{})
					var errorObj = make(map[string]interface{})
					errorObj["code"] = "500"
					errorObj["description"] = fmt.Sprintf(messages.UNKNOWN_ERROR, detokenizeTag, common.AppendRequestId(string(data), requestId))
					error["error"] = errorObj
					error["token"] = singleRecord.Token
					responseChannel <- error
				} else {
					errorResult := result["error"]
					if errorResult != nil {
						logger.Error(fmt.Sprintf(messages.DETOKENIZING_FAILED, detokenizeTag, common.AppendRequestId(singleRecord.Token, requestId)))
						var generatedError = (errorResult).(map[string]interface{})
						var error = make(map[string]interface{})
						var errorObj = make(map[string]interface{})
						errorObj["code"] = fmt.Sprintf("%v", (errorResult.(map[string]interface{}))["http_code"])
						errorObj["description"] = common.AppendRequestId((generatedError["message"]).(string), requestId)
						error["error"] = errorObj
						error["token"] = singleRecord.Token
						responseChannel <- error
					} else {
						logger.Info(fmt.Sprintf(messages.DETOKENIZING_SUCCESS, detokenizeTag, singleRecord.Token))
						var generatedResult = (result["records"]).([]interface{})
						var record = (generatedResult[0]).(map[string]interface{})
						delete(record, "valueType")
						responseChannel <- record
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
			finalSuccess = append(finalSuccess, response)
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
