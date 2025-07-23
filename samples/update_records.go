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
Example demonstrating how to use the Skyflow Go SDK to update existing records in a Vault.
Steps:
1. Configure the skyflow client.
2. Get the flowservice client.
3. Specify records to update using Skyflow IDs.
4. Call the update API with the new values.
5. Handle and print the response.
*/

// updateRecords updates existing records in a specified table in the vault.
func updateRecords(client *flowservice.Client) {
    // Step 1: Set up context, vault ID, and table name
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Step 2: Create records with data to update
    records := []*api.V1UpdateRecordData{
        {
            SkyflowId: stringPtr("<SKYFLOW_ID>"),
            Data: map[string]interface{}{
                "<COLUMN_NAME_1>":       "<COLUMN_VALUE_1>",
                "<COLUMN_NAME_2>":       "<COLUMN_VALUE_2>",
                "<COLUMN_NAME_3>":       "<COLUMN_VALUE_3>",
            },
            Tokens: map[string]interface{}{
                "<COLUMN_NAME>": "<TOKEN>",
                // Add more columns and tokens as needed
            },
        },
    }

    // Step 3: Configure Create the update request & parameters
    request := &api.V1UpdateRequest{
        VaultId:   &vaultID,
        TableName: &tableName,
        Records:   records,
    }

    // Step 4: Call the Update API
    response, err := client.Update(ctx, request)
    if err != nil {
        fmt.Println("Error during update:", err)
        return
    }

    fmt.Println("Update response:", response)
}

func stringPtr(s string) *string {
	return &s
}

func main() {
    // Step 1: Configure the skyflow client.
	skyflowClient := SkyflowClient.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // vault url
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer " + "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
	)
    // Step 2: Get the flowservice client.
    var flowserviceClient *flowservice.Client = skyflowClient.Flowservice

    // Step 3: Call the updateRecords function
	updateRecords(flowserviceClient)
}