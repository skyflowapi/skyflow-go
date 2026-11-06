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
 * This example re-identifies vault tokens in a string with the Detect API.
 *
 * Only vault tokens can be re-identified: values de-identified with
 * DeidentificationTypeVaultToken are stored in a token group, so their tokens map back to
 * the original values. Placeholders such as [NAME_1] (ENTITY_UNIQUE_COUNTER) or masks are not
 * stored and can't be reversed. So this example first de-identifies the text into a token
 * group, then re-identifies the result.
 *
 * Without RedactionLevel every token is replaced by its original value. In a FlowVault
 * vault, a RedactionLevel selects tokens by TokenGroupName and renders them with a
 * RedactionPattern; both are configured in the vault.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. De-identify the text into vault tokens.
 * 4. Re-identify the tokens, in full and through a redaction pattern.
 * 5. Handle the response and errors.
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

	// Step 3: De-identify names into vault tokens in a token group. A stored configuration
	// (ConfigurationId) that tokenizes into a token group works the same way.
	deidentified, err := service.DeidentifyString(context.Background(), common.DeidentifyStringRequest{
		Text: "My name is John Doe.",
		Configuration: &common.DetectConfiguration{
			Detect: &common.Detect{
				Entities: []common.Entity{{
					EntityType:           common.EntityTypeName,
					DeidentificationType: common.DeidentificationTypeVaultToken,
					Destination:          "<TOKEN_GROUP_NAME>",
				}},
			},
		},
	}, common.DeidentifyStringOptions{})
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	tokenized := deidentified.ProcessedText // the name is replaced by a vault token

	// Step 4: Re-identify the tokens. With no RedactionLevel, the original values come back.
	plain, err := service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{
		Text: tokenized,
	}, common.ReidentifyStringOptions{})
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println(plain.ProcessedText) // My name is John Doe.

	// ...or render the token group's values through a redaction pattern configured in the vault
	redacted, err := service.ReidentifyString(context.Background(), common.ReidentifyStringRequest{
		Text: tokenized,
		RedactionLevel: []common.RedactionRule{common.RedactionLevel{
			TokenGroupName:   "<TOKEN_GROUP_NAME>",
			RedactionPattern: "<REDACTION_PATTERN_NAME>",
		}},
	}, common.ReidentifyStringOptions{})

	// Step 5: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println(redacted.ProcessedText) // the name as <REDACTION_PATTERN_NAME> renders it
}
