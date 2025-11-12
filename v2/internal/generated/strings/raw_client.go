package strings

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

func (r *RawClient) DeidentifyString(
	ctx context.Context,
	request *generated.DeidentifyStringRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.DeidentifyStringResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/deidentify/string"
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
	var response *generated.DeidentifyStringResponse
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
	return &core.Response[*generated.DeidentifyStringResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}

func (r *RawClient) ReidentifyString(
	ctx context.Context,
	request *generated.ReidentifyStringRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.IdentifyResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/detect/reidentify/string"
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
	var response *generated.IdentifyResponse
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
	return &core.Response[*generated.IdentifyResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}
