package common

type UpdateRequest struct {
	TableName  string // set here OR on every record
	Records    []UpdateRequestRecord
	UpdateType UpdateType // request level only
}

type UpdateRequestRecord struct {
	SkyflowId string                 // required
	Data      map[string]interface{} // required
	Tokens    map[string]interface{}
	TableName string // only if not set on the request
}

type UpdateResponseRecord = InsertResponseRecord

type UpdateResponse struct {
	Records []UpdateResponseRecord `json:"records"`
}

type UpdateOptions struct {
	Interceptor Interceptor
}
