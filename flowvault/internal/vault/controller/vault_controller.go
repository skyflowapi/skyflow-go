package controller

import (
	"context"
	"fmt"

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
// Embeds BaseController for shared auth/HTTP logic — same as Java's
// VaultController extends AbstractBaseVaultController.
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
func (v *VaultController) prepare(method, operation string, interceptor fvcommon.Interceptor, validate func() *fverrors.SkyflowError) (*RequestSetup, *fverrors.SkyflowError) {
	fvlogger.Info(fmt.Sprintf(logs.METHOD_TRIGGERED, method))
	fvlogger.Info(fmt.Sprintf(logs.VALIDATING_REQUEST, method))
	if err := validate(); err != nil {
		return nil, err
	}
	return v.SetupRequest(operation, interceptor)
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
	setup, err := v.prepare(fvconsts.METHOD_INSERT, fvconsts.OPERATION_INSERT, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateInsert(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.InsertRecords(ctx, toWireInsert(setup.VaultID, request))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.InsertResponse) int { return len(b.Records) }))
	if err := finish(fvconsts.METHOD_INSERT, err); err != nil {
		return nil, err
	}
	return &fvcommon.InsertResponse{Records: toRecordResults(body.Records, header)}, nil
}

// Update updates existing records by Skyflow ID.
func (v *VaultController) Update(ctx context.Context, request fvcommon.UpdateRequest, options fvcommon.UpdateOptions) (*fvcommon.UpdateResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(fvconsts.METHOD_UPDATE, fvconsts.OPERATION_UPDATE, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateUpdate(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.UpdateRecords(ctx, toWireUpdate(setup.VaultID, request))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.UpdateResponse) int { return len(b.Records) }))
	if err := finish(fvconsts.METHOD_UPDATE, err); err != nil {
		return nil, err
	}
	return &fvcommon.UpdateResponse{Records: toRecordResults(body.Records, header)}, nil
}

// Get fetches records by Skyflow IDs or unique values, from one table or several.
func (v *VaultController) Get(ctx context.Context, request fvcommon.GetRequest, options fvcommon.GetOptions) (*fvcommon.GetResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(fvconsts.METHOD_GET, fvconsts.OPERATION_GET, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateGet(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.GetRecords(ctx, toWireGet(setup.VaultID, request))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.GetResponse) int { return len(b.Records) }))
	if err := finish(fvconsts.METHOD_GET, err); err != nil {
		return nil, err
	}
	return &fvcommon.GetResponse{Records: toRecordResults(body.Records, header)}, nil
}

// Delete deletes records by Skyflow IDs or unique values.
func (v *VaultController) Delete(ctx context.Context, request fvcommon.DeleteRequest, options fvcommon.DeleteOptions) (*fvcommon.DeleteResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(fvconsts.METHOD_DELETE, fvconsts.OPERATION_DELETE, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateDelete(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Records.WithRawResponse.DeleteRecords(ctx, toWireDelete(setup.VaultID, request))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.DeleteResponse) int { return len(b.Records) }))
	if err := finish(fvconsts.METHOD_DELETE, err); err != nil {
		return nil, err
	}
	return &fvcommon.DeleteResponse{Records: toDeleteResults(body.Records, header)}, nil
}

// Detokenize returns the values behind tokens, applying per-token-group redaction.
func (v *VaultController) Detokenize(ctx context.Context, request fvcommon.DetokenizeRequest, options fvcommon.DetokenizeOptions) (*fvcommon.DetokenizeResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(fvconsts.METHOD_DETOKENIZE, fvconsts.OPERATION_DETOKENIZE, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateDetokenize(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Tokens.WithRawResponse.Detokenize(ctx, toWireDetokenize(setup.VaultID, request))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.DetokenizeResponse) int { return len(b.Response) }))
	if err := finish(fvconsts.METHOD_DETOKENIZE, err); err != nil {
		return nil, err
	}
	return &fvcommon.DetokenizeResponse{Records: toDetokenizeResults(body.Response, header)}, nil
}

// GetTokens returns existing tokens for values in deterministic token groups.
func (v *VaultController) GetTokens(ctx context.Context, request fvcommon.GetTokensRequest, options fvcommon.GetTokensOptions) (*fvcommon.GetTokensResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(fvconsts.METHOD_GET_TOKENS, fvconsts.OPERATION_GET_TOKENS, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateGetTokens(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Tokens.WithRawResponse.GetTokens(ctx, toWireGetTokens(setup.VaultID, request))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.GetTokensFromValuesResponse) int { return len(b.Records) }))
	if err := finish(fvconsts.METHOD_GET_TOKENS, err); err != nil {
		return nil, err
	}
	return &fvcommon.GetTokensResponse{Records: toGetTokensResults(body.Records, header)}, nil
}

// Query runs a SELECT query. It returns at most 25 rows per call.
func (v *VaultController) Query(ctx context.Context, request fvcommon.QueryRequest, options fvcommon.QueryOptions) (*fvcommon.QueryResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(fvconsts.METHOD_QUERY, fvconsts.OPERATION_QUERY, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateQuery(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Query.WithRawResponse.ExecuteQuery(ctx, &api.ExecuteQueryRequest{VaultID: setup.VaultID, Query: request.Query})
	body, header, err := result(resp, callErr, nil)
	if err := finish(fvconsts.METHOD_QUERY, err); err != nil {
		return nil, err
	}
	fields := make([]map[string]interface{}, 0, len(body.Records))
	for _, record := range body.Records {
		if record != nil {
			fields = append(fields, record.Data)
		}
	}
	response := &fvcommon.QueryResponse{Fields: fields, RequestId: header.Get(fvconsts.RESPONSE_HEADER_REQUEST_ID)}
	if body.Metadata != nil {
		response.Metadata.Columns = body.Metadata.Columns
	}
	return response, nil
}

// UploadFiles uploads files into file columns. It requests a signed upload URL per column,
// then PUTs each file to its URL, at most MAX_CONCURRENT_UPLOADS at once. The PUTs use
// fvhelpers.UploadHTTPClient: no retries, no overall timeout and no SDK headers. A column
// that can't be uploaded is reported as FAILED or SKIPPED; the other columns are unaffected.
func (v *VaultController) UploadFiles(ctx context.Context, request fvcommon.UploadFilesRequest, options fvcommon.UploadFilesOptions) (*fvcommon.UploadFilesResponse, *fverrors.SkyflowError) {
	files := make([][]fvhelpers.ResolvedFile, len(request.Records))
	setup, err := v.prepare(fvconsts.METHOD_UPLOAD_FILES, fvconsts.OPERATION_UPLOAD_FILES, options.Interceptor, func() *fverrors.SkyflowError {
		if err := fvvalidation.ValidateUploadFiles(request); err != nil {
			return err
		}
		// Read every file up front so a bad file fails the call before anything is uploaded.
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
	resp, callErr := setup.API.Files.WithRawResponse.UploadFiles(ctx, toWireUploadFiles(setup.VaultID, request, files))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.FileUploadResponse) int { return len(b.Records) }))
	if err := finish(fvconsts.METHOD_UPLOAD_FILES, err); err != nil {
		return nil, err
	}

	// Records come back in request order, so record i holds the signed URLs for request record i.
	records := make([]fvcommon.UploadFilesResponseRecord, len(body.Records))
	var uploads []func()
	for i, wireRecord := range body.Records {
		if wireRecord == nil {
			continue
		}
		recordErr := deref(wireRecord.Error)
		records[i] = fvcommon.UploadFilesResponseRecord{
			SkyflowId: wireRecord.SkyflowID,
			TableName: wireRecord.TableName,
			HttpCode:  wireRecord.HTTPCode,
			Error:     recordErr,
			RequestId: fvhelpers.RequestIdFor(header, recordErr),
		}
		if i >= len(request.Records) {
			continue
		}
		columns := request.Records[i].Columns
		records[i].Columns = make([]fvcommon.UploadFilesColumnResult, len(columns))
		for j, column := range columns {
			file := files[i][j]
			columnResult := &records[i].Columns[j]
			*columnResult = fvcommon.UploadFilesColumnResult{Column: column.Column, FileName: file.Name, UploadStatus: fvcommon.UploadStatusSkipped}
			if recordErr != "" {
				columnResult.Error = recordErr
				continue
			}
			signedURL, _ := wireRecord.Data[column.Column].(string)
			if signedURL == "" {
				continue
			}
			uploads = append(uploads, func() {
				columnResult.UploadStatus, columnResult.Error = fvhelpers.PutFile(ctx, fvhelpers.UploadHTTPClient(), signedURL, file)
			})
		}
	}
	fvhelpers.RunConcurrently(uploads, fvconsts.MAX_CONCURRENT_UPLOADS)
	return &fvcommon.UploadFilesResponse{Records: records}, nil
}

// DeleteFiles deletes files from file columns of existing records, including the stored objects.
func (v *VaultController) DeleteFiles(ctx context.Context, request fvcommon.DeleteFilesRequest, options fvcommon.DeleteFilesOptions) (*fvcommon.DeleteFilesResponse, *fverrors.SkyflowError) {
	setup, err := v.prepare(fvconsts.METHOD_DELETE_FILES, fvconsts.OPERATION_DELETE_FILES, options.Interceptor, func() *fverrors.SkyflowError {
		return fvvalidation.ValidateDeleteFiles(request)
	})
	if err != nil {
		return nil, err
	}
	resp, callErr := setup.API.Files.WithRawResponse.DeleteFiles(ctx, toWireDeleteFiles(setup.VaultID, request))
	body, header, err := result(resp, callErr, hasRecords(func(b *api.FileDeleteResponse) int { return len(b.Records) }))
	if err := finish(fvconsts.METHOD_DELETE_FILES, err); err != nil {
		return nil, err
	}
	return &fvcommon.DeleteFilesResponse{Records: toDeleteFilesResults(body.Records, header)}, nil
}
