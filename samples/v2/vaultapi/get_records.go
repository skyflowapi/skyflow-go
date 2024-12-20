package vaultapi

import (
	"context"
	"fmt"
	. "skyflow-go/v2/client"
	. "skyflow-go/v2/utils/common"
	"skyflow-go/v2/utils/logger"
)

func main() {
	vaultConfig1 := VaultConfig{VaultId: "<VAULT_ID>", ClusterId: "<CLUSTER_ID>", Env: DEV, Credentials: Credentials{Token: "<BEARER_TOKEN>"}}
	skyflow1 := Skyflow{}
	client1, err := skyflow1.Builder().WithVaultConfig(vaultConfig1).WithLogLevel(logger.DEBUG).Build()
	if err != nil {
		fmt.Println(err)
	}
	service, _ := client1.Vault("<VAULT_ID>")
	ctx := context.TODO()
	res, getErr := service.Get(ctx, GetRequest{
		Table: "<TABLE_NAME>",
		Ids: []string{
			"<SKYFLOW_ID_1>",
		},
	}, GetOptions{
		ReturnTokens: true,
	})
	if getErr != nil {
		fmt.Println("ERROR: ", getErr)
	} else {
		fmt.Println("RESPONSE:", res.Data)
	}
}
