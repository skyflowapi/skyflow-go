package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	api "github.com/skyflowapi/skyflow-go/flowvault/internal/generated"
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvvalidation "github.com/skyflowapi/skyflow-go/flowvault/internal/validation"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/flowvault/utils/messages"
)

// VaultController is flowvault's vault controller.
// Embeds BaseController for shared auth/HTTP logic.
//
// Every operation validates its request, prepares auth, headers and the HTTP client, then
// calls the FlowService /v2 API through the generated client in flowvault/internal/generated.
type VaultController struct {
	BaseController
}

func NewVaultController(config fvcommon.VaultConfig, client ClientSettings) *VaultController {
	return &VaultController{BaseController: NewBaseController(config, client)}
}

// prepare logs the call, validates the request and resolves auth, headers and HTTP settings.
func (v *VaultController) prepare(ctx context.Context, method, operation string, headers map[fvcommon.CustomHeaderKey]string, validate func() *fverrors.SkyflowError) (*RequestSetup, *fverrors.SkyflowError) {
	fvlogger.Info(fmt.Sprintf(logs.METHOD_TRIGGERED, method))
	fvlogger.Info(fmt.Sprintf(logs.VALIDATING_REQUEST, method))
	if err := validate(); err != nil {
		return nil, err
	}
	return v.SetupRequest(ctx, operation, headers)
}

// finish logs how the call ended and passes its error through.
func finish(method string, err *fverrors.SkyflowError) *fverrors.SkyflowError {
	if err != nil {
		fvlogger.Error(fmt.Sprintf(logs.REQUEST_REJECTED, method))
		return err
	}
	fvlogger.Info(fmt.Sprintf(logs.REQUEST_RESOLVED, method))
	return nil
}

func hasRecords[T any](records func(*T) int) func(*T) bool {
	return func(body *T) bool { return records(body) > 0 }
}

// Insert inserts records, optionally upserting on unique columns.
func (v *VaultController) Insert(ctx context.Context, request fvcommon.InsertRequest, options fvcommon.InsertOptions) (*fvcommon.InsertResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_INSERT, fvconsts.OPERATION_INSERT, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateInsert(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.InsertRecords(ctx, toWireInsert(setup.VaultID, request))
	body, header, err := result(ctx, resp, callErr, hasRecords(func(b *api.InsertResponse) int { return len(b.Records) }), fvconsts.ERROR_KEY_RECORDS)
	if err := finish(fvconsts.METHOD_INSERT, err); err != nil {
		return nil, err
	}
	if status := failedCallStatus(callErr); status != 0 {
		return &fvcommon.InsertResponse{Records: failedRecordResults(failedRecordMaps(callErr, fvconsts.ERROR_KEY_RECORDS), status, header)}, nil
	}
	return &fvcommon.InsertResponse{Records: toRecordResults(body.Records, header)}, nil
}

// Update updates existing records by Skyflow ID.
func (v *VaultController) Update(ctx context.Context, request fvcommon.UpdateRequest, options fvcommon.UpdateOptions) (*fvcommon.UpdateResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_UPDATE, fvconsts.OPERATION_UPDATE, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateUpdate(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.UpdateRecords(ctx, toWireUpdate(setup.VaultID, request))
	body, header, err := result(ctx, resp, callErr, hasRecords(func(b *api.UpdateResponse) int { return len(b.Records) }), fvconsts.ERROR_KEY_RECORDS)
	if err := finish(fvconsts.METHOD_UPDATE, err); err != nil {
		return nil, err
	}
	if status := failedCallStatus(callErr); status != 0 {
		return &fvcommon.UpdateResponse{Records: failedRecordResults(failedRecordMaps(callErr, fvconsts.ERROR_KEY_RECORDS), status, header)}, nil
	}
	return &fvcommon.UpdateResponse{Records: toRecordResults(body.Records, header)}, nil
}

// Get fetches records by Skyflow IDs or unique values, from one table or several.
func (v *VaultController) Get(ctx context.Context, request fvcommon.GetRequest, options fvcommon.GetOptions) (*fvcommon.GetResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_GET, fvconsts.OPERATION_GET, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateGet(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.GetRecords(ctx, toWireGet(setup.VaultID, request))
	body, header, err := result(ctx, resp, callErr, hasRecords(func(b *api.GetResponse) int { return len(b.Records) }), fvconsts.ERROR_KEY_RECORDS)
	if err := finish(fvconsts.METHOD_GET, err); err != nil {
		return nil, err
	}
	if status := failedCallStatus(callErr); status != 0 {
		return &fvcommon.GetResponse{Records: failedRecordResults(failedRecordMaps(callErr, fvconsts.ERROR_KEY_RECORDS), status, header)}, nil
	}
	return &fvcommon.GetResponse{Records: toRecordResults(body.Records, header)}, nil
}

// Delete deletes records by Skyflow IDs or unique values.
func (v *VaultController) Delete(ctx context.Context, request fvcommon.DeleteRequest, options fvcommon.DeleteOptions) (*fvcommon.DeleteResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_DELETE, fvconsts.OPERATION_DELETE, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateDelete(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.DeleteRecords(ctx, toWireDelete(setup.VaultID, request))
	body, header, err := result(ctx, resp, callErr, hasRecords(func(b *api.DeleteResponse) int { return len(b.Records) }), fvconsts.ERROR_KEY_RECORDS)
	if err := finish(fvconsts.METHOD_DELETE, err); err != nil {
		return nil, err
	}
	if status := failedCallStatus(callErr); status != 0 {
		return &fvcommon.DeleteResponse{Records: failedDeleteResults(failedRecordMaps(callErr, fvconsts.ERROR_KEY_RECORDS), status, header)}, nil
	}
	return &fvcommon.DeleteResponse{Records: toDeleteResults(body.Records, header)}, nil
}

// Detokenize returns the values behind tokens, applying per-token-group redaction.
func (v *VaultController) Detokenize(ctx context.Context, request fvcommon.DetokenizeRequest, options fvcommon.DetokenizeOptions) (*fvcommon.DetokenizeResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_DETOKENIZE, fvconsts.OPERATION_DETOKENIZE, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateDetokenize(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Tokens.WithRawResponse.Detokenize(ctx, toWireDetokenize(setup.VaultID, request))
	body, header, err := result(ctx, resp, callErr, hasRecords(func(b *api.DetokenizeResponse) int { return len(b.Response) }), fvconsts.ERROR_KEY_RESPONSE)
	if err := finish(fvconsts.METHOD_DETOKENIZE, err); err != nil {
		return nil, err
	}
	if status := failedCallStatus(callErr); status != 0 {
		return &fvcommon.DetokenizeResponse{Records: failedDetokenizeResults(failedRecordMaps(callErr, fvconsts.ERROR_KEY_RESPONSE), status, header)}, nil
	}
	return &fvcommon.DetokenizeResponse{Records: toDetokenizeResults(body.Response, header)}, nil
}

// GetTokens returns existing tokens for values in deterministic token groups.
func (v *VaultController) GetTokens(ctx context.Context, request fvcommon.GetTokensRequest, options fvcommon.GetTokensOptions) (*fvcommon.GetTokensResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_GET_TOKENS, fvconsts.OPERATION_GET_TOKENS, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateGetTokens(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Tokens.WithRawResponse.GetTokens(ctx, toWireGetTokens(setup.VaultID, request))
	body, header, err := result(ctx, resp, callErr, hasRecords(func(b *api.GetTokensFromValuesResponse) int { return len(b.Records) }), "")
	if err := finish(fvconsts.METHOD_GET_TOKENS, err); err != nil {
		return nil, err
	}
	return &fvcommon.GetTokensResponse{Records: toGetTokensResults(body.Records, header, failedCallStatus(callErr), failedRecordMaps(callErr, fvconsts.ERROR_KEY_RECORDS))}, nil
}

// Query runs a SELECT query. It returns at most 25 rows per call, so page with LIMIT and
// OFFSET in the query; a warning is logged when a call returns that many.
func (v *VaultController) Query(ctx context.Context, request fvcommon.QueryRequest, options fvcommon.QueryOptions) (*fvcommon.QueryResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_QUERY, fvconsts.OPERATION_QUERY, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateQuery(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Query.WithRawResponse.ExecuteQuery(ctx, &api.ExecuteQueryRequest{VaultID: setup.VaultID, Query: request.Query})
	body, header, err := result(ctx, resp, callErr, nil, "")
	if err := finish(fvconsts.METHOD_QUERY, err); err != nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, 0, len(body.Records))
	for _, record := range body.Records {
		if record != nil {
			rows = append(rows, record.Data)
		}
	}
	if len(rows) >= fvconsts.QUERY_RECORD_LIMIT {
		fvlogger.Warn(fmt.Sprintf(logs.QUERY_RECORD_LIMIT_REACHED, fvconsts.QUERY_RECORD_LIMIT))
	}
	response := &fvcommon.QueryResponse{Records: rows, RequestId: header.Get(fvconsts.RESPONSE_HEADER_REQUEST_ID)}
	if body.Metadata != nil {
		response.Metadata.Columns = body.Metadata.Columns
	}
	return response, nil
}

// UploadFiles uploads files into file columns. It requests a signed upload URL per column,
// then PUTs each file to its URL, at most MAX_CONCURRENT_UPLOADS at once. The PUTs use
// vaultclient.SignedURLClient: no retries, no overall timeout and no SDK headers, but the
// vault's ConnectTimeout, ReadTimeout and WriteTimeout. A column that can't be uploaded is
// reported as FAILED or SKIPPED; the other columns are unaffected.
func (v *VaultController) UploadFiles(ctx context.Context, request fvcommon.UploadFilesRequest, options fvcommon.UploadFilesOptions) (*fvcommon.UploadFilesResponse, *fverrors.SkyflowError) {
	files := make([][]fvhelpers.ResolvedFile, len(request.Records))
	setup, err := v.prepare(ctx, fvconsts.METHOD_UPLOAD_FILES, fvconsts.OPERATION_UPLOAD_FILES, options.CustomHeaders, func() *fverrors.SkyflowError {
		if err := fvvalidation.ValidateUploadFiles(request); err != nil {
			return err
		}
		// Check every file up front so a bad file fails the call before anything is uploaded.
		// Nothing is read into memory here: files are read, and base64 decoded, only while
		// they are uploaded.
		for i, record := range request.Records {
			for _, column := range record.Columns {
				file, err := fvhelpers.ResolveFile(column)
				if err != nil {
					return err
				}
				files[i] = append(files[i], file)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// The vault creates a record for every entry without a SkyflowId, so the call is not
	// retried on a 5xx, which could create duplicates; 408 and 429 are.
	resp, callErr := setup.API.Files.WithRawResponse.UploadFiles(vaultclient.WithoutServerErrorRetry(ctx), toWireUploadFiles(setup.VaultID, request, files))
	// Unlike getTokens and deleteFiles, an error status fails the whole call even when the
	// body carries records, so no file is uploaded for a failed call.
	body, header, err := result(ctx, resp, callErr, nil, "")
	if err != nil {
		err = uploadFilesCallError(callErr, err)
	}
	if err := finish(fvconsts.METHOD_UPLOAD_FILES, err); err != nil {
		return nil, err
	}

	return &fvcommon.UploadFilesResponse{Records: uploadToSignedURLs(ctx, vaultclient.SignedURLClient(setup.HTTPClient), request.Records, files, body.Records, header)}, nil
}

// uploadToSignedURLs PUTs each file to the signed URL the vault returned for its column and
// builds one result per requested record:
//   - Returned records are matched to requested ones by position. If the counts differ, no
//     position can be trusted, so nothing is uploaded rather than risk sending a file to
//     another record. A returned record whose table (or skyflowId, when one was requested)
//     doesn't match its request is not uploaded either.
//   - A record the vault failed has every column SKIPPED with the record's error.
//   - A column with no signed URL is SKIPPED with NO_SIGNED_URL_ERROR.
//   - When any column of a record wasn't uploaded, the record's error says how many, and its
//     httpCode stays the one the vault returned.
func uploadToSignedURLs(ctx context.Context, client *http.Client, requested []fvcommon.UploadFilesRequestRecord, files [][]fvhelpers.ResolvedFile,
	returned []*api.FileUploadResponseObject, header http.Header) []fvcommon.UploadFilesResponseRecord {
	requestID := header.Get(fvconsts.RESPONSE_HEADER_REQUEST_ID)
	records := make([]fvcommon.UploadFilesResponseRecord, len(requested))
	if len(returned) != len(requested) {
		fvlogger.Warn(logs.INCOMPLETE_UPLOAD_FILES_RESPONSE)
		for i, record := range requested {
			records[i] = skippedUploadRecord(record, files[i], record.SkyflowId, record.TableName,
				http.StatusInternalServerError, fvconsts.UPLOAD_RESPONSE_MISMATCH_ERROR, requestID)
		}
		return records
	}

	var uploads []func()
	for i, wireRecord := range returned {
		record := requested[i]
		recordErr, httpCode := "", 0
		skyflowID, tableName := record.SkyflowId, record.TableName
		if wireRecord != nil {
			recordErr, httpCode = deref(wireRecord.Error), wireRecord.HTTPCode
			skyflowID, tableName = wireRecord.SkyflowID, wireRecord.TableName
		}
		if recordErr == "" && !uploadResponseMatches(record, wireRecord) {
			fvlogger.Warn(fmt.Sprintf(logs.UPLOAD_FILES_RECORD_MISMATCH, i))
			recordErr = fvconsts.UPLOAD_RESPONSE_MISMATCH_ERROR
		}
		if httpCode == 0 {
			httpCode = http.StatusOK
			if recordErr != "" {
				httpCode = http.StatusInternalServerError
			}
		}
		if recordErr != "" {
			records[i] = skippedUploadRecord(record, files[i], skyflowID, tableName, httpCode, recordErr, requestID)
			continue
		}
		records[i] = fvcommon.UploadFilesResponseRecord{
			SkyflowId: skyflowID,
			TableName: tableName,
			HttpCode:  httpCode,
			Columns:   make([]fvcommon.UploadFilesColumnResult, len(record.Columns)),
		}
		for j, column := range record.Columns {
			file := files[i][j]
			columnResult := &records[i].Columns[j]
			*columnResult = fvcommon.UploadFilesColumnResult{Column: column.Column, FileName: file.Name, UploadStatus: fvcommon.UploadStatusSkipped}
			signedURL, _ := wireRecord.Data[column.Column].(string)
			if signedURL == "" {
				columnResult.Error = fvconsts.NO_SIGNED_URL_ERROR
				continue
			}
			uploads = append(uploads, func() {
				columnResult.UploadStatus, columnResult.Error = fvhelpers.PutFile(ctx, client, signedURL, file)
			})
		}
	}
	fvhelpers.RunConcurrently(uploads, fvconsts.MAX_CONCURRENT_UPLOADS)

	// The vault accepted these records, but a file that never reached storage is still a
	// failure the caller must see on the record, not only on its column.
	for i := range records {
		if records[i].Error != "" {
			continue
		}
		notUploaded := 0
		for _, column := range records[i].Columns {
			if column.UploadStatus != fvcommon.UploadStatusUploaded {
				notUploaded++
			}
		}
		if notUploaded > 0 {
			records[i].Error = fmt.Sprintf(fvconsts.UPLOAD_COLUMNS_FAILED_ERROR, notUploaded, len(records[i].Columns))
			records[i].RequestId = requestID
		}
	}
	return records
}

// uploadResponseMatches reports whether a returned record belongs to its requested record:
// same table, and the same skyflowId when the request named one.
func uploadResponseMatches(requested fvcommon.UploadFilesRequestRecord, returned *api.FileUploadResponseObject) bool {
	if returned == nil || returned.TableName != requested.TableName {
		return false
	}
	return requested.SkyflowId == "" || returned.SkyflowID == requested.SkyflowId
}

// skippedUploadRecord is the result for a record none of whose files was uploaded.
func skippedUploadRecord(record fvcommon.UploadFilesRequestRecord, files []fvhelpers.ResolvedFile, skyflowID, tableName string,
	httpCode int, recordErr, requestID string) fvcommon.UploadFilesResponseRecord {
	columns := make([]fvcommon.UploadFilesColumnResult, len(record.Columns))
	for j, column := range record.Columns {
		columns[j] = fvcommon.UploadFilesColumnResult{Column: column.Column, FileName: files[j].Name, UploadStatus: fvcommon.UploadStatusSkipped, Error: recordErr}
	}
	return fvcommon.UploadFilesResponseRecord{
		SkyflowId: skyflowID,
		TableName: tableName,
		Columns:   columns,
		HttpCode:  httpCode,
		Error:     recordErr,
		RequestId: requestID,
	}
}

// DeleteFiles deletes files from file columns of existing records, including the stored objects.
func (v *VaultController) DeleteFiles(ctx context.Context, request fvcommon.DeleteFilesRequest, options fvcommon.DeleteFilesOptions) (*fvcommon.DeleteFilesResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(ctx, fvconsts.METHOD_DELETE_FILES, fvconsts.OPERATION_DELETE_FILES, options.CustomHeaders, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateDeleteFiles(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Files.WithRawResponse.DeleteFiles(ctx, toWireDeleteFiles(setup.VaultID, request))
	body, header, err := result(ctx, resp, callErr, hasRecords(func(b *api.FileDeleteResponse) int { return len(b.Records) }), "")
	if err := finish(fvconsts.METHOD_DELETE_FILES, err); err != nil {
		return nil, err
	}
	return &fvcommon.DeleteFilesResponse{Records: toDeleteFilesResults(body.Records, header, failedCallStatus(callErr), failedRecordMaps(callErr, fvconsts.ERROR_KEY_RECORDS))}, nil
}
