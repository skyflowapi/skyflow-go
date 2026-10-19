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
func ParseTokens(raw map[string]interface{}) map[string][]fvcommon.Token {
	if raw == nil {
		return nil
	}
	parsed := make(map[string][]fvcommon.Token, len(raw))
	for column, value := range raw {
		var tokens []fvcommon.Token
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
		if tokens != nil {
			parsed[column] = tokens
		}
	}
	return parsed
}

// ParseHashedData accepts a column value that is a hash string, a {data, hashName} object or a list of either.
func ParseHashedData(raw map[string]interface{}) map[string][]fvcommon.HashedValue {
	if raw == nil {
		return nil
	}
	parsed := make(map[string][]fvcommon.HashedValue, len(raw))
	for column, value := range raw {
		var hashes []fvcommon.HashedValue
		for _, item := range asList(value) {
			switch entry := item.(type) {
			case map[string]interface{}:
				hashes = append(hashes, fvcommon.HashedValue{Data: stringField(entry, "data"), HashName: stringField(entry, "hashName")})
			case nil:
			default:
				hashes = append(hashes, fvcommon.HashedValue{Data: fmt.Sprint(entry)})
			}
		}
		if hashes != nil {
			parsed[column] = hashes
		}
	}
	return parsed
}

// ParseDetokenizeMetadata accepts both "table" and "tableName" keys.
func ParseDetokenizeMetadata(raw map[string]interface{}) *fvcommon.DetokenizeRecordMetadata {
	if raw == nil {
		return nil
	}
	return &fvcommon.DetokenizeRecordMetadata{
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

// ParseDeleteFilesColumns converts {column: {"status": "DELETED"}} into results sorted by column.
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
		columns = append(columns, fvcommon.DeleteFilesColumnResult{Column: name, Status: status})
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
