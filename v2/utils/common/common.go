package common

import (
	"os"

	cfg "github.com/skyflowapi/skyflow-go/common/config"
	sa "github.com/skyflowapi/skyflow-go/common/serviceaccount"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
)

// Platform-level configuration types are defined once in
// skyflow-go/common/config and shared by every SDK in this repository.
// Aliases keep them the same types under v2's import path.
type Env = cfg.Env

const (
	PROD    Env = cfg.PROD
	STAGE   Env = cfg.STAGE
	SANDBOX Env = cfg.SANDBOX
	DEV     Env = cfg.DEV
)

// Service-account types are defined once in skyflow-go/common/serviceaccount
// and shared by every SDK in this repository. Aliases keep them the same
// types under v2's import path.
type TokenResponse = sa.TokenResponse
type BearerTokenOptions = sa.BearerTokenOptions

type SignedDataTokensOptions = sa.SignedDataTokensOptions

type SignedDataTokensResponse = sa.SignedDataTokensResponse

type VaultConfig = cfg.VaultConfig

type Credentials = cfg.Credentials
type ConnectionConfig = cfg.ConnectionConfig

type BYOT string

const (
	DISABLE       BYOT = "DISABLE"
	ENABLE        BYOT = "ENABLE"
	ENABLE_STRICT BYOT = "ENABLE_STRICT"
)

type InvokeConnectionResponse struct {
	Data     interface{}
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
	Body        interface{}
	Headers     map[string]string
}
type ContentType string

const (
	APPLICATIONORJSON ContentType = "application/json"
	TEXTORPLAIN       ContentType = "text/plain"
	FORMURLENCODED    ContentType = "application/x-www-form-urlencoded"
	FORMDATA          ContentType = "multipart/form-data"
	TEXTORXML         ContentType = "text/xml"
	APPLICATIONXML    ContentType = "application/xml"
	TEXTHTML          ContentType = "text/html"
)

type OrderByEnum string

const (
	ASCENDING  OrderByEnum = "ASCENDING"
	DESCENDING OrderByEnum = "DESCENDING"
	NONE       OrderByEnum = "NONE"
)

// RedactionType is shared with the Detect APIs (RedactionLevel.RedactionType), so it is an
// alias of the Detect type declared in detect.go.

// Constants for RedactionType
const (
	PLAIN_TEXT RedactionType = "PLAIN_TEXT"
	DEFAULT    RedactionType = "DEFAULT"
	MASKED     RedactionType = "MASKED"
	REDACTED   RedactionType = "REDACTED"
)

type CustomHeaderKey = cfg.CustomHeaderKey

const (
	SkyflowAccountId   CustomHeaderKey = cfg.SkyflowAccountId
	SkyflowAccountName CustomHeaderKey = cfg.SkyflowAccountName
	RequestIdHeader    CustomHeaderKey = cfg.RequestIdHeader

	// Deprecated: Use SkyflowAccountId instead.
	SkyflowAccountID = SkyflowAccountId
	// Deprecated: Use RequestIdHeader instead.
	RequestIDHeader = RequestIdHeader
)

// RequestContext is passed to a Detect options Interceptor before each API call.
type RequestContext = vaultclient.RequestContext

// Interceptor customises a single Detect call, e.g. by adding headers.
type Interceptor = vaultclient.Interceptor

type InsertOptions struct {
	ReturnTokens    bool
	Upsert          string
	Homogeneous     bool
	TokenMode       BYOT
	ContinueOnError bool
	Tokens          []map[string]interface{}
	CustomHeaders   map[CustomHeaderKey]string
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
	DownloadUrl     bool
	// Deprecated: Use DownloadUrl instead.
	DownloadURL   bool
	CustomHeaders map[CustomHeaderKey]string
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
	//revive:disable-next-line:var-naming
	Ids []string
}
type DeleteOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

type DeleteResponse struct {
	// Response fields
	//revive:disable-next-line:var-naming
	DeletedIds []string
	Errors     []map[string]interface{}
}

type UpdateRequest struct {
	Table  string
	Data   map[string]interface{}
	Tokens map[string]interface{}
}
type UpdateOptions struct {
	ReturnTokens  bool
	TokenMode     BYOT
	CustomHeaders map[CustomHeaderKey]string
}
type UpdateResponse struct {
	// Response fields
	UpdatedField map[string]interface{}
	Errors       []map[string]interface{}
}

type GetRequest struct {
	Table string
	//revive:disable-next-line:var-naming
	Ids []string
}
type GetOptions struct {
	RedactionType RedactionType
	ReturnTokens  bool
	Fields        []string
	Offset        string
	Limit         string
	DownloadUrl   bool
	// Deprecated: Use DownloadUrl instead.
	DownloadURL   bool
	ColumnName    string
	ColumnValues  []string
	OrderBy       OrderByEnum
	CustomHeaders map[CustomHeaderKey]string
}

type GetResponse struct {
	// Response fields
	Data   []map[string]interface{}
	Errors []map[string]interface{}
}

type QueryRequest struct {
	Query string
}
type QueryOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}
type TokenizeOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}
type FileUploadOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}
type TokenizeRequest struct {
	ColumnGroup string
	Value       string
}
type FileUploadRequest struct {
	Table      string
	SkyflowId  string
	ColumnName string
	FilePath   string
	Base64     string
	FileName   string
	FileObject os.File
}

type FileUploadResponse struct {
	SkyflowId string
	Errors    []map[string]interface{}
}

type TokenizeResponse struct {
	Tokens []string
	Errors []map[string]interface{}
}
type QueryResponse struct {
	Fields []map[string]interface{}
	Errors []map[string]interface{}
}
