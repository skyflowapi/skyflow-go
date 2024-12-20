package vaultapi

import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func main() {
	vaultConfig1 := common.VaultConfig{VaultId: "<VAULT_ID1>", ClusterId: "<CLUSTER_ID1>", Env: common.DEV, Credentials: common.Credentials{Token: "<BEARER_TOKEN1>"}}
	vaultConfig2 := common.VaultConfig{VaultId: "<VAULT_ID2>", ClusterId: "<CLUSTER_ID2>", Env: common.DEV, Credentials: common.Credentials{Token: "<BEARER_TOKEN2>"}}
	var arr []common.VaultConfig
	arr = append(arr, vaultConfig2, vaultConfig1)
	skyflowInstance, err := client.NewSkyflow(
		client.WithVaults(arr...),
		client.WithCredentials(common.Credentials{}), // pass credentials if not provided in vault config
		client.WithLogLevel(logger.DEBUG),
	)
	if err != nil {
		fmt.Println(err)
	} else {
		service, serviceError := skyflowInstance.Vault("<VAULT_ID>")
		if serviceError != nil {
			fmt.Println(serviceError)
		} else {
			ctx := context.TODO()
			res, queryErr := service.Query(ctx, common.QueryRequest{Query: "SELECT * FROM demo WHERE skyflow_id='<ID>'"}) // Pass query here
			if queryErr != nil {
				fmt.Println("ERROR: ", *queryErr)
			} else {
				fmt.Println("RESPONSE: ", res)
			}
		}
	}

}
