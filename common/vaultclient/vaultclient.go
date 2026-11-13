// Package vaultclient is the per-vault request plumbing shared by the Skyflow Go SDKs:
// which credentials a call uses, the cached bearer token, the HTTP client with timeout and
// retries, and the headers every call sends. flowvault's vault controller and the shared
// detect controller (common/detect) both build their calls on a Base.
//
// This package exists to be shared by the SDKs in the skyflow-go repository. It is not part
// of the public SDK API: use the client package of the SDK you depend on instead. Its API
// may change without notice.
package vaultclient

import (
	"context"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/skyflowapi/skyflow-go/common/auth"
	"github.com/skyflowapi/skyflow-go/common/config"
	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
	sa "github.com/skyflowapi/skyflow-go/common/serviceaccount"
)

// Profile describes the SDK a Base works for: the prefixes its errors and logs carry, its
// service-account authenticator, how it builds a vault URL from a cluster ID, the
// sky-metadata header value and its HTTP defaults.
type Profile struct {
	// SDKPrefix is prepended to error messages, e.g. "Skyflow FlowVault Go SDK v1.0.0".
	SDKPrefix string
	// LogPrefix is prepended to log lines, e.g. "[ Skyflow FlowVault Go SDK v1.0.0 ] ".
	LogPrefix string
	// Auth generates bearer tokens from service-account credentials.
	Auth *auth.Service
	// ClusterURL builds the vault base URL when the config has no BaseVaultUrl.
	ClusterURL func(env config.Env, clusterId string) string
	// SDKMetadata returns the sky-metadata header value.
	SDKMetadata func() string
	// Defaults apply when neither the vault nor the client sets an HTTP value.
	Defaults HttpDefaults

	// ValidateHeaders checks custom headers; tag names their source in the error. Nil uses
	// Profile.ValidateCustomHeaders.
	ValidateHeaders func(headers map[config.CustomHeaderKey]string, tag string) *skyflowError.SkyflowError
	// ClientHeadersTag names the client-level headers in validation errors. When set, they
	// are validated on every call that has them.
	ClientHeadersTag string
	// ErrTokenExpired, when set, replaces the expired-bearer-token error message (SDK prefix
	// included).
	ErrTokenExpired string
	// ErrEmptyCredentials, when set, replaces the no-credentials error message (SDK prefix
	// included).
	ErrEmptyCredentials string
	// Logs, when set, replace or add log lines.
	Logs Logs
	// ErrorOptions are applied to every error built for this SDK, by Base and by the
	// controllers that share it (common/detect), e.g. errors.RawMessage() for flowvault.
	ErrorOptions []skyflowError.ErrorOption
}

// Logs are an SDK's own log lines, SDK log prefix included. An empty field uses the
// default line, or logs nothing for the lines that have no default.
type Logs struct {
	TokenExpired           string // replaces the default
	GenerateBearerToken    string // replaces the default
	ReuseBearerToken       string // replaces the default
	UsingApiKey            string // info, when a call uses an API key; no default
	UsingBearerToken       string // info, when a call uses a token from the credentials; no default
	DeprecatedBaseVaultURL string // warning, when the config sets BaseVaultURL; no default
}

func (p Profile) errMsg(base string) string { return p.SDKPrefix + base }

// NewError is errors.NewSkyflowError with the profile's ErrorOptions.
func (p Profile) NewError(code skyflowError.ErrorCodesEnum, message string) *skyflowError.SkyflowError {
	return skyflowError.NewSkyflowError(code, message, p.ErrorOptions...)
}

// NewErrorWithCause is errors.NewSkyflowErrorWithCause with the profile's ErrorOptions.
func (p Profile) NewErrorWithCause(code skyflowError.ErrorCodesEnum, message string, cause error) *skyflowError.SkyflowError {
	return skyflowError.NewSkyflowErrorWithCause(code, message, cause, p.ErrorOptions...)
}

// ApiError is errors.SkyflowErrorApi with the profile's ErrorOptions.
func (p Profile) ApiError(err error, header http.Header) *skyflowError.SkyflowError {
	return skyflowError.SkyflowErrorApi(err, header, p.ErrorOptions...)
}
func (p Profile) logMsg(base string) string { return p.LogPrefix + base }

// logLine returns the SDK's own line when set, else the default with the log prefix.
func (p Profile) logLine(own, base string) string {
	if own != "" {
		return own
	}
	return p.logMsg(base)
}

func (p Profile) emptyCredentialsMsg() string {
	if p.ErrEmptyCredentials != "" {
		return p.ErrEmptyCredentials
	}
	return p.errMsg(errEmptyCredentials)
}

func (p Profile) tokenExpiredMsg() string {
	if p.ErrTokenExpired != "" {
		return p.ErrTokenExpired
	}
	return p.errMsg(errBearerTokenExpired)
}

// ClientSettings are the client-level settings every vault falls back to.
type ClientSettings struct {
	Credentials *config.Credentials
	HttpConfig  config.HttpConfig
	// Headers are sent with every call. A call's own header of the same name replaces one.
	Headers map[config.CustomHeaderKey]string
}

// Call is everything one API call needs: the vault ID, where to send it, how to
// authenticate, which headers to add and the HTTP client with timeout and retries applied.
type Call struct {
	VaultID    string
	BaseURL    string
	AuthToken  string
	Headers    http.Header
	HTTPClient *http.Client
}

// Base holds one vault's settings and its cached bearer token.
//
// It is safe for concurrent use: calls read a snapshot of the settings, and the client can
// change them (SetConfig, SetClientSettings) while calls are running. Settings are copied
// on the way in and out, so a caller that later changes a credentials struct, roles slice or
// headers map it passed in or got back does not change the Base.
type Base struct {
	profile Profile

	settingsMu sync.RWMutex
	config     config.VaultConfig
	client     ClientSettings
	// authVersion changes whenever the credentials may have changed, so a token generated
	// from older credentials is never reused.
	authVersion uint64

	// tokenLock holds a value while one call reads or generates the bearer token. It is a
	// channel rather than a mutex so a call whose context ends stops waiting for it.
	tokenLock    chan struct{}
	token        string
	tokenVersion uint64
	// refreshFailedFor is the rejected token whose replacement after a 401 failed, so other
	// calls still holding it don't try again. Guarded by tokenLock.
	refreshFailedFor string
}

// settings is a consistent copy of the Base's settings for one call.
type settings struct {
	config      config.VaultConfig
	client      ClientSettings
	authVersion uint64
}

// New returns a Base for one vault.
func New(profile Profile, vaultConfig config.VaultConfig, client ClientSettings) *Base {
	return &Base{
		profile:   profile,
		config:    cloneVaultConfig(vaultConfig),
		client:    cloneClientSettings(client),
		tokenLock: make(chan struct{}, 1),
	}
}

// Profile returns the SDK profile this Base works for.
func (b *Base) Profile() Profile {
	return b.profile
}

// Config returns a copy of the vault config.
func (b *Base) Config() config.VaultConfig {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return cloneVaultConfig(b.config)
}

// SetConfig replaces the vault config. A change of the vault credentials drops the cached
// token; resetAuth drops it in any case, for credentials re-supplied unchanged whose source
// may have changed (for example a credentials file rewritten at the same Path).
func (b *Base) SetConfig(vaultConfig config.VaultConfig, resetAuth bool) {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	if resetAuth || !sameCredentials(&vaultConfig.Credentials, &b.config.Credentials) {
		b.authVersion++
	}
	b.config = cloneVaultConfig(vaultConfig)
}

// ClientSettings returns a copy of the client-level settings this vault uses.
func (b *Base) ClientSettings() ClientSettings {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return cloneClientSettings(b.client)
}

// SetClientSettings replaces the client-level settings. A change of client credentials
// drops the cached token.
func (b *Base) SetClientSettings(client ClientSettings) {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	if !sameCredentials(client.Credentials, b.client.Credentials) {
		b.authVersion++
	}
	b.client = cloneClientSettings(client)
}

// ResetAuth drops the cached bearer token so the next call re-authenticates.
func (b *Base) ResetAuth() {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	b.authVersion++
}

// sameCredentials reports whether two client credentials hold the same values.
func sameCredentials(a, b *config.Credentials) bool {
	if a == nil || b == nil {
		return a == b
	}
	return reflect.DeepEqual(*a, *b)
}

func cloneCredentials(creds config.Credentials) config.Credentials {
	creds.Roles = slices.Clone(creds.Roles)
	return creds
}

func cloneRetries(retries *int) *int {
	if retries == nil {
		return nil
	}
	n := *retries
	return &n
}

func cloneVaultConfig(vaultConfig config.VaultConfig) config.VaultConfig {
	vaultConfig.Credentials = cloneCredentials(vaultConfig.Credentials)
	vaultConfig.MaxRetries = cloneRetries(vaultConfig.MaxRetries)
	return vaultConfig
}

func cloneClientSettings(client ClientSettings) ClientSettings {
	if client.Credentials != nil {
		creds := cloneCredentials(*client.Credentials)
		client.Credentials = &creds
	}
	client.HttpConfig.MaxRetries = cloneRetries(client.HttpConfig.MaxRetries)
	client.Headers = maps.Clone(client.Headers)
	return client
}

func (b *Base) snapshot() settings {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return settings{config: b.config, client: b.client, authVersion: b.authVersion}
}

// Prepare resolves auth, headers and HTTP settings for one call. headers are the call's own
// custom headers, added after the client-level ones; operation names the call in a header
// validation error.
func (b *Base) Prepare(operation string, headers map[config.CustomHeaderKey]string) (*Call, *skyflowError.SkyflowError) {
	return b.PrepareContext(context.Background(), operation, headers)
}

// PrepareContext is Prepare that stops waiting for a bearer token when ctx ends.
func (b *Base) PrepareContext(ctx context.Context, operation string, callHeaders map[config.CustomHeaderKey]string) (*Call, *skyflowError.SkyflowError) {
	current := b.snapshot()
	headers, err := b.buildHeaders(operation, current.client.Headers, callHeaders)
	if err != nil {
		return nil, err
	}
	authToken, err := b.authToken(ctx, current)
	if err != nil {
		return nil, err
	}
	httpConfig := ResolveHttpConfig(current.config.HttpConfig(), current.client.HttpConfig, b.profile.Defaults)
	return &Call{
		VaultID:    current.config.VaultId,
		BaseURL:    b.baseURL(current.config),
		AuthToken:  authToken,
		Headers:    headers,
		HTTPClient: b.newCallClient(httpConfig),
	}, nil
}

// RefreshAuth returns a copy of call with the auth token and base URL the vault's settings
// give now: the cached token while it is valid, else a new one. Both come from one snapshot,
// so a token is never sent to an endpoint other than the one configured with it. A call that
// runs for long, such as polling a run, uses it so its token neither expires nor outlives a
// settings change. It stops waiting for a token when ctx ends.
func (b *Base) RefreshAuth(ctx context.Context, call *Call) (*Call, *skyflowError.SkyflowError) {
	current := b.snapshot()
	authToken, err := b.authToken(ctx, current)
	if err != nil {
		return nil, err
	}
	refreshed := *call
	refreshed.AuthToken = authToken
	refreshed.BaseURL = b.baseURL(current.config)
	return &refreshed, nil
}

// GenerateToken returns the token a set of credentials authenticates with: the token or
// API key as given, or a bearer token generated from service-account credentials.
func (p Profile) GenerateToken(creds config.Credentials) (string, *skyflowError.SkyflowError) {
	switch {
	case creds.Token != "":
		return creds.Token, nil
	case creds.ApiKey != "":
		return creds.ApiKey, nil
	}
	opts := sa.BearerTokenOptions{RoleIds: creds.Roles, Ctx: creds.Context}
	switch {
	case creds.Path != "":
		resp, err := p.Auth.GenerateBearerToken(creds.Path, opts)
		if err != nil {
			return "", err
		}
		return resp.AccessToken, nil
	case creds.CredentialsString != "":
		resp, err := p.Auth.GenerateBearerTokenFromCreds(creds.CredentialsString, opts)
		if err != nil {
			return "", err
		}
		return resp.AccessToken, nil
	}
	return "", p.NewError(skyflowError.INVALID_INPUT_CODE, p.emptyCredentialsMsg())
}

// tokenRefreshMargin is how long before it expires a cached bearer token is replaced, so it
// does not expire during a call and clock skew with the server does not make it fail early.
const tokenRefreshMargin = 2 * time.Minute

// tokenResult is the outcome of generating a bearer token.
type tokenResult struct {
	token string
	err   *skyflowError.SkyflowError
}

// authToken returns the API key or bearer token to send, generating and caching a bearer
// token from service-account credentials when needed. When ctx ends first it returns an
// error wrapping ctx.Err(); a generation already started still finishes and is cached.
func (b *Base) authToken(ctx context.Context, current settings) (string, *skyflowError.SkyflowError) {
	creds, err := b.resolveCredentials(current)
	if err != nil {
		return "", err
	}
	if creds.ApiKey != "" {
		logOptional(logger.Info, b.profile.Logs.UsingApiKey)
		return creds.ApiKey, nil
	}
	if creds.Token != "" {
		if b.profile.Auth.IsExpired(creds.Token) {
			logger.Error(b.profile.logLine(b.profile.Logs.TokenExpired, logBearerTokenExpired))
			return "", b.profile.NewError(skyflowError.INVALID_INPUT_CODE, b.profile.tokenExpiredMsg())
		}
		logOptional(logger.Info, b.profile.Logs.UsingBearerToken)
		return creds.Token, nil
	}
	// Checked first: select picks at random when the lock is also free, so a call whose
	// context already ended could otherwise start a token request.
	if ctx.Err() != nil {
		return "", b.cancelled(ctx)
	}
	select {
	case b.tokenLock <- struct{}{}:
	case <-ctx.Done():
		return "", b.cancelled(ctx)
	}
	if b.token != "" && b.tokenVersion == current.authVersion && !b.profile.Auth.ExpiresWithin(b.token, tokenRefreshMargin) {
		token := b.token
		<-b.tokenLock
		logger.Info(b.profile.logLine(b.profile.Logs.ReuseBearerToken, logReuseBearerToken))
		return token, nil
	}
	logger.Info(b.profile.logLine(b.profile.Logs.GenerateBearerToken, logGenerateBearerTokenTriggered))
	result := make(chan tokenResult, 1)
	go func() {
		defer func() { <-b.tokenLock }()
		token, err := b.profile.GenerateToken(*creds)
		// A call that started before a credentials change must not replace a newer token.
		if err == nil && current.authVersion >= b.tokenVersion {
			b.storeToken(token, current.authVersion)
		}
		result <- tokenResult{token: token, err: err}
	}()
	select {
	case r := <-result:
		return r.token, r.err
	case <-ctx.Done():
		return "", b.cancelled(ctx)
	}
}

// cancelled is the error for a call whose context ended while it waited for a bearer token.
func (b *Base) cancelled(ctx context.Context) *skyflowError.SkyflowError {
	logger.Error(b.profile.logMsg(logBearerTokenCancelled))
	return b.profile.NewErrorWithCause(skyflowError.REQUEST_CANCELLED, b.profile.errMsg(errBearerTokenCancelled), ctx.Err())
}

// resolveCredentials picks vault credentials, then client credentials, then SKYFLOW_CREDENTIALS
// from the environment or the .env file.
func (b *Base) resolveCredentials(current settings) (*config.Credentials, *skyflowError.SkyflowError) {
	if !config.IsCredentialsEmpty(current.config.Credentials) {
		creds := current.config.Credentials
		return &creds, nil
	}
	if current.client.Credentials != nil && !config.IsCredentialsEmpty(*current.client.Credentials) {
		return current.client.Credentials, nil
	}
	if envCreds := EnvCredentials(); envCreds != "" {
		return &config.Credentials{CredentialsString: envCreds}, nil
	}
	logger.Error(b.profile.logMsg(logEmptyCredentials))
	return nil, b.profile.NewError(skyflowError.INVALID_INPUT_CODE, b.profile.emptyCredentialsMsg())
}

// baseURL picks BaseVaultUrl, then the deprecated BaseVaultURL, then the cluster URL.
func (b *Base) baseURL(vaultConfig config.VaultConfig) string {
	if vaultConfig.BaseVaultURL != "" { //nolint:staticcheck // the deprecated field is still honoured
		logOptional(logger.Warn, b.profile.Logs.DeprecatedBaseVaultURL)
	}
	if vaultConfig.BaseVaultUrl != "" {
		return vaultConfig.BaseVaultUrl
	}
	if vaultConfig.BaseVaultURL != "" { //nolint:staticcheck // the deprecated field is still honoured
		return vaultConfig.BaseVaultURL //nolint:staticcheck // the deprecated field is still honoured
	}
	return b.profile.ClusterURL(vaultConfig.Env, vaultConfig.ClusterId)
}

// logOptional logs line when the SDK set one.
func logOptional(log func(...interface{}), line string) {
	if line != "" {
		log(line)
	}
}
