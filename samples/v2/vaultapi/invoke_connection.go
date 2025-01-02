package vaultapi

import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"

	. "github.com/skyflowapi/skyflow-go/v2/client"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

func main() {
	// Add connection configurations 1
	connConfig1 := ConnectionConfig{ConnectionId: "<CONNECTION_ID1>", ConnectionUrl: "<CONNECTION_URL1>", Credentials: Credentials{Token: "<BEARER_TOKEN1>"}}

	// Add connection configurations 2
	connConfig2 := ConnectionConfig{ConnectionId: "<CONNECTION_ID2>", ConnectionUrl: "<CONNECTION_URL2>", Credentials: Credentials{Token: "<BEARER_TOKEN2>"}}

	var arr []ConnectionConfig
	arr = append(arr, connConfig1, connConfig2)

	// Initialize Skyflow client
	client1, clientError := NewSkyflow(
		WithConnections(arr...),
		WithLogLevel(logger.DEBUG),
	)
	if clientError != nil {
		fmt.Println("Error:", clientError)
	} else {
		service, conError := client1.Connection("<CONNECTION_ID1>")
		if conError != nil {
			fmt.Println("Error:", conError)
		} else {
			ctx := context.TODO()
			body := map[string]interface{}{ // Set your data
				"<HEADER_NAME_1>": "<HEADER_VALUE_1>",
				"<HEADER_NAME_2>": "<HEADER_VALUE_2>",
			}
			headers := map[string]string{
				"Content-Type": "application/json", // Set the content type
			}
			queryParams := map[string]interface{}{ // Set your params
				"<YOUR_QUERY_PARAM_KEY_1>": "<YOUR_QUERY_PARAM_VALUE_1>",
				"<YOUR_QUERY_PARAM_KEY_2>": "<YOUR_QUERY_PARAM_VALUE_2>",
			}
			pathParams := map[string]string{"<YOUR_PATH_PARAM_KEY_1>": "<YOUR_PATH_PARAM_VALUE_1>"}
			req := InvokeConnectionRequest{
				Method:      POST, // set the request method
				Headers:     headers,
				Body:        body,
				QueryParams: queryParams,
				PathParams:  pathParams,
			}
			res, invokeError := service.Invoke(ctx, req)
			if invokeError != nil {
				fmt.Println("ERROR: ", *invokeError)
			} else {
				fmt.Println("RESPONSE", res)
			}
		}
	}

}
