package audit

import (
	context "context"
	http "net/http"

	core "github.com/skyflowapi/skyflow-go/v2/internal/generated/core"
	generated "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	internal "github.com/skyflowapi/skyflow-go/v2/internal/generated/internal"
	option "github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
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

func (r *RawClient) AuditServiceListAuditEvents(
	ctx context.Context,
	request *generated.AuditServiceListAuditEventsRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1AuditResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/audit/events"
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
	var response *generated.V1AuditResponse
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
	return &core.Response[*generated.V1AuditResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}
