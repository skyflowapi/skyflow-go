package controller

import (
	"errors"
	"io"
	"net/http"
	"os"
	"sync"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	apiclient "github.com/skyflowapi/skyflow-go/flowvault/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/generated/option"
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvvalidation "github.com/skyflowapi/skyflow-go/flowvault/internal/validation"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/flowvault/utils/messages"
)

// BaseController holds shared auth state and HTTP client logic for flowvault vault operations.
// flowvault's VaultController embeds this — equivalent to Java's AbstractBaseVaultController.
//
// It is safe for concurrent use: calls read a snapshot of the settings, and the client can
// change them (SetConfig, SetClientSettings) while calls are running.
type BaseController struct {
	settingsMu sync.RWMutex
	config     fvcommon.VaultConfig
	client     ClientSettings
	// authVersion changes whenever the credentials may have changed, so a token generated
	// from older credentials is never reused.
	authVersion uint64

	tokenMu      sync.Mutex
	token        string
	tokenVersion uint64
}

// ClientSettings are the client-level settings every vault falls back to.
type ClientSettings struct {
	Credentials *fvcommon.Credentials
	HttpConfig  fvcommon.HttpConfig
}

// settings is a consistent copy of the controller's settings for one call.
type settings struct {
	config      fvcommon.VaultConfig
	client      ClientSettings
	authVersion uint64
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
	opts := fvcommon.BearerTokenOptions{}
	if creds.Roles != nil {
		opts.RoleIds = creds.Roles
	}
	if creds.Context != nil {
		opts.Ctx = creds.Context
	}
	var token string
	switch {
	case creds.Token != "":
		token = creds.Token
	case creds.ApiKey != "":
		token = creds.ApiKey
	case creds.Path != "":
		resp, err := fvhelpers.Auth().GenerateBearerToken(creds.Path, opts)
		if err != nil {
			return nil, err
		}
		token = resp.AccessToken
	case creds.CredentialsString != "":
		resp, err := fvhelpers.Auth().GenerateBearerTokenFromCreds(creds.CredentialsString, opts)
		if err != nil {
			return nil, err
		}
		token = resp.AccessToken
	default:
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.EMPTY_CREDENTIALS)
	}
	return &token, nil
}

// Config returns a copy of the vault config.
func (b *BaseController) Config() fvcommon.VaultConfig {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return b.config
}

// SetConfig replaces the vault config. Set resetAuth when the credentials changed.
func (b *BaseController) SetConfig(config fvcommon.VaultConfig, resetAuth bool) {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	b.config = config
	if resetAuth {
		b.authVersion++
	}
}

// ClientSettings returns the client-level settings this vault uses.
func (b *BaseController) ClientSettings() ClientSettings {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return b.client
}

// SetClientSettings replaces the client-level settings. A change of client credentials
// drops the cached token.
func (b *BaseController) SetClientSettings(client ClientSettings) {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	if client.Credentials != b.client.Credentials {
		b.authVersion++
	}
	b.client = client
}

// ResetAuth drops the cached bearer token so the next call re-authenticates.
func (b *BaseController) ResetAuth() {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	b.authVersion++
}

func (b *BaseController) snapshot() settings {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return settings{config: b.config, client: b.client, authVersion: b.authVersion}
}

// SetupRequest resolves auth, headers and HTTP settings for one call. The interceptor,
// when set, can add per-call headers.
func (b *BaseController) SetupRequest(operation string, interceptor fvcommon.Interceptor) (*RequestSetup, *fverrors.SkyflowError) {
	current := b.snapshot()
	headers, err := buildHeaders(operation, interceptor)
	if err != nil {
		return nil, err
	}
	authToken, err := b.authToken(current)
	if err != nil {
		return nil, err
	}
	httpConfig := fvhelpers.ResolveHttpConfig(current.config.HttpConfig(), current.client.HttpConfig)
	setup := &RequestSetup{
		VaultID:    current.config.VaultId,
		BaseURL:    baseURL(current.config),
		AuthToken:  authToken,
		Headers:    headers,
		HTTPClient: fvhelpers.NewHTTPClient(httpConfig),
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

func (c finalResponseClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil || !fvhelpers.ShouldRetry(resp.StatusCode) {
		return resp, err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nil, core.NewAPIError(resp.StatusCode, resp.Header, errors.New(string(raw)))
}

// buildHeaders merges the interceptor's per-call headers and the SDK metadata header.
func buildHeaders(operation string, interceptor fvcommon.Interceptor) (http.Header, *fverrors.SkyflowError) {
	headers := http.Header{}
	if interceptor != nil {
		rc := &fvcommon.RequestContext{Operation: operation}
		interceptor(rc)
		if added := rc.Headers(); len(added) > 0 {
			if err := fvvalidation.ValidateCustomHeaders(added, operation); err != nil {
				return nil, err
			}
			for key, value := range added {
				headers.Set(string(key), value)
			}
		}
	}
	headers.Set(fvconsts.SDK_METRICS_HEADER_KEY, fvhelpers.CreateSDKMetadata())
	return headers, nil
}

// authToken returns the API key or bearer token to send, generating and caching a
// bearer token from service-account credentials when needed.
func (b *BaseController) authToken(current settings) (string, *fverrors.SkyflowError) {
	creds, err := resolveCredentials(current)
	if err != nil {
		return "", err
	}
	if creds.ApiKey != "" {
		return creds.ApiKey, nil
	}
	if creds.Token != "" {
		if fvhelpers.Auth().IsExpired(creds.Token) {
			fvlogger.Error(logs.BEARER_TOKEN_EXPIRED)
			return "", fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.BEARER_TOKEN_EXPIRED)
		}
		return creds.Token, nil
	}
	b.tokenMu.Lock()
	defer b.tokenMu.Unlock()
	if b.token != "" && b.tokenVersion == current.authVersion && !fvhelpers.Auth().IsExpired(b.token) {
		fvlogger.Info(logs.REUSE_BEARER_TOKEN)
		return b.token, nil
	}
	fvlogger.Info(logs.GENERATE_BEARER_TOKEN_TRIGGERED)
	token, err := GenerateToken(*creds)
	if err != nil {
		return "", err
	}
	// A call that started before a credentials change must not replace a newer token.
	if current.authVersion >= b.tokenVersion {
		b.token, b.tokenVersion = *token, current.authVersion
	}
	return *token, nil
}

func baseURL(config fvcommon.VaultConfig) string {
	if config.BaseVaultUrl != "" {
		return config.BaseVaultUrl
	}
	if config.BaseVaultURL != "" {
		return config.BaseVaultURL
	}
	return fvhelpers.GetURLWithEnv(config.Env, config.ClusterId)
}

// resolveCredentials picks vault credentials, then client credentials, then SKYFLOW_CREDENTIALS.
func resolveCredentials(current settings) (*fvcommon.Credentials, *fverrors.SkyflowError) {
	if !fvcommon.IsCredentialsEmpty(current.config.Credentials) {
		creds := current.config.Credentials
		return &creds, nil
	}
	if current.client.Credentials != nil && !fvcommon.IsCredentialsEmpty(*current.client.Credentials) {
		return current.client.Credentials, nil
	}
	if envCreds := os.Getenv(fvconsts.SKYFLOW_CREDENTIALS_ENV); envCreds != "" {
		return &fvcommon.Credentials{CredentialsString: envCreds}, nil
	}
	fvlogger.Error(logs.EMPTY_CREDENTIALS)
	return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.EMPTY_CREDENTIALS)
}
