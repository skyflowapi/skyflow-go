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
Example demonstrating how to use the Skyflow Go SDK to detokenize records from a Vault.
Steps:
1. Configure the skyflow client.
2. Call the detokenize API.
3. Handle and print the response.
*/

// detokenizeRecords detokenizes tokens to retrieve original data.
func detokenizeRecords(client *recordservice.Client) {
	// Step 1: Set up context and vault ID
	ctx := context.Background()
	vaultID := "<VAULT_ID>"

	// Step 2: Configure tokens to detokenize and redactions
	tokens := []string{"<TOKEN_1>", "<TOKEN_2>", "<TOKEN_3>"}

	// Optional token group redactions
	tokenGroupRedactions := []*api.TokenGroupRedactions{
		{
			TokenGroupName: stringPtr("<TOKEN_GROUP_1>"),
			Redaction:      stringPtr("<REDACTION_TYPE_1>"),
		},
	}

	// Create the detokenize request
	request := &api.DetokenizeRequest{
		VaultId:              &vaultID,
		Tokens:               tokens,
		TokenGroupRedactions: tokenGroupRedactions,
	}

	// Step 3: Call the Detokenize function
	response, err := client.Detokenize(ctx, request)

	// Step 4: Handle and print the response.
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
	// Step 1: Initialize the skyflow client
	skyflowClient := SkyflowClient.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // vault url
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer " + "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
	)
	var recordserviceClient *recordservice.Client = skyflowClient.Recordservice

	// Step 2: Call the detokenize API and handle response
	detokenizeRecords(recordserviceClient)
}
