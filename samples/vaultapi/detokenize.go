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
	vaultConfig1 := common.VaultConfig{VaultId: "<VAULT_ID1>", ClusterId: "<CLUSTER_ID1>", Env: common.DEV, Credentials: common.Credentials{Token: "<BEARER_TOKEN1>"}}
	vaultConfig2 := common.VaultConfig{VaultId: "<VAULT_ID2>", ClusterId: "<CLUSTER_ID2>", Env: common.DEV, Credentials: common.Credentials{Token: "<BEARER_TOKEN2>"}}
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
			res, errDetokenize := service.Detokenize(ctx, common.DetokenizeRequest{
				Tokens:        []string{"<TOKEN>", "<TOKEN>"},
				RedactionType: common.REDACTED,
			}, common.DetokenizeOptions{
				ContinueOnError: true,
			})
			if errDetokenize != nil {
				fmt.Println("ERROR: ", errDetokenize)
			} else {
				fmt.Println("RESPONSE: ", res)
			}
		}
	}

}
