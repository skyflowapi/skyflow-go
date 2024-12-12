package vaultapi

import (
	"context"
	"fmt"
	. "skyflow-go/v2/client"
	. "skyflow-go/v2/utils/common"
	"skyflow-go/v2/utils/logger"
)

func main() {

	// Connection 1 configuration
	connConfig1 := ConnectionConfig{ConnectionId: "<CONNECTION_ID_1>", ConnectionUrl: "<CONNECTION_ID_URL>", Credentials: Credentials{Token: "<TOKEN>"}}

	// Connection 2 configuration
	connConfig2 := ConnectionConfig{
		ConnectionId:  "<CONNECTION_ID_2>",
		ConnectionUrl: "<CONNECTION_URL_2>",
		Credentials: Credentials{
			Token: "<TOKEN>",
		},
	}
	skyflow1 := Skyflow{}
	client1, _ := skyflow1.Builder().WithConnectionConfig(connConfig1).WithConnectionConfig(connConfig2).WithLogLevel(logger.DEBUG).Build()
	service, _ := client1.Connection("<CONNECTION_ID>")
	ctx := context.TODO()
	res, getErr := service.Invoke(ctx, InvokeConnectionRequest{
		Method: POST, // set the request method
		Headers: map[string]string{
			"Content-Type": "application/json", // Set the content type
		},
		Body: map[string]interface{}{ // Set your data
			"<KEY_1>": "<VALUE_1>",
			"<KEY_2>": "<VALUE_2>",
		},
	})
	if getErr != nil {
		fmt.Println("ERROR: ", *getErr)
	} else {
		fmt.Println("RESPONSE", res)
	}

}
