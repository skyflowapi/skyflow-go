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

// deleteTokens deletes tokens from the vault.
func deleteTokens(client *flowservice.Client) {
    vaultID := "<VAULT_ID>"

    // Tokens to delete
    tokens := []string{"<TOKEN_1>", "<TOKEN_2>"}

    // Create the delete token request
    request := &api.V1FlowDeleteTokenRequest{
        VaultId: &vaultID,
        Tokens:  tokens,
    }

    // Call the DeleteToken function
    ctx := context.Background()
    response, err := client.Deletetoken(ctx, request)
    if err != nil {
        fmt.Println("Error during token deletion:", err)
        return
    }

    fmt.Println("DeleteToken response:", response)
}
func main() {
	// Initialize the client
	client := SkyflowClient.NewClient(
		option.WithBaseURL("<VAULT_URL>"), // vault url
		option.WithHTTPHeader(http.Header{
			"Authorization": []string{"Bearer "+ "<BEARER_TOKEN>"}, // Bearer token
		}),
		option.WithMaxAttempts(1),
		
	)
    var flowserviceClient *flowservice.Client= client.Flowservice

	// Call the delete tokens function
	deleteTokens(flowserviceClient)
}