package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/skyflowapi/skyflow-go/api"
	SkyflowClient "github.com/skyflowapi/skyflow-go/client"
	"github.com/skyflowapi/skyflow-go/option"
	"github.com/skyflowapi/skyflow-go/recordservice"
)

/*
Example demonstrating how to use the Skyflow Go SDK to retrieve specific column values with redaction from the vault.
Steps:
1. Configure the skyflow client.
2. Configure columns and redaction rules.
3. Set pagination parameters (limit and offset).
4. Call the get API and handle the response.
*/

// getRecords retrieves records from a specified table in the vault.
func getRecords(client *recordservice.Client) {
    // Step 1: Set up context, vault ID, and table name
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Step 2: Configure columns and redactions
    // Define the columns to fetch
    columns := []string{"<COLUMN_1>", "<COLUMN_2>", "<COLUMN_3>"}

    // Define column redactions (optional)
    columnRedactions := []*api.ColumnRedactions{
        {
            ColumnName: stringPtr("<COLUMN_1>"),
            Redaction:  stringPtr("plain_text"),
        },
        {
            ColumnName: stringPtr("<COLUMN_2>"),
            Redaction:  stringPtr("redacted"),
        },
    }
    limit := 10 // Set the limit for the number of records to fetch
    offset := 2 // Set the offset 
	// Create the GetRequest object
	request := &api.GetRequest{
        VaultId:          &vaultID,
        TableName:        &tableName,
        Columns:          columns,
        ColumnRedactions: columnRedactions,
        Limit:           &limit,
        Offset:          &offset,
    }

    // Step 3: Execute get request with pagination
    response, err := client.Get(ctx, request)

    // Step 4: Handle and print the response
    if err != nil {
        fmt.Println("Error during get:", err)
        return
    }

    fmt.Println("Get response:", response)
}

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
    return &s
}

func main() {
    // Step 1: Configure the skyflow client.
    skyflowClient := SkyflowClient.NewClient(
        option.WithBaseURL("<BASE_URL>"), // base URL
		option.WithMaxAttempts(1),
        option.WithHTTPHeader(http.Header{
            "Authorization": []string{"Bearer <ACCESS_TOKEN>"},
        }),
    )
	var recordserviceClient *recordservice.Client = skyflowClient.Recordservice

	// Step 2: Call the getRecords function
	getRecords(recordserviceClient)
}
