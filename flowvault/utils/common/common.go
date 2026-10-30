package common

import (
	cfg "github.com/skyflowapi/skyflow-go/common/config"
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
	Credentials      = cfg.Credentials
	VaultConfig      = cfg.VaultConfig
	HttpConfig       = cfg.HttpConfig
	ConnectionConfig = cfg.ConnectionConfig
	CustomHeaderKey  = cfg.CustomHeaderKey
)

const (
	SkyflowAccountId   CustomHeaderKey = cfg.SkyflowAccountId
	SkyflowAccountName CustomHeaderKey = cfg.SkyflowAccountName
	RequestIdHeader    CustomHeaderKey = cfg.RequestIdHeader
)

type (
	TokenResponse            = sa.TokenResponse
	BearerTokenOptions       = sa.BearerTokenOptions
	SignedDataTokensOptions  = sa.SignedDataTokensOptions
	SignedDataTokensResponse = sa.SignedDataTokensResponse
)

// IsCredentialsEmpty reports whether no credential source is set.
func IsCredentialsEmpty(c Credentials) bool { return cfg.IsCredentialsEmpty(c) }
