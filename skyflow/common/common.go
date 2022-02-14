package common

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

type ConnectionConfig struct {
	ConnectionURL string
	MethodName    RequestMethod
	PathParams    map[string]string
	QueryParams   map[string]interface{}
	RequestBody   map[string]interface{}
	RequestHeader map[string]string
}

type InsertOptions struct {
	Tokens bool
}

type Configuration struct {
	VaultID       string
	VaultURL      string
	TokenProvider TokenProvider
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
