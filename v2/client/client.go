package client

import (
	"skyflow-go/v2/common/logger"
	"skyflow-go/v2/vault/controller"
	"skyflow-go/v2/vault/utils"
)

type SkyflowClient struct {
	vaultConfigs     map[string]utils.VaultConfig // Vault ID -> VaultConfig
	connectionConfig utils.ConnectionConfig
	credentials      utils.Credentials
	logLevel         logger.LogLevel
}

// Default log level if not specified
const defaultLogLevel logger.LogLevel = 2

type Skyflow struct {
	client SkyflowClient
}

func (b *Skyflow) WithVaultConfig(vaultConfig utils.VaultConfig) *Skyflow {
	if b.client.vaultConfigs == nil {
		b.client.vaultConfigs = make(map[string]utils.VaultConfig)
	}
	b.client.vaultConfigs[vaultConfig.ID] = vaultConfig
	return b
}

func (b *Skyflow) WithConnectionConfig(connConfig utils.ConnectionConfig) *Skyflow {
	b.client.connectionConfig = connConfig
	return b
}

func (b *Skyflow) WithSkyflowCredentials(credentials utils.Credentials) *Skyflow {
	b.client.credentials = credentials
	return b
}

func (b *Skyflow) WithLogLevel(logLevel logger.LogLevel) *Skyflow {
	b.client.logLevel = logLevel
	return b
}

func (b *Skyflow) Build() (*SkyflowClient, error) {
	return &b.client, nil
}

func (c *SkyflowClient) Vault(vaultID string) (*controller.VaultService, error) {
	config, _ := c.vaultConfigs[vaultID]

	return &controller.VaultService{
		Config: config,
	}, nil
}
