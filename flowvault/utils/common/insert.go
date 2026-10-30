package common

type UpsertOptions struct {
	UniqueColumns []string // required
	UpdateType    UpdateType
}

type InsertRequest struct {
	TableName string // set here OR on every record
	Records   []InsertRequestRecord
	Upsert    *UpsertOptions // only with request-level TableName
}

type InsertRequestRecord struct {
	Data      map[string]interface{} // required
	Tokens    map[string]interface{}
	TableName string         // only if not set on the request
	Upsert    *UpsertOptions // only with record-level TableName
}

// InsertResponseRecord is the per-record result of Insert, Update and Get.
type InsertResponseRecord struct {
	TableName  string                   `json:"tableName"`
	SkyflowId  string                   `json:"skyflowId"`
	Tokens     map[string][]Token       `json:"tokens"`
	Data       map[string]interface{}   `json:"data"`
	HashedData map[string][]HashedValue `json:"hashedData"`
	HttpCode   int                      `json:"httpCode"`
	Error      string                   `json:"error"`
	RequestId  string                   `json:"requestId"`
}

type InsertResponse struct {
	Records []InsertResponseRecord `json:"records"`
}

type InsertOptions struct {
	Interceptor Interceptor
}
