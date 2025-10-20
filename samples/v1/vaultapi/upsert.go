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
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()
	logger.SetLogLevel(logger.INFO) // Set loglevel to INFO
	configuration := common.Configuration{VaultID: "<vault_id>", VaultURL: "<vault_url>", TokenProvider: GetToken}
	var client = Skyflow.Init(configuration)

	var records = make(map[string]interface{})
	var record = make(map[string]interface{})
	record["table"] = "cards"
	var fields = make(map[string]interface{})
	fields["cardNumber"] = "4111111111111111"
	fields["fullname"] = "name"
	record["fields"] = fields
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray

	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table:"cards",Column:"cardNumber"}
	upsertArray = append(upsertArray,upsertOption)
	
    var options = common.InsertOptions {
        Tokens: true
        Upsert: upsertArray
    }

	res, err := client.Insert(records, options)
	if err == nil {
		fmt.Println("Records : ", res.Records)
	} else {
		panic(err.GetMessage())
	}
}
