package common

import (
	"os"

	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

type Env int

const (
	PROD Env = iota
	STAGE
	SANDBOX
	DEV
)

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}
type DeidentifyTextRequest struct {
	Text string
	// ConfigurationId    ConfigurationId
	Entities          []DetectEntities
	TokenFormat       TokenFormat
	AllowRegexList    []string
	RestrictRegexList []string
	Transformations   Transformations
}

type Transformations struct {
	ShiftDates DateTransformation
}

type TokenFormat struct {
	DefaultType         TokenTypeDefault
	VaultToken          []DetectEntities
	EntityUniqueCounter []DetectEntities
	EntityOnly          []DetectEntities
}

type TokenTypeDefault string

const (
	TokenTypeDefaultEntityOnly       TokenTypeDefault = "entity_only"
	TokenTypeDefaultEntityUnqCounter TokenTypeDefault = "entity_unq_counter"
	TokenTypeDefaultVaultToken       TokenTypeDefault = "vault_token"
)

type DateTransformation struct {
	MaxDays  int
	MinDays  int
	Entities []TransformationsShiftDatesEntityTypesItem
}

type TransformationsShiftDatesEntityTypesItem string

const (
	TransformationsShiftDatesEntityTypesItemDate         TransformationsShiftDatesEntityTypesItem = "date"
	TransformationsShiftDatesEntityTypesItemDateInterval TransformationsShiftDatesEntityTypesItem = "date_interval"
	TransformationsShiftDatesEntityTypesItemDob          TransformationsShiftDatesEntityTypesItem = "dob"
)

type DeidentifyFileStatus string

const (
	IN_PROGRESS DeidentifyFileStatus = "in_progress"
	FAILED      DeidentifyFileStatus = "failed"
	SUCCESS     DeidentifyFileStatus = "success"
)

type DetectEntities string

const (
	AccountNumber               DetectEntities = "account_number"
	Age                         DetectEntities = "age"
	All                         DetectEntities = "all"
	BankAccount                 DetectEntities = "bank_account"
	BloodType                   DetectEntities = "blood_type"
	Condition                   DetectEntities = "condition"
	CorporateAction             DetectEntities = "corporate_action"
	CreditCard                  DetectEntities = "credit_card"
	CreditCardExpiration        DetectEntities = "credit_card_expiration"
	Cvv                         DetectEntities = "cvv"
	Date                        DetectEntities = "date"
	Day                         DetectEntities = "day"
	DateInterval                DetectEntities = "date_interval"
	Dob                         DetectEntities = "dob"
	Dose                        DetectEntities = "dose"
	DriverLicense               DetectEntities = "driver_license"
	Drug                        DetectEntities = "drug"
	Duration                    DetectEntities = "duration"
	Effect                      DetectEntities = "effect"
	EmailAddress                DetectEntities = "email_address"
	Event                       DetectEntities = "event"
	Filename                    DetectEntities = "filename"
	FinancialMetric             DetectEntities = "financial_metric"
	Gender                      DetectEntities = "gender"
	HealthcareNumber            DetectEntities = "healthcare_number"
	Injury                      DetectEntities = "injury"
	IpAddress                   DetectEntities = "ip_address"
	Language                    DetectEntities = "language"
	Location                    DetectEntities = "location"
	LocationAddress             DetectEntities = "location_address"
	LocationAddressStreet       DetectEntities = "location_address_street"
	LocationCity                DetectEntities = "location_city"
	LocationCoordinate          DetectEntities = "location_coordinate"
	LocationCountry             DetectEntities = "location_country"
	LocationState               DetectEntities = "location_state"
	LocationZip                 DetectEntities = "location_zip"
	MaritalStatus               DetectEntities = "marital_status"
	MedicalCode                 DetectEntities = "medical_code"
	MedicalProcess              DetectEntities = "medical_process"
	Money                       DetectEntities = "money"
	Month                       DetectEntities = "month"
	Name                        DetectEntities = "name"
	NameFamily                  DetectEntities = "name_family"
	NameGiven                   DetectEntities = "name_given"
	NameMedicalProfessional     DetectEntities = "name_medical_professional"
	NumericalPii                DetectEntities = "numerical_pii"
	Occupation                  DetectEntities = "occupation"
	Organization                DetectEntities = "organization"
	OrganizationId              DetectEntities = "organization_id"
	OrganizationMedicalFacility DetectEntities = "organization_medical_facility"
	Origin                      DetectEntities = "origin"
	PassportNumber              DetectEntities = "passport_number"
	Password                    DetectEntities = "password"
	PhoneNumber                 DetectEntities = "phone_number"
	Project                     DetectEntities = "project"
	PhysicalAttribute           DetectEntities = "physical_attribute"
	PoliticalAffiliation        DetectEntities = "political_affiliation"
	Product                     DetectEntities = "product"
	Religion                    DetectEntities = "religion"
	RoutingNumber               DetectEntities = "routing_number"
	Sexuality                   DetectEntities = "sexuality"
	Ssn                         DetectEntities = "ssn"
	Statistics                  DetectEntities = "statistics"
	Time                        DetectEntities = "time"
	Trend                       DetectEntities = "trend"
	Url                         DetectEntities = "url"
	Username                    DetectEntities = "username"
	VehicleId                   DetectEntities = "vehicle_id"
	Year                        DetectEntities = "year"
	ZodiacSign                  DetectEntities = "zodiac_sign"
)

type MaskingMethod string

const (
	BLACKBOX MaskingMethod = "blackbox"
	BLUR     MaskingMethod = "blur"
)

type DetectOutputTranscriptions string

const (
	DIARIZED_TRANSCRIPTION         DetectOutputTranscriptions = "diarized_transcription"
	MEDICAL_DIARIZED_TRANSCRIPTION DetectOutputTranscriptions = "medical_diarized_transcription"
	MEDICAL_TRANSCRIPTION          DetectOutputTranscriptions = "medical_transcription"
	PLAINTEXT_TRANSCRIPTION        DetectOutputTranscriptions = "plaintext_transcription"
	TRANSCRIPTION                  DetectOutputTranscriptions = "transcription"
)

type BearerTokenOptions struct {
	Ctx      string
	RoleIDs  []string
	LogLevel logger.LogLevel
}

type SignedDataTokensOptions struct {
	DataTokens []string
	TimeToLive int
	Ctx        string
	LogLevel   logger.LogLevel
}

type SignedDataTokensResponse struct {
	Token       string
	SignedToken string
}

type VaultConfig struct {
	VaultId     string
	ClusterId   string
	BaseVaultURL string
	Env         Env
	Credentials Credentials
}

type Credentials struct {
	Path              string
	Roles             []string
	Context           string
	CredentialsString string
	Token             string
	ApiKey            string
}
type ConnectionConfig struct {
	ConnectionId  string
	ConnectionUrl string
	Credentials   Credentials
}
type DetectConfig struct {
	VaultId     string
	ClusterId   string
	Env         Env
	Credentials Credentials
}

type BYOT string

const (
	DISABLE       BYOT = "DISABLE"
	ENABLE        BYOT = "ENABLE"
	ENABLE_STRICT BYOT = "ENABLE_STRICT"
)

type InvokeConnectionResponse struct {
	Data     map[string]interface{}
	Metadata map[string]interface{}
	Errors   map[string]interface{}
}
type RequestMethod string

const (
	GET    RequestMethod = "GET"
	POST   RequestMethod = "POST"
	PUT    RequestMethod = "PUT"
	PATCH  RequestMethod = "PATCH"
	DELETE RequestMethod = "DELETE"
)

type DeidentifyTextResponse struct {
	ProcessedText  string
	Entities       []EntityInfo
	WordCount      int
	CharCount int
}

type EntityInfo struct {
	Token          string
	Value          string
	Entity         string
	Scores         map[string]float64
	ProcessedIndex TextIndex
	TextIndex      TextIndex
}
type TextIndex struct {
	Start int
	End   int
}
type ReidentifyTextResponse struct {
	ProcessedText string
}

type ReidentifyTextRequest struct {
	Text              string
	RedactedEntities  []DetectEntities
	MaskedEntities    []DetectEntities
	PlainTextEntities []DetectEntities
}

type DeidentifyFileRequest struct {
	Entities             []DetectEntities
	AllowRegexList       []string
	RestrictRegexList    []string
	TokenFormat          TokenFormat
	Transformations      Transformations
	OutputProcessedImage bool
	OutputOcrText        bool
	MaskingMethod        MaskingMethod
	PixelDensity         int
	MaxResolution        int
	OutputProcessedAudio bool
	OutputTranscription  DetectOutputTranscriptions
	Bleep                AudioBleep
	File                 FileInput
	OutputDirectory      string
	WaitTime             int
}

type FileInput struct {
	FilePath string
	File     *os.File
}

type AudioBleep struct {
	Gain         int
	Frequency    int
	StartPadding float64
	StopPadding  float64
}

type FileInfo struct {
	Name         string
	Size         int64
	Type         string
	LastModified int64
}

type FileEntityInfo struct {
	File      string
	Type      string
	Extension string
}

type DeidentifyFileResponse struct {
	File              FileInfo
	FileBase64        string
	Type              string
	Extension         string
	WordCount         int
	CharCount         int
	SizeInKb          float64
	DurationInSeconds float64
	PageCount         int
	SlideCount        int
	Entities          []FileEntityInfo
	RunId             string
	Status            string
}

type GetDetectRunRequest struct {
	RunId string
}

func (m RequestMethod) IsValid() bool {
	validMethods := []RequestMethod{
		GET,
		POST,
		PUT,
		PATCH,
		DELETE,
	}
	for _, method := range validMethods {
		if m == method {
			return true
		}
	}
	return false
}

type InvokeConnectionRequest struct {
	Method      RequestMethod
	QueryParams map[string]interface{}
	PathParams  map[string]string
	Body        map[string]interface{}
	Headers     map[string]string
}
type ContentType string

const (
	APPLICATIONORJSON ContentType = "application/json"
	TEXTORPLAIN       ContentType = "text/plain"
	FORMURLENCODED    ContentType = "application/x-www-form-urlencoded"
	FORMDATA          ContentType = "multipart/form-data"
	TEXTORXML         ContentType = "text/xml"
)

type OrderByEnum string

const (
	ASCENDING  OrderByEnum = "ASCENDING"
	DESCENDING OrderByEnum = "DESCENDING"
	NONE       OrderByEnum = "NONE"
)

type RedactionType string

// Constants for RedactionType
const (
	PLAIN_TEXT RedactionType = "PLAIN_TEXT"
	DEFAULT    RedactionType = "DEFAULT"
	MASKED     RedactionType = "MASKED"
	REDACTED   RedactionType = "REDACTED"
)

type InsertOptions struct {
	ReturnTokens    bool
	Upsert          string
	Homogeneous     bool
	TokenMode       BYOT
	ContinueOnError bool
	Tokens          []map[string]interface{}
}

type InsertRequest struct {
	Table  string
	Values []map[string]interface{}
}

type InsertResponse struct {
	// Response fields
	InsertedFields []map[string]interface{}
	Errors         []map[string]interface{}
}

type DetokenizeRequest struct {
	DetokenizeData []DetokenizeData
}
type DetokenizeData struct {
	Token         string
	RedactionType RedactionType
}
type DetokenizeOptions struct {
	ContinueOnError bool
	DownloadURL     bool
}

type DetokenizeRecordResponse struct {
	Token     string
	Value     string
	Type      string
	Error     string
	RequestId string
}

type DetokenizeResponse struct {
	DetokenizedFields []DetokenizeRecordResponse
	Errors            []DetokenizeRecordResponse
}

type DeleteRequest struct {
	Table string
	Ids   []string
}

type DeleteResponse struct {
	// Response fields
	DeletedIds []string
	Errors     []map[string]interface{}
}

type UpdateRequest struct {
	Table  string
	Data   map[string]interface{}
	Tokens map[string]interface{}
}
type UpdateOptions struct {
	ReturnTokens bool
	TokenMode    BYOT
}
type UpdateResponse struct {
	// Response fields
	UpdatedField map[string]interface{} 
	Errors       []map[string]interface{}
}

type GetRequest struct {
	Table string
	Ids   []string
}
type GetOptions struct {
	RedactionType RedactionType
	ReturnTokens  bool
	Fields        []string
	Offset        string
	Limit         string
	DownloadURL   bool
	ColumnName    string
	ColumnValues  []string
	OrderBy       OrderByEnum
}

type GetResponse struct {
	// Response fields
	Data   []map[string]interface{}
	Errors []map[string]interface{}
}

type QueryRequest struct {
	Query string
}
type TokenizeRequest struct {
	ColumnGroup string
	Value       string
}
type FileUploadRequest struct {
	Table  string
	SkyflowId  string
	ColumnName string
	FilePath   string
	Base64     string
	FileName   string
	FileObject os.File
}

type FileUploadResponse struct {
	SkyflowId string
	Errors  []map[string]interface{}
}

type TokenizeResponse struct {
	Tokens []string
	Errors []map[string]interface{}
}
type QueryResponse struct {
	Fields        []map[string]interface{}
	TokenizedData []map[string]interface{}
	Errors        []map[string]interface{}
}
