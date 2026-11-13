package common

import (
	cfg "github.com/skyflowapi/skyflow-go/common/config"
	"github.com/skyflowapi/skyflow-go/common/logger"
	sa "github.com/skyflowapi/skyflow-go/common/serviceaccount"
)

// ---------------------------------------------------------------------------
// Platform-level types shared with the other Skyflow Go SDKs (same types, via
// aliases): environment, credentials, configs, custom headers, service account.
// ---------------------------------------------------------------------------

type Env = cfg.Env

const (
	PROD    Env = cfg.PROD
	STAGE   Env = cfg.STAGE
	SANDBOX Env = cfg.SANDBOX
	DEV     Env = cfg.DEV
)

type (
	Credentials     = cfg.Credentials
	VaultConfig     = cfg.VaultConfig
	HttpConfig      = cfg.HttpConfig
	CustomHeaderKey = cfg.CustomHeaderKey
)

const (
	SkyflowAccountId   CustomHeaderKey = cfg.SkyflowAccountId
	SkyflowAccountName CustomHeaderKey = cfg.SkyflowAccountName
	RequestIdHeader    CustomHeaderKey = cfg.RequestIdHeader
)

type (
	TokenResponse            = sa.TokenResponse
	SignedDataTokensOptions  = sa.SignedDataTokensOptions
	SignedDataTokensResponse = sa.SignedDataTokensResponse
)

// BearerTokenOptions customise bearer token generation.
type BearerTokenOptions struct {
	// Ctx is the context to embed in the token: a string, or a map for a context object.
	Ctx interface{}
	// RoleIds scopes the token to these roles.
	RoleIds  []string
	LogLevel logger.LogLevel
}

// IsCredentialsEmpty reports whether no credential source is set.
func IsCredentialsEmpty(c Credentials) bool { return cfg.IsCredentialsEmpty(c) }
