package serviceaccount

import (
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func ExampleTokenGenerationWithScope() {
	// Generate bearer token using file path
	var filePath = "<FILE_PATH>"
	res, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{LogLevel: logger.DEBUG, RoleIDs: []string{"<ROLE_ID_1>", "<ROLE_ID_2>", "<ROLE_ID_3>"}})
	if err != nil {
		fmt.Println("errors", *err)
	} else {
		fmt.Println("here it is", res.AccessToken)

	}

	// Generate bearer token using cred as string
	var credString = "<CRED_STRING>"
	res1, err1 := serviceaccount.GenerateBearerTokenFromCreds(credString, common.BearerTokenOptions{LogLevel: logger.DEBUG, RoleIDs: []string{"<ROLE_ID_1>", "<ROLE_ID_2>", "<ROLE_ID_3>"}})
	if err1 != nil {
		fmt.Println("errors", *err1)
	} else {
		fmt.Println("Token", res1.AccessToken)

	}
}
