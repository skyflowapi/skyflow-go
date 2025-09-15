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
 * This example demonstrates how to use the Skyflow Go SDK to query records from the vault
 * from your vault using query and conditions.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault.
 * 4. Querying records from vault using query and handling results.
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
			// Step 4: Query records from vault using query and handle results
			queryRes, queryErr := service.Query(ctx, common.QueryRequest{Query: "SELECT * FROM demo WHERE skyflow_id='<ID>'"}) // Pass query here

			// Step 5: Handle the response and errors
			if queryErr != nil {
				fmt.Println("ERROR: ", *queryErr)
			} else {
				fmt.Println("RESPONSE: ", queryRes)
			}
		}
	}
}
