/*
Copyright (c) 2022 Skyflow, Inc.
*/

package main

import (
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * Example program to generate a Bearer Token with context.
 * The token can be generated in three ways:
 * 1. Using a string context identifier.
 * 2. Using a JSON object context (map) for conditional data access policies.
 * 3. Using credentials as a string with context.
 */

func ExampleTokenGenerationWithContext() {
	var filePath = "<FILE_PATH>"

	// Approach 1: Bearer token with string context
	// Use a simple string identifier when your policy references a single context value.
	res, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{
		LogLevel: logger.DEBUG,
		Ctx:      "user_12345",
	})
	if err != nil {
		fmt.Println("errors", *err)
	} else {
		fmt.Println("Token (string context):", res.AccessToken)
	}

	// Approach 2: Bearer token with JSON object context
	// Use a map when your policy needs multiple context values for conditional data access.
	// Each key maps to a Skyflow CEL policy variable under request.context.*
	// For example: request.context.role == "admin" && request.context.department == "finance"
	ctxMap := map[string]interface{}{
		"role":       "admin",
		"department": "finance",
		"user_id":    "user_12345",
	}
	res2, err2 := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{
		LogLevel: logger.DEBUG,
		Ctx:      ctxMap,
	})
	if err2 != nil {
		fmt.Println("errors", *err2)
	} else {
		fmt.Println("Token (object context):", res2.AccessToken)
	}

	// Approach 3: Bearer token from credentials string with context
	var credString = "<CRED_STRING>"
	res3, err3 := serviceaccount.GenerateBearerTokenFromCreds(credString, common.BearerTokenOptions{
		LogLevel: logger.DEBUG,
		Ctx:      "user_12345",
	})
	if err3 != nil {
		fmt.Println("errors", *err3)
	} else {
		fmt.Println("Token (creds string):", res3.AccessToken)
	}
}

func main() {
	ExampleTokenGenerationWithContext()
}
