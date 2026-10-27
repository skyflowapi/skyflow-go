/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example de-identifies a file with the Detect V2 API.
 *
 * By default DeidentifyFile only submits the file and returns a run ID: fetch the result
 * later with GetRun (see get_run.go). Set Poll in the options to wait for the run instead;
 * for a local file (FilePath or FileObject) the outputs of a successful run are then
 * written to OutputDirectory as processed-<name>.<ext> and processed-entities-<name>.json.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Submit a file in the vault and keep its run ID.
 * 4. Submit a local file and wait for the run.
 * 5. Handle the response and errors.
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

	// Step 3: Submit a file already in the vault; only the run ID comes back
	submitted, err := service.DeidentifyFile(context.Background(), common.DeidentifyFileRequest{
		DataSource:      common.DataSourceTypeSkyflowID, // or DataSourceTypePresignedURL / DataSourceTypeBase64
		Value:           "<SKYFLOW_ID>",
		DataFormat:      common.FileDataFormatPDF,
		ConfigurationId: "<CONFIGURATION_ID>",
	}, common.DeidentifyFileOptions{})
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println("Run ID:", submitted.RunId) // retrieve the result later with GetRun

	// Step 4: Submit a local file and wait up to 2 minutes for the run
	response, err := service.DeidentifyFile(context.Background(), common.DeidentifyFileRequest{
		FilePath:        "<PATH_TO_FILE>/patient-notes.pdf", // sent as BASE64, format inferred from the extension
		ConfigurationId: "<CONFIGURATION_ID>",
	}, common.DeidentifyFileOptions{
		Poll:            &common.PollOptions{WaitTime: 2 * time.Minute}, // &common.PollOptions{} uses 60s and 23 attempts
		OutputDirectory: "<OUTPUT_DIRECTORY>",
	})

	// Step 5: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	switch response.Status {
	case common.DetectRunStatusSuccess: // outputs are written to OutputDirectory
		fmt.Println("Outputs:", len(response.Output))
		if response.Metrics != nil && response.Metrics.Pages != nil {
			fmt.Println("Pages:", *response.Metrics.Pages)
		}
	case common.DetectRunStatusInProgress, common.DetectRunStatusQueued: // the wait ran out: poll later with GetRun
		fmt.Println("Still running, run ID:", response.RunId)
	default: // FAILED or UNKNOWN
		fmt.Println("Run failed:", response.Message)
	}
}
