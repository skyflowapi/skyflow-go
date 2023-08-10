/*
Copyright (c) 2022 Skyflow, Inc.
*/
package common

// Internal
type ResponseBody map[string]interface{}

// Helper function that retrieves a Skyflow bearer token from your backend.
type TokenProvider func() (string, error)

// Supported request methods.
type RequestMethod int

const (
	GET RequestMethod = iota
	POST
	PUT
	PATCH
	DELETE
)

// Returns the string representation of the RequestMethod.
func (requestMethod RequestMethod) String() string {
	return [...]string{"GET", "POST", "PUT", "PATCH", "DELETE"}[requestMethod]
}

// Supported redaction types.
type RedactionType string

const (
	DEFAULT    RedactionType = "DEFAULT"
	PLAIN_TEXT RedactionType = "PLAIN_TEXT"
	MASKED     RedactionType = "MASKED"
	REDACTED   RedactionType = "REDACTED"
)

// Supported connection configurations.
type ConnectionConfig struct {
	ConnectionURL string
	MethodName    RequestMethod
	PathParams    map[string]string
	QueryParams   map[string]interface{}
	RequestBody   map[string]interface{}
	RequestHeader map[string]string
}

// Wrapper for parameters required by insert options.
type InsertOptions struct {
	Tokens bool
	Upsert []UpsertOptions
}

// Wrapper for parameters required by upsert options.
type UpsertOptions struct {
	Table  string
	Column string
}

// Contains the parameters required for Skyflow client initialisation.
type Configuration struct {
	VaultID       string
	VaultURL      string
	TokenProvider TokenProvider
}

// Internal
type InsertRecords struct {
	Records []InsertRecord
}

// Internal
type InsertRecord struct {
	Table  string
	Fields map[string]interface{}
}

// Internal
type DetokenizeInput struct {
	Records []RevealRecord
}

// Internal
type RevealRecord struct {
	Token     string
	Redaction string
}

// Internal
type DetokenizeRecords struct {
	Records []DetokenizeRecord
	Errors  []DetokenizeError
}

// Internal
type DetokenizeRecord struct {
	Token string
	Value string
}

// Internal
type DetokenizeError struct {
	Token string
	Error ResponseError
}

// Internal
type ResponseError struct {
	Code        string
	Description string
}

// Internal
type GetByIdInput struct {
	Records []SkyflowIdRecord
}

// Internal
type GetByIdRecords struct {
	Records []GetByIdRecord
	Errors  []GetByIdError
}

// Internal
type GetByIdRecord struct {
	Fields map[string]interface{}
	Table  string
}

// Internal
type GetByIdError struct {
	Ids   []string
	Error ResponseError
}

// Internal
type SkyflowIdRecord struct {
	Ids       []string
	Redaction RedactionType
	Table     string
}

// Supported content types.
type ContentType string

const (
	APPLICATIONORJSON ContentType = "application/json"
	TEXTORPLAIN       ContentType = "text/plain"
	FORMURLENCODED    ContentType = "application/x-www-form-urlencoded"
	FORMDATA          ContentType = "multipart/form-data"
	TEXTORXML         ContentType = "text/xml"
)

const sdk_name="skyflow-go"
const sdk_version="1.6.0"