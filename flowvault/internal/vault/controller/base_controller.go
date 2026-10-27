package controller

import (
	"net/http"

	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	apiclient "github.com/skyflowapi/skyflow-go/flowvault/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/option"
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// BaseController holds one vault's settings and cached bearer token. flowvault's
// VaultController and DetectController share one per vault, so both reuse the same token.
// Equivalent to Java's AbstractBaseVaultController.
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
// when set, can add per-call headers.
func (b BaseController) SetupRequest(operation string, interceptor fvcommon.Interceptor) (*RequestSetup, *fverrors.SkyflowError) {
	call, err := b.Prepare(operation, interceptor)
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
		option.WithHTTPClient(setup.HTTPClient),
		option.WithHTTPHeader(setup.Headers),
		option.WithToken(setup.AuthToken),
		option.WithMaxAttempts(1),
	)
	return setup, nil
}
