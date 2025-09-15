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
 * This example demonstrates how to use the Skyflow Go SDK to detokenize records from your vault
 * by converting tokens back into their original sensitive data values.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault.
 * 4. Detokenizing records by providing tokens and receiving original values.
 * 5. Handling the response and errors.
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
			fmt.Println(*serviceError)
		} else {
			ctx := context.TODO()
			req := common.DetokenizeRequest{DetokenizeData: []common.DetokenizeData{
				{
					Token:         "<TOKEN1>",
					RedactionType: common.PLAIN_TEXT,
				},
				{
					Token:         "<TOKEN2>",
					RedactionType: common.PLAIN_TEXT,
				},
			}}
			// Step 4: Detokenize records by providing tokens and receiving original values
			detokenizeRes, errDetokenize := service.Detokenize(ctx, req, common.DetokenizeOptions{
				ContinueOnError: true,
			})
			// Step 5: Handling the response and errors
			if errDetokenize != nil {
				fmt.Println("ERROR: ", *errDetokenize)
			} else {
				fmt.Println("RESPONSE: ", detokenizeRes)
			}
		}
	}

}
