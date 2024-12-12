package client

import (
	"fmt"
	"os"
	"skyflow-go/v2/internal/validation"
	"skyflow-go/v2/internal/vault/controller"
	vaultutils "skyflow-go/v2/utils/common"
	errors "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
	logs "skyflow-go/v2/utils/messages"
)

type Skyflow struct {
	builder *SkyflowBuilder
}

const defaultLogLevel logger.LogLevel = 0

type SkyflowBuilder struct {
	vaultConfigs       map[string][]vaultutils.VaultConfig
	vaultServices      map[string]*vaultService
	connectionServices map[string]*connectionService
	connectionConfig   map[string][]vaultutils.ConnectionConfig
	credentials        *vaultutils.Credentials
	logLevel           logger.LogLevel
}

func (c *Skyflow) Builder() *SkyflowBuilder {
	return &SkyflowBuilder{
		vaultConfigs:       make(map[string][]vaultutils.VaultConfig),
		vaultServices:      make(map[string]*vaultService),
		connectionServices: make(map[string]*connectionService),
		connectionConfig:   make(map[string][]vaultutils.ConnectionConfig),
		logLevel:           defaultLogLevel,
	}
}

func (b *SkyflowBuilder) WithVaultConfig(vaultConfig vaultutils.VaultConfig) *SkyflowBuilder {
	b.vaultConfigs[vaultConfig.VaultId] = append(b.vaultConfigs[vaultConfig.VaultId], vaultConfig)
	return b
}

func (b *SkyflowBuilder) WithConnectionConfig(connConfig vaultutils.ConnectionConfig) *SkyflowBuilder {
	b.connectionConfig[connConfig.ConnectionId] = append(b.connectionConfig[connConfig.ConnectionId], connConfig)
	return b
}

func (b *SkyflowBuilder) WithSkyflowCredentials(credentials vaultutils.Credentials) *SkyflowBuilder {
	b.credentials = &credentials
	return b
}

func (b *SkyflowBuilder) WithLogLevel(logLevel logger.LogLevel) *SkyflowBuilder {
	b.logLevel = logLevel
	logger.SetLogLevel(b.logLevel)
	return b
}

func (b *SkyflowBuilder) Build() (*Skyflow, *errors.SkyflowError) {
	for _, configs := range b.vaultConfigs {
		if len(configs) > 1 {
			logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_EXISTS, configs[0].VaultId))
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.VAULT_ID_ALREADY_IN_CONFIG_LIST)
		}
	}
	for _, configs := range b.connectionConfig {
		if len(configs) > 1 {
			logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_EXISTS, configs[0].ConnectionId))
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.CONNECTION_ID_ALREADY_IN_CONFIG_LIST)
		}
	}
	if len(b.vaultConfigs) != 0 {
		for vaultID, vaultConfig := range b.vaultConfigs {
			logger.Info(logs.VALIDATING_VAULT_CONFIG)
			validationErr := validation.ValidateVaultConfig(vaultConfig[0])
			if validationErr != nil {
				return nil, validationErr
			}
			if _, exists := b.vaultServices[vaultID]; !exists {
				logger.Info(logs.VALIDATE_CONNECTION_CONFIG)
				b.vaultServices[vaultID] = &vaultService{
					config:   &vaultConfig[0],
					logLevel: &b.logLevel,
				}
			}
		}
	}
	if len(b.connectionConfig) != 0 {
		for connectionId, connectionConfig := range b.connectionConfig {
			validationErr := validation.ValidateConnectionConfig(connectionConfig[0])
			if validationErr != nil {
				return nil, validationErr
			}
			if _, exists := b.vaultServices[connectionId]; !exists {
				b.connectionServices[connectionId] = &connectionService{
					config:   connectionConfig[0],
					logLevel: &b.logLevel,
				}
			}
		}
	}
	if b.credentials != nil {
		validationErr := validation.ValidateCredentials(*b.credentials)
		if validationErr != nil {
			return nil, validationErr
		}
	}

	logger.Info(logs.CLIENT_INITIALIZED)
	return &Skyflow{
		builder: b,
	}, nil
}

// Vault vaultapi method
func (c *Skyflow) Vault(vaultID ...string) (*vaultService, *errors.SkyflowError) {
	// get vaultapi config if available in vaultapi configs, skyflow or env
	config, err := getVaultConfig(*c.builder, vaultID...)
	if err != nil {
		return nil, err
	}
	cred, err := setVaultCredentials(config, *c.builder)
	config.Credentials = cred
	if err != nil {
		return nil, err
	}
	err1 := validation.ValidateVaultConfig(config)
	if err1 != nil {
		return nil, err1
	}
	vaultService := &vaultService{}
	// Get the VaultController from the builder's vaultServices map
	vaultService, exists := c.builder.vaultServices[config.VaultId]
	if !exists {
		vaultService.config = &config
		vaultService.logLevel = &c.builder.logLevel
	}
	// Update the config in the vaultapi service
	vaultService.controller = &controller.VaultController{
		Config:   config,
		Loglevel: &c.builder.logLevel,
	}
	vaultService.config = &config
	return vaultService, nil
}
func (c *Skyflow) Connection(connectionId ...string) (*connectionService, *errors.SkyflowError) {
	config, err := getConnectionConfig(c.builder, connectionId...)
	if err != nil {
		return nil, err
	}
	err = setConnectionCredentials(config, *c.builder)
	if err != nil {
		return nil, err
	}
	err1 := validation.ValidateConnectionConfig(*config)
	if err1 != nil {
		return nil, err1
	}
	connectionService := &connectionService{}
	connectionService, exists := c.builder.connectionServices[config.ConnectionId]
	if !exists {
		connectionService.config = *config
		connectionService.logLevel = &c.builder.logLevel
		c.builder.connectionServices[config.ConnectionId] = connectionService
	}
	connectionService.controller = controller.ConnectionController{
		Config:   *config,
		Loglevel: &c.builder.logLevel,
	}
	connectionService.config = *config
	return connectionService, nil
}

// UpdateLogLevel update methods
func (c *Skyflow) UpdateLogLevel(logLevel logger.LogLevel) {
	logger.Info(fmt.Sprintf(logs.CURRENT_LOG_LEVEL, c.builder.logLevel))
	c.builder.logLevel = logLevel
	for _, service := range c.builder.vaultServices {
		service.logLevel = &c.builder.logLevel
	}
	logger.SetLogLevel(logLevel)
}
func (c *Skyflow) UpdateSkyflowCredentials(credentials vaultutils.Credentials) *errors.SkyflowError {
	err := validation.ValidateCredentials(credentials)
	if err != nil {
		return err
	}
	c.builder.credentials = &credentials
	return nil
}
func (c *Skyflow) UpdateVaultConfig(updatedConfig vaultutils.VaultConfig) *errors.SkyflowError {
	logger.Info(logs.VALIDATING_VAULT_CONFIG)
	e := validation.ValidateVaultConfig(updatedConfig)
	if e != nil {
		return e
	}
	if _, exists := c.builder.vaultConfigs[updatedConfig.VaultId]; !exists {
		logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, updatedConfig.VaultId))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}

	c.builder.vaultConfigs[updatedConfig.VaultId][0] = updatedConfig
	c.builder.vaultServices[updatedConfig.VaultId].config = &updatedConfig
	return nil
}
func (c *Skyflow) UpdateConnectionConfig(updatedConfig vaultutils.ConnectionConfig) *errors.SkyflowError {
	logger.Info(logs.VALIDATING_CONNECTION_CONFIG)
	err := validation.ValidateConnectionConfig(updatedConfig)
	if err != nil {
		return err
	}
	if _, exists := c.builder.connectionConfig[updatedConfig.ConnectionId]; !exists {
		logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_DOES_NOT_EXIST, updatedConfig.ConnectionId))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.CONNECTION_ID_NOT_IN_CONFIG_LIST)
	}

	c.builder.connectionConfig[updatedConfig.ConnectionId][0] = updatedConfig
	c.builder.connectionServices[updatedConfig.ConnectionId].config = updatedConfig
	return nil
}
func (c *Skyflow) GetVaultConfig(vaultId string) (vaultutils.VaultConfig, *errors.SkyflowError) {
	config, exists := c.builder.vaultConfigs[vaultId]
	if !exists {
		return vaultutils.VaultConfig{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	return config[0], nil
}
func (c *Skyflow) GetConnectionConfig(connectionId string) (vaultutils.ConnectionConfig, *errors.SkyflowError) {

	config, exists := c.builder.connectionConfig[connectionId]
	if !exists {
		return vaultutils.ConnectionConfig{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), fmt.Sprintf("vaultapi ID %s not found", connectionId))
	}
	return config[0], nil
}
func (c *Skyflow) GetLoglevel() (*logger.LogLevel, *errors.SkyflowError) {
	loglevel := c.builder.logLevel
	return &loglevel, nil
}
func (c *Skyflow) AddVaultConfig(config vaultutils.VaultConfig) *errors.SkyflowError {
	logger.Info(logs.VALIDATING_VAULT_CONFIG)
	e := validation.ValidateVaultConfig(config)
	if e != nil {
		return e
	}
	// add new config
	if _, exists := c.builder.vaultConfigs[config.VaultId]; exists {
		logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, config.VaultId))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.VAULT_ID_ALREADY_IN_CONFIG_LIST)
	}
	c.builder.vaultConfigs[config.VaultId] = append(c.builder.vaultConfigs[config.VaultId], config)

	// add service instance for new config
	if _, exists := c.builder.vaultServices[config.VaultId]; !exists {
		c.builder.vaultServices[config.VaultId] = &vaultService{
			config:   &config,
			logLevel: &c.builder.logLevel,
		}
	}
	logger.Info(logs.VAULT_CONTROLLER_INITIALIZED)
	return nil
}
func (c *Skyflow) AddSkyflowCredentials(config vaultutils.Credentials) *errors.SkyflowError {
	err := validation.ValidateCredentials(config)
	if err != nil {
		return err
	}
	c.builder.credentials = &config
	return nil
}
func (c *Skyflow) AddConnectionConfig(config vaultutils.ConnectionConfig) *errors.SkyflowError {
	logger.Info(logs.VALIDATING_CONNECTION_CONFIG)
	err := validation.ValidateConnectionConfig(config)
	if err != nil {
		return err
	}
	// add new config
	if _, exists := c.builder.connectionConfig[config.ConnectionId]; exists {
		logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_EXISTS, config.ConnectionId))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.CONNECTION_ID_EXITS_IN_CONFIG_LIST)
	}
	c.builder.connectionConfig[config.ConnectionId] = append(c.builder.connectionConfig[config.ConnectionId], config)
	// add service instance for new config
	if _, exists := c.builder.connectionServices[config.ConnectionId]; !exists {
		c.builder.connectionServices[config.ConnectionId] = &connectionService{
			config:   config,
			logLevel: &c.builder.logLevel,
		}
	}
	logger.Info(logs.CONNECTION_CONTROLLER_INITIALIZED)
	return nil
}
func (c *Skyflow) RemoveVaultConfig(vaultId string) *errors.SkyflowError {
	if _, exists := c.builder.vaultConfigs[vaultId]; !exists {
		logger.Error(fmt.Sprintf(logs.VAULT_ID_CONFIG_DOES_NOT_EXIST, vaultId))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}

	delete(c.builder.vaultConfigs, vaultId)
	delete(c.builder.vaultServices, vaultId)
	return nil
}
func (c *Skyflow) RemoveConnectionConfig(connectionId string) *errors.SkyflowError {
	if _, exists := c.builder.connectionConfig[connectionId]; !exists {
		logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_DOES_NOT_EXIST, connectionId))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.CONNECTION_ID_NOT_IN_CONFIG_LIST)
	}
	delete(c.builder.connectionConfig, connectionId)
	delete(c.builder.connectionServices, connectionId)
	return nil
}

// vault utils or helper func
func isCredentialsEmpty(creds vaultutils.Credentials) bool {
	return creds.Path == "" &&
		creds.CredentialsString == "" &&
		creds.Token == "" && creds.ApiKey == ""
}
func getConnectionConfig(builder *SkyflowBuilder, connectionId ...string) (*vaultutils.ConnectionConfig, *errors.SkyflowError) {
	// if connection configs are empty
	if len(builder.connectionConfig) == 0 {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.EMPTY_CONNECTION_CONFIG)
	}

	// if connection id is passed
	if len(connectionId) > 0 && len(builder.connectionConfig) > 0 {
		config, exists := builder.connectionConfig[connectionId[0]]
		if !exists {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.CONNECTION_ID_NOT_IN_CONFIG_LIST)
		}
		return &config[0], nil
	}

	// No conenction ID passed, return the first config available
	for _, cfg := range builder.connectionConfig {
		return &cfg[0], nil
	}

	return nil, nil
}
func getVaultConfig(builder SkyflowBuilder, vaultID ...string) (vaultutils.VaultConfig, *errors.SkyflowError) {
	// if vaultapi configs are empty
	if len(builder.vaultConfigs) == 0 {
		return vaultutils.VaultConfig{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.EMPTY_VAULT_CONFIG)
	}

	// if vaultapi is passed
	if len(vaultID) > 0 && len(builder.vaultConfigs) > 0 {
		config, exists := builder.vaultConfigs[vaultID[0]]
		if !exists {
			return vaultutils.VaultConfig{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.VAULT_ID_NOT_IN_CONFIG_LIST)
		}
		return config[0], nil
	}

	// No vaultapi ID passed, return the first vaultapi config available
	for _, cfg := range builder.vaultConfigs {
		return cfg[0], nil
	}

	return vaultutils.VaultConfig{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.VAULT_ID_NOT_IN_CONFIG_LIST)
}
func setVaultCredentials(config vaultutils.VaultConfig, builderCreds SkyflowBuilder) (vaultutils.Credentials, *errors.SkyflowError) {
	// here if credentials are empty in the vaultapi config\
	if isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if builderCreds.credentials != nil && !isCredentialsEmpty(*builderCreds.credentials) {
			config.Credentials = *builderCreds.credentials
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return vaultutils.Credentials{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.EMPTY_CREDENTIALS)
		}
	}
	return config.Credentials, nil
}
func setConnectionCredentials(config *vaultutils.ConnectionConfig, builderCreds SkyflowBuilder) *errors.SkyflowError {
	// here if credentials are empty in the vaultapi config
	if config == nil || isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if builderCreds.credentials != nil && !isCredentialsEmpty(*builderCreds.credentials) {
			config.Credentials = *builderCreds.credentials
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.EMPTY_CREDENTIALS)
		}
	}
	return nil
}
