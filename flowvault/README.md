# Skyflow FlowVault Go SDK

The `flowvault` module is the Skyflow Go SDK for **FlowVault** vaults. It shares its credentials, configuration, logger and error types with the [SkyVault Go SDK](../README.md) (both depend on the `common` module) but exposes a different, narrower surface: **unary** vault operations (insert, get, update, delete, detokenize, query, get tokens, upload files and delete files) plus Detect.

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
- [Detect](#detect)
  - [Deidentify String](#deidentify-string)
  - [Deidentify File](#deidentify-file)
  - [Get Run](#get-run)
  - [Reidentify String](#reidentify-string)
  - [Reidentify File](#reidentify-file)
  - [Check Guardrails](#check-guardrails)
  - [Detect configuration reference](#detect-configuration-reference)
  - [Detect request headers](#detect-request-headers)
- [Error Handling](#error-handling)
  - [Two layers of errors](#two-layers-of-errors)
  - [Per-record success and failure](#per-record-success-and-failure)
  - [Handling SkyflowError](#handling-skyflowerror)
  - [SkyflowError methods](#skyflowerror-methods)
  - [Retrying the failed records](#retrying-the-failed-records)
  - [Detect run status](#detect-run-status)
  - [Common Detect errors](#common-detect-errors)
- [Samples](#samples)

# Overview

- Authenticate with a Skyflow service account, an API key or a bearer token. See [Authenticate](#authenticate).
- Perform vault operations: insert, get, update, delete, detokenize, query, get tokens, upload files and delete files. See [Vault operations](#vault-operations).
- De-identify and re-identify strings and files, and check text against guardrails, with Detect. See [Detect](#detect).
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
	"github.com/skyflowapi/skyflow-go/flowvault/client"         // Skyflow client, VaultService, DetectService
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

`Vault("<VAULT_ID>")` returns the `*client.VaultService` for a registered vault. `Vault()` with no argument returns the first vault registered. To talk to more than one vault from one client, pass each config to `WithVaults(...)` (or add it later with `AddVaultConfig`) and fetch each service by ID.

# Authenticate

Requests are authorized with Skyflow credentials, set as a `common.Credentials` value. Credential handling comes from the shared `common` module, so it works the same way as in the SkyVault SDK.

## Credential types

Set **exactly one** of the following fields. Setting more than one, or none, fails validation with a `*SkyflowError`, and so does a field that holds only spaces. Credentials are checked when you add them.

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
| `Context` | Attach context to the generated token for context-aware authorization: a non-blank `string`, or a non-empty `map[string]interface{}` whose keys contain only letters, digits and underscores. Any other type fails validation. |

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
| `ConnectTimeout` | `time.Duration` | optional | Per-attempt connection timeout. |
| `ReadTimeout` | `time.Duration` | optional | Per-attempt wait for the server to start responding. |
| `WriteTimeout` | `time.Duration` | optional | Per-attempt limit on each write of the request. |
| `MaxRetries` | `*int` | optional | Retry attempts after the first failure. |
| `InitialRetryDelay` | `time.Duration` | optional | Backoff before the first retry. |
| `MaxRetryDelay` | `time.Duration` | optional | Ceiling the exponential backoff grows to. |
| `ConnectTimeout` | `time.Duration` | optional | Per attempt: bound on connecting, including the TLS handshake. |
| `ReadTimeout` | `time.Duration` | optional | Per attempt: bound on waiting for the response once the request is sent. |
| `WriteTimeout` | `time.Duration` | optional | Per attempt: bound on each write of the request that makes no progress. |

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
| `client.WithVaults(configs ...common.VaultConfig)` | Register one or more vaults. The first one registered is what `Vault()` (no argument) returns. Fails if called with no configs. |
| `client.WithCredentials(credentials)` | Client-wide credentials for vaults that don't set their own. |
| `client.WithLogLevel(level)` | `logger.DEBUG`, `logger.INFO`, `logger.WARN`, `logger.ERROR` (default) or `logger.OFF`. |
| `client.WithHttpConfig(common.HttpConfig{...})` | Client-wide timeouts and retry policy. Same fields as the HTTP fields on `VaultConfig`. |

Methods on `*client.Skyflow`:

| Method | Description |
|---|---|
| `Vault(vaultId ...string)` | The `*VaultService` for a registered vault. With no ID, or `""`, it returns the default vault, the first one added. |
| `Detect(vaultId ...string)` | The `*DetectService` for a registered vault, or the default vault when no ID is given. See [Detect](#detect). |
| `AddVaultConfig(config)` | Register another vault. Fails if the vault ID is already registered. |
| `UpdateVaultConfig(config)` | Update a registered vault in place. Only the fields you set change, except `Env`, which is always applied (leaving it out moves the vault to `PROD`). Setting `ClusterId` without a URL clears a stored `BaseVaultUrl`, so the vault moves to that cluster. |
| `RemoveVaultConfig(vaultId)` | Unregister a vault. |
| `GetVaultConfig(vaultId)` | A copy of a registered vault's config; `""` means the default vault. |
| `AddSkyflowCredentials(credentials)` / `UpdateSkyflowCredentials(credentials)` | Set or replace the client-wide credentials. |
| `GetSkyflowCredentials()` | A copy of the client-wide credentials, or `nil`. |
| `SetLogLevel(level)` / `GetLogLevel()` | Change or read the log level. `UpdateLogLevel` and `GetLoglevel` are deprecated aliases. |

Each of these that can fail returns a `*SkyflowError`.

## Timeouts and retries

Each HTTP setting resolves **most specific first**: the value on the `VaultConfig`, else the client-wide value from `WithHttpConfig`, else the SDK default. Negative values fail validation.

| Setting | SDK default |
|---|---|
| `Timeout` (overall call, retries included) | 60 s |
| `ConnectTimeout` / `ReadTimeout` / `WriteTimeout` (per attempt) | 10 s |
| `MaxRetries` | `0`: retries are **opt-in**, so non-idempotent writes are never replayed silently |
| `InitialRetryDelay` | 500 ms |
| `MaxRetryDelay` | 2000 ms |
| `ConnectTimeout`, `ReadTimeout`, `WriteTimeout` (each attempt) | 10 s each |

The per-attempt timeouts also apply to the file uploads `UploadFiles` sends to signed URLs, which have no overall timeout.

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
		Timeout:     300 * time.Second, // per-vault overrides
		ReadTimeout: 30 * time.Second,
	}),
)
```

When retries are enabled, retryable responses (HTTP `408`, `429` and `5xx`) are retried with exponential backoff and jitter, bounded by `InitialRetryDelay`/`MaxRetryDelay` and the overall `Timeout`. `UploadFiles` is the exception: it is not retried on a `5xx`, because the vault may already have created the records before failing. It is still retried on `408` and `429`. A deadline on the `context.Context` you pass to a call also applies when it is shorter. The file uploads in [Upload Files](#upload-files) are not retried and have no overall timeout, so large files aren't cut off.

## Logging

The SDK logs at `logger.ERROR` by default, to standard error. Set the level with `client.WithLogLevel(logger.DEBUG)` or `skyflowClient.SetLogLevel(logger.DEBUG)`, or turn logging off with `logger.OFF`. Send the output elsewhere with `logger.SetOutput(w io.Writer)`.

The level and output are **process-wide**: they are shared by every client of the FlowVault and SkyVault Go SDKs in the process.

## Concurrency and resource lifecycle

- **Concurrency**: `*client.Skyflow` and the `*VaultService` / `*DetectService` values it returns are safe for concurrent use. Vaults and client settings can be changed while calls are running; a call that is already running keeps the config it started with, and a `VaultService` obtained earlier sees the change on its next call.
- **Lifecycle**: create the client once and reuse it. The underlying HTTP client and the cached bearer token are reused across calls rather than recreated per request. There is nothing to close.
- **Cancellation**: every call takes a `context.Context`. If the context is cancelled or its deadline passes, the call stops and returns a `*SkyflowError`. The error has HTTP code `499` and unwraps to the context's error, so `errors.Is(err, context.Canceled)` and `errors.Is(err, context.DeadlineExceeded)` work.

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

Pass the zero value of the options struct (for example `common.InsertOptions{}`) when you don't need it. Its only field is `CustomHeaders`; see [Custom Request Headers](#custom-request-headers).

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

- `Records` must not be empty, and each record's `Data` must not be empty. No column name or value may be blank or nil.
- `TableName` and `Upsert` follow a one-level rule: set them on the request (applies to every record) or on every record, never both.
- `Upsert` takes a `*common.UpsertOptions` with non-empty `UniqueColumns` and an optional `UpdateType`: `common.UpdateTypeUpdate` (merge) or `common.UpdateTypeReplace` (overwrite).
- `Tokens` is optional bring-your-own-token; when set, the map must not be empty and no key or value may be blank.

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

Records come back in the order you submitted them. `RequestId` is empty on success and the failing call's `x-request-id` on error. `Tokens` maps each column to a **slice** of `common.Token`, one entry per token group configured on that column, each with `Token`, `TokenGroupName` and `Path` (the location within a structured column's value the token came from, for example `"phone_numbers[0].type"`; empty for a flat column). `HashedData` is the vault's `hashedData` passed through as returned: a `map[string]interface{}` from column name to that column's hashes (typically a list of `{"data", "hashName"}` objects).

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
- `TokenGroupRedactions` is optional; each `common.TokenGroupRedactions` needs a non-blank `TokenGroupName` and `Redaction`.

```go
request := common.DetokenizeRequest{
	Tokens: []string{"5479-4229-4622-1393", "<TOKEN_1>"},
	TokenGroupRedactions: []common.TokenGroupRedactions{
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

`Value` is an `interface{}`. `Metadata` is a `*common.DetokenizeResponseRecordMetadata` with `SkyflowId` and `TableName` (`nil` on records that failed).

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

- A `common.GetRequest` works in one of two modes, which are mutually exclusive: **single-table** (`TableName`, `SkyflowIds`/`UniqueValues`, `Columns`, `ColumnRedactions`, `Limit`, `Offset`) or **multi-table** (`Records`, a slice of `common.GetRequestRecord`). Setting fields from both fails validation. An empty slice counts as not set, so an empty `Records` makes the request single-table. `Limit` and `Offset` aren't sent in multi-table mode, so setting them with `Records` fails validation rather than being ignored.
- `TableName` is required, and exactly one of `SkyflowIds` or `UniqueValues` must be set; both, or neither, fails validation. This holds per record in multi-table mode.
- `UniqueValues` is a `[]map[string]interface{}`: one map per record, holding the unique column-name/value pairs that identify it.
- `Columns` selects the columns to return; leave it nil or empty for all of them.
- `Limit` and `Offset` apply to the call as a whole, are **only used in single-table mode**, and must not be negative. They are `*int`: `nil` means "not sent", and an explicit `0` is sent.
- Each `common.ColumnRedactions` needs a non-blank `ColumnName` and `Redaction`.

```go
// Single table, by Skyflow ID
limit := 10
response, err := vault.Get(ctx, common.GetRequest{
	TableName:        "table1",
	SkyflowIds:       []string{"9fac9201-7b8a-4446-93f8-5244e1213bd1"},
	Columns:          []string{"card_number", "cardholder_name"},
	ColumnRedactions: []common.ColumnRedactions{{ColumnName: "card_number", Redaction: "MASKED"}},
	Limit:            &limit,
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
- `Data` holds the columns to change; it must not be empty and no key or value may be blank. `Tokens` is optional bring-your-own-token.
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
- A call returns at most 25 rows; page through more with the SQL `LIMIT` and `OFFSET` keywords. When a call returns 25 rows, the SDK logs a warning that more rows may match.
- Returned values may be masked per your vault's policies; they are never tokens or file URLs.
- Unlike the other operations, `Query` has **no per-record results**: any failure comes back as the `*SkyflowError`, and `RequestId` is always set.

```go
request := common.QueryRequest{Query: "SELECT name, email FROM table1 WHERE name = 'john doe' LIMIT 25"}
response, err := vault.Query(ctx, request, common.QueryOptions{})
```

Sample response:

```json
{
  "records": [
    { "skyflow_id": "9fac9201-7b8a-4446-93f8-5244e1213bd1", "name": "john doe", "email": "john@example.com" }
  ],
  "metadata": { "columns": ["skyflow_id", "name", "email"] },
  "requestId": "a1b2c3d4-..."
}
```

`common.QueryResponse` has `Records` (a `[]map[string]interface{}`, one map of column to value per row, never nil), `Metadata` (`common.QueryResponseMetadata` with `Columns`) and `RequestId`. `Fields` holds the same rows but is deprecated, because these are rows, not fields; use `Records`.

```go
for _, row := range response.Records {
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
- Each `common.UploadFilesRequestColumn` needs a `Column` name and **exactly one** file source: `FilePath`, `Base64` or `FileObject` (an `*os.File`, which you close). `Base64` also requires `FileName`. The content type comes from the file name's extension, using the SDK's own list, so a file gets the same type on every OS. An unknown or missing extension is sent as `application/octet-stream`.
- Leave `SkyflowId` empty to create a new record (CREATE permission); set it to upload into an existing record (UPDATE permission). A `SkyflowId` of only spaces fails validation rather than creating a new record.
- `FileName` is derived from `FilePath`/`FileObject` when empty; if it is still unset, the server generates one.
- Every file is checked before anything is uploaded, so a missing or unreadable file, or invalid base64, fails the whole call. Files are then uploaded **concurrently** (at most 5 at once) and streamed: nothing is read into memory up front, and base64 content is decoded while it is uploaded. A `FileObject` is uploaded whole, from its start, whatever its current offset, and its offset is left unchanged, so one `*os.File` can back several columns. The uploads to the signed URLs are not retried and have no overall timeout. The final virus-scan status is asynchronous; read it later with `Get`.
- Errors:
  - An error status from the first step (requesting the signed URLs) fails the whole call with a `*SkyflowError`, and nothing is uploaded, even when the response lists records. When those records carry errors, the message is the distinct record errors joined with `; `.
  - A record that fails in the first step comes back with its `Error` set, and every one of its columns is `SKIPPED` with that error.
  - A column the vault returned no signed URL for is `SKIPPED`, with the error `"No signed URL was returned for this column; file was not uploaded."`.
  - A column whose upload fails is `FAILED`, with an error such as `"PUT failed: 403"`. Other columns and records are unaffected.
  - When any column of a record was not uploaded, the record's `Error` is set to `"<n> of <m> files were not uploaded. See each column for details."` and its `RequestId` is filled in. Its `HttpCode` stays the one the vault returned.
  - Results are matched to your records by position. If the vault returns a different number of records, or a record for another table (or another `SkyflowId`, when you set one), nothing is uploaded for the records affected and they come back with the error `"Upload response did not match the request; file was not uploaded."`. Two new records in the same table can't be told apart, so if the vault ever returned those in a different order their files would be reported against each other.

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
      "error": "1 of 3 files were not uploaded. See each column for details.",
      "requestId": "a1b2c3d4-..."
    }
  ]
}
```

Fields on each `common.UploadFilesResponseRecord`: `SkyflowId`, `TableName`, `Columns`, `HttpCode`, `Error`, `RequestId`. Each entry in `Columns` is a `common.UploadFilesColumnResult` with `Column`, `FileName`, `UploadStatus` (`common.UploadStatusUploaded`, `common.UploadStatusFailed` or `common.UploadStatusSkipped`) and `Error`. A column is `SKIPPED` when its record failed or no signed URL came back for it. Compare `UploadStatus` with these constants rather than with strings.

A record's `Error` is set whenever any of its files wasn't uploaded; check each column to see which:

```go
for _, record := range response.Records {
	if record.Error != "" {
		fmt.Println("record:", record.HttpCode, record.Error, record.RequestId)
	}
	for _, column := range record.Columns {
		switch column.UploadStatus {
		case common.UploadStatusUploaded:
			fmt.Println(column.Column, "uploaded")
		case common.UploadStatusFailed, common.UploadStatusSkipped:
			fmt.Println(column.Column, column.UploadStatus, column.Error)
		}
	}
}
```

[Example](../samples/flowvault/vaultapi/upload_files.go)

# Delete Files

Delete files from file columns of existing records, and remove the stored objects, in a single API call.

> **Vault type supported:** structured (schema) vaults. See [Vault type support](#vault-type-support).

**Note:**

- `Records` is required and must be a non-empty slice of `common.DeleteFilesRequestRecord`; each needs a `TableName` and a non-empty `Columns` slice of file-column names, with no column repeated.
- Each record sets **exactly one** of `SkyflowId` or `UniqueValues`; both, or neither, fails validation. A `SkyflowId` of only spaces fails validation too.
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
      "columns": [ { "column": "resumePDF", "status": "DELETED", "rawStatus": "DELETED" } ],
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

Fields on each `common.DeleteFilesResponseRecord`: `SkyflowId`, `TableName`, `Columns`, `HttpCode`, `Error`, `RequestId`. On a successful record `Columns` is a slice of `common.DeleteFilesColumnResult` with `Column`, `Status` and `RawStatus`; on a failed record it is `nil`. `Status` is a `common.DeleteFilesStatus`: `common.DeleteFilesStatusDeleted`, or `common.DeleteFilesStatusUnknown` when the vault reported a status this SDK version does not recognize (or none). `RawStatus` is the status exactly as the vault sent it.

When the whole call fails but its error body still lists per-record results, those records are returned instead of an error. Such a record counts as failed unless it reports a `2xx` `HttpCode` of its own; one that reports neither an error nor a status gets the error `"Unknown error"` and the call's status. The same applies to [Get Tokens](#get-tokens).

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

To add custom HTTP headers to one call, set `CustomHeaders` on that operation's options struct. It is a `map[common.CustomHeaderKey]string`, and its keys are these constants:

| `CustomHeaderKey` | HTTP header name |
|---|---|
| `common.SkyflowAccountId` | `x-skyflow-account-id` |
| `common.SkyflowAccountName` | `x-skyflow-account-name` |
| `common.RequestIdHeader` | `x-request-id` |

Any other key, or a blank value, fails the call with a `*SkyflowError` before it is sent. The error message names the operation, for example `INSERT`.

[Example](../samples/flowvault/vaultapi/custom_header_example.go):

```go
options := common.InsertOptions{
	CustomHeaders: map[common.CustomHeaderKey]string{
		common.RequestIdHeader: "req-001",
	},
}
response, err := vault.Insert(ctx, request, options)
```

These headers are sent along with the SDK's own headers (SDK metadata and `Authorization`).

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

# Detect

Skyflow Detect finds sensitive data in text and in files, including documents, spreadsheets, images and audio, and replaces it with tokens or placeholders. Vault tokens can be re-identified later. It also screens text with guardrails for toxicity and denied topics. Get a vault's detect service from the client. Every method takes a context, a request and an options struct, and returns a response and a `*SkyflowError`.

```go
service, err := skyflowClient.Detect("<VAULT_ID>") // a vault from the client config
if err != nil {
	fmt.Println(err.GetMessage())
	return
}
```

| Method | Request | Options | Response |
|---|---|---|---|
| `DeidentifyString` | `DeidentifyStringRequest` | `DeidentifyStringOptions` | `DeidentifyStringResponse` |
| `DeidentifyFile` | `DeidentifyFileRequest` | `DeidentifyFileOptions` | `DeidentifyFileResponse` |
| `GetRun` | `GetRunRequest` | `GetRunOptions` | `GetRunResponse` |
| `ReidentifyString` | `ReidentifyStringRequest` | `ReidentifyStringOptions` | `ReidentifyStringResponse` |
| `ReidentifyFile` | `ReidentifyFileRequest` | `ReidentifyFileOptions` | `ReidentifyFileResponse` |
| `CheckGuardrails` | `CheckGuardrailsRequest` | `CheckGuardrailsOptions` | `CheckGuardrailsResponse` |

All types are in `github.com/skyflowapi/skyflow-go/flowvault/utils/common`.

**Before you start**, have these ready:
- **A vault ID and cluster ID.** The vault ID is on the vault's details page in Skyflow Studio. The cluster ID is the first part of the vault URL: `https://{clusterId}.vault.skyflowapis.com`.
- **A role.** The service account needs the Detect Invoker, Detect Manager or Vault Owner role on the vault. Detect Invoker can de-identify data and write vault tokens; Detect Manager can also create, update and delete Detect configurations. Re-identifying vault tokens works like detokenization: the caller also needs a policy that grants the Detokenization action on the destination, unless it is a Vault Owner.
- **A configuration**, unless you pass one inline. A stored configuration is created in Studio or with the [Management API](https://docs.skyflow.com/management/), and its ID goes in `ConfigurationId`; the SDK doesn't create configurations. See [De-identification](https://docs.skyflow.com/docs/de-identification).
- **A token group**, if you use `DeidentificationTypeVaultToken`. Its name goes in the entity's `Destination`. The token group must already exist in the vault: Detect writes into it but doesn't create it.
- **A redaction pattern**, if you want re-identified values masked rather than in full. It is also configured in the vault, and its name goes in `RedactionPattern`.

**Tokens in a FlowVault vault.** A value de-identified with `DeidentificationTypeVaultToken` is stored in a **token group**, named in the entity's `Destination`. Re-identification looks tokens up in their token group and can render them through a **redaction pattern**. Both token groups and redaction patterns are configured in the vault. The other de-identification types (`ENTITY_UNIQUE_COUNTER`, `ENTITY_ONLY`, `MASK`) produce placeholders such as `[NAME_1]` that are not stored, so they can't be re-identified.

## Deidentify String
`DeidentifyString` replaces sensitive data in `Text`. Choose the configuration with:

- `ConfigurationId`: a stored configuration, created in Studio or with the [Management API](https://docs.skyflow.com/management/).
- `Configuration`: an inline `DetectConfiguration` (see the [Detect configuration reference](#detect-configuration-reference)).

Both are optional and sent as given. With neither, the SDK sends only the vault ID and the run uses the vault's default Detect settings. Set only one: the API rejects a request with both (HTTP 400).

```go
// With a stored configuration
response, err := service.DeidentifyString(ctx, common.DeidentifyStringRequest{
	Text:            "My name is John Doe, and my email is johndoe@acme.com.",
	ConfigurationId: "<CONFIGURATION_ID>",
}, common.DeidentifyStringOptions{})

// ...or with an inline configuration. Optional numbers and flags are pointers (set them with common.Ptr): nil is left to the server.
response, err = service.DeidentifyString(ctx, common.DeidentifyStringRequest{
	Text: "My name is John Doe, and my email is johndoe@acme.com.",
	Configuration: &common.DetectConfiguration{
		Detect: &common.Detect{
			Entities: []common.Entity{
				// Names become vault tokens in a token group, so they can be re-identified
				{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TOKEN_GROUP_NAME>"},
				{
					EntityType:           common.EntityTypeDOB,
					DeidentificationType: common.DeidentificationTypeEntityUniqueCounter,
					Transformation:       &common.Transformation{ShiftDates: &common.ShiftDates{MinDays: common.Ptr(10), MaxDays: common.Ptr(30)}},
				},
				{EntityType: common.EntityTypeAll, DeidentificationType: common.DeidentificationTypeEntityUniqueCounter},
			},
			Restrict:       []string{"[0-9]{3}-[0-9]{2}-[0-9]{4}"}, // always treat as sensitive
			Skip:           []string{"Skyflow"},                    // never treat as sensitive
			CustomTypes:    []common.CustomType{{Label: "EMPLOYEE_ID", Match: []string{"EMP-[0-9]{6}"}}},
			ReturnEntities: common.ReturnEntitiesTypeExcludeSensitiveData,
		},
	},
}, common.DeidentifyStringOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
fmt.Println(response.ProcessedText)
// Metrics and each of its counts are pointers, nil when the server leaves them out.
if response.Metrics != nil && response.Metrics.WordCount != nil {
	fmt.Println("Words:", *response.Metrics.WordCount)
}
```

Sample response (entity-counter placeholders):
```json
{
  "processedText": "My name is [NAME_1], and my email is [EMAIL_ADDRESS_1].",
  "entities": [
    {
      "token": "NAME_1",
      "value": "",
      "location": { "startIndex": 11, "endIndex": 19, "startIndexProcessed": 11, "endIndexProcessed": 19 },
      "entityType": "NAME",
      "entityScores": { "NAME": 0.95 }
    }
  ],
  "metrics": { "size": 0.05, "wordCount": 11, "characterCount": 54 }
}
```

`Value` is set only when the configuration's `ReturnEntities` is `ReturnEntitiesTypeAll`. `Location` is a `*common.EntityLocation`: nil when the server sends no location.

> [!TIP]
> See the full example in the samples directory: [deidentify_string.go](../samples/flowvault/detectapi/deidentify_string.go)

## Deidentify File
`DeidentifyFile` submits a file and returns its run ID. Supply the file as exactly one of:

| Field | Input |
|---|---|
| `FilePath` | A local file, sent as base64. The format is inferred from the extension, and `DataSource` is ignored. |
| `FileObject` | An open `*os.File`, sent the same way. The caller closes it. |
| `DataSource` + `Value` | `DeidentifySourceTypeBase64` with base64 content, or `DeidentifySourceTypePresignedURL` with a presigned URL. |

Before calling the API, the SDK checks that exactly one is set, that a `Value` has a `DataSource` and isn't blank, and that a local file exists and is a regular file. A local file is checked before the `Poll` options, and one that exists but can't be read fails with "Failed to encode the file". It sends `DataSource` as given; the server rejects one the vault doesn't accept.

`ConfigurationId` and `Configuration` work as for [Deidentify String](#deidentify-string). `DataFormat` is optional: the server infers it, and for a local file the SDK infers it from the extension. Set it to override either.

> [!WARNING]
> **`DeidentifyFile` doesn't wait for the result by default.** It returns as soon as the file is submitted, with only `RunId` set: `Status` is empty and there is no `Output`. Get the result later with [GetRun](#get-run), or set `Poll` to wait for the run.

Outputs are written to disk only when all three hold: the input is a local file (`FilePath` or `FileObject`), the polled run reached `SUCCESS`, and its outputs are base64. They are named `processed-<name>.<ext>`, `processed-entities-<name>.json` and `processed-object-entities-<name>.json`, where `<name>` is the input's name without its extension (a dotfile such as `.env` keeps its whole name). A file that can't be written fails the call with "Failed to write the processed output to <path>." and code `"0"`; the underlying error is `Unwrap()`. `OutputDirectory` requires `Poll`. It is created if missing and defaults to the current directory (an empty or blank value means the current directory). Each output goes to a hidden temporary file first and is then renamed into place, so a process that dies mid-write can leave a `.processed-*.tmp` file behind. Presigned-URL outputs are not fetched: read them from `Output`.

`DeidentificationTypeVaultToken` is supported only for `txt`, `csv`, `json`, `jsonl`, `xml`, `docx`, `xlsx` and `pptx` files.

```go
// Submit by presigned URL and return: only RunId is set
submitted, err := service.DeidentifyFile(ctx, common.DeidentifyFileRequest{
	DataSource:      common.DeidentifySourceTypePresignedURL,
	Value:           "<PRESIGNED_URL>",
	DataFormat:      common.FileDataFormatPDF, // optional
	ConfigurationId: "<CONFIGURATION_ID>",
}, common.DeidentifyFileOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
fmt.Println(submitted.RunId)

// Submit a local file and wait up to 2 minutes
response, err := service.DeidentifyFile(ctx, common.DeidentifyFileRequest{
	FilePath:        "<PATH_TO_FILE>/patient-notes.pdf",
	ConfigurationId: "<CONFIGURATION_ID>",
}, common.DeidentifyFileOptions{
	Poll:            &common.PollOptions{WaitTime: common.Ptr(2 * time.Minute)},
	OutputDirectory: "<OUTPUT_DIRECTORY>",
})
if err != nil {
	// A lookup or output write that fails mid-wait still returns the response, so the run ID isn't lost
	if response != nil {
		fmt.Println("Run ID:", response.RunId)
	}
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
switch response.Status {
case common.DetectRunStatusSuccess:
	fmt.Println("Outputs:", len(response.Output))
case common.DetectRunStatusInProgress, common.DetectRunStatusQueued: // the wait ran out
	fmt.Println("Still running, run ID:", response.RunId)
default: // FAILED or UNKNOWN
	fmt.Println("Run failed:", response.Message)
}
```

| `PollOptions` | Default | Limit |
|---|---|---|
| `WaitTime` | 60s | 1s–300s |
| `MaxAttempts` | 23 | 1–23 |

Polling backs off 1s, 2s, 4s … up to 32s between lookups. It stops at whichever comes first: `WaitTime`, `MaxAttempts`, or a final status (`SUCCESS`, `FAILED`, `UNKNOWN`). `WaitTime` counts the pauses between lookups, not the time each lookup takes, so a call can run a little longer than `WaitTime`. With this backoff at most 7 lookups fit in 60s and 15 in 300s, so `MaxAttempts` only stops polling earlier when it is set below that. `&common.PollOptions{}` uses the defaults: a field left nil uses its default, and a value that is set must be within the limit, so `0` is rejected (`common.Ptr(2 * time.Minute)` sets one).

Sample response (polled, successful):
```json
{
  "runId": "89699e2c-4301-4f9f-bcff-0a8fd3057489",
  "status": "SUCCESS",
  "outputType": "BASE64",
  "output": [
    { "processedFile": "TXkgU1NOIGlzIFtTU0==", "processedFileType": "REDACTED_FILE", "processedFileExtension": "PDF" },
    { "processedFile": "0X2xhYmVsIjoiQ1JFRElUX0NB==", "processedFileType": "ENTITIES", "processedFileExtension": "JSON" }
  ],
  "metrics": { "size": 12.4, "wordCount": 210, "characterCount": 1180, "pages": 2 },
  "message": ""
}
```

> [!TIP]
> See the full examples in the samples directory: [deidentify_file.go](../samples/flowvault/detectapi/deidentify_file.go) and [deidentify_file_advanced.go](../samples/flowvault/detectapi/deidentify_file_advanced.go) (`FileObject`, media options, per-call `CustomHeaders`).

## Get Run
`GetRun` returns the current state of a `DeidentifyFile` run. It's how you get the result of a run you didn't poll. `Output`, `Metrics` and `OutputType` are set once the run succeeds. Each output is base64 or a presigned URL, as `OutputType` says, and the SDK writes nothing to disk. `OutputType` is a `DeidentifySourceType`; it is nil for a value the SDK doesn't know. `Metrics` and each of its counts are pointers, nil when the server leaves them out. When the run failed, `Message` says why.

```go
run, err := service.GetRun(ctx, common.GetRunRequest{RunId: "<RUN_ID>"}, common.GetRunOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
switch run.Status {
case common.DetectRunStatusSuccess:
	for i, output := range run.Output {
		if run.OutputType != common.DeidentifySourceTypeBase64 {
			continue // a presigned URL: fetch it, and don't log it
		}
		decoded, decodeErr := base64.StdEncoding.DecodeString(output.ProcessedFile)
		if decodeErr != nil {
			fmt.Println(decodeErr)
			return
		}
		// The type and extension come from the server: keep only safe characters (namePart
		// in the sample allows a-z, 0-9 and _), so the name can't leave this directory.
		name := fmt.Sprintf("run-output-%d-%s.%s", i,
			namePart(string(output.ProcessedFileType), "output"), namePart(string(output.ProcessedFileExtension), "bin"))
		if writeErr := os.WriteFile(name, decoded, 0o600); writeErr != nil {
			fmt.Println(writeErr)
			return
		}
	}
case common.DetectRunStatusInProgress, common.DetectRunStatusQueued:
	fmt.Println("Still running: look it up again later")
default: // FAILED or UNKNOWN
	fmt.Println("Run failed:", run.Message)
}
```

Sample response (successful run):
```json
{
  "runId": "89699e2c-4301-4f9f-bcff-0a8fd3057489",
  "status": "SUCCESS",
  "outputType": "BASE64",
  "output": [
    { "processedFile": "TXkgU1NOIGlzIFtTU0==", "processedFileType": "REDACTED_FILE", "processedFileExtension": "PDF" }
  ],
  "metrics": { "size": 12.4, "wordCount": 210, "characterCount": 1180, "pages": 2 },
  "message": ""
}
```

A run still in progress has `"status": "IN_PROGRESS"` and no `output`. A failed run has `"status": "FAILED"` and the reason in `message`.

> [!TIP]
> See the full example in the samples directory: [get_run.go](../samples/flowvault/detectapi/get_run.go)

## Reidentify String
`ReidentifyString` replaces the vault tokens in `Text`. Only values de-identified with `DeidentificationTypeVaultToken` can be re-identified. `ENTITY_UNIQUE_COUNTER`, `ENTITY_ONLY` and `MASK` placeholders such as `[NAME_1]` are not stored, so they come back unchanged.

A vault token in the text looks like `[NAME_<token>]`: the entity-type prefix, then the token stored in the entity's `Destination` token group. A token de-identified with `IncludePrefixNo` has no prefix and can't be re-identified.

Without `RedactionLevel`, every token gets its original value. `RedactionLevel` is a `[]common.RedactionRule` of `common.RedactionLevel` values. Each one has two fields, both configured in the vault:

| Field | Meaning |
|---|---|
| `TokenGroupName` | The token group whose tokens the rule applies to: the `Destination` used when de-identifying. |
| `RedactionPattern` | The named redaction pattern that renders the token group's values, for example a partial mask. |

```go
// De-identify into vault tokens first: only these can be re-identified
deidentified, err := service.DeidentifyString(ctx, common.DeidentifyStringRequest{
	Text: "My name is John Doe.",
	Configuration: &common.DetectConfiguration{
		Detect: &common.Detect{
			Entities: []common.Entity{
				{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TOKEN_GROUP_NAME>"},
			},
		},
	},
}, common.DeidentifyStringOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
fmt.Println(deidentified.ProcessedText) // My name is [NAME_<token>].

// Re-identify, rendering the token group through a redaction pattern
response, err := service.ReidentifyString(ctx, common.ReidentifyStringRequest{
	Text: deidentified.ProcessedText,
	RedactionLevel: []common.RedactionRule{common.RedactionLevel{
		TokenGroupName:   "<TOKEN_GROUP_NAME>",
		RedactionPattern: "<REDACTION_PATTERN_NAME>",
	}},
}, common.ReidentifyStringOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
fmt.Println(response.ProcessedText) // the name, as <REDACTION_PATTERN_NAME> renders it
```

Sample response, with a redaction pattern that masks the middle of the name:
```json
{
  "processedText": "My name is Jo****oe.",
  "metrics": { "size": 0.02, "wordCount": 4, "characterCount": 20 }
}
```

> [!TIP]
> See the full example in the samples directory: [reidentify_string.go](../samples/flowvault/detectapi/reidentify_string.go)

## Reidentify File
`ReidentifyFile` is synchronous: there's no run to poll. `DataSource` and `Value` are required: either `ReidentifySourceTypeBase64` with base64 content, or `ReidentifySourceTypePresignedURL` with a URL. There's no `FilePath` or `FileObject`, so encode a local file into `Value` yourself. `DataFormat` is optional: a `ReidentifyFileDataFormat*` constant (`CSV`, `DOCX`, `JSON`, `JSONL`, `PPTX`, `TXT`, `XLSX`, `XML`), the formats re-identification works with. Each output is a `common.ReidentifyFileOutput`, whose `ProcessedFileExtension` is the same type: empty for any other format. `RedactionLevel` works as in [Reidentify String](#reidentify-string). The response's `OutputType` is a `ReidentifySourceType`, or nil for a value the SDK doesn't know.

The SDK writes nothing to disk: the re-identified output is in `response.Output` for you to handle.

```go
content, readErr := os.ReadFile("<PATH_TO_FILE>/processed-notes.txt")
if readErr != nil {
	fmt.Println(readErr)
	return
}
response, err := service.ReidentifyFile(ctx, common.ReidentifyFileRequest{
	DataSource: common.ReidentifySourceTypeBase64,
	Value:      base64.StdEncoding.EncodeToString(content),
	DataFormat: common.ReidentifyFileDataFormatTXT,
	RedactionLevel: []common.RedactionRule{common.RedactionLevel{
		TokenGroupName:   "<TOKEN_GROUP_NAME>",
		RedactionPattern: "<REDACTION_PATTERN_NAME>",
	}},
}, common.ReidentifyFileOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
for i, output := range response.Output {
	if response.OutputType != common.ReidentifySourceTypeBase64 {
		// A presigned URL: anyone holding it can download the data, so don't log it.
		fmt.Println("Output", i, "of type", output.ProcessedFileType, "is at a presigned URL")
		continue
	}
	// Decode output.ProcessedFile and save it under a sanitized name, as in the sample.
}
```

Sample response:
```json
{
  "status": "SUCCESS",
  "outputType": "BASE64",
  "output": [
    { "processedFile": "TXkgbmFtZSBpcyBKb2huIERvZS4=", "processedFileType": "REIDENTIFIED_FILE", "processedFileExtension": "TXT" }
  ],
  "metrics": { "size": 0.03, "wordCount": 5, "characterCount": 20 }
}
```

> [!TIP]
> See the full examples in the samples directory: [reidentify_file.go](../samples/flowvault/detectapi/reidentify_file.go), and [round_trip.go](../samples/flowvault/detectapi/round_trip.go) for a file de-identified into vault tokens and then re-identified.

## Check Guardrails
`CheckGuardrails` screens text for toxicity and denied topics. `CheckToxicity` is a `*bool`: toxicity is checked only when it is `true`. Nil leaves the field out of the request, and the server then doesn't check toxicity.

```go
response, err := service.CheckGuardrails(ctx, common.CheckGuardrailsRequest{
	Text:          "<USER_PROMPT>",
	CheckToxicity: common.Ptr(true),
	DenyTopics:    []string{"politics", "medical advice"},
}, common.CheckGuardrailsOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
if response.Validation == common.GuardrailsValidationFailed {
	// Toxic and DeniedTopic are *bool: nil when the API leaves them out.
	fmt.Println("Blocked. Toxic:", response.Toxic != nil && *response.Toxic, "Denied topic:", response.DeniedTopic != nil && *response.DeniedTopic)
}
```

Sample response:
```json
{ "text": "<USER_PROMPT>", "validation": "FAILED", "toxic": false, "deniedTopic": true }
```

> [!TIP]
> See the full example in the samples directory: [check_guardrails.go](../samples/flowvault/detectapi/check_guardrails.go)

## Detect configuration reference
An inline `DetectConfiguration` (the `Configuration` field) has these parts. Optional numbers and flags are pointers: nil leaves them to the server default, and `common.Ptr(v)` sets them. The SDK fills in the vault ID.

| `DetectConfiguration` field | Type | Meaning |
|---|---|---|
| `Name`, `Description` | `string` | Optional labels. |
| `Detect` | `*Detect` | What is detected and how it's de-identified. |
| `FileMapping` | `[]FileMapping` | Maps a source file location to where its outputs go: `Source`, `Destination`, `OutputFileNameSuffix`, `EntitiesDestination`, `ObjectEntitiesDestination`. |
| `Media` | `*Media` | Per-media options. File calls only. |

| `Detect` field | Type | Meaning |
|---|---|---|
| `Entities` | `[]Entity` | One rule per entity type (below). `EntityTypeAll` matches every type without its own rule. |
| `ObjectEntities` | `[]ObjectEntity` | Objects in images and video. `EntityType`: `ObjectEntityTypeFace`, `ObjectEntityTypeLicensePlate`, `ObjectEntityTypeLogo`, `ObjectEntityTypeSignature`, `ObjectEntityTypeAll`. `DeidentificationType`: `ObjectDeidentificationTypeRedact` or `ObjectDeidentificationTypeUnredact`. File calls only. |
| `Restrict` | `[]string` | Regexes always treated as sensitive. |
| `Skip` | `[]string` | Regexes never treated as sensitive. |
| `CustomTypes` | `[]CustomType` | Custom entities: a `Label` (reported as the entity type), `Match` regexes, and the same `DeidentificationType`, `Destination`, `IncludePrefix` and `PrefixOverride` fields as `Entity`. |
| `ReturnEntities` | `ReturnEntitiesType` | What `Entities` in the response carries: `ReturnEntitiesTypeAll` (with values), `ReturnEntitiesTypeExcludeSensitiveData` (without values) or `ReturnEntitiesTypeNone`. |

| `Entity` field | Meaning |
|---|---|
| `EntityType` | An `EntityType*` constant, for example `EntityTypeName`, `EntityTypeSSN` or `EntityTypeEmailAddress`. See `flowvault/utils/common/detect.go` for the full list. |
| `DeidentificationType` | `DeidentificationTypeVaultToken` (stored in the token group and reversible), `DeidentificationTypeEntityUniqueCounter` (`[NAME_1]`), `DeidentificationTypeEntityOnly` (`[NAME]`) or `DeidentificationTypeMask`. |
| `Destination` | The token group, for `DeidentificationTypeVaultToken`. |
| `Transformation` | `ShiftDates{MinDays, MaxDays}` shifts dates by a random number of days in that range, instead of replacing them. |
| `IncludePrefix` | `IncludePrefixYes` / `IncludePrefixNo`: whether a vault token carries the entity-type prefix, as in `[NAME_<token>]`. Applies only to `DeidentificationTypeVaultToken`; the server default is yes. **`IncludePrefixNo` returns the raw vault token, and the value can't be re-identified later.** |
| `PrefixOverride` | A custom prefix instead of the entity type, for example `NAAM` gives `[NAAM_<token>]`. Not valid with `IncludePrefixNo`. |

| `Media` field | Options |
|---|---|
| `Audio` | `OutputProcessedAudio *bool`; `OutputTranscription` (`OutputTranscriptionType*`: none, transcription, medical, diarized, medical diarized); `Bleep` with `StartPadding` / `StopPadding` (seconds, 0–20; server defaults 0.5 and 0.2), `Frequency` (Hz) and `Gain`. |
| `Document.Pdf` | `ProcessingMode` (`PdfProcessingModeOCR`, `PdfProcessingModeTextLayer`, `PdfProcessingModeNone`), `Density`, `MaxResolution`. |
| `Image` | `OutputProcessedImage *bool`, `OutputOcrText *bool`, `MaskingMethod` (`MaskingMethodBlur`, `MaskingMethodBlackbox`, `MaskingMethodNone`). |

The supported file formats are the `FileDataFormat*` constants: `pdf`, `txt`, `csv`, `json`, `jsonl`, `xml`, `doc`, `docx`, `xls`, `xlsx`, `ppt`, `pptx`, `jpg`, `jpeg`, `png`, `bmp`, `gif`, `tif`, `tiff`, `dcm`, `mp3`, `wav` and `zip`.

The SDK doesn't limit text or file size. Any limit is enforced by the server, and a request over it comes back as a `*SkyflowError`.

## Detect request headers
To add headers to a Detect call, set `CustomHeaders` in its options, as for the vault operations (see [Custom Request Headers](#custom-request-headers)).

```go
options := common.DeidentifyStringOptions{
	CustomHeaders: map[common.CustomHeaderKey]string{
		common.RequestIdHeader: "<REQUEST_ID>",
	},
}
```

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
| `HttpCode` | Per-item status (`int`). `2xx` for success; `4xx`/`5xx` for failure. When the vault leaves it out, it is `200`, or `500` for an item with an error. |
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
| `GetHttpCode()` | `string` | HTTP status code, for example `"400"`, `"404"`, `"500"`. `"499"` when the call's context was cancelled or its deadline passed. `"0"` (with an empty `GetHttpStatusCode()`) when no usable HTTP response came back: a network failure, the call ran past its `Timeout`, or the response body couldn't be read. Also `"0"` when there was no HTTP exchange at all, for example when Detect couldn't write an output file. |
| `GetHttpCodeInt()` | `int` | `GetHttpCode()` as a number; `0` when there is no HTTP code. |
| `GetMessage()` | `string` | Human-readable description of the error, as it is. For example `"Skyflow FlowVault Go SDK 1.0.0 Validation error. 'query' can't be empty. Specify a query"`. |
| `GetGrpcCode()` | `string` | gRPC status code from the server. |
| `GetGrpcCodeInt()` | `int` | `GetGrpcCode()` as a number; `0` when the server sent none. |
| `GetHttpStatusCode()` | `string` | HTTP status text from the server. |
| `GetRequestId()` | `string` | The `x-request-id` response header. Useful for support requests. |
| `GetDetails()` | `[]interface{}` | Additional error context from the server. Empty for validation errors. |
| `GetResponseBody()` | `map[string]interface{}` | The server's error response body as it was sent, decoded, when it is a JSON object; `nil` otherwise (for example a validation error or an HTML page from a proxy). |
| `Error()` | `string` | Implements `error`: the message, followed by `": "` and the cause when the SDK kept one. |
| `Unwrap()` | `error` | The underlying error when the SDK kept one (the context's error for a `"499"`, the network or timeout error for a `"0"`), so `errors.Is` / `errors.As` work. |

**Validation errors** (table name on both levels, empty token list and similar) are returned before any network call, with HTTP code `"400"`. **API errors** come from the Skyflow server, with the fields filled in from the response body and headers; `GetHttpCode()` is always the response's HTTP status. When the body isn't a Skyflow JSON error, the message is, in order: an `error` string in the body, a top-level `message`, the body itself (for example an HTML page from a proxy; only its first 512 bytes are kept), or `"Error with status code <N>"` for an empty body.

When a call fails (a non-`2xx` status) but its body still lists per-record results, `Insert`, `Update`, `Get`, `Delete`, `Detokenize`, `GetTokens` and `DeleteFiles` return those records instead of an error. For `Insert`, `Update`, `Get`, `Delete` and `Detokenize` every such record is failed: its `HttpCode` is its own status or else the call's, its `Error` is its own error or message or else `"Unknown error"`, every record carries the request ID, and no data, tokens or detokenized values are returned. For `GetTokens` and `DeleteFiles` a record counts as a success only when it reports no error and a `2xx` status of its own. `UploadFiles` instead fails the whole call, so no file is uploaded.

A successful (`2xx`) response with an empty body is treated as an empty result. One whose body isn't a JSON object fails the call with its real status code (for example `"200"`) and a message that includes the first 512 bytes of the body.

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

## Detect run status

A `*SkyflowError` covers Detect request failures as it does for vault operations. A de-identify file run, though, can also finish without an error and still not have succeeded: check `Status` on the `*common.DeidentifyFileResponse` (when you set `Poll`) or the `*common.GetRunResponse`.

| `Status` | Meaning |
|---|---|
| `common.DetectRunStatusSuccess` | The run finished; `Output` and `Metrics` are set. |
| `common.DetectRunStatusFailed` / `common.DetectRunStatusUnknown` | The run did not succeed. `Message` has the failure detail. |
| `common.DetectRunStatusInProgress` / `common.DetectRunStatusQueued` | The poll wait ran out before the run finished. Check again later with `GetRun(RunId)`. |

`Status` is empty when `DeidentifyFile` was called without `Poll`: only `RunId` is set.

```go
detect, err := skyflowClient.Detect("<VAULT_ID>")
if err != nil {
	return
}
response, err := detect.DeidentifyFile(ctx, common.DeidentifyFileRequest{
	FilePath:        "<PATH_TO_FILE>",
	ConfigurationId: "<CONFIGURATION_ID>",
}, common.DeidentifyFileOptions{Poll: &common.PollOptions{}}) // PollOptions{} waits up to 60s (at most 7 lookups)
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
switch response.Status {
case common.DetectRunStatusSuccess:
	fmt.Println("outputs:", len(response.Output))
case common.DetectRunStatusInProgress, common.DetectRunStatusQueued:
	fmt.Println("still running; poll later with GetRun:", response.RunId)
default: // FAILED or UNKNOWN
	fmt.Println("run failed:", response.Message)
}
```

## Common Detect errors

These come back from the server as a `*SkyflowError`. `GetMessage()` has the server's reason; quote `GetRequestId()` if you contact support.

| Cause | Status | Fix |
|---|---|---|
| Both `ConfigurationId` and `Configuration` are set. The SDK sends both, and the API rejects the pair. | `400` | Set one. |
| An entity or custom type uses `DeidentificationTypeVaultToken` without a `Destination`, or names a token group the vault doesn't have. Detect doesn't create token groups. | `400` | Set `Destination` to an existing token group. |
| A `RedactionLevel` rule names a token group or redaction pattern the vault doesn't have. | `400` | Check both names in the vault. |
| `DeidentificationTypeVaultToken` in a file whose format doesn't support it. Only `txt`, `csv`, `json`, `jsonl`, `xml`, `docx`, `xlsx` and `pptx` do. | `400` | Use another de-identification type for that file, or convert it. |
| `PrefixOverride` set together with `IncludePrefixNo`. | `400` | Drop one of them. |
| A missing role, or re-identifying vault tokens without permission to detokenize. Detect needs the Detect Invoker, Detect Manager or Vault Owner role; re-identification also needs a policy granting the Detokenization action on the destination, unless the caller is a Vault Owner. | `403` | Grant the role or policy. |

The SDK rejects some mistakes itself, before any request, with a `400` `*SkyflowError` and no request ID. Examples: no file source in `DeidentifyFile`, `OutputDirectory` without `Poll`, and a `Poll` limit out of range.

# Samples

Runnable examples live in [samples/flowvault](../samples/flowvault/): one file per vault operation and Detect method, plus custom-header, timeout/retry and service account examples. See [samples/flowvault/README.md](../samples/flowvault/README.md) to run them.
