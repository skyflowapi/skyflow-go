/*
Copyright (c) 2026 Skyflow, Inc.
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
 * This example de-identifies sensitive data in a string with the Detect V2 API.
 *
 * The first call uses a stored configuration (created in Studio or through the REST API);
 * the second passes an inline configuration instead. Set exactly one of the two.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. De-identify with a stored configuration, then with an inline one.
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

	// Step 3: De-identify with a stored configuration
	text := "My name is John Doe, and my email is johndoe@acme.com."
	response, err := service.DeidentifyString(context.Background(), common.DeidentifyStringRequest{
		Text:            text,
		ConfigurationId: "<CONFIGURATION_ID>",
	}, common.DeidentifyStringOptions{})
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println(response.ProcessedText) // My name is [NAME_1], and my email is [EMAIL_ADDRESS_1].

	// ...or with an inline configuration: no stored configuration needed
	configuration := &common.DetectConfiguration{
		Detect: &common.Detect{
			Entities: []common.Entity{
				{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TABLE_NAME>.<COLUMN_NAME>"},
				{
					EntityType:           common.EntityTypeDOB,
					DeidentificationType: common.DeidentificationTypeEntityUniqueCounter,
					Transformation:       &common.Transformation{ShiftDates: &common.ShiftDates{MinDays: 10, MaxDays: 30}},
				},
				{EntityType: common.EntityTypeAll, DeidentificationType: common.DeidentificationTypeEntityUniqueCounter},
			},
			Restrict:       []string{"[0-9]{3}-[0-9]{2}-[0-9]{4}"},
			Skip:           []string{"Skyflow"},
			CustomTypes:    []common.CustomType{{Label: "EMPLOYEE_ID", Match: []string{"EMP-[0-9]{6}"}}},
			ReturnEntities: common.ReturnEntitiesTypeExcludeSensitiveData,
		},
	}
	response, err = service.DeidentifyString(context.Background(), common.DeidentifyStringRequest{
		Text:          text,
		Configuration: configuration,
	}, common.DeidentifyStringOptions{})

	// Step 4: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	for _, entity := range response.Entities {
		fmt.Printf("%-16s %-18s [%d,%d] %v\n", entity.EntityType, entity.Token,
			entity.Location.StartIndex, entity.Location.EndIndex, entity.EntityScores)
	}
	fmt.Println("Words:", response.Metrics.WordCount, "Characters:", response.Metrics.CharacterCount)
}
