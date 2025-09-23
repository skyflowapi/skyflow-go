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
 * This example demonstrates how to use the Skyflow Go SDK to retrieve the status and results
 * of a detect run for sensitive data of files.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault with detect service.
 * 4. Retrieve the detect run status and results.
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
		fmt.Println(*err)
	} else {
		// Step 3: Configure the vault with detect service
		service, serviceErr := skyflowInstance.Detect("<VAULT_ID>") // Replace with your vault ID from the vault config
		if serviceErr != nil {
			fmt.Println(*serviceErr)
		} else {
			ctx := context.TODO()
			// Step 4: Retrieve the detect run status and results
			getDetectRunRes, deidentifyFileErr := service.GetDetectRun(ctx, common.GetDetectRunRequest{
				RunId: "<RUN_ID_FROM_DEIDENTIFY_FILE_RESPONSE>",
			})

			// Step 5: Handle the response and errors.
			if deidentifyFileErr != nil {
				fmt.Println("ERROR: ", *deidentifyFileErr)
			} else {
				fmt.Println("RESPONSE: ", getDetectRunRes)
			}
		}
	}

}
