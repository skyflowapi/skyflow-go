package main

import (
    "context"
    "fmt"
    "github.com/skyflowapi/skyflow-go/flowservice"
    "github.com/skyflowapi/skyflow-go/api"
    "github.com/skyflowapi/skyflow-go/option"
    SkyflowClient "github.com/skyflowapi/skyflow-go/client"
    "net/http"
)

/*
Example demonstrating how to use the Skyflow Go SDK to retrieve records from a Vault using Skyflow IDs.
Steps:
1. Configure the skyflow client.
2. Get the flowservice client.
3. Call the get API with vault ID and Skyflow IDs.
*/

// getRecords retrieves records from a specified table in the vault.
func getRecords(client *flowservice.Client) {
    // Step 1: Set up the context, vault ID, and table name
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Step 2: Create get request with Skyflow IDs
    request := &api.V1GetRequest{
        VaultId:    &vaultID,
        TableName:  &tableName,
        SkyflowIDs: []string{"<SKYFLOW_ID_1>", "<SKYFLOW_ID_2>"},
    }

    // Step 3: Execute get request and handle response
    response, err := client.Get(ctx, request)
    if err != nil {
        fmt.Println("Error during get:", err)
        return
    }

    fmt.Println("Get response:", response)
}
func main() {
	// Step 1: Configure the skyflow client.
	skyflowClient := SkyflowClient.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // Vault URL
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer "+ "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
		
	)
    // Step 2: Get the flowservice client
    var flowserviceClient *flowservice.Client = skyflowClient.Flowservice

    // Step 3: Call the get API with vault ID and Skyflow IDs.
	getRecords(flowserviceClient)
}