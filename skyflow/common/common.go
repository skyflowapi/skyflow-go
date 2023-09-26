/*
Copyright (c) 2022 Skyflow, Inc.
*/
package common

import "context"

type ResponseBody map[string]interface{}
type TokenProvider func() (string, error)

type RequestMethod int

const (
	GET RequestMethod = iota
	POST
	PUT
	PATCH
	DELETE
)

func (requestMethod RequestMethod) String() string {
	return [...]string{"GET", "POST", "PUT", "PATCH", "DELETE"}[requestMethod]
}

type RedactionType string

const (
	DEFAULT    RedactionType = "DEFAULT"
	PLAIN_TEXT RedactionType = "PLAIN_TEXT"
	MASKED     RedactionType = "MASKED"
	REDACTED   RedactionType = "REDACTED"
)

type BYOT string

const (
	DISABLE       BYOT = "DISABLE"
	ENABLE        BYOT = "ENABLE"
	ENABLE_STRICT BYOT = "ENABLE_STRICT"
)

type ConnectionConfig struct {
	ConnectionURL string
	MethodName    RequestMethod
	PathParams    map[string]string
	QueryParams   map[string]interface{}
	RequestBody   map[string]interface{}
	RequestHeader map[string]string
}

type InsertOptions struct {
	Tokens          bool
	Upsert          []UpsertOptions
	Context         context.Context
	ContinueOnError bool
	Byot            BYOT
}

type DetokenizeOptions struct {
	Context         context.Context
	ContinueOnError bool
}

type GetByIdOptions struct {
	Context context.Context
}

type UpsertOptions struct {
	Table  string
	Column string
}

type Configuration struct {
	VaultID       string
	VaultURL      string
	TokenProvider TokenProvider
}

type InsertRecords struct {
	Records []InsertRecord
	Errors  []InsertError
}
type InsertError struct {
	Error InsertResponseError
}
type InsertRecord struct {
	RequestIndex int `json:"request_index"`
	Table        string
	Fields       map[string]interface{}
	Tokens       map[string]interface{}
}

type DetokenizeInput struct {
	Records []RevealRecord
}

type RevealRecord struct {
	Token     string
	Redaction string
}

type DetokenizeRecords struct {
	Records []DetokenizeRecord
	Errors  []DetokenizeError
}

type DetokenizeRecord struct {
	Token string
	Value string
}

type DetokenizeError struct {
	Token string
	Error ResponseError
}

type InsertResponseError struct {
	RequestIndex int `json:"request_index"`
	Code         string
	Description  string
}
type ResponseError struct {
	Code        string
	Description string
}
type GetByIdInput struct {
	Records []SkyflowIdRecord
}

type GetByIdRecords struct {
	Records []GetByIdRecord
	Errors  []GetByIdError
}

type GetByIdRecord struct {
	Fields map[string]interface{}
	Table  string
}

type GetByIdError struct {
	Ids   []string
	Error ResponseError
}
type SkyflowIdRecord struct {
	Ids       []string
	Redaction RedactionType
	Table     string
}

type ContentType string

const (
	APPLICATIONORJSON ContentType = "application/json"
	TEXTORPLAIN       ContentType = "text/plain"
	FORMURLENCODED    ContentType = "application/x-www-form-urlencoded"
	FORMDATA          ContentType = "multipart/form-data"
	TEXTORXML         ContentType = "text/xml"
)

const sdk_name = "skyflow-go"
const sdk_version = "1.8.1"
