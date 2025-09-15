package records

import (
	context "context"
	http "net/http"

	generated "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	core "github.com/skyflowapi/skyflow-go/v2/internal/generated/core"
	option "github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	internal "github.com/skyflowapi/skyflow-go/v2/internal/generated/internal"
)

type RawClient struct {
	baseURL string
	caller  *internal.Caller
	header  http.Header
}

func NewRawClient(options *core.RequestOptions) *RawClient {
	return &RawClient{
		baseURL: options.BaseURL,
		caller: internal.NewCaller(
			&internal.CallerParams{
				Client:      options.HTTPClient,
				MaxAttempts: options.MaxAttempts,
			},
		),
		header: options.ToHeader(),
	}
}

func (r *RawClient) RecordServiceBatchOperation(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	request *generated.RecordServiceBatchOperationBody,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1BatchOperationResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v",
		vaultId,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1BatchOperationResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodPost,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1BatchOperationResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) RecordServiceBulkGetRecord(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table that contains the records.
	objectName string,
	request *generated.RecordServiceBulkGetRecordRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1BulkGetRecordResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v",
		vaultId,
		objectName,
	)
	queryParams, err := internal.QueryValues(request)
	if err != nil {
		return nil, err
	}
	if len(queryParams) > 0 {
		endpointURL += "?" + queryParams.Encode()
	}
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1BulkGetRecordResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodGet,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1BulkGetRecordResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) RecordServiceInsertRecord(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	objectName string,
	request *generated.RecordServiceInsertRecordBody,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1InsertRecordResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v",
		vaultId,
		objectName,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1InsertRecordResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodPost,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1InsertRecordResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) RecordServiceBulkDeleteRecord(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	objectName string,
	request *generated.RecordServiceBulkDeleteRecordBody,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1BulkDeleteRecordResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v",
		vaultId,
		objectName,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1BulkDeleteRecordResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodDelete,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1BulkDeleteRecordResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) RecordServiceGetRecord(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	objectName string,
	// `skyflow_id` of the record.
	id string,
	request *generated.RecordServiceGetRecordRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1FieldRecords], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v/%v",
		vaultId,
		objectName,
		id,
	)
	queryParams, err := internal.QueryValues(request)
	if err != nil {
		return nil, err
	}
	if len(queryParams) > 0 {
		endpointURL += "?" + queryParams.Encode()
	}
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1FieldRecords
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodGet,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1FieldRecords]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) RecordServiceUpdateRecord(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	objectName string,
	// `skyflow_id` of the record.
	id string,
	request *generated.RecordServiceUpdateRecordBody,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1UpdateRecordResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v/%v",
		vaultId,
		objectName,
		id,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1UpdateRecordResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodPut,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1UpdateRecordResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) RecordServiceDeleteRecord(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	objectName string,
	// `skyflow_id` of the record to delete.
	id string,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1DeleteRecordResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v/%v",
		vaultId,
		objectName,
		id,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1DeleteRecordResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodDelete,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1DeleteRecordResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) FileServiceUploadFile(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	objectName string,
	// `skyflow_id` of the record.
	id string,
	request *generated.FileServiceUploadFileRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1UpdateRecordResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v/%v/files",
		vaultId,
		objectName,
		id,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "multipart/form-data")
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	writer := internal.NewMultipartWriter()
	if err := writer.WriteFile("file", request.File); err != nil {
		return nil, err
	}
	if request.ColumnName != nil {
		if err := writer.WriteField("columnName", *request.ColumnName); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	headers.Set("Content-Type", writer.ContentType())

	var response *generated.V1UpdateRecordResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodPost,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         writer.Buffer(),
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1UpdateRecordResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) FileServiceDeleteFile(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	tableName string,
	// `skyflow_id` of the record.
	id string,
	// Name of the column that contains the file.
	columnName string,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1DeleteFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v/%v/files/%v",
		vaultId,
		tableName,
		id,
		columnName,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1DeleteFileResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodDelete,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1DeleteFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) FileServiceGetFileScanStatus(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	// Name of the table.
	tableName string,
	// `skyflow_id` of the record.
	id string,
	// Name of the column that contains the file.
	columnName string,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1GetFileScanStatusResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/%v/%v/files/%v/scan-status",
		vaultId,
		tableName,
		id,
		columnName,
	)
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1GetFileScanStatusResponse
	raw, err := r.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodGet,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	)
	if err != nil {
		return nil, err
	}
	return &core.Response[*generated.V1GetFileScanStatusResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}
