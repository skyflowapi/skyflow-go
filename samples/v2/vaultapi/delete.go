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
		service, serviceErr := skyflowInstance.Vault("<VAULT_ID>")
		if serviceErr != nil {
			fmt.Println(serviceErr)
		} else {
			ctx := context.TODO()
			res, deleteErr := service.Delete(ctx, common.DeleteRequest{
				Table: "<TABLE_NAME>",
				Ids:   []string{"<SKYFLOW_ID>"},
			})
			if deleteErr != nil {
				fmt.Println(deleteErr)
			} else {
				fmt.Println(res)
			}
		}
	}

}
