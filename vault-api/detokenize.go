package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/skyflowapi/skyflow-go/errors"
)

type detokenizeApi struct {
	configuration Configuration
	records       DetokenizeInput
	token         string
}

func (detokenize *detokenizeApi) get() (map[string]interface{}, *errors.SkyflowError) {
	err := detokenize.validateRecords(detokenize.records)
	if err != nil {
		return nil, err
	}
	res, err := detokenize.sendRequest()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (detokenize *detokenizeApi) validateRecords(records DetokenizeInput) *errors.SkyflowError {
	fmt.Println(records)
	if len(records.Records) == 0 {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.EMPTY_RECORDS)

	}
	for i := 0; i < len(records.Records); i++ {
		singleRecord := records.Records[0]
		if singleRecord.Token == "" {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.EMPTY_TOKEN_ID)
		}
	}
	return nil
}

func (detokenize *detokenizeApi) sendRequest() (map[string]interface{}, *errors.SkyflowError) {

	var wg = sync.WaitGroup{}
	var finalSuccess []map[string]interface{}
	var finalError []map[string]interface{}
	for i := 0; i < len(detokenize.records.Records); i++ {
		wg.Add(1)
		singleRecord := detokenize.records.Records[i]
		requestUrl := fmt.Sprintf("%s/v1/vaults/%s/detokenize", detokenize.configuration.VaultURL, detokenize.configuration.VaultID)
		var detokenizeParameter = []RevealRecord{singleRecord}
		var body = make(map[string]interface{})
		body["detokenizationParameters"] = detokenizeParameter
		requestBody, err := json.Marshal(body)
		if err == nil {
			request, _ := http.NewRequest(
				"POST",
				requestUrl,
				strings.NewReader(string(requestBody)),
			)
			bearerToken := fmt.Sprintf("Bearer %s", detokenize.token)
			request.Header.Add("Authorization", bearerToken)
			res, err := http.DefaultClient.Do(request)
			if err != nil {
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
				error["error"] = fmt.Sprintf(errors.UNKNOWN_ERROR, err)
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
	finalRecord["success"] = finalSuccess
	finalRecord["errors"] = finalError
	return finalRecord, nil
}
