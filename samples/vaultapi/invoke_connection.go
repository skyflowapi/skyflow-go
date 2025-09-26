/*
Copyright (c) 2022 Skyflow, Inc.
*/
package main

import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"

	. "github.com/skyflowapi/skyflow-go/v2/client"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to use the Skyflow Go SDK to invoke API connections
 * <p>
 * Steps include:
 * 1. Setting up Connection credentials.
 * 2. Configure the Skyflow client.
 * 3. Configure Skyflow client with connection ID.
 * 4. Invoking connections with proper parameters and handling responses.
 * 5. Handle the response and errors.
 */

func main() {
	// Step 1: Setting up Connection credentials
	// Add connection configurations 1
	connConfig1 := ConnectionConfig{ConnectionId: "<CONNECTION_ID1>", ConnectionUrl: "<CONNECTION_URL1>", Credentials: Credentials{CredentialsString: "<STRINGIFIED_JSON_VALUE>"}}

	// Add connection configurations 2
	connConfig2 := ConnectionConfig{ConnectionId: "<CONNECTION_ID2>", ConnectionUrl: "<CONNECTION_URL2>", Credentials: Credentials{CredentialsString: "<STRINGIFIED_JSON_VALUE>"}}

	var arr []ConnectionConfig
	arr = append(arr, connConfig1, connConfig2)

	// Step 2: Configure the Skyflow client
	// Initialize Skyflow client
	skyflowClient, clientError := NewSkyflow(
		WithConnections(arr...),
		WithLogLevel(logger.DEBUG),
	)
	if clientError != nil {
		fmt.Println("Error:", clientError)
	} else {
		// Step 3: Configure Skyflow client with connection ID
		service, conError := skyflowClient.Connection("<CONNECTION_ID1>")
		if conError != nil {
			fmt.Println("Error:", conError)
		} else {
			ctx := context.TODO()
			body := map[string]interface{}{ // Set your data
				"<KEY_1>": "<VALUE_1>",
				"<KEY_2>": "<VALUE_2>",
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
			// Step 4: Invoke connections with proper parameters and handle responses
			res, invokeError := service.Invoke(ctx, req)

			// Step 5: Handle the response and errors
			if invokeError != nil {
				fmt.Println("ERROR: ", *invokeError)
			} else {
				fmt.Println("RESPONSE", res)
			}
		}
	}

}
