# FlowVault Go samples

Runnable examples for the Skyflow FlowVault Go SDK (`github.com/skyflowapi/skyflow-go/flowvault`), one file per operation.

## Prerequisites

- Go 1.25+
- A Skyflow account. If you don't have one, register on the [Try Skyflow](https://skyflow.com/try-skyflow) page.
- A FlowVault vault and Skyflow credentials: a service-account `credentials.json`, an API key, or a bearer token.

### Find the vault details

In Skyflow Studio, open the vault, click the gear icon and select **Edit Vault Details**. Note the **Vault ID** and **Vault URL**. The cluster ID is the subdomain of the vault URL: for `https://a1b2c3d4.vault.skyflowapis.com`, it is `a1b2c3d4`.

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

## Detect

| Sample | Demonstrates |
|---|---|
| [deidentify_string.go](detectapi/deidentify_string.go) | De-identify text with a stored configuration, or an inline one |
| [deidentify_file.go](detectapi/deidentify_file.go) | Submit a file by presigned URL and keep the run ID, or send a local file and wait for the run |
| [deidentify_file_advanced.go](detectapi/deidentify_file_advanced.go) | `FileObject`, media and object-entity options, and a per-call `Interceptor` |
| [get_run.go](detectapi/get_run.go) | Look up a run by ID, save its base64 outputs, read `Message` when it failed |
| [reidentify_string.go](detectapi/reidentify_string.go) | Tokenize text into a token group, then re-identify it in full or through a redaction pattern |
| [reidentify_file.go](detectapi/reidentify_file.go) | Re-identify a base64 file and save the output |
| [round_trip.go](detectapi/round_trip.go) | De-identify a local file into vault tokens, then re-identify its redacted output |
| [check_guardrails.go](detectapi/check_guardrails.go) | Screen text for toxicity and denied topics |

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
