package common

// UpdateType controls how an update or upsert applies data to an existing record.
type UpdateType string

const (
	// UpdateTypeUpdate merges the given columns into the record (server default).
	UpdateTypeUpdate UpdateType = "UPDATE"
	// UpdateTypeReplace overwrites the record with the given columns.
	UpdateTypeReplace UpdateType = "REPLACE"
)

// NotBatched is RequestContext's BatchIndex and TotalBatches for an operation that was not
// split into batches, as in the Java and Python SDKs. Every flowvault Go operation sends a
// single request, so it is the only value today.
const NotBatched = -1

// RequestContext is passed to an options Interceptor before each API call.
type RequestContext struct {
	Operation string
	// BatchIndex is the zero-based position of this batch within the caller's request, and
	// TotalBatches the number of batches; both are NotBatched (-1) for an unbatched call.
	BatchIndex   int
	TotalBatches int
	headers      map[CustomHeaderKey]string
}

// AddHeader adds a custom header to the outgoing request.
func (rc *RequestContext) AddHeader(key CustomHeaderKey, value string) {
	if rc.headers == nil {
		rc.headers = make(map[CustomHeaderKey]string)
	}
	rc.headers[key] = value
}

// Headers returns a copy of the headers added so far.
func (rc *RequestContext) Headers() map[CustomHeaderKey]string {
	headers := make(map[CustomHeaderKey]string, len(rc.headers))
	for key, value := range rc.headers {
		headers[key] = value
	}
	return headers
}

// Interceptor customises a single call, e.g. by adding headers.
type Interceptor func(rc *RequestContext)

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
