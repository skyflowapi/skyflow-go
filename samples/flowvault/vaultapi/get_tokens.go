/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"context"
	"fmt"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example looks up the existing tokens for values in a FlowVault vault.
 *
 * Only deterministic token groups are supported, and no new tokens are created.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Get tokens for values.
 * 4. Handle the response and errors.
 */
func main() {
	// Step 1: Set up credentials and the vault config
	credentials := common.Credentials{Path: "<PATH_TO_YOUR_CREDENTIALS_JSON>"} // or ApiKey / Token / CredentialsString
	vaultConfig := common.VaultConfig{
		VaultId:     "<YOUR_VAULT_ID>",
		ClusterId:   "<YOUR_CLUSTER_ID>", // from the vault URL: https://{clusterId}.skyvault.skyflowapis.com
		Env:         common.PROD,         // DEV, STAGE, SANDBOX or PROD
		Credentials: credentials,
	}

	// Step 2: Configure the Skyflow client
	skyflowClient, err := client.NewSkyflow(client.WithVaults(vaultConfig), client.WithLogLevel(logger.ERROR))
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	vault, err := skyflowClient.Vault(vaultConfig.VaultId)
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}

	// Step 3: Get tokens for values
	request := common.GetTokensRequest{Records: []common.GetTokensRequestRecord{
		{Value: "john@example.com", TokenGroupName: "<TOKEN_GROUP_NAME>"},
		{Value: "jane@example.com", TokenGroupName: "<TOKEN_GROUP_NAME>"},
	}}
	response, err := vault.GetTokens(context.Background(), request, common.GetTokensOptions{})

	// Step 4: Handle the response and errors.
	// One record per input, in order. Each has Value, TokenGroupName, Token, HttpCode, Error
	// and RequestId. A value with no token has Error set.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error == "" {
			fmt.Printf("%v -> %s\n", record.Value, record.Token)
		} else {
			fmt.Printf("%v failed (%d): %s [requestId=%s]\n", record.Value, record.HttpCode, record.Error, record.RequestId)
		}
	}
}
