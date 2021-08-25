package main

import (
	"fmt"

	saUtil "github.com/skyflowapi/skyflow-go/service-account/util"
)

func main() {
	token, err := saUtil.GenerateToken("/Users/bandi/Downloads/sa_credentials.json")
	if err != nil {
		panic(err)
	}

	fmt.Printf("token %v", token)
}
