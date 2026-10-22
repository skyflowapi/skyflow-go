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
 * This example deletes records from a FlowVault vault by Skyflow ID.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Delete records.
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

	// Step 3: Delete records. Set exactly one of SkyflowIds or UniqueValues.
	request := common.DeleteRequest{
		TableName:  "<SENSITIVE_DATA_TABLE>",
		SkyflowIds: []string{"<SKYFLOW_ID1>", "<SKYFLOW_ID2>"},
		// UniqueValues: []map[string]interface{}{{"email": "john@example.com"}},
	}
	response, err := vault.Delete(context.Background(), request, common.DeleteOptions{})

	// Step 4: Handle the response and errors.
	// One record per input, in order. Each has SkyflowId, HttpCode, Error and RequestId.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error == "" {
			fmt.Println(record.SkyflowId, "deleted")
		} else {
			fmt.Printf("%s failed (%d): %s [requestId=%s]\n", record.SkyflowId, record.HttpCode, record.Error, record.RequestId)
		}
	}
}
