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
 * This example generates a bearer token in two ways, and reuses it until it expires:
 * 1. From the path to a credentials.json file.
 * 2. From the credentials as a JSON string.
 */

var (
	filePath    = "<YOUR_CREDENTIALS_FILE_PATH>"
	bearerToken = ""
)

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

func bearerTokenFromFilePath() string {
	if bearerToken != "" && !serviceaccount.IsExpired(bearerToken) {
		return bearerToken
	}
	response, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{})
	if err != nil {
		fmt.Println("Error generating token from file path:", err.GetMessage())
		return ""
	}
	bearerToken = response.AccessToken
	return bearerToken
}

func bearerTokenFromCredentialsString() string {
	if bearerToken != "" && !serviceaccount.IsExpired(bearerToken) {
		return bearerToken
	}
	response, err := serviceaccount.GenerateBearerTokenFromCreds(credentialsString(), common.BearerTokenOptions{})
	if err != nil {
		fmt.Println("Error generating token from credentials string:", err.GetMessage())
		return ""
	}
	bearerToken = response.AccessToken
	return bearerToken
}

func main() {
	fmt.Println("Generated Bearer Token (from file):", bearerTokenFromFilePath())
	fmt.Println("Generated Bearer Token (from string):", bearerTokenFromCredentialsString())
}
