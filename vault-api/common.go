package vaultapi

type responseBody map[string]interface{}
type TokenProvider func() string

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

type LogLevel int

const (
	WARN LogLevel = iota
	INFO
	DEBUG
	ERROR
	NONE
)

func (logLevel LogLevel) String() string {
	return [...]string{"WARN", "INFO", "DEBUG", "ERROR", "NONE"}[logLevel]
}

type MessageType int

const (
	LOG MessageType = iota
)

func (messageType MessageType) String() string {
	return [...]string{"LOG", "WARN", "ERROR"}[messageType]
}

type RedactionType string

const (
	DEFAULT    RedactionType = "DEFAULT"
	PLAIN_TEXT RedactionType = "PLAIN_TEXT"
	MASKED     RedactionType = "MASKED"
	REDACTED   RedactionType = "REDACTED"
)

type ConnectionConfig struct {
	connectionURL string
	methodName    RequestMethod
	pathParams    map[string]interface{}
	queryParams   map[string]interface{}
	requestBody   map[string]interface{}
	requestHeader map[string]interface{}
}

type Options struct {
	LogLevel LogLevel
}
type InsertOptions struct {
	Tokens bool
}

type Configuration struct {
	VaultID       string
	VaultURL      string
	TokenProvider TokenProvider
	Options       Options
}

type InsertRecord struct {
	Records []SingleRecord
}

type SingleRecord struct {
	Table  string
	Fields map[string]interface{}
}

type DetokenizeInput struct {
	Records []RevealRecord
}

type RevealRecord struct {
	Token string
}

type GetByIdInput struct {
	Records []SkyflowIdRecord
}

type SkyflowIdRecord struct {
	Ids       []string
	Redaction RedactionType
	Table     string
}
