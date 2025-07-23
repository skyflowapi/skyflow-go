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
Example demonstrating how to use the Skyflow Go SDK to insert records into a Vault.
Steps:
1. Configure the skyflow client.
2. Prepare records data for insertion.
3. Call the insert API with the records.
4. Handle and print the response.
*/

// insertRecords inserts new records into a specified table in the vault.
func insertRecords(client *flowservice.Client) {
    // Step 1: Set up the context, vault ID, and table name
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Step 2: Create records with data
    records := []*api.V1InsertRecordData{
        {
            Data: map[string]interface{}{
                "<COLUMN_NAME_1>":       "<COLUMN_VALUE_1>",
                "<COLUMN_NAME_2>":       "<COLUMN_VALUE_2>",
                "<COLUMN_NAME_3>":       "<COLUMN_VALUE_3>",
            },
        },
    }

    // Step 3: Create and execute the insert request
    request := &api.V1InsertRequest{
        VaultId:   &vaultID,
        TableName: &tableName,
        Records:   records,
    }

    // Call the Insert function
    response, err := client.Insert(ctx, request)

    // Step 4: Handle and print the response.
    if err != nil {
        fmt.Println("Error during insert:", err)
        return
    }

    fmt.Println("Insert response:", response)
}

func main() {
	// Step 1: Configure the skyflow client.
	skyflowClient := SkyflowClient.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // Vault URL
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer " + "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
	)
    var flowserviceClient *flowservice.Client = skyflowClient.Flowservice

    // Step 2: Call the insertRecords function
	insertRecords(flowserviceClient)
}