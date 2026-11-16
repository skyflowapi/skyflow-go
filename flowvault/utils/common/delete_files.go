package common

import "strings"

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

// DeleteFilesStatus is the outcome of deleting the file in one column, as the vault reported it.
type DeleteFilesStatus string

const (
	// DeleteFilesStatusDeleted means the file was deleted.
	DeleteFilesStatusDeleted DeleteFilesStatus = "DELETED"
	// DeleteFilesStatusUnknown means the vault reported a status this SDK version does not
	// recognize, or none at all. The value as received is in DeleteFilesColumnResult.RawStatus.
	DeleteFilesStatusUnknown DeleteFilesStatus = "UNKNOWN"
)

// DeleteFilesStatusFromWire returns the status for a wire value, matched case-insensitively,
// or DeleteFilesStatusUnknown when there is none.
func DeleteFilesStatusFromWire(value string) DeleteFilesStatus {
	if strings.EqualFold(strings.TrimSpace(value), string(DeleteFilesStatusDeleted)) {
		return DeleteFilesStatusDeleted
	}
	return DeleteFilesStatusUnknown
}

type DeleteFilesColumnResult struct {
	Column    string            `json:"column"`
	Status    DeleteFilesStatus `json:"status"`
	RawStatus string            `json:"rawStatus"` // the status exactly as the vault sent it
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
	CustomHeaders map[CustomHeaderKey]string
}
