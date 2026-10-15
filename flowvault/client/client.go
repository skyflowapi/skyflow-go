package client

import (
	"fmt"

	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/flowvault/utils/messages"
)

// Skyflow is the flowvault client entry point.
type Skyflow struct {
	vaultServices map[string]*VaultService
	credentials   *fvcommon.Credentials
	logLevel      fvlogger.LogLevel
	customHeaders map[fvcommon.CustomHeaderKey]string
}

type Option func(*Skyflow) *fverrors.SkyflowError

// NewSkyflow initializes a flowvault Skyflow client with the given options.
func NewSkyflow(opts ...Option) (*Skyflow, *fverrors.SkyflowError) {
	client := &Skyflow{
		vaultServices: make(map[string]*VaultService),
	}
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}
	fvlogger.Info(logs.CLIENT_INITIALIZED)
	return client, nil
}

// WithVaults registers vault configurations.
func WithVaults(configs ...fvcommon.VaultConfig) Option {
	return func(s *Skyflow) *fverrors.SkyflowError {
		if len(configs) == 0 {
			fvlogger.Error(logs.EMPTY_VAULT_ARRAY)
			return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.EMPTY_VAULT_CONFIG)
		}
		for _, cfg := range configs {
			if err := s.AddVaultConfig(cfg); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithCredentials sets client-level credentials, used by vaults whose config has none.
func WithCredentials(credentials fvcommon.Credentials) Option {
	return func(s *Skyflow) *fverrors.SkyflowError {
		return s.AddSkyflowCredentials(credentials)
	}
}

// WithLogLevel sets the logging level.
func WithLogLevel(logLevel fvlogger.LogLevel) Option {
	return func(s *Skyflow) *fverrors.SkyflowError {
		s.logLevel = logLevel
		fvlogger.SetLogLevel(logLevel)
		return nil
	}
}

// WithCustomHeaders sets custom headers sent with every request from this client.
func WithCustomHeaders(headers map[fvcommon.CustomHeaderKey]string) Option {
	return func(s *Skyflow) *fverrors.SkyflowError {
		s.customHeaders = headers
		return nil
	}
}

// Vault returns the vault service registered under vaultId.
func (s *Skyflow) Vault(vaultId string) (*VaultService, *fverrors.SkyflowError) {
	svc, ok := s.vaultServices[vaultId]
	if !ok {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	return svc, nil
}

// ---------------------------------------------------------------------------
// Vault config management
// ---------------------------------------------------------------------------

// AddVaultConfig registers a new vault config with the client.
// Returns an error if the config is invalid or the vault ID is already registered.
func (s *Skyflow) AddVaultConfig(cfg fvcommon.VaultConfig) *fverrors.SkyflowError {
	fvlogger.Info(logs.VALIDATING_VAULT_CONFIG)
	if err := fvhelpers.ConfigValidator().ValidateVaultConfig(cfg); err != nil {
		return err
	}
	if _, exists := s.vaultServices[cfg.VaultId]; exists {
		fvlogger.Error(fmt.Sprintf(logs.VAULT_CONFIG_EXISTS, cfg.VaultId))
		return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fmt.Sprintf(fverrors.VAULT_ID_EXISTS_IN_CONFIG_LIST, cfg.VaultId))
	}
	s.vaultServices[cfg.VaultId] = newVaultService(s, &cfg)
	fvlogger.Info(fmt.Sprintf(logs.VAULT_CONTROLLER_INITIALIZED, cfg.VaultId))
	return nil
}

// UpdateVaultConfig updates the credentials, cluster ID, or environment of a registered vault config.
// Returns an error if vaultId is empty or the config is not registered.
func (s *Skyflow) UpdateVaultConfig(updatedConfig fvcommon.VaultConfig) *fverrors.SkyflowError {
	fvlogger.Info(logs.VALIDATING_VAULT_CONFIG)
	if err := fvhelpers.ConfigValidator().ValidateUpdateVaultConfig(updatedConfig); err != nil {
		return err
	}
	svc, exists := s.vaultServices[updatedConfig.VaultId]
	if !exists {
		fvlogger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, updatedConfig.VaultId))
		return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	// The service holds a pointer to its config, so in-place updates reach it
	// and any controller it has built; cached tokens are dropped on change.
	if !fvcommon.IsCredentialsEmpty(updatedConfig.Credentials) {
		svc.config.Credentials = updatedConfig.Credentials
		svc.resetAuth()
	}
	if updatedConfig.ClusterId != "" {
		svc.config.ClusterId = updatedConfig.ClusterId
	}
	if updatedConfig.BaseVaultUrl != "" {
		svc.config.BaseVaultUrl = updatedConfig.BaseVaultUrl
	}
	svc.config.Env = updatedConfig.Env
	return nil
}

// RemoveVaultConfig removes the vault config registered under vaultId.
// Returns an error if the vault ID is not registered.
func (s *Skyflow) RemoveVaultConfig(vaultId string) *fverrors.SkyflowError {
	if _, exists := s.vaultServices[vaultId]; !exists {
		fvlogger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, vaultId))
		return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	delete(s.vaultServices, vaultId)
	return nil
}

// GetVaultConfig returns the VaultConfig registered under vaultId, or an error if it is not found.
func (s *Skyflow) GetVaultConfig(vaultId string) (*fvcommon.VaultConfig, *fverrors.SkyflowError) {
	svc, exists := s.vaultServices[vaultId]
	if !exists {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	return svc.config, nil
}

// ---------------------------------------------------------------------------
// Credentials
// ---------------------------------------------------------------------------

// AddSkyflowCredentials sets or replaces client-level credentials and propagates them to all vaults.
func (s *Skyflow) AddSkyflowCredentials(credentials fvcommon.Credentials) *fverrors.SkyflowError {
	fvlogger.Info(logs.VALIDATING_CRED)
	if err := fvhelpers.ConfigValidator().ValidateCredentials(credentials); err != nil {
		return err
	}
	s.credentials = &credentials
	for _, svc := range s.vaultServices {
		svc.resetAuth()
	}
	return nil
}

// UpdateSkyflowCredentials replaces client-level credentials and propagates them to all vaults.
func (s *Skyflow) UpdateSkyflowCredentials(credentials fvcommon.Credentials) *fverrors.SkyflowError {
	return s.AddSkyflowCredentials(credentials)
}

// GetSkyflowCredentials returns the client-level credentials, or nil if none are set.
func (s *Skyflow) GetSkyflowCredentials() *fvcommon.Credentials {
	return s.credentials
}

// ---------------------------------------------------------------------------
// Logging
// ---------------------------------------------------------------------------

// UpdateLogLevel changes the log level for the client.
func (s *Skyflow) UpdateLogLevel(logLevel fvlogger.LogLevel) {
	fvlogger.Info(fmt.Sprintf(logs.CURRENT_LOG_LEVEL, s.logLevel))
	s.logLevel = logLevel
	fvlogger.SetLogLevel(logLevel)
}

// GetLoglevel returns the current log level of the client.
func (s *Skyflow) GetLoglevel() *fvlogger.LogLevel {
	logLevel := s.logLevel
	return &logLevel
}
