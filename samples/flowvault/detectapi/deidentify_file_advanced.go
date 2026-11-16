/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example de-identifies an open file with the Detect API, using an inline
 * configuration with media options and per-call custom headers.
 *
 * - FileObject sends a file you already opened (you close it); its format is inferred from
 *   the name. FilePath and Value (with DataSource) are the other two sources: set exactly one.
 * - Media holds per-media-type options. Its numbers and flags are pointers (common.Ptr): nil
 *   leaves the option to the server default.
 * - ObjectEntities redact objects such as faces in images.
 * - CustomHeaders adds headers to this call only.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Build the configuration and the per-call options.
 * 4. De-identify the file and wait for the run.
 * 5. Handle the response and errors.
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

	// Step 3: Build the configuration and the per-call options
	configuration := &common.DetectConfiguration{
		Detect: &common.Detect{
			Entities: []common.Entity{
				{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TOKEN_GROUP_NAME>"},
				{EntityType: common.EntityTypeAll, DeidentificationType: common.DeidentificationTypeEntityUniqueCounter},
			},
			ObjectEntities: []common.ObjectEntity{
				{EntityType: common.ObjectEntityTypeFace, DeidentificationType: common.ObjectDeidentificationTypeRedact},
			},
			ReturnEntities: common.ReturnEntitiesTypeExcludeSensitiveData,
		},
		Media: &common.Media{
			Document: &common.Document{Pdf: &common.Pdf{ProcessingMode: common.PdfProcessingModeOCR, Density: common.Ptr(300)}},
			Image:    &common.Image{OutputProcessedImage: common.Ptr(true), MaskingMethod: common.MaskingMethodBlackbox},
		},
	}
	requestID := makeRequestID()
	fmt.Println("request id =>", requestID)
	options := common.DeidentifyFileOptions{
		Poll:            &common.PollOptions{WaitTime: common.Ptr(3 * time.Minute), MaxAttempts: common.Ptr(10)},
		OutputDirectory: "<OUTPUT_DIRECTORY>",
		CustomHeaders:   map[common.CustomHeaderKey]string{common.RequestIdHeader: requestID},
	}

	// Step 4: De-identify the open file and wait for the run
	file, openErr := os.Open("<PATH_TO_FILE>/scanned-intake.pdf")
	if openErr != nil {
		fmt.Println("Could not open the file:", openErr)
		return
	}
	defer file.Close()
	response, err := service.DeidentifyFile(context.Background(), common.DeidentifyFileRequest{
		FileObject:    file,
		Configuration: configuration,
	}, options)

	// Step 5: Handle the response and errors. On an error the response can still hold the run
	// ID (for example when a lookup failed during the wait), so check it before giving up.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		if response != nil {
			fmt.Println("Run ID:", response.RunId) // look it up later with GetRun
		}
		return
	}
	switch response.Status {
	case common.DetectRunStatusSuccess: // outputs are written to OutputDirectory
		for _, output := range response.Output {
			fmt.Println("Output:", output.ProcessedFileType, output.ProcessedFileExtension)
		}
		if response.Metrics != nil && response.Metrics.Pages != nil {
			fmt.Println("Pages:", *response.Metrics.Pages)
		}
	case common.DetectRunStatusInProgress, common.DetectRunStatusQueued: // the wait ran out: poll later with GetRun
		fmt.Println("Still running, run ID:", response.RunId)
	default: // FAILED or UNKNOWN
		fmt.Println("Run failed:", response.Message)
	}
}

// makeRequestID returns a random ID to trace this call.
func makeRequestID() string {
	id := make([]byte, 16)
	_, _ = rand.Read(id)
	return hex.EncodeToString(id)
}
