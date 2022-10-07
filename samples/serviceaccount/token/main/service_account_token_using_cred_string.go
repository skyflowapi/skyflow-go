/*
Copyright (c) 2022 Skyflow, Inc.
*/
package main

import (
	"fmt"

	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	saUtil "github.com/skyflowapi/skyflow-go/serviceaccount/util"
)

var token = ""

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()
	logger.SetLogLevel(logger.INFO) //set loglevel to INFO
	credentials:= "<credentials_in_string_format>"
	if saUtil.IsExpired(token) {
		newToken, err := saUtil.GenerateBearerTokenFromCreds(credentials)
		if err != nil {
			panic(err)
		} else {
			token = newToken.AccessToken
		}
		fmt.Println("%v", token)
	}
}
