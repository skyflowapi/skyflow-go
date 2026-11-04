package detect

// DetectConfiguration is an inline detect policy for DeidentifyString and DeidentifyFile, used
// instead of a stored ConfigurationId. Create stored configurations in Studio or through
// the REST API; the SDK does not manage them. The SDK fills in the vault ID.
type DetectConfiguration struct {
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Detect      *Detect       `json:"detect,omitempty"`
	FileMapping []FileMapping `json:"fileMapping,omitempty"`
	Media       *Media        `json:"media,omitempty"` // file calls only
}

// Detect chooses what is detected and how it is de-identified.
type Detect struct {
	Entities       []Entity           `json:"entities,omitempty"`
	ObjectEntities []ObjectEntity     `json:"objectEntities,omitempty"` // file calls only (image/video)
	Restrict       []string           `json:"restrict,omitempty"`       // regexes always treated as restricted
	Skip           []string           `json:"skip,omitempty"`           // regexes never treated as sensitive
	CustomTypes    []CustomType       `json:"customTypes,omitempty"`
	ReturnEntities ReturnEntitiesType `json:"returnEntities,omitempty"`
}

// Entity sets how one entity type is de-identified.
type Entity struct {
	EntityType           EntityType           `json:"entityType,omitempty"`
	DeidentificationType DeidentificationType `json:"deidentificationType,omitempty"`
	// Destination is the token group (FlowVault) or "table.column" (SkyVault) for
	// DeidentificationTypeVaultToken.
	Destination    string          `json:"destination,omitempty"`
	Transformation *Transformation `json:"transformation,omitempty"`
	IncludePrefix  IncludePrefix   `json:"includePrefix,omitempty"`
	PrefixOverride string          `json:"prefixOverride,omitempty"` // not valid with IncludePrefixNo
}

// Transformation changes a detected value instead of replacing it.
type Transformation struct {
	ShiftDates *ShiftDates `json:"shiftDates,omitempty"`
}

// ShiftDates moves dates by a random number of days in [MinDays, MaxDays]. A nil bound is
// left to the server.
type ShiftDates struct {
	MinDays *int `json:"minDays,omitempty"`
	MaxDays *int `json:"maxDays,omitempty"`
}

// CustomType detects values matching regexes as a custom entity. Label becomes the
// EntityType on detected entities.
type CustomType struct {
	Label                string               `json:"label,omitempty"`
	Match                []string             `json:"match,omitempty"`
	DeidentificationType DeidentificationType `json:"deidentificationType,omitempty"`
	Destination          string               `json:"destination,omitempty"`
	IncludePrefix        IncludePrefix        `json:"includePrefix,omitempty"`
	PrefixOverride       string               `json:"prefixOverride,omitempty"`
}

// ObjectEntity sets how one object type in images and video is treated.
type ObjectEntity struct {
	EntityType           ObjectEntityType           `json:"entityType,omitempty"`
	DeidentificationType ObjectDeidentificationType `json:"deidentificationType,omitempty"`
}

// FileMapping maps a source file location to where its outputs go.
type FileMapping struct {
	Source                    string `json:"source,omitempty"`
	Destination               string `json:"destination,omitempty"`
	OutputFileNameSuffix      string `json:"outputFileNameSuffix,omitempty"`
	EntitiesDestination       string `json:"entitiesDestination,omitempty"`
	ObjectEntitiesDestination string `json:"objectEntitiesDestination,omitempty"`
}

// Media holds the per-media-type options of a file run. Their option fields are pointers:
// nil leaves the option to the server default, and a pointer to a zero value sends zero.
type Media struct {
	Audio    *Audio    `json:"audio,omitempty"`
	Document *Document `json:"document,omitempty"`
	Image    *Image    `json:"image,omitempty"`
}

// Audio options.
type Audio struct {
	OutputProcessedAudio *bool                   `json:"outputProcessedAudio,omitempty"`
	OutputTranscription  OutputTranscriptionType `json:"outputTranscription,omitempty"`
	Bleep                *Bleep                  `json:"bleep,omitempty"`
}

// Bleep sets the tone that replaces sensitive audio.
type Bleep struct {
	StartPadding *float32 `json:"startPadding,omitempty"` // seconds, 0–20; server default 0.5
	StopPadding  *float32 `json:"stopPadding,omitempty"`  // seconds, 0–20; server default 0.2
	Frequency    *int     `json:"frequency,omitempty"`    // Hz
	Gain         *int     `json:"gain,omitempty"`
}

// Document options.
type Document struct {
	Pdf *Pdf `json:"pdf,omitempty"`
}

// Pdf options.
type Pdf struct {
	ProcessingMode PdfProcessingMode `json:"processingMode,omitempty"`
	Density        *int              `json:"density,omitempty"`
	MaxResolution  *int              `json:"maxResolution,omitempty"`
}

// Image options.
type Image struct {
	OutputProcessedImage *bool         `json:"outputProcessedImage,omitempty"`
	OutputOcrText        *bool         `json:"outputOcrText,omitempty"`
	MaskingMethod        MaskingMethod `json:"maskingMethod,omitempty"`
}
