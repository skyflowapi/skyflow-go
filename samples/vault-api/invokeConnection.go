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
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()

	configuration := common.Configuration{TokenProvider: GetToken}
	var client = Skyflow.Init(configuration)

	connectionUrl := "<CONNECTION_URL>"
	pathParams := make(map[string]string)
	pathParams["card_number"] = "<card_number>"

	//queryParams := make(map[string]interface{})
	//["cc"] = true

	requestBody := make(map[string]interface{})
	expiryDate := make(map[string]interface{})
	expiryDate["mm"] = "06"
	expiryDate["yy"] = "22"
	requestBody["expirationDate"] = expiryDate

	requestHeader := make(map[string]string)
	requestHeader["Authorization"] = "<Your-Authorization-Value>"

	var connectionConfig = common.ConnectionConfig{ConnectionURL: connectionUrl, MethodName: common.POST,
		PathParams: pathParams, RequestBody: requestBody, RequestHeader: requestHeader}

	res, err := client.InvokeConnection(connectionConfig)

	if err == nil {
		jsonRes, err := json.Marshal(res)
		if err == nil {
			fmt.Println("result: ", string(jsonRes))
		}
	} else {
		panic(err.GetMessage())
	}
}
