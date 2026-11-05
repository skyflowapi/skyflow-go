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
 * This example updates a record in a FlowVault vault.
 *
 * UpdateType is set on the request: UPDATE (the default) changes only the given columns,
 * REPLACE overwrites the whole record.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Update the record.
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

	// Step 3: Update the record. Set the table name on the request or on every record, not both.
	request := common.UpdateRequest{
		TableName:  "<SENSITIVE_DATA_TABLE>",
		UpdateType: common.UpdateTypeUpdate,
		Records: []common.UpdateRequestRecord{
			{SkyflowId: "<SKYFLOW_ID>", Data: map[string]interface{}{"name": "Jane Doe"}},
		},
	}
	response, err := vault.Update(context.Background(), request, common.UpdateOptions{})

	// Step 4: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error == "" {
			fmt.Printf("%s updated -> tokens=%v\n", record.SkyflowId, record.Tokens)
		} else {
			fmt.Printf("%s failed (%d): %s [requestId=%s]\n", record.SkyflowId, record.HttpCode, record.Error, record.RequestId)
		}
	}
}
