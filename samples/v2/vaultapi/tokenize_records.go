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
	var reqArray []TokenizeRequest
	reqArray = append(reqArray, TokenizeRequest{
		ColumnGroup: "<COLUMN_GROUP_NAME>",
		Value:       "<VALUE>",
	})
	res, tokenizeErr := service.Tokenize(ctx, reqArray)
	if tokenizeErr != nil {
		fmt.Println("ERROR: ", tokenizeErr)
	} else {
		fmt.Println("RESPONSE: ", res)
	}
}
