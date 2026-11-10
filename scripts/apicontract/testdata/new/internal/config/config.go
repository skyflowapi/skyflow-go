// Package config holds the type that vault re-exports as an alias, so the fixture covers
// a type moved out of the public package.
package config

// VaultConfig identifies a vault.
type VaultConfig struct {
	VaultId   string
	clusterId string
}

// GetVaultId returns the vault ID.
func (c VaultConfig) GetVaultId() string { return c.VaultId }

// SetVaultId sets the vault ID.
func (c *VaultConfig) SetVaultId(vaultId string) { c.VaultId = vaultId }
