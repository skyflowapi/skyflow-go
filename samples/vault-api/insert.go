package main

import (
	"fmt"

	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	saUtil "github.com/skyflowapi/skyflow-go/service-account/util"
	Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func GetToken() (string, error) {
	filePath := "<file_path>"
	token, err := saUtil.GenerateBearerToken(filePath)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()
	logger.SetLogLevel(logger.INFO) //set loglevel to INFO
	configuration := common.Configuration{VaultID: "<vault_id>", VaultURL: "<vault_url>", TokenProvider: GetToken}
	var client = Skyflow.Init(configuration)
	var options = common.InsertOptions{Tokens: false}
	var records = make(map[string]interface{})
	var record = make(map[string]interface{})
	record["table"] = "cards"
	var fields = make(map[string]interface{})
	fields["cvv"] = "123"
	fields["fullname"] = "name"
	record["fields"] = fields
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	res, err := client.Insert(records, options)
	if err == nil {
		fmt.Println("Records : ", res.Records)
	} else {
		panic(err.GetMessage())
	}
}
