package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	}
	requestBody, err := json.Marshal(record)
	if err != nil {
	}
	requestUrl := fmt.Sprintf("%s/v1/vaults/%s", insertApi.configuration.VaultURL, insertApi.configuration.VaultID)
	request, _ := http.NewRequest(
		"POST",
		requestUrl,
		strings.NewReader(string(requestBody)),
	)
	bearerToken := fmt.Sprintf("Bearer %s", insertApi.token)
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
	}
	return insertApi.buildResponse((result["responses"]).([]interface{})), nil
}

func (InsertApi *insertApi) constructRequestBody(record InsertRecord, options InsertOptions) (map[string]interface{}, error) {
	postPayload := []interface{}{}
	records := record.Records

	if len(records) == 0 {
	}
	for index, value := range records {
		singleRecord := value
		table := singleRecord.Table

		if table == "" {
		}
		fields := singleRecord.Fields

		if len(fields) == 0 {
		}

		for column, _ := range fields {
			if column == "" {
			}
		}

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

//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
