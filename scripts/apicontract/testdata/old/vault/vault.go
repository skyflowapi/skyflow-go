// Package vault is a test fixture for the apicontract tool: a public API with one of
// each declaration kind the snapshot records. This is the version before VaultConfig
// moved to internal/config.
package vault

// VaultConfig identifies a vault.
type VaultConfig struct {
	VaultId   string
	clusterId string
}

// GetVaultId returns the vault ID.
func (c VaultConfig) GetVaultId() string { return c.VaultId }

// SetVaultId sets the vault ID.
func (c *VaultConfig) SetVaultId(vaultId string) { c.VaultId = vaultId }

// NewVaultConfig builds a VaultConfig.
func NewVaultConfig(vaultId string, roleIds ...string) (*VaultConfig, error) {
	return &VaultConfig{VaultId: vaultId}, nil
}

// Validator is implemented by callers.
type Validator interface{ Validate(VaultConfig) error }

// SdkName is an exported constant.
const SdkName = "skyflow-go"

// LogLevel is a named string.
type LogLevel string

// TokenGroupName is an alias of a basic type.
type TokenGroupName = string

// DefaultVaultConfig is a package variable.
var DefaultVaultConfig VaultConfig

// Response is generic.
type Response[T any] struct{ Records []T }

// Credentials can't be implemented outside the module.
type Credentials interface {
	GetToken() string
	isCredentials()
}

// Numeric is a type-set constraint.
type Numeric interface{ ~int | ~float64 }
