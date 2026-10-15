package common

// InsertRequest is the FlowVault insert request. Unlike v2, ContinueOnError
// is part of the request rather than the options.
type InsertRequest struct {
	Table           string
	Values          []map[string]interface{}
	ContinueOnError bool
}

// InsertOptions are the FlowVault insert options. CustomHeaders uses plain
// string keys (v2 uses a typed CustomHeaderKey).
type InsertOptions struct {
	Upsert        string
	CustomHeaders map[string]string
}

// InsertResponse is the FlowVault insert response.
type InsertResponse struct {
	InsertedFields []map[string]interface{}
	Tokens         []map[string]interface{}
	Errors         []map[string]interface{}
}
