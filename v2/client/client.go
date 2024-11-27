package client

import (
	"fmt"
	"os"
	vaultutils "skyflow-go/v2/utils/common"
	errors "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
	"skyflow-go/v2/vault/controller"
)

type Skyflow struct {
	builder *SkyflowBuilder
}

const defaultLogLevel logger.LogLevel = 2

type SkyflowBuilder struct {
	vaultConfigs       map[string]vaultutils.VaultConfig
	vaultServices      map[string]*vaultService
	connectionServices map[string]*connectionService
	connectionConfig   map[string]vaultutils.ConnectionConfig
	credentials        vaultutils.Credentials
	logLevel           logger.LogLevel
}

func (c *Skyflow) Builder() *SkyflowBuilder {
	return &SkyflowBuilder{
		vaultConfigs:       make(map[string]vaultutils.VaultConfig),
		vaultServices:      make(map[string]*vaultService),
		connectionServices: make(map[string]*connectionService),
		connectionConfig:   make(map[string]vaultutils.ConnectionConfig),
		logLevel:           defaultLogLevel,
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
		if _, exists := b.vaultServices[vaultID]; !exists {
			b.vaultServices[vaultID] = &vaultService{
				config:   vaultConfig,
				logLevel: &b.logLevel,
			}
		}
	}
	for connectionId, connectionConfig := range b.connectionConfig {
		if _, exists := b.vaultServices[connectionId]; !exists {
			b.connectionServices[connectionId] = &connectionService{
				config:   connectionConfig,
				logLevel: &b.logLevel,
			}
		}
	}

	return &Skyflow{
		builder: b,
	}, nil
}

// Vault vaultapi method
func (c *Skyflow) Vault(vaultID ...string) (*vaultService, *errors.SkyflowError) {
	// get vaultapi config if available in vaultapi configs, skyflow or env
	config, err := getVaultConfig(c.builder, vaultID...)
	if err != nil {
		return nil, err
	}

	err = setVaultCredentials(&config, c.builder.credentials)
	if err != nil {
		return nil, err
	}

	// Get the VaultController from the builder's vaultServices map
	vaultService, exists := c.builder.vaultServices[config.VaultId]
	if !exists {
		return nil, errors.NewSkyflowError("400", fmt.Sprintf("vaultapi service not found for vaultapi ID %s", config.VaultId))
	}
	// Update the config in the vaultapi service
	vaultService.controller = controller.VaultController{
		Config:   config,
		Loglevel: &c.builder.logLevel,
	}
	vaultService.config = config
	return vaultService, nil
}
func (c *Skyflow) Connection(connectionId ...string) (*connectionService, *errors.SkyflowError) {
	config, err := getConnectionConfig(c.builder, connectionId...)
	if err != nil {
		return nil, err
	}
	err = setConnectionCredentials(&config, c.builder.credentials)
	if err != nil {
		return nil, err
	}
	connectionService, exists := c.builder.connectionServices[config.ConnectionId]
	if !exists {
		return nil, errors.NewSkyflowError("400", fmt.Sprintf("connection service not found for connectionid %s", config.ConnectionId))
	}
	connectionService.controller = controller.ConnectionController{
		Config:   config,
		Loglevel: &c.builder.logLevel,
	}
	connectionService.config = config
	return connectionService, nil
}

// vaultutils or helper func
func isCredentialsEmpty(creds vaultutils.Credentials) bool {
	return creds.Path == "" &&
		creds.CredentialsString == "" &&
		creds.Token == ""
}
func getConnectionConfig(builder *SkyflowBuilder, connectionId ...string) (vaultutils.ConnectionConfig, *errors.SkyflowError) {
	// if connection configs are empty
	if len(builder.connectionConfig) == 0 {
		return vaultutils.ConnectionConfig{}, errors.NewSkyflowError("400", "no connection configurations available")
	}

	// if connection id is passed
	if len(connectionId) > 0 && len(builder.connectionConfig) > 0 {
		config, exists := builder.connectionConfig[connectionId[0]]
		if !exists {
			return vaultutils.ConnectionConfig{}, errors.NewSkyflowError("400", "connection ID %s not found")
		}
		return config, nil
	}

	// No conenction ID passed, return the first config available
	for _, cfg := range builder.connectionConfig {
		return cfg, nil
	}

	return vaultutils.ConnectionConfig{}, errors.NewSkyflowError("400", "no connection configuration found")
}
func getVaultConfig(builder *SkyflowBuilder, vaultID ...string) (vaultutils.VaultConfig, *errors.SkyflowError) {
	// if vaultapi configs are empty
	if len(builder.vaultConfigs) == 0 {
		return vaultutils.VaultConfig{}, errors.NewSkyflowError("400", "no vaultapi configurations available")
	}

	// if vaultapi is passed
	if len(vaultID) > 0 && len(builder.vaultConfigs) > 0 {
		config, exists := builder.vaultConfigs[vaultID[0]]
		if !exists {
			return vaultutils.VaultConfig{}, errors.NewSkyflowError("400", "vaultapi ID %s not found")
		}
		return config, nil
	}

	// No vaultapi ID passed, return the first vaultapi config available
	for _, cfg := range builder.vaultConfigs {
		return cfg, nil
	}

	return vaultutils.VaultConfig{}, errors.NewSkyflowError("400", "no vaultapi configuration found")
}
func setVaultCredentials(config *vaultutils.VaultConfig, builderCreds vaultutils.Credentials) *errors.SkyflowError {
	// here if credentials are empty in the vaultapi config
	if isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if !isCredentialsEmpty(builderCreds) {
			config.Credentials = builderCreds
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return errors.NewSkyflowError("400", "no credentials available")
		}
	}
	return nil
}
func setConnectionCredentials(config *vaultutils.ConnectionConfig, builderCreds vaultutils.Credentials) *errors.SkyflowError {
	// here if credentials are empty in the vaultapi config
	if isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if !isCredentialsEmpty(builderCreds) {
			config.Credentials = builderCreds
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return errors.NewSkyflowError("400", "no credentials available")
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
		return fmt.Errorf("vaultapi ID %s not found", updatedConfig.VaultId)
	}

	c.builder.vaultConfigs[updatedConfig.VaultId] = updatedConfig
	c.builder.vaultServices[updatedConfig.VaultId].config = updatedConfig
	return nil
}

func (c *Skyflow) UpdateConnectionConfig(updatedConfig vaultutils.ConnectionConfig) error {
	if _, exists := c.builder.connectionConfig[updatedConfig.ConnectionId]; !exists {
		return fmt.Errorf("connection ID %s not found", updatedConfig.ConnectionId)
	}

	c.builder.connectionConfig[updatedConfig.ConnectionId] = updatedConfig
	return nil
}

func (s *Skyflow) GetVaultConfig(vaultId string) (vaultutils.VaultConfig, error) {

	config, exists := s.builder.vaultConfigs[vaultId]
	if !exists {
		return vaultutils.VaultConfig{}, fmt.Errorf("vault config with ID %s not found", vaultId)
	}
	return config, nil
}
func (s *Skyflow) AddVaultConfig(config vaultutils.VaultConfig) error {

	if _, exists := s.builder.vaultConfigs[config.VaultId]; exists {
		return fmt.Errorf("vault config with ID %s already exists", config.VaultId)
	}

	s.builder.vaultConfigs[config.VaultId] = config
	return nil
}
func (s *Skyflow) RemoveVaultConfig(vaultId string) error {
	if _, exists := s.builder.vaultConfigs[vaultId]; !exists {
		return fmt.Errorf("vault config with ID %s not found", vaultId)
	}

	delete(s.builder.vaultConfigs, vaultId)
	return nil
}
