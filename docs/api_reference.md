# API Reference

A reference for the public Skyflow Go SDK v2 surface: client-management methods, config structs, request and response objects, helper types, enums, service-account utilities, and error handling. For task-oriented usage and examples, see the [README](../README.md).

All types, fields, and enum values below are taken directly from the SDK source.

## Table of Contents

- [Client management methods](#client-management-methods)
- [Config classes](#config-classes)
- [Request objects](#request-objects)
- [Response objects](#response-objects)
- [Helper classes](#helper-classes)
- [Enums](#enums)
- [Service account utilities](#service-account-utilities)
- [Error handling](#error-handling)

---

## Client management methods

### Constructor

| Function | Returns | Description |
|----------|---------|-------------|
| `NewSkyflow(opts ...Option)` | `(*Skyflow, *SkyflowError)` | Build and return a configured Skyflow client. |

### Option functions

Pass `Option` values to `NewSkyflow` to configure the client.

| Function | Description |
|----------|-------------|
| `WithVaults(config ...VaultConfig) Option` | Add one or more vault configurations. |
| `WithConnections(config ...ConnectionConfig) Option` | Add one or more connection configurations. |
| `WithCredentials(credentials Credentials) Option` | Set client-level credentials applied when a vault or connection config does not specify its own. |
| `WithLogLevel(logLevel LogLevel) Option` | Set the initial log level. See [`LogLevel`](#loglevel). |
| `WithCustomHeaders(headers map[CustomHeaderKey]string) Option` | Set client-level custom headers sent with every request. See [`CustomHeaderKey`](#customheaderkey). |

```go
import (
    "github.com/skyflowapi/skyflow-go/v2/client"
    "github.com/skyflowapi/skyflow-go/v2/utils/common"
    "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

skyflowClient, err := client.NewSkyflow(
    client.WithVaults(common.VaultConfig{
        VaultId:   "<VAULT_ID>",
        ClusterId: "<CLUSTER_ID>",
        Env:       common.PROD,
        Credentials: common.Credentials{
            ApiKey: "<API_KEY>",
        },
    }),
    client.WithLogLevel(logger.ERROR),
)
```

### Controller accessors

| Method | Returns | Description |
|--------|---------|-------------|
| `Vault(vaultID ...string)` | `(*VaultService, *SkyflowError)` | Get the vault service. Uses the first configured vault if no ID is given. |
| `Connection(connectionId ...string)` | `(*ConnectionService, *SkyflowError)` | Get the connection service. Uses the first configured connection if no ID is given. |
| `Detect(vaultID ...string)` | `(Detect service, *SkyflowError)` | Get the Detect service for a vault. Uses the first configured vault if no ID is given. See [Detect API](#detect-api). |

### Instance management methods

All mutating methods return `*SkyflowError` (nil on success) unless the table notes otherwise.

| Method | Returns | Description |
|--------|---------|-------------|
| `AddVaultConfig(VaultConfig)` | `*SkyflowError` | Add a vault after initialization. |
| `GetVaultConfig(vaultId string)` | `(*VaultConfig, *SkyflowError)` | Retrieve a vault configuration by ID. |
| `UpdateVaultConfig(VaultConfig)` | `*SkyflowError` | Update an existing vault configuration (matched by `VaultId`). Only the `ClusterId`, `Env`, and `Credentials` you set change; fields you leave empty keep their current values. |
| `RemoveVaultConfig(vaultId string)` | `*SkyflowError` | Remove a vault configuration. |
| `AddConnectionConfig(ConnectionConfig)` | `*SkyflowError` | Add a connection after initialization. |
| `GetConnectionConfig(connId string)` | `(*ConnectionConfig, *SkyflowError)` | Retrieve a connection configuration by ID. |
| `UpdateConnectionConfig(ConnectionConfig)` | `*SkyflowError` | Replace an existing connection configuration. |
| `RemoveConnectionConfig(connectionId string)` | `*SkyflowError` | Remove a connection configuration. |
| `AddSkyflowCredentials(Credentials)` | `*SkyflowError` | Set client-level credentials. |
| `UpdateSkyflowCredentials(Credentials)` | `*SkyflowError` | Replace client-level credentials. |
| `GetSkyflowCredentials()` | `*Credentials` | Get the current client-level credentials. |
| `GetLoglevel()` | `*LogLevel` | Get the current log level. |
| `UpdateLogLevel(LogLevel)` | — | Change the log level at runtime. |

```go
// Manage configuration after the client is built
skyflowClient.AddVaultConfig(common.VaultConfig{
    VaultId:   "<ANOTHER_VAULT_ID>",
    ClusterId: "<CLUSTER_ID>",
    Credentials: common.Credentials{ApiKey: "<API_KEY>"},
})
skyflowClient.UpdateLogLevel(logger.DEBUG)
level := skyflowClient.GetLoglevel()
```

---

## Config classes

### `VaultConfig`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — passed to `WithVaults()`, `AddVaultConfig()`, and `UpdateVaultConfig()`.

| Field | Type | Description |
|-------|------|-------------|
| `VaultId` | `string` | _(required)_ Vault ID. |
| `ClusterId` | `string` | _(required)_ Cluster ID — the first segment of the vault URL. |
| `Env` | `Env` | Deployment environment. Default: `PROD`. See [`Env`](#env). |
| `Credentials` | `Credentials` | Vault-specific credentials. Overrides client-level credentials for this vault. |
| `BaseVaultUrl` | `string` | _(optional)_ Override the base vault URL. |

### `ConnectionConfig`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — passed to `WithConnections()`, `AddConnectionConfig()`, and `UpdateConnectionConfig()`.

| Field | Type | Description |
|-------|------|-------------|
| `ConnectionId` | `string` | _(required)_ Connection ID. |
| `ConnectionUrl` | `string` | _(required)_ Connection URL. |
| `Credentials` | `Credentials` | Connection-specific credentials. Overrides client-level credentials for this connection. |

### `Credentials`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — use exactly one authentication field; the others should be left empty.

| Field | Type | Description |
|-------|------|-------------|
| `ApiKey` | `string` | API key for direct authentication. |
| `Token` | `string` | Static bearer token. |
| `Path` | `string` | Path to a service account `credentials.json` file. |
| `CredentialsString` | `string` | Service account credentials as a JSON string. |
| `Roles` | `[]string` | _(optional)_ Role IDs to scope the generated bearer token. |
| `Context` | `interface{}` | _(optional)_ Context value embedded in the bearer token for context-aware authorization. |

---

## Request objects

### Vault API

#### `InsertRequest`

Passed to `vault.Insert()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Table` | `string` | _(required)_ | Target table name. |
| `Values` | `[]map[string]interface{}` | _(required)_ | List of records to insert. Each map is a `column → value` mapping. |

#### `InsertOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ReturnTokens` | `bool` | `false` | Return tokens for the inserted values. |
| `Upsert` | `string` | `""` | Column name to use as the upsert key (column must have a `unique` constraint). |
| `Homogeneous` | `bool` | `false` | Treat the batch as homogeneous (all records share the same columns). |
| `TokenMode` | `BYOT` | `DISABLE` | Bring-your-own-token mode. See [`BYOT`](#byot). |
| `ContinueOnError` | `bool` | `false` | Continue the batch despite partial per-record errors. |
| `Tokens` | `[]map[string]interface{}` | `nil` | BYOT token values aligned positionally with `Values` (used with `TokenMode`). |
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. See [`CustomHeaderKey`](#customheaderkey). |

---

#### `DetokenizeRequest`

Passed to `vault.Detokenize()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DetokenizeData` | `[]DetokenizeData` | _(required)_ | List of token-redaction pairs to detokenize. See [`DetokenizeData`](#detokenizedata). |

#### `DetokenizeOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ContinueOnError` | `bool` | `false` | Continue despite per-token errors. |
| `DownloadUrl` | `bool` | `false` | Return file download URLs for file-type tokens. |
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `GetRequest`

Passed to `vault.Get()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Table` | `string` | _(required)_ | Target table name. |
| `Ids` | `[]string` | `nil` | Skyflow IDs to retrieve. Mutually exclusive with `ColumnName`/`ColumnValues` in `GetOptions`. |

#### `GetOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `RedactionType` | `RedactionType` | `DEFAULT` | Redaction applied to returned values. See [`RedactionType`](#redactiontype). |
| `ReturnTokens` | `bool` | `false` | Return tokens instead of plain values. |
| `Fields` | `[]string` | `nil` | Specific columns to return. Returns all columns if empty. |
| `Offset` | `string` | `""` | Pagination offset. |
| `Limit` | `string` | `""` | Pagination limit. |
| `DownloadUrl` | `bool` | `false` | Return file download URLs for file columns. |
| `ColumnName` | `string` | `""` | Unique column to look up by value. Mutually exclusive with `Ids`. |
| `ColumnValues` | `[]string` | `nil` | Values for `ColumnName`. |
| `OrderBy` | `OrderByEnum` | `NONE` | Sort order. See [`OrderByEnum`](#orderbyenum). |
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `UpdateRequest`

Passed to `vault.Update()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Table` | `string` | _(required)_ | Target table name. |
| `Data` | `map[string]interface{}` | _(required)_ | Map containing `"SkyflowId"` (the record to update) plus the columns and their new values. |
| `Tokens` | `map[string]interface{}` | `nil` | BYOT token values for the updated columns. |

#### `UpdateOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ReturnTokens` | `bool` | `false` | Return tokens for the updated record. |
| `TokenMode` | `BYOT` | `DISABLE` | Bring-your-own-token mode. See [`BYOT`](#byot). |
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `DeleteRequest`

Passed to `vault.Delete()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Table` | `string` | _(required)_ | Target table name. |
| `Ids` | `[]string` | _(required)_ | Skyflow IDs of the records to delete. |

#### `DeleteOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `QueryRequest`

Passed to `vault.Query()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Query` | `string` | _(required)_ | SQL-like query string to execute against the vault. |

#### `QueryOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `TokenizeRequest`

Each element in the `[]TokenizeRequest` slice passed to `vault.Tokenize()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Value` | `string` | _(required)_ | The value to tokenize. |
| `ColumnGroup` | `string` | _(required)_ | The column group that defines the tokenization policy. |

#### `TokenizeOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `FileUploadRequest`

Passed to `vault.UploadFile()`. Provide exactly one file source: `FilePath`, `Base64`, or `FileObject`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Table` | `string` | _(required)_ | Target table name. |
| `SkyflowId` | `string` | `""` | Existing record ID to attach the file to. Omit to create a new record. |
| `ColumnName` | `string` | `""` | File column name. |
| `FilePath` | `string` | `""` | Path to a local file to upload. |
| `Base64` | `string` | `""` | Base64-encoded file content. |
| `FileName` | `string` | `""` | Override the file name sent to the vault. |
| `FileObject` | `os.File` | — | A file object to upload. |

#### `FileUploadOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

### Connection API

#### `InvokeConnectionRequest`

Passed to `connection.Invoke()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Method` | `RequestMethod` | `POST` | HTTP method. See [`RequestMethod`](#requestmethod). |
| `PathParams` | `map[string]string` | `nil` | Path parameter substitutions. |
| `QueryParams` | `map[string]interface{}` | `nil` | Query string parameters. |
| `Headers` | `map[string]string` | `nil` | Additional request headers. |
| `Body` | `interface{}` | `nil` | Request body. |

---

### Detect API

The Detect methods are on the service that `client.Detect()` returns. Each method takes a `context.Context`, a request struct, and an options struct, and returns a response and a `*SkyflowError`. All Detect types are in `github.com/skyflowapi/skyflow-go/v2/utils/common`.

These are the Detect V2 types. To move from the V1 methods and types they replace, see [Migrate from Detect V1](../README.md#migrate-from-detect-v1).

#### `DeidentifyStringRequest`

Passed to `DeidentifyString()`. Set exactly one of `ConfigurationId` and `Configuration`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Text` | `string` | _(required)_ | Text to de-identify. |
| `ConfigurationId` | `string` | `""` | ID of a stored Detect configuration. |
| `Configuration` | `*DetectConfiguration` | `nil` | Inline configuration to use instead of a stored one. See [`DetectConfiguration`](#detectconfiguration). |

#### `DeidentifyFileRequest`

Passed to `DeidentifyFile()`. Supply the file as exactly one of `Value` (with `DataSource`), `FilePath`, or `FileObject`. Set exactly one of `ConfigurationId` and `Configuration`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DataSource` | `DataSourceType` | `""` | What `Value` holds: `DataSourceTypeBase64`, `DataSourceTypeSkyflowID`, or `DataSourceTypePresignedURL`. Required with `Value`. With `FilePath` or `FileObject`, leave it empty or set `DataSourceTypeBase64`. |
| `Value` | `string` | `""` | Base64 file content, a Skyflow ID, or a presigned URL. |
| `FilePath` | `string` | `""` | Path to a local file. The SDK reads it and sends it as base64. |
| `FileObject` | `*os.File` | `nil` | An open local file. The SDK reads it and sends it as base64. The caller closes it. |
| `DataFormat` | `FileDataFormat` | `""` | File format. Optional: the server infers it, and for a local file the SDK infers it from the extension. Set it to override either. See [Detect enums](#detect-enums). |
| `ConfigurationId` | `string` | `""` | ID of a stored Detect configuration. |
| `Configuration` | `*DetectConfiguration` | `nil` | Inline configuration. See [`DetectConfiguration`](#detectconfiguration). |

#### `GetRunRequest`

Passed to `GetRun()` to look up a `DeidentifyFile` run.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `RunId` | `string` | _(required)_ | The run ID returned by `DeidentifyFile()`. |

#### `ReidentifyStringRequest`

Passed to `ReidentifyString()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Text` | `string` | _(required)_ | Text that contains tokens to re-identify. |
| `RedactionLevel` | `[]RedactionLevel` | `nil` | How to show each group of tokens. Without it, every token gets its original value. See [`RedactionLevel`](#redactionlevel). |

#### `ReidentifyFileRequest`

Passed to `ReidentifyFile()`. The call is synchronous, so there is no run to poll. Supply the file as exactly one of `Value` (with `DataSource`), `FilePath`, or `FileObject`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DataSource` | `DataSourceType` | `""` | What `Value` holds. Same rules as in [`DeidentifyFileRequest`](#deidentifyfilerequest). |
| `Value` | `string` | `""` | Base64 file content, a Skyflow ID, or a presigned URL. |
| `FilePath` | `string` | `""` | Path to a local file, sent as base64. |
| `FileObject` | `*os.File` | `nil` | An open local file, sent as base64. The caller closes it. |
| `DataFormat` | `FileDataFormat` | `""` | File format. Optional. |
| `RedactionLevel` | `[]RedactionLevel` | `nil` | How to show each group of tokens. See [`RedactionLevel`](#redactionlevel). |

#### `CheckGuardrailsRequest`

Passed to `CheckGuardrails()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Text` | `string` | _(required)_ | Text to screen. |
| `CheckToxicity` | `bool` | `false` | Check the text for toxicity. |
| `DenyTopics` | `[]string` | `nil` | Topics the text must not cover. Values can't be empty. |

#### Detect options

Each Detect method has its own options struct: `DeidentifyStringOptions`, `DeidentifyFileOptions`, `GetRunOptions`, `ReidentifyStringOptions`, `ReidentifyFileOptions`, and `CheckGuardrailsOptions`. Every one has this field:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Interceptor` | `Interceptor` | `nil` | A `func(*RequestContext)` that runs before the call is sent. Add a header to this call with `rc.AddHeader(key, value)`. See [`CustomHeaderKey`](#customheaderkey). Client-level headers set with `WithCustomHeaders` are also sent. |

`DeidentifyFileOptions` also has:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Poll` | `*PollOptions` | `nil` | Wait for the run to finish. When `nil`, the call returns after the file is submitted, with only `RunId` set. See [`PollOptions`](#polloptions). |
| `OutputDirectory` | `string` | `""` | Where to write a successful run's outputs for a `FilePath` or `FileObject` input. Empty means the current directory. Requires `Poll`. |

`ReidentifyFileOptions` also has:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `OutputDirectory` | `string` | `""` | Where to write the outputs for a `FilePath` or `FileObject` input. Empty means the current directory. |

The SDK writes outputs only when they're base64. When the outputs are presigned URLs, it writes nothing.

#### `PollOptions`

Used in `DeidentifyFileOptions.Poll`. A zero value uses the default.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `WaitTime` | `time.Duration` | 60 seconds | The longest `DeidentifyFile()` waits. Between 1 and 300 seconds. |
| `MaxAttempts` | `int` | `23` | The most run lookups. Between 1 and 23. |

The first lookup is immediate. After that, the wait between lookups starts at 1 second and doubles up to 32 seconds. Polling stops at a final status (`SUCCESS`, `FAILED`, or `UNKNOWN`), or when `WaitTime` or `MaxAttempts` runs out. In that case, the response has the run's current status and its `RunId`, so you can call `GetRun()` later.

---

## Response objects

> **The `Errors` field** is present on most responses. It is populated only on partial failure (for example when `ContinueOnError: true` is set); it is nil or empty when there are no errors.

### `InsertResponse`

Returned by `vault.Insert()`.

| Field | Type | Description |
|-------|------|-------------|
| `InsertedFields` | `[]map[string]interface{}` | One entry per inserted record. Each map contains `skyflow_id`; with `ReturnTokens: true`, also a token per column; with `ContinueOnError: true`, also a `request_index`. |
| `Errors` | `[]map[string]interface{}` | Per-record errors when `ContinueOnError: true`. Each map contains `request_index`, `error`, and `http_code`. |

### `DetokenizeResponse`

Returned by `vault.Detokenize()`.

| Field | Type | Description |
|-------|------|-------------|
| `DetokenizedFields` | `[]DetokenizeRecordResponse` | One entry per successfully detokenized token. See [`DetokenizeRecordResponse`](#detokenizerecordresponse). |
| `Errors` | `[]DetokenizeRecordResponse` | Per-token errors when `ContinueOnError: true`. |

### `GetResponse`

Returned by `vault.Get()`.

| Field | Type | Description |
|-------|------|-------------|
| `Data` | `[]map[string]interface{}` | Retrieved records as `column → value` maps. Returns tokens when `ReturnTokens: true`. |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

### `UpdateResponse`

Returned by `vault.Update()`.

| Field | Type | Description |
|-------|------|-------------|
| `UpdatedField` | `map[string]interface{}` | The updated record's `skyflow_id` and, when `ReturnTokens: true`, a token per updated column. |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

### `DeleteResponse`

Returned by `vault.Delete()`.

| Field | Type | Description |
|-------|------|-------------|
| `DeletedIds` | `[]string` | Skyflow IDs of the deleted records. |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

### `QueryResponse`

Returned by `vault.Query()`.

| Field | Type | Description |
|-------|------|-------------|
| `Fields` | `[]map[string]interface{}` | Matching records. Each map includes a `tokenized_data` entry. |
| `Errors` | `[]map[string]interface{}` | Always nil (errors from query throw `SkyflowError` directly). |

### `TokenizeResponse`

Returned by `vault.Tokenize()`.

| Field | Type | Description |
|-------|------|-------------|
| `Tokens` | `[]string` | One token per input `TokenizeRequest`, in order. |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

### `FileUploadResponse`

Returned by `vault.UploadFile()`.

| Field | Type | Description |
|-------|------|-------------|
| `SkyflowId` | `string` | Skyflow ID of the record the file was attached to (or the newly created record). |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

### `InvokeConnectionResponse`

Returned by `connection.Invoke()`.

| Field | Type | Description |
|-------|------|-------------|
| `Data` | `interface{}` | The response body from the downstream service. |
| `Metadata` | `map[string]interface{}` | Response metadata (e.g. forwarded HTTP headers). |
| `Errors` | `map[string]interface{}` | Errors, if any. |

### `DeidentifyStringResponse`

Returned by `DeidentifyString()`.

| Field | Type | Description |
|-------|------|-------------|
| `ProcessedText` | `string` | The text with sensitive values replaced. |
| `Entities` | `[]DetectedEntity` | The entities found. See [`DetectedEntity`](#detectedentity). |
| `Metrics` | `*Metrics` | Size and counts of the input. See [`Metrics`](#metrics). |

### `DeidentifyFileResponse`

Returned by `DeidentifyFile()`. `RunId` is always set. The other fields are set only when the call polls (`Poll` is set). `Output`, `Metrics`, and `OutputType` are set once the run succeeds.

| Field | Type | Description |
|-------|------|-------------|
| `RunId` | `string` | The run ID. Pass it to `GetRun()`. |
| `Status` | `DetectRunStatus` | The run status. Empty when the call didn't poll. See [Detect enums](#detect-enums). |
| `OutputType` | `DataSourceType` | Whether each output is base64 (`BASE64`) or a presigned URL (`PRESIGNED_URL`). |
| `Output` | `[]FileOutput` | The run outputs. See [`FileOutput`](#fileoutput). |
| `Metrics` | `*Metrics` | Size and counts of the input. |
| `Message` | `string` | Failure detail when `Status` is `FAILED`. |

### `GetRunResponse`

Returned by `GetRun()`. `Output`, `Metrics`, and `OutputType` are set once the run succeeds.

| Field | Type | Description |
|-------|------|-------------|
| `RunId` | `string` | The run ID from the request. |
| `Status` | `DetectRunStatus` | The run status. |
| `OutputType` | `DataSourceType` | Whether each output is base64 or a presigned URL. |
| `Output` | `[]FileOutput` | The run outputs. |
| `Metrics` | `*Metrics` | Size and counts of the input. |
| `Message` | `string` | Failure detail when `Status` is `FAILED`. |

### `ReidentifyStringResponse`

Returned by `ReidentifyString()`.

| Field | Type | Description |
|-------|------|-------------|
| `ProcessedText` | `string` | The text with tokens replaced. |
| `Metrics` | `*Metrics` | Size and counts of the input. |

### `ReidentifyFileResponse`

Returned by `ReidentifyFile()`.

| Field | Type | Description |
|-------|------|-------------|
| `Status` | `DetectRunStatus` | The status of the re-identification. |
| `OutputType` | `DataSourceType` | Whether each output is base64 or a presigned URL. |
| `Output` | `[]FileOutput` | The outputs. |
| `Metrics` | `*Metrics` | Size and counts of the input. |

### `CheckGuardrailsResponse`

Returned by `CheckGuardrails()`.

| Field | Type | Description |
|-------|------|-------------|
| `Text` | `string` | The screened text. |
| `Validation` | `GuardrailsValidation` | `GuardrailsValidationPassed` or `GuardrailsValidationFailed`. |
| `Toxic` | `bool` | Whether the text is toxic. |
| `DeniedTopic` | `bool` | Whether the text covers a denied topic. |

---

## Helper classes

### `DetokenizeData`

Used inside [`DetokenizeRequest`](#detokenizerequest) to pair a token with a redaction type.

| Field | Type | Description |
|-------|------|-------------|
| `Token` | `string` | The token to detokenize. |
| `RedactionType` | `RedactionType` | Redaction applied to the returned value. See [`RedactionType`](#redactiontype). |

### `DetokenizeRecordResponse`

Each element returned in `DetokenizeResponse.DetokenizedFields` and `DetokenizeResponse.Errors`.

| Field | Type | Description |
|-------|------|-------------|
| `Token` | `string` | The input token. |
| `Value` | `string` | The detokenized value. Empty on error. |
| `Type` | `string` | The value type (e.g. `"STRING"`). Empty on error. |
| `Error` | `string` | Error message. Empty on success. |
| `RequestId` | `string` | Server request ID for this token — useful for support escalations. |

### `DetectConfiguration`

An inline Detect configuration for `DeidentifyString()` and `DeidentifyFile()`, used instead of a stored `ConfigurationId`. Create stored configurations in Studio or through the REST API; the SDK doesn't manage them. The SDK fills in the vault ID.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Configuration name. |
| `Description` | `string` | Configuration description. |
| `Detect` | `*Detect` | What to detect and how to de-identify it. See [`Detect`](#detect). |
| `FileMapping` | `[]FileMapping` | Where a file's outputs go. See [`FileMapping`](#filemapping). |
| `Media` | `*Media` | Options per media type. File calls only. See [`Media`](#media). |

### `Detect`

Used in [`DetectConfiguration`](#detectconfiguration).

| Field | Type | Description |
|-------|------|-------------|
| `Entities` | `[]Entity` | How to de-identify each entity type. See [`Entity`](#entity). |
| `ObjectEntities` | `[]ObjectEntity` | How to treat objects in images and video. File calls only. See [`ObjectEntity`](#objectentity). |
| `Restrict` | `[]string` | Regex patterns whose matches are always treated as sensitive. |
| `Skip` | `[]string` | Regex patterns whose matches are never treated as sensitive. |
| `CustomTypes` | `[]CustomType` | Custom entity types matched by regex. See [`CustomType`](#customtype). |
| `ReturnEntities` | `ReturnEntitiesType` | Which detected values the response carries. The server default is `EXCLUDE_SENSITIVE_DATA`. |

### `Entity`

Used in [`Detect`](#detect).

| Field | Type | Description |
|-------|------|-------------|
| `EntityType` | `EntityType` | The entity type. `EntityTypeAll` selects every type. See [`EntityType`](#entitytype). |
| `DeidentificationType` | `DeidentificationType` | How the value is replaced. The server default is `ENTITY_UNIQUE_COUNTER`. |
| `Destination` | `string` | For `DeidentificationTypeVaultToken`, the `<TABLE_NAME>.<COLUMN_NAME>` that stores the value. |
| `Transformation` | `*Transformation` | Change the value instead of replacing it. See [`Transformation`](#transformation). |
| `IncludePrefix` | `IncludePrefix` | Whether the replacement carries the entity-type prefix. The server default is `YES`. |
| `PrefixOverride` | `string` | Prefix to use instead of the entity type. Not valid with `IncludePrefixNo`. |

### `Transformation`

| Field | Type | Description |
|-------|------|-------------|
| `ShiftDates` | `*ShiftDates` | Shift dates by a random number of days. |

### `ShiftDates`

Shifts each date by a random number of days between `MinDays` and `MaxDays`.

| Field | Type | Description |
|-------|------|-------------|
| `MinDays` | `int` | Fewest days to shift. |
| `MaxDays` | `int` | Most days to shift. |

### `CustomType`

Detects values that match regex patterns as a custom entity type. `Label` becomes the `EntityType` of the detected entities.

| Field | Type | Description |
|-------|------|-------------|
| `Label` | `string` | Name of the custom entity type. |
| `Match` | `[]string` | Regex patterns to match. |
| `DeidentificationType` | `DeidentificationType` | How the value is replaced. |
| `Destination` | `string` | As in [`Entity`](#entity). |
| `IncludePrefix` | `IncludePrefix` | As in [`Entity`](#entity). |
| `PrefixOverride` | `string` | As in [`Entity`](#entity). |

### `ObjectEntity`

| Field | Type | Description |
|-------|------|-------------|
| `EntityType` | `ObjectEntityType` | The object type, such as a face or a license plate. |
| `DeidentificationType` | `ObjectDeidentificationType` | Whether to redact or unredact the object. |

### `FileMapping`

Maps a source file location to where its outputs go.

| Field | Type |
|-------|------|
| `Source` | `string` |
| `Destination` | `string` |
| `OutputFileNameSuffix` | `string` |
| `EntitiesDestination` | `string` |
| `ObjectEntitiesDestination` | `string` |

### `Media`

Options per media type for file calls.

| Field | Type | Description |
|-------|------|-------------|
| `Audio` | `*Audio` | Audio options. |
| `Document` | `*Document` | Document options. `Document.Pdf` is a `*Pdf`. |
| `Image` | `*Image` | Image options. |

### `Audio`

| Field | Type | Description |
|-------|------|-------------|
| `OutputProcessedAudio` | `bool` | Return the processed audio. |
| `OutputTranscription` | `OutputTranscriptionType` | The transcription to return. |
| `Bleep` | `*Bleep` | The tone that replaces sensitive audio. |

### `Bleep`

| Field | Type | Description |
|-------|------|-------------|
| `StartPadding` | `float32` | Padding before the bleep. |
| `StopPadding` | `float32` | Padding after the bleep. |
| `Frequency` | `int` | Frequency of the tone. |
| `Gain` | `int` | Gain of the tone. |

### `Pdf`

| Field | Type | Description |
|-------|------|-------------|
| `ProcessingMode` | `PdfProcessingMode` | How PDF text is read: `PdfProcessingModeOCR` or `PdfProcessingModeTextLayer`. |
| `Density` | `int` | Pixel density for processing. |
| `MaxResolution` | `int` | Maximum resolution for processing. |

### `Image`

| Field | Type | Description |
|-------|------|-------------|
| `OutputProcessedImage` | `bool` | Return the processed image. |
| `OutputOcrText` | `bool` | Return the text read from the image. |
| `MaskingMethod` | `MaskingMethod` | How detected regions are hidden: `MaskingMethodBlur` or `MaskingMethodBlackbox`. |

### `RedactionLevel`

Used in `ReidentifyStringRequest` and `ReidentifyFileRequest` to choose how a group of tokens is shown. In a v2 vault, set `EntityName` and `RedactionType`.

| Field | Type | Description |
|-------|------|-------------|
| `EntityName` | `EntityType` | The entity type of the tokens. |
| `RedactionType` | `RedactionType` | How to show them: `PLAIN_TEXT`, `MASKED`, `REDACTED`, or `DEFAULT`. See [`RedactionType`](#redactiontype). |
| `TokenGroupName` | `string` | For Flow Vault vaults only. |
| `RedactionPattern` | `string` | For Flow Vault vaults only. |

Each level needs exactly one of `EntityName` and `TokenGroupName`, and exactly one of `RedactionType` and `RedactionPattern`.

### `DetectedEntity`

Each element in `DeidentifyStringResponse.Entities`.

| Field | Type | Description |
|-------|------|-------------|
| `Token` | `string` | The token that replaced the value, such as `NAME_1`. |
| `Value` | `string` | The original value. Empty unless the configuration's `ReturnEntities` is `ReturnEntitiesTypeAll`. |
| `Location` | `EntityLocation` | Where the entity is. See [`EntityLocation`](#entitylocation). |
| `EntityType` | `string` | The entity type, or a custom type's `Label`. |
| `EntityScores` | `map[string]float64` | Confidence score per entity type, from 0.0 to 1.0. |

### `EntityLocation`

| Field | Type | Description |
|-------|------|-------------|
| `StartIndex` | `int` | Start of the entity in the original text. |
| `EndIndex` | `int` | End of the entity in the original text. |
| `StartIndexProcessed` | `int` | Start of the token in the processed text. |
| `EndIndexProcessed` | `int` | End of the token in the processed text. |

### `FileOutput`

Each element in the `Output` of a file response.

| Field | Type | Description |
|-------|------|-------------|
| `ProcessedFile` | `string` | The output, as base64 or a presigned URL, per the response's `OutputType`. |
| `ProcessedFileType` | `ProcessedFileType` | The kind of output, such as `REDACTED_FILE` or `ENTITIES`. |
| `ProcessedFileExtension` | `FileDataFormat` | The output's file extension. |

### `Metrics`

Describes the processed input. At most one of `Slides`, `Pages`, and `Duration` is set.

| Field | Type | Description |
|-------|------|-------------|
| `Size` | `float64` | Size in kilobytes. |
| `WordCount` | `int` | Word count. |
| `CharacterCount` | `int` | Character count. |
| `Slides` | `*int` | Slide count. |
| `Pages` | `*int` | Page count. |
| `Duration` | `*float64` | Duration in seconds. |

---

## Enums

### `LogLevel`

`github.com/skyflowapi/skyflow-go/v2/utils/logger`

| Value | Description |
|-------|-------------|
| `ERROR` | Log errors only. _(default)_ |
| `INFO` | Log informational messages, warnings, and errors. |
| `DEBUG` | Log everything (DEBUG, INFO, WARN, ERROR). |
| `WARN` | Log warnings and errors. |
| `OFF` | Disable all logging. |

### `Env`

`github.com/skyflowapi/skyflow-go/v2/utils/common`

| Value | Description |
|-------|-------------|
| `PROD` | Production environment. _(default)_ |
| `STAGE` | Staging environment. |
| `SANDBOX` | Sandbox environment. |
| `DEV` | Development environment. |

An unset `Env` (the zero value) resolves to `PROD`. Use the named constants rather than numeric values: `PROD` is `1`, not `0`.

### `RedactionType`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — used in [`DetokenizeData`](#detokenizedata), [`GetOptions`](#getoptions), and the Detect [`RedactionLevel`](#redactionlevel). The Detect constants `RedactionTypePlainText`, `RedactionTypeMasked`, `RedactionTypeRedacted`, and `RedactionTypeDefault` hold the same values.

| Value | Description |
|-------|-------------|
| `PLAIN_TEXT` | Return the original, unmasked value. |
| `MASKED` | Return a partially masked value (e.g. `****1234`). |
| `REDACTED` | Return a fully redacted placeholder. |
| `DEFAULT` | Use the redaction type configured on the vault column. |

### `BYOT`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — controls bring-your-own-token behavior for `Insert` and `Update`.

| Value | Description |
|-------|-------------|
| `DISABLE` | Do not use BYOT tokens. _(default)_ |
| `ENABLE` | Use provided tokens where supplied; generate tokens for fields that do not supply one. |
| `ENABLE_STRICT` | All fields must supply a BYOT token; missing tokens cause an error. |

### `RequestMethod`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — used in [`InvokeConnectionRequest`](#invokeconnectionrequest).

| Value | Description |
|-------|-------------|
| `GET` | HTTP GET. |
| `POST` | HTTP POST. |
| `PUT` | HTTP PUT. |
| `PATCH` | HTTP PATCH. |
| `DELETE` | HTTP DELETE. |

### `OrderByEnum`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — used in [`GetOptions`](#getoptions).

| Value | Description |
|-------|-------------|
| `ASCENDING` | Return records in ascending order. |
| `DESCENDING` | Return records in descending order. |
| `NONE` | No specific ordering. _(default)_ |

### `CustomHeaderKey`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — keys for custom header maps and for `RequestContext.AddHeader` in a Detect `Interceptor`.

| Value | Wire header | Description |
|-------|-------------|-------------|
| `SkyflowAccountId` | `x-skyflow-account-id` | Skyflow account ID header. |
| `SkyflowAccountName` | `x-skyflow-account-name` | Skyflow account name header. |
| `RequestIdHeader` | `x-request-id` | Custom request ID for tracing. |

### Detect enums

`github.com/skyflowapi/skyflow-go/v2/utils/common` — used in the [Detect API](#detect-api) types. Each type is a `string`, and each constant holds the API value shown.

| Type | Constants (value) |
|------|-------------------|
| `DeidentificationType` | `DeidentificationTypeEntityUniqueCounter` (`ENTITY_UNIQUE_COUNTER`, the server default), `DeidentificationTypeEntityOnly` (`ENTITY_ONLY`), `DeidentificationTypeVaultToken` (`VAULT_TOKEN`), `DeidentificationTypeUnknown` (`UNKNOWN`) |
| `IncludePrefix` | `IncludePrefixYes` (`YES`, the server default), `IncludePrefixNo` (`NO`), `IncludePrefixIncludePrefixUnspecified` (`INCLUDE_PREFIX_UNSPECIFIED`) |
| `ReturnEntitiesType` | `ReturnEntitiesTypeExcludeSensitiveData` (`EXCLUDE_SENSITIVE_DATA`, the server default), `ReturnEntitiesTypeAll` (`ALL`), `ReturnEntitiesTypeNone` (`NONE`) |
| `DataSourceType` | `DataSourceTypeBase64` (`BASE64`), `DataSourceTypeSkyflowID` (`SKYFLOW_ID`), `DataSourceTypePresignedURL` (`PRESIGNED_URL`), `DataSourceTypeFile` (`FILE`) |
| `FileDataFormat` | `FileDataFormatMP3`, `FileDataFormatWAV`, `FileDataFormatPDF`, `FileDataFormatTXT`, `FileDataFormatCSV`, `FileDataFormatJSON`, `FileDataFormatJSONL`, `FileDataFormatXML`, `FileDataFormatJPG`, `FileDataFormatJPEG`, `FileDataFormatTIF`, `FileDataFormatTIFF`, `FileDataFormatPNG`, `FileDataFormatBMP`, `FileDataFormatGIF`, `FileDataFormatXLS`, `FileDataFormatXLSX`, `FileDataFormatDOC`, `FileDataFormatDOCX`, `FileDataFormatPPT`, `FileDataFormatPPTX`, `FileDataFormatDCM`, `FileDataFormatZIP`, `FileDataFormatUnknown`. Each value is the lowercase extension, such as `pdf`. |
| `DetectRunStatus` | `DetectRunStatusQueued` (`QUEUED`), `DetectRunStatusInProgress` (`IN_PROGRESS`), `DetectRunStatusSuccess` (`SUCCESS`), `DetectRunStatusFailed` (`FAILED`), `DetectRunStatusUnknown` (`UNKNOWN`). `SUCCESS`, `FAILED`, and `UNKNOWN` are final. |
| `ProcessedFileType` | `ProcessedFileTypeRedactedFile` (`REDACTED_FILE`), `ProcessedFileTypeRedactedText` (`REDACTED_TEXT`), `ProcessedFileTypeEntities` (`ENTITIES`), `ProcessedFileTypeReidentifiedFile` (`REIDENTIFIED_FILE`), `ProcessedFileTypeRedactedTranscription` (`REDACTED_TRANSCRIPTION`), `ProcessedFileTypeRedactedDeepgramTranscriptionDiarizeJSON` (`REDACTED_DEEPGRAM_TRANSCRIPTION_DIARIZE_JSON`), `ProcessedFileTypeObjectEntities` (`OBJECT_ENTITIES`), `ProcessedFileTypeNone` (`NONE`) |
| `ObjectEntityType` | `ObjectEntityTypeFace` (`FACE`), `ObjectEntityTypeLicensePlate` (`LICENSE_PLATE`), `ObjectEntityTypeLogo` (`LOGO`), `ObjectEntityTypeSignature` (`SIGNATURE`), `ObjectEntityTypeAll` (`ALL`), `ObjectEntityTypeNone` (`NONE`) |
| `ObjectDeidentificationType` | `ObjectDeidentificationTypeRedact` (`REDACT`), `ObjectDeidentificationTypeUnredact` (`UNREDACT`), `ObjectDeidentificationTypeUnknown` (`UNKNOWN`) |
| `OutputTranscriptionType` | `OutputTranscriptionTypeTranscription` (`TRANSCRIPTION`), `OutputTranscriptionTypeMedicalTranscription` (`MEDICAL_TRANSCRIPTION`), `OutputTranscriptionTypeDiarizedTranscription` (`DIARIZED_TRANSCRIPTION`), `OutputTranscriptionTypeMedicalDiarizedTranscription` (`MEDICAL_DIARIZED_TRANSCRIPTION`), `OutputTranscriptionTypeNone` (`NONE`) |
| `PdfProcessingMode` | `PdfProcessingModeOCR` (`OCR`), `PdfProcessingModeTextLayer` (`TEXT_LAYER`), `PdfProcessingModeNone` (`NONE`) |
| `MaskingMethod` | `MaskingMethodBlur` (`BLUR`), `MaskingMethodBlackbox` (`BLACKBOX`), `MaskingMethodNone` (`NONE`) |
| `GuardrailsValidation` | `GuardrailsValidationPassed` (`PASSED`), `GuardrailsValidationFailed` (`FAILED`) |

### `EntityType`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — the entity types Detect can find, used in [`Entity`](#entity) and [`RedactionLevel`](#redactionlevel). Each value is the uppercase name, such as `EMAIL_ADDRESS` for `EntityTypeEmailAddress`. Use `EntityTypeAll` to select every type.

| Category | Constants |
|----------|-----------|
| Personal identity | `EntityTypeName`, `EntityTypeNameGiven`, `EntityTypeNameFamily`, `EntityTypeNameMedicalProfessional`, `EntityTypeDOB`, `EntityTypeAge`, `EntityTypeGender`, `EntityTypeMaritalStatus`, `EntityTypeSexuality`, `EntityTypeReligion`, `EntityTypePoliticalAffiliation`, `EntityTypePhysicalAttribute`, `EntityTypeOrigin` |
| Contact | `EntityTypeEmailAddress`, `EntityTypePhoneNumber`, `EntityTypeLocation`, `EntityTypeLocationAddress`, `EntityTypeLocationAddressStreet`, `EntityTypeLocationCity`, `EntityTypeLocationState`, `EntityTypeLocationCountry`, `EntityTypeLocationZIP`, `EntityTypeLocationCoordinate` |
| Financial | `EntityTypeCreditCard`, `EntityTypeCreditCardExpiration`, `EntityTypeCVV`, `EntityTypeBankAccount`, `EntityTypeAccountNumber`, `EntityTypeRoutingNumber`, `EntityTypeMoney`, `EntityTypeFinancialMetric`, `EntityTypeCorporateAction` |
| Government ID | `EntityTypeSSN`, `EntityTypeDriverLicense`, `EntityTypePassportNumber`, `EntityTypeHealthcareNumber`, `EntityTypeOrganizationID`, `EntityTypeNumericalPII` |
| Medical | `EntityTypeBloodType`, `EntityTypeCondition`, `EntityTypeDrug`, `EntityTypeDose`, `EntityTypeEffect`, `EntityTypeInjury`, `EntityTypeMedicalCode`, `EntityTypeMedicalProcess`, `EntityTypeOrganizationMedicalFacility` |
| Date and time | `EntityTypeDate`, `EntityTypeDay`, `EntityTypeMonth`, `EntityTypeYear`, `EntityTypeTime`, `EntityTypeDuration`, `EntityTypeDateInterval`, `EntityTypeEvent` |
| Technology | `EntityTypeIPAddress`, `EntityTypeURL`, `EntityTypeUsername`, `EntityTypePassword`, `EntityTypeFilename`, `EntityTypeVehicleID` |
| Organization | `EntityTypeOrganization`, `EntityTypeOccupation`, `EntityTypeProject`, `EntityTypeProduct` |
| Other | `EntityTypeLanguage`, `EntityTypeStatistics`, `EntityTypeTrend`, `EntityTypeZodiacSign`, `EntityTypeAll`, `EntityTypeNone` |

---

## Service account utilities

`github.com/skyflowapi/skyflow-go/v2/serviceaccount`

These functions generate bearer tokens and signed data tokens from Skyflow service account credentials. Bearer tokens are valid for 60 minutes.

### Functions

| Function | Parameters | Returns | Description |
|----------|------------|---------|-------------|
| `GenerateBearerToken(credentialsFilePath string, options BearerTokenOptions)` | path to `credentials.json` | `(*TokenResponse, *SkyflowError)` | Generate a bearer token from a credentials file. |
| `GenerateBearerTokenFromCreds(credentials string, options BearerTokenOptions)` | credentials JSON string | `(*TokenResponse, *SkyflowError)` | Generate a bearer token from a credentials JSON string. |
| `GenerateSignedDataTokens(credentialsFilePath string, options SignedDataTokensOptions)` | path to `credentials.json` | `([]SignedDataTokensResponse, *SkyflowError)` | Sign data tokens using a credentials file. |
| `GenerateSignedDataTokensFromCreds(credentials string, options SignedDataTokensOptions)` | credentials JSON string | `([]SignedDataTokensResponse, *SkyflowError)` | Sign data tokens using a credentials JSON string. |
| `IsExpired(tokenString string)` | — | `bool` | Returns `true` if the token is nil or has expired. Check before every API call. |

### `BearerTokenOptions`

| Field | Type | Description |
|-------|------|-------------|
| `Ctx` | `interface{}` | Context value embedded in the token for context-aware authorization. |
| `RoleIds` | `[]string` | Role IDs to scope the generated token. |
| `LogLevel` | `LogLevel` | Log level for this call. |

### `SignedDataTokensOptions`

| Field | Type | Description |
|-------|------|-------------|
| `DataTokens` | `[]string` | Data tokens to sign. |
| `TimeToLive` | `int` | Token validity in seconds. Default: `60`. |
| `Ctx` | `interface{}` | Context value embedded in the signed token. |
| `LogLevel` | `LogLevel` | Log level for this call. |

### `TokenResponse`

Returned by `GenerateBearerToken` and `GenerateBearerTokenFromCreds`.

| Field | Type | Description |
|-------|------|-------------|
| `AccessToken` | `string` | The bearer token string. Pass as `Authorization: Bearer <token>` or in `Credentials.Token`. |
| `TokenType` | `string` | Token type (always `"Bearer"`). |

### `SignedDataTokensResponse`

Each element returned by `GenerateSignedDataTokens` and `GenerateSignedDataTokensFromCreds`.

| Field | Type | Description |
|-------|------|-------------|
| `Token` | `string` | The original data token. |
| `SignedToken` | `string` | The signed token string (JWT). Use this for detokenization requests. |

```go
import "github.com/skyflowapi/skyflow-go/v2/serviceaccount"
import "github.com/skyflowapi/skyflow-go/v2/utils/common"
import "github.com/skyflowapi/skyflow-go/v2/utils/logger"

var token string

func getToken() (string, error) {
    if serviceaccount.IsExpired(token) {
        resp, err := serviceaccount.GenerateBearerToken("<PATH_TO_CREDENTIALS_JSON>", common.BearerTokenOptions{
            LogLevel: logger.ERROR,
        })
        if err != nil {
            return "", err
        }
        token = resp.AccessToken
    }
    return token, nil
}
```

---

## Error handling

### `SkyflowError`

All SDK operations return `*SkyflowError` instead of Go's built-in `error`. It is `nil` on success.

`github.com/skyflowapi/skyflow-go/v2/utils/error`

| Method | Return type | Description |
|--------|-------------|-------------|
| `Error() string` | `string` | Implements the `error` interface; returns the error message. |
| `GetMessage() string` | `string` | Human-readable error message. |
| `GetHttpStatusCode() string` | `string` | HTTP status code. Preferred over `GetCode()`. |
| `GetHttpCode() string` | `string` | HTTP status code string. |
| `GetCode() string` | `string` | _(deprecated)_ Use `GetHttpStatusCode()`. |
| `GetRequestId() string` | `string` | Server request ID — include in support escalations. |
| `GetGrpcCode() string` | `string` | gRPC status code (where applicable). |
| `GetDetails() []interface{}` | `[]interface{}` | Structured error detail objects. |
| `GetResponseBody() map[string]interface{}` | `map[string]interface{}` | Raw response body as a map. |

```go
resp, skyflowErr := vault.Insert(ctx, insertRequest, insertOptions)
if skyflowErr != nil {
    fmt.Println("Error message:   ", skyflowErr.GetMessage())
    fmt.Println("HTTP status:     ", skyflowErr.GetHttpStatusCode())
    fmt.Println("Request ID:      ", skyflowErr.GetRequestId())
    fmt.Println("Details:         ", skyflowErr.GetDetails())
    return
}
fmt.Println("Inserted fields:", resp.InsertedFields)
```
