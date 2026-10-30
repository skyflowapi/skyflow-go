package common

import "os"

// UploadStatus is the outcome of uploading one file column.
type UploadStatus string

const (
	UploadStatusUploaded UploadStatus = "UPLOADED"
	UploadStatusFailed   UploadStatus = "FAILED"
	UploadStatusSkipped  UploadStatus = "SKIPPED"
)

type UploadFilesRequest struct {
	Records []UploadFilesRequestRecord
}

type UploadFilesRequestRecord struct {
	TableName string // required
	SkyflowId string // empty -> new record; set -> update existing record
	Columns   []UploadFilesRequestColumn
}

// UploadFilesRequestColumn takes exactly one of FilePath, Base64 or FileObject.
type UploadFilesRequestColumn struct {
	Column     string // required
	FilePath   string
	Base64     string   // requires FileName
	FileObject *os.File // caller closes
	FileName   string   // derived from path/object if empty; server UUID if none
}

type UploadFilesColumnResult struct {
	Column       string       `json:"column"`
	FileName     string       `json:"fileName"`
	UploadStatus UploadStatus `json:"uploadStatus"`
	Error        string       `json:"error"`
}

type UploadFilesResponseRecord struct {
	SkyflowId string                    `json:"skyflowId"`
	TableName string                    `json:"tableName"`
	Columns   []UploadFilesColumnResult `json:"columns"`
	HttpCode  int                       `json:"httpCode"`
	Error     string                    `json:"error"`
	RequestId string                    `json:"requestId"`
}

type UploadFilesResponse struct {
	Records []UploadFilesResponseRecord `json:"records"`
}

type UploadFilesOptions struct {
	Interceptor Interceptor
}
