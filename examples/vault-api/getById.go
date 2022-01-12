package main

import (
	"encoding/json"
	"fmt"

	Skyflow "github.com/skyflowapi/skyflow-go/vault-api"
)

func GetToken() string {
	return "<token>"
}
func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error: ", r)
		}
	}()
	configuration := Skyflow.Configuration{VaultID: "<vauld_id>", VaultURL: "<vault_url>", TokenProvider: GetToken, Options: Skyflow.Options{LogLevel: Skyflow.WARN}}
	var client = Skyflow.Init(configuration)
	var records = make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["ids"] = []interface{}{"<skyflow_id>", "<skyflow_id>"}
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
