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
 * This example de-identifies sensitive data in a string with the Detect API.
 *
 * The first call uses a stored configuration (created in Studio or through the REST API);
 * the second passes an inline configuration instead. Both fields are optional and sent as
 * given; with neither, the server decides whether the vault ID alone is enough.
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

	// ...or with an inline configuration: no stored configuration needed. Optional numbers
	// and flags in a configuration are pointers (common.Ptr), so nil is left to the server.
	configuration := &common.DetectConfiguration{
		Detect: &common.Detect{
			Entities: []common.Entity{
				{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TOKEN_GROUP_NAME>"},
				{
					EntityType:           common.EntityTypeDOB,
					DeidentificationType: common.DeidentificationTypeEntityUniqueCounter,
					Transformation:       &common.Transformation{ShiftDates: &common.ShiftDates{MinDays: common.Ptr(10), MaxDays: common.Ptr(30)}},
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
		// Location is nil when the server doesn't send one.
		location := "-"
		if entity.Location != nil {
			location = fmt.Sprintf("[%d,%d]", entity.Location.StartIndex, entity.Location.EndIndex)
		}
		fmt.Printf("%-16s %-18s %s %v\n", entity.EntityType, entity.Token, location, entity.EntityScores)
	}
	// Metrics and each count are pointers: nil when the server leaves them out.
	if m := response.Metrics; m != nil && m.WordCount != nil && m.CharacterCount != nil {
		fmt.Println("Words:", *m.WordCount, "Characters:", *m.CharacterCount)
	}
}
