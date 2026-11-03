package common

import (
	cfg "github.com/skyflowapi/skyflow-go/common/config"
	sa "github.com/skyflowapi/skyflow-go/common/serviceaccount"
)

// ---------------------------------------------------------------------------
// Platform-level types shared with the other Skyflow Go SDKs (same types, via
// aliases): environment, credentials, configs, custom headers, service account.
// ---------------------------------------------------------------------------

type Env = cfg.Env

const (
	PROD    Env = cfg.PROD
	STAGE   Env = cfg.STAGE
	SANDBOX Env = cfg.SANDBOX
	DEV     Env = cfg.DEV
)

type (
	Credentials      = cfg.Credentials
	VaultConfig      = cfg.VaultConfig
	ConnectionConfig = cfg.ConnectionConfig
	CustomHeaderKey  = cfg.CustomHeaderKey
)

const (
	SkyflowAccountId   CustomHeaderKey = cfg.SkyflowAccountId
	SkyflowAccountName CustomHeaderKey = cfg.SkyflowAccountName
	RequestIdHeader    CustomHeaderKey = cfg.RequestIdHeader
)

type (
	TokenResponse            = sa.TokenResponse
	BearerTokenOptions       = sa.BearerTokenOptions
	SignedDataTokensOptions  = sa.SignedDataTokensOptions
	SignedDataTokensResponse = sa.SignedDataTokensResponse
)

// ---------------------------------------------------------------------------
// FlowVault vault API types. These belong to flowvault alone and evolve with
// the FlowVault API, independently of the v2 (SkyVault) SDK.
// ---------------------------------------------------------------------------

type RedactionType string

const (
	PLAIN_TEXT RedactionType = "PLAIN_TEXT"
	DEFAULT    RedactionType = "DEFAULT"
	MASKED     RedactionType = "MASKED"
	REDACTED   RedactionType = "REDACTED"
)

type OrderByEnum string

const (
	ASCENDING  OrderByEnum = "ASCENDING"
	DESCENDING OrderByEnum = "DESCENDING"
	NONE       OrderByEnum = "NONE"
)

type BYOT string

const (
	DISABLE       BYOT = "DISABLE"
	ENABLE        BYOT = "ENABLE"
	ENABLE_STRICT BYOT = "ENABLE_STRICT"
)

// Get

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
	DownloadUrl   bool
	ColumnName    string
	ColumnValues  []string
	OrderBy       OrderByEnum
	CustomHeaders map[CustomHeaderKey]string
}

type GetResponse struct {
	Data   []map[string]interface{}
	Errors []map[string]interface{}
}

// Detokenize

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
	CustomHeaders   map[CustomHeaderKey]string
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

// Delete

type DeleteRequest struct {
	Table string
	Ids   []string
}

type DeleteOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

type DeleteResponse struct {
	DeletedIds []string
	Errors     []map[string]interface{}
}

// Update

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
	UpdatedField map[string]interface{}
	Errors       []map[string]interface{}
}

// Query

type QueryRequest struct {
	Query string
}

type QueryOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

type QueryResponse struct {
	Fields []map[string]interface{}
	Errors []map[string]interface{}
}

// Tokenize

type TokenizeRequest struct {
	ColumnGroup string
	Value       string
}

type TokenizeOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}

type TokenizeResponse struct {
	Tokens []string
	Errors []map[string]interface{}
}

// IsCredentialsEmpty reports whether no credential source is set.
func IsCredentialsEmpty(c Credentials) bool { return cfg.IsCredentialsEmpty(c) }
