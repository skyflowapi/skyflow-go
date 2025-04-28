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
		service, serviceError := skyflowInstance.Vault("<VAULT_ID>")
		if serviceError != nil {
			fmt.Println(serviceError)
		} else {
			ctx := context.TODO()
			values := make([]map[string]interface{}, 0)
			values = append(values, map[string]interface{}{
				"<FIELD_NAME1_1>": "<VALUE_1>",
			})
			values = append(values, map[string]interface{}{
				"<FIELD_NAME_2>": "<VALUE_1>",
				"<FIELD_NAME_3>": "<VALUE_2>",
			})

			tokens := make([]map[string]interface{}, 0)
			tokens = append(values, map[string]interface{}{
				"<FIELD_NAME1_1>": "<TOKEN>",
			})
			tokens = append(tokens, map[string]interface{}{
				"<FIELD_NAME_2>": "<TOKEN>",
			})

			insert, err4 := service.Insert(ctx, common.InsertRequest{
				Table:  "<TABLE_NAME>",
				Values: values,
			}, common.InsertOptions{ContinueOnError: false, ReturnTokens: true, TokenMode: common.ENABLE, Tokens: tokens})
			if err4 != nil {
				fmt.Println("ERROR: ", *err4)
			} else {
				fmt.Println("RESPONSE", insert)
			}
		}
	}

}
