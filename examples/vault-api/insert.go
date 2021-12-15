package main

import (
	"encoding/json"
	"fmt"

	vaultapi "github.com/skyflowapi/skyflow-go/vault-api"
)

func main1() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()

	configuration := vaultapi.Configuration{"b359c43f1b844ff4bea0f03", "https://sb1.area51.vault.skyflowapis.tech", getToken, vaultapi.Options{vaultapi.WARN}}
	var record = `{"records":[{"table":"cards","fields":{"cvv":"123"}}]}`
	var client = vaultapi.Init(configuration)
	var options = vaultapi.InsertOptions{true}
	var records map[string]interface{}
	json.Unmarshal([]byte(record), &records)
	res, err := client.Insert(records, options)
	if err == nil {
		result, jsonErr := json.Marshal(res)
		if jsonErr == nil {
			fmt.Println(result)
		} else {
			fmt.Println("unable to parse :", result)
		}
	} else {
		panic(err)
	}
}
