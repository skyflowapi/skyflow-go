package vaultapi

import (
	"context"
	"fmt"
	. "skyflow-go/v2/client"
	. "skyflow-go/v2/utils/common"
	utils "skyflow-go/v2/utils/common"
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
	values := make([]map[string]interface{}, 0)
	values = append(values, map[string]interface{}{
		"<FIELD_NAME1_1>": "<VALUE_1>",
	})
	values = append(values, map[string]interface{}{
		"<FIELD_NAME_2>": "<VALUE_1>",
		"<FIELD_NAME_3>": "<VALUE_2>",
	})

	tokens := make([]map[string]interface{}, 0)
	tokens = append(values, map[string]interface{}{
		"<FIELD_NAME1_1>": "<TOKEN>",
	})
	tokens = append(tokens, map[string]interface{}{
		"<FIELD_NAME_2>": "<TOKEN>",
	})

	insert, err4 := service.Insert(ctx, utils.InsertRequest{
		Table:  "<TABLE_NAME>",
		Values: values,
	}, utils.InsertOptions{ContinueOnError: false, ReturnTokens: true, TokenMode: ENABLE, Tokens: tokens})
	if err4 != nil {
		fmt.Println("ERROR: ", *err4)
	} else {
		fmt.Println("RESPONSE", insert)
	}
}
