/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example sends custom headers with one call, using the CustomHeaders option.
 *
 * Every operation's options take CustomHeaders, a map of the headers to send with that call.
 * The supported keys are common.SkyflowAccountId, common.SkyflowAccountName and
 * common.RequestIdHeader.
 *
 * Steps:
 * 1. Set up credentials and the vault config.
 * 2. Configure the Skyflow client.
 * 3. Insert a record with a request ID header.
 * 4. Handle the response and errors.
 */
func main() {
	// Step 1: Set up credentials and the vault config
	vaultConfig := common.VaultConfig{
		VaultId:     "<VAULT_ID>",
		ClusterId:   "<CLUSTER_ID>",
		Env:         common.DEV,
		Credentials: common.Credentials{Token: "<BEARER_TOKEN>"},
	}

	// Step 2: Configure the Skyflow client
	skyflowClient, err := client.NewSkyflow(client.WithVaults(vaultConfig), client.WithLogLevel(logger.DEBUG))
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), "requestId:", err.GetRequestId(), err.GetDetails())
		return
	}
	vault, err := skyflowClient.Vault(vaultConfig.VaultId)
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), "requestId:", err.GetRequestId(), err.GetDetails())
		return
	}

	// Step 3: Insert a record with a request ID header
	request := common.InsertRequest{
		TableName: "<TABLE_NAME>",
		Upsert:    &common.UpsertOptions{UniqueColumns: []string{"<UPSERT_COLUMN_NAME>"}},
		Records:   []common.InsertRequestRecord{{Data: map[string]interface{}{"<YOUR_COLUMN_NAME>": "<YOUR_VALUE>"}}},
	}
	requestID := makeRequestID()
	fmt.Println("request id =>", requestID)
	options := common.InsertOptions{
		CustomHeaders: map[common.CustomHeaderKey]string{common.RequestIdHeader: requestID},
	}
	response, err := vault.Insert(context.Background(), request, options)

	// Step 4: Handle the response and errors
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), "requestId:", err.GetRequestId(), err.GetDetails())
		return
	}
	fmt.Println("Records:", response.Records)
}

// makeRequestID returns a random ID to trace this call.
func makeRequestID() string {
	id := make([]byte, 16)
	_, _ = rand.Read(id)
	return hex.EncodeToString(id)
}
