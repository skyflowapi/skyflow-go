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
 * This example retrieves records from a FlowVault vault by Skyflow ID.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Get records.
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

	// Step 3: Get records by Skyflow ID. Set exactly one of SkyflowIds or UniqueValues.
	request := common.GetRequest{
		TableName:  "<SENSITIVE_DATA_TABLE>",
		SkyflowIds: []string{"<SKYFLOW_ID1>", "<SKYFLOW_ID2>"},
		// Columns:          []string{"name", "email"},                                        // return only these columns
		// ColumnRedactions: []common.ColumnRedaction{{ColumnName: "email", Redaction: "<REDACTION>"}},
		// Limit: 25, Offset: 0,
	}
	// To look records up by a unique column instead of by Skyflow ID:
	// request := common.GetRequest{TableName: "<SENSITIVE_DATA_TABLE>", UniqueValues: []map[string]interface{}{{"email": "john@example.com"}}}
	//
	// To get from several tables in one call, use Records instead of the fields above:
	// request := common.GetRequest{Records: []common.GetRequestRecord{
	// 	{TableName: "<TABLE_1>", SkyflowIds: []string{"<SKYFLOW_ID1>"}},
	// 	{TableName: "<TABLE_2>", UniqueValues: []map[string]interface{}{{"email": "john@example.com"}}},
	// }}
	response, err := vault.Get(context.Background(), request, common.GetOptions{})

	// Step 4: Handle the response and errors.
	// One record per input, in order. Each has SkyflowId, TableName, Data, Tokens, HashedData,
	// HttpCode, Error and RequestId.
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	for _, record := range response.Records {
		if record.Error == "" {
			fmt.Printf("%s -> data=%v\n", record.SkyflowId, record.Data)
		} else {
			fmt.Printf("failed (%d): %s [requestId=%s]\n", record.HttpCode, record.Error, record.RequestId)
		}
	}
}
