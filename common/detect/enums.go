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
	IncludePrefixIncludePrefixUnspecified IncludePrefix = "INCLUDE_PREFIX_UNSPECIFIED"
	IncludePrefixYes                      IncludePrefix = "YES"
	IncludePrefixNo                       IncludePrefix = "NO"
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

// DeidentifySourceType is what a DeidentifyFileRequest's Value holds.
type DeidentifySourceType string

const (
	DeidentifySourceTypeBase64       DeidentifySourceType = "BASE64"
	DeidentifySourceTypeSkyflowID    DeidentifySourceType = "SKYFLOW_ID"
	DeidentifySourceTypePresignedURL DeidentifySourceType = "PRESIGNED_URL"
)

// ReidentifySourceType is what a ReidentifyFileRequest's Value holds. Re-identification does
// not accept a Skyflow ID.
type ReidentifySourceType string

const (
	ReidentifySourceTypeBase64       ReidentifySourceType = "BASE64"
	ReidentifySourceTypePresignedURL ReidentifySourceType = "PRESIGNED_URL"
)

// DataSourceType is the form a file run's outputs take.
type DataSourceType string

const (
	DataSourceTypeBase64       DataSourceType = "BASE64"
	DataSourceTypeSkyflowID    DataSourceType = "SKYFLOW_ID"
	DataSourceTypePresignedURL DataSourceType = "PRESIGNED_URL"
)

// FileDataFormat is a file format. The wire values are lowercase.
type FileDataFormat string

const (
	FileDataFormatUnknown FileDataFormat = "unknown"
	FileDataFormatMP3     FileDataFormat = "mp3"
	FileDataFormatWAV     FileDataFormat = "wav"
	FileDataFormatPDF     FileDataFormat = "pdf"
	FileDataFormatTXT     FileDataFormat = "txt"
	FileDataFormatCSV     FileDataFormat = "csv"
	FileDataFormatJSON    FileDataFormat = "json"
	FileDataFormatJPG     FileDataFormat = "jpg"
	FileDataFormatJPEG    FileDataFormat = "jpeg"
	FileDataFormatTIF     FileDataFormat = "tif"
	FileDataFormatTIFF    FileDataFormat = "tiff"
	FileDataFormatPNG     FileDataFormat = "png"
	FileDataFormatBMP     FileDataFormat = "bmp"
	FileDataFormatXLS     FileDataFormat = "xls"
	FileDataFormatXLSX    FileDataFormat = "xlsx"
	FileDataFormatDOC     FileDataFormat = "doc"
	FileDataFormatDOCX    FileDataFormat = "docx"
	FileDataFormatPPT     FileDataFormat = "ppt"
	FileDataFormatPPTX    FileDataFormat = "pptx"
	FileDataFormatXML     FileDataFormat = "xml"
	FileDataFormatDCM     FileDataFormat = "dcm"
	FileDataFormatJSONL   FileDataFormat = "jsonl"
	FileDataFormatZIP     FileDataFormat = "zip"
	FileDataFormatGIF     FileDataFormat = "gif"
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
)
