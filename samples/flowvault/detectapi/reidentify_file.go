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
 * This example re-identifies a file with the Detect V2 API. ReidentifyFile is synchronous:
 * there is no run to poll. For a local file the outputs are written to OutputDirectory.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Re-identify the file.
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

	// Step 3: Re-identify the file
	response, err := service.ReidentifyFile(context.Background(), common.ReidentifyFileRequest{
		FilePath: "<PATH_TO_FILE>/processed-notes.txt",
		RedactionLevel: []common.RedactionLevel{{
			TokenGroupName:   "<TOKEN_GROUP_NAME>",
			RedactionPattern: "<REDACTION_PATTERN_NAME>",
		}},
	}, common.ReidentifyFileOptions{OutputDirectory: "<OUTPUT_DIRECTORY>"})

	// Step 4: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println("Status:", response.Status, "Outputs:", len(response.Output))
}
