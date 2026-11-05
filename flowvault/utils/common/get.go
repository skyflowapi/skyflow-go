package common

type GetRequest struct {
	TableName        string
	SkyflowIds       []string
	UniqueValues     []map[string]interface{}
	Columns          []string
	ColumnRedactions []ColumnRedaction
	Limit, Offset    int                // 0 = not sent
	Records          []GetRequestRecord // multi-table mode; excludes the fields above
}

type GetRequestRecord struct {
	TableName        string
	SkyflowIds       []string
	UniqueValues     []map[string]interface{}
	Columns          []string
	ColumnRedactions []ColumnRedaction
}

type GetResponseRecord = InsertResponseRecord

type GetResponse struct {
	Records []GetResponseRecord `json:"records"`
}

type GetOptions struct {
	Interceptor Interceptor
}
