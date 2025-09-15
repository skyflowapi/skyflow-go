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
 * This example demonstrates how to use the Skyflow Go SDK to update existing records
 * in your vault using record IDs and new values.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault.
 * 4. Updating records with new values using record IDs.
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
		service, serviceErr := skyflowInstance.Vault("<VAULT_ID>") // Replace with your vault ID from the vault config
		if serviceErr != nil {
			fmt.Println(*serviceErr)
		} else {
			ctx := context.TODO()
			// Step 4: Update records with new values using record IDs
			resUpdate, errUpdate := service.Update(ctx, common.UpdateRequest{
				Table: "<TABLE_NAME>",
				Data: map[string]interface{}{
					"skyflow_id": "<SKYFLOW_ID>", // Replace with the actual id of the record to be updated
					"<FIELD1>":   "<VALUE1>",     // Replace with the actual field and value to be updated
					"<FIELD2>":   "<VALUE2>",     // Replace with the actual field and value to be updated
				},
			}, common.UpdateOptions{
				ReturnTokens: true,
				TokenMode:    common.DISABLE,
			})

			// Step 5: Handling the response and errors
			if errUpdate != nil {
				fmt.Println("ERROR: ", *errUpdate)
			} else {
				fmt.Println("response: ", resUpdate)
			}
		}
	}
}
