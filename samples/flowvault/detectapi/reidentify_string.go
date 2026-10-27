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
 * This example re-identifies tokens in a string with the Detect V2 API.
 *
 * Without RedactionLevel every token is replaced by its original value. In a FlowVault
 * vault, a RedactionLevel selects tokens by TokenGroupName and shows them through a
 * RedactionPattern.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Re-identify the string.
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

	// Step 2: Configure the Skyflow client and get the vault's detect service
	skyflowClient, err := client.NewSkyflow(client.WithVaults(vaultConfig), client.WithLogLevel(logger.ERROR))
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	service, err := skyflowClient.Detect(vaultConfig.VaultId)
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}

	// Step 3: Re-identify the string
	response, err := service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{
		Text: "My name is [NAME_1].",
		RedactionLevel: []common.RedactionLevel{{
			TokenGroupName:   "<TOKEN_GROUP_NAME>",
			RedactionPattern: "<REDACTION_PATTERN_NAME>",
		}},
	}, common.ReidentifyStringOptions{})

	// Step 4: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println(response.ProcessedText) // My name is J**n D*e.
}
