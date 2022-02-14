package main

import (
	"encoding/json"
	"fmt"

	Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func GetToken() (string, error) {
	return "<token>", nil
}

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error: ", r)
		}
	}()
	configuration := common.Configuration{VaultID: "<vault_id>", VaultURL: "<vault_url>", TokenProvider: GetToken}
	var client = Skyflow.Init(configuration)
	var records = make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["ids"] = []interface{}{"<id1>", "<id2>"}
	record1["table"] = "cards"
	record1["redaction"] = "PLAIN_TEXT"

	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	res, err := client.GetById(records)
	if err == nil {
		jsonRes, err := json.Marshal(res)
		if err == nil {
			fmt.Println("result: ", string(jsonRes))
		}
	} else {
		panic(err.GetMessage())
	}
}
