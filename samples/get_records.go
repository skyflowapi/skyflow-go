package main

import (
    "context"
    "fmt"
    "github.com/skyflowapi/skyflow-go/flowservice"
    "github.com/skyflowapi/skyflow-go/api"
    "github.com/skyflowapi/skyflow-go/option"
    "net/http"
)

// getRecords retrieves records from a specified table in the vault.
func getRecords(client *flowservice.Client) {
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Create the get request
    request := &api.V1GetRequest{
        VaultId:    &vaultID,
        TableName:  &tableName,
        SkyflowIDs: []string{"<SKYFLOW_ID_1>", "<SKYFLOW_ID_2>"},
    }

    // Call the Get function
    response, err := client.Get(ctx, request)
    if err != nil {
        fmt.Println("Error during get:", err)
        return
    }

    fmt.Println("Get response:", response)
}
func main() {
	// Initialize the client
	client := flowservice.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // vault url
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer "+ "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
		
	)

	// Call the deleteRecords function
	getRecords(client)
}