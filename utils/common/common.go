package common

import (
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
	Entities []string
	AllowRegexList []string
	RestrictRegexList []string
	TokenFormat string
	Transformations []string
}

type DeidentifyTextResponse struct {
	DeidentifiedText string   `json:"deidentifiedText"`
	Errors           []string `json:"errors,omitempty"`
}

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
	DEFAULT                  = "DEFAULT"
	MASKED                   = "MASKED"
	REDACTED                 = "REDACTED"
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
type DetokenizeResponse struct {
	DetokenizedFields []map[string]interface{}
	Errors            []map[string]interface{}
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
	Id     string
	Values map[string]interface{}
	Tokens map[string]interface{}
}
type UpdateOptions struct {
	ReturnTokens bool
	TokenMode    BYOT
}
type UpdateResponse struct {
	// Response fields
	SkyflowId string
	Tokens    map[string]interface{}
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

type UploadFileRequest struct {
	TableName  string
	SkyflowId  string
	ColumnName string
	FilePath   string
}

type QueryRequest struct {
	Query string
}
type TokenizeRequest struct {
	ColumnGroup string
	Value       string
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
