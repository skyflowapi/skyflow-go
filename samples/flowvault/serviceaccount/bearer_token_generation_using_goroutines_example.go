/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/skyflowapi/skyflow-go/flowvault/serviceaccount"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

/**
 * This example generates bearer tokens concurrently from two goroutines, one using a credentials
 * file path and one using a credentials string, with a shared context. Each goroutine generates
 * and prints a token five times.
 */

var (
	filePath = "<YOUR_CREDENTIALS_FILE_PATH>"
	options  = common.BearerTokenOptions{Ctx: "abc"}
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

func generateFromFilePath() {
	for i := 0; i < 5; i++ {
		response, err := serviceaccount.GenerateBearerToken(filePath, options)
		if err != nil {
			fmt.Println("Error generating token from file path:", err.GetMessage())
			continue
		}
		fmt.Println(response.AccessToken)
	}
}

func generateFromCredentialsString() {
	credentials := credentialsString()
	for i := 0; i < 5; i++ {
		response, err := serviceaccount.GenerateBearerTokenFromCreds(credentials, options)
		if err != nil {
			fmt.Println("Error generating token from credentials string:", err.GetMessage())
			continue
		}
		fmt.Println(response.AccessToken)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		generateFromFilePath()
	}()
	go func() {
		defer wg.Done()
		generateFromCredentialsString()
	}()
	wg.Wait()
}
