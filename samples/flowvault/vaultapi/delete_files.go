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
 * the stored files. It is a hard delete and can't be undone: a PENDING upload is cancelled, a
 * READY file's object is removed, and a FAILED or SCAN_ERROR file just has its metadata cleared.
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
		ClusterId:   "<YOUR_CLUSTER_ID>", // from the vault URL: https://{clusterId}.vault.skyflowapis.com
		Env:         common.PROD,         // DEV, STAGE, SANDBOX or PROD
		Credentials: credentials,
	}

	// Step 2: Configure the Skyflow client
	skyflowClient, err := client.NewSkyflow(client.WithVaults(vaultConfig), client.WithLogLevel(logger.ERROR))
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), "requestId:", err.GetRequestId(), err.GetDetails())
		return
	}
	vault, err := skyflowClient.Vault(vaultConfig.VaultId)
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), "requestId:", err.GetRequestId(), err.GetDetails())
		return
	}

	// Step 3: Delete the files. Each record sets exactly one of SkyflowId or UniqueValues.
	request := common.DeleteFilesRequest{Records: []common.DeleteFilesRequestRecord{
		// By Skyflow ID
		{TableName: "<SENSITIVE_DATA_TABLE>", Columns: []string{"<FILE_COLUMN>"}, SkyflowId: "<SKYFLOW_ID>"},
		// By unique column value; each map can match more than one record
		{TableName: "<SENSITIVE_DATA_TABLE>", Columns: []string{"<FILE_COLUMN>"},
			UniqueValues: []map[string]interface{}{{"<UNIQUE_COLUMN>": "<UNIQUE_VALUE>"}}},
	}}
	response, err := vault.DeleteFiles(context.Background(), request, common.DeleteFilesOptions{})

	// Step 4: Handle the response and errors.
	// One record per matched record. Each has SkyflowId, TableName, Columns (each with Column
	// and Status "DELETED"; empty on a failed record), HttpCode, Error and RequestId.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), "requestId:", err.GetRequestId(), err.GetDetails())
		return
	}
	// Records fail independently: check each one.
	for _, record := range response.Records {
		if record.Error != "" {
			fmt.Printf("%s failed (%d): %s [requestId=%s]\n", record.SkyflowId, record.HttpCode, record.Error, record.RequestId)
			continue
		}
		for _, column := range record.Columns {
			fmt.Printf("%s %s: %s\n", record.SkyflowId, column.Column, column.Status)
		}
	}
}
