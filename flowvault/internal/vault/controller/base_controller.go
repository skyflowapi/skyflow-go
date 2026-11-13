package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	apiclient "github.com/skyflowapi/skyflow-go/flowvault/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/option"
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/flowvault/utils/messages"
)

// BaseController holds one vault's settings and cached bearer token. flowvault's
// VaultController and DetectController share one per vault, so both reuse the same token.
//
// The credential, token and HTTP logic is the shared skyflow-go/common/vaultclient Base;
// this type adds the generated FlowService client. It is safe for concurrent use.
type BaseController struct {
	*vaultclient.Base
}

// ClientSettings are the client-level settings every vault falls back to.
type ClientSettings = vaultclient.ClientSettings

// NewBaseController returns the controller state for one vault.
func NewBaseController(config fvcommon.VaultConfig, client ClientSettings) BaseController {
	return BaseController{Base: vaultclient.New(fvhelpers.Profile(), config, client)}
}

// RequestSetup is everything one API call needs: the vault ID, where to send it, how to
// authenticate, which headers to add, the HTTP client (timeout and retries applied) and
// the generated API client built from all of these.
type RequestSetup struct {
	VaultID    string
	BaseURL    string
	AuthToken  string
	Headers    http.Header
	HTTPClient *http.Client
	API        *apiclient.Client
}

// GenerateToken resolves a bearer token from credentials.
// Uses the shared skyflow-go/common/auth implementation.
func GenerateToken(creds fvcommon.Credentials) (*string, *fverrors.SkyflowError) {
	token, err := fvhelpers.Profile().GenerateToken(creds)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// SetupRequest resolves auth, headers and HTTP settings for one call. The interceptor,
// when set, can add per-call headers. It stops waiting for a bearer token when ctx ends.
func (b BaseController) SetupRequest(ctx context.Context, operation string, interceptor fvcommon.Interceptor) (*RequestSetup, *fverrors.SkyflowError) {
	if ctx == nil {
		fvlogger.Error(logs.NIL_CONTEXT)
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.NIL_CONTEXT)
	}
	call, err := b.PrepareContext(ctx, operation, interceptor)
	if err != nil {
		return nil, err
	}
	setup := &RequestSetup{
		VaultID:    call.VaultID,
		BaseURL:    call.BaseURL,
		AuthToken:  call.AuthToken,
		Headers:    call.Headers,
		HTTPClient: call.HTTPClient,
	}
	// Retries happen in setup.HTTPClient, so the generated client makes a single attempt.
	setup.API = apiclient.NewClient(
		option.WithBaseURL(setup.BaseURL),
		option.WithHTTPClient(finalResponseClient{setup.HTTPClient}),
		option.WithHTTPHeader(setup.Headers),
		option.WithToken(setup.AuthToken),
		option.WithMaxAttempts(1),
	)
	return setup, nil
}

// finalResponseClient returns a response that would be retried as the generated client's
// APIError, because retries are already done. The generated retrier would otherwise sleep
// before giving up, outside the call's timeout and ignoring cancellation.
type finalResponseClient struct {
	client *http.Client
}

// It also checks 2xx bodies before the generated client parses them: a blank body becomes
// {}, an empty response, instead of a decoding failure; a body that is not a JSON object
// becomes an APIError that keeps the real status but not the body, which may hold vault data.
func (c finalResponseClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return successResponse(resp)
	}
	if !vaultclient.ShouldRetry(resp.StatusCode) {
		return resp, nil
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nil, core.NewAPIError(resp.StatusCode, resp.Header, errors.New(string(raw)))
}

// successResponse normalizes a 2xx response body; see finalResponseClient.
func successResponse(resp *http.Response) (*http.Response, error) {
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(string(raw))
	if text == "" {
		text = fvconsts.EMPTY_JSON_OBJECT
	} else {
		var object map[string]json.RawMessage
		if json.Unmarshal(raw, &object) != nil || object == nil {
			body, _ := json.Marshal(map[string]interface{}{fvconsts.ERROR_KEY_ERROR: map[string]interface{}{
				fvconsts.ERROR_KEY_MESSAGE:   fverrors.UNPARSEABLE_SUCCESS_RESPONSE,
				fvconsts.ERROR_KEY_HTTP_CODE: resp.StatusCode,
			}})
			return nil, core.NewAPIError(resp.StatusCode, resp.Header, errors.New(string(body)))
		}
		text = string(raw)
	}
	resp.Body = io.NopCloser(strings.NewReader(text))
	resp.ContentLength = int64(len(text))
	return resp, nil
}
