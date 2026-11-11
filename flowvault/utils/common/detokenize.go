package common

type DetokenizeRequest struct {
	Tokens               []string // required
	TokenGroupRedactions []TokenGroupRedactions
}

type DetokenizeResponseRecordMetadata struct {
	SkyflowId string `json:"skyflowId"`
	TableName string `json:"tableName"`
}

type DetokenizeResponseRecord struct {
	Token          string                            `json:"token"`
	Value          interface{}                       `json:"value"`
	TokenGroupName string                            `json:"tokenGroupName"`
	Metadata       *DetokenizeResponseRecordMetadata `json:"metadata"`
	HttpCode       int                               `json:"httpCode"`
	Error          string                            `json:"error"`
	RequestId      string                            `json:"requestId"`
}

type DetokenizeResponse struct {
	Records []DetokenizeResponseRecord `json:"records"`
}

type DetokenizeOptions struct {
	Interceptor Interceptor
}
