package common

type DeleteRequest struct {
	TableName    string // required
	SkyflowIds   []string
	UniqueValues []map[string]interface{}
}

type DeleteResponseRecord struct {
	SkyflowId string `json:"skyflowId"`
	HttpCode  int    `json:"httpCode"`
	Error     string `json:"error"`
	RequestId string `json:"requestId"`
}

type DeleteResponse struct {
	Records []DeleteResponseRecord `json:"records"`
}

type DeleteOptions struct {
	Interceptor Interceptor
}
