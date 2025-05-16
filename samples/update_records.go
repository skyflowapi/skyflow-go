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

// updateRecords updates existing records in a specified table in the vault.
func updateRecords(client *flowservice.Client) {
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Create the records to update
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

    // Create the update request
    request := &api.V1UpdateRequest{
        VaultId:   &vaultID,
        TableName: &tableName,
        Records:   records,
    }

    // Call the Update function
    response, err := client.Update(ctx, request)
    if err != nil {
        fmt.Println("Error during update:", err)
        return
    }

    fmt.Println("Update response:", response)
}

func main() {
	// Initialize the client
	skyflowClient := SkyflowClient.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // vault url
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer " + "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
	)
    var flowserviceClient *flowservice.Client = skyflowClient.Flowservice

	// Call the insertRecords function
	updateRecords(flowserviceClient)
}