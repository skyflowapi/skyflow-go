package main

import (
	"encoding/json"
	"fmt"

	vaultapi "github.com/skyflowapi/skyflow-go/vault-api"
)

func main2() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error: ", r)
		}
	}()
	configuration := vaultapi.Configuration{"b359c43f1b913", "https://sb1.area51.vault.skyflowapis.tech", getToken, vaultapi.Options{vaultapi.WARN}}
	var record = `{"records":[{"ids":["e1a84d29-a2c3-41a3-96bf-038feef5175b", "81fb2a6b-d2e2-4772-905f-b185b1ae0c9b"],"redaction":"PLAIN_TEXT","table":"cards"}]}`
	var client = vaultapi.Init(configuration)
	var records map[string]interface{}
	json.Unmarshal([]byte(record), &records)
	res, err := client.Detokenize(records)
	if err == nil {
		jsonRes, err := json.Marshal(res)
		if err == nil {
			fmt.Println("result: ", string(jsonRes))
		}
	} else {
		panic(err.Error())
	}
}
