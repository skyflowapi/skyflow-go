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
 * The token can be generated in two ways:
 * 1. Using the file path to a credentials.json file.
 * 2. Using the stringify JSON content of the credential file.
 */

func ExampleTokenGenerationWithContext() {
	// Generate bearer token using file path
	var filePath = "<FILE_PATH>"
	res, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{LogLevel: logger.DEBUG, Ctx: "<CONTEXT>"})
	if err != nil {
		fmt.Println("errors", *err)
	} else {
		fmt.Println("Token", res.AccessToken)

	}
	// Generate bearer token using cred as string
	var credString = "<CRED_STRING>"
	res1, err1 := serviceaccount.GenerateBearerTokenFromCreds(credString, common.BearerTokenOptions{LogLevel: logger.DEBUG, Ctx: "<CONTEXT>"})
	if err1 != nil {
		fmt.Println("errors", *err1)
	} else {
		fmt.Println("Token", res1.AccessToken)

	}

}

func main() {
	ExampleTokenGenerationWithContext()
}
