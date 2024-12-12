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
	res, err := service.Detokenize(ctx, DetokenizeRequest{
		Tokens:        []string{"<TOKEN>", "<TOKEN>"},
		RedactionType: REDACTED,
	}, DetokenizeOptions{
		ContinueOnError: true,
	})
	if err != nil {
		fmt.Println("ERROR: ", err)
	} else {
		fmt.Println("RESPONSE: ", res)
	}
}
