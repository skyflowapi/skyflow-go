package common

type GetRequest struct {
	TableName        string
	SkyflowIds       []string
	UniqueValues     []map[string]interface{}
	Columns          []string
	ColumnRedactions []ColumnRedactions
	Limit, Offset    *int               // nil = not sent; 0 is sent
	Records          []GetRequestRecord // multi-table mode; excludes the fields above
}

type GetRequestRecord struct {
	TableName        string
	SkyflowIds       []string
	UniqueValues     []map[string]interface{}
	Columns          []string
	ColumnRedactions []ColumnRedactions
}

type GetResponseRecord = InsertResponseRecord

type GetResponse struct {
	Records []GetResponseRecord `json:"records"`
}

type GetOptions struct {
	Interceptor Interceptor
}
