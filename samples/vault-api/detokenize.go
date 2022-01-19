package main

import (
	"encoding/json"
	"fmt"

	Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error: ", r)
		}
	}()

	configuration := common.Configuration{VaultID: "<vault_id>", VaultURL: "<vault_url>", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
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
		jsonRes, err := json.Marshal(res)
		if err == nil {
			fmt.Println("result: ", string(jsonRes))
		}
	} else {
		panic(err.GetMessage())
	}
}
