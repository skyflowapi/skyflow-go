package query

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

func (r *RawClient) QueryServiceExecuteQuery(
	ctx context.Context,
	// ID of the vault.
	vaultId string,
	request *generated.QueryServiceExecuteQueryBody,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1GetQueryResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/v1/vaults/%v/query",
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
	var response *generated.V1GetQueryResponse
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
	return &core.Response[*generated.V1GetQueryResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}
