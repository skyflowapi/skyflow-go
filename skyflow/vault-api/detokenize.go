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

func (detokenize *DetokenizeApi) Get() (map[string]interface{}, *errors.SkyflowError) {

	// logger.Debug("Useful debugging information.")
	// logger.Info("Something noteworthy happened!")
	// logger.Warn("You should probably take a look at this.")
	// logger.Error("Something failed but I'm not quitting.")
	err := detokenize.doValidations()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(detokenize.Records)
	var detokenizeRecord common.DetokenizeInput
	if err := json.Unmarshal(jsonRecord, &detokenizeRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.INVALID_RECORDS)
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

	logger.Info(messages.VALIDATE_DETOKENIZE_INPUT)

	var totalRecords = detokenizeApi.Records["records"]
	if totalRecords == nil {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.RECORDS_KEY_NOT_FOUND)
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.EMPTY_RECORDS)
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var token = singleRecord["token"]
		if token == nil {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.MISSING_TOKEN)
		} else if token == "" {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), messages.EMPTY_TOKEN_ID)
		}
	}
	return nil

}

func (detokenize *DetokenizeApi) sendRequest(records common.DetokenizeInput) (map[string]interface{}, *errors.SkyflowError) {

	var finalSuccess []map[string]interface{}
	var finalError []map[string]interface{}

	responseChannel := make(chan map[string]interface{})

	for i := 0; i < len(records.Records); i++ {
		logger.Info(fmt.Sprintf(messages.DETOKENIZING_RECORDS, records.Records[i].Token))
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

				if err != nil {
					logger.Error(fmt.Sprintf(messages.DETOKENIZING_FAILED, singleRecord.Token))
					var error = make(map[string]interface{})
					error["error"] = fmt.Sprintf(messages.SERVER_ERROR, err)
					error["token"] = singleRecord.Token
					responseChannel <- error
					//continue
					return
				}
				data, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()
				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				if err != nil {
					logger.Error(fmt.Sprintf(messages.DETOKENIZING_FAILED, singleRecord.Token))
					var error = make(map[string]interface{})
					error["error"] = fmt.Sprintf(messages.UNKNOWN_ERROR, string(data))
					error["token"] = singleRecord.Token
					responseChannel <- error
				} else {
					errorResult := result["error"]
					if errorResult != nil {
						logger.Error(fmt.Sprintf(messages.DETOKENIZING_FAILED, singleRecord.Token))
						var generatedError = (errorResult).(map[string]interface{})
						var error = make(map[string]interface{})
						error["error"] = generatedError["message"]
						error["token"] = singleRecord.Token
						responseChannel <- error
					} else {
						logger.Info(fmt.Sprintf(messages.DETOKENIZING_SUCCESS, singleRecord.Token))
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
