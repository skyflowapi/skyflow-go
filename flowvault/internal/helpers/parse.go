package helpers

import (
	"fmt"
	"net/http"
	"sort"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

const statusKey = "status"

// ParseTokens accepts a column value that is a token string, a token object or a list of either.
// A column whose value is null is left out, while one whose list
// is empty is kept with an empty (non-nil) list.
func ParseTokens(raw map[string]interface{}) map[string][]fvcommon.Token {
	if raw == nil {
		return nil
	}
	parsed := make(map[string][]fvcommon.Token, len(raw))
	for column, value := range raw {
		if value == nil {
			continue
		}
		tokens := []fvcommon.Token{}
		for _, item := range asList(value) {
			switch entry := item.(type) {
			case map[string]interface{}:
				tokens = append(tokens, fvcommon.Token{
					Token:          stringField(entry, "token"),
					TokenGroupName: stringField(entry, "tokenGroupName"),
					Path:           stringField(entry, "path"),
				})
			case nil:
			default:
				tokens = append(tokens, fvcommon.Token{Token: fmt.Sprint(entry)})
			}
		}
		parsed[column] = tokens
	}
	return parsed
}

// ParseDetokenizeMetadata accepts both "table" and "tableName" keys.
func ParseDetokenizeMetadata(raw map[string]interface{}) *fvcommon.DetokenizeResponseRecordMetadata {
	if raw == nil {
		return nil
	}
	return &fvcommon.DetokenizeResponseRecordMetadata{
		SkyflowId: stringField(raw, "skyflowID", "skyflowId"),
		TableName: stringField(raw, "tableName", "table"),
	}
}

func asList(value interface{}) []interface{} {
	if list, ok := value.([]interface{}); ok {
		return list
	}
	return []interface{}{value}
}

// stringField returns the first of keys present in entry as a string.
func stringField(entry map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := entry[key]; ok && value != nil {
			if text, isString := value.(string); isString {
				return text
			}
			return fmt.Sprint(value)
		}
	}
	return ""
}

// ParseDeleteFilesColumns converts {column: {"status": "DELETED"}} into results sorted by
// column. An unrecognized or missing status is DeleteFilesStatusUnknown, with the wire value
// kept in RawStatus.
func ParseDeleteFilesColumns(data map[string]interface{}) []fvcommon.DeleteFilesColumnResult {
	if data == nil {
		return nil
	}
	columns := make([]fvcommon.DeleteFilesColumnResult, 0, len(data))
	for name, value := range data {
		status := ""
		switch entry := value.(type) {
		case map[string]interface{}:
			status = stringField(entry, statusKey)
		case string:
			status = entry
		}
		columns = append(columns, fvcommon.DeleteFilesColumnResult{
			Column:    name,
			Status:    fvcommon.DeleteFilesStatusFromWire(status),
			RawStatus: status,
		})
	}
	sort.Slice(columns, func(i, j int) bool { return columns[i].Column < columns[j].Column })
	return columns
}

// RequestIdFor returns the call's request ID for a record that failed, and "" otherwise.
func RequestIdFor(header http.Header, recordErr string) string {
	if recordErr == "" || header == nil {
		return ""
	}
	return header.Get(fvconsts.RESPONSE_HEADER_REQUEST_ID)
}
