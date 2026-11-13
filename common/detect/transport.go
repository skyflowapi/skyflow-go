package detect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/internal/constants"
	gen "github.com/skyflowapi/skyflow-go/common/internal/generated"
	apiclient "github.com/skyflowapi/skyflow-go/common/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/common/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/common/internal/generated/option"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
)

// The Detect endpoints are called through the Fern-generated client in
// common/internal/generated. convert.go maps between the public types and the generated ones.

// newAPI builds the generated client for one call. Retries happen in call.HTTPClient, so the
// generated client makes a single attempt.
func newAPI(call *vaultclient.Call) *apiclient.Client {
	return apiclient.NewClient(
		option.WithBaseURL(call.BaseURL),
		option.WithHTTPClient(finalResponseClient{call.HTTPClient}),
		option.WithHTTPHeader(call.Headers),
		option.WithToken(call.AuthToken),
		option.WithMaxAttempts(1),
	)
}

// finalResponseClient returns a response that would be retried as the generated client's
// APIError, because retries are already done. The generated retrier would otherwise sleep
// before giving up, outside the call's timeout and ignoring cancellation.
type finalResponseClient struct {
	client *http.Client
}

func (c finalResponseClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil || !vaultclient.ShouldRetry(resp.StatusCode) {
		return resp, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nil, core.NewAPIError(resp.StatusCode, resp.Header, errors.New(string(raw)))
}

// apiError turns a generated-client error into a SkyflowError, as flowvault's vault
// operations do:
//   - An API error (the server answered) has the response's HTTP status as its code, even
//     when the body carries its own http_code, and keeps the request ID and the server's body
//     (GetResponseBody). See responseError for the message.
//   - Any other error got no usable HTTP response. It is REQUEST_CANCELLED ("499") when the
//     caller's context ended, else TRANSPORT_ERROR ("0", no status): a network failure or the
//     client's Timeout. ctx decides, not the error: net/http's timeout error also matches
//     context.DeadlineExceeded. The error is kept as the cause, so errors.Is works.
func (c *Controller) apiError(ctx context.Context, err error) *skyflowError.SkyflowError {
	var apiErr *core.APIError
	if errors.As(err, &apiErr) {
		return c.responseError(apiErr)
	}
	code := skyflowError.TRANSPORT_ERROR
	if ctx.Err() != nil {
		code = skyflowError.REQUEST_CANCELLED
	}
	return c.base.Profile().NewErrorWithCause(code, err.Error(), err)
}

// responseError converts an error response. When the body has no Skyflow error object, the
// message is the first of: an "error" string, a top-level "message", the raw body (its first
// MAX_RAW_ERROR_BODY bytes), or "Error with status code N" for an empty body.
func (c *Controller) responseError(apiErr *core.APIError) *skyflowError.SkyflowError {
	header := apiErr.Header
	if header == nil {
		header = http.Header{}
	}
	raw := ""
	if apiErr.Unwrap() != nil {
		raw = apiErr.Unwrap().Error()
	}
	var body map[string]interface{}
	if json.Unmarshal([]byte(raw), &body) != nil {
		body = nil
	}
	errorBody := map[string]interface{}{}
	if serverError, ok := body[constants.ERROR_KEY_ERROR].(map[string]interface{}); ok {
		for key, value := range serverError {
			errorBody[key] = value
		}
		// grpcCode and httpStatus are read when the snake_case keys are missing.
		for snake, camel := range map[string]string{constants.ERROR_KEY_GRPC_CODE: errorKeyGrpcCodeCamel, constants.ERROR_KEY_HTTP_STATUS: errorKeyHttpStatusCamel} {
			if _, present := errorBody[snake]; !present && errorBody[camel] != nil {
				errorBody[snake] = errorBody[camel]
			}
		}
		// An error-from-client header is added to the details.
		if fromClient := header.Get(constants.ERROR_FROM_CLIENT); fromClient != "" {
			details, _ := errorBody[constants.ERROR_KEY_DETAILS].([]interface{})
			errorBody[constants.ERROR_KEY_DETAILS] = append(slices.Clone(details), map[string]interface{}{constants.ERROR_KEY_FROM_CLIENT: fromClient})
		}
	} else {
		errorBody[constants.ERROR_KEY_MESSAGE] = fallbackMessage(body, raw, apiErr.StatusCode)
	}
	errorBody[constants.ERROR_KEY_HTTP_CODE] = apiErr.StatusCode
	normalized, _ := json.Marshal(map[string]interface{}{constants.ERROR_KEY_ERROR: errorBody}) // decoded JSON always marshals
	opts := append(slices.Clone(c.base.Profile().ErrorOptions), skyflowError.WithResponseBody(body))
	return skyflowError.SkyflowErrorApi(fmt.Errorf(apiErrorFormat, apiErr.StatusCode, normalized), header, opts...)
}

// fallbackMessage is the message for an error body without a Skyflow error object.
func fallbackMessage(body map[string]interface{}, raw string, status int) string {
	for _, key := range []string{constants.ERROR_KEY_ERROR, constants.ERROR_KEY_MESSAGE} {
		if text, isText := body[key].(string); isText && strings.TrimSpace(text) != "" {
			return truncateBody(text)
		}
	}
	if strings.TrimSpace(raw) != "" {
		return truncateBody(raw)
	}
	return fmt.Sprintf(errorWithStatusCode, status)
}

// truncateBody keeps the first MAX_RAW_ERROR_BODY bytes of an error body, without splitting a
// multi-byte character. A gateway or proxy can echo the request, which carries sensitive data.
func truncateBody(body string) string {
	if len(body) <= constants.MAX_RAW_ERROR_BODY {
		return body
	}
	cut := constants.MAX_RAW_ERROR_BODY
	for cut > 0 && !utf8.RuneStart(body[cut]) {
		cut--
	}
	return body[:cut] + constants.TRUNCATED_SUFFIX
}

func (c *Controller) emptyResponse(method string) *skyflowError.SkyflowError {
	return c.base.Profile().NewError(skyflowError.SERVER, fmt.Sprintf(c.errMsg(errEmptyResponse), method))
}

func (c *Controller) callDeidentifyString(ctx context.Context, call *vaultclient.Call, request DeidentifyStringRequest) (*DeidentifyStringResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).StringsV2.DeidentifyStringV2(ctx, &gen.DeidentifyStringRequestV2{
		Text:            request.Text,
		ConfigurationID: optString(request.ConfigurationId),
		Configuration:   requestConfiguration(request.ConfigurationId, request.Configuration, call.VaultID),
	})
	if err != nil {
		return nil, c.apiError(ctx, err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodDeidentifyString)
	}
	return fromGenDeidentifyString(resp), nil
}

func (c *Controller) callDeidentifyFile(ctx context.Context, call *vaultclient.Call, request DeidentifyFileRequest) (string, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).FilesV2.DeidentifyFileV2(ctx, &gen.DeidentifyFileRequestV2{
		DataSource:      gen.DeidentifyFileRequestV2DataSource(deidentifySourceName(request.DataSource)),
		Value:           request.Value,
		DataFormat:      optEnum[gen.DeidentifyFileRequestV2DataFormat](wireFormat(request.DataFormat)),
		ConfigurationID: optString(request.ConfigurationId),
		Configuration:   requestConfiguration(request.ConfigurationId, request.Configuration, call.VaultID),
	})
	if err != nil {
		return "", c.apiError(ctx, err)
	}
	if resp == nil || resp.RunID == nil || *resp.RunID == "" {
		return "", c.emptyResponse(methodDeidentifyFile)
	}
	return *resp.RunID, nil
}

func (c *Controller) callGetRun(ctx context.Context, call *vaultclient.Call, runId string) (*GetRunResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).FilesV2.GetRunV2(ctx, &gen.GetRunV2Request{RunID: runId, VaultID: call.VaultID})
	if err != nil {
		return nil, c.apiError(ctx, err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodGetRun)
	}
	return fromGenRun(runId, resp, c.sources.Deidentify), nil
}

func (c *Controller) callReidentifyString(ctx context.Context, call *vaultclient.Call, request ReidentifyStringRequest) (*ReidentifyStringResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).StringsV2.ReidentifyStringV2(ctx, &gen.ReidentifyStringRequestV2{
		VaultID:        call.VaultID,
		Text:           request.Text,
		RedactionLevel: toGenRedactionLevels(request.RedactionLevel),
	})
	if err != nil {
		return nil, c.apiError(ctx, err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodReidentifyString)
	}
	return &ReidentifyStringResponse{ProcessedText: deref(resp.ProcessedText), Metrics: fromGenMetrics(resp.Metrics)}, nil
}

func (c *Controller) callReidentifyFile(ctx context.Context, call *vaultclient.Call, request ReidentifyFileRequest) (*ReidentifyFileResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).FilesV2.ReidentifyFileV2(ctx, &gen.ReidentifyFileRequestV2{
		VaultID:        call.VaultID,
		DataSource:     gen.ReidentifyFileRequestV2DataSource(reidentifySourceName(request.DataSource)),
		Value:          request.Value,
		DataFormat:     optEnum[gen.ReidentifyFileRequestV2DataFormat](strings.ToLower(string(request.DataFormat))),
		RedactionLevel: toGenRedactionLevels(request.RedactionLevel),
	})
	if err != nil {
		return nil, c.apiError(ctx, err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodReidentifyFile)
	}
	return &ReidentifyFileResponse{
		Status:     runStatus(derefEnum(resp.Status)),
		OutputType: resolveReidentifySource(derefEnum(resp.OutputType), c.sources.Reidentify),
		Output:     fromGenReidentifyOutputs(resp.Output),
		Metrics:    fromGenMetrics(resp.Metrics),
	}, nil
}

func (c *Controller) callCheckGuardrails(ctx context.Context, call *vaultclient.Call, request CheckGuardrailsRequest) (*CheckGuardrailsResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).GuardrailsV2.CheckGuardrailsV2(ctx, &gen.DetectGuardrailsRequestV2{
		VaultID:       call.VaultID,
		Text:          request.Text,
		CheckToxicity: request.CheckToxicity,
		DenyTopics:    request.DenyTopics,
	})
	if err != nil {
		return nil, c.apiError(ctx, err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodCheckGuardrails)
	}
	return &CheckGuardrailsResponse{
		Text:        resp.Text,
		Validation:  knownOr(string(resp.Validation), guardrailsValidations, GuardrailsValidationUnknown),
		Toxic:       resp.Toxic,
		DeniedTopic: resp.DeniedTopic,
	}, nil
}
