package client

import (
	"fmt"
	"skyflow-go/v2/common/logger"
	"skyflow-go/v2/vault/controller"
	"skyflow-go/v2/vault/utils"
)

type SkyflowClient struct {
	vaultConfigs     map[string]utils.VaultConfig      // Vault ID -> VaultConfig
	connectionConfig map[string]utils.ConnectionConfig // Connection ID -> ConnectionConfig
	credentials      utils.Credentials
	logLevel         logger.LogLevel
}

const defaultLogLevel logger.LogLevel = 2

type SkyflowClientBuilder struct {
	client SkyflowClient
}

func (c *SkyflowClient) Builder() *SkyflowClientBuilder {
	return &SkyflowClientBuilder{
		client: SkyflowClient{
			vaultConfigs:     make(map[string]utils.VaultConfig),
			connectionConfig: make(map[string]utils.ConnectionConfig),
			logLevel:         defaultLogLevel,
		},
	}
}

func (b *SkyflowClientBuilder) WithVaultConfig(vaultConfig utils.VaultConfig) *SkyflowClientBuilder {
	b.client.vaultConfigs[vaultConfig.VaultId] = vaultConfig
	return b
}

func (b *SkyflowClientBuilder) WithConnectionConfig(connConfig utils.ConnectionConfig) *SkyflowClientBuilder {
	b.client.connectionConfig[connConfig.ConnectionId] = connConfig
	return b
}

func (b *SkyflowClientBuilder) WithSkyflowCredentials(credentials utils.Credentials) *SkyflowClientBuilder {
	b.client.credentials = credentials
	return b
}

func (b *SkyflowClientBuilder) WithLogLevel(logLevel logger.LogLevel) *SkyflowClientBuilder {
	b.client.logLevel = logLevel
	return b
}

func (b *SkyflowClientBuilder) Build() (*SkyflowClient, error) {
	return &b.client, nil
}

func (c *SkyflowClient) Vault(vaultID string) (*controller.VaultService, error) {
	config, exists := c.vaultConfigs[vaultID]
	if !exists {
		return nil, fmt.Errorf("vault ID %s not found", vaultID) // throw skyflow error
	}

	return &controller.VaultService{
		Config:   config,
		Loglevel: c.logLevel,
	}, nil
}

// update client level props methods
func (c *SkyflowClient) UpdateVaultConfig(updatedConfig utils.VaultConfig) error {
	if _, exists := c.vaultConfigs[updatedConfig.VaultId]; !exists {
		return fmt.Errorf("vault ID %s not found", updatedConfig.VaultId) // throw skyflow error
	}

	c.vaultConfigs[updatedConfig.VaultId] = updatedConfig // Update the vault configuration
	return nil
}
func (c *SkyflowClient) UpdateConnectionConfig(updatedConfig utils.ConnectionConfig) error {
	if _, exists := c.connectionConfig[updatedConfig.ConnectionId]; !exists {
		return fmt.Errorf("connection ID %s not found", updatedConfig.ConnectionId) // throw skyflow error
	}

	c.connectionConfig[updatedConfig.ConnectionId] = updatedConfig // Update the connection configuration
	return nil
}
func (c *SkyflowClient) UpdateLogLevel(logLevel logger.LogLevel) {
	c.logLevel = logLevel
}

func (c *SkyflowClient) UpdateCredentials(credentials utils.Credentials) {
	c.credentials = credentials
}
