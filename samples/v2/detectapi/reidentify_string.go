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
 * This example re-identifies vault tokens in a string with the Detect API.
 *
 * Only vault tokens can be re-identified: values de-identified with
 * DeidentificationTypeVaultToken are stored in the vault, so their tokens map back to the
 * original values. Placeholders such as [NAME_1] (ENTITY_UNIQUE_COUNTER) or masks are not
 * stored and can't be reversed. So this example first de-identifies the text into vault
 * tokens, then re-identifies the result.
 *
 * Without RedactionLevel every token is replaced by its original value. In a SkyVault
 * vault, a RedactionLevel selects tokens by EntityName and shows them through a
 * RedactionType.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. De-identify the text into vault tokens.
 * 4. Re-identify the tokens.
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

	// Step 3: De-identify the name and email into vault tokens. A stored configuration
	// (ConfigurationId) that tokenizes into the vault works the same way.
	deidentified, err := service.DeidentifyString(context.Background(), common.DeidentifyStringRequest{
		Text: "My name is John Doe, and my email is johndoe@acme.com.",
		Configuration: &common.DetectConfiguration{
			Detect: &common.Detect{
				Entities: []common.Entity{
					{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TABLE_NAME>.<COLUMN_NAME>"},
					{EntityType: common.EntityTypeEmailAddress, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TABLE_NAME>.<COLUMN_NAME>"},
				},
			},
		},
	}, common.DeidentifyStringOptions{})
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}

	// Step 4: Re-identify the tokens: names in plain text, emails masked
	response, err := service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{
		Text: deidentified.ProcessedText,
		RedactionLevel: []common.RedactionRule{
			common.RedactionLevel{EntityName: common.EntityTypeName, RedactionType: common.PLAIN_TEXT},
			common.RedactionLevel{EntityName: common.EntityTypeEmailAddress, RedactionType: common.MASKED},
		},
	}, common.ReidentifyStringOptions{})

	// Step 5: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println(response.ProcessedText) // My name is John Doe, and my email is the masked address.
}
