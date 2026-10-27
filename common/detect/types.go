package detect

import (
	"os"
	"time"

	"github.com/skyflowapi/skyflow-go/common/vaultclient"
)

// Interceptor customises a single call, e.g. by adding headers. It is the same type the
// vault options use.
type Interceptor = vaultclient.Interceptor

// ---------------------------------------------------------------------------
// Shared response types
// ---------------------------------------------------------------------------

// Metrics describe the processed input. At most one of Slides, Pages and Duration is set.
type Metrics struct {
	Size           float64  `json:"size"` // kilobytes
	WordCount      int      `json:"wordCount"`
	CharacterCount int      `json:"characterCount"`
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

// RedactionLevel controls how one group of tokens is re-identified. Set exactly one source
// (TokenGroupName for FlowVault, EntityName for SkyVault) and exactly one replacement
// (RedactionPattern for FlowVault, RedactionType for SkyVault). The server checks that
// they suit the vault type.
type RedactionLevel struct {
	TokenGroupName   string        `json:"tokenGroupName,omitempty"`
	EntityName       EntityType    `json:"entityName,omitempty"`
	RedactionPattern string        `json:"redactionPattern,omitempty"`
	RedactionType    RedactionType `json:"redactionType,omitempty"`
}

// ---------------------------------------------------------------------------
// DeidentifyString
// ---------------------------------------------------------------------------

// DeidentifyStringRequest de-identifies sensitive data in Text. Set exactly one of
// ConfigurationId and Configuration.
type DeidentifyStringRequest struct {
	Text            string               `json:"text"`
	ConfigurationId string               `json:"configurationId,omitempty"`
	Configuration   *DetectConfiguration `json:"configuration,omitempty"`
}

// DeidentifyStringOptions customise a DeidentifyString call.
type DeidentifyStringOptions struct {
	Interceptor Interceptor
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
	Value    string         `json:"value"`
	Location EntityLocation `json:"location"`
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
// one of Value (with DataSource), FilePath or FileObject. Set exactly one of
// ConfigurationId and Configuration.
type DeidentifyFileRequest struct {
	// DataSource says what Value holds. A local file (FilePath or FileObject) is sent as
	// DataSourceTypeBase64, which is also the only value allowed with one.
	DataSource DataSourceType `json:"dataSource"`
	// Value is base64 content, a Skyflow ID or a presigned URL.
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
	Interceptor Interceptor
	// Poll, when set, waits for the run to finish (V1's behaviour). Nil, the default,
	// returns as soon as the run is submitted, with only RunId set.
	Poll *PollOptions
	// OutputDirectory is where a successful polled run's outputs are written for a local
	// file input. Empty means the current directory. Requires Poll.
	OutputDirectory string
}

// PollOptions bound how long DeidentifyFile waits for a run. Zero values use the defaults.
type PollOptions struct {
	// WaitTime is the longest DeidentifyFile waits. Default 60s, at most 300s.
	WaitTime time.Duration
	// MaxAttempts is the most run lookups made. Default 23, at most 23.
	MaxAttempts int
}

// DeidentifyFileResponse always has RunId. The other fields are set only when the run
// was polled: Status is empty when it wasn't, and Output, Metrics and OutputType are set
// once the run succeeded.
type DeidentifyFileResponse struct {
	RunId      string
	Status     DetectRunStatus
	OutputType DataSourceType
	Output     []FileOutput
	Metrics    *Metrics
	Message    string // failure detail when Status is DetectRunStatusFailed
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
	Interceptor Interceptor
}

// GetRunResponse is the current state of a run. Output, Metrics and OutputType are set
// once it succeeded.
type GetRunResponse struct {
	RunId      string          `json:"runId"` // echoed from the request
	Status     DetectRunStatus `json:"status"`
	OutputType DataSourceType  `json:"outputType"`
	Output     []FileOutput    `json:"output"`
	Metrics    *Metrics        `json:"metrics"`
	Message    string          `json:"message"` // failure detail when Status is DetectRunStatusFailed
}

// ---------------------------------------------------------------------------
// ReidentifyString
// ---------------------------------------------------------------------------

// ReidentifyStringRequest replaces the tokens in Text with their original values, or a
// redacted form per RedactionLevel.
type ReidentifyStringRequest struct {
	Text           string           `json:"text"`
	RedactionLevel []RedactionLevel `json:"redactionLevel,omitempty"`
}

// ReidentifyStringOptions customise a ReidentifyString call.
type ReidentifyStringOptions struct {
	Interceptor Interceptor
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
// Supply the file as exactly one of Value (with DataSource), FilePath or FileObject.
type ReidentifyFileRequest struct {
	DataSource     DataSourceType   `json:"dataSource"`
	Value          string           `json:"value"`
	DataFormat     FileDataFormat   `json:"dataFormat,omitempty"`
	RedactionLevel []RedactionLevel `json:"redactionLevel,omitempty"`

	FilePath   string   `json:"-"`
	FileObject *os.File `json:"-"`
}

// ReidentifyFileOptions customise a ReidentifyFile call.
type ReidentifyFileOptions struct {
	Interceptor Interceptor
	// OutputDirectory is where the outputs are written for a local file input. Empty means
	// the current directory.
	OutputDirectory string
}

// ReidentifyFileResponse is the re-identified file.
type ReidentifyFileResponse struct {
	Status     DetectRunStatus `json:"status"`
	OutputType DataSourceType  `json:"outputType"`
	Output     []FileOutput    `json:"output"`
	Metrics    *Metrics        `json:"metrics"`
}

// ---------------------------------------------------------------------------
// CheckGuardrails
// ---------------------------------------------------------------------------

// CheckGuardrailsRequest screens Text for toxicity and denied topics.
type CheckGuardrailsRequest struct {
	Text          string   `json:"text"`
	CheckToxicity bool     `json:"checkToxicity,omitempty"`
	DenyTopics    []string `json:"denyTopics,omitempty"`
}

// CheckGuardrailsOptions customise a CheckGuardrails call.
type CheckGuardrailsOptions struct {
	Interceptor Interceptor
}

// CheckGuardrailsResponse is the screening result.
type CheckGuardrailsResponse struct {
	Text        string               `json:"text"`
	Validation  GuardrailsValidation `json:"validation"`
	Toxic       bool                 `json:"toxic"`
	DeniedTopic bool                 `json:"deniedTopic"`
}
