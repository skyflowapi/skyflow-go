package common

type QueryRequest struct {
	Query string
}

type QueryResponseMetadata struct {
	Columns []string `json:"columns"`
}

type QueryResponse struct {
	Fields    []map[string]interface{} `json:"fields"` // one map per row
	Metadata  QueryResponseMetadata    `json:"metadata"`
	RequestId string                   `json:"requestId"` // call-level, always populated
}

type QueryOptions struct {
	Interceptor Interceptor
}
