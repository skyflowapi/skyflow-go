/*
Copyright (c) 2022 Skyflow, Inc.
*/
package main

import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func main() {
	vaultConfig1 := common.VaultConfig{VaultId: "<VAULT_ID1>", ClusterId: "<CLUSTER_ID1>", Env: common.DEV, Credentials: common.Credentials{CredentialsString: "<STRINGIFIED_JSON_VALUE>"}}
	vaultConfig2 := common.VaultConfig{VaultId: "<VAULT_ID2>", ClusterId: "<CLUSTER_ID2>", Env: common.DEV, Credentials: common.Credentials{CredentialsString: "<STRINGIFIED_JSON_VALUE>"}}
	var arr []common.VaultConfig
	arr = append(arr, vaultConfig2, vaultConfig1)
	skyflowInstance, err := client.NewSkyflow(
		client.WithVaults(arr...),
		client.WithCredentials(common.Credentials{}), // pass credentials if not provided in vault config
		client.WithLogLevel(logger.DEBUG),
	)
	if err != nil {
		fmt.Println(err)
	} else {
		service, serviceErr := skyflowInstance.Vault("<VAULT_ID>")
		if serviceErr != nil {
			fmt.Println(serviceErr)
		} else {
			ctx := context.TODO()
			var reqArray []common.TokenizeRequest
			reqArray = append(reqArray, common.TokenizeRequest{
				ColumnGroup: "<COLUMN_GROUP_NAME>",
				Value:       "<VALUE>",
			})
			res, tokenizeErr := service.Tokenize(ctx, reqArray)
			if tokenizeErr != nil {
				fmt.Println("ERROR: ", tokenizeErr)
			} else {
				fmt.Println("RESPONSE: ", res)
			}
		}
	}

}
