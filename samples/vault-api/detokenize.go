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
	res, err := client.Detokenize(records)
	if err == nil {
		fmt.Println("Records : ", res.Records)
	} else {
		panic(err.GetMessage())
	}
}
