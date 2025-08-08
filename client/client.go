package client

import (
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	vaultutils "github.com/skyflowapi/skyflow-go/v2/utils/common"
	error "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
	"os"
)

type Skyflow struct {
	vaultServices      map[string]*vaultService
	connectionServices map[string]*connectionService
    detectServices     map[string]*detectService
	credentials        *vaultutils.Credentials
	logLevel           logger.LogLevel
}

type Option func(*Skyflow) *error.SkyflowError

// NewSkyflow initializes a Skyflow client with the given options.
func NewSkyflow(opts ...Option) (*Skyflow, *error.SkyflowError) {
	client := &Skyflow{
		vaultServices:      make(map[string]*vaultService),
		connectionServices: make(map[string]*connectionService),
		credentials:        nil,
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	logger.Info(logs.CLIENT_INITIALIZED)
	return client, nil
}

// WithVaults sets a vault configuration.
func WithVaults(config ...vaultutils.VaultConfig) Option {
	return func(s *Skyflow) *error.SkyflowError {
		if config == nil {
			logger.Error(logs.EMPTY_VAULT_ARRAY)
			return error.NewSkyflowError(error.INVALID_INPUT_CODE, fmt.Sprintf(error.EMPTY_VAULT_CONFIG))
		} else if len(config) == 0 {
			logger.Error(logs.EMPTY_VAULT_ARRAY)
			return error.NewSkyflowError(error.INVALID_INPUT_CODE, fmt.Sprintf(error.EMPTY_VAULT_CONFIG))
		}

		for _, vaultConfig := range config {
			if _, exists := s.vaultServices[vaultConfig.VaultId]; exists {
				logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_EXISTS, vaultConfig.VaultId))
				return error.NewSkyflowError(error.INVALID_INPUT_CODE, fmt.Sprintf(error.VAULT_ID_EXITS_IN_CONFIG_LIST, vaultConfig.VaultId))
			}
			if _, exists := s.detectServices[vaultConfig.VaultId]; exists {
				logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_EXISTS, vaultConfig.VaultId))
				return error.NewSkyflowError(error.INVALID_INPUT_CODE, fmt.Sprintf(error.VAULT_ID_EXITS_IN_CONFIG_LIST, vaultConfig.VaultId))
			}
			// validate the config
			logger.Info(logs.VALIDATING_VAULT_CONFIG)
			if err := validation.ValidateVaultConfig(vaultConfig); err != nil {
				return err
			}

			// create vault service for config
			s.vaultServices[vaultConfig.VaultId] = &vaultService{
				config:   &vaultConfig,
				logLevel: &s.logLevel,
			}

			s.detectServices[vaultConfig.VaultId] = &detectService{
				config:   &vaultConfig,
				logLevel: &s.logLevel,
			}
		}
		return nil
	}
}

// WithConnections sets a connection configuration.
func WithConnections(config ...vaultutils.ConnectionConfig) Option {
	return func(s *Skyflow) *error.SkyflowError {

		if config == nil {
			logger.Error(logs.EMPTY_CONNECTION_ARRAY)
			return error.NewSkyflowError(error.INVALID_INPUT_CODE, fmt.Sprintf(error.EMPTY_CONNECTION_CONFIG))
		} else if len(config) == 0 {
			logger.Error(logs.EMPTY_CONNECTION_ARRAY)
			return error.NewSkyflowError(error.INVALID_INPUT_CODE, fmt.Sprintf(error.EMPTY_CONNECTION_CONFIG))
		}

		for _, connectionConfig := range config {
			if _, exists := s.connectionServices[connectionConfig.ConnectionId]; exists {
				logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_EXISTS, connectionConfig.ConnectionId))
				return error.NewSkyflowError(error.INVALID_INPUT_CODE, fmt.Sprintf(error.CONNECTION_ID_EXITS_IN_CONFIG_LIST, connectionConfig.ConnectionId))
			}
			// validate the config
			logger.Info(logs.VALIDATING_CONNECTION_CONFIG)
			if err := validation.ValidateConnectionConfig(connectionConfig); err != nil {
				return err
			}

			// create the connection service
			s.connectionServices[connectionConfig.ConnectionId] = &connectionService{
				config:   connectionConfig,
				logLevel: &s.logLevel,
			}
		}
		return nil
	}
}

// WithCredentials sets credentials for the Skyflow client.
func WithCredentials(credentials vaultutils.Credentials) Option {
	return func(s *Skyflow) *error.SkyflowError {
		logger.Info(logs.VALIDATING_CRED)
		if err := validation.ValidateCredentials(credentials); err != nil {
			return err
		}
		s.credentials = &credentials
		return nil
	}
}

// WithLogLevel sets the logging level.
func WithLogLevel(logLevel logger.LogLevel) Option {
	return func(s *Skyflow) *error.SkyflowError {
		s.logLevel = logLevel
		logger.SetLogLevel(logLevel)
		return nil
	}
}

// Vault retrieves a vault service by ID.
func (s *Skyflow) Vault(vaultID ...string) (*vaultService, *error.SkyflowError) {
	// get vaultapi config if available in vaultapi configs, skyflow or env
	config, err := getVaultConfig(s.vaultServices, vaultID...)
	if err != nil {
		return nil, err
	}

	err = setVaultCredentials(config, s.credentials)
	if err != nil {
		return nil, err
	}
	err1 := validation.ValidateVaultConfig(*config)
	if err1 != nil {
		return nil, err1
	}
	vaultService := &vaultService{}
	// Get the VaultController from the builder's vaultServices map
	vaultService, exists := s.vaultServices[config.VaultId]
	if !exists {
		vaultService.config = config
		vaultService.logLevel = &s.logLevel
	}
	// Update the config in the vaultapi service
	vaultService.controller = &controller.VaultController{
		Config:   *config,
		Loglevel: &s.logLevel,
	}
	vaultService.config = config
	return vaultService, nil
}

func (s *Skyflow) Connection(connectionId ...string) (*connectionService, *error.SkyflowError) {
	config, err := getConnectionConfig(s.connectionServices, connectionId...)
	if err != nil {
		return nil, err
	}
	err = setConnectionCredentials(config, s.credentials)
	if err != nil {
		return nil, err
	}
	err1 := validation.ValidateConnectionConfig(*config)
	if err1 != nil {
		return nil, err1
	}
	connectionService := &connectionService{}
	connectionService, exists := s.connectionServices[config.ConnectionId]
	if !exists {
		connectionService.config = *config
		connectionService.logLevel = &s.logLevel
		s.connectionServices[config.ConnectionId] = connectionService
	}
	connectionService.controller = controller.ConnectionController{
		Config:   *config,
		Loglevel: &s.logLevel,
	}
	connectionService.config = *config
	return connectionService, nil
}

func (s *Skyflow) Detect(vaultID ...string) (*detectService, *error.SkyflowError) {
		// get vaultapi config if available in vaultapi configs, skyflow or env
	config, err := getDetectConfig(s.detectServices, vaultID...)
	if err != nil {
		return nil, err
	}

	err = setDetectCredentials(config, s.credentials)
	if err != nil {
		return nil, err
	}
	err1 := validation.ValidateVaultConfig(*config)
	if err1 != nil {
		return nil, err1
	}
	detectService := &detectService{}
	// Get the VaultController from the builder's vaultServices map
	detectService, exists := s.detectServices[config.VaultId]
	if !exists {
		detectService.config = config
		detectService.logLevel = &s.logLevel
	}
	// Update the config in the vaultapi service
	detectService.controller = &controller.DetectController{
		Config:   *config,
		Loglevel: &s.logLevel,
	}
	detectService.config = config
	return detectService, nil
}

func (s *Skyflow) GetVault(vaultId string) (*vaultutils.VaultConfig, *error.SkyflowError) {
	config, exist := s.vaultServices[vaultId]
	if !exist {
		return nil, error.NewSkyflowError(error.INVALID_INPUT_CODE, error.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	return config.config, nil
}

func (s *Skyflow) GetConnection(connId string) (*vaultutils.ConnectionConfig, *error.SkyflowError) {
	config, exist := s.connectionServices[connId]
	if !exist {
		return nil, error.NewSkyflowError(error.INVALID_INPUT_CODE, error.CONNECTION_ID_NOT_IN_CONFIG_LIST)
	}
	return &config.config, nil
}

// UpdateLogLevel update methods
func (s *Skyflow) UpdateLogLevel(logLevel logger.LogLevel) {
	logger.Info(fmt.Sprintf(logs.CURRENT_LOG_LEVEL, s.logLevel))
	s.logLevel = logLevel
	for _, service := range s.vaultServices {
		service.logLevel = &s.logLevel
	}
	for _, service := range s.detectServices {
		service.logLevel = &s.logLevel
	}
	logger.SetLogLevel(logLevel)
}

func (s *Skyflow) UpdateSkyflowCredentials(credentials vaultutils.Credentials) *error.SkyflowError {
	err := validation.ValidateCredentials(credentials)
	if err != nil {
		return err
	}
	s.credentials = &credentials
	return nil
}

func (s *Skyflow) UpdateVault(updatedConfig vaultutils.VaultConfig) *error.SkyflowError {
	logger.Info(logs.VALIDATING_VAULT_CONFIG)
	e := validation.ValidateVaultConfig(updatedConfig)
	if e != nil {
		return e
	}
	if _, exists := s.vaultServices[updatedConfig.VaultId]; !exists {
		logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, updatedConfig.VaultId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
	}

	s.vaultServices[updatedConfig.VaultId].config = &updatedConfig

	if _, exists := s.detectServices[updatedConfig.VaultId]; !exists {
		logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, updatedConfig.VaultId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
	}

	s.detectServices[updatedConfig.VaultId].config = &updatedConfig
	return nil
}

func (s *Skyflow) UpdateConnection(updatedConfig vaultutils.ConnectionConfig) *error.SkyflowError {
	logger.Info(logs.VALIDATING_CONNECTION_CONFIG)
	err := validation.ValidateConnectionConfig(updatedConfig)
	if err != nil {
		return err
	}
	if _, exists := s.connectionServices[updatedConfig.ConnectionId]; !exists {
		logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_DOES_NOT_EXIST, updatedConfig.ConnectionId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.CONNECTION_ID_NOT_IN_CONFIG_LIST)
	}

	s.connectionServices[updatedConfig.ConnectionId].config = updatedConfig
	return nil
}

func (s *Skyflow) GetLoglevel() *logger.LogLevel {
	loglevel := s.logLevel
	return &loglevel
}

func (s *Skyflow) AddVault(config vaultutils.VaultConfig) *error.SkyflowError {
	logger.Info(logs.VALIDATING_VAULT_CONFIG)
	e := validation.ValidateVaultConfig(config)
	if e != nil {
		return e
	}
	// add new config
	if _, exists := s.vaultServices[config.VaultId]; exists {
		logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, config.VaultId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_ALREADY_IN_CONFIG_LIST)
	}
	// add service instance for new config
	if _, exists := s.vaultServices[config.VaultId]; !exists {
		s.vaultServices[config.VaultId] = &vaultService{
			config:   &config,
			logLevel: &s.logLevel,
		}
	}
	if _, exists := s.detectServices[config.VaultId]; exists {
		logger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, config.VaultId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_ALREADY_IN_CONFIG_LIST)
	}
	// add service instance for new config
	if _, exists := s.detectServices[config.VaultId]; !exists {
		s.detectServices[config.VaultId] = &detectService{
			config:   &config,
			logLevel: &s.logLevel,
		}
	}
	logger.Info(fmt.Sprintf(logs.VAULT_CONTROLLER_INITIALIZED, config.VaultId))
	return nil
}

func (s *Skyflow) AddSkyflowCredentials(config vaultutils.Credentials) *error.SkyflowError {
	err := validation.ValidateCredentials(config)
	if err != nil {
		return err
	}
	s.credentials = &config
	return nil
}

func (s *Skyflow) AddConnection(config vaultutils.ConnectionConfig) *error.SkyflowError {
	logger.Info(logs.VALIDATING_CONNECTION_CONFIG)
	err := validation.ValidateConnectionConfig(config)
	if err != nil {
		return err
	}
	// add new config
	if _, exists := s.connectionServices[config.ConnectionId]; exists {
		logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_EXISTS, config.ConnectionId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.CONNECTION_ID_EXITS_IN_CONFIG_LIST)
	}
	// add service instance for new config
	if _, exists := s.connectionServices[config.ConnectionId]; !exists {
		s.connectionServices[config.ConnectionId] = &connectionService{
			config:   config,
			logLevel: &s.logLevel,
		}
	}
	logger.Info(fmt.Sprintf(logs.CONNECTION_CONTROLLER_INITIALIZED, config.ConnectionId))
	return nil
}

func (s *Skyflow) RemoveVault(vaultId string) *error.SkyflowError {
	if _, exists := s.vaultServices[vaultId]; !exists {
		logger.Error(fmt.Sprintf(logs.VAULT_ID_CONFIG_DOES_NOT_EXIST, vaultId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	delete(s.vaultServices, vaultId)

	if _, exists := s.detectServices[vaultId]; !exists {
		logger.Error(fmt.Sprintf(logs.VAULT_ID_CONFIG_DOES_NOT_EXIST, vaultId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	delete(s.detectServices, vaultId)
	return nil
}

func (s *Skyflow) RemoveConnection(connectionId string) *error.SkyflowError {
	if _, exists := s.connectionServices[connectionId]; !exists {
		logger.Error(fmt.Sprintf(logs.CONNECTION_CONFIG_DOES_NOT_EXIST, connectionId))
		return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.CONNECTION_ID_NOT_IN_CONFIG_LIST)
	}
	delete(s.connectionServices, connectionId)
	return nil
}

// vault utils or helper func
func getVaultConfig(builder map[string]*vaultService, vaultID ...string) (*vaultutils.VaultConfig, *error.SkyflowError) {
	// if vaultapi configs are empty
	if len(builder) == 0 {
		return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.EMPTY_VAULT_CONFIG)
	}

	// if vaultapi is passed
	if len(vaultID) > 0 && len(builder) > 0 {
		config, exists := builder[vaultID[0]]
		if !exists {
			return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
		}
		return config.config, nil
	}

	// No vaultapi ID passed, return the first vaultapi config available
	for _, cfg := range builder {
		return cfg.config, nil
	}

	return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
}
func getDetectConfig(builder map[string]*detectService, vaultID ...string) (*vaultutils.VaultConfig, *error.SkyflowError) {
	// if vaultapi configs are empty
	if len(builder) == 0 {
		return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.EMPTY_VAULT_CONFIG)
	}

	// if vaultapi is passed
	if len(vaultID) > 0 && len(builder) > 0 {
		config, exists := builder[vaultID[0]]
		if !exists {
			return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
		}
		return config.config, nil
	}

	// No vaultapi ID passed, return the first vaultapi config available
	for _, cfg := range builder {
		return cfg.config, nil
	}

	return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.VAULT_ID_NOT_IN_CONFIG_LIST)
}
func setVaultCredentials(config *vaultutils.VaultConfig, builderCreds *vaultutils.Credentials) *error.SkyflowError {
	// here if credentials are empty in the vaultapi config
	if config == nil || isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if builderCreds != nil {
			if !isCredentialsEmpty(*builderCreds) {
				config.Credentials = *builderCreds
			}
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.EMPTY_CREDENTIALS)
		}
	}
	return nil
}
func isCredentialsEmpty(creds vaultutils.Credentials) bool {
	return creds.Path == "" &&
		creds.CredentialsString == "" &&
		creds.Token == "" && creds.ApiKey == ""
}
func setConnectionCredentials(config *vaultutils.ConnectionConfig, builderCreds *vaultutils.Credentials) *error.SkyflowError {
	// here if credentials are empty in the vaultapi config
	if config == nil || isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if !isCredentialsEmpty(*builderCreds) {
			config.Credentials = *builderCreds
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.EMPTY_CREDENTIALS)
		}
	}
	return nil
}
func setDetectCredentials(config *vaultutils.VaultConfig, builderCreds *vaultutils.Credentials) *error.SkyflowError {
	// here if credentials are empty in the vaultapi config
	if config == nil || isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if !isCredentialsEmpty(*builderCreds) {
			config.Credentials = *builderCreds
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			config.Credentials.CredentialsString = envCreds
		} else {
			return error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.EMPTY_CREDENTIALS)
		}
	}
	return nil
}
func getConnectionConfig(builder map[string]*connectionService, connectionId ...string) (*vaultutils.ConnectionConfig, *error.SkyflowError) {
	// if connection configs are empty
	if len(builder) == 0 {
		return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.EMPTY_CONNECTION_CONFIG)
	}

	// if connection id is passed
	if len(connectionId) > 0 && len(builder) > 0 {
		config, exists := builder[connectionId[0]]
		if !exists {
			return nil, error.NewSkyflowError(error.ErrorCodesEnum(error.INVALID_INPUT_CODE), error.CONNECTION_ID_NOT_IN_CONFIG_LIST)
		}
		return &config.config, nil
	}

	// No conenction ID passed, return the first config available
	for _, cfg := range builder {
		return &cfg.config, nil
	}

	return nil, nil
}
