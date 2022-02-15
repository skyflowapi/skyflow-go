package main

import (
	"fmt"

	saUtil "github.com/skyflowapi/skyflow-go/service-account/util"
)

func main() {
	filePath := ""
	token, err := saUtil.GenerateBearerToken(filePath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("token %v", token)
}
