/*
Copyright (c) 2022 Skyflow, Inc.
*/

package serviceaccount

import (
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func ExampleTokenGeneration() {
	// Generate bearer token using file path
	var filePath = "<FILE_PATH>"
	res, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{LogLevel: logger.DEBUG})
	if err != nil {
		fmt.Println("errors", *err)
	} else {
		fmt.Println("here it is", res.AccessToken)

	}

	// Generate bearer token using cred as string
	var credString = "<CRED_STRING>"
	res1, err1 := serviceaccount.GenerateBearerTokenFromCreds(credString, common.BearerTokenOptions{LogLevel: logger.DEBUG})
	if err1 != nil {
		fmt.Println("errors", *err1)
	} else {
		fmt.Println("Token", res1.AccessToken)

	}
}
