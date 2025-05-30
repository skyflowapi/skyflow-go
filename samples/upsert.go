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

// insertRecords inserts new records into a specified table in the vault.
func upsertRecords(client *flowservice.Client) {
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Create the records to insert
    records := []*api.V1InsertRecordData{
        {
            Data: map[string]interface{}{
                "<COLUMN_NAME_1>":       "<COLUMN_VALUE_1>",
                "<COLUMN_NAME_2>":       "<COLUMN_VALUE_2>",
                "<COLUMN_NAME_3>":       "<COLUMN_VALUE_3>",
            },
        },
    }
	upsert := api.V1Upsert{
		UpdateType: api.EnumUpdateTypeUpdate.Ptr(),
		UniqueColumns: []string{
			"<COLUMN_NAME_1>",
			"<COLUMN_NAME_2>",
		},
	}

    // Create the insert request
    request := &api.V1InsertRequest{
        VaultId:   &vaultID,
        TableName: &tableName,
        Records:   records,
		Upsert:    &upsert,
    }

    // Call the Insert function
    response, err := client.Insert(ctx, request)
    if err != nil {
        fmt.Println("Error during insert:", err)
        return
    }

    fmt.Println("Insert response:", response)
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
	upsertRecords(flowserviceClient)
}