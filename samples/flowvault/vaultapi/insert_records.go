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
 * This example inserts records into a FlowVault vault and prints the tokens returned for each.
 *
 * The table name and upsert are set at exactly ONE level: on the request (applies to every
 * record) or on every record. Setting them on both, or mixing the two, is a validation error.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Insert records.
 * 4. Handle the response and errors.
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

	// Step 2: Configure the Skyflow client
	skyflowClient, err := client.NewSkyflow(client.WithVaults(vaultConfig), client.WithLogLevel(logger.ERROR))
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	vault, err := skyflowClient.Vault(vaultConfig.VaultId)
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}

	// Step 3: Insert records. The table name and upsert here apply to every record.
	request := common.InsertRequest{
		TableName: "<SENSITIVE_DATA_TABLE>",
		Upsert:    &common.UpsertOptions{UniqueColumns: []string{"email"}, UpdateType: common.UpdateTypeUpdate},
		Records: []common.InsertRequestRecord{
			{Data: map[string]interface{}{"name": "John Doe", "email": "john@example.com"}},
		},
	}
	// To set the table (and upsert) per record instead, for example to insert into several
	// tables in one call, leave TableName and Upsert empty on the request:
	//
	// request := common.InsertRequest{
	// 	Records: []common.InsertRequestRecord{
	// 		{TableName: "<TABLE_1>", Data: map[string]interface{}{"name": "Jane Doe"}},
	// 		{TableName: "<TABLE_2>", Data: map[string]interface{}{"email": "jane@example.com"},
	// 			Upsert: &common.UpsertOptions{UniqueColumns: []string{"email"}, UpdateType: common.UpdateTypeReplace}},
	// 	},
	// }
	response, err := vault.Insert(context.Background(), request, common.InsertOptions{})

	// Step 4: Handle the response and errors.
	// Validation errors and whole-call failures come back as err. Otherwise there is one record
	// per input, in order; a record that failed has Error, HttpCode and RequestId set.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error == "" {
			fmt.Printf("%s -> tokens=%v\n", record.SkyflowId, record.Tokens)
		} else {
			fmt.Printf("failed (%d): %s [requestId=%s]\n", record.HttpCode, record.Error, record.RequestId)
		}
	}
}
