package authentication

import (
	context "context"
	generated "github.com/skyflowapi/skyflow-go/common/internal/generated"
	core "github.com/skyflowapi/skyflow-go/common/internal/generated/core"
	internal "github.com/skyflowapi/skyflow-go/common/internal/generated/internal"
	option "github.com/skyflowapi/skyflow-go/common/internal/generated/option"
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

func (r *RawClient) GetBearerToken(
	ctx context.Context,
	request *generated.GetAuthTokenRequest,
	opts ...option.RequestOption,
) (*core.Response[*generated.GetAuthTokenResponse], error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		r.baseURL,
		"https://%7B%7Bvault_url%7D%7D",
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
		500: func(apiError *core.APIError) error {
			return &generated.InternalServerError{
				APIError: apiError,
			}
		},
	}
	var response *generated.GetAuthTokenResponse
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
	return &core.Response[*generated.GetAuthTokenResponse]{
		StatusCode: raw.StatusCode,
		Header:     raw.Header,
		Body:       response,
	}, nil
}
