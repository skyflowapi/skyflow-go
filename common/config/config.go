// Package config holds the platform-level configuration types shared by every
// Skyflow Go SDK: environment, credentials, vault and connection configs, and
// custom header keys. Each SDK re-exports them under its own utils/common
// package through type aliases, so they are the same types across SDKs.
//
// Request, option and response types are deliberately NOT here: each SDK owns
// its own, because the vault APIs differ between products.
package config

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
	VaultId      string
	ClusterId    string
	BaseVaultUrl string
	// Deprecated: Use BaseVaultUrl instead.
	BaseVaultURL string
	Env          Env
	Credentials  Credentials
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
