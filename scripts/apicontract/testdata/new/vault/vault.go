// Package vault is a test fixture for the apicontract tool: the same public API as
// testdata/old, with VaultConfig moved to internal/config behind an alias. Its snapshot
// must match the old one.
package vault

import "github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/internal/config"

// VaultConfig keeps the old name for the moved type.
type VaultConfig = config.VaultConfig

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
