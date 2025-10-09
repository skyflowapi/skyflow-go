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
 * Example program to generate a Bearer Token
 * The token can be generated in two ways:
 * 1. Using the file path to a credentials.json file.
 * 2. Using the stringify JSON content of the credential file.
 */

func ExampleTokenGeneration() {
	// Generate bearer token using file path
	var filePath = "<FILE_PATH>"
	tokenResUsingCredFilePath, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{LogLevel: logger.DEBUG})
	if err != nil {
		fmt.Println("errors", *err)
	} else {
		fmt.Println("Token using file path:", tokenResUsingCredFilePath.AccessToken)
	}

	// Generate bearer token using cred as string
	var credString = "<CRED_STRING>"
	tokenResUsingCredString, errr := serviceaccount.GenerateBearerTokenFromCreds(credString, common.BearerTokenOptions{LogLevel: logger.DEBUG})
	if errr != nil {
		fmt.Println("errors", *errr)
	} else {
		fmt.Println("Token using cred string:", tokenResUsingCredString.AccessToken)
	}
}

func main() {
	ExampleTokenGeneration()
}
