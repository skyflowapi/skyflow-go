/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example uploads files into the file columns of a FlowVault vault.
 *
 * The SDK requests a signed upload URL for each column, then uploads each file to its URL.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Upload the files.
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

	// Step 3: Upload the files. Each column takes exactly one of FilePath, Base64 (needs
	// FileName) or FileObject. Leave SkyflowId empty to create a new record, or set it to upload
	// into an existing record.
	file, openErr := os.Open("<PATH_TO_ANOTHER_FILE>")
	if openErr != nil {
		fmt.Println("Could not open file:", openErr)
		return
	}
	defer file.Close() // the caller closes a FileObject

	request := common.UploadFilesRequest{Records: []common.UploadFilesRequestRecord{
		{
			TableName: "<SENSITIVE_DATA_TABLE>",
			Columns: []common.UploadFilesRequestColumn{
				{Column: "<FILE_COLUMN_1>", FilePath: "<PATH_TO_FILE>"},
				{Column: "<FILE_COLUMN_2>", Base64: "<BASE64_FILE_CONTENT>", FileName: "<FILE_NAME>"},
				{Column: "<FILE_COLUMN_3>", FileObject: file},
			},
		},
		// {TableName: "<SENSITIVE_DATA_TABLE>", SkyflowId: "<SKYFLOW_ID>",
		// 	Columns: []common.UploadFilesRequestColumn{{Column: "<FILE_COLUMN_1>", FilePath: "<PATH_TO_FILE>"}}},
	}}
	response, err := vault.UploadFiles(context.Background(), request, common.UploadFilesOptions{})

	// Step 4: Handle the response and errors.
	// One record per input record, in order. Each has SkyflowId, TableName, Columns, HttpCode,
	// Error and RequestId. Each column has Column, FileName, UploadStatus (UPLOADED, FAILED or
	// SKIPPED) and Error.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error != "" {
			fmt.Printf("record failed (%d): %s [requestId=%s]\n", record.HttpCode, record.Error, record.RequestId)
		}
		for _, column := range record.Columns {
			fmt.Printf("%s %s: %s %s\n", record.SkyflowId, column.Column, column.UploadStatus, column.Error)
		}
	}
}
