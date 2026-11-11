package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	api "github.com/skyflowapi/skyflow-go/flowvault/internal/generated"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/core"
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// result returns the response body and headers. A non-2xx response whose body still
// carries per-record results (hasRecords) is returned as a body, not an error, so callers
// see which records failed.
func result[T any](ctx context.Context, resp *core.Response[*T], err error, hasRecords func(*T) bool) (*T, http.Header, *fverrors.SkyflowError) {
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
		return nil, nil, transportError(ctx, err)
	}
	if hasRecords != nil && apiErr.Unwrap() != nil {
		body := new(T)
		if json.Unmarshal([]byte(apiErr.Unwrap().Error()), body) == nil && hasRecords(body) {
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

// transportError converts a call that got no usable HTTP response. If the caller's context
// ended, it's REQUEST_CANCELLED ("499"). Anything else (network failure, the client's
// Timeout, an empty or unreadable body) is TRANSPORT_ERROR: code "0" and no status, as in
// the Java SDK. The cause is kept, so errors.Is and errors.As work on the result.
func transportError(ctx context.Context, err error) *fverrors.SkyflowError {
	code := fverrors.TRANSPORT_ERROR
	if ctx.Err() != nil {
		code = fverrors.REQUEST_CANCELLED
	}
	return fverrors.NewSkyflowErrorWithCause(code, err.Error(), err)
}

// rawBodyMessage keeps the first MAX_RAW_ERROR_BODY bytes of a non-JSON error body,
// without splitting a multi-byte character.
func rawBodyMessage(body string) string {
	if len(body) <= fvconsts.MAX_RAW_ERROR_BODY {
		return body
	}
	cut := fvconsts.MAX_RAW_ERROR_BODY
	for cut > 0 && !utf8.RuneStart(body[cut]) {
		cut--
	}
	return body[:cut] + fvconsts.TRUNCATED_SUFFIX
}

// apiError converts a failed call. The status code and request ID are kept even when the
// body isn't a Skyflow error object. The message is then, as in the Java SDK, the first of:
// an "error" string, a top-level "message", the raw body (truncated by rawBodyMessage), or
// "Error with status code N" for an empty body.
func apiError(apiErr *core.APIError, header http.Header) *fverrors.SkyflowError {
	raw := ""
	if apiErr.Unwrap() != nil {
		raw = apiErr.Unwrap().Error()
	}
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(raw), &body)
	errorBody, ok := body[fvconsts.ERROR_KEY_ERROR].(map[string]interface{})
	if !ok {
		errorBody = map[string]interface{}{fvconsts.ERROR_KEY_MESSAGE: fallbackMessage(body, raw, apiErr.StatusCode)}
	}
	if _, ok := errorBody[fvconsts.ERROR_KEY_HTTP_CODE].(float64); !ok {
		errorBody[fvconsts.ERROR_KEY_HTTP_CODE] = apiErr.StatusCode
	}
	normalized, _ := json.Marshal(map[string]interface{}{fvconsts.ERROR_KEY_ERROR: errorBody})
	return fverrors.SkyflowErrorApi(fmt.Errorf("%d: %s", apiErr.StatusCode, normalized), header)
}

// fallbackMessage is the message for an error body without a Skyflow error object.
func fallbackMessage(body map[string]interface{}, raw string, statusCode int) string {
	for _, key := range []string{fvconsts.ERROR_KEY_ERROR, fvconsts.ERROR_KEY_MESSAGE} {
		if text, isText := body[key].(string); isText && strings.TrimSpace(text) != "" {
			return rawBodyMessage(text)
		}
	}
	if strings.TrimSpace(raw) != "" {
		return rawBodyMessage(raw)
	}
	return fmt.Sprintf(fvconsts.ERROR_WITH_STATUS_CODE, statusCode)
}

// uploadFilesCallError returns the error of a failed uploadFiles call. When the body is a
// {"records": [...]} list with record errors, the message is the distinct record errors
// joined, as in the Java SDK, and capped like a raw error body; otherwise it is err unchanged.
func uploadFilesCallError(callErr error, err *fverrors.SkyflowError) *fverrors.SkyflowError {
	var apiErr *core.APIError
	if !errors.As(callErr, &apiErr) || apiErr.Unwrap() == nil {
		return err
	}
	var body map[string][]*struct {
		Error string `json:"error"`
	}
	if json.Unmarshal([]byte(apiErr.Unwrap().Error()), &body) != nil {
		return err
	}
	seen := map[string]bool{}
	var messages []string
	for _, record := range body[fvconsts.ERROR_KEY_RECORDS] {
		if record != nil && record.Error != "" && !seen[record.Error] {
			seen[record.Error] = true
			messages = append(messages, record.Error)
		}
	}
	if len(messages) == 0 {
		return err
	}
	header := apiErr.Header
	if header == nil {
		header = http.Header{}
	}
	normalized, _ := json.Marshal(map[string]interface{}{fvconsts.ERROR_KEY_ERROR: map[string]interface{}{
		fvconsts.ERROR_KEY_HTTP_CODE: apiErr.StatusCode,
		fvconsts.ERROR_KEY_MESSAGE:   rawBodyMessage(strings.Join(messages, fvconsts.RECORD_ERRORS_SEPARATOR)),
	}})
	return fverrors.SkyflowErrorApi(fmt.Errorf("%d: %s", apiErr.StatusCode, normalized), header)
}

// failedCallStatus is the HTTP status of a call that failed but whose body still carried
// per-record results, or 0 when the call succeeded.
func failedCallStatus(callErr error) int {
	var apiErr *core.APIError
	if errors.As(callErr, &apiErr) {
		return apiErr.StatusCode
	}
	return 0
}

// recordHTTPCode is a record's httpCode in a successful call. When the vault leaves it out,
// it is 500 for a record with an error and 200 otherwise, as in the Java SDK.
func recordHTTPCode(httpCode int, recordErr string) int {
	if httpCode != 0 {
		return httpCode
	}
	if recordErr != "" {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// recordOutcome returns a record's httpCode and error. In a successful call (failedStatus 0)
// they are the record's own, with recordHTTPCode filling in a missing httpCode. In a failed call, as in the Java SDK, a record counts as a
// success only when it says so itself: no error and a 2xx status of its own. A record with
// neither gets UNKNOWN_RECORD_ERROR, and a missing httpCode becomes the call's status.
func recordOutcome(httpCode int, recordErr string, failedStatus int) (int, string) {
	if failedStatus == 0 {
		return recordHTTPCode(httpCode, recordErr), recordErr
	}
	if recordErr == "" && (httpCode < http.StatusOK || httpCode >= http.StatusMultipleChoices) {
		recordErr = fvconsts.UNKNOWN_RECORD_ERROR
	}
	if httpCode == 0 {
		httpCode = failedStatus
	}
	return httpCode, recordErr
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

func toWireColumnRedactions(redactions []fvcommon.ColumnRedactions) []*api.ColumnRedactions {
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
	if len(request.Records) > 0 {
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
	wire.Limit = request.Limit
	wire.Offset = request.Offset
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
			HashedData: record.HashedData,
			HttpCode:   recordHTTPCode(record.HTTPCode, recordErr),
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
			HttpCode:  recordHTTPCode(record.HTTPCode, recordErr),
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
			HttpCode:       recordHTTPCode(deref(record.HTTPCode), recordErr),
			Error:          recordErr,
			RequestId:      fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}

// toGetTokensResults converts getTokens records; failedStatus is failedCallStatus of the call.
func toGetTokensResults(records []*api.TokenizeResponseObject, header http.Header, failedStatus int) []fvcommon.GetTokensResponseRecord {
	results := make([]fvcommon.GetTokensResponseRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		httpCode, recordErr := recordOutcome(deref(record.HTTPCode), deref(record.Error), failedStatus)
		results = append(results, fvcommon.GetTokensResponseRecord{
			Value:          deref(record.Value),
			TokenGroupName: deref(record.TokenGroupName),
			Token:          record.Token,
			HttpCode:       httpCode,
			Error:          recordErr,
			RequestId:      fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}

// toDeleteFilesResults converts deleteFiles records; failedStatus is failedCallStatus of the call.
func toDeleteFilesResults(records []*api.FileDeleteResponseObject, header http.Header, failedStatus int) []fvcommon.DeleteFilesResponseRecord {
	results := make([]fvcommon.DeleteFilesResponseRecord, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		httpCode, recordErr := recordOutcome(record.HTTPCode, deref(record.Error), failedStatus)
		results = append(results, fvcommon.DeleteFilesResponseRecord{
			SkyflowId: deref(record.SkyflowID),
			TableName: record.TableName,
			Columns:   fvhelpers.ParseDeleteFilesColumns(record.Data),
			HttpCode:  httpCode,
			Error:     recordErr,
			RequestId: fvhelpers.RequestIdFor(header, recordErr),
		})
	}
	return results
}
