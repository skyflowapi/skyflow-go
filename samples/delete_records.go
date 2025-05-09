package main

import (
    "context"
    "fmt"
    "github.com/skyflowapi/skyflow-go/flowservice"
    "github.com/skyflowapi/skyflow-go/api"
)

// deleteRecords deletes records from a specified table in the vault.
func deleteRecords(client *flowservice.Client) {
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Create the delete request
    request := &api.V1DeleteRequest{
        VaultId:    &vaultID,
        TableName:  &tableName,
        SkyflowIDs: []string{"<SKYFLOW_ID_1>", "<SKYFLOW_ID_2>"},
    }

    // Call the Delete function
    response, err := client.Delete(ctx, request)
    if err != nil {
        fmt.Println("Error during delete:", err)
        return
    }

    fmt.Println("Delete response:", response)
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
	deleteRecords(client)
}