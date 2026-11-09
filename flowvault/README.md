# Skyflow FlowVault Go SDK

The `flowvault` module is the Skyflow Go SDK for **FlowVault** vaults. It shares its credentials, configuration, logger and error types with the [SkyVault Go SDK](../README.md) (both depend on the `common` module) but exposes a different, narrower surface: **unary** vault operations (insert, get, update, delete, detokenize, query, get tokens, upload files and delete files).

> Meant for **FlowVault** vaults.

> **`github.com/skyflowapi/skyflow-go/flowvault` is versioned independently of the SkyVault SDK (`github.com/skyflowapi/skyflow-go/v2`).** It is a separate Go module with its own tags (`flowvault/vX.Y.Z`), so a lower `flowvault` version number does not mean it is older or behind. Upgrade each module on its own.

[![CI](https://img.shields.io/static/v1?label=CI&message=passing&color=green?style=plastic&logo=github)](https://github.com/skyflowapi/skyflow-go/actions)
[![License](https://img.shields.io/github/license/skyflowapi/skyflow-go)](https://github.com/skyflowapi/skyflow-go/blob/main/LICENSE)

# Table of Contents

- [Table of Contents](#table-of-contents)
- [Overview](#overview)
- [Install](#install)
  - [Requirements](#requirements)
  - [Configuration](#configuration)
- [Quickstart](#quickstart)
- [Authenticate](#authenticate)
  - [Credential types](#credential-types)
  - [Where credentials can be set](#where-credentials-can-be-set)
  - [Generate a bearer token](#generate-a-bearer-token)
  - [Generate bearer tokens with context](#generate-bearer-tokens-with-context)
  - [Generate scoped bearer tokens](#generate-scoped-bearer-tokens)
  - [Generate signed data tokens](#generate-signed-data-tokens)
- [Initialize the client](#initialize-the-client)
  - [VaultConfig reference](#vaultconfig-reference)
  - [Client options and methods reference](#client-options-and-methods-reference)
  - [Timeouts and retries](#timeouts-and-retries)
  - [Logging](#logging)
  - [Concurrency and resource lifecycle](#concurrency-and-resource-lifecycle)
- [Vault operations](#vault-operations)
  - [Vault type support](#vault-type-support)
- [Insert](#insert)
- [Detokenize](#detokenize)
- [Get](#get)
- [Update](#update)
- [Delete](#delete)
- [Query](#query)
- [Get Tokens](#get-tokens)
- [Upload Files](#upload-files)
- [Delete Files](#delete-files)
- [Custom Request Headers](#custom-request-headers)
- [Error Handling](#error-handling)
  - [Two layers of errors](#two-layers-of-errors)
  - [Per-record success and failure](#per-record-success-and-failure)
  - [Handling SkyflowError](#handling-skyflowerror)
  - [SkyflowError methods](#skyflowerror-methods)
  - [Retrying the failed records](#retrying-the-failed-records)
- [Samples](#samples)

# Overview

- Authenticate with a Skyflow service account, an API key or a bearer token. See [Authenticate](#authenticate).
- Perform vault operations: insert, get, update, delete, detokenize, query, get tokens, upload files and delete files. See [Vault operations](#vault-operations).
- **Per-record reporting, not all-or-nothing.** A call succeeds as a call even when individual records fail; every response reports the outcome of each record or token. See [Error Handling](#error-handling).
- Every call takes a `context.Context`, so you control cancellation and deadlines.

# Install

## Requirements

- Go 1.25 and above

## Configuration

Make sure your project uses Go modules, then add the SDK:

```bash
go get github.com/skyflowapi/skyflow-go/flowvault
```

Import the packages you need:

```go
import (
	"github.com/skyflowapi/skyflow-go/flowvault/client"         // Skyflow client and VaultService
	"github.com/skyflowapi/skyflow-go/flowvault/serviceaccount" // bearer and signed data token generation
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"   // configs, credentials, request/response types
	skyflowError "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)
```

The module path is different from the SkyVault SDK's (`github.com/skyflowapi/skyflow-go/v2`), so both can be used in one program if you need them.

# Quickstart

```go
package main

import (
	"context"
	"fmt"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

func main() {
	credentials := common.Credentials{ApiKey: "<API_KEY>"} // or Token / Path / CredentialsString

	vaultConfig := common.VaultConfig{
		VaultId:     "<VAULT_ID>",
		ClusterId:   "<CLUSTER_ID>", // part of the vault URL: https://{clusterId}.vault.skyflowapis.com
		Env:         common.PROD,    // DEV, STAGE, SANDBOX or PROD (the default)
		Credentials: credentials,
	}

	skyflowClient, err := client.NewSkyflow(client.WithVaults(vaultConfig))
	if err != nil {
		fmt.Println(err.GetHttpCode(), err.GetMessage())
		return
	}

	// Returns the service for the given vault
	vault, err := skyflowClient.Vault("<VAULT_ID>")
	if err != nil {
		fmt.Println(err.GetHttpCode(), err.GetMessage())
		return
	}

	response, err := vault.Insert(context.Background(), common.InsertRequest{
		TableName: "cards",
		Records:   []common.InsertRequestRecord{{Data: map[string]interface{}{"card_number": "4111111111111111"}}},
	}, common.InsertOptions{})
	if err != nil {
		fmt.Println(err.GetHttpCode(), err.GetMessage())
		return
	}
	fmt.Println(response.Records)
}
```

`Vault("<VAULT_ID>")` returns the `*client.VaultService` for a registered vault. To talk to more than one vault from one client, pass each config to `WithVaults(...)` (or add it later with `AddVaultConfig`) and fetch each service by ID.

# Authenticate

Requests are authorized with Skyflow credentials, set as a `common.Credentials` value. Credential handling comes from the shared `common` module, so it works the same way as in the SkyVault SDK.

## Credential types

Set **exactly one** of the following fields. Setting more than one, or none, fails validation with a `*SkyflowError`.

| Field | What it is |
|---|---|
| `ApiKey` | A long-lived key that authenticates and authorizes requests. Simplest option. Must have the form `sky-<5 chars>-<32 hex chars>`. |
| `Token` | A short-lived bearer token, typically generated from service account credentials. See [Generate a bearer token](#generate-a-bearer-token). |
| `Path` | Filesystem path to a service account `credentials.json`. The SDK generates and refreshes bearer tokens from it. |
| `CredentialsString` | The contents of a service account `credentials.json` as a JSON string. Use it when the credentials come from a secret store rather than a file. |

Two optional fields apply when the SDK generates tokens for you (with `Path` or `CredentialsString`):

| Field | Description |
|---|---|
| `Roles` | Restrict the generated token to specific role IDs (a scoped token). When set, it must not be empty or contain a blank role. |
| `Context` | Attach context to the generated token for context-aware authorization: a `string` or a `map[string]interface{}`. |

```go
// API key
credentials := common.Credentials{ApiKey: "<API_KEY>"}

// Bearer token you generated yourself
credentials = common.Credentials{Token: "<BEARER_TOKEN>"}

// Service account credentials file: the SDK handles token generation and refresh
credentials = common.Credentials{Path: "<PATH_TO_CREDENTIALS_JSON>"}

// Service account credentials as a JSON string, with a scoped role and context
credentials = common.Credentials{
	CredentialsString: "<CREDENTIALS_JSON_STRING>",
	Roles:             []string{"<ROLE_ID>"},
	Context:           "<CONTEXT_ID>",
}
```

When the SDK generates the token, it caches it per vault, regenerates it when it expires, and resends a request once with a fresh token if the vault rejects a token that was about to expire with `401`.

## Where credentials can be set

Credentials resolve **most specific first**:

1. **Per vault**: `VaultConfig.Credentials`. Wins for that vault.
2. **Client-wide**: `client.WithCredentials(credentials)` or `skyflowClient.AddSkyflowCredentials(credentials)`. Used by any vault that has none of its own.
3. **Environment**: the `SKYFLOW_CREDENTIALS` environment variable, holding a service account credentials JSON string. If the environment doesn't set it, the SDK reads it from a `.env` file in the working directory (the file is only read, never loaded into the environment).

If none of the three yields credentials, the call fails with a `*SkyflowError`.

## Generate a bearer token

If you'd rather manage tokens yourself, the service account functions live in the `flowvault/serviceaccount` package:

- `GenerateBearerToken(path, options)`: mint a token from a service account credentials **file path**.
- `GenerateBearerTokenFromCreds(credentialsString, options)`: the same, from the credentials JSON as a **string**.
- `IsExpired(token)`: `true` when the token is empty, unreadable or past its expiry. Use it to reuse a cached token until it expires.
- `GenerateSignedDataTokens(path, options)` / `GenerateSignedDataTokensFromCreds(credentialsString, options)`: see [Generate signed data tokens](#generate-signed-data-tokens).

`GenerateBearerToken*` returns a `*common.TokenResponse`; the token is in `AccessToken`. Tokens are valid for 60 minutes and carry the service account's permissions.

[Example](../samples/flowvault/serviceaccount/bearer_token_generation_example.go):

```go
import (
	"github.com/skyflowapi/skyflow-go/flowvault/serviceaccount"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

var bearerToken = ""

func getToken(filePath string) (string, error) {
	if bearerToken != "" && !serviceaccount.IsExpired(bearerToken) {
		return bearerToken, nil // reuse until it expires
	}
	response, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{})
	if err != nil {
		return "", err
	}
	bearerToken = response.AccessToken
	return bearerToken, nil
}

// ...or from a credentials JSON string:
// response, err := serviceaccount.GenerateBearerTokenFromCreds(credentialsString, common.BearerTokenOptions{})
```

To use the token with the SDK, set it on the credentials:

```go
credentials := common.Credentials{Token: bearerToken}
```

To generate bearer tokens concurrently from several goroutines, see [bearer_token_generation_using_goroutines_example.go](../samples/flowvault/serviceaccount/bearer_token_generation_using_goroutines_example.go).

## Generate bearer tokens with context

**Context-aware authorization** embeds context values into a bearer token when it is generated, so your policies can reference them. This allows more flexible access control, such as tracking end-user identity when calling through a service account, and is required for detokenizing signed data tokens.

Set `Ctx` on `common.BearerTokenOptions`, as either a **string** or a **map**:

```go
// String context: a single value your policy references as request.context
response, err := serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{Ctx: "user_12345"})

// Map context: each key maps to a Skyflow CEL policy variable under request.context.*
response, err = serviceaccount.GenerateBearerToken(filePath, common.BearerTokenOptions{
	Ctx: map[string]interface{}{
		"role":       "admin",
		"department": "finance",
		"user_id":    "user_12345",
	},
})
```

With the map above, your policies can reference `request.context.role`, `request.context.department` and `request.context.user_id` to make conditional access decisions.

When the SDK generates tokens for you, set the same value on `Credentials.Context` instead.

[Full example](../samples/flowvault/serviceaccount/bearer_token_generation_with_context_example.go)

## Generate scoped bearer tokens

A service account with several roles can generate bearer tokens limited to specific roles by passing their role IDs in `RoleIds`. This is useful for services with several responsibilities, such as separating billing access from analytics access. The tokens are valid for 60 minutes and can only perform operations permitted by the designated roles.

[Example](../samples/flowvault/serviceaccount/scoped_token_generation_example.go):

```go
options := common.BearerTokenOptions{RoleIds: []string{"<ROLE_ID1>", "<ROLE_ID2>"}}
response, err := serviceaccount.GenerateBearerToken(filePath, options)
// ...or from a credentials string: serviceaccount.GenerateBearerTokenFromCreds(credentialsString, options)
```

When the SDK generates tokens for you, set `Credentials.Roles` instead.

## Generate signed data tokens

Skyflow generates data tokens when sensitive data is inserted into the vault. Those data tokens can be digitally signed with the private key of the service account credentials, which adds a further layer of protection. A signed token can only be detokenized together with a bearer token generated from service account credentials that hold the matching context and permissions.

`common.SignedDataTokensOptions` takes `Ctx` (string or map, as for bearer tokens), `DataTokens` (the tokens to sign) and `TimeToLive` (seconds; 60 when not set):

```go
results, err := serviceaccount.GenerateSignedDataTokens(filePath, common.SignedDataTokensOptions{
	Ctx:        "user_12345",
	DataTokens: []string{"<DATA_TOKEN1>", "<DATA_TOKEN2>"},
	TimeToLive: 30, // seconds
})
// ...or from a credentials string: serviceaccount.GenerateSignedDataTokensFromCreds(credentialsString, options)
if err != nil {
	fmt.Println(err.GetMessage())
	return
}
for _, result := range results {
	fmt.Println(result.Token, "->", result.SignedToken)
}
```

`GenerateSignedDataTokens*` returns a `[]common.SignedDataTokensResponse`, each with `Token` and `SignedToken`.

[Full example](../samples/flowvault/serviceaccount/signed_token_generation_example.go)

# Initialize the client

`*client.Skyflow` is the client. Create it once with `client.NewSkyflow(...)`, keep it for the lifetime of your application, and get a `*client.VaultService` from it with `Vault(...)`.

```go
import (
	"time"

	"github.com/skyflowapi/skyflow-go/flowvault/client"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	"github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

// Step 1: Credentials, exactly one credential type
credentials := common.Credentials{Path: "<PATH_TO_CREDENTIALS_JSON>"}

// Step 2: Vault configuration
maxRetries := 2
vaultConfig := common.VaultConfig{
	VaultId:     "<VAULT_ID>",
	ClusterId:   "<CLUSTER_ID>",
	Env:         common.PROD, // DEV, STAGE, SANDBOX or PROD (the default)
	Credentials: credentials,
	// Optional: vault-level HTTP overrides
	Timeout:    120 * time.Second, // overall call timeout
	MaxRetries: &maxRetries,       // retries after the first failure
}

// Step 3: Create the client
skyflowClient, err := client.NewSkyflow(
	client.WithVaults(vaultConfig),
	client.WithLogLevel(logger.INFO), // default is ERROR
)
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage())
	return
}

// Step 4: Get the vault service and make calls
vault, err := skyflowClient.Vault("<VAULT_ID>")
```

## VaultConfig reference

`common.VaultConfig` is a struct. A zero value (or nil `MaxRetries`) means "not set".

| Field | Type | Required | Description |
|---|---|---|---|
| `VaultId` | `string` | required | The vault's ID. |
| `ClusterId` | `string` | required unless `BaseVaultUrl` is set | The cluster part of the vault URL: `https://{clusterId}.vault.skyflowapis.com`. |
| `Env` | `common.Env` | optional | `common.DEV`, `common.STAGE`, `common.SANDBOX` or `common.PROD`. The zero value is `PROD`. |
| `Credentials` | `common.Credentials` | optional | Credentials for this vault. Falls back to client-wide credentials, then `SKYFLOW_CREDENTIALS`. |
| `BaseVaultUrl` | `string` | optional | Full vault URL, used instead of the one built from `ClusterId` and `Env`. Must use `https` (`http` is accepted only for localhost and loopback addresses). `BaseVaultURL` is a deprecated alias. |
| `Timeout` | `time.Duration` | optional | Overall call timeout, retries included. |
| `MaxRetries` | `*int` | optional | Retry attempts after the first failure. |
| `InitialRetryDelay` | `time.Duration` | optional | Backoff before the first retry. |
| `MaxRetryDelay` | `time.Duration` | optional | Ceiling the exponential backoff grows to. |

The vault URL for each `Env`:

| `Env` | Vault URL |
|---|---|
| `PROD` | `https://{clusterId}.vault.skyflowapis.com` |
| `SANDBOX` | `https://{clusterId}.vault.skyflowapis-preview.com` |
| `STAGE` | `https://{clusterId}.vault.skyflowapis.tech` |
| `DEV` | `https://{clusterId}.vault.skyflowapis.dev` |

## Client options and methods reference

Options for `client.NewSkyflow(opts ...client.Option)`:

| Option | Description |
|---|---|
| `client.WithVaults(configs ...common.VaultConfig)` | Register one or more vaults. Fails if called with no configs. |
| `client.WithCredentials(credentials)` | Client-wide credentials for vaults that don't set their own. |
| `client.WithLogLevel(level)` | `logger.DEBUG`, `logger.INFO`, `logger.WARN`, `logger.ERROR` (default) or `logger.OFF`. |
| `client.WithHttpConfig(common.HttpConfig{...})` | Client-wide timeouts and retry policy. Same fields as the HTTP fields on `VaultConfig`. |

Methods on `*client.Skyflow`:

| Method | Description |
|---|---|
| `Vault(vaultId string)` | The `*VaultService` for a registered vault. |
| `AddVaultConfig(config)` | Register another vault. Fails if the vault ID is already registered. |
| `UpdateVaultConfig(config)` | Update a registered vault in place. Only the fields you set change, except `Env`, which is always applied (leaving it out moves the vault to `PROD`). |
| `RemoveVaultConfig(vaultId)` | Unregister a vault. |
| `GetVaultConfig(vaultId)` | A copy of a registered vault's config. |
| `AddSkyflowCredentials(credentials)` / `UpdateSkyflowCredentials(credentials)` | Set or replace the client-wide credentials. |
| `GetSkyflowCredentials()` | A copy of the client-wide credentials, or `nil`. |
| `UpdateLogLevel(level)` / `GetLoglevel()` | Change or read the log level. |

Each of these that can fail returns a `*SkyflowError`.

## Timeouts and retries

Each HTTP setting resolves **most specific first**: the value on the `VaultConfig`, else the client-wide value from `WithHttpConfig`, else the SDK default. Negative values fail validation.

| Setting | SDK default |
|---|---|
| `Timeout` (overall call, retries included) | 60 s |
| `MaxRetries` | `0`: retries are **opt-in**, so non-idempotent writes are never replayed silently |
| `InitialRetryDelay` | 500 ms |
| `MaxRetryDelay` | 2000 ms |

[Example](../samples/flowvault/vaultapi/timeout_and_retry_config_example.go):

```go
clientRetries := 3
skyflowClient, err := client.NewSkyflow(
	client.WithHttpConfig(common.HttpConfig{ // client-wide policy
		Timeout:           60 * time.Second,
		MaxRetries:        &clientRetries,
		InitialRetryDelay: 500 * time.Millisecond,
		MaxRetryDelay:     4 * time.Second,
	}),
	client.WithVaults(common.VaultConfig{
		VaultId:     "<VAULT_ID>",
		ClusterId:   "<CLUSTER_ID>",
		Env:         common.PROD,
		Credentials: common.Credentials{ApiKey: "<API_KEY>"},
		Timeout:     300 * time.Second, // per-vault override
	}),
)
```

When retries are enabled, retryable responses (HTTP `408`, `429` and `5xx`) are retried with exponential backoff and jitter, bounded by `InitialRetryDelay`/`MaxRetryDelay` and the overall `Timeout`. A deadline on the `context.Context` you pass to a call also applies when it is shorter. The file uploads in [Upload Files](#upload-files) are not retried and have no overall timeout, so large files aren't cut off.

## Logging

The SDK logs at `logger.ERROR` by default, to standard error. Set the level with `client.WithLogLevel(logger.DEBUG)` or `skyflowClient.UpdateLogLevel(logger.DEBUG)`, or turn logging off with `logger.OFF`. Send the output elsewhere with `logger.SetOutput(w io.Writer)`.

The level and output are **process-wide**: they are shared by every client of the FlowVault and SkyVault Go SDKs in the process.

## Concurrency and resource lifecycle

- **Concurrency**: `*client.Skyflow` and the `*VaultService` values it returns are safe for concurrent use. Vaults and client settings can be changed while calls are running; a call that is already running keeps the config it started with, and a `VaultService` obtained earlier sees the change on its next call.
- **Lifecycle**: create the client once and reuse it. The underlying HTTP client and the cached bearer token are reused across calls rather than recreated per request. There is nothing to close.
- **Cancellation**: every call takes a `context.Context`. If the context is cancelled or its deadline passes, the call stops and returns a `*SkyflowError`. When that happens while the call waits for a bearer token, the error has HTTP code `499` and unwraps to the context's error.

# Vault operations

`*client.VaultService` is returned by `skyflowClient.Vault(...)` and exposes these operations. Each is a single method call that returns the result directly (`UploadFiles` orchestrates the signed-URL upload across a couple of calls internally). Every method has the shape `Method(ctx, request, options) (*Response, *SkyflowError)`. Request, options and response types are in `flowvault/utils/common`.

| Method | Request | Options | Response | Description |
|--------|---------|---------|----------|-------------|
| `Insert` | `InsertRequest` | `InsertOptions` | `InsertResponse` | Insert records, optionally across several tables, in one call |
| `Detokenize` | `DetokenizeRequest` | `DetokenizeOptions` | `DetokenizeResponse` | Detokenize tokens, optionally with a redaction override per token group |
| `Get` | `GetRequest` | `GetOptions` | `GetResponse` | Read records by Skyflow ID or unique value, optionally with a redaction override per column |
| `Update` | `UpdateRequest` | `UpdateOptions` | `UpdateResponse` | Update records by Skyflow ID |
| `Delete` | `DeleteRequest` | `DeleteOptions` | `DeleteResponse` | Delete records by Skyflow ID or unique value |
| `Query` | `QueryRequest` | `QueryOptions` | `QueryResponse` | Run a read-only SQL `SELECT`; returns up to 25 rows |
| `GetTokens` | `GetTokensRequest` | `GetTokensOptions` | `GetTokensResponse` | Look up existing tokens for values (deterministic token groups) |
| `UploadFiles` | `UploadFilesRequest` | `UploadFilesOptions` | `UploadFilesResponse` | Upload files to file columns (the SDK fetches signed URLs and uploads the bytes) |
| `DeleteFiles` | `DeleteFilesRequest` | `DeleteFilesOptions` | `DeleteFilesResponse` | Delete files from file columns by Skyflow ID or unique value |

Pass the zero value of the options struct (for example `common.InsertOptions{}`) when you don't need it. Its only field is `Interceptor`; see [Custom Request Headers](#custom-request-headers).

The snippets below assume `vault` from [Quickstart](#quickstart), `ctx := context.Background()`, and the `common` package imported.

## Vault type support

Which operations make sense depends on whether the vault is **structured** (has a schema: tables and columns) or **schemaless** (stores standalone tokens with no table structure). The record and file operations address a table's columns, so they need a structured vault; the token operations need only the token or token group, so they work against either kind:

| Operation | Supported on |
|---|---|
| `Insert`, `Get`, `Update`, `Delete`, `Query` | Structured (schema) vaults |
| `UploadFiles`, `DeleteFiles` | Structured vaults, on a table's file columns |
| `Detokenize` | Both: detokenizing needs only the token itself |
| `GetTokens` | Both: looks up a token for a value within a token group, not a table |

# Insert

Insert records in a single API call. Each record is a `common.InsertRequestRecord` with its own `Data` and, optionally, its own `TableName`, `Tokens` and `Upsert`.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- `Records` must not be empty, and each record's `Data` must not be empty. No column name may be blank.
- `TableName` and `Upsert` follow a one-level rule: set them on the request (applies to every record) or on every record, never both.
- `Upsert` takes a `*common.UpsertOptions` with non-empty `UniqueColumns` and an optional `UpdateType`: `common.UpdateTypeUpdate` (merge) or `common.UpdateTypeReplace` (overwrite).
- `Tokens` is optional bring-your-own-token: a map of column name to token.

```go
// Insert several records into one table: set TableName once on the request
request := common.InsertRequest{
	TableName: "cards",
	Records: []common.InsertRequestRecord{
		{Data: map[string]interface{}{"card_number": "4111111111111111", "cardholder_name": "john doe"}},
		{Data: map[string]interface{}{"card_number": "4222222222222222", "cardholder_name": "jane doe"}},
	},
}
response, err := vault.Insert(ctx, request, common.InsertOptions{})
```

To set the table name on each record instead (the same one-level rule), leave the request-level `TableName` empty and set it on **every** record. This is also how you insert across tables in one call, and where per-record `Upsert` goes:

```go
request := common.InsertRequest{Records: []common.InsertRequestRecord{
	{TableName: "table1", Data: map[string]interface{}{"card_number": "4111111111111111", "cardholder_name": "john doe"}},
	{
		TableName: "table2",
		Data:      map[string]interface{}{"email": "jane.doe@example.com"},
		Upsert:    &common.UpsertOptions{UniqueColumns: []string{"email"}, UpdateType: common.UpdateTypeUpdate},
	},
}}
response, err := vault.Insert(ctx, request, common.InsertOptions{})
```

Sample response:

```json
{
  "records": [
    {
      "tableName": "cards",
      "skyflowId": "9fac9201-7b8a-4446-93f8-5244e1213bd1",
      "tokens": {
        "card_number": [ { "token": "5484-7829-1702-9110", "tokenGroupName": "card_number_tg" } ]
      },
      "data": { "card_number": "4111111111111111", "cardholder_name": "john doe" },
      "hashedData": { "card_number": [ { "data": "b6e6d...c3f9", "hashName": "hash1" } ] },
      "httpCode": 200,
      "error": "",
      "requestId": ""
    }
  ]
}
```

Records come back in the order you submitted them. `RequestId` is empty on success and the failing call's `x-request-id` on error. `Tokens` maps each column to a **slice** of `common.Token`, one entry per token group configured on that column, each with `Token`, `TokenGroupName` and `Path` (the location within a structured column's value the token came from, for example `"phone_numbers[0].type"`; empty for a flat column). `HashedData` maps each column to a slice of `common.HashedValue` (`Data`, `HashName`).

```go
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.SkyflowId, record.Tokens)
	} else {
		fmt.Println("insert failed", record.HttpCode, record.Error)
	}
}
```

Fields on each `common.InsertResponseRecord`: `TableName`, `SkyflowId`, `Tokens`, `Data`, `HashedData`, `HttpCode`, `Error`, `RequestId`.

[Example](../samples/flowvault/vaultapi/insert_records.go)

# Detokenize

Detokenize tokens in a single API call, optionally overriding the redaction applied per token group with `TokenGroupRedactions`.

> **Vault type supported:** both. See [Vault type support](#vault-type-support).

**Note:**

- `Tokens` is required and must not be empty, and no entry may be blank.
- `TokenGroupRedactions` is optional; each `common.TokenGroupRedaction` needs a non-blank `TokenGroupName` and `Redaction`.

```go
request := common.DetokenizeRequest{
	Tokens: []string{"5479-4229-4622-1393", "<TOKEN_1>"},
	TokenGroupRedactions: []common.TokenGroupRedaction{
		{TokenGroupName: "card_number_tg", Redaction: "MASKED"},
	},
}
response, err := vault.Detokenize(ctx, request, common.DetokenizeOptions{})
```

Sample response:

```json
{
  "records": [
    {
      "token": "5479-4229-4622-1393",
      "value": "4111111111111111",
      "tokenGroupName": "card_number_tg",
      "metadata": { "skyflowId": "9fac9201-7b8a-4446-93f8-5244e1213bd1", "tableName": "table1" },
      "httpCode": 200,
      "error": "",
      "requestId": ""
    },
    {
      "token": "<TOKEN_1>",
      "value": null,
      "tokenGroupName": "",
      "metadata": null,
      "httpCode": 404,
      "error": "Token Not Found",
      "requestId": "a1b2c3d4-..."
    }
  ]
}
```

`Value` is an `interface{}`. `Metadata` is a `*common.DetokenizeRecordMetadata` with `SkyflowId` and `TableName` (`nil` on records that failed).

```go
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.Token, "->", record.Value, "(", record.TokenGroupName, ")")
	} else {
		fmt.Println(record.Token, "failed", record.HttpCode, record.Error)
	}
}
```

Fields on each `common.DetokenizeResponseRecord`: `Token`, `Value`, `TokenGroupName`, `Metadata`, `HttpCode`, `Error`, `RequestId`.

[Example](../samples/flowvault/vaultapi/detokenize_records.go)

# Get

Read records back from a table by Skyflow ID or by unique value, optionally overriding the redaction applied per column with `ColumnRedactions`.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- A `common.GetRequest` works in one of two modes, which are mutually exclusive: **single-table** (`TableName`, `SkyflowIds`/`UniqueValues`, `Columns`, `ColumnRedactions`, `Limit`, `Offset`) or **multi-table** (`Records`, a slice of `common.GetRequestRecord`). Setting fields from both fails validation.
- `TableName` is required, and exactly one of `SkyflowIds` or `UniqueValues` must be set; both, or neither, fails validation. This holds per record in multi-table mode.
- `UniqueValues` is a `[]map[string]interface{}`: one map per record, holding the unique column-name/value pairs that identify it.
- `Columns` selects the columns to return; leave it nil or empty for all of them.
- `Limit` and `Offset` apply to the call as a whole, are **only used in single-table mode**, and must not be negative. `0` means "not sent".
- Each `common.ColumnRedaction` needs a non-blank `ColumnName` and `Redaction`.

```go
// Single table, by Skyflow ID
response, err := vault.Get(ctx, common.GetRequest{
	TableName:        "table1",
	SkyflowIds:       []string{"9fac9201-7b8a-4446-93f8-5244e1213bd1"},
	Columns:          []string{"card_number", "cardholder_name"},
	ColumnRedactions: []common.ColumnRedaction{{ColumnName: "card_number", Redaction: "MASKED"}},
	Limit:            10,
}, common.GetOptions{})

// Single table, by unique value
response, err = vault.Get(ctx, common.GetRequest{
	TableName:    "table2",
	UniqueValues: []map[string]interface{}{{"email": "jane.doe@example.com"}},
}, common.GetOptions{})

// Several tables in one call: each GetRequestRecord carries its own table and lookup fields
response, err = vault.Get(ctx, common.GetRequest{Records: []common.GetRequestRecord{
	{TableName: "table1", SkyflowIds: []string{"<SKYFLOW_ID>"}, Columns: []string{"card_number"}},
	{TableName: "table2", UniqueValues: []map[string]interface{}{{"email": "jane.doe@example.com"}}},
}}, common.GetOptions{})
```

Sample response:

```json
{
  "records": [
    {
      "tableName": "table1",
      "skyflowId": "9fac9201-7b8a-4446-93f8-5244e1213bd1",
      "tokens": {
        "card_number": [ { "token": "5484-7829-1702-9110", "tokenGroupName": "card_number_tg" } ]
      },
      "data": { "card_number": "4111-XXXX-XXXX-1111", "cardholder_name": "John Doe" },
      "hashedData": null,
      "httpCode": 200,
      "error": "",
      "requestId": ""
    }
  ]
}
```

`common.GetResponseRecord` is the same type as `common.InsertResponseRecord`: `TableName`, `SkyflowId`, `Tokens`, `Data`, `HashedData`, `HttpCode`, `Error`, `RequestId`.

```go
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.SkyflowId, "->", record.Data)
	} else {
		fmt.Println(record.SkyflowId, "failed", record.HttpCode, record.Error)
	}
}
```

[Example](../samples/flowvault/vaultapi/get_records.go)

# Update

Update records in a table by Skyflow ID, in a single API call.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- `TableName` follows the same one-level rule as [Insert](#insert): set it on the request or on every record, never both.
- Every `common.UpdateRequestRecord` needs a non-blank `SkyflowId`.
- `Data` holds the columns to change; it must not be empty and no column name may be blank. `Tokens` is optional bring-your-own-token.
- `UpdateType` is a request-level setting: `common.UpdateTypeUpdate` (merge the given columns) or `common.UpdateTypeReplace` (overwrite the whole record). Leaving it empty sends no update type, which the vault treats as an update.

```go
request := common.UpdateRequest{
	TableName:  "table1",
	UpdateType: common.UpdateTypeUpdate,
	Records: []common.UpdateRequestRecord{
		{SkyflowId: "9fac9201-7b8a-4446-93f8-5244e1213bd1", Data: map[string]interface{}{"cardholder_name": "jane doe"}},
	},
}
response, err := vault.Update(ctx, request, common.UpdateOptions{})
```

Sample response:

```json
{
  "records": [
    {
      "tableName": "table1",
      "skyflowId": "9fac9201-7b8a-4446-93f8-5244e1213bd1",
      "tokens": {
        "cardholder_name": [ { "token": "<TOKEN_2>", "tokenGroupName": "deterministic_string" } ]
      },
      "data": { "cardholder_name": "Jane Doe" },
      "hashedData": null,
      "httpCode": 200,
      "error": "",
      "requestId": ""
    }
  ]
}
```

`common.UpdateResponseRecord` is the same type as `common.InsertResponseRecord`. Successes and failures come back in one `Records` slice, each with its own `HttpCode` and `Error`.

```go
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.SkyflowId, "updated")
	} else {
		fmt.Println(record.SkyflowId, "failed", record.HttpCode, record.Error)
	}
}
```

[Example](../samples/flowvault/vaultapi/update_record.go)

# Delete

Delete records from a table by Skyflow ID or unique value, in a single API call.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- `TableName` is required, and exactly one of `SkyflowIds` or `UniqueValues` must be set; both, or neither, fails validation.
- `UniqueValues` has the same shape as in [Get](#get): one map per record, holding the unique column-name/value pairs that identify it.

```go
request := common.DeleteRequest{
	TableName:  "table1",
	SkyflowIds: []string{"9fac9201-7b8a-4446-93f8-5244e1213bd1", "b2308e2a-c1f5-469b-97b7-1f193159399b"},
}
response, err := vault.Delete(ctx, request, common.DeleteOptions{})
```

Sample response:

```json
{
  "records": [
    { "skyflowId": "9fac9201-7b8a-4446-93f8-5244e1213bd1", "httpCode": 200, "error": "", "requestId": "" },
    { "skyflowId": "b2308e2a-c1f5-469b-97b7-1f193159399b", "httpCode": 404, "error": "Record Not Found", "requestId": "a1b2c3d4-..." }
  ]
}
```

`common.DeleteResponseRecord` is flatter than the insert-shaped records above: the vault returns no data, tokens or hashed data for a delete. Fields: `SkyflowId`, `HttpCode`, `Error`, `RequestId`.

```go
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.SkyflowId, "deleted")
	} else {
		fmt.Println(record.SkyflowId, "failed", record.HttpCode, record.Error)
	}
}
```

[Example](../samples/flowvault/vaultapi/delete_records.go)

# Query

Run a read-only SQL `SELECT` against the vault in a single API call.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- `Query` is required and must not be blank. Only `SELECT` is supported.
- A call returns at most 25 rows; page through more with the SQL `OFFSET` keyword.
- Returned values may be masked per your vault's policies; they are never tokens or file URLs.
- Unlike the other operations, `Query` has **no per-record results**: any failure comes back as the `*SkyflowError`, and `RequestId` is always set.

```go
request := common.QueryRequest{Query: "SELECT name, email FROM table1 WHERE name = 'john doe' LIMIT 25"}
response, err := vault.Query(ctx, request, common.QueryOptions{})
```

Sample response:

```json
{
  "fields": [
    { "skyflow_id": "9fac9201-7b8a-4446-93f8-5244e1213bd1", "name": "john doe", "email": "john@example.com" }
  ],
  "metadata": { "columns": ["skyflow_id", "name", "email"] },
  "requestId": "a1b2c3d4-..."
}
```

`common.QueryResponse` has `Fields` (a `[]map[string]interface{}`, one map of column to value per row), `Metadata` (`common.QueryResponseMetadata` with `Columns`) and `RequestId`.

```go
for _, row := range response.Fields {
	fmt.Println(row)
}
```

[Example](../samples/flowvault/vaultapi/query_records.go)

# Get Tokens

Look up the existing token for one or more values, one result per value, in a single API call.

> **Vault type supported:** both. See [Vault type support](#vault-type-support).

**Note:**

- `Records` is required and must be a non-empty slice of `common.GetTokensRequestRecord`.
- Each record needs a non-nil, non-blank `Value` and a non-blank `TokenGroupName`. Only **deterministic** token groups are supported.
- One result is returned per input value, in request order; duplicate inputs are not collapsed.

```go
request := common.GetTokensRequest{Records: []common.GetTokensRequestRecord{
	{Value: "john@example.com", TokenGroupName: "email_tg"},
	{Value: "jane@example.com", TokenGroupName: "email_tg"},
}}
response, err := vault.GetTokens(ctx, request, common.GetTokensOptions{})
```

Sample response:

```json
{
  "records": [
    { "value": "john@example.com",    "tokenGroupName": "email_tg", "token": "1R9kNnLOPM", "httpCode": 200, "error": "", "requestId": "" },
    { "value": "jane@example.com",    "tokenGroupName": "email_tg", "token": "",           "httpCode": 404, "error": "Token not found.", "requestId": "a1b2c3d4-..." }
  ]
}
```

Fields on each `common.GetTokensResponseRecord`: `Value`, `TokenGroupName`, `Token`, `HttpCode`, `Error`, `RequestId`.

```go
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.Value, "->", record.Token)
	} else {
		fmt.Println(record.Value, "failed", record.HttpCode, record.Error)
	}
}
```

[Example](../samples/flowvault/vaultapi/get_tokens.go)

# Upload Files

Upload one or more files to file columns. The SDK **orchestrates** this in two steps: it requests a signed upload URL per column, then uploads each file's bytes to its URL. You just supply the files.

The signed URL is a credential that expires after about 15 minutes. The SDK uses it straight away and never returns it to you.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- `Records` is required and must be a non-empty slice of `common.UploadFilesRequestRecord`; each needs a `TableName` and a non-empty `Columns` slice, with no column repeated.
- Each `common.UploadFilesRequestColumn` needs a `Column` name and **exactly one** file source: `FilePath`, `Base64` or `FileObject` (an `*os.File`, which you close). `Base64` also requires `FileName`. The content type is inferred from the file name's extension.
- Leave `SkyflowId` empty to create a new record (CREATE permission); set it to upload into an existing record (UPDATE permission).
- `FileName` is derived from `FilePath`/`FileObject` when empty; if it is still unset, the server generates one.
- Every file is checked before anything is uploaded, so a missing or unreadable file fails the whole call. Files are then uploaded **concurrently** (at most 5 at once) and streamed from disk. The uploads to the signed URLs are not retried and have no overall timeout. The final virus-scan status is asynchronous; read it later with `Get`.
- Errors:
  - An error status from the first step (requesting the signed URLs) fails the whole call with a `*SkyflowError`, and nothing is uploaded.
  - A record that fails in the first step comes back with its `Error` set, and every one of its columns is `SKIPPED` with that error.
  - A column the vault returned no signed URL for is `SKIPPED`, with no error.
  - A column whose upload fails is `FAILED`, with an error such as `"PUT failed: 403"`. Other columns and records are unaffected.
  - A record's `Error` stays empty when only some of its columns failed, so check each column's `UploadStatus`.

```go
file, openErr := os.Open("<PATH_TO_FILE>")
if openErr != nil {
	return
}
defer file.Close() // the caller closes a FileObject

request := common.UploadFilesRequest{Records: []common.UploadFilesRequestRecord{
	{
		TableName: "onboarding",
		Columns: []common.UploadFilesRequestColumn{
			{Column: "resumePDF", FilePath: "/path/to/resume.pdf"},
			{Column: "photoID", Base64: "<BASE64_CONTENT>", FileName: "photo.jpg"},
			{Column: "addressProof", FileObject: file},
		},
	},
}}
response, err := vault.UploadFiles(ctx, request, common.UploadFilesOptions{})
```

Sample response (partial success):

```json
{
  "records": [
    {
      "skyflowId": "2a62a1fd-...",
      "tableName": "onboarding",
      "columns": [
        { "column": "resumePDF",    "fileName": "resume.pdf", "uploadStatus": "UPLOADED", "error": "" },
        { "column": "photoID",      "fileName": "photo.jpg",  "uploadStatus": "FAILED",   "error": "PUT failed: 403" },
        { "column": "addressProof", "fileName": "proof.pdf",  "uploadStatus": "UPLOADED", "error": "" }
      ],
      "httpCode": 200,
      "error": "",
      "requestId": ""
    }
  ]
}
```

Fields on each `common.UploadFilesResponseRecord`: `SkyflowId`, `TableName`, `Columns`, `HttpCode`, `Error`, `RequestId`. Each entry in `Columns` is a `common.UploadFilesColumnResult` with `Column`, `FileName`, `UploadStatus` (`common.UploadStatusUploaded`, `common.UploadStatusFailed` or `common.UploadStatusSkipped`) and `Error`. A column is `SKIPPED` when its record failed or no signed URL came back for it. Compare `UploadStatus` with these constants rather than with strings.

```go
for _, record := range response.Records {
	if record.Error != "" {
		fmt.Println("record failed:", record.HttpCode, record.Error, record.RequestId) // every column is SKIPPED
		continue
	}
	for _, column := range record.Columns {
		switch column.UploadStatus {
		case common.UploadStatusUploaded:
			fmt.Println(column.Column, "uploaded")
		case common.UploadStatusFailed:
			fmt.Println(column.Column, "failed:", column.Error)
		case common.UploadStatusSkipped:
			fmt.Println(column.Column, "skipped: no signed URL came back")
		}
	}
}
```

[Example](../samples/flowvault/vaultapi/upload_files.go)

# Delete Files

Delete files from file columns of existing records, and remove the stored objects, in a single API call.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- `Records` is required and must be a non-empty slice of `common.DeleteFilesRequestRecord` (at most 20,000); each needs a `TableName` and a non-empty `Columns` slice of file-column names, with no column repeated.
- Each record sets **exactly one** of `SkyflowId` or `UniqueValues`; both, or neither, fails validation.
- `UniqueValues` has the same shape as in [Get](#get); each map may resolve to more than one record.
- It is a **hard delete** and can't be undone: afterwards `Get` no longer returns a file for that column. It works whatever the file's `fileStatus`: a `PENDING` upload is cancelled and its signed URL invalidated, a `READY` file's stored object is removed, and a `FAILED` or `SCAN_ERROR` file just has its metadata cleared.
- Record-level failures come back on the response. Whole-call failures come back as the `*SkyflowError`.

```go
request := common.DeleteFilesRequest{Records: []common.DeleteFilesRequestRecord{
	{TableName: "onboarding", Columns: []string{"resumePDF"}, SkyflowId: "<SKYFLOW_ID>"},
	{TableName: "onboarding", Columns: []string{"photoID"}, UniqueValues: []map[string]interface{}{{"email": "jane.doe@example.com"}}},
}}
response, err := vault.DeleteFiles(ctx, request, common.DeleteFilesOptions{})
```

Sample response (partial success):

```json
{
  "records": [
    {
      "skyflowId": "9fac9201-...",
      "tableName": "onboarding",
      "columns": [ { "column": "resumePDF", "status": "DELETED" } ],
      "httpCode": 200,
      "error": "",
      "requestId": ""
    },
    {
      "skyflowId": "invalid-id",
      "tableName": "onboarding",
      "columns": null,
      "httpCode": 404,
      "error": "Invalid request. skyflowID invalid-id is invalid.",
      "requestId": "a1b2c3d4-..."
    }
  ]
}
```

Fields on each `common.DeleteFilesResponseRecord`: `SkyflowId`, `TableName`, `Columns`, `HttpCode`, `Error`, `RequestId`. On a successful record `Columns` is a slice of `common.DeleteFilesColumnResult` (`Column`, `Status`: `"DELETED"`); on a failed record it is `nil`.

```go
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.SkyflowId, "deleted", record.Columns)
	} else {
		fmt.Println(record.SkyflowId, "failed", record.HttpCode, record.Error)
	}
}
```

[Example](../samples/flowvault/vaultapi/delete_files.go)

# Custom Request Headers

To add custom HTTP headers to one call, set an `Interceptor` on that operation's options struct. The interceptor is a `func(rc *common.RequestContext)`; it runs once per call, just before the request is sent, and adds headers with `rc.AddHeader(key, value)`. `rc.Operation` names the operation (for example `"INSERT"`). The supported keys are `common.CustomHeaderKey` constants:

| `CustomHeaderKey` | HTTP header name |
|---|---|
| `common.SkyflowAccountId` | `x-skyflow-account-id` |
| `common.SkyflowAccountName` | `x-skyflow-account-name` |
| `common.RequestIdHeader` | `x-request-id` |

Any other key, or a blank value, fails the call with a `*SkyflowError` before it is sent.

[Example](../samples/flowvault/vaultapi/custom_header_example.go):

```go
options := common.InsertOptions{
	Interceptor: func(rc *common.RequestContext) {
		rc.AddHeader(common.RequestIdHeader, "req-001")
	},
}
response, err := vault.Insert(ctx, request, options)
```

Headers the interceptor adds are merged with the SDK's own headers (SDK metadata and `Authorization`).

The same pattern applies to every operation, through its options struct:

| Operation | Options struct |
|---|---|
| `Insert` | `common.InsertOptions` |
| `Detokenize` | `common.DetokenizeOptions` |
| `Get` | `common.GetOptions` |
| `Update` | `common.UpdateOptions` |
| `Delete` | `common.DeleteOptions` |
| `Query` | `common.QueryOptions` |
| `GetTokens` | `common.GetTokensOptions` |
| `UploadFiles` | `common.UploadFilesOptions` |
| `DeleteFiles` | `common.DeleteFilesOptions` |

# Error Handling

## Two layers of errors

This is the model to keep in mind for every operation:

| Layer | What it covers | How you see it |
|---|---|---|
| **Request level** | The call could not be made or the whole call failed: invalid request, missing credentials, auth failure, a cancelled context, or a whole-call API rejection. | A non-nil `*SkyflowError`, and a nil response. |
| **Record level** | The call succeeded, but individual records or tokens inside it did not. | A nil error and a response. Each entry in `response.Records` reports its own `HttpCode` and `Error`. |

The second layer is what sets `flowvault` apart from an all-or-nothing API: **a call that returns a nil error can still contain failures, and a call where every record failed still returns a nil error.** Checking only `err` will silently miss failed records, so always read the per-record results.

## Per-record success and failure

Every response (except `Query`) has `Records`, with one entry per submitted item in submitted order. Each entry carries:

| Field | Meaning |
|---|---|
| `HttpCode` | Per-item status (`int`). `2xx` for success; `4xx`/`5xx` for failure. |
| `Error` | Error message for this item. Empty means this item succeeded. |
| `RequestId` | The `x-request-id` of the call this item was part of, set only on failure. Quote it in support requests. |

The success payload sits alongside those fields on the same struct: `SkyflowId`/`Tokens`/`Data`/`HashedData` for `Insert`, `Get` and `Update`; `Value`/`TokenGroupName`/`Metadata` for `Detokenize`; `Value`/`TokenGroupName`/`Token` for `GetTokens`; `SkyflowId`/`TableName`/`Columns` for `UploadFiles` and `DeleteFiles`; and `SkyflowId` alone for `Delete`. (`Query` has no per-record results; any failure comes back as the `*SkyflowError`.)

The idiomatic way to consume a response:

```go
response, err := vault.Insert(ctx, request, common.InsertOptions{})
if err != nil {
	fmt.Println("call failed:", err.GetHttpCode(), err.GetMessage())
	return
}
for _, record := range response.Records {
	if record.Error == "" {
		fmt.Println(record.SkyflowId)
	} else {
		fmt.Println("failed", record.HttpCode, record.Error, "(request id", record.RequestId, ")")
	}
}
```

## Handling SkyflowError

`*SkyflowError` (from `flowvault/utils/error`) covers the request-level layer only: client-side validation errors and whole-call API errors. It is the same type the SkyVault SDK returns, from the shared `common` module.

Every method returns `*SkyflowError`, not `error`. Check it with `err != nil` directly. Don't assign it to a variable of type `error` first: a nil `*SkyflowError` stored in an `error` is not `== nil`.

```go
response, err := vault.Insert(ctx, request, common.InsertOptions{})
if err != nil {
	fmt.Println("HTTP code :", err.GetHttpCode())
	fmt.Println("Message   :", err.GetMessage())
	fmt.Println("Request ID:", err.GetRequestId())
	fmt.Println("Details   :", err.GetDetails())
	return
}
// reaching here means the CALL succeeded; individual records may still have failed
```

## SkyflowError methods

| Method | Returns | Description |
|---|---|---|
| `GetHttpCode()` | `string` | HTTP status code, for example `"400"`, `"404"`, `"500"`. `"499"` when the call's context ended while it waited for a bearer token. |
| `GetMessage()` | `string` | Human-readable description of the error. |
| `GetGrpcCode()` | `string` | gRPC status code from the server. |
| `GetHttpStatusCode()` | `string` | HTTP status text from the server. |
| `GetRequestId()` | `string` | The `x-request-id` response header. Useful for support requests. |
| `GetDetails()` | `[]interface{}` | Additional error context from the server. Empty for validation errors. |
| `GetResponseBody()` | `map[string]interface{}` | The parsed error response body, when there is one. |
| `Error()` | `string` | Implements `error`. |
| `Unwrap()` | `error` | The underlying error when the SDK kept one (for example the context's error for a `"499"`), so `errors.Is` / `errors.As` work. |

**Validation errors** (table name on both levels, empty token list and similar) are returned before any network call, with HTTP code `"400"`. **API errors** come from the Skyflow server, with the fields filled in from the response body and headers.

## Retrying the failed records

Because failures are reported per record, a partial failure can be retried without resubmitting the whole payload. Resend only the records worth retrying: **server-side failures (HTTP 500–599), excluding 529**, which is a permanent capacity-limit code. Client-side (`4xx`) failures are left out on purpose; they need a fix to the data, not a retry. This is separate from the `MaxRetries` setting in [Timeouts and retries](#timeouts-and-retries), which retries whole HTTP attempts and is off by default.

Records come back in input order, so you can filter `response.Records` yourself with this predicate:

```go
func isRetryable(httpCode int) bool {
	return httpCode >= 500 && httpCode <= 599 && httpCode != 529
}
```

For `Insert` (and likewise `Update`, `Get` and `Delete`), match each response record to your input by position, then resubmit the retryable ones:

```go
request := common.InsertRequest{TableName: "cards", Records: records}
response, err := vault.Insert(ctx, request, common.InsertOptions{})
if err != nil {
	return
}

var retryable []common.InsertRequestRecord
for i, record := range response.Records {
	if isRetryable(record.HttpCode) {
		retryable = append(retryable, request.Records[i])
	}
}
if len(retryable) > 0 {
	retryResponse, err := vault.Insert(ctx, common.InsertRequest{TableName: "cards", Records: retryable}, common.InsertOptions{})
	if err == nil {
		fmt.Println(retryResponse.Records) // handle as above
	}
}
```

> **Make insert retries idempotent.** A record can succeed on the server even when the response never reaches you (a dropped connection, or a `5xx` returned after the row was written), so a blind resubmit can create a duplicate. Guard against it with `Upsert` on a unique column, set at the same level as `TableName`: a retried record then updates the existing row instead of inserting a second one.

For `Detokenize`, each response record carries its own `Token`, so filter on that directly, with no position bookkeeping:

```go
response, err := vault.Detokenize(ctx, request, common.DetokenizeOptions{})
if err != nil {
	return
}

var retryTokens []string
for _, record := range response.Records {
	if isRetryable(record.HttpCode) {
		retryTokens = append(retryTokens, record.Token)
	}
}
if len(retryTokens) > 0 {
	retryResponse, err := vault.Detokenize(ctx, common.DetokenizeRequest{Tokens: retryTokens}, common.DetokenizeOptions{})
	if err == nil {
		fmt.Println(retryResponse.Records) // handle as above
	}
}
```

# Samples

Runnable examples live in [samples/flowvault](../samples/flowvault/): one file per vault operation, plus custom-header, timeout/retry and service account examples. See [samples/flowvault/README.md](../samples/flowvault/README.md) to run them.
