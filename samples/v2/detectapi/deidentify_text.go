/*
Copyright (c) 2022 Skyflow, Inc.
*/

package main

import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * This example demonstrates how to use the Skyflow Go SDK to deidentify sensitive data in text
 * by masking or transforming detected values according to your configuration.
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the skyflow client.
 * 3. Configure the vault with detect service.
 * 4. Deidentifying sensitive data in the text and returning the output.
 * 5. Handling the response and errors.
 */

func main() {
	// Step 1: Set up Skyflow vault credentials
	vaultConfig1 := common.VaultConfig{VaultId: "<VAULT_ID1>", ClusterId: "<CLUSTER_ID1>", Env: common.PROD, Credentials: common.Credentials{Token: "<BEARER_TOKEN1>"}}
	vaultConfig2 := common.VaultConfig{VaultId: "<VAULT_ID2>", ClusterId: "<CLUSTER_ID2>", Env: common.SANDBOX, Credentials: common.Credentials{Token: "<BEARER_TOKEN2>"}}
	var arr []common.VaultConfig
	arr = append(arr, vaultConfig2, vaultConfig1)

	// Step 2: Configure the skyflow client
	skyflowInstance, err := client.NewSkyflow(
		client.WithVaults(arr...),
		client.WithCredentials(common.Credentials{}), // Pass credentials if not provided in vault config
		client.WithLogLevel(logger.ERROR),            // Use LogLevel.ERROR in production
	)
	if err != nil {
		fmt.Println(*err)
	} else {
		// Step 3: Configure the vault with detect service.
		service, serviceErr := skyflowInstance.Detect("<VAULT_ID>") // Replace with your vault ID from the vault config
		if serviceErr != nil {
			fmt.Println(*serviceErr)
		} else {
			ctx := context.TODO()
			// Step 4: Deidentify sensitive data in the text and return the output
			deidentifyTextRes, deidentifyTextErr := service.DeidentifyText(ctx, common.DeidentifyTextRequest{
				Text: "My SSN is 123-45-6789 and my card is 4111 1111 1111 1111.",
				Entities: []common.DetectEntities{ // If not provided, all entities will be detected
					common.Ssn,
					common.CreditCard,
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
					VaultToken: []common.DetectEntities{ // Specify entities to use vault tokens
						common.CreditCard,
						common.Ssn,
						common.Name,
						common.CreditCardExpiration,
					},
					EntityUniqueCounter: []common.DetectEntities{
						common.Statistics,
					},
					EntityOnly: []common.DetectEntities{
						common.Dob,
					},
				},
			})
			// Step 5: Handle the response and errors.
			if deidentifyTextErr != nil {
				fmt.Println("ERROR: ", *deidentifyTextErr)

			} else {
				fmt.Println("RESPONSE: ", deidentifyTextRes)

			}
		}

	}
}
