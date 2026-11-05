package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	api "github.com/skyflowapi/skyflow-go/flowvault/internal/generated"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/core"
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// result returns the response body and headers. A non-2xx response whose body still
// carries per-record results (in the shape fallback describes) is returned as a body, not
// an error, so callers see which records failed.
func result[T any](resp *core.Response[*T], err error, fallback *recordsFallback) (*T, http.Header, *fverrors.SkyflowError) {
	if err == nil {
		body, header := new(T), http.Header{}
		if resp != nil {
			if resp.Body != nil {
				body = resp.Body
			}
			if resp.Header != nil {
				header = resp.Header
			}
		}
		return body, header, nil
	}
	var apiErr *core.APIError
	if !errors.As(err, &apiErr) {
		return nil, nil, fverrors.NewSkyflowError(fverrors.SERVER, err.Error())
	}
	if fallback != nil && apiErr.Unwrap() != nil {
		body := new(T)
		if normalized, ok := fallback.normalize(apiErr.Unwrap().Error(), apiErr.StatusCode); ok && json.Unmarshal(normalized, body) == nil {
			header := apiErr.Header
			if header == nil {
				header = http.Header{}
			}
			return body, header, nil
		}
	}
	header := apiErr.Header
	if header == nil {
		header = http.Header{}
	}
	return nil, nil, apiError(apiErr, header)
}

// apiError converts a failed call. The status code and request ID are kept even when the
// body isn't a Skyflow error: such a body becomes the message.
func apiError(apiErr *core.APIError, header http.Header) *fverrors.SkyflowError {
	raw := ""
	if apiErr.Unwrap() != nil {
		raw = apiErr.Unwrap().Error()
	}
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(raw), &body)
	errorBody, ok := body[fvconsts.ERROR_KEY_ERROR].(map[string]interface{})
	if !ok {
		message := raw
		if text, isText := body[fvconsts.ERROR_KEY_ERROR].(string); isText {
			message = text
		} else if recordErrors := recordErrorMessages(body); recordErrors != "" {
			message = recordErrors
		}
		if strings.TrimSpace(message) == "" {
			message = http.StatusText(apiErr.StatusCode)
		}
		errorBody = map[string]interface{}{fvconsts.ERROR_KEY_MESSAGE: message}
	}
	if _, ok := errorBody[fvconsts.ERROR_KEY_HTTP_CODE].(float64); !ok {
		errorBody[fvconsts.ERROR_KEY_HTTP_CODE] = apiErr.StatusCode
	}
	normalized, _ := json.Marshal(map[string]interface{}{fvconsts.ERROR_KEY_ERROR: errorBody})
	return fverrors.SkyflowErrorApi(fmt.Errorf("%d: %s", apiErr.StatusCode, normalized), header)
}

// recordErrorMessages joins the distinct record errors of a records-shaped error body
// ({"records":[{"error":"..."}]}) with "; ", or returns "" when there are none.
func recordErrorMessages(body map[string]interface{}) string {
	records, _ := body[fvconsts.ERROR_KEY_RECORDS].([]interface{})
	var messages []string
	for _, record := range records {
		fields, _ := record.(map[string]interface{})
		if text, _ := fields[fvconsts.ERROR_KEY_ERROR].(string); text != "" && !slices.Contains(messages, text) {
			messages = append(messages, text)
		}
	}
	return strings.Join(messages, "; ")
}

// recordsFallback is the records-shaped error body of an operation: the vault can fail a
// call outright yet still list its records, for example when the request's only record has
// an invalid column. As in the Java SDK, each listed record becomes a failed record of the
// response:
//   - httpCode is the record's http_code, httpCode or statusCode (a number or numeric
//     string), else the call's HTTP status;
//   - error is the record's error, else its message, else "Unknown error";
//   - only the identifying fields in keep are carried over; data, tokens and values are not.
//
// The mappers then set the request ID on every record, since each one has an error.
type recordsFallback struct {
	key  string   // the records array: "records", or "response" for Detokenize
	keep []string // wire fields carried over from each record
}

var (
	recordFallback      = &recordsFallback{key: fvconsts.ERROR_KEY_RECORDS, keep: []string{wireSkyflowID, wireTableName}}
	deleteFallback      = &recordsFallback{key: fvconsts.ERROR_KEY_RECORDS, keep: []string{wireSkyflowID}}
	detokenizeFallback  = &recordsFallback{key: wireDetokenizeRecords, keep: []string{wireToken, wireTokenGroupName}}
	getTokensFallback   = &recordsFallback{key: fvconsts.ERROR_KEY_RECORDS, keep: []string{wireValue, wireTokenGroupName, wireToken}}
	deleteFilesFallback = &recordsFallback{key: fvconsts.ERROR_KEY_RECORDS, keep: []string{wireSkyflowID, wireTableName}}
)

// Wire field names used by the fallback.
const (
	wireDetokenizeRecords = "response"
	wireSkyflowID         = "skyflowID"
	wireTableName         = "tableName"
	wireToken             = "token"
	wireTokenGroupName    = "tokenGroupName"
	wireValue             = "value"
	wireHttpCode          = "httpCode"
	wireMessage           = "message"
	unknownRecordError    = "Unknown error"
)

// fallbackHttpCodeKeys are the spellings a record's status is read from, in order.
var fallbackHttpCodeKeys = []string{fvconsts.ERROR_KEY_HTTP_CODE, wireHttpCode, "statusCode"}

// normalize rewrites raw into the success body shape, or reports false when raw lists no
// records, so the call fails with an error instead.
func (f *recordsFallback) normalize(raw string, status int) ([]byte, bool) {
	var body map[string]interface{}
	if json.Unmarshal([]byte(raw), &body) != nil {
		return nil, false
	}
	listed, _ := body[f.key].([]interface{})
	records := make([]interface{}, 0, len(listed))
	for _, item := range listed {
		fields, isObject := item.(map[string]interface{})
		if !isObject {
			continue
		}
		record := map[string]interface{}{}
		for _, key := range f.keep {
			if value, present := fields[key]; present {
				record[key] = value
			}
		}
		record[wireHttpCode] = fallbackHttpCode(fields, status)
		record[fvconsts.ERROR_KEY_ERROR] = fallbackError(fields)
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil, false
	}
	normalized, err := json.Marshal(map[string]interface{}{f.key: records})
	return normalized, err == nil
}

func fallbackHttpCode(fields map[string]interface{}, status int) int {
	for _, key := range fallbackHttpCodeKeys {
		switch value := fields[key].(type) {
		case float64:
			return int(value)
		case string:
			if code, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				return code
			}
		}
	}
	return status
}

func fallbackError(fields map[string]interface{}) string {
	for _, key := range []string{fvconsts.ERROR_KEY_ERROR, wireMessage} {
		switch value := fields[key].(type) {
		case nil:
		case string:
			if value != "" {
				return value
			}
		default:
			if text, err := json.Marshal(value); err == nil {
				return string(text)
			}
		}
	}
	return unknownRecordError
}

// recordHttpCode is a record's status on a 2xx response. The vault can leave it out, which
// reads as 0; as in the Java SDK it then follows the record's outcome: 500 for a record with
// an error, 200 for one without.
func recordHttpCode(httpCode int, recordErr string) int {
	switch {
	case httpCode != 0:
		return httpCode
	case recordErr != "":
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func deref[T any](value *T) T {
	var zero T
	if value == nil {
		return zero
	}
	return *value
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func optionalInt(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func toWireUniqueValues(values []map[string]interface{}) []*api.UniqueValue {
	if values == nil {
		return nil
	}
	wire := make([]*api.UniqueValue, 0, len(values))
	for _, value := range values {
		wire = append(wire, &api.UniqueValue{Data: value})
	}
	return wire
}

func toWireColumnRedactions(redactions []fvcommon.ColumnRedaction) []*api.ColumnRedactions {
	if redactions == nil {
		return nil
	}
	wire := make([]*api.ColumnRedactions, 0, len(redactions))
	for _, redaction := range redactions {
		wire = append(wire, &api.ColumnRedactions{ColumnName: redaction.ColumnName, Redaction: redaction.Redaction})
	}
	return wire
}

func toWireUpsert(upsert *fvcommon.UpsertOptions) *api.Upsert {
	if upsert == nil {
		return nil
	}
	wire := &api.Upsert{UniqueColumns: upsert.UniqueColumns}
	if upsert.UpdateType != "" {
		wire.UpdateType = api.UpsertUpdateType(upsert.UpdateType).Ptr()
	}
	return wire
}

func toWireInsert(vaultID string, request fvcommon.InsertRequest) *api.InsertRequest {
	wire := &api.InsertRequest{
		VaultID:   vaultID,
		TableName: request.TableName,
		Upsert:    toWireUpsert(request.Upsert),
		Records:   make([]*api.InsertRecordData, 0, len(request.Records)),
	}
	for _, record := range request.Records {
		wire.Records = append(wire.Records, &api.InsertRecordData{
			Data:      record.Data,
			Tokens:    record.Tokens,
			TableName: optionalString(record.TableName),
			Upsert:    toWireUpsert(record.Upsert),
		})
	}
	return wire
}

func toWireUpdate(vaultID string, request fvcommon.UpdateRequest) *api.UpdateRequest {
	wire := &api.UpdateRequest{
		VaultID:   vaultID,
		TableName: request.TableName,
		Records:   make([]*api.UpdateRecordData, 0, len(request.Records)),
	}
	if request.UpdateType != "" {
		wire.UpdateType = api.UpdateRequestUpdateType(request.UpdateType).Ptr()
	}
	for _, record := range request.Records {
		wire.Records = append(wire.Records, &api.UpdateRecordData{
			SkyflowID: record.SkyflowId,
			Data:      record.Data,
			Tokens:    record.Tokens,
			TableName: optionalString(record.TableName),
		})
	}
	return wire
}

func toWireGet(vaultID string, request fvcommon.GetRequest) *api.GetRequest {
	wire := &api.GetRequest{VaultID: vaultID}
	if request.Records != nil {
		wire.Records = make([]*api.GetRequestData, 0, len(request.Records))
		for _, record := range request.Records {
			wire.Records = append(wire.Records, &api.GetRequestData{
				TableName:        record.TableName,
				SkyflowIDs:       record.SkyflowIds,
				Columns:          record.Columns,
				ColumnRedactions: toWireColumnRedactions(record.ColumnRedactions),
				UniqueValues:     toWireUniqueValues(record.UniqueValues),
			})
		}
		return wire
	}
	wire.TableName = optionalString(request.TableName)
	wire.SkyflowIDs = request.SkyflowIds
	wire.UniqueValues = toWireUniqueValues(request.UniqueValues)
	wire.Columns = request.Columns
	wire.ColumnRedactions = toWireColumnRedactions(request.ColumnRedactions)
	wire.Limit = optionalInt(request.Limit)
	wire.Offset = optionalInt(request.Offset)
	return wire
}

func toWireDelete(vaultID string, request fvcommon.DeleteRequest) *api.DeleteRequest {
	return &api.DeleteRequest{
		VaultID:      vaultID,
		TableName:    optionalString(request.TableName),
		SkyflowIDs:   request.SkyflowIds,
		UniqueValues: toWireUniqueValues(request.UniqueValues),
	}
}

func toWireDetokenize(vaultID string, request fvcommon.DetokenizeRequest) *api.DetokenizeRequest {
	wire := &api.DetokenizeRequest{VaultID: vaultID, Tokens: request.Tokens}
	for _, redaction := range request.TokenGroupRedactions {
		wire.TokenGroupRedactions = append(wire.TokenGroupRedactions, &api.TokenGroupRedactions{
			TokenGroupName: optionalString(redaction.TokenGroupName),
			Redaction:      optionalString(redaction.Redaction),
		})
	}
	return wire
}

func toWireGetTokens(vaultID string, request fvcommon.GetTokensRequest) *api.GetTokensFromValuesRequest {
	wire := &api.GetTokensFromValuesRequest{
		VaultID: vaultID,
		Records: make([]*api.GetTokensFromValuesRequestObject, 0, len(request.Records)),
	}
	for _, record := range request.Records {
		wire.Records = append(wire.Records, &api.GetTokensFromValuesRequestObject{Value: record.Value, TokenGroupName: record.TokenGroupName})
	}
	return wire
}

func toWireUploadFiles(vaultID string, request fvcommon.UploadFilesRequest, files [][]fvhelpers.ResolvedFile) *api.FileUploadRequest {
	wire := &api.FileUploadRequest{VaultID: vaultID, Records: make([]*api.FileUploadRecord, 0, len(request.Records))}
	for i, record := range request.Records {
		columns := make([]*api.FileUploadColumn, 0, len(record.Columns))
		for j, column := range record.Columns {
			columns = append(columns, &api.FileUploadColumn{Column: column.Column, FileName: optionalString(files[i][j].Name)})
		}
		wire.Records = append(wire.Records, &api.FileUploadRecord{
			TableName: record.TableName,
			SkyflowID: optionalString(record.SkyflowId),
			Columns:   columns,
		})
	}
	return wire
}

func toWireDeleteFiles(vaultID string, request fvcommon.DeleteFilesRequest) *api.FileDeleteRequest {
	wire := &api.FileDeleteRequest{VaultID: vaultID, Records: make([]*api.FileDeleteRecord, 0, len(request.Records))}
	for _, record := range request.Records {
		wire.Records = append(wire.Records, &api.FileDeleteRecord{
			TableName:    record.TableName,
			SkyflowID:    optionalString(record.SkyflowId),
			Columns:      record.Columns,
			UniqueValues: toWireUniqueValues(record.UniqueValues),
		})
	}
	return wire
}

// toRecordResults converts insert, update and get records.
func toRecordResults(records []*api.RecordResponseObject, header http.Header) []fvcommon.InsertResponseRecord {
	results := make([]fvcommon.InsertResponseRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		recordErr := deref(record.Error)
		results = append(results, fvcommon.InsertResponseRecord{
			TableName:  deref(record.TableName),
			SkyflowId:  deref(record.SkyflowID),
			Tokens:     fvhelpers.ParseTokens(record.Tokens),
			Data:       record.Data,
			HashedData: fvhelpers.ParseHashedData(record.HashedData),
			HttpCode:   recordHttpCode(record.HTTPCode, recordErr),
			Error:      recordErr,
			RequestId:  fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}

func toDeleteResults(records []*api.DeleteResponseObject, header http.Header) []fvcommon.DeleteResponseRecord {
	results := make([]fvcommon.DeleteResponseRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		recordErr := deref(record.Error)
		results = append(results, fvcommon.DeleteResponseRecord{
			SkyflowId: record.SkyflowID,
			HttpCode:  recordHttpCode(record.HTTPCode, recordErr),
			Error:     recordErr,
			RequestId: fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}

func toDetokenizeResults(records []*api.DetokenizeResponseObject, header http.Header) []fvcommon.DetokenizeResponseRecord {
	results := make([]fvcommon.DetokenizeResponseRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		recordErr := deref(record.Error)
		results = append(results, fvcommon.DetokenizeResponseRecord{
			Token:          record.Token,
			Value:          deref(record.Value),
			TokenGroupName: deref(record.TokenGroupName),
			Metadata:       fvhelpers.ParseDetokenizeMetadata(record.Metadata),
			HttpCode:       recordHttpCode(deref(record.HTTPCode), recordErr),
			Error:          recordErr,
			RequestId:      fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}

func toGetTokensResults(records []*api.TokenizeResponseObject, header http.Header) []fvcommon.GetTokensResponseRecord {
	results := make([]fvcommon.GetTokensResponseRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		recordErr := deref(record.Error)
		results = append(results, fvcommon.GetTokensResponseRecord{
			Value:          deref(record.Value),
			TokenGroupName: deref(record.TokenGroupName),
			Token:          record.Token,
			HttpCode:       recordHttpCode(deref(record.HTTPCode), recordErr),
			Error:          recordErr,
			RequestId:      fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}

func toDeleteFilesResults(records []*api.FileDeleteResponseObject, header http.Header) []fvcommon.DeleteFilesResponseRecord {
	results := make([]fvcommon.DeleteFilesResponseRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		recordErr := deref(record.Error)
		results = append(results, fvcommon.DeleteFilesResponseRecord{
			SkyflowId: deref(record.SkyflowID),
			TableName: record.TableName,
			Columns:   fvhelpers.ParseDeleteFilesColumns(record.Data),
			HttpCode:  recordHttpCode(record.HTTPCode, recordErr),
			Error:     recordErr,
			RequestId: fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}
