package serviceaccount

import (
	"fmt"
	"skyflow-go/v2/serviceaccount"
	"skyflow-go/v2/utils/common"
	"skyflow-go/v2/utils/logger"
)

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
