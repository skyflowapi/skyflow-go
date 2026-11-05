package common

import "github.com/skyflowapi/skyflow-go/common/vaultclient"

// UpdateType controls how an update or upsert applies data to an existing record.
type UpdateType string

const (
	// UpdateTypeUpdate merges the given columns into the record (server default).
	UpdateTypeUpdate UpdateType = "UPDATE"
	// UpdateTypeReplace overwrites the record with the given columns.
	UpdateTypeReplace UpdateType = "REPLACE"
)

// RequestContext is passed to an options Interceptor before each API call. It is the
// shared skyflow-go/common/vaultclient type, so detect and vault options use the same one.
type RequestContext = vaultclient.RequestContext

// Interceptor customises a single call, e.g. by adding headers.
type Interceptor = vaultclient.Interceptor

// Token is one token issued for a column value.
type Token struct {
	Token          string `json:"token"`
	TokenGroupName string `json:"tokenGroupName,omitempty"`
	Path           string `json:"path,omitempty"`
}

// HashedValue is one hash computed for a column value.
type HashedValue struct {
	Data     string `json:"data"`
	HashName string `json:"hashName,omitempty"`
}

type ColumnRedaction struct {
	ColumnName string
	Redaction  string
}

type TokenGroupRedaction struct {
	TokenGroupName string
	Redaction      string
}
