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
 * This example deletes files from the file columns of records in a FlowVault vault, including
 * the stored files.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Delete the files.
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

	// Step 3: Delete the files. Each record sets exactly one of SkyflowId or UniqueValues.
	request := common.DeleteFilesRequest{Records: []common.DeleteFilesRequestRecord{
		{TableName: "<SENSITIVE_DATA_TABLE>", Columns: []string{"<FILE_COLUMN>"}, SkyflowId: "<SKYFLOW_ID>"},
		// {TableName: "<SENSITIVE_DATA_TABLE>", Columns: []string{"<FILE_COLUMN>"},
		// 	UniqueValues: []map[string]interface{}{{"email": "john@example.com"}}},
	}}
	response, err := vault.DeleteFiles(context.Background(), request, common.DeleteFilesOptions{})

	// Step 4: Handle the response and errors.
	// One record per matched record. Each has SkyflowId, TableName, Columns (each with Column
	// and Status "DELETED"; empty on a failed record), HttpCode, Error and RequestId.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error == "" {
			fmt.Printf("%s -> %v\n", record.SkyflowId, record.Columns)
		} else {
			fmt.Printf("%s failed (%d): %s [requestId=%s]\n", record.SkyflowId, record.HttpCode, record.Error, record.RequestId)
		}
	}
}
