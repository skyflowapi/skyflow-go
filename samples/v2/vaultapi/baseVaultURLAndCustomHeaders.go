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
	// Step 1: Set up Skyflow vault credentials
	vaultConfig1 := common.VaultConfig{
		VaultId:      "<VAULT_ID>",
		BaseVaultURL: "<VAULT_URL>", // Custom base vault URL
		Env:          common.SANDBOX,
		Credentials: common.Credentials{
			Token: "<TOKEN>",
		},
	}
	var arr []common.VaultConfig
	arr = append(arr, vaultConfig1)

	var customHeaders = make(map[string]string) // Create a map for custom headers
	customHeaders["x-custom-header"] = "custom-header-value"
	customHeaders["X-Gateway-Route-ID"] = "UNIQUE_VAULT_ROUTING_ID"
	customHeaders["X-Application-Source"] = "sample-application"

	// Step 2: Configure the skyflow client
	skyflowInstance, err := client.NewSkyflow(
		client.WithVaults(arr...),
		client.WithLogLevel(logger.INFO),
		client.WithCustomHeaders(customHeaders), // Added custom headers
	)
	if err != nil {
		fmt.Println(*err)
	} else {
		// Step 3: Configure the vault
		service, serviceError := skyflowInstance.Vault("<VAULT_ID>") // Replace with your vault ID from the vault config
		if serviceError != nil {
			fmt.Println(*serviceError)
		} else {
			ctx := context.TODO()
			values := make([]map[string]interface{}, 0)
			values = append(values, map[string]interface{}{
				"<FIELD_1>": "<VALUE_1>",
			})
			values = append(values, map[string]interface{}{
				"<FIELD_2>": "<VALUE_2>",
				"<FIELD_3>": "<VALUE_3>",
			})

			// Step 4: Insert records with proper data and receive tokens
			insert, insertErr := service.Insert(ctx, common.InsertRequest{
				Table:  "<TABLE>", // Replace with actual table
				Values: values,
			}, common.InsertOptions{ContinueOnError: false, ReturnTokens: true})

			// Step 5: Handle the response and errors
			if insertErr != nil {
				fmt.Println("ERROR: ", *insertErr)
			} else {
				fmt.Println("RESPONSE: ", insert)
			}
		}
	}

}
