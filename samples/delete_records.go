package main

import (
    "context"
    "fmt"
    "github.com/skyflowapi/skyflow-go/flowservice"
    "github.com/skyflowapi/skyflow-go/api"
    "github.com/skyflowapi/skyflow-go/option"
    "github.com/skyflowapi/skyflow-go/client"
    "net/http"

)
/*
Example demonstrating how to use the Skyflow Go SDK to delete records from a Vault using Skyflow IDs.
Steps:
1. Configure the skyflow client.
2. Call the delete API with vault ID and Skyflow IDs.
3. Handle and print the response.
*/


// deleteRecords deletes records from a specified table in the vault.
func deleteRecords(client *flowservice.Client) {
    // Step 1: Set up the context, vault ID, and table name
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Step 2: Create the delete request with Skyflow IDs
    request := &api.V1DeleteRequest{
        VaultId:    &vaultID,
        TableName:  &tableName,
        SkyflowIDs: []string{"<SKYFLOW_ID_1>", "<SKYFLOW_ID_2>"},
    }

    // Step 3: Call the Delete API
    response, err := client.Delete(ctx, request)
    if err != nil {
        fmt.Println("Error during delete:", err)
        return
    }

    // Step 4: Handle the response
    fmt.Println("Delete response:", response)
}

func main() {
	// Step1: Configure the skyflow client.
	SkyflowClient := client.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // Vault URL
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer "+ "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
	)
    var flowserviceClient *flowservice.Client= SkyflowClient.Flowservice

	// Step 2: Call the deleteRecords function
	deleteRecords(flowserviceClient)
}