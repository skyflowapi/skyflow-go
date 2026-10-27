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
	"net/http"
	"os"
	"sync"

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
	// SDKPrefix is prepended to error messages, e.g. "Skyflow FlowVault Go SDK v0.1.0".
	SDKPrefix string
	// LogPrefix is prepended to log lines, e.g. "[ Skyflow FlowVault Go SDK v0.1.0 ] ".
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
	// Logs, when set, replace or add log lines.
	Logs Logs
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
func (p Profile) logMsg(base string) string { return p.LogPrefix + base }

// logLine returns the SDK's own line when set, else the default with the log prefix.
func (p Profile) logLine(own, base string) string {
	if own != "" {
		return own
	}
	return p.logMsg(base)
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
	// Headers are sent with every call, before any the call's interceptor adds.
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
// change them (SetConfig, SetClientSettings) while calls are running.
type Base struct {
	profile Profile

	settingsMu sync.RWMutex
	config     config.VaultConfig
	client     ClientSettings
	// authVersion changes whenever the credentials may have changed, so a token generated
	// from older credentials is never reused.
	authVersion uint64

	tokenMu      sync.Mutex
	token        string
	tokenVersion uint64
}

// settings is a consistent copy of the Base's settings for one call.
type settings struct {
	config      config.VaultConfig
	client      ClientSettings
	authVersion uint64
}

// New returns a Base for one vault.
func New(profile Profile, vaultConfig config.VaultConfig, client ClientSettings) *Base {
	return &Base{profile: profile, config: vaultConfig, client: client}
}

// Profile returns the SDK profile this Base works for.
func (b *Base) Profile() Profile {
	return b.profile
}

// Config returns a copy of the vault config.
func (b *Base) Config() config.VaultConfig {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return b.config
}

// SetConfig replaces the vault config. Set resetAuth when the credentials changed.
func (b *Base) SetConfig(vaultConfig config.VaultConfig, resetAuth bool) {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	b.config = vaultConfig
	if resetAuth {
		b.authVersion++
	}
}

// ClientSettings returns the client-level settings this vault uses.
func (b *Base) ClientSettings() ClientSettings {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return b.client
}

// SetClientSettings replaces the client-level settings. A change of client credentials
// drops the cached token.
func (b *Base) SetClientSettings(client ClientSettings) {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	if client.Credentials != b.client.Credentials {
		b.authVersion++
	}
	b.client = client
}

// ResetAuth drops the cached bearer token so the next call re-authenticates.
func (b *Base) ResetAuth() {
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()
	b.authVersion++
}

func (b *Base) snapshot() settings {
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return settings{config: b.config, client: b.client, authVersion: b.authVersion}
}

// Prepare resolves auth, headers and HTTP settings for one call. The interceptor, when set,
// can add per-call headers; operation is passed to it.
func (b *Base) Prepare(operation string, interceptor Interceptor) (*Call, *skyflowError.SkyflowError) {
	current := b.snapshot()
	headers, err := b.buildHeaders(operation, current.client.Headers, interceptor)
	if err != nil {
		return nil, err
	}
	authToken, err := b.authToken(current)
	if err != nil {
		return nil, err
	}
	httpConfig := ResolveHttpConfig(current.config.HttpConfig(), current.client.HttpConfig, b.profile.Defaults)
	return &Call{
		VaultID:    current.config.VaultId,
		BaseURL:    b.baseURL(current.config),
		AuthToken:  authToken,
		Headers:    headers,
		HTTPClient: NewHTTPClient(httpConfig),
	}, nil
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
	return "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, p.errMsg(errEmptyCredentials))
}

// authToken returns the API key or bearer token to send, generating and caching a bearer
// token from service-account credentials when needed.
func (b *Base) authToken(current settings) (string, *skyflowError.SkyflowError) {
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
			return "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, b.profile.tokenExpiredMsg())
		}
		logOptional(logger.Info, b.profile.Logs.UsingBearerToken)
		return creds.Token, nil
	}
	b.tokenMu.Lock()
	defer b.tokenMu.Unlock()
	if b.token != "" && b.tokenVersion == current.authVersion && !b.profile.Auth.IsExpired(b.token) {
		logger.Info(b.profile.logLine(b.profile.Logs.ReuseBearerToken, logReuseBearerToken))
		return b.token, nil
	}
	logger.Info(b.profile.logLine(b.profile.Logs.GenerateBearerToken, logGenerateBearerTokenTriggered))
	token, err := b.profile.GenerateToken(*creds)
	if err != nil {
		return "", err
	}
	// A call that started before a credentials change must not replace a newer token.
	if current.authVersion >= b.tokenVersion {
		b.token, b.tokenVersion = token, current.authVersion
	}
	return token, nil
}

// resolveCredentials picks vault credentials, then client credentials, then SKYFLOW_CREDENTIALS.
func (b *Base) resolveCredentials(current settings) (*config.Credentials, *skyflowError.SkyflowError) {
	if !config.IsCredentialsEmpty(current.config.Credentials) {
		creds := current.config.Credentials
		return &creds, nil
	}
	if current.client.Credentials != nil && !config.IsCredentialsEmpty(*current.client.Credentials) {
		return current.client.Credentials, nil
	}
	if envCreds := os.Getenv(CredentialsEnv); envCreds != "" {
		return &config.Credentials{CredentialsString: envCreds}, nil
	}
	logger.Error(b.profile.logMsg(logEmptyCredentials))
	return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, b.profile.errMsg(errEmptyCredentials))
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
