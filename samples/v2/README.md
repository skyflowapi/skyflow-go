# SkyVault Go samples

Runnable examples for the Skyflow Go SDK (`github.com/skyflowapi/skyflow-go/v2`), one file per operation. For FlowVault vaults, see [samples/flowvault](../flowvault/README.md).

## Prerequisites

- Go 1.25+
- A Skyflow account. If you don't have one, register on the [Try Skyflow](https://skyflow.com/try-skyflow) page.
- A SkyVault vault and Skyflow credentials: a service-account `credentials.json`, an API key, or a bearer token.

### Create a vault

1. In a browser, sign in to Skyflow Studio.
2. Create a vault: **Create Vault** > **Start With a Template** > **Quickstart vault**.
3. Once the vault is ready, click the gear icon and select **Edit Vault Details**.
4. Note the **Vault ID** and **Vault URL**. The cluster ID is the subdomain of the vault URL: for `https://a1b2c3d4.vault.skyflowapis.com`, it is `a1b2c3d4`.

### Create a service account

1. In Skyflow Studio, in the side navigation: **IAM** > **Service Accounts** > **New Service Account**.
2. Name it and choose its **Roles**: **Vault Editor** for the vault samples, plus **Detect Invoker** for the Detect samples. Re-identifying vault tokens also needs a policy that grants the Detokenization action.
3. Click **Create**. Your browser downloads a `credentials.json`. Keep it secure: most samples need its path or contents.

## Configure

Each sample has placeholders near the top. Replace them with your own values:

```go
credentials := common.Credentials{Path: "<PATH_TO_YOUR_CREDENTIALS_JSON>"} // or ApiKey / Token / CredentialsString
vaultConfig := common.VaultConfig{
	VaultId:     "<YOUR_VAULT_ID>",
	ClusterId:   "<YOUR_CLUSTER_ID>", // from the vault URL: https://{clusterId}.vault.skyflowapis.com
	Env:         common.PROD,         // DEV, STAGE, SANDBOX or PROD
	Credentials: credentials,
}
```

## Run

Each file is its own program. Run one from this folder:

```bash
cd samples/v2
go run detectapi/deidentify_string.go
```

## Detect

| Sample | Demonstrates |
|---|---|
| [deidentify_string.go](detectapi/deidentify_string.go) | De-identify text with a stored configuration, or an inline one |
| [deidentify_file.go](detectapi/deidentify_file.go) | Submit a file and keep the run ID, or send a local file and wait for the run with `Poll` |
| [get_run.go](detectapi/get_run.go) | Look up a run by the ID `DeidentifyFile` returned |
| [reidentify_string.go](detectapi/reidentify_string.go) | De-identify text into vault tokens, then re-identify them |
| [reidentify_file.go](detectapi/reidentify_file.go) | Re-identify a base64 file and save the output |
| [check_guardrails.go](detectapi/check_guardrails.go) | Screen text for toxicity and denied topics |

Moving from the earlier Detect methods? See the [migration guide](../../docs/migrate_detect.md).

## Vault operations

| Sample | Demonstrates |
|---|---|
| [insert_records.go](vaultapi/insert_records.go) | Insert records |
| [insert_byot.go](vaultapi/insert_byot.go) | Insert records with your own tokens (BYOT) |
| [get_records.go](vaultapi/get_records.go) | Get records by Skyflow ID |
| [get_column_values.go](vaultapi/get_column_values.go) | Get records by column values |
| [update_record.go](vaultapi/update_record.go) | Update a record |
| [delete.go](vaultapi/delete.go) | Delete records |
| [detokenize.go](vaultapi/detokenize.go) | Detokenize tokens |
| [tokenize_records.go](vaultapi/tokenize_records.go) | Tokenize values |
| [query_record.go](vaultapi/query_record.go) | Query records with SQL |
| [upload_file.go](vaultapi/upload_file.go) | Upload a file to a file column |
| [invoke_connection.go](vaultapi/invoke_connection.go) | Invoke a connection |

## Service account (token generation)

| Sample | Demonstrates |
|---|---|
| [token_generation_example.go](serviceaccount/token_generation_example.go) | Bearer token from a credentials file or string |
| [token_generation_with_context.go](serviceaccount/token_generation_with_context.go) | Bearer token with a context |
| [scoped_token_generation.go](serviceaccount/scoped_token_generation.go) | Scoped token via role IDs |
| [signed_token_generation.go](serviceaccount/signed_token_generation.go) | Signed data tokens |
| [bearer_token_expiry_example.go](serviceaccount/bearer_token_expiry_example.go) | Retry a call when the bearer token expires mid-request |
