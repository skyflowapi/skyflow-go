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
| `Detect(vaultID ...string)` | `(*DetectService, *SkyflowError)` | Get the Detect service. Uses the first configured vault if no ID is given. |

### Instance management methods

All mutating methods return `*SkyflowError` (nil on success) unless the table notes otherwise.

| Method | Returns | Description |
|--------|---------|-------------|
| `AddVaultConfig(VaultConfig)` | `*SkyflowError` | Add a vault after initialization. |
| `GetVaultConfig(vaultId string)` | `(*VaultConfig, *SkyflowError)` | Retrieve a vault configuration by ID. |
| `UpdateVaultConfig(VaultConfig)` | `*SkyflowError` | Replace an existing vault configuration (matched by `VaultId`). |
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

#### `DeidentifyTextRequest`

Passed to `detect.DeidentifyText()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Text` | `string` | _(required)_ | Text to de-identify. |
| `Entities` | `[]DetectEntities` | `nil` | Entity types to detect. Detects all types if empty. See [`DetectEntities`](#detectentities). |
| `AllowRegexList` | `[]string` | `nil` | Regex patterns to always treat as detectable. |
| `RestrictRegexList` | `[]string` | `nil` | Regex patterns to exclude from detection. |
| `TokenFormat` | `TokenFormat` | — | Token format per entity type. See [`TokenFormat`](#tokenformat). |
| `Transformations` | `Transformations` | — | Data transformations (e.g. date shifting). See [`Transformations`](#transformations). |

#### `DeidentifyTextOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `ReidentifyTextRequest`

Passed to `detect.ReidentifyText()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Text` | `string` | _(required)_ | The de-identified text to re-identify. |
| `RedactedEntities` | `[]DetectEntities` | `nil` | Entity types to keep redacted in the output. |
| `MaskedEntities` | `[]DetectEntities` | `nil` | Entity types to mask in the output. |
| `PlainTextEntities` | `[]DetectEntities` | `nil` | Entity types to reveal as plain text. |

#### `ReidentifyTextOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `DeidentifyFileRequest`

Passed to `detect.DeidentifyFile()`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `File` | `FileInput` | _(required)_ | File source. See [`FileInput`](#fileinput). |
| `Entities` | `[]DetectEntities` | `nil` | Entity types to detect. |
| `AllowRegexList` | `[]string` | `nil` | Regex patterns to always detect. |
| `RestrictRegexList` | `[]string` | `nil` | Regex patterns to exclude. |
| `TokenFormat` | `TokenFormat` | — | Token format per entity type. |
| `Transformations` | `Transformations` | — | Transformations (not supported for images/PDFs). |
| `OutputProcessedImage` | `bool` | `false` | Include the processed image in the response. |
| `OutputOcrText` | `bool` | `false` | Include OCR-extracted text in the response. |
| `MaskingMethod` | `MaskingMethod` | — | Visual masking method for images. See [`MaskingMethod`](#maskingmethod). |
| `PixelDensity` | `int` | `0` | Pixel density for PDF processing. |
| `MaxResolution` | `int` | `0` | Maximum resolution for PDF processing. |
| `OutputProcessedAudio` | `bool` | `false` | Include processed audio in the response. |
| `OutputTranscription` | `DetectOutputTranscriptions` | — | Transcription mode for audio. See [`DetectOutputTranscriptions`](#detectoutputtranscriptions). |
| `Bleep` | `AudioBleep` | — | Audio bleep configuration. See [`AudioBleep`](#audiobleep). |
| `OutputDirectory` | `string` | `""` | Directory to write the processed file. |
| `WaitTime` | `int` | `0` | Maximum seconds to wait for async file processing (≤ 64). |

#### `DeidentifyFileOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

---

#### `GetDetectRunRequest`

Passed to `detect.GetDetectRun()` to poll the status of an async `DeidentifyFile` call.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `RunId` | `string` | _(required)_ | The `RunId` returned by a prior `DeidentifyFile` response. |

#### `GetDetectRunOptions`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `CustomHeaders` | `map[CustomHeaderKey]string` | `nil` | Per-request custom headers. |

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

### `DeidentifyTextResponse`

Returned by `detect.DeidentifyText()`.

| Field | Type | Description |
|-------|------|-------------|
| `ProcessedText` | `string` | The de-identified text with entity values replaced by tokens. |
| `Entities` | `[]EntityInfo` | Detected entities. See [`EntityInfo`](#entityinfo). |
| `WordCount` | `int` | Word count of the input text. |
| `CharCount` | `int` | Character count of the input text. |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

### `ReidentifyTextResponse`

Returned by `detect.ReidentifyText()`.

| Field | Type | Description |
|-------|------|-------------|
| `ProcessedText` | `string` | The re-identified text with tokens replaced by their original values. |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

### `DeidentifyFileResponse`

Returned by both `detect.DeidentifyFile()` and `detect.GetDetectRun()`.

| Field | Type | Description |
|-------|------|-------------|
| `File` | `FileInfo` | Metadata about the processed file. See [`FileInfo`](#fileinfo). |
| `FileBase64` | `string` | Base64-encoded processed file content (when `OutputProcessedImage: true`). |
| `Type` | `string` | MIME type of the output file. |
| `Extension` | `string` | File extension of the output file. |
| `WordCount` | `int` | Word count (text and document files). |
| `CharCount` | `int` | Character count (text and document files). |
| `SizeInKb` | `float64` | Output file size in kilobytes. |
| `DurationInSeconds` | `float64` | Duration in seconds (audio and video files). |
| `PageCount` | `int` | Page count (PDF files). |
| `SlideCount` | `int` | Slide count (presentation files). |
| `Entities` | `[]FileEntityInfo` | Detected entities. See [`FileEntityInfo`](#fileentityinfo). |
| `RunId` | `string` | Run ID for polling async processing with `GetDetectRun()`. |
| `Status` | `string` | Processing status. See [`DeidentifyFileStatus`](#deidentifyfilestatus). |
| `Errors` | `[]map[string]interface{}` | Errors, if any. |

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

### `EntityInfo`

Each element in `DeidentifyTextResponse.Entities`.

| Field | Type | Description |
|-------|------|-------------|
| `Token` | `string` | The token that replaced the original entity value. |
| `Value` | `string` | The original entity value. |
| `Entity` | `string` | Entity type label (e.g. `"email_address"`). |
| `Scores` | `map[string]float64` | Confidence scores per entity type. |
| `ProcessedIndex` | `TextIndex` | Character offsets of the token in the processed text. |
| `TextIndex` | `TextIndex` | Character offsets of the entity in the original text. |

### `TextIndex`

Used in [`EntityInfo`](#entityinfo).

| Field | Type | Description |
|-------|------|-------------|
| `Start` | `int` | Start character offset (inclusive). |
| `End` | `int` | End character offset (exclusive). |

### `FileEntityInfo`

Each element in `DeidentifyFileResponse.Entities`.

| Field | Type | Description |
|-------|------|-------------|
| `File` | `string` | File name or identifier. |
| `Type` | `string` | Output file type. |
| `Extension` | `string` | Output file extension. |

### `FileInfo`

Returned in `DeidentifyFileResponse.File`.

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Original file name. |
| `Size` | `int64` | File size in bytes. |
| `Type` | `string` | MIME type. |
| `LastModified` | `int64` | Last-modified timestamp (milliseconds since Unix epoch). |

### `FileInput`

Used inside [`DeidentifyFileRequest`](#deidentifyfilerequest). Provide exactly one source field.

| Field | Type | Description |
|-------|------|-------------|
| `FilePath` | `string` | Path to a local file. |
| `File` | `*os.File` | An open file handle. |

### `AudioBleep`

Used inside [`DeidentifyFileRequest`](#deidentifyfilerequest) to configure audio bleeping.

| Field | Type | Description |
|-------|------|-------------|
| `Gain` | `int` | Gain level of the bleep tone. |
| `Frequency` | `int` | Frequency (Hz) of the bleep tone. |
| `StartPadding` | `float64` | Seconds of silence before the bleep. |
| `StopPadding` | `float64` | Seconds of silence after the bleep. |

### `TokenFormat`

Used inside [`DeidentifyTextRequest`](#deidentifytextrequest) and [`DeidentifyFileRequest`](#deidentifyfilerequest) to control the token type per entity.

| Field | Type | Description |
|-------|------|-------------|
| `DefaultType` | `TokenTypeDefault` | Default token type for entities not explicitly listed. See [`TokenTypeDefault`](#tokentypedefault). |
| `VaultToken` | `[]DetectEntities` | Entities to tokenize as vault tokens. |
| `EntityUniqueCounter` | `[]DetectEntities` | Entities to tokenize as entity-unique-counter tokens. |
| `EntityOnly` | `[]DetectEntities` | Entities to tokenize as entity-only tokens. |

### `Transformations`

Used inside [`DeidentifyTextRequest`](#deidentifytextrequest) and [`DeidentifyFileRequest`](#deidentifyfilerequest).

| Field | Type | Description |
|-------|------|-------------|
| `ShiftDates` | `DateTransformation` | Date-shifting transformation to apply. |

### `DateTransformation`

Used inside [`Transformations`](#transformations).

| Field | Type | Description |
|-------|------|-------------|
| `MaxDays` | `int` | Maximum number of days to shift. |
| `MinDays` | `int` | Minimum number of days to shift. |
| `Entities` | `[]TransformationsShiftDatesEntityTypesItem` | Entity types to shift. Valid values: `Date`, `DateInterval`, `Dob`. |

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

### `RedactionType`

`github.com/skyflowapi/skyflow-go/v2/utils/common`

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

### `MaskingMethod`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — used in [`DeidentifyFileRequest`](#deidentifyfilerequest).

| Value | Description |
|-------|-------------|
| `BLACKBOX` | Cover detected entities with a solid black rectangle. |
| `BLUR` | Blur detected entities with a Gaussian blur. |

### `DetectOutputTranscriptions`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — used in [`DeidentifyFileRequest`](#deidentifyfilerequest).

| Value | Description |
|-------|-------------|
| `PLAINTEXT_TRANSCRIPTION` | Standard plain-text transcription. |
| `DIARIZED_TRANSCRIPTION` | Transcription with speaker diarization. |
| `TRANSCRIPTION` | Basic transcription. |
| `MEDICAL_TRANSCRIPTION` | Medical-domain transcription. |
| `MEDICAL_DIARIZED_TRANSCRIPTION` | Medical transcription with speaker diarization. |

### `DeidentifyFileStatus`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — returned in `DeidentifyFileResponse.Status`.

| Value | Description |
|-------|-------------|
| `IN_PROGRESS` | File processing is ongoing. Poll with `GetDetectRun()`. |
| `SUCCESS` | Processing completed successfully. |
| `FAILED` | Processing failed. |

### `TokenTypeDefault`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — used in [`TokenFormat`](#tokenformat).

| Value | Description |
|-------|-------------|
| `TokenTypeDefaultEntityOnly` | Token represents the entity type only; no value is stored in the vault. |
| `TokenTypeDefaultEntityUnqCounter` | Deterministic token unique to the entity value (consistent across occurrences). |
| `TokenTypeDefaultVaultToken` | Token stored in the vault and associated with a `skyflow_id`. |

### `CustomHeaderKey`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — keys for custom header maps.

| Value | Wire header | Description |
|-------|-------------|-------------|
| `SkyflowAccountId` | `x-skyflow-account-id` | Skyflow account ID header. |
| `SkyflowAccountName` | `x-skyflow-account-name` | Skyflow account name header. |
| `RequestIdHeader` | `x-request-id` | Custom request ID for tracing. |

### `DetectEntities`

`github.com/skyflowapi/skyflow-go/v2/utils/common` — entity types for Detect operations. Use `All` to detect all supported types.

| Category | Values |
|----------|--------|
| Personal identity | `Name`, `NameGiven`, `NameFamily`, `NameMedicalProfessional`, `Dob`, `Age`, `Gender`, `MaritalStatus`, `Sexuality`, `Religion`, `PoliticalAffiliation`, `PhysicalAttribute`, `Origin` |
| Contact | `EmailAddress`, `PhoneNumber`, `Location`, `LocationAddress`, `LocationAddressStreet`, `LocationCity`, `LocationState`, `LocationCountry`, `LocationZip`, `LocationCoordinate` |
| Financial | `CreditCard`, `CreditCardExpiration`, `Cvv`, `BankAccount`, `AccountNumber`, `RoutingNumber`, `Money`, `FinancialMetric`, `CorporateAction` |
| Government ID | `Ssn`, `DriverLicense`, `PassportNumber`, `HealthcareNumber`, `OrganizationId`, `NumericalPii` |
| Medical | `BloodType`, `Condition`, `Drug`, `Dose`, `Effect`, `Injury`, `MedicalCode`, `MedicalProcess`, `OrganizationMedicalFacility` |
| Date & time | `Date`, `Day`, `Month`, `Year`, `Time`, `Duration`, `DateInterval`, `Event` |
| Technology | `IpAddress`, `Url`, `Username`, `Password`, `Filename`, `VehicleId` |
| Organization | `Organization`, `Occupation`, `Project`, `Product` |
| Other | `Language`, `Statistics`, `Trend`, `ZodiacSign`, `All` |

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
