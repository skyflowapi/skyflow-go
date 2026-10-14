package common

// Base types embedded by flowvault's request/option/response structs.

// BaseInsertRequest holds the insert request fields shared across SDK versions.
type BaseInsertRequest struct {
	Table  string
	Values []map[string]interface{}
}

// BaseInsertOptions is a marker for options shared across SDK versions.
// CustomHeaders is intentionally omitted: v2 uses map[CustomHeaderKey]string,
// flowvault uses map[string]string — each version defines it as its own field.
type BaseInsertOptions struct{}

// BaseInsertResponse holds the insert response fields shared across SDK versions.
type BaseInsertResponse struct {
	Errors []map[string]interface{}
}
