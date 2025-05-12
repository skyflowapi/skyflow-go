package main

import (
    "context"
    "fmt"
    "github.com/skyflowapi/skyflow-go/flowservice"
    "github.com/skyflowapi/skyflow-go/api"
    "github.com/skyflowapi/skyflow-go/option"
    "net/http"
)

// detokenizeRecords detokenizes tokens to retrieve original data.
func detokenizeRecords(client *flowservice.Client) {
    ctx := context.Background()
    vaultID := "<VAULT_ID>"

    // Tokens to detokenize
    tokens := []string{"<TOKEN_1>", "<TOKEN_2>", "<TOKEN_3>"}

    // Optional token group redactions
    tokenGroupRedactions := []*api.V1TokenGroupRedactions{
        {
            TokenGroupName: stringPtr("<TOKEN_GROUP_1>"),
            Redaction:      stringPtr("<REDACTION_TYPE_1>"),
        },
    }

    // Create the detokenize request
    request := &api.V1FlowDetokenizeRequest{
        VaultId:             &vaultID,
        Tokens:              tokens,
        TokenGroupRedactions: tokenGroupRedactions,
    }

    // Call the Detokenize function
    response, err := client.Detokenize(ctx, request)
    if err != nil {
        fmt.Println("Error during detokenize:", err)
        return
    }

    fmt.Println("Detokenize response:", response)
}

func stringPtr(s string) *string {
    return &s
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
	detokenizeRecords(client)
}