package main

import (
	"encoding/json"
	"fmt"

	Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()

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
		result, jsonErr := json.Marshal(res)
		if jsonErr == nil {
			fmt.Println("result", string(result))
		} else {
			fmt.Println("unable to parse :", result)
		}
	} else {
		panic(err.GetMessage())
	}
}
