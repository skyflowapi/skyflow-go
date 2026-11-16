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
 * This example generates scoped bearer tokens, limited to specific roles, from a credentials
 * file path and from a credentials string.
 */

var (
	filePath = "<YOUR_CREDENTIALS_FILE_PATH>"
	// Scoped tokens grant only the access of the roles listed here.
	options = common.BearerTokenOptions{RoleIds: []string{"<YOUR_ROLE_ID1>", "<YOUR_ROLE_ID2>"}}
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

func main() {
	fromFile, err := serviceaccount.GenerateBearerToken(filePath, options)
	if err != nil {
		fmt.Println("Error generating token from file path:", err.GetMessage())
	} else {
		fmt.Println("Scoped Token (using file path):", fromFile.AccessToken)
	}

	fromString, err := serviceaccount.GenerateBearerTokenFromCreds(credentialsString(), options)
	if err != nil {
		fmt.Println("Error generating token from credentials string:", err.GetMessage())
	} else {
		fmt.Println("Scoped Token (using credentials string):", fromString.AccessToken)
	}
}
