// Package config holds the platform-level configuration types shared by every
// Skyflow Go SDK: environment, credentials, vault and connection configs, and
// custom header keys. Each SDK re-exports them under its own utils/common
// package through type aliases, so they are the same types across SDKs.
//
// Request, option and response types are deliberately NOT here: each SDK owns
// its own, because the vault APIs differ between products.
package config

import "time"

// Env is the Skyflow environment a vault runs in. The zero value is PROD, the default.
type Env int

const (
	PROD Env = iota
	STAGE
	SANDBOX
	DEV
)

type Credentials struct {
	Path              string
	Roles             []string
	Context           interface{}
	CredentialsString string
	Token             string
	ApiKey            string
}

type VaultConfig struct {
	VaultId   string
	ClusterId string
	// VaultUrl overrides the URL built from ClusterId and Env. It must use https; http
	// is accepted only for localhost and loopback addresses.
	VaultUrl    string
	Env         Env
	Credentials Credentials

	// Optional HTTP settings for this vault. Zero values (and a nil MaxRetries)
	// mean "not set": the client-level HttpConfig applies, then the SDK default.
	// See HttpConfig for what each one bounds.
	Timeout           time.Duration
	ConnectTimeout    time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxRetries        *int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
}

// HttpConfig holds client-level call timeout and retry settings. A value set on
// VaultConfig wins over the one here.
type HttpConfig struct {
	// Timeout bounds the whole call, retries included.
	Timeout time.Duration
	// ConnectTimeout bounds opening a connection, per attempt.
	ConnectTimeout time.Duration
	// ReadTimeout bounds the wait for the server to start responding once the request is
	// sent, per attempt.
	ReadTimeout time.Duration
	// WriteTimeout bounds each write of the request to the connection, per attempt.
	WriteTimeout      time.Duration
	MaxRetries        *int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
}

// HttpConfig returns the vault's own HTTP settings.
func (c VaultConfig) HttpConfig() HttpConfig {
	return HttpConfig{
		Timeout:           c.Timeout,
		ConnectTimeout:    c.ConnectTimeout,
		ReadTimeout:       c.ReadTimeout,
		WriteTimeout:      c.WriteTimeout,
		MaxRetries:        c.MaxRetries,
		InitialRetryDelay: c.InitialRetryDelay,
		MaxRetryDelay:     c.MaxRetryDelay,
	}
}

type ConnectionConfig struct {
	ConnectionId  string
	ConnectionUrl string
	Credentials   Credentials
}

type CustomHeaderKey string

const (
	SkyflowAccountId   CustomHeaderKey = "x-skyflow-account-id"
	SkyflowAccountName CustomHeaderKey = "x-skyflow-account-name"
	RequestIdHeader    CustomHeaderKey = "x-request-id"

	// Deprecated: Use SkyflowAccountId instead.
	SkyflowAccountID = SkyflowAccountId
	// Deprecated: Use RequestIdHeader instead.
	RequestIDHeader = RequestIdHeader
)
