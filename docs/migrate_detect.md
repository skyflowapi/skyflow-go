# Migrate to the newer Detect version (SkyVault)

This guide is for SkyVault users of the Go SDK (`github.com/skyflowapi/skyflow-go/v2`) who call Detect V1 methods such as `DeidentifyText`, `ReidentifyText` and `GetDetectRun`. Those methods and their types are removed. `client.Detect(vaultID)` now returns the newer version of the Detect service.

> [!NOTE]
> Using a FlowVault vault? FlowVault Detect is new and has nothing to migrate: see [Detect in the FlowVault README](../flowvault/README.md#detect).

## Table of Contents

- [What changed](#what-changed)
- [Method mapping](#method-mapping)
- [Configuration: stored or inline](#configuration-stored-or-inline)
- [De-identify text](#de-identify-text)
- [Re-identify text](#re-identify-text)
- [De-identify a file](#de-identify-a-file)
- [Get a run](#get-a-run)
- [Request headers](#request-headers)
- [New methods](#new-methods)
- [Migration checklist](#migration-checklist)

## What changed

- **Methods are renamed.** `DeidentifyText` is now `DeidentifyString`, `ReidentifyText` is `ReidentifyString`, and `GetDetectRun` is `GetRun`.
- **Detection settings move into a configuration.** In Detect V1, entities, token formats, regex lists and transformations were fields on the request. Now they go in a stored configuration (`ConfigurationId`) or an inline one (`Configuration`).
- **Token type is set per entity.** `TokenFormat` and its entity lists are gone. Each `Entity` carries its own `DeidentificationType`.
- **`DeidentifyFile` no longer waits by default.** It submits the file and returns the run ID. To wait, set `Poll` in the options. `WaitTime` is now a `time.Duration` with a 300s maximum, up from 64 seconds.
- **Responses are restructured.** Word and character counts move to `Metrics`. File outputs are a list in `Output`, entity positions are in `Location`, and the run status is a typed `DetectRunStatus`.
- **Per-call headers use an interceptor.** `CustomHeaders` in the options is replaced by `Interceptor`.
- **Two methods are new:** `ReidentifyFile` and `CheckGuardrails`.

## Method mapping

| Detect V1 | Newer version | Notes |
|---|---|---|
| `DeidentifyText(ctx, DeidentifyTextRequest, DeidentifyTextOptions)` | `DeidentifyString(ctx, DeidentifyStringRequest, DeidentifyStringOptions)` | Settings move into a configuration. |
| `ReidentifyText(ctx, ReidentifyTextRequest, ReidentifyTextOptions)` | `ReidentifyString(ctx, ReidentifyStringRequest, ReidentifyStringOptions)` | The three entity lists become `RedactionLevel` rules. |
| `DeidentifyFile(ctx, DeidentifyFileRequest, DeidentifyFileOptions)` | `DeidentifyFile(ctx, DeidentifyFileRequest, DeidentifyFileOptions)` | Same name, new request and response. It waits only when `Poll` is set. |
| `GetDetectRun(ctx, GetDetectRunRequest, GetDetectRunOptions)` | `GetRun(ctx, GetRunRequest, GetRunOptions)` | Returns `GetRunResponse`. |
| — | `ReidentifyFile`, `CheckGuardrails` | New. |

The call shape is unchanged: every method still takes `(ctx, request, options)` and returns `(*Response, *SkyflowError)`, with a nil `*SkyflowError` on success. Only the request, options and response types change. All types are still in `github.com/skyflowapi/skyflow-go/v2/utils/common`.

## Configuration: stored or inline

`DeidentifyString` and `DeidentifyFile` take their settings from a configuration. Both are optional and sent as given, as in the other Skyflow SDKs. With neither, the SDK sends a configuration holding only the vault ID and the server decides whether that is enough; with both, both are sent.

- `ConfigurationId`: a configuration stored in Studio or created through the REST API. Use it to manage detection settings outside your code.
- `Configuration`: an inline `*common.DetectConfiguration`. This is the closest match to Detect V1, where the settings were on the request.

Detect V1 request fields map into an inline configuration like this:

| Detect V1 field | Newer version |
|---|---|
| `Entities []DetectEntities` | `Configuration.Detect.Entities []Entity`: one `Entity` per type |
| `TokenFormat.VaultToken` / `EntityUniqueCounter` / `EntityOnly` | `Entity.DeidentificationType`: `DeidentificationTypeVaultToken`, `DeidentificationTypeEntityUniqueCounter` or `DeidentificationTypeEntityOnly` |
| `TokenFormat.DefaultType` | An `Entity` with `EntityType: EntityTypeAll` and the default `DeidentificationType`. It applies to every type without its own rule. |
| `AllowRegexList` | `Configuration.Detect.Skip` |
| `RestrictRegexList` | `Configuration.Detect.Restrict` |
| `Transformations.ShiftDates{MinDays, MaxDays, Entities}` | `Entity.Transformation.ShiftDates{MinDays, MaxDays}` on each date entity (`EntityTypeDate`, `EntityTypeDateInterval`, `EntityTypeDOB`). The bounds are `*int`: use `common.Ptr`. |
| — | New: `CustomTypes` (regex-based custom entities), `ReturnEntities`, and per-entity `IncludePrefix` / `PrefixOverride` |

Entity constants gain an `EntityType` prefix, and acronyms are upper-cased:

| Detect V1 | Newer version |
|---|---|
| `common.Name`, `common.EmailAddress`, `common.CreditCard`, ... | `common.EntityTypeName`, `common.EntityTypeEmailAddress`, `common.EntityTypeCreditCard`, ... |
| `common.Ssn`, `common.Dob`, `common.Cvv` | `common.EntityTypeSSN`, `common.EntityTypeDOB`, `common.EntityTypeCVV` |
| `common.IpAddress`, `common.Url`, `common.NumericalPii` | `common.EntityTypeIPAddress`, `common.EntityTypeURL`, `common.EntityTypeNumericalPII` |
| `common.VehicleId`, `common.OrganizationId`, `common.LocationZip` | `common.EntityTypeVehicleID`, `common.EntityTypeOrganizationID`, `common.EntityTypeLocationZIP` |
| `common.All` | `common.EntityTypeAll` |

## De-identify text

**Detect V1**

```go
response, err := service.DeidentifyText(ctx, common.DeidentifyTextRequest{
	Text:     "My name is John Doe, and my SSN is 123-45-6789. I was born on 1990-01-01.",
	Entities: []common.DetectEntities{common.Name, common.Ssn, common.Dob},
	TokenFormat: common.TokenFormat{
		DefaultType: common.TokenTypeDefaultEntityUnqCounter,
		VaultToken:  []common.DetectEntities{common.Ssn},
	},
	AllowRegexList:    []string{"Skyflow"},
	RestrictRegexList: []string{"EMP-[0-9]{6}"},
	Transformations: common.Transformations{
		ShiftDates: common.DateTransformation{MinDays: 10, MaxDays: 30, Entities: []common.TransformationsShiftDatesEntityTypesItem{common.TransformationsShiftDatesEntityTypesItemDob}},
	},
}, common.DeidentifyTextOptions{})
fmt.Println(response.ProcessedText, response.WordCount, response.CharCount)
```

**Newer version**

```go
response, err := service.DeidentifyString(ctx, common.DeidentifyStringRequest{
	Text: "My name is John Doe, and my SSN is 123-45-6789. I was born on 1990-01-01.",
	Configuration: &common.DetectConfiguration{
		Detect: &common.Detect{
			Entities: []common.Entity{
				{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeEntityUniqueCounter},
				{EntityType: common.EntityTypeSSN, DeidentificationType: common.DeidentificationTypeVaultToken, Destination: "<TABLE_NAME>.<COLUMN_NAME>"},
				{
					EntityType:           common.EntityTypeDOB,
					DeidentificationType: common.DeidentificationTypeEntityUniqueCounter,
					Transformation:       &common.Transformation{ShiftDates: &common.ShiftDates{MinDays: common.Ptr(10), MaxDays: common.Ptr(30)}},
				},
			},
			Skip:     []string{"Skyflow"},      // was AllowRegexList
			Restrict: []string{"EMP-[0-9]{6}"}, // was RestrictRegexList
		},
	},
	// ...or, instead of Configuration: ConfigurationId: "<CONFIGURATION_ID>",
}, common.DeidentifyStringOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
fmt.Println(response.ProcessedText)
if m := response.Metrics; m != nil && m.WordCount != nil && m.CharacterCount != nil {
	fmt.Println(*m.WordCount, *m.CharacterCount)
}
```

A vault token needs a `Destination`: the `"<TABLE_NAME>.<COLUMN_NAME>"` it is stored in.

**Response fields**

| Detect V1 (`DeidentifyTextResponse`) | Newer version (`DeidentifyStringResponse`) |
|---|---|
| `ProcessedText` | `ProcessedText` |
| `WordCount`, `CharCount` | `Metrics.WordCount`, `Metrics.CharacterCount`. `Metrics` and both counts are pointers (`*int`), nil when the server leaves them out. |
| `Entities []EntityInfo` | `Entities []DetectedEntity` |
| `EntityInfo.Entity` | `DetectedEntity.EntityType` (a string, so it can carry custom-type labels) |
| `EntityInfo.Scores` | `DetectedEntity.EntityScores` |
| `EntityInfo.TextIndex.Start` / `.End` | `DetectedEntity.Location.StartIndex` / `.EndIndex` |
| `EntityInfo.ProcessedIndex.Start` / `.End` | `DetectedEntity.Location.StartIndexProcessed` / `.EndIndexProcessed` |
| `EntityInfo.Token`, `EntityInfo.Value` | `Token`, `Value`. `Value` is set only with `ReturnEntities: common.ReturnEntitiesTypeAll`. |
| `Errors` | Removed: failures come back as the `*SkyflowError`. |

## Re-identify text

Only vault tokens can be re-identified. Placeholders from `ENTITY_UNIQUE_COUNTER`, `ENTITY_ONLY` or `MASK` aren't stored, so they come back unchanged.

The three Detect V1 entity lists become `RedactionLevel` rules, one per entity type:

| Detect V1 field | Newer version |
|---|---|
| `PlainTextEntities` | `common.RedactionLevel{EntityName: <type>, RedactionType: common.PLAIN_TEXT}` |
| `MaskedEntities` | `common.RedactionLevel{EntityName: <type>, RedactionType: common.MASKED}` |
| `RedactedEntities` | `common.RedactionLevel{EntityName: <type>, RedactionType: common.REDACTED}` |

Without `RedactionLevel`, every token gets its original value.

**Detect V1**

```go
response, err := service.ReidentifyText(ctx, common.ReidentifyTextRequest{
	Text:              "<TEXT_WITH_VAULT_TOKENS>",
	PlainTextEntities: []common.DetectEntities{common.Name},
	MaskedEntities:    []common.DetectEntities{common.Ssn},
}, common.ReidentifyTextOptions{})
```

**Newer version**

```go
response, err := service.ReidentifyString(ctx, common.ReidentifyStringRequest{
	Text: "<TEXT_WITH_VAULT_TOKENS>", // e.g. ProcessedText from a DeidentifyString call that used vault tokens
	RedactionLevel: []common.RedactionRule{
		common.RedactionLevel{EntityName: common.EntityTypeName, RedactionType: common.PLAIN_TEXT},
		common.RedactionLevel{EntityName: common.EntityTypeSSN, RedactionType: common.MASKED},
	},
}, common.ReidentifyStringOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
fmt.Println(response.ProcessedText)
```

**Response fields**

| Detect V1 (`ReidentifyTextResponse`) | Newer version (`ReidentifyStringResponse`) |
|---|---|
| `ProcessedText` | `ProcessedText` (unchanged) |
| `Errors` | Removed: failures come back as the `*SkyflowError`. |
| — | New: `Metrics *Metrics`, with the size, word count and character count of the processed text. Check it for nil before use. |

## De-identify a file

The biggest behaviour change: **`DeidentifyFile` no longer waits for the run.** By default it submits the file and returns only `RunId`. Fetch the result later with [`GetRun`](#get-a-run). To keep the Detect V1 behaviour, set `Poll`.

| Detect V1 | Newer version |
|---|---|
| `File: common.FileInput{FilePath: ...}` | `FilePath` on the request |
| `File: common.FileInput{File: f}` | `FileObject` on the request (you still close the file) |
| — | New: `DataSource` + `Value`, to send base64 content (`DeidentifySourceTypeBase64`), a presigned URL (`DeidentifySourceTypePresignedURL`) or a Skyflow ID (`DeidentifySourceTypeSkyflowID`) |
| Always waited | Waits only when `Options.Poll` is set |
| `WaitTime int` (seconds; default and max 64) | `Options.Poll.WaitTime time.Duration` (default 60s, max 300s), plus `Poll.MaxAttempts` (default and max 23) |
| `OutputDirectory` on the request | `Options.OutputDirectory`. Requires `Poll`. |
| `Entities`, `TokenFormat`, `AllowRegexList`, `RestrictRegexList`, `Transformations` | The configuration, as for [de-identify text](#de-identify-text) |
| `OutputProcessedImage`, `OutputOcrText`, `MaskingMethod` | `Configuration.Media.Image{OutputProcessedImage, OutputOcrText, MaskingMethod}`. The flags are `*bool`, so use `common.Ptr`. Masking methods are `common.MaskingMethodBlackbox` and `common.MaskingMethodBlur`. |
| `PixelDensity`, `MaxResolution` | `Configuration.Media.Document.Pdf{Density, MaxResolution}` (`*int`) |
| `OutputProcessedAudio`, `OutputTranscription` | `Configuration.Media.Audio{OutputProcessedAudio, OutputTranscription}` (`common.OutputTranscriptionType*`) |
| `Bleep{Gain, Frequency, StartPadding, StopPadding}` | `Configuration.Media.Audio.Bleep{...}`. The fields are pointers, and the paddings are `*float32`. |

`PLAINTEXT_TRANSCRIPTION` has no equivalent. The newer transcription types are `None`, `Transcription`, `MedicalTranscription`, `DiarizedTranscription` and `MedicalDiarizedTranscription`.

**Detect V1**

```go
response, err := service.DeidentifyFile(ctx, common.DeidentifyFileRequest{
	File:            common.FileInput{FilePath: "<PATH_TO_FILE>/scan.pdf"},
	Entities:        []common.DetectEntities{common.Name, common.Ssn},
	PixelDensity:    300,
	OutputDirectory: "<OUTPUT_DIRECTORY>",
	WaitTime:        30,
}, common.DeidentifyFileOptions{})
fmt.Println(response.Status, response.PageCount)
```

**Newer version**

```go
response, err := service.DeidentifyFile(ctx, common.DeidentifyFileRequest{
	FilePath: "<PATH_TO_FILE>/scan.pdf",
	Configuration: &common.DetectConfiguration{
		Detect: &common.Detect{
			Entities: []common.Entity{
				{EntityType: common.EntityTypeName, DeidentificationType: common.DeidentificationTypeEntityUniqueCounter},
				{EntityType: common.EntityTypeSSN, DeidentificationType: common.DeidentificationTypeEntityUniqueCounter},
			},
		},
		Media: &common.Media{Document: &common.Document{Pdf: &common.Pdf{Density: common.Ptr(300)}}},
	},
}, common.DeidentifyFileOptions{
	Poll:            &common.PollOptions{WaitTime: 30 * time.Second}, // leave Poll out to return as soon as the file is submitted
	OutputDirectory: "<OUTPUT_DIRECTORY>",
})
if err != nil {
	if response != nil { // a failure while waiting still returns the run ID
		fmt.Println("Run ID:", response.RunId)
	}
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
switch response.Status {
case common.DetectRunStatusSuccess:
	if response.Metrics != nil && response.Metrics.Pages != nil {
		fmt.Println("Pages:", *response.Metrics.Pages)
	}
case common.DetectRunStatusInProgress, common.DetectRunStatusQueued: // the wait ran out
	fmt.Println("Still running, run ID:", response.RunId)
default: // FAILED or UNKNOWN
	fmt.Println("Run failed:", response.Message)
}
```

**Response fields** (`DeidentifyFileResponse`, and `GetRunResponse` in the newer version)

| Detect V1 | Newer version |
|---|---|
| `RunId` | `RunId` |
| `Status string` (`IN_PROGRESS`, `FAILED`, `SUCCESS`) | `Status common.DetectRunStatus`: `DetectRunStatusSuccess`, `DetectRunStatusFailed`, `DetectRunStatusInProgress`, `DetectRunStatusQueued` or `DetectRunStatusUnknown`. It's empty when the call didn't poll. |
| `FileBase64`, `File`, `Type`, `Extension` | `Output []FileOutput`, with one entry per output: `ProcessedFile`, `ProcessedFileType` and `ProcessedFileExtension`. `OutputType` says whether `ProcessedFile` is base64 or a presigned URL. |
| `Entities []FileEntityInfo` | The `Output` entry whose `ProcessedFileType` is `ProcessedFileTypeEntities` |
| `WordCount`, `CharCount` | `Metrics.WordCount`, `Metrics.CharacterCount` (`*int`) |
| `SizeInKb` | `Metrics.Size` (`*float64`) |
| `PageCount`, `SlideCount`, `DurationInSeconds` | `Metrics.Pages`, `Metrics.Slides`, `Metrics.Duration`. These are pointers, and at most one is set. |
| `Errors` | `Message` for a failed run. Request failures come back as the `*SkyflowError`. |

Written outputs are still named `processed-<name>.<ext>`, and there's now also `processed-entities-<name>.json` (plus `processed-object-entities-<name>.json` for images). They're written only for a local file (`FilePath` or `FileObject`) whose polled run succeeded with base64 outputs.

## Get a run

| Detect V1 | Newer version |
|---|---|
| `GetDetectRunRequest{RunId}` | `GetRunRequest{RunId}`. Same field. A blank `RunId` is rejected before the call. |
| `GetDetectRunOptions{CustomHeaders}` | `GetRunOptions{Interceptor}`: see [Request headers](#request-headers) |
| Returned `*DeidentifyFileResponse` | Returns `*GetRunResponse`, with the same fields as the newer `DeidentifyFileResponse`: see the [file response fields](#de-identify-a-file). `RunId` is echoed from the request. |

**Detect V1**

```go
response, err := service.GetDetectRun(ctx, common.GetDetectRunRequest{RunId: "<RUN_ID>"}, common.GetDetectRunOptions{})
```

**Newer version**

```go
run, err := service.GetRun(ctx, common.GetRunRequest{RunId: "<RUN_ID>"}, common.GetRunOptions{})
if err != nil {
	fmt.Println(err.GetHttpCode(), err.GetMessage(), err.GetRequestId())
	return
}
if run.Status == common.DetectRunStatusSuccess {
	for _, output := range run.Output {
		fmt.Println(output.ProcessedFileType, output.ProcessedFileExtension)
	}
} else if run.Status == common.DetectRunStatusFailed {
	fmt.Println("Run failed:", run.Message)
}
```

`GetRun` writes nothing to disk. Each output in `Output` is base64 or a presigned URL, as `OutputType` says. Don't log presigned URLs: anyone holding one can download the file.

## Request headers

Client-level headers set with `WithCustomHeaders` are still sent with every Detect call. The per-call `CustomHeaders` map in the Detect V1 options is replaced by an `Interceptor`:

**Detect V1**

```go
options := common.DeidentifyTextOptions{
	CustomHeaders: map[common.CustomHeaderKey]string{common.RequestIdHeader: "<REQUEST_ID>"},
}
```

**Newer version**

```go
options := common.DeidentifyStringOptions{
	Interceptor: func(rc *common.RequestContext) {
		rc.AddHeader(common.RequestIdHeader, "<REQUEST_ID>")
	},
}
```

Every newer-version options struct (`DeidentifyStringOptions`, `DeidentifyFileOptions`, `GetRunOptions`, `ReidentifyStringOptions`, `ReidentifyFileOptions`, `CheckGuardrailsOptions`) has `Interceptor`.

## New methods

- **`ReidentifyFile`** re-identifies a file synchronously, sent as base64 or a presigned URL. The output comes back in the response, and the SDK writes nothing to disk. See [Reidentify File](../README.md#reidentify-file).
- **`CheckGuardrails`** screens text for toxicity and denied topics. See [Check Guardrails](../README.md#check-guardrails).

## Migration checklist

- [ ] Rename `DeidentifyText` to `DeidentifyString`, `ReidentifyText` to `ReidentifyString`, and `GetDetectRun` to `GetRun`.
- [ ] Move entities, token formats, regex lists and transformations into a `Configuration`, or switch to a stored `ConfigurationId`.
- [ ] Replace `TokenFormat` with a `DeidentificationType` on each `Entity`. Give vault-token entities a `Destination`.
- [ ] Rename entity constants (`common.Name` becomes `common.EntityTypeName`, `common.Ssn` becomes `common.EntityTypeSSN`, and so on).
- [ ] Replace `PlainTextEntities`, `MaskedEntities` and `RedactedEntities` with `RedactionLevel` rules.
- [ ] Move the `DeidentifyFile` input from `File` to `FilePath` or `FileObject`.
- [ ] Decide whether each `DeidentifyFile` call should wait. If it should, set `Poll`, with `WaitTime` as a `time.Duration`, and move `OutputDirectory` into the options.
- [ ] Move image, PDF and audio options into `Configuration.Media`.
- [ ] Read counts from `Metrics` (with a nil check), and file outputs from `Output`.
- [ ] Handle the `DetectRunStatus` values, including `QUEUED` and `UNKNOWN`, and read `Message` when a run fails.
- [ ] Replace `CustomHeaders` in options with `Interceptor`.
- [ ] Drop any reads of the removed `Errors` fields: check the `*SkyflowError` instead.
