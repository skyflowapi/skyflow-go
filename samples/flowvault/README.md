# FlowVault Go samples

Runnable examples for the Skyflow FlowVault Go SDK (`github.com/skyflowapi/skyflow-go/flowvault`), one file per operation.

## Prerequisites

- Go 1.25+
- A FlowVault vault and Skyflow credentials: a service-account `credentials.json`, an API key, or a bearer token.

## Configure

Each sample has placeholders near the top. Replace them with your own values:

```go
credentials := common.Credentials{Path: "<PATH_TO_YOUR_CREDENTIALS_JSON>"} // or ApiKey / Token / CredentialsString
vaultConfig := common.VaultConfig{
	VaultId:     "<YOUR_VAULT_ID>",
	ClusterId:   "<YOUR_CLUSTER_ID>", // from the vault URL: https://{clusterId}.skyvault.skyflowapis.com
	Env:         common.PROD,         // DEV, STAGE, SANDBOX or PROD
	Credentials: credentials,
}
```

## Run

Each file is its own program. Run one from this folder:

```bash
cd samples/flowvault
go run vaultapi/insert_records.go
```

## Vault operations

| Sample | Demonstrates |
|---|---|
| [insert_records.go](vaultapi/insert_records.go) | Insert records (table name and upsert on the request, or on each record) |
| [get_records.go](vaultapi/get_records.go) | Get records by Skyflow ID, by unique value, or from several tables |
| [update_record.go](vaultapi/update_record.go) | Update a record (`UPDATE` or `REPLACE`) |
| [delete_records.go](vaultapi/delete_records.go) | Delete records |
| [detokenize_records.go](vaultapi/detokenize_records.go) | Detokenize tokens |
| [query_records.go](vaultapi/query_records.go) | Run a read-only SQL `SELECT` query |
| [get_tokens.go](vaultapi/get_tokens.go) | Look up existing tokens for values (deterministic token groups) |
| [upload_files.go](vaultapi/upload_files.go) | Upload files to file columns (`FilePath` / `Base64` / `FileObject`) |
| [delete_files.go](vaultapi/delete_files.go) | Delete files from file columns (by `SkyflowId` or `UniqueValues`) |

## Custom headers and HTTP config

| Sample | Demonstrates |
|---|---|
| [custom_header_example.go](vaultapi/custom_header_example.go) | Add headers to one call with an `InsertOptions` interceptor |
| [timeout_and_retry_config_example.go](vaultapi/timeout_and_retry_config_example.go) | Client-wide `WithHttpConfig` and per-vault `Timeout` / `MaxRetries` |

## Service account (token generation)

| Sample | Demonstrates |
|---|---|
| [bearer_token_generation_example.go](serviceaccount/bearer_token_generation_example.go) | Bearer token from a credentials file path or a credentials string |
| [bearer_token_generation_with_context_example.go](serviceaccount/bearer_token_generation_with_context_example.go) | Context-aware bearer tokens (string context and object context) |
| [bearer_token_generation_using_goroutines_example.go](serviceaccount/bearer_token_generation_using_goroutines_example.go) | Concurrent token generation from goroutines with a shared context |
| [scoped_token_generation_example.go](serviceaccount/scoped_token_generation_example.go) | Scoped tokens via `RoleIds` |
| [signed_token_generation_example.go](serviceaccount/signed_token_generation_example.go) | Signed data tokens (`GenerateSignedDataTokens`) |
