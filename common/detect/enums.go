package detect

// The values below mirror the enums in detect.proto. Each enum is a string holding the
// proto value name, so a value the server adds later still decodes and round-trips.

// EntityType is an entity Detect can find in text, files and media. EntityTypeAll selects every entity type.
type EntityType string

const (
	EntityTypeNone                        EntityType = "NONE"
	EntityTypeBankAccount                 EntityType = "BANK_ACCOUNT"
	EntityTypeCreditCard                  EntityType = "CREDIT_CARD"
	EntityTypeCreditCardExpiration        EntityType = "CREDIT_CARD_EXPIRATION"
	EntityTypeCVV                         EntityType = "CVV"
	EntityTypeDate                        EntityType = "DATE"
	EntityTypeDateInterval                EntityType = "DATE_INTERVAL"
	EntityTypeDOB                         EntityType = "DOB"
	EntityTypeDriverLicense               EntityType = "DRIVER_LICENSE"
	EntityTypeEmailAddress                EntityType = "EMAIL_ADDRESS"
	EntityTypeHealthcareNumber            EntityType = "HEALTHCARE_NUMBER"
	EntityTypeIPAddress                   EntityType = "IP_ADDRESS"
	EntityTypeLocation                    EntityType = "LOCATION"
	EntityTypeName                        EntityType = "NAME"
	EntityTypeNumericalPII                EntityType = "NUMERICAL_PII"
	EntityTypePhoneNumber                 EntityType = "PHONE_NUMBER"
	EntityTypeSSN                         EntityType = "SSN"
	EntityTypeURL                         EntityType = "URL"
	EntityTypeVehicleID                   EntityType = "VEHICLE_ID"
	EntityTypeMedicalCode                 EntityType = "MEDICAL_CODE"
	EntityTypeNameFamily                  EntityType = "NAME_FAMILY"
	EntityTypeNameGiven                   EntityType = "NAME_GIVEN"
	EntityTypeAccountNumber               EntityType = "ACCOUNT_NUMBER"
	EntityTypeEvent                       EntityType = "EVENT"
	EntityTypeFilename                    EntityType = "FILENAME"
	EntityTypeGender                      EntityType = "GENDER"
	EntityTypeLanguage                    EntityType = "LANGUAGE"
	EntityTypeLocationAddress             EntityType = "LOCATION_ADDRESS"
	EntityTypeLocationCity                EntityType = "LOCATION_CITY"
	EntityTypeLocationCoordinate          EntityType = "LOCATION_COORDINATE"
	EntityTypeLocationCountry             EntityType = "LOCATION_COUNTRY"
	EntityTypeLocationState               EntityType = "LOCATION_STATE"
	EntityTypeLocationZIP                 EntityType = "LOCATION_ZIP"
	EntityTypeMaritalStatus               EntityType = "MARITAL_STATUS"
	EntityTypeMoney                       EntityType = "MONEY"
	EntityTypeNameMedicalProfessional     EntityType = "NAME_MEDICAL_PROFESSIONAL"
	EntityTypeOccupation                  EntityType = "OCCUPATION"
	EntityTypeOrganization                EntityType = "ORGANIZATION"
	EntityTypeOrganizationMedicalFacility EntityType = "ORGANIZATION_MEDICAL_FACILITY"
	EntityTypeOrigin                      EntityType = "ORIGIN"
	EntityTypePassportNumber              EntityType = "PASSPORT_NUMBER"
	EntityTypePassword                    EntityType = "PASSWORD"
	EntityTypePhysicalAttribute           EntityType = "PHYSICAL_ATTRIBUTE"
	EntityTypePoliticalAffiliation        EntityType = "POLITICAL_AFFILIATION"
	EntityTypeReligion                    EntityType = "RELIGION"
	EntityTypeTime                        EntityType = "TIME"
	EntityTypeUsername                    EntityType = "USERNAME"
	EntityTypeZodiacSign                  EntityType = "ZODIAC_SIGN"
	EntityTypeBloodType                   EntityType = "BLOOD_TYPE"
	EntityTypeCondition                   EntityType = "CONDITION"
	EntityTypeDose                        EntityType = "DOSE"
	EntityTypeDrug                        EntityType = "DRUG"
	EntityTypeInjury                      EntityType = "INJURY"
	EntityTypeMedicalProcess              EntityType = "MEDICAL_PROCESS"
	EntityTypeStatistics                  EntityType = "STATISTICS"
	EntityTypeRoutingNumber               EntityType = "ROUTING_NUMBER"
	EntityTypeCorporateAction             EntityType = "CORPORATE_ACTION"
	EntityTypeFinancialMetric             EntityType = "FINANCIAL_METRIC"
	EntityTypeProduct                     EntityType = "PRODUCT"
	EntityTypeTrend                       EntityType = "TREND"
	EntityTypeDuration                    EntityType = "DURATION"
	EntityTypeLocationAddressStreet       EntityType = "LOCATION_ADDRESS_STREET"
	EntityTypeAge                         EntityType = "AGE"
	EntityTypeSexuality                   EntityType = "SEXUALITY"
	EntityTypeEffect                      EntityType = "EFFECT"
	EntityTypeProject                     EntityType = "PROJECT"
	EntityTypeOrganizationID              EntityType = "ORGANIZATION_ID"
	EntityTypeDay                         EntityType = "DAY"
	EntityTypeMonth                       EntityType = "MONTH"
	EntityTypeYear                        EntityType = "YEAR"
	EntityTypeAll                         EntityType = "ALL"
)

// DeidentificationType is how a detected entity is replaced. The server default is ENTITY_UNIQUE_COUNTER.
type DeidentificationType string

const (
	DeidentificationTypeUnknown             DeidentificationType = "UNKNOWN"
	DeidentificationTypeEntityUniqueCounter DeidentificationType = "ENTITY_UNIQUE_COUNTER"
	DeidentificationTypeEntityOnly          DeidentificationType = "ENTITY_ONLY"
	DeidentificationTypeVaultToken          DeidentificationType = "VAULT_TOKEN"
	DeidentificationTypeMask                DeidentificationType = "MASK"
)

// IncludePrefix is whether the replacement carries the entity-type prefix. The server default is YES.
type IncludePrefix string

const (
	IncludePrefixYes IncludePrefix = "YES"
	IncludePrefixNo  IncludePrefix = "NO"
)

// ReturnEntitiesType is which detected values the response carries. The server default is EXCLUDE_SENSITIVE_DATA.
type ReturnEntitiesType string

const (
	ReturnEntitiesTypeExcludeSensitiveData ReturnEntitiesType = "EXCLUDE_SENSITIVE_DATA"
	ReturnEntitiesTypeAll                  ReturnEntitiesType = "ALL"
	ReturnEntitiesTypeNone                 ReturnEntitiesType = "NONE"
)

// RedactionType is how a re-identified value is shown. SkyVault vaults only; FlowVault vaults use RedactionLevel.RedactionPattern.
type RedactionType string

const (
	// Deprecated: the server rejects DEFAULT in a redaction level. Use REDACTED, MASKED or PLAIN_TEXT.
	RedactionTypeDefault   RedactionType = "DEFAULT"
	RedactionTypeRedacted  RedactionType = "REDACTED"
	RedactionTypeMasked    RedactionType = "MASKED"
	RedactionTypePlainText RedactionType = "PLAIN_TEXT"
)

// DeidentifyFileSource is what a DeidentifyFileRequest's Value holds, and the OutputType of
// DeidentifyFileResponse and GetRunResponse. Each SDK exports its own DeidentifySourceType
// that implements it, with the sources its vault type accepts: FlowVault takes BASE64 and
// PRESIGNED_URL (FlowDeidentifySourceType); SkyVault also takes SKYFLOW_ID
// (SkyDeidentifySourceType). Only those types implement it.
type DeidentifyFileSource interface {
	// deidentifySource returns the wire value.
	deidentifySource() string
}

// base64Value and presignedURLValue are wire values every vault type's source sets have.
const (
	base64Value       = "BASE64"
	presignedURLValue = "PRESIGNED_URL"
)

// FlowDeidentifySourceType is a DeidentifyFileSource for a FlowVault vault.
type FlowDeidentifySourceType string

const (
	FlowDeidentifySourceTypeBase64       FlowDeidentifySourceType = base64Value
	FlowDeidentifySourceTypePresignedURL FlowDeidentifySourceType = presignedURLValue
)

// SkyDeidentifySourceType is a DeidentifyFileSource for a SkyVault vault.
type SkyDeidentifySourceType string

const (
	SkyDeidentifySourceTypeBase64       SkyDeidentifySourceType = base64Value
	SkyDeidentifySourceTypeSkyflowID    SkyDeidentifySourceType = "SKYFLOW_ID"
	SkyDeidentifySourceTypePresignedURL SkyDeidentifySourceType = presignedURLValue
)

// ReidentifyFileSource is what a ReidentifyFileRequest's Value holds, and the OutputType of
// ReidentifyFileResponse. Each SDK exports its own ReidentifySourceType that implements it:
// FlowReidentifySourceType for FlowVault and SkyReidentifySourceType for SkyVault, both with
// BASE64 and PRESIGNED_URL (re-identification takes no Skyflow ID). Only those types
// implement it.
type ReidentifyFileSource interface {
	// reidentifySource returns the wire value.
	reidentifySource() string
}

// FlowReidentifySourceType is a ReidentifyFileSource for a FlowVault vault.
type FlowReidentifySourceType string

const (
	FlowReidentifySourceTypeBase64       FlowReidentifySourceType = base64Value
	FlowReidentifySourceTypePresignedURL FlowReidentifySourceType = presignedURLValue
)

// SkyReidentifySourceType is a ReidentifyFileSource for a SkyVault vault.
type SkyReidentifySourceType string

const (
	SkyReidentifySourceTypeBase64       SkyReidentifySourceType = base64Value
	SkyReidentifySourceTypePresignedURL SkyReidentifySourceType = presignedURLValue
)

// VaultSources are a vault type's de-identify and re-identify sources. Responses resolve
// their OutputType against them. Pass FlowSources or SkySources to NewController.
type VaultSources struct {
	Deidentify []DeidentifyFileSource // must include BASE64, which local files are sent as
	Reidentify []ReidentifyFileSource
}

// FlowSources and SkySources are the FlowVault and SkyVault source sets.
var (
	FlowSources = VaultSources{
		Deidentify: []DeidentifyFileSource{FlowDeidentifySourceTypeBase64, FlowDeidentifySourceTypePresignedURL},
		Reidentify: []ReidentifyFileSource{FlowReidentifySourceTypeBase64, FlowReidentifySourceTypePresignedURL},
	}
	SkySources = VaultSources{
		Deidentify: []DeidentifyFileSource{SkyDeidentifySourceTypeBase64, SkyDeidentifySourceTypeSkyflowID, SkyDeidentifySourceTypePresignedURL},
		Reidentify: []ReidentifyFileSource{SkyReidentifySourceTypeBase64, SkyReidentifySourceTypePresignedURL},
	}
)

// FileDataFormat is a file format. The values are upper case, as in the Java SDK;
// the API's wire values are lower case, and the SDK converts in both directions (see
// wireFormat). A request's DataFormat is matched without regard to case.
type FileDataFormat string

const (
	FileDataFormatMP3   FileDataFormat = "MP3"
	FileDataFormatWAV   FileDataFormat = "WAV"
	FileDataFormatPDF   FileDataFormat = "PDF"
	FileDataFormatTXT   FileDataFormat = "TXT"
	FileDataFormatCSV   FileDataFormat = "CSV"
	FileDataFormatJSON  FileDataFormat = "JSON"
	FileDataFormatJPG   FileDataFormat = "JPG"
	FileDataFormatJPEG  FileDataFormat = "JPEG"
	FileDataFormatTIF   FileDataFormat = "TIF"
	FileDataFormatTIFF  FileDataFormat = "TIFF"
	FileDataFormatPNG   FileDataFormat = "PNG"
	FileDataFormatBMP   FileDataFormat = "BMP"
	FileDataFormatXLS   FileDataFormat = "XLS"
	FileDataFormatXLSX  FileDataFormat = "XLSX"
	FileDataFormatDOC   FileDataFormat = "DOC"
	FileDataFormatDOCX  FileDataFormat = "DOCX"
	FileDataFormatPPT   FileDataFormat = "PPT"
	FileDataFormatPPTX  FileDataFormat = "PPTX"
	FileDataFormatXML   FileDataFormat = "XML"
	FileDataFormatDCM   FileDataFormat = "DCM"
	FileDataFormatJSONL FileDataFormat = "JSONL"
	FileDataFormatZIP   FileDataFormat = "ZIP"
	FileDataFormatGIF   FileDataFormat = "GIF"
)

// DetectRunStatus is the state of a de-identify file run. SUCCESS, FAILED and UNKNOWN are terminal.
type DetectRunStatus string

const (
	DetectRunStatusUnknown    DetectRunStatus = "UNKNOWN"
	DetectRunStatusFailed     DetectRunStatus = "FAILED"
	DetectRunStatusSuccess    DetectRunStatus = "SUCCESS"
	DetectRunStatusInProgress DetectRunStatus = "IN_PROGRESS"
	DetectRunStatusQueued     DetectRunStatus = "QUEUED"
)

// ProcessedFileType is the kind of a run output.
type ProcessedFileType string

const (
	ProcessedFileTypeNone                                     ProcessedFileType = "NONE"
	ProcessedFileTypeRedactedFile                             ProcessedFileType = "REDACTED_FILE"
	ProcessedFileTypeRedactedText                             ProcessedFileType = "REDACTED_TEXT"
	ProcessedFileTypeEntities                                 ProcessedFileType = "ENTITIES"
	ProcessedFileTypeReidentifiedFile                         ProcessedFileType = "REIDENTIFIED_FILE"
	ProcessedFileTypeRedactedTranscription                    ProcessedFileType = "REDACTED_TRANSCRIPTION"
	ProcessedFileTypeRedactedDeepgramTranscriptionDiarizeJSON ProcessedFileType = "REDACTED_DEEPGRAM_TRANSCRIPTION_DIARIZE_JSON"
	ProcessedFileTypeObjectEntities                           ProcessedFileType = "OBJECT_ENTITIES"
)

// knownFileTypes holds every ProcessedFileType; add new values here too, or output file names
// leave out the type.
var knownFileTypes = func() map[ProcessedFileType]bool {
	types := map[ProcessedFileType]bool{}
	for _, t := range []ProcessedFileType{
		ProcessedFileTypeNone, ProcessedFileTypeRedactedFile, ProcessedFileTypeRedactedText, ProcessedFileTypeEntities,
		ProcessedFileTypeReidentifiedFile, ProcessedFileTypeRedactedTranscription,
		ProcessedFileTypeRedactedDeepgramTranscriptionDiarizeJSON, ProcessedFileTypeObjectEntities,
	} {
		types[t] = true
	}
	return types
}()

// ObjectEntityType is an object Detect can find in images and video.
type ObjectEntityType string

const (
	ObjectEntityTypeNone         ObjectEntityType = "NONE"
	ObjectEntityTypeFace         ObjectEntityType = "FACE"
	ObjectEntityTypeLicensePlate ObjectEntityType = "LICENSE_PLATE"
	ObjectEntityTypeLogo         ObjectEntityType = "LOGO"
	ObjectEntityTypeSignature    ObjectEntityType = "SIGNATURE"
	ObjectEntityTypeAll          ObjectEntityType = "ALL"
)

// ObjectDeidentificationType is how a detected object is treated.
type ObjectDeidentificationType string

const (
	ObjectDeidentificationTypeUnknown  ObjectDeidentificationType = "UNKNOWN"
	ObjectDeidentificationTypeRedact   ObjectDeidentificationType = "REDACT"
	ObjectDeidentificationTypeUnredact ObjectDeidentificationType = "UNREDACT"
)

// OutputTranscriptionType is the transcription an audio run returns.
type OutputTranscriptionType string

const (
	OutputTranscriptionTypeNone                         OutputTranscriptionType = "NONE"
	OutputTranscriptionTypeTranscription                OutputTranscriptionType = "TRANSCRIPTION"
	OutputTranscriptionTypeMedicalTranscription         OutputTranscriptionType = "MEDICAL_TRANSCRIPTION"
	OutputTranscriptionTypeDiarizedTranscription        OutputTranscriptionType = "DIARIZED_TRANSCRIPTION"
	OutputTranscriptionTypeMedicalDiarizedTranscription OutputTranscriptionType = "MEDICAL_DIARIZED_TRANSCRIPTION"
)

// PdfProcessingMode is how PDF text is read.
type PdfProcessingMode string

const (
	PdfProcessingModeNone      PdfProcessingMode = "NONE"
	PdfProcessingModeOCR       PdfProcessingMode = "OCR"
	PdfProcessingModeTextLayer PdfProcessingMode = "TEXT_LAYER"
)

// MaskingMethod is how detected regions of an image are hidden.
type MaskingMethod string

const (
	MaskingMethodNone     MaskingMethod = "NONE"
	MaskingMethodBlur     MaskingMethod = "BLUR"
	MaskingMethodBlackbox MaskingMethod = "BLACKBOX"
)

// GuardrailsValidation is the overall result of a guardrails check.
type GuardrailsValidation string

const (
	GuardrailsValidationFailed GuardrailsValidation = "FAILED"
	GuardrailsValidationPassed GuardrailsValidation = "PASSED"
	// GuardrailsValidationUnknown is set when the API returns no validation, or one this SDK
	// does not know.
	GuardrailsValidationUnknown GuardrailsValidation = "UNKNOWN"
)
