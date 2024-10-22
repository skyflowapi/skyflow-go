package client

import (
	"fmt"
	"os"
	vaultutils "skyflow-go/v2/utils/common"
	"skyflow-go/v2/utils/logger"
	_ "skyflow-go/v2/vault/controller"
)

type Skyflow struct {
	builder *SkyflowBuilder
}

const defaultLogLevel logger.LogLevel = 2

type SkyflowBuilder struct {
	vaultConfigs     map[string]vaultutils.VaultConfig
	vaultControllers map[string]*vaultService
	connectionConfig map[string]vaultutils.ConnectionConfig
	credentials      vaultutils.Credentials
	logLevel         logger.LogLevel
}

func (c *Skyflow) Builder() *SkyflowBuilder {
	return &SkyflowBuilder{
		vaultConfigs:     make(map[string]vaultutils.VaultConfig),
		vaultControllers: make(map[string]*vaultService),
		connectionConfig: make(map[string]vaultutils.ConnectionConfig),
		logLevel:         defaultLogLevel,
	}
}

func (b *SkyflowBuilder) WithVaultConfig(vaultConfig vaultutils.VaultConfig) *SkyflowBuilder {
	if _, exists := b.vaultConfigs[vaultConfig.VaultId]; exists {
		return b
	}
	b.vaultConfigs[vaultConfig.VaultId] = vaultConfig
	return b
}

func (b *SkyflowBuilder) WithConnectionConfig(connConfig vaultutils.ConnectionConfig) *SkyflowBuilder {
	if _, exists := b.connectionConfig[connConfig.ConnectionId]; exists {
		return b
	}

	b.connectionConfig[connConfig.ConnectionId] = connConfig
	return b
}

func (b *SkyflowBuilder) WithSkyflowCredentials(credentials vaultutils.Credentials) *SkyflowBuilder {
	b.credentials = credentials
	return b
}

func (b *SkyflowBuilder) WithLogLevel(logLevel logger.LogLevel) *SkyflowBuilder {
	b.logLevel = logLevel
	return b
}

func (b *SkyflowBuilder) Build() (*Skyflow, error) {
	for vaultID, vaultConfig := range b.vaultConfigs {
		if _, exists := b.vaultControllers[vaultID]; !exists {
			b.vaultControllers[vaultID] = &vaultService{
				config:   vaultConfig,
				logLevel: &b.logLevel,
			}
		}
	}
	return &Skyflow{
		builder: b,
	}, nil
}

// Vault vault method
func (c *Skyflow) Vault(vaultID ...string) (*vaultService, error) {
	// get vault config if available in vault configs, skyflow or env
	config, err := getVaultConfig(c.builder, vaultID...)
	if err != nil {
		return nil, err
	}

	err = setVaultCredentials(&config, c.builder.credentials)
	if err != nil {
		return nil, err
	}

	// Get the VaultController from the builder's VaultControllers map
	vaultController, exists := c.builder.vaultControllers[config.VaultId]
	if !exists {
		return nil, fmt.Errorf("vault service not found for vault ID %s", config.VaultId)
	}
	// Update the config in the vault service
	vaultController.config = config

	return vaultController, nil
}

// vaultutils or helper func will move later
func isCredentialsEmpty(creds vaultutils.Credentials) bool {
	return creds.Path == "" &&
		creds.CredentialsString == "" &&
		creds.Token == ""
}

func getVaultConfig(builder *SkyflowBuilder, vaultID ...string) (vaultutils.VaultConfig, error) {
	// if vault configs are empty
	if len(builder.vaultConfigs) == 0 {
		return vaultutils.VaultConfig{}, fmt.Errorf("no vault configurations available")
	}

	// if vault is passed
	if len(vaultID) > 0 && len(builder.vaultConfigs) > 0 {
		config, exists := builder.vaultConfigs[vaultID[0]]
		if !exists {
			return vaultutils.VaultConfig{}, fmt.Errorf("vault ID %s not found", vaultID[0])
		}
		return config, nil
	}

	// No vault ID passed, return the first vault config available
	for _, cfg := range builder.vaultConfigs {
		return cfg, nil
	}

	return vaultutils.VaultConfig{}, fmt.Errorf("no vault configuration found")
}
func setVaultCredentials(config *vaultutils.VaultConfig, builderCreds vaultutils.Credentials) error {
	// here if credentials are empty in the vault config
	if isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if !isCredentialsEmpty(builderCreds) {
			config.Credentials = builderCreds
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return fmt.Errorf("no credentials available")
		}
	}
	return nil
}

// UpdateLogLevel update methods
func (c *Skyflow) UpdateLogLevel(logLevel logger.LogLevel) {
	c.builder.logLevel = logLevel
}

func (c *Skyflow) UpdateCredentials(credentials vaultutils.Credentials) {
	c.builder.credentials = credentials
}
func (c *Skyflow) UpdateVaultConfig(updatedConfig vaultutils.VaultConfig) error {
	if _, exists := c.builder.vaultConfigs[updatedConfig.VaultId]; !exists {
		return fmt.Errorf("vault ID %s not found", updatedConfig.VaultId)
	}

	c.builder.vaultConfigs[updatedConfig.VaultId] = updatedConfig
	c.builder.vaultControllers[updatedConfig.VaultId].config = updatedConfig
	return nil
}

func (c *Skyflow) UpdateConnectionConfig(updatedConfig vaultutils.ConnectionConfig) error {
	if _, exists := c.builder.connectionConfig[updatedConfig.ConnectionId]; !exists {
		return fmt.Errorf("connection ID %s not found", updatedConfig.ConnectionId)
	}

	c.builder.connectionConfig[updatedConfig.ConnectionId] = updatedConfig
	return nil
}
