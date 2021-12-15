package main

import (
	"encoding/json"
	"fmt"

	vaultapi "github.com/skyflowapi/skyflow-go/vault-api"
)

func getToken() string {
	return "token"
}
func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error: ", r)
		}
	}()

	configuration := vaultapi.Configuration{"b359c43f1b84", "https://sb1.area51.vault.skyflowapis.tech", getToken, vaultapi.Options{vaultapi.WARN}}
	var record = `{"records":[{"token":"342d993c-e83c-4f3c-a98c-de49309af382"},{"token":"e6b6d252-9f99-4c21-b08a-5a62533725f7"},{"token":"xxxx"}]}`
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
