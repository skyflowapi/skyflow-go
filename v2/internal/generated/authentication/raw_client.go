package authentication

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

func (r *RawClient) AuthenticationServiceGetAuthToken(
	ctx context.Context,
	request *generated.V1GetAuthTokenRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.V1GetAuthTokenResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://identifier.vault.skyflowapis.com",
	)
	endpointURL := baseURL + "/v1/auth/sa/oauth/token"
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
		404: func(apiError *core.APIError) error {
			return &generated.NotFoundError{
				APIError: apiError,
			}
		},
	}
	var response *generated.V1GetAuthTokenResponse
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
	return &core.Response[*generated.V1GetAuthTokenResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}
