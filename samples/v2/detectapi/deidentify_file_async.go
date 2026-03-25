/*
Copyright (c) 2022 Skyflow, Inc.
*/

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * This example demonstrates how to use the Skyflow Go SDK to deidentify sensitive data in a file
 * asynchronously by masking or transforming detected values according to your configuration.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault with detect service.
 * 4. Deidentify sensitive data in the file and save the output by configuring detection options.
 * 5. Handle the response and errors.
 */

type DeidentifyFileResult struct {
	Response *common.DeidentifyFileResponse
	Error    error
}

func main() {
	// Step 1: Set up Skyflow vault credentials.
	vaultConfig1 := common.VaultConfig{VaultId: "<VAULT_ID1>", ClusterId: "<CLUSTER_ID1>", Env: common.PROD, Credentials: common.Credentials{Token: "<BEARER_TOKEN1>"}}
	vaultConfig2 := common.VaultConfig{VaultId: "<VAULT_ID2>", ClusterId: "<CLUSTER_ID2>", Env: common.SANDBOX, Credentials: common.Credentials{Token: "<BEARER_TOKEN2>"}}
	var arr []common.VaultConfig
	arr = append(arr, vaultConfig2, vaultConfig1)

	// Step 2: Configure the skyflow client.
	skyflowInstance, err := client.NewSkyflow(
		client.WithVaults(arr...),
		client.WithCredentials(common.Credentials{}), // Pass credentials if not provided in vault config
		client.WithLogLevel(logger.ERROR), // Use LogLevel.ERROR in production
	)
	if err != nil {
		fmt.Println(*err)
	} else {
		// Step 3: Configure the vault with detect service
		service, serviceErr := skyflowInstance.Detect("<VAULT_ID>") // Replace with your vault ID from the vault config
		if serviceErr != nil {
			fmt.Println(*serviceErr)
		} else {
			filePath := "<PATH_TO_FILE>" // Replace with your file path to deidentify
			outputDir := "<OUTPUT_DIRECTORY_PATH>" // Replace with your directory path to save the output file

			resultChan := make(chan DeidentifyFileResult, 1)

			go func() {
				defer close(resultChan)

				ctx := context.TODO()
				file, fileErr := os.Open(filePath)
				if fileErr != nil {
					resultChan <- DeidentifyFileResult{
						Response: nil,
						Error:    fmt.Errorf("Failed to open file: %v", fileErr),
					}
					return
				}
				defer file.Close()

				// Deidentify sensitive data in the file
				deidentifyFileRes, deidentifyFileErr := service.DeidentifyFile(ctx, common.DeidentifyFileRequest{
					File: common.FileInput{
						File: file,
						// FilePath: filePath, // Provide FilePath or File at a time
					},
					OutputDirectory: outputDir,
					WaitTime:        20,
					Entities: []common.DetectEntities{
						common.Ssn,
						common.CreditCard,
						common.EmailAddress,
					},
					AllowRegexList:    []string{"<ALLOW_REGEX_PATTERN1>", "<ALLOW_REGEX_PATTERN2>"},       // Replace with the regex patterns you want to allow during deidentification
					RestrictRegexList: []string{"<RESTRICT_REGEX_PATTERN1>", "<RESTRICT_REGEX_PATTERN2>"}, // Replace with the regex patterns you want to restrict during deidentification
					// Transformations: common.Transformations{
					// 	ShiftDates: common.DateTransformation{
					// 		MaxDays: 15, // Maximum days to shift
					// 		MinDays: 5,  //  Minimum days to shift
					// 		Entities: []common.TransformationsShiftDatesEntityTypesItem{ // Apply shift to DOB entities
					// 			common.TransformationsShiftDatesEntityTypesItemDob,
					// 		},
					// 	},
					// },
					TokenFormat: common.TokenFormat{
						DefaultType: common.TokenTypeDefaultEntityOnly,
						EntityUniqueCounter: []common.DetectEntities{
							common.Statistics,
						},
						EntityOnly: []common.DetectEntities{
							common.Dob,
						},
					},
					// Image file related options
					MaskingMethod:        common.BLACKBOX,
					OutputProcessedImage: true,
					OutputOcrText:        true,
					// PDF file related options
					PixelDensity:  30,
					MaxResolution: 3,
					// Audio file related options
					// OutputProcessedAudio: true,
					Bleep: common.AudioBleep{
						Gain:         70,
						Frequency:    100,
						StartPadding: 2,
						StopPadding:  8,
					},
					// OutputTranscription: common.PLAINTEXT_TRANSCRIPTION,
				})

				// Send the result through the channel
				if deidentifyFileErr != nil {
					resultChan <- DeidentifyFileResult{
						Response: nil,
						Error:    fmt.Errorf("Deidentify error: %v", *deidentifyFileErr),
					}
				} else {
					resultChan <- DeidentifyFileResult{
						Response: deidentifyFileRes,
						Error:    nil,
					}
				}
			}()

			// Handle the response and errors.
			result := <-resultChan

			if result.Error != nil {
				fmt.Println("ERROR: ", result.Error)
			} else if result.Response != nil {
				fmt.Println("RESPONSE: ", result.Response)
			} else {
				fmt.Println("Unexpected result: both error and response are nil")
			}
		}
	}
}
