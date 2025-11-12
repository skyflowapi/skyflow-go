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
	record1["token"] = "<token>"
	var record2 = make(map[string]interface{})
	record2["token"] = "<token>"

	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	recordsArray = append(recordsArray, record2)
	records["records"] = recordsArray
	//default value for ContinueOnError is true
	var options = common.DetokenizeOptions{ ContinueOnError: false };
	res, err := client.Detokenize(records, options)
	if err == nil {
		fmt.Println("Records : ", res.Records)
	} else {
		panic(err.GetMessage())
	}
}
