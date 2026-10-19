package common

type GetTokensRequest struct {
	Records []GetTokensRequestRecord
}

type GetTokensRequestRecord struct {
	Value          interface{} // required
	TokenGroupName string      // required, deterministic group
}

type GetTokensResponseRecord struct {
	Value          interface{} `json:"value"`
	TokenGroupName string      `json:"tokenGroupName"`
	Token          string      `json:"token"`
	HttpCode       int         `json:"httpCode"`
	Error          string      `json:"error"`
	RequestId      string      `json:"requestId"`
}

type GetTokensResponse struct {
	Records []GetTokensResponseRecord `json:"records"`
}

type GetTokensOptions struct {
	Interceptor Interceptor
}
