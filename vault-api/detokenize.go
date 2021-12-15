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
		return nil
	}
	for i := 0; i < len(records.Records); i++ {
		singleRecord := records.Records[0]
		if singleRecord.Token == "" {
			return nil
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
		var r = make(map[string]interface{})
		r["detokenizationParameters"] = detokenizeParameter
		requestBody, err := json.Marshal(r)
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
				fmt.Println("error from server: ", err)
			}
			data, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				fmt.Println(err)
				//return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.INVALID_FIELD)
			} else {
				errorResult := result["error"]
				if errorResult != nil {
					var generatedError = (errorResult).(map[string]interface{})
					fmt.Println(generatedError)
					var error = make(map[string]interface{})
					//var skyflowError = errors.NewSkyflowError("404", (generatedError["message"]).(string))
					error["error"] = "skyflowError"
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
