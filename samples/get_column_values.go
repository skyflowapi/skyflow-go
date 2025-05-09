package main

import (
    "context"
    "fmt"
    "net/http"
    "github.com/skyflowapi/skyflow-go/flowservice"
    "github.com/skyflowapi/skyflow-go/api"
    "github.com/skyflowapi/skyflow-go/option"
)

// getRecords retrieves records from a specified table in the vault.
func getRecords(client *flowservice.Client) {
    ctx := context.Background()
    vaultID := "<VAULT_ID>"
    tableName := "<TABLE_NAME>"

    // Define the columns to fetch
    columns := []string{"<COLUMN_1>", "<COLUMN_2>", "<COLUMN_3>"}

    // Define column redactions (optional)
    columnRedactions := []*api.V1ColumnRedactions{
        {
            ColumnName: stringPtr("<COLUMN_1>"),
            Redaction:  stringPtr("plain_text"),
        },
        {
            ColumnName: stringPtr("<COLUMN_2>"),
            Redaction:  stringPtr("redacted"),
        },
    }

    // Create the V1GetRequest object
    request := &api.V1GetRequest{
        VaultId:          &vaultID,
        TableName:        &tableName,
        Columns:          columns,
        ColumnRedactions: columnRedactions,
    }

    // Call the Get function
    response, err := client.Get(ctx, request)
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
    // Initialize the client
    client := flowservice.NewClient(
        option.WithBaseURL("<BASE_URL>"), // base URL
		option.WithMaxAttempts(1),
        option.WithHTTPHeader(http.Header{
            "Authorization": []string{"Bearer <ACCESS_TOKEN>"},
        }),
    )

    // Call the getRecords function
    getRecords(client)
}