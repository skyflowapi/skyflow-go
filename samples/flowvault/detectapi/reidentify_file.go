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
 * This example re-identifies a file with the Detect API. ReidentifyFile is synchronous:
 * there is no run to poll. The file is sent as base64 in Value, and the re-identified
 * output comes back in the response: the SDK writes nothing to disk.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Re-identify the file.
 * 4. Handle the response and errors, and save the output.
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

	// Step 3: Re-identify the file, sent as base64
	content, readErr := os.ReadFile("<PATH_TO_FILE>/processed-notes.txt")
	if readErr != nil {
		fmt.Println("Could not read the file:", readErr)
		return
	}
	response, err := service.ReidentifyFile(context.Background(), common.ReidentifyFileRequest{
		DataSource: common.ReidentifySourceTypeBase64, // or ReidentifySourceTypePresignedURL with a URL in Value
		Value:      base64.StdEncoding.EncodeToString(content),
		DataFormat: common.FileDataFormatTXT,
		RedactionLevel: []common.RedactionRule{common.RedactionLevel{
			TokenGroupName:   "<TOKEN_GROUP_NAME>",
			RedactionPattern: "<REDACTION_PATTERN_NAME>",
		}},
	}, common.ReidentifyFileOptions{})

	// Step 4: Handle the response and errors, and save the output yourself
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	for i, output := range response.Output {
		if response.OutputType != common.ReidentifySourceTypeBase64 {
			// output.ProcessedFile is a presigned URL: anyone holding it can download the
			// re-identified data, so fetch it here and don't log it.
			fmt.Println("Output", i, "of type", output.ProcessedFileType, "is available at a presigned URL")
			continue
		}
		decoded, decodeErr := base64.StdEncoding.DecodeString(output.ProcessedFile)
		if decodeErr != nil {
			fmt.Println("Could not decode the output:", decodeErr)
			return
		}
		// One file per output, named by its index, type and extension, e.g.
		// reidentified-notes-0-redacted_file.txt. The type and extension come from the
		// server, so only safe characters are kept: the name can't leave this directory.
		name := fmt.Sprintf("reidentified-notes-%d-%s.%s", i,
			namePart(string(output.ProcessedFileType), "output"), namePart(string(output.ProcessedFileExtension), "bin"))
		if writeErr := os.WriteFile(name, decoded, 0o600); writeErr != nil {
			fmt.Println("Could not write the output:", writeErr)
			return
		}
	}
	fmt.Println("Status:", response.Status, "Outputs:", len(response.Output))
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
