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
 * This example demonstrates how to use the Skyflow Go SDK to retrieve records
 * from your vault using record IDs and table names.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault.
 * 4. Retrieve records using record IDs and table names.
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
			fmt.Println(*serviceError)
		} else {
			ctx := context.TODO()
			// Step 4: Retrieve records using record IDs and table names
			getRes, getErr := service.Get(ctx, common.GetRequest{
				Table: "<TABLE_NAME>", // Name of the table
				Ids: []string{
					"<SKYFLOW_ID_1>", // List of Skyflow IDs to be fetched
					"<SKYFLOW_ID_2>",
				},
			}, common.GetOptions{
				ReturnTokens: true,
			})
			// Step 5: Handle the response and errors
			if getErr != nil {
				fmt.Println("ERROR: ", *getErr)
			} else {
				fmt.Println("RESPONSE: ", getRes.Data)
			}
		}
	}

}
