package files

import (
	context "context"
	generated "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	core "github.com/skyflowapi/skyflow-go/v2/internal/generated/core"
	internal "github.com/skyflowapi/skyflow-go/v2/internal/generated/internal"
	option "github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	http "net/http"
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

func (r *RawClient) DeidentifyFile(
	ctx context.Context,
	request *generated.DeidentifyFileRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifyAudio(
	ctx context.Context,
	request *generated.DeidentifyFileAudioRequestDeidentifyAudio,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/audio"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifyDocument(
	ctx context.Context,
	request *generated.DeidentifyFileRequestDeidentifyDocument,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/document"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifyPdf(
	ctx context.Context,
	request *generated.DeidentifyFileDocumentPdfRequestDeidentifyPdf,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/document/pdf"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifyImage(
	ctx context.Context,
	request *generated.DeidentifyFileImageRequestDeidentifyImage,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/image"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifyPresentation(
	ctx context.Context,
	request *generated.DeidentifyFileRequestDeidentifyPresentation,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/presentation"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifySpreadsheet(
	ctx context.Context,
	request *generated.DeidentifyFileRequestDeidentifySpreadsheet,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/spreadsheet"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifyStructuredText(
	ctx context.Context,
	request *generated.DeidentifyFileRequestDeidentifyStructuredText,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/structured_text"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) DeidentifyText(
	ctx context.Context,
	request *generated.DeidentifyFileRequestDeidentifyText,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/file/text"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DeidentifyFileResponse
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
	return &core.Response[*generated.DeidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) ReidentifyFile(
	ctx context.Context,
	request *generated.ReidentifyFileRequestReidentifyFile,
	opts ...option.RequestOption,
) (*core.Response[*generated.ReidentifyFileResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/reidentify/file"
	headers := internal.MergeHeaders(
		r.header.Clone(),
		options.ToHeader(),
	)
	headers.Add("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.ReidentifyFileResponse
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
	return &core.Response[*generated.ReidentifyFileResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) GetRun(
	ctx context.Context,
	runId string,
	request *generated.GetRunRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.DetectRunsResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/detect/runs/%v",
		runId,
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
		400: func(apiError *core.APIError) error {
			return &generated.BadRequestError{
				APIError: apiError,
			}
		},
		401: func(apiError *core.APIError) error {
			return &generated.UnauthorizedError{
				APIError: apiError,
			}
		},
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.DetectRunsResponse
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
	return &core.Response[*generated.DetectRunsResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}
