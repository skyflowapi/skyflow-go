package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/skyflowapi/skyflow-go/errors"
)

type insertApi struct {
	configuration Configuration
	records       InsertRecord
	options       InsertOptions
	token         string
}

func (insertApi *insertApi) post() (map[string]interface{}, *errors.SkyflowError) {
	record, err := insertApi.constructRequestBody(insertApi.records, insertApi.options)
	if err != nil {
		return nil, err
	}
	requestBody, err1 := json.Marshal(record)
	if err1 != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(errors.UNKNOWN_ERROR, err1))
	}
	requestUrl := fmt.Sprintf("%s/v1/vaults/%s", insertApi.configuration.VaultURL, insertApi.configuration.VaultID)
	request, _ := http.NewRequest(
		"POST",
		requestUrl,
		strings.NewReader(string(requestBody)),
	)
	bearerToken := fmt.Sprintf("Bearer %s", insertApi.token)
	request.Header.Add("Authorization", bearerToken)
	res, err2 := http.DefaultClient.Do(request)
	if err2 != nil {
		code := strconv.Itoa(res.StatusCode)
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(code), fmt.Sprintf(errors.SERVER_ERROR, err2))
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var result map[string]interface{}
	err2 = json.Unmarshal(data, &result)
	if err2 != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(errors.UNKNOWN_ERROR, string(data)))
	} else if result["error"] != nil {
		var generatedError = (result["error"]).(map[string]interface{})
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(fmt.Sprintf("%v", generatedError["http_code"])), fmt.Sprintf(errors.SERVER_ERROR, generatedError["message"]))
	}
	return insertApi.buildResponse((result["responses"]).([]interface{})), nil
}

func (InsertApi *insertApi) constructRequestBody(record InsertRecord, options InsertOptions) (map[string]interface{}, *errors.SkyflowError) {
	postPayload := []interface{}{}
	records := record.Records

	for index, value := range records {
		singleRecord := value
		table := singleRecord.Table
		fields := singleRecord.Fields
		var finalRecord = make(map[string]interface{})
		finalRecord["tableName"] = table
		finalRecord["fields"] = fields
		finalRecord["method"] = "POST"
		finalRecord["quorum"] = true
		postPayload = append(postPayload, finalRecord)
		if options.Tokens {
			temp2 := make(map[string]interface{})
			temp2["method"] = "GET"
			temp2["tableName"] = table
			temp2["ID"] = fmt.Sprintf("$responses.%v.records.0.skyflow_id", index)
			temp2["tokenization"] = true
			postPayload = append(postPayload, temp2)
		}

	}
	body := make(map[string]interface{})
	body["records"] = postPayload
	return body, nil
}

func (insertApi *insertApi) buildResponse(responseJson []interface{}) map[string]interface{} {

	var inputRecords = insertApi.records.Records
	var recordsArray = []interface{}{}
	var responseObject = make(map[string]interface{})
	if insertApi.options.Tokens {
		for i := (len(responseJson) / 2); i < len(responseJson); i++ {
			var skyflowIDsObject = (responseJson[i-
				(len(responseJson)-len(responseJson)/2)]).(map[string]interface{})
			var skyflowIDs = (skyflowIDsObject["records"]).([]interface{})
			var skyflowID = (skyflowIDs[0]).(map[string]interface{})["skyflow_id"]
			var record = (responseJson[i]).(map[string]interface{})
			var inputRecord = inputRecords[i-len(responseJson)/2]
			record["table"] = inputRecord.Table
			var fields = (record["fields"]).(map[string]interface{})
			fields["skyflow_id"] = skyflowID
			record["fields"] = fields
			recordsArray = append(recordsArray, record)
		}
	} else {
		for i := 0; i < len(responseJson); i++ {
			var inputRecord = inputRecords[i]
			var record = ((responseJson[i]).(map[string]interface{})["records"]).([]interface{})
			((record[0]).(map[string]interface{}))["table"] = inputRecord.Table
			recordsArray = append(recordsArray, record[0])

		}
	}
	responseObject["records"] = recordsArray
	return responseObject
}
