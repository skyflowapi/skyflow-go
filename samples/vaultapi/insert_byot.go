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

/**
 * This example demonstrates how to use the Skyflow Go SDK to insert records using BYOT
 * (Bring Your Own Token) functionality into your vault.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault.
 * 4. Inserting records with custom tokens using BYOT functionality.
 * 5. Handle the response and errors.
 */

func main() {
	// Step 1: Set up Skyflow vault credentials
	vaultConfig1 := common.VaultConfig{VaultId: "<VAULT_ID1>", ClusterId: "<CLUSTER_ID1>", Env: common.PROD, Credentials: common.Credentials{Token: "<BEARER_TOKEN1>"}}
	vaultConfig2 := common.VaultConfig{VaultId: "<VAULT_ID2>", ClusterId: "<CLUSTER_ID2>", Env: common.SANDBOX, Credentials: common.Credentials{Token: "<BEARER_TOKEN2>"}}
	var arr []common.VaultConfig
	arr = append(arr, vaultConfig2, vaultConfig1)

	// Step 2: Configure the skyflow client
	skyflowInstance, err := client.NewSkyflow(
		client.WithVaults(arr...),
		client.WithCredentials(common.Credentials{}), // Pass credentials if not provided in vault config
		client.WithLogLevel(logger.ERROR),            // Use LogLevel.ERROR in production
	)
	if err != nil {
		fmt.Println(err)
	} else {
		// Step 3: Configure the vault
		service, serviceError := skyflowInstance.Vault("<VAULT_ID>") // Replace with your vault ID from the vault config
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
			// Step 4: Insert records with custom tokens using BYOT functionality
			insert, insertErr := service.Insert(ctx, common.InsertRequest{
				Table:  "<TABLE_NAME>",
				Values: values,
			}, common.InsertOptions{ContinueOnError: false, ReturnTokens: true, TokenMode: common.ENABLE, Tokens: tokens})

			// Step 5: Handle the response and errors
			if insertErr != nil {
				fmt.Println("ERROR: ", *insertErr)
			} else {
				fmt.Println("RESPONSE: ", insert)
			}
		}
	}

}
