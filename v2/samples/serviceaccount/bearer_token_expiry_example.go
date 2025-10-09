package main

import (
	"context"
	"fmt"

	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	errors "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * This example demonstrates how to use the Skyflow Go SDK to detokenize sensitive data stored a Skyflow vault.
 * by converting tokens back into their original sensitive data values. It also retries detokenization on unauthorized access errors (HTTP 401).
 * <p>
 * Steps include:
 * 1. Set up Skyflow vault credentials.
 * 2. Configure the vault.
 * 3. Create a new Skyflow client.
 * 4. Detokenizing records by providing tokens and receiving original values.
 * 5. Handling the response and errors and retrying on token expiry.
 */

func DetokenizeData(skyflowClient *client.Skyflow, vaultID string) *errors.SkyflowError {
	service, serviceError := skyflowClient.Vault(vaultID)

	if serviceError != nil {
		fmt.Println(*serviceError)
	}

	detokenizeRes := &common.DetokenizeResponse{}
	errDetokenize := &errors.SkyflowError{}

	if serviceError != nil {
		fmt.Println(*serviceError)
	} else {
		ctx := context.TODO()
		detokenizeDataReq := common.DetokenizeRequest{DetokenizeData: []common.DetokenizeData{
			{
				Token:         "<TOKEN1>",
				RedactionType: common.PLAIN_TEXT,
			},
			{
				Token:         "<TOKEN2>",
				RedactionType: common.PLAIN_TEXT,
			},
		}}
		// Step 4: Detokenize records by providing tokens and receiving original values
		detokenizeRes, errDetokenize = service.Detokenize(ctx, detokenizeDataReq, common.DetokenizeOptions{
			ContinueOnError: true,
		})
		fmt.Println("RESPONSE: ", detokenizeRes)
	}

	return errDetokenize
}

func main() {
	// Step 1: Set up Skyflow vault credentials.
	credentials := common.Credentials{
		CredentialsString: "<STRINGIFIED_JSON_VALUE>",
	}

	// Step 2: Configure the vault.
	privaryVaultConfig := common.VaultConfig{
		VaultId:     "<YOUR_VAULT_ID1>",
		ClusterId:   "<YOUR_CLUSTER_ID1>",
		Env:         common.DEV,
		Credentials: credentials,
	}
	// Step 3: Create a new Skyflow client
	skyflowClient, err := client.NewSkyflow(
		client.WithVaults(privaryVaultConfig),
		client.WithCredentials(credentials),
		client.WithLogLevel(logger.ERROR),
	)

	if err != nil {
		fmt.Println("Error creating Skyflow client:", err)
		return
	}
	//  Attempting to detokenize data using the Skyflow client
	err = DetokenizeData(skyflowClient, "<VAULT_ID>")
	// Step 5: Handling the response and errors and retrying on token expiry.
	if err != nil {
		fmt.Println("Error detokenizing data:", *err)
		// Retry detokenization if the error is due to unauthorized access (HTTP 401)
		if err.GetCode() == "401" {
			fmt.Println("Unauthorized access. Retrying...")
			err2 := DetokenizeData(skyflowClient, "<VAULT_ID>")
			if err2 != nil {
				fmt.Println("Error detokenizing data on retry:", err2)
			} else {
				fmt.Println("Detokenization successful on retry")
			}
		}
		return
	} else {
		fmt.Println("Detokenization successful")
	}
}
