package common

type QueryRequest struct {
	Query string
}

type QueryResponseMetadata struct {
	Columns []string `json:"columns"`
}

type QueryResponse struct {
	// Records holds the rows the query returned, one map of column name to value per row.
	// Never nil. The vault returns at most 25 rows per call, so page with LIMIT and OFFSET in
	// the query.
	Records   []map[string]interface{} `json:"records"`
	Metadata  QueryResponseMetadata    `json:"metadata"`
	RequestId string                   `json:"requestId"` // call-level, always populated
}

type QueryOptions struct {
	CustomHeaders map[CustomHeaderKey]string
}
