/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example takes a file through de-identification and back with the Detect API.
 *
 * A local text file is de-identified into vault tokens in a token group, and the run is
 * polled until it finishes. Its redacted output (base64) is then re-identified as is. Only
 * vault tokens can be re-identified, so the configuration tokenizes names with
 * DeidentificationTypeVaultToken.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. De-identify the file into vault tokens and wait for the run.
 * 4. Re-identify the redacted output.
 * 5. Handle the response and errors, and save the output.
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

	// Step 3: De-identify the file into vault tokens and wait up to 2 minutes for the run
	deidentified, err := service.DeidentifyFile(context.Background(), common.DeidentifyFileRequest{
		FilePath: "<PATH_TO_FILE>/notes.txt",
		Configuration: &common.DetectConfiguration{
			Detect: &common.Detect{
				Entities: []common.Entity{{
					EntityType:           common.EntityTypeName,
					DeidentificationType: common.DeidentificationTypeVaultToken,
					Destination:          "<TOKEN_GROUP_NAME>",
				}},
			},
		},
	}, common.DeidentifyFileOptions{
		Poll:            &common.PollOptions{WaitTime: 2 * time.Minute},
		OutputDirectory: "<OUTPUT_DIRECTORY>", // processed-notes.txt is also written here
	})
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	if deidentified.Status != common.DetectRunStatusSuccess {
		fmt.Println("Run", deidentified.RunId, "did not succeed:", deidentified.Status, deidentified.Message)
		return
	}

	// Step 4: Re-identify the redacted output. A local file's outputs come back as base64,
	// so the redacted file can be sent as it is.
	var redacted *common.FileOutput
	for i := range deidentified.Output {
		if deidentified.Output[i].ProcessedFileType == common.ProcessedFileTypeRedactedFile {
			redacted = &deidentified.Output[i]
			break
		}
	}
	if redacted == nil || deidentified.OutputType != common.DeidentifySourceTypeBase64 {
		fmt.Println("The run returned no base64 redacted file")
		return
	}
	response, err := service.ReidentifyFile(context.Background(), common.ReidentifyFileRequest{
		DataSource: common.ReidentifySourceTypeBase64,
		Value:      redacted.ProcessedFile,
		DataFormat: redacted.ProcessedFileExtension,
	}, common.ReidentifyFileOptions{})

	// Step 5: Handle the response and errors, and save the output yourself
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	if len(response.Output) == 0 || response.OutputType != common.ReidentifySourceTypeBase64 {
		fmt.Println("Status:", response.Status, "- no base64 output to save")
		return
	}
	decoded, decodeErr := base64.StdEncoding.DecodeString(response.Output[0].ProcessedFile)
	if decodeErr != nil {
		fmt.Println("Could not decode the output:", decodeErr)
		return
	}
	if writeErr := os.WriteFile("reidentified-notes.txt", decoded, 0o600); writeErr != nil {
		fmt.Println("Could not write the output:", writeErr)
		return
	}
	fmt.Println("Status:", response.Status, "- saved reidentified-notes.txt")
}
