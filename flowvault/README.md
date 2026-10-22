# Skyflow Flow Vault Go SDK

> **The `flowvault` module isn't published yet.** Until it is, build against a local checkout of this repository, as described in [Add the module](#add-the-module).

Use the Skyflow Flow Vault Go SDK (`github.com/skyflowapi/skyflow-go/flowvault`) to work with a Skyflow Flow Vault from a Go backend. You can insert, read, update, and delete records, and run read-only SQL queries. You can also detokenize tokens, look up existing tokens, and upload and delete files in file columns. This guide covers installing the module, authenticating, configuring the client, each vault operation, and handling errors.

# Table of contents

- [Overview](#overview)
- [Install](#install)
  - [Requirements](#requirements)
  - [Add the module](#add-the-module)
- [Quickstart](#quickstart)
- [Authenticate](#authenticate)
  - [Credential types](#credential-types)
  - [Where to set credentials](#where-to-set-credentials)
  - [Generate tokens yourself](#generate-tokens-yourself)
- [Initialize the client](#initialize-the-client)
  - [Vault configuration](#vault-configuration)
  - [Client options](#client-options)
- [Vault operations](#vault-operations)
  - [Insert and upsert](#insert-and-upsert)
  - [Update](#update)
  - [Get](#get)
  - [Delete](#delete)
  - [Detokenize](#detokenize)
  - [Get tokens](#get-tokens)
  - [Query](#query)
  - [Upload files](#upload-files)
  - [Delete files](#delete-files)
- [Timeouts and retries](#timeouts-and-retries)
- [Custom request headers](#custom-request-headers)
- [Manage the client](#manage-the-client)
- [Logging](#logging)
- [Error handling](#error-handling)
  - [Whole-call errors](#whole-call-errors)
  - [Per-record results](#per-record-results)
  - [Retry failed records](#retry-failed-records)
- [Samples](#samples)

## Overview

This SDK is for Flow Vaults. A Flow Vault's URL uses the `skyvault` subdomain, for example `https://<CLUSTER_ID>.skyvault.skyflowapis.com`. For Skyflow's other vaults, use the v2 SDK described in the [root README](https://github.com/skyflowapi/skyflow-go/blob/main/README.md).

The SDK works like this:

- **One client, many vaults.** You create a `Skyflow` client, register one or more vaults, and get a `VaultService` for each vault. Every operation is a method on `VaultService`.
- **One request, one response.** Each operation takes a `context.Context`, a request struct, and an options struct. It returns a response or a `*SkyflowError`.
- **Results per record.** When some records in a call fail and others succeed, the call still returns a response. Each record in it carries its own error, HTTP code, and request ID. See [Error handling](#error-handling).
- **Retries are off by default.** You can turn them on for the client or for one vault. See [Timeouts and retries](#timeouts-and-retries).
- **Safe for concurrent use.** You can share one client across goroutines, and change its vaults and credentials while calls are running.

## Install

### Requirements

- Go 1.25.0 or later.
- A Flow Vault, with its vault ID and cluster ID.
- A service account credentials file, an API key, or a bearer token for the vault. See [Authenticate](#authenticate).

### Add the module

Make sure your project uses Go modules, then add the module:

```bash
go get github.com/skyflowapi/skyflow-go/flowvault
```

Until the module is published, point your `go.mod` at a local checkout of this repository instead, the way the [samples](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/go.mod) do:

```text
require github.com/skyflowapi/skyflow-go/flowvault v0.0.0

replace github.com/skyflowapi/skyflow-go/flowvault => <PATH_TO_SKYFLOW_GO_CHECKOUT>/flowvault
```

The module's packages are:

| Package | Use it for |
| --- | --- |
| `github.com/skyflowapi/skyflow-go/flowvault/client` | The `Skyflow` client and `VaultService` |
| `github.com/skyflowapi/skyflow-go/flowvault/utils/common` | Configuration, credential, request, option, and response types |
| `github.com/skyflowapi/skyflow-go/flowvault/utils/error` | `SkyflowError` |
| `github.com/skyflowapi/skyflow-go/flowvault/utils/logger` | Log levels |
| `github.com/skyflowapi/skyflow-go/flowvault/serviceaccount` | Generating bearer tokens and signed data tokens |

## Quickstart

This program registers one vault, inserts a record, and prints the tokens for the inserted record or the error for a failed one:

```go
package main

import (
	"context"
	"fmt"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

func main() {
	vaultConfig := common.VaultConfig{
		VaultId:     "<YOUR_VAULT_ID>",   // Replace with your vault ID
		ClusterId:   "<YOUR_CLUSTER_ID>", // Replace with the cluster ID from your vault URL
		Env:         common.PROD,
		Credentials: common.Credentials{Path: "<PATH_TO_YOUR_CREDENTIALS_JSON>"}, // Replace with your credentials file path
	}

	skyflowClient, err := client.NewSkyflow(client.WithVaults(vaultConfig))
	if err != nil {
		fmt.Println("client error:", err.GetMessage())
		return
	}
	vault, err := skyflowClient.Vault(vaultConfig.VaultId)
	if err != nil {
		fmt.Println("vault error:", err.GetMessage())
		return
	}

	request := common.InsertRequest{
		TableName: "<SENSITIVE_DATA_TABLE>", // Replace with a table in your vault
		Records: []common.InsertRequestRecord{
			{Data: map[string]interface{}{"name": "John Doe", "email": "john@example.com"}},
		},
	}
	response, err := vault.Insert(context.Background(), request, common.InsertOptions{})
	if err != nil {
		// The whole call failed, so there are no per-record results.
		fmt.Println("insert failed:", err.GetHttpCode(), err.GetMessage(), "request ID:", err.GetRequestId())
		return
	}
	for _, record := range response.Records {
		if record.Error != "" {
			fmt.Printf("record failed (%d): %s [request ID: %s]\n", record.HttpCode, record.Error, record.RequestId)
			continue
		}
		fmt.Println(record.SkyflowId, record.Tokens)
	}
}
```

## Authenticate

### Credential types

A `common.Credentials` value holds one of four credential types:

| Field | Credential | What the SDK sends |
| --- | --- | --- |
| `ApiKey` | An API key | The API key |
| `Token` | A bearer token you generated | The token as is. If the token has expired, the call returns a `SkyflowError` without contacting the vault |
| `Path` | The path to a service account credentials JSON file | A bearer token it generates from the file. It reuses the token until it expires, then generates a new one |
| `CredentialsString` | The contents of a service account credentials JSON file, as a string | A bearer token it generates from the string, reused the same way as for `Path` |

Set exactly one of these fields. `WithCredentials`, `AddSkyflowCredentials`, and `UpdateSkyflowCredentials` return a `SkyflowError` when the `Credentials` value sets more than one of them, or none.

With `Path` or `CredentialsString`, you can also shape the tokens the SDK generates:

- `Roles` limits each token to the listed role IDs.
- `Context` embeds a context value in each token.

```go
credentials := common.Credentials{
	Path:  "<PATH_TO_YOUR_CREDENTIALS_JSON>",
	Roles: []string{"<YOUR_ROLE_ID>"}, // Optional
}
```

### Where to set credentials

For each vault, the SDK uses the first of these that's set:

1. `Credentials` in the vault's `VaultConfig`.
2. Client-level credentials, set with `client.WithCredentials` or `AddSkyflowCredentials`.
3. The `SKYFLOW_CREDENTIALS` environment variable, read as a service account credentials JSON string.

If none is set, the call returns a `SkyflowError`.

### Generate tokens yourself

To manage bearer tokens in your own code, generate them with the `serviceaccount` package and pass them as `Token` credentials:

```go
response, err := serviceaccount.GenerateBearerToken("<PATH_TO_YOUR_CREDENTIALS_JSON>", common.BearerTokenOptions{
	RoleIds: []string{"<YOUR_ROLE_ID>"}, // Optional: limit the token to these roles
})
if err != nil {
	fmt.Println("token error:", err.GetMessage())
	return
}
credentials := common.Credentials{Token: response.AccessToken}
```

The package also has `GenerateBearerTokenFromCreds` for a credentials string, `GenerateSignedDataTokens` and `GenerateSignedDataTokensFromCreds` for signed data tokens, and `IsExpired` to check a token before you reuse it.

> [!TIP]
> See the [service account samples](https://github.com/skyflowapi/skyflow-go/tree/main/samples/flowvault/serviceaccount) for tokens with context, scoped tokens, signed data tokens, and token generation from several goroutines.

## Initialize the client

### Vault configuration

Describe each vault with a `common.VaultConfig`:

| Field | Required | Description |
| --- | --- | --- |
| `VaultId` | Yes | The vault's ID. |
| `ClusterId` | Yes, unless `BaseVaultUrl` is set | The cluster ID from the vault URL `https://<CLUSTER_ID>.skyvault.skyflowapis.com`. |
| `Env` | No | The Skyflow environment: `common.PROD`, `common.SANDBOX`, `common.STAGE`, or `common.DEV`. When you leave it unset, the SDK uses `common.PROD`. |
| `BaseVaultUrl` | No | A full `http` or `https` base URL for the vault. When it's set, the SDK uses it instead of building the URL from `ClusterId` and `Env`. |
| `Credentials` | No | Credentials for this vault only. See [Where to set credentials](#where-to-set-credentials). |
| `Timeout`, `MaxRetries`, `InitialRetryDelay`, `MaxRetryDelay` | No | Timeout and retry settings for this vault. See [Timeouts and retries](#timeouts-and-retries). |

`Env` selects the domain in the vault URL:

| `Env` | Vault URL |
| --- | --- |
| `common.PROD` | `https://<CLUSTER_ID>.skyvault.skyflowapis.com` |
| `common.SANDBOX` | `https://<CLUSTER_ID>.skyvault.skyflowapis-preview.com` |
| `common.STAGE` | `https://<CLUSTER_ID>.skyvault.skyflowapis.tech` |
| `common.DEV` | `https://<CLUSTER_ID>.skyvault.skyflowapis.dev` |

### Client options

Pass options to `client.NewSkyflow`:

| Option | Effect |
| --- | --- |
| `client.WithVaults(configs ...common.VaultConfig)` | Registers vaults. Returns a `SkyflowError` if you pass no configs, a config is invalid, or two configs share a vault ID. |
| `client.WithCredentials(credentials common.Credentials)` | Sets credentials for every vault whose config has none. |
| `client.WithHttpConfig(httpConfig common.HttpConfig)` | Sets timeout and retry defaults for every vault. See [Timeouts and retries](#timeouts-and-retries). |
| `client.WithLogLevel(logLevel logger.LogLevel)` | Sets the log level. See [Logging](#logging). |

Then call `Vault` with a registered vault ID to get its `VaultService`. This fragment registers two vaults that share client-level credentials:

```go
skyflowClient, err := client.NewSkyflow(
	client.WithVaults(primaryVaultConfig, secondaryVaultConfig), // VaultConfig values without Credentials
	client.WithCredentials(common.Credentials{ApiKey: "<YOUR_API_KEY>"}),
	client.WithLogLevel(logger.INFO),
)
if err != nil {
	fmt.Println("client error:", err.GetMessage())
	return
}
vault, err := skyflowClient.Vault("<YOUR_VAULT_ID>")
```

`Vault` returns a `SkyflowError` if no vault with that ID is registered.

## Vault operations

The operation examples are fragments. Each assumes `vault`, the `VaultService` from the [Quickstart](#quickstart), and `ctx`, a `context.Context`. Each operation section links to a complete program in [`samples/flowvault`](https://github.com/skyflowapi/skyflow-go/tree/main/samples/flowvault).

The SDK checks every request before sending it. A request that breaks one of the rules for its operation returns a `SkyflowError`, and nothing is sent to the vault.

### Insert and upsert

`Insert` adds records to a table. With upsert, it updates a record that has the same values in the unique columns you name, and inserts a new record when none matches.

Set the table name, and upsert if you use it, at exactly one level:

- **On the request**, to apply them to every record. Records then can't set `TableName` or `Upsert`.
- **On every record**, for example to insert into several tables in one call. The request then can't set `TableName` or `Upsert`.

Each record needs `Data` with at least one column. In `UpsertOptions`, `UniqueColumns` is required, and `UpdateType` is optional. For what `UpdateType` does, see [Update](#update).

```go
request := common.InsertRequest{
	TableName: "<SENSITIVE_DATA_TABLE>",
	Upsert: &common.UpsertOptions{
		UniqueColumns: []string{"email"},
		UpdateType:    common.UpdateTypeUpdate,
	},
	Records: []common.InsertRequestRecord{
		{Data: map[string]interface{}{"name": "John Doe", "email": "john@example.com"}},
	},
}
response, err := vault.Insert(ctx, request, common.InsertOptions{})
```

To set the table and upsert on each record instead:

```go
request := common.InsertRequest{
	Records: []common.InsertRequestRecord{
		{TableName: "<TABLE_1>", Data: map[string]interface{}{"name": "Jane Doe"}},
		{
			TableName: "<TABLE_2>",
			Data:      map[string]interface{}{"email": "jane@example.com"},
			Upsert:    &common.UpsertOptions{UniqueColumns: []string{"email"}, UpdateType: common.UpdateTypeReplace},
		},
	},
}
```

`response.Records` has one `InsertResponseRecord` per request record, in request order. Each record has `TableName`, `SkyflowId`, `Tokens`, `Data`, `HashedData`, `HttpCode`, `Error`, and `RequestId`. `Tokens` maps each column to its tokens, and each `Token` has `Token`, `TokenGroupName`, and `Path`.

> [!TIP]
> See the full sample in [insert_records.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/insert_records.go).

### Update

`Update` changes existing records, which you identify by Skyflow ID.

- Each record needs `SkyflowId` and `Data` with at least one column.
- Set `TableName` on the request or on every record, not both.
- `UpdateType` applies to the whole request:
  - `common.UpdateTypeUpdate` changes only the columns in `Data`. The vault uses this type when you leave `UpdateType` empty.
  - `common.UpdateTypeReplace` overwrites the record with the columns in `Data`.

```go
request := common.UpdateRequest{
	TableName:  "<SENSITIVE_DATA_TABLE>",
	UpdateType: common.UpdateTypeUpdate,
	Records: []common.UpdateRequestRecord{
		{SkyflowId: "<SKYFLOW_ID>", Data: map[string]interface{}{"name": "Jane Doe"}},
	},
}
response, err := vault.Update(ctx, request, common.UpdateOptions{})
```

Each record in `response.Records` has the same fields as an [insert](#insert-and-upsert) result.

> [!TIP]
> See the full sample in [update_record.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/update_record.go).

### Get

`Get` returns records from one table, or from several tables in one call.

To read from one table, set these fields on `GetRequest`:

- `TableName`, which is required.
- Exactly one of `SkyflowIds` or `UniqueValues`. Each `UniqueValues` entry maps a unique column to the value to look up.
- Optionally, `Columns` to return only those columns, and `ColumnRedactions` to apply a redaction type to a column. Each redaction needs both `ColumnName` and `Redaction`.
- Optionally, `Limit` and `Offset`. Neither can be negative, and the SDK doesn't send a value of `0`.

```go
request := common.GetRequest{
	TableName:  "<SENSITIVE_DATA_TABLE>",
	SkyflowIds: []string{"<SKYFLOW_ID1>", "<SKYFLOW_ID2>"},
	Columns:    []string{"name", "email"}, // Optional
}
response, err := vault.Get(ctx, request, common.GetOptions{})
```

To look up records by a unique column instead:

```go
request := common.GetRequest{
	TableName:    "<SENSITIVE_DATA_TABLE>",
	UniqueValues: []map[string]interface{}{{"email": "john@example.com"}},
}
```

To read from several tables, set only `Records` and leave the single-table fields empty. Each `GetRequestRecord` follows the single-table rules, without `Limit` and `Offset`:

```go
request := common.GetRequest{Records: []common.GetRequestRecord{
	{TableName: "<TABLE_1>", SkyflowIds: []string{"<SKYFLOW_ID1>"}},
	{TableName: "<TABLE_2>", UniqueValues: []map[string]interface{}{{"email": "john@example.com"}}},
}}
```

Each record in `response.Records` has the same fields as an [insert](#insert-and-upsert) result, with the column values in `Data`.

> [!TIP]
> See the full sample in [get_records.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/get_records.go).

### Delete

`Delete` removes records from a table. Set `TableName` and exactly one of `SkyflowIds` or `UniqueValues`.

```go
request := common.DeleteRequest{
	TableName:  "<SENSITIVE_DATA_TABLE>",
	SkyflowIds: []string{"<SKYFLOW_ID1>", "<SKYFLOW_ID2>"},
}
response, err := vault.Delete(ctx, request, common.DeleteOptions{})
```

Each record in `response.Records` has `SkyflowId`, `HttpCode`, `Error`, and `RequestId`.

> [!TIP]
> See the full sample in [delete_records.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/delete_records.go).

### Detokenize

`Detokenize` returns the values behind tokens. `Tokens` must contain at least one token and no empty strings. To redact the values of a token group, add a `TokenGroupRedaction` with both `TokenGroupName` and `Redaction`.

```go
request := common.DetokenizeRequest{
	Tokens: []string{"<TOKEN1>", "<TOKEN2>"},
	TokenGroupRedactions: []common.TokenGroupRedaction{ // Optional
		{TokenGroupName: "<TOKEN_GROUP_NAME>", Redaction: "<REDACTION>"},
	},
}
response, err := vault.Detokenize(ctx, request, common.DetokenizeOptions{})
```

Each record in `response.Records` has `Token`, `Value`, `TokenGroupName`, `HttpCode`, `Error`, and `RequestId`. `Metadata`, when present, has the `SkyflowId` and `TableName` of the record the token belongs to.

> [!TIP]
> See the full sample in [detokenize_records.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/detokenize_records.go).

### Get tokens

`GetTokens` returns the token the vault already issued for each value in a deterministic token group. It doesn't create tokens. Each record needs a `Value` and a `TokenGroupName`.

```go
request := common.GetTokensRequest{Records: []common.GetTokensRequestRecord{
	{Value: "john@example.com", TokenGroupName: "<TOKEN_GROUP_NAME>"},
	{Value: "jane@example.com", TokenGroupName: "<TOKEN_GROUP_NAME>"},
}}
response, err := vault.GetTokens(ctx, request, common.GetTokensOptions{})
```

Each record in `response.Records` has `Value`, `TokenGroupName`, `Token`, `HttpCode`, `Error`, and `RequestId`. A value with no token has `Error` set.

> [!TIP]
> See the full sample in [get_tokens.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/get_tokens.go).

### Query

`Query` runs a read-only SQL query. Only `SELECT` is supported, and a call returns at most 25 records. To get more, run the query again with `OFFSET`. A query can't return tokens or file URLs.

```go
request := common.QueryRequest{Query: "SELECT name, email FROM <SENSITIVE_DATA_TABLE> LIMIT 25"}
response, err := vault.Query(ctx, request, common.QueryOptions{})
```

`response.Fields` has one map per row, `response.Metadata.Columns` lists the columns, and `response.RequestId` identifies the call. A query has no per-record results, so any failure returns a `SkyflowError`.

> [!TIP]
> See the full sample in [query_records.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/query_records.go).

### Upload files

`UploadFiles` stores files in a table's file columns. Leave `SkyflowId` empty on a record to create a new record, or set it to upload into an existing record. Each record needs `TableName` and at least one column.

Each `UploadFilesRequestColumn` needs `Column` and exactly one file source:

- `FilePath`, a path to a file.
- `Base64`, the file's content encoded as base64. `FileName` is required with `Base64`.
- `FileObject`, an open `*os.File`. Close it yourself after the call.

If you leave `FileName` empty, the SDK takes the name from `FilePath` or `FileObject`. If no name is available, the vault generates one.

```go
file, openErr := os.Open("<PATH_TO_ANOTHER_FILE>")
if openErr != nil {
	fmt.Println("could not open file:", openErr)
	return
}
defer file.Close()

request := common.UploadFilesRequest{Records: []common.UploadFilesRequestRecord{
	{
		TableName: "<SENSITIVE_DATA_TABLE>",
		Columns: []common.UploadFilesRequestColumn{
			{Column: "<FILE_COLUMN_1>", FilePath: "<PATH_TO_FILE>"},
			{Column: "<FILE_COLUMN_2>", Base64: "<BASE64_FILE_CONTENT>", FileName: "<FILE_NAME>"},
			{Column: "<FILE_COLUMN_3>", FileObject: file},
		},
	},
}}
response, err := vault.UploadFiles(ctx, request, common.UploadFilesOptions{})
```

The SDK uploads files in these steps:

1. It reads every file into memory. If a file can't be read or its base64 is invalid, the call returns a `SkyflowError` before anything is sent.
2. It requests a signed upload URL for each column from the vault. This request follows your timeout, retry, and header settings.
3. It uploads each file to its signed URL, at most 5 at a time. These uploads aren't retried, have no overall timeout, and carry none of the SDK's headers or your interceptor's headers. The SDK waits up to 10 seconds to connect to the storage service, and up to 10 seconds for its response after sending a file. To stop an upload, cancel `ctx`.

Each record in `response.Records` has `SkyflowId`, `TableName`, `HttpCode`, `Error`, and `RequestId`, plus a `Columns` entry for each request column. Each column result has `Column`, `FileName`, `UploadStatus`, and `Error`. `UploadStatus` is one of these values:

- `common.UploadStatusUploaded`: the file was uploaded.
- `common.UploadStatusFailed`: the upload failed, and `Error` says why.
- `common.UploadStatusSkipped`: the SDK didn't upload the file, because the record failed or the vault returned no signed URL for the column.

A failed column doesn't affect the other columns.

> [!TIP]
> See the full sample in [upload_files.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/upload_files.go).

### Delete files

`DeleteFiles` removes files from the file columns of existing records, including the stored files. A request can have up to 20,000 records. Each record needs:

- `TableName`.
- `Columns`, the file columns to clear, each listed once.
- Exactly one of `SkyflowId` or `UniqueValues`.

```go
request := common.DeleteFilesRequest{Records: []common.DeleteFilesRequestRecord{
	{TableName: "<SENSITIVE_DATA_TABLE>", Columns: []string{"<FILE_COLUMN>"}, SkyflowId: "<SKYFLOW_ID>"},
}}
response, err := vault.DeleteFiles(ctx, request, common.DeleteFilesOptions{})
```

Each record in `response.Records` has `SkyflowId`, `TableName`, `HttpCode`, `Error`, and `RequestId`. A successful record has `Columns`, with the `Column` and `Status` of each deleted file. A failed record has no `Columns`.

> [!TIP]
> See the full sample in [delete_files.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/delete_files.go).

## Timeouts and retries

Four settings control how long a call can take and whether the SDK retries it:

| Setting | Default | Description |
| --- | --- | --- |
| `Timeout` | 60 seconds | The limit for the whole call, including retries and the waits between them. |
| `MaxRetries` | `0`, so no retries | How many times to retry after the first attempt. The SDK retries only responses with HTTP status 408, 429, or 5xx. |
| `InitialRetryDelay` | 500 milliseconds | The wait before the first retry. The wait doubles for each later retry, with up to 20% random variation. |
| `MaxRetryDelay` | 2 seconds | The longest wait between retries. |

Set the client-wide values with `client.WithHttpConfig` and a `common.HttpConfig`. Set values for one vault with the same fields on its `VaultConfig`. The SDK picks each setting separately: the vault's value, then the client's, then the default.

```go
clientRetries := 3
vaultRetries := 2

vaultConfig := common.VaultConfig{
	VaultId:     "<YOUR_VAULT_ID>",
	ClusterId:   "<YOUR_CLUSTER_ID>",
	Credentials: common.Credentials{Path: "<PATH_TO_YOUR_CREDENTIALS_JSON>"},
	// For this vault only. These values win over the client-wide ones.
	Timeout:    30 * time.Second,
	MaxRetries: &vaultRetries,
}

skyflowClient, err := client.NewSkyflow(
	client.WithHttpConfig(common.HttpConfig{
		Timeout:           60 * time.Second,
		MaxRetries:        &clientRetries,
		InitialRetryDelay: 500 * time.Millisecond,
		MaxRetryDelay:     4 * time.Second,
	}),
	client.WithVaults(vaultConfig),
)
```

Keep these points in mind:

- A zero duration or a nil `MaxRetries` means "not set", so the next level applies. `MaxRetries` is a pointer so that a vault can set `0` to turn off retries that the client turns on.
- A negative value returns a `SkyflowError` from `WithHttpConfig`, `WithVaults`, `AddVaultConfig`, or `UpdateVaultConfig`.
- A deadline on `ctx` still applies when it's shorter than `Timeout`.
- Retries send the same request again, for every operation, including writes.
- A failure that returns no HTTP response, such as a refused connection, isn't retried.
- File uploads in `UploadFiles` aren't retried and have no overall timeout. See [Upload files](#upload-files).

> [!TIP]
> See the full sample in [timeout_and_retry_config_example.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/timeout_and_retry_config_example.go).

## Custom request headers

To add headers to one call, set `Interceptor` in that operation's options struct. Every options struct has this field. The interceptor runs once before the call is sent. It receives a `*common.RequestContext`, whose `Operation` field names the operation, such as `INSERT` or `GET_TOKENS`. Add headers with `AddHeader`.

```go
options := common.InsertOptions{
	Interceptor: func(rc *common.RequestContext) {
		rc.AddHeader(common.RequestIdHeader, "<YOUR_REQUEST_ID>")
	},
}
response, err := vault.Insert(ctx, request, options)
```

You can add these headers:

| Key | Header |
| --- | --- |
| `common.SkyflowAccountId` | `x-skyflow-account-id` |
| `common.SkyflowAccountName` | `x-skyflow-account-name` |
| `common.RequestIdHeader` | `x-request-id` |

A header with any other key or an empty value makes the call return a `SkyflowError`, and nothing is sent. Retries send the same headers. The SDK has no client-level headers, so set the interceptor on each call that needs them.

> [!TIP]
> See the full sample in [custom_header_example.go](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/vaultapi/custom_header_example.go).

## Manage the client

After you create the client, you can change its vaults and credentials with these methods:

| Method | Effect |
| --- | --- |
| `AddVaultConfig(config)` | Registers another vault. Returns a `SkyflowError` if the config is invalid or the vault ID is already registered. |
| `UpdateVaultConfig(config)` | Changes the settings of a registered vault. Returns a `SkyflowError` if `VaultId` is empty or isn't registered. |
| `RemoveVaultConfig(vaultId)` | Removes a registered vault. Returns a `SkyflowError` if the vault ID isn't registered. |
| `GetVaultConfig(vaultId)` | Returns a copy of a vault's config, or a `SkyflowError` if the vault ID isn't registered. |
| `AddSkyflowCredentials(credentials)`, `UpdateSkyflowCredentials(credentials)` | Replace the client-level credentials. |
| `GetSkyflowCredentials()` | Returns a copy of the client-level credentials, or `nil` if none are set. |

To update a vault, pass a `VaultConfig` with its `VaultId` and only the fields you want to change. Fields left at their zero value, including an unset `Env`, keep their current values. Because a zero value means "keep", `UpdateVaultConfig` can't clear a field. To clear one, remove the vault and add it again.

```go
err := skyflowClient.UpdateVaultConfig(common.VaultConfig{
	VaultId:     "<YOUR_VAULT_ID>",
	Credentials: common.Credentials{ApiKey: "<YOUR_NEW_API_KEY>"},
})
```

You can change credentials at the client or for one vault. Either way, the SDK drops the bearer token it generated from the old credentials and authenticates again on the next call.

The client is safe for concurrent use. You can change vaults and credentials while calls are running. A call that has already started keeps the settings it started with. A `VaultService` you got earlier uses the new settings on its next call.

## Logging

The SDK logs at the `logger.ERROR` level by default. Set another level with `client.WithLogLevel`, or change it later with `UpdateLogLevel`. The levels are `logger.ERROR`, `logger.WARN`, `logger.INFO`, `logger.DEBUG`, and `logger.OFF`.

```go
skyflowClient, err := client.NewSkyflow(
	client.WithVaults(vaultConfig),
	client.WithLogLevel(logger.INFO),
)
```

Logs go to standard error. To write them somewhere else, pass an `io.Writer` to `logger.SetOutput`. The log level and output apply to the whole process, including other Skyflow Go SDK modules in the same program.

## Error handling

A call can fail in two ways:

- **The whole call fails.** The operation returns a `*SkyflowError` and no response.
- **Some records fail.** The operation returns a response and no error, and each failed record carries its own error. Checking only the returned error misses these failures.

### Whole-call errors

An operation returns a `*SkyflowError` in these cases:

- The request breaks a validation rule. The SDK sends nothing, and `GetHttpCode` returns `"400"`.
- The credentials are missing or invalid, or the bearer token in `Token` has expired.
- The vault returns an error without per-record results.
- The request doesn't reach the vault or gets no response, for example because of a network failure. `GetHttpCode` returns `"500"`.

`SkyflowError` has these methods:

| Method | Returns |
| --- | --- |
| `GetHttpCode()` | The HTTP code as a string |
| `GetMessage()` | The error message |
| `GetRequestId()` | The request ID of the failed call, when the vault returned one |
| `GetGrpcCode()` | The gRPC code from the vault's error, if any |
| `GetHttpStatusCode()` | The HTTP status from the vault's error, if any |
| `GetDetails()` | Details from the vault's error, if any |

Operations return the concrete `*SkyflowError` type. Compare the returned value with `nil` directly. If you store a nil `*SkyflowError` in a variable of type `error`, that variable isn't equal to `nil`.

```go
response, err := vault.Insert(ctx, request, common.InsertOptions{})
if err != nil {
	fmt.Println("HTTP code:", err.GetHttpCode())
	fmt.Println("message:", err.GetMessage())
	fmt.Println("request ID:", err.GetRequestId())
	fmt.Println("details:", err.GetDetails())
	return
}
```

### Per-record results

Every operation except `Query` returns per-record results. The vault can answer with an error status and still return results for each record. In that case the SDK returns those results as the response, not as a `SkyflowError`.

Check `Error` on each record. A failed record has:

- `Error`, the error message.
- `HttpCode`, the HTTP code for that record.
- `RequestId`, the request ID of the call. The SDK sets it only on failed records.

```go
for _, record := range response.Records {
	if record.Error != "" {
		fmt.Printf("record failed (%d): %s [request ID: %s]\n", record.HttpCode, record.Error, record.RequestId)
		continue
	}
	fmt.Println(record.SkyflowId, record.Tokens)
}
```

For `UploadFiles`, also check `UploadStatus` and `Error` on each column. See [Upload files](#upload-files).

### Retry failed records

The SDK's retries act on the HTTP status of the whole call. It doesn't retry a record that failed inside a successful response. To retry such records, send them again in a new request. Use `HttpCode` and `Error` to decide which failures to send again.

This fragment sends the failed records of an insert again:

```go
response, err := vault.Insert(ctx, request, common.InsertOptions{})
if err != nil {
	fmt.Println("insert failed:", err.GetMessage())
	return
}
var failed []common.InsertRequestRecord
for i, record := range response.Records {
	if record.Error != "" {
		failed = append(failed, request.Records[i]) // Results follow request order
	}
}
if len(failed) > 0 {
	retryRequest := request // Keeps the request-level TableName and Upsert
	retryRequest.Records = failed
	response, err = vault.Insert(ctx, retryRequest, common.InsertOptions{})
	// Check err and response.Records again, as for the first call.
}
```

## Samples

The [`samples/flowvault`](https://github.com/skyflowapi/skyflow-go/tree/main/samples/flowvault) directory has a runnable program for each operation, plus custom headers, timeout and retry settings, and service account token generation. Its [README](https://github.com/skyflowapi/skyflow-go/blob/main/samples/flowvault/README.md) lists the samples and explains how to run them.
