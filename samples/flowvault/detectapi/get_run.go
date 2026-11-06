/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example looks up a de-identify file run with the Detect API, using the run ID
 * DeidentifyFile returned, and saves its outputs.
 *
 * GetRun writes nothing to disk: each output comes back as base64 or as a presigned URL,
 * as the run's OutputType says, for you to save.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Look up the run.
 * 4. Handle the response and errors, and save the outputs.
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

	// Step 3: Look up the run
	run, err := service.GetRun(context.Background(), common.GetRunRequest{RunId: "<RUN_ID>"}, common.GetRunOptions{})

	// Step 4: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	switch run.Status {
	case common.DetectRunStatusSuccess:
		saveOutputs(run)
	case common.DetectRunStatusInProgress, common.DetectRunStatusQueued:
		fmt.Println("Still running: look it up again later")
	default: // FAILED or UNKNOWN: Message says why
		fmt.Println("Run failed:", run.Status, run.Message)
	}
}

// saveOutputs writes each base64 output of a successful run to the current directory.
func saveOutputs(run *common.GetRunResponse) {
	for i, output := range run.Output {
		if run.OutputType != common.DeidentifySourceTypeBase64 {
			// output.ProcessedFile is a presigned URL: anyone holding it can download the
			// output, so fetch it here and don't log it.
			fmt.Println("Output", i, "of type", output.ProcessedFileType, "is available at a presigned URL")
			continue
		}
		decoded, decodeErr := base64.StdEncoding.DecodeString(output.ProcessedFile)
		if decodeErr != nil {
			fmt.Println("Could not decode output", i, decodeErr)
			return
		}
		// One file per output, named by its index, type and extension, e.g.
		// run-output-0-redacted_file.pdf. The type and extension come from the server, so
		// only safe characters are kept: the name can't leave this directory.
		name := fmt.Sprintf("run-output-%d-%s.%s", i,
			namePart(string(output.ProcessedFileType), "output"), namePart(string(output.ProcessedFileExtension), "bin"))
		if writeErr := os.WriteFile(name, decoded, 0o600); writeErr != nil {
			fmt.Println("Could not write output", i, writeErr)
			return
		}
		fmt.Println("Saved", name)
	}
}

// namePart lowercases s and keeps only letters, digits and underscores, so it is safe in a
// file name. It returns fallback when nothing is left.
func namePart(s, fallback string) string {
	kept := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, strings.ToLower(s))
	if kept == "" {
		return fallback
	}
	return kept
}
