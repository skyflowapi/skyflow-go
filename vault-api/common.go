package vaultapi

type token string
type responseBody map[string]interface{}
type TokenProvider interface {
	getBearerToken() (token, error)
}

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
	PLAIN_TEXT               = "PLAIN_TEXT"
	MASKED                   = "MASKED"
	REDACTED                 = "REDACTED"
)

//

type ConnectionConfig struct {
	connectionURL string
	methodName    RequestMethod
	pathParams    map[string]interface{}
	queryParams   map[string]interface{}
	requestBody   map[string]interface{}
	requestHeader map[string]interface{}
}

type Configuration struct {
	vaultID       string
	vaultURL      string
	tokenProvider TokenProvider
	options       map[string]interface{}
}
