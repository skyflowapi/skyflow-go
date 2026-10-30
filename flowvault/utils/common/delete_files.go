package common

// DeleteFilesRequestRecord takes exactly one of SkyflowId or UniqueValues.
type DeleteFilesRequestRecord struct {
	TableName    string   // required
	Columns      []string // required
	SkyflowId    string
	UniqueValues []map[string]interface{}
}

type DeleteFilesRequest struct {
	Records []DeleteFilesRequestRecord
}

type DeleteFilesColumnResult struct {
	Column string `json:"column"`
	Status string `json:"status"`
}

type DeleteFilesResponseRecord struct {
	SkyflowId string                    `json:"skyflowId"`
	TableName string                    `json:"tableName"`
	Columns   []DeleteFilesColumnResult `json:"columns"` // nil on a failed record
	HttpCode  int                       `json:"httpCode"`
	Error     string                    `json:"error"`
	RequestId string                    `json:"requestId"`
}

type DeleteFilesResponse struct {
	Records []DeleteFilesResponseRecord `json:"records"`
}

type DeleteFilesOptions struct {
	Interceptor Interceptor
}
