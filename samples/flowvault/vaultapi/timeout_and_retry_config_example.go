/*
Copyright (c) 2026 Skyflow, Inc.
*/
package main

import (
	"fmt"
	"time"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

/**
 * This example sets HTTP timeout and retry settings, for the whole client and for one vault.
 *
 * Each setting is resolved separately: the vault's value, then the client's, then the SDK
 * default.
 *
 *   Timeout            overall limit for one call, including retries and backoff. Default 60s.
 *   MaxRetries         retries after the first attempt, for 408, 429 and 5xx. Default 0 (off).
 *   InitialRetryDelay  wait before the first retry. Default 500ms.
 *   MaxRetryDelay      the most the wait can grow to; it doubles each retry. Default 2000ms.
 *
 * A context deadline still applies when it is shorter. File uploads in UploadFiles are not
 * retried and have no overall timeout, so large files aren't cut off.
 *
 * Steps:
 * 1. Set up credentials and the vault config, with per-vault overrides.
 * 2. Configure the Skyflow client with client-wide defaults.
 */
func main() {
	// Step 1: Set up credentials and the vault config, with per-vault overrides
	vaultRetries := 2
	vaultConfig := common.VaultConfig{
		VaultId:     "<YOUR_VAULT_ID>",
		ClusterId:   "<YOUR_CLUSTER_ID>",
		Env:         common.PROD,
		Credentials: common.Credentials{Path: "<PATH_TO_YOUR_CREDENTIALS_JSON>"},
		// These win over the client-wide values below, for this vault only.
		Timeout:    30 * time.Second,
		MaxRetries: &vaultRetries,
	}

	// Step 2: Configure the Skyflow client with client-wide defaults.
	// Any vault that doesn't set a field uses these.
	clientRetries := 3
	skyflowClient, err := client.NewSkyflow(
		client.WithHttpConfig(common.HttpConfig{
			Timeout:           60 * time.Second,
			MaxRetries:        &clientRetries,
			InitialRetryDelay: 500 * time.Millisecond,
			MaxRetryDelay:     4 * time.Second,
		}),
		client.WithVaults(vaultConfig),
		client.WithLogLevel(logger.ERROR),
	)
	if err != nil {
		fmt.Println("Skyflow Specific Error:", err.GetHttpCode(), err.GetMessage(), err.GetDetails())
		return
	}
	fmt.Println("Skyflow client configured with custom timeout and retry settings:", skyflowClient != nil)

	// Use the client as usual; calls now fail at the configured timeouts and retry 408, 429 and
	// 5xx responses with exponential backoff and jitter.
	// vault, _ := skyflowClient.Vault(vaultConfig.VaultId)
	// response, err := vault.Detokenize(context.Background(), detokenizeRequest, common.DetokenizeOptions{})
}
