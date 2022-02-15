package main

import (
	"fmt"

	saUtil "github.com/skyflowapi/skyflow-go/service-account/util"
)

func main() {
	filePath := ""
	token, err := saUtil.GenerateBearerToken(filePath)
	// token, err := saUtil.GenerateBearerTokenFromCreds("<creds_as_String>")
	if err != nil {
		panic(err)
	}

	fmt.Printf("token %v", token)
}
