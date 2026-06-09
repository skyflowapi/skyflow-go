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
 * Example program to generate a Signed Token
 * The token can be generated in two ways:
 * 1. Using the file path to a credentials.json file.
 * 2. Using the JSON content of the credentials file as a string.
 */

func SignedDataTokenGenerationSample() {
	var filePath = "<FILE_PATH>"

	// signed data token generation using cred file
	var tokens []string
	tokens = append(tokens, "<TOKEN>")
	res, err := serviceaccount.GenerateSignedDataTokens(filePath, common.SignedDataTokensOptions{
		DataTokens: tokens,
		TimeToLive: 60, // in seconds
		LogLevel:   logger.ERROR,
	})
	if err != nil {
		fmt.Println("ERROR: ", err)
	} else {
		fmt.Println("RESPONSE:", res)
	}

	// signed data token generation using cred string
	var credString = "<CRED_STRING>"
	var tokens2 []string
	tokens2 = append(tokens2, "<TOKEN>")
	res2, err1 := serviceaccount.GenerateSignedDataTokensFromCreds(credString, common.SignedDataTokensOptions{
		DataTokens: tokens2,
		TimeToLive: 60, // in seconds
		LogLevel:   logger.ERROR,
	})
	if err1 != nil {
		fmt.Println("ERROR: ", err1)
	} else {
		fmt.Println("RESPONSE:", res2)
	}

}

func main() {
	SignedDataTokenGenerationSample()
}
