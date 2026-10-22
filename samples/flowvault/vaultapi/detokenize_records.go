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
 * This example returns the values behind tokens in a FlowVault vault.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Detokenize tokens.
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

	// Step 3: Detokenize tokens. Redactions are set per token group.
	request := common.DetokenizeRequest{
		Tokens: []string{"<TOKEN1>", "<TOKEN2>"},
		// TokenGroupRedactions: []common.TokenGroupRedaction{{TokenGroupName: "<TOKEN_GROUP_NAME>", Redaction: "<REDACTION>"}},
	}
	response, err := vault.Detokenize(context.Background(), request, common.DetokenizeOptions{})

	// Step 4: Handle the response and errors.
	// One record per input token, in order. Each has Token, Value, TokenGroupName, Metadata
	// (SkyflowId and TableName), HttpCode, Error and RequestId.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error == "" {
			fmt.Printf("%s -> %v (%s)\n", record.Token, record.Value, record.TokenGroupName)
		} else {
			fmt.Printf("%s failed (%d): %s [requestId=%s]\n", record.Token, record.HttpCode, record.Error, record.RequestId)
		}
	}
}
