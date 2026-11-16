/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"encoding/json"
	"fmt"

	"github.com/skyflowapi/skyflow-go/flowvault/serviceaccount"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

/**
 * This example generates signed data tokens, with a string context and with an object context.
 * Each result pairs a data token with its signed token.
 */

var filePath = "<YOUR_CREDENTIALS_FILE_PATH>"

// credentialsString returns the credentials as a JSON string.
func credentialsString() string {
	credentials, _ := json.Marshal(map[string]string{
		"clientID":   "<YOUR_CLIENT_ID>",
		"clientName": "<YOUR_CLIENT_NAME>",
		"tokenURI":   "<YOUR_TOKEN_URI>",
		"keyID":      "<YOUR_KEY_ID>",
		"privateKey": "<YOUR_PRIVATE_KEY>",
	})
	return string(credentials)
}

func printResults(label string, results []common.SignedDataTokensResponse) {
	fmt.Println(label)
	for _, result := range results {
		fmt.Printf("  Token: %s, Signed Token: %s\n", result.Token, result.SignedToken)
	}
}

// Approach 1: a string context, from a credentials file path.
func signedTokensWithStringContext() {
	results, err := serviceaccount.GenerateSignedDataTokens(filePath, common.SignedDataTokensOptions{
		Ctx:        "user_12345",
		DataTokens: []string{"<DATA_TOKEN1>", "<DATA_TOKEN2>"},
		TimeToLive: 30, // seconds
	})
	if err != nil {
		fmt.Println("Error:", err.GetMessage())
		return
	}
	printResults("Signed tokens (string context):", results)
}

// Approach 2: an object context, from a credentials string. Each key maps to a Skyflow CEL
// policy variable under request.context.*, for example:
//
//	request.context.role == "analyst" && request.context.department == "research"
func signedTokensWithObjectContext() {
	results, err := serviceaccount.GenerateSignedDataTokensFromCreds(credentialsString(), common.SignedDataTokensOptions{
		Ctx: map[string]interface{}{
			"role":       "analyst",
			"department": "research",
			"user_id":    "user_67890",
		},
		DataTokens: []string{"<DATA_TOKEN1>", "<DATA_TOKEN2>"},
		TimeToLive: 30,
	})
	if err != nil {
		fmt.Println("Error:", err.GetMessage())
		return
	}
	printResults("Signed tokens (object context):", results)
}

func main() {
	signedTokensWithStringContext()
	signedTokensWithObjectContext()
}
