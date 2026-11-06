package controller

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

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

	// tokenSlotCh holds a value while one call reads or generates the bearer token. It is a
	// channel rather than a mutex so a call whose context ends stops waiting for it. Use
	// tokenSlot(), which creates it, so a zero BaseController works too.
	tokenSlotCh   chan struct{}
	tokenSlotOnce sync.Once
	token         string
	tokenVersion  uint64
}

// tokenRefreshMargin is how long before it expires a cached bearer token is replaced, so it
// does not expire during a call and clock skew with the server does not make it fail early.
const tokenRefreshMargin = 2 * time.Minute

// tokenResult is the outcome of generating a bearer token.
type tokenResult struct {
	token string
	err   *fverrors.SkyflowError
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
func (b *BaseController) SetupRequest(ctx context.Context, operation string, interceptor fvcommon.Interceptor) (*RequestSetup, *fverrors.SkyflowError) {
	if ctx == nil {
		fvlogger.Error(logs.NIL_CONTEXT)
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.NIL_CONTEXT)
	}
	current := b.snapshot()
	headers, err := buildHeaders(operation, interceptor)
	if err != nil {
		return nil, err
	}
	authToken, err := b.authToken(ctx, current)
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
// bearer token from service-account credentials when needed. When ctx ends first it returns
// an error wrapping ctx.Err(); a generation already started still finishes and is cached.
func (b *BaseController) authToken(ctx context.Context, current settings) (string, *fverrors.SkyflowError) {
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
	return b.bearerToken(ctx, *creds, current.authVersion)
}

// bearerToken returns the cached bearer token for the credentials at authVersion, or
// generates one. Generation runs on its own goroutine, which owns the token slot until it
// ends, so a call whose context ends can return while the token is still cached for others.
func (b *BaseController) bearerToken(ctx context.Context, creds fvcommon.Credentials, authVersion uint64) (string, *fverrors.SkyflowError) {
	release, err := b.acquireTokenSlot(ctx)
	if err != nil {
		return "", err
	}
	if token, ok := b.cachedToken(authVersion); ok {
		release()
		fvlogger.Info(logs.REUSE_BEARER_TOKEN)
		return token, nil
	}
	fvlogger.Info(logs.GENERATE_BEARER_TOKEN_TRIGGERED)
	result := b.generateInBackground(creds, authVersion, release)
	select {
	case r := <-result:
		return r.token, r.err
	case <-ctx.Done():
		return "", cancelled(ctx)
	}
}

// acquireTokenSlot waits for the token slot and returns the function that frees it, or an
// error when ctx ends first.
func (b *BaseController) acquireTokenSlot(ctx context.Context) (release func(), err *fverrors.SkyflowError) {
	// Checked first: select picks at random when the slot is also free.
	if ctx.Err() != nil {
		return nil, cancelled(ctx)
	}
	slot := b.tokenSlot()
	select {
	case slot <- struct{}{}:
		return func() { <-slot }, nil
	case <-ctx.Done():
		return nil, cancelled(ctx)
	}
}

// cachedToken returns the cached token when it belongs to authVersion and isn't about to
// expire. The caller holds the token slot.
func (b *BaseController) cachedToken(authVersion uint64) (string, bool) {
	if b.token == "" || b.tokenVersion != authVersion || fvhelpers.Auth().ExpiresWithin(b.token, tokenRefreshMargin) {
		return "", false
	}
	return b.token, true
}

// generateInBackground generates a token on a new goroutine, caches it, then calls release.
// The returned channel receives the outcome.
func (b *BaseController) generateInBackground(creds fvcommon.Credentials, authVersion uint64, release func()) <-chan tokenResult {
	result := make(chan tokenResult, 1)
	go func() {
		defer release()
		token, err := GenerateToken(creds)
		if err != nil {
			result <- tokenResult{err: err}
			return
		}
		// A call that started before a credentials change must not replace a newer token.
		if authVersion >= b.tokenVersion {
			b.token, b.tokenVersion = *token, authVersion
		}
		result <- tokenResult{token: *token}
	}()
	return result
}

// tokenSlot returns the one-slot channel that guards the cached token, creating it on
// first use.
func (b *BaseController) tokenSlot() chan struct{} {
	b.tokenSlotOnce.Do(func() { b.tokenSlotCh = make(chan struct{}, 1) })
	return b.tokenSlotCh
}

// cancelled is the error for a call whose context ended while it waited for a bearer token.
func cancelled(ctx context.Context) *fverrors.SkyflowError {
	fvlogger.Error(logs.BEARER_TOKEN_CANCELLED)
	return fverrors.NewSkyflowErrorWithCause(fverrors.REQUEST_CANCELLED, fverrors.BEARER_TOKEN_CANCELLED, ctx.Err())
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
