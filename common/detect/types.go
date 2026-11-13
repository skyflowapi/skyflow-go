package detect

import (
	"os"
	"time"

	"github.com/skyflowapi/skyflow-go/common/config"
)

// CustomHeaderKey names a custom header a call can send. It is the type the vault options use.
type CustomHeaderKey = config.CustomHeaderKey

// ---------------------------------------------------------------------------
// Shared response types
// ---------------------------------------------------------------------------

// Metrics describe the processed input. Every field is optional in the API, so each is nil
// when the server leaves it out. At most one of Slides, Pages and
// Duration is set.
type Metrics struct {
	Size           *float64 `json:"size,omitempty"` // kilobytes
	WordCount      *int     `json:"wordCount,omitempty"`
	CharacterCount *int     `json:"characterCount,omitempty"`
	Slides         *int     `json:"slides,omitempty"`
	Pages          *int     `json:"pages,omitempty"`
	Duration       *float64 `json:"duration,omitempty"` // seconds
}

// FileOutput is one output of a file run: a base64 blob or a presigned URL, per the run's
// OutputType.
type FileOutput struct {
	ProcessedFile          string            `json:"processedFile"`
	ProcessedFileType      ProcessedFileType `json:"processedFileType"`
	ProcessedFileExtension FileDataFormat    `json:"processedFileExtension"`
}

// ReidentifyFileOutput is one output of ReidentifyFile, as FileOutput is of a de-identify
// file run. ProcessedFileExtension is empty for a format re-identification doesn't work with.
type ReidentifyFileOutput struct {
	ProcessedFile          string                   `json:"processedFile"`
	ProcessedFileType      ProcessedFileType        `json:"processedFileType"`
	ProcessedFileExtension ReidentifyFileDataFormat `json:"processedFileExtension"`
}

// RedactionRule is one element of ReidentifyStringRequest.RedactionLevel and
// ReidentifyFileRequest.RedactionLevel. Each SDK exports its own RedactionLevel that
// implements it, carrying the pair of fields its vault type uses: FlowVault names a
// TokenGroupName and a RedactionPattern (FlowRedactionLevel); SkyVault names an EntityName
// and a RedactionType (SkyRedactionLevel). Only those types implement it.
type RedactionRule interface {
	// apiFields returns the wire fields the rule carries; the others are empty.
	apiFields() (tokenGroupName, redactionPattern string, entityName EntityType, redactionType RedactionType)
}

// FlowRedactionLevel controls how one group of tokens is rendered when re-identifying on a
// FlowVault vault: the token group the rule applies to and the named redaction pattern,
// configured in the vault, to render its values with. Both fields are optional on the API
// and are sent as given.
type FlowRedactionLevel struct {
	TokenGroupName   string `json:"tokenGroupName,omitempty"`
	RedactionPattern string `json:"redactionPattern,omitempty"`
}

// SkyRedactionLevel controls how one group of tokens is rendered when re-identifying on a
// SkyVault vault: the entity type the rule applies to and the RedactionType to render it
// with. Both fields are optional on the API and are sent as given.
type SkyRedactionLevel struct {
	EntityName    EntityType    `json:"entityName,omitempty"`
	RedactionType RedactionType `json:"redactionType,omitempty"`
}

// ---------------------------------------------------------------------------
// DeidentifyString
// ---------------------------------------------------------------------------

// DeidentifyStringRequest de-identifies sensitive data in Text. ConfigurationId (a stored
// configuration) and Configuration (an inline one) are optional and sent as given; with
// neither, a configuration carrying only the vault ID is sent.
type DeidentifyStringRequest struct {
	Text            string               `json:"text"`
	ConfigurationId string               `json:"configurationId,omitempty"`
	Configuration   *DetectConfiguration `json:"configuration,omitempty"`
}

// DeidentifyStringOptions customise a DeidentifyString call.
type DeidentifyStringOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

// DeidentifyStringResponse is the de-identified text and what was found in it.
type DeidentifyStringResponse struct {
	ProcessedText string           `json:"processedText"`
	Entities      []DetectedEntity `json:"entities"`
	Metrics       *Metrics         `json:"metrics"`
}

// DetectedEntity is one entity found in the text.
type DetectedEntity struct {
	Token string `json:"token"` // e.g. "NAME_1"
	// Value is empty unless the configuration returns entities (ReturnEntitiesTypeAll).
	Value string `json:"value"`
	// Location is nil when the server doesn't send one.
	Location *EntityLocation `json:"location"`
	// EntityType is a string, not an EntityType, because it also carries custom-type labels.
	EntityType   string             `json:"entityType"`
	EntityScores map[string]float64 `json:"entityScores"` // 0.0 - 1.0
}

// EntityLocation gives an entity's position in the original and the processed text.
type EntityLocation struct {
	StartIndex          int `json:"startIndex"`
	EndIndex            int `json:"endIndex"`
	StartIndexProcessed int `json:"startIndexProcessed"`
	EndIndexProcessed   int `json:"endIndexProcessed"`
}

// ---------------------------------------------------------------------------
// DeidentifyFile
// ---------------------------------------------------------------------------

// DeidentifyFileRequest submits a file for de-identification. Supply the file as exactly
// one of Value (with DataSource), FilePath or FileObject. ConfigurationId (a stored
// configuration) and Configuration (an inline one) are optional and sent as given; with
// neither, a configuration carrying only the vault ID is sent.
type DeidentifyFileRequest struct {
	// DataSource says what Value holds, and is required with it: the SDK's
	// DeidentifySourceType. It is ignored for a local file (FilePath or FileObject), which
	// is always sent as BASE64.
	DataSource DeidentifyFileSource `json:"dataSource"`
	// Value is what DataSource names: base64 content, a presigned URL or (SkyVault) a
	// Skyflow ID.
	Value string `json:"value"`
	// DataFormat is optional: the server infers it, and for a local file the SDK infers
	// it from the extension. Set it to override either.
	DataFormat      FileDataFormat       `json:"dataFormat,omitempty"`
	ConfigurationId string               `json:"configurationId,omitempty"`
	Configuration   *DetectConfiguration `json:"configuration,omitempty"`

	// FilePath is a local file to read and send as base64.
	FilePath string `json:"-"`
	// FileObject is an open local file to read and send as base64. The caller closes it.
	FileObject *os.File `json:"-"`
}

// DeidentifyFileOptions customise a DeidentifyFile call.
type DeidentifyFileOptions struct {
	CustomHeaders map[CustomHeaderKey]string
	// Poll, when set, waits for the run to finish (V1's behaviour). Nil, the default,
	// returns as soon as the run is submitted, with only RunId set.
	Poll *PollOptions
	// OutputDirectory is where a successful polled run's outputs are written for a local
	// file input. Empty or blank means the current directory. Requires Poll. Each output is written
	// to a hidden temporary file first and renamed into place; a process that dies while
	// writing can leave such a ".processed-*.tmp" file behind.
	OutputDirectory string
}

// PollOptions bound how long DeidentifyFile waits for a run. A nil field uses its default;
// a set field must be in range, so a zero is rejected rather than taken
// as the default. &PollOptions{} polls with both defaults.
type PollOptions struct {
	// WaitTime caps the total pause between run lookups: 1s to 300s, default 60s. The
	// lookups' own time isn't counted, so a call can take a little longer.
	WaitTime *time.Duration
	// MaxAttempts is the most run lookups made: 1 to 23, default 23. The backoff fits at
	// most 7 lookups in 60s and 15 in 300s, so only a lower value stops polling earlier.
	MaxAttempts *int
}

// DeidentifyFileResponse always has RunId. The other fields are set only when the run
// was polled: Status is empty when it wasn't, and Output, Metrics and OutputType are set
// once the run succeeded.
type DeidentifyFileResponse struct {
	RunId      string               `json:"runId"`
	Status     DetectRunStatus      `json:"status"`
	OutputType DeidentifyFileSource `json:"outputType"`
	Output     []FileOutput         `json:"output"`
	Metrics    *Metrics             `json:"metrics"`
	Message    string               `json:"message"` // failure detail when Status is DetectRunStatusFailed
}

// ---------------------------------------------------------------------------
// GetRun
// ---------------------------------------------------------------------------

// GetRunRequest looks up a de-identify file run.
type GetRunRequest struct {
	RunId string
}

// GetRunOptions customise a GetRun call.
type GetRunOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

// GetRunResponse is the current state of a run. Output, Metrics and OutputType are set
// once it succeeded.
type GetRunResponse struct {
	RunId      string               `json:"runId"` // echoed from the request
	Status     DetectRunStatus      `json:"status"`
	OutputType DeidentifyFileSource `json:"outputType"`
	Output     []FileOutput         `json:"output"`
	Metrics    *Metrics             `json:"metrics"`
	Message    string               `json:"message"` // failure detail when Status is DetectRunStatusFailed
}

// ---------------------------------------------------------------------------
// ReidentifyString
// ---------------------------------------------------------------------------

// ReidentifyStringRequest replaces the tokens in Text with their original values, or a
// redacted form per RedactionLevel, built with the SDK's RedactionLevel.
type ReidentifyStringRequest struct {
	Text           string          `json:"text"`
	RedactionLevel []RedactionRule `json:"redactionLevel,omitempty"`
}

// ReidentifyStringOptions customise a ReidentifyString call.
type ReidentifyStringOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

// ReidentifyStringResponse is the re-identified text.
type ReidentifyStringResponse struct {
	ProcessedText string   `json:"processedText"`
	Metrics       *Metrics `json:"metrics"`
}

// ---------------------------------------------------------------------------
// ReidentifyFile
// ---------------------------------------------------------------------------

// ReidentifyFileRequest re-identifies a file. It is synchronous: there is no run to poll.
// DataSource and Value are required; to send a local file, base64-encode it into Value.
type ReidentifyFileRequest struct {
	// DataSource says what Value holds: the SDK's ReidentifySourceType.
	DataSource     ReidentifyFileSource     `json:"dataSource"`
	Value          string                   `json:"value"` // base64 content or a presigned URL
	DataFormat     ReidentifyFileDataFormat `json:"dataFormat,omitempty"`
	RedactionLevel []RedactionRule          `json:"redactionLevel,omitempty"`
}

// ReidentifyFileOptions customise a ReidentifyFile call.
type ReidentifyFileOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

// ReidentifyFileResponse is the re-identified file. The SDK writes nothing to disk: Output
// holds base64 content or presigned URLs, as OutputType says, for the caller to handle.
type ReidentifyFileResponse struct {
	Status DetectRunStatus `json:"status"`
	// OutputType is the SDK's ReidentifySourceType, or nil for a value the SDK doesn't know.
	OutputType ReidentifyFileSource   `json:"outputType"`
	Output     []ReidentifyFileOutput `json:"output"`
	Metrics    *Metrics               `json:"metrics"`
}

// ---------------------------------------------------------------------------
// CheckGuardrails
// ---------------------------------------------------------------------------

// CheckGuardrailsRequest screens Text for toxicity and denied topics.
type CheckGuardrailsRequest struct {
	Text string `json:"text"`
	// CheckToxicity nil leaves the field out of the request, and the server then doesn't
	// check toxicity: it checks only when CheckToxicity is true.
	CheckToxicity *bool    `json:"checkToxicity,omitempty"`
	DenyTopics    []string `json:"denyTopics,omitempty"`
}

// CheckGuardrailsOptions customise a CheckGuardrails call.
type CheckGuardrailsOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

// CheckGuardrailsResponse is the screening result.
type CheckGuardrailsResponse struct {
	Text       string               `json:"text"`
	Validation GuardrailsValidation `json:"validation"`
	// Toxic and DeniedTopic are nil when the API leaves them out.
	Toxic       *bool `json:"toxic"`
	DeniedTopic *bool `json:"deniedTopic"`
}
