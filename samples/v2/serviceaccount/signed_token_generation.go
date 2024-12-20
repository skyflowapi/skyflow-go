package serviceaccount

import (
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

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
	var tokens2 []string
	tokens2 = append(tokens2, "<TOKEN>")
	res2, err1 := serviceaccount.GenerateSignedDataTokensFromCreds(filePath, common.SignedDataTokensOptions{
		DataTokens: tokens2,
		TimeToLive: 0,
		LogLevel:   logger.ERROR,
	})
	if err1 != nil {
		fmt.Println("ERROR: ", err)
	} else {
		fmt.Println("RESPONSE:", res2)
	}

}
