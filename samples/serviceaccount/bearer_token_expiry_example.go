package serviceaccount

import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
//   * This example demonstrates how to configure and use the Skyflow SDK
//   * to detokenize sensitive data stored in a Skyflow vault.
//   * It includes setting up credentials, configuring the vault, and
//   * making a detokenization request. The code also implements a retry
//   * mechanism to handle unauthorized access errors (HTTP 401).

func DetokenizeData(skyflowClient, vaultID) error.SkyflowError {
	service, serviceError := skyflowInstance.Vault(vaultID)
	if serviceError != nil {
		fmt.Println(serviceError)
		return serviceError
	} else {
		ctx := context.TODO()
		//  Creating a list of tokens to be detokenized
		detokenizeData := []common.DetokenizeData{
			{
				Token:         "<TOKEN1>",
				RedactionType: common.REDACTED,
			},
			{
				Token:         "<TOKEN2>",
				RedactionType: common.MASKED,
			},
		}
		//  Building a detokenization request
		req := common.DetokenizeRequest{DetokenizeData: detokenizeData}
		//  Sending the detokenization request and receiving the response
		res, errDetokenize := service.Detokenize(ctx, req, common.DetokenizeOptions{
			ContinueOnError: true,
		})
		if errDetokenize != nil {
			
			fmt.Println("Unexpected error occurred: ", errDetokenize)
			return errDetokenize
		} else {
			// Printing the detokenized response
			fmt.Println("Skyflow error occurred: ", res)
		}
	}
	return nil
}
func main(){
	// Setting up credentials for accessing the Skyflow vault
	//  Credentials string for authentication
	credentials := common.Credentials{
		CredentialsString: "<STRINGIFIED_JSON_VALUE>", 
	}

	// Configuring the Skyflow vault with necessary details
	privaryVaultConfig := common.VaultConfig{
		VaultId:     "<YOUR_VAULT_ID1>",
		ClusterId:   "<YOUR_CLUSTER_ID1>",
		Env:         common.DEV,
		Credentials: credentials,
	}
	// Create a new Skyflow client
	skyflowClient, err := client.NewSkyflow(
		client.WithVaults(privaryVaultConfig),
		client.WithCredentials(credentials),
		client.WithLogLevel(logger.DEBUG),
	)
	if err != nil {
		fmt.Println("Error creating Skyflow client:", err)
		return
	}
	//  Attempting to detokenize data using the Skyflow client
	err = DetokenizeData(skyflowClient, "<VAULT_ID>")
	if err != nil {
		fmt.Println("Error detokenizing data:", err)
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
