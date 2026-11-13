/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"fmt"

	"github.com/skyflowapi/skyflow-go/flowvault/serviceaccount"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

/**
 * This example generates bearer tokens that carry a context, for context-aware access policies.
 * 1. A string context, referenced in a Skyflow policy as request.context.
 * 2. An object context, whose keys are referenced as request.context.<key>.
 */

var filePath = "<YOUR_CREDENTIALS_FILE_PATH>"

// Approach 1: a string context.
// Use a simple string identifier when your policy references a single context value.
func bearerTokenWithStringContext() string {
	response, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{Ctx: "user_12345"})
	if err != nil {
		fmt.Println("Error generating token:", err.GetMessage())
		return ""
	}
	return response.AccessToken
}

// Approach 2: an object context.
// Use a map when your policy needs several context values. Each key maps to a Skyflow CEL policy
// variable under request.context.*, for example:
//
//	request.context.role == "admin" && request.context.department == "finance"
func bearerTokenWithObjectContext() string {
	response, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{
		Ctx: map[string]interface{}{
			"role":       "admin",
			"department": "finance",
			"user_id":    "user_12345",
		},
	})
	if err != nil {
		fmt.Println("Error generating token:", err.GetMessage())
		return ""
	}
	return response.AccessToken
}

func main() {
	fmt.Println("Bearer token (string context):", bearerTokenWithStringContext())
	fmt.Println("Bearer token (object context):", bearerTokenWithObjectContext())
}
