package common

// UpdateType controls how an update or upsert applies data to an existing record.
type UpdateType string

const (
	// UpdateTypeUpdate merges the given columns into the record (server default).
	UpdateTypeUpdate UpdateType = "UPDATE"
	// UpdateTypeReplace overwrites the record with the given columns.
	UpdateTypeReplace UpdateType = "REPLACE"
)

// Token is one token issued for a column value.
type Token struct {
	Token          string `json:"token"`
	TokenGroupName string `json:"tokenGroupName,omitempty"`
	Path           string `json:"path,omitempty"`
}

type ColumnRedactions struct {
	ColumnName string
	Redaction  string
}

type TokenGroupRedactions struct {
	TokenGroupName string
	Redaction      string
}
