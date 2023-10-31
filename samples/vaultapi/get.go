/*
Copyright (c) 2022 Skyflow, Inc.
*/
package main

import (
	"fmt"

	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	saUtil "github.com/skyflowapi/skyflow-go/serviceaccount/util"
	Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

var bearerToken = ""

func GetToken() (string, error) {

	filePath := "<file_path>"
	if saUtil.IsExpired(bearerToken) {
		newToken, err := saUtil.GenerateBearerToken(filePath)
		if err != nil {
			return "", err
		} else {
			bearerToken = newToken.AccessToken
			return bearerToken, nil
		}
	}
	return bearerToken, nil
}

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error: ", r)
		}
	}()
	logger.SetLogLevel(logger.INFO) //set loglevel to INFO
	configuration := common.Configuration{VaultID: "<vault_id>", VaultURL: "<vault_url>", TokenProvider: GetToken}
	var client = Skyflow.Init(configuration)
	var records = make(map[string]interface{})
	var record1 = make(map[string]interface{})
	var record2 = make(map[string]interface{})
	var record3 = make(map[string]interface{})

	record1["ids"] = []interface{}{"<id1>", "<id2>"}
	record1["table"] = "<table_name>"
	record1["redaction"] = common.PLAIN_TEXT

	record2["columnValues"] = []interface{}{"<column_value1>", "<column_value2>"}
	record2["columnName"] = "<column_name>"
	record2["table"] = "<table_name>"
	record2["redaction"] = common.PLAIN_TEXT
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	recordsArray = append(recordsArray, record2)

	records["records"] = recordsArray
	res, err := client.Get(records)
	if err == nil {
		fmt.Println("Records : ", res)
	} else {
		fmt.Println("result is", err)
		panic(err.GetMessage())
	}
}
