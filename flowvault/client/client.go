package client

import (
	"fmt"
	"slices"
	"sync"

	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvvalidation "github.com/skyflowapi/skyflow-go/flowvault/internal/validation"
	"github.com/skyflowapi/skyflow-go/flowvault/internal/vault/controller"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/flowvault/utils/messages"
)

// Skyflow is the flowvault client entry point. It is safe for concurrent use: vaults and
// client settings can be changed while calls are running.
type Skyflow struct {
	mu            sync.RWMutex
	vaultServices map[string]*VaultService
	// vaultOrder holds the registered vault IDs in the order they were added; Vault() with
	// no ID returns the first.
	vaultOrder  []string
	credentials *fvcommon.Credentials
	logLevel    fvlogger.LogLevel
	httpConfig  fvcommon.HttpConfig
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
// The level is process-wide: it is shared by every client of the v2 and flowvault SDKs.
func WithLogLevel(logLevel fvlogger.LogLevel) Option {
	return func(s *Skyflow) *fverrors.SkyflowError {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.logLevel = logLevel
		fvlogger.SetLogLevel(logLevel)
		fvlogger.Info(fmt.Sprintf(logs.CURRENT_LOG_LEVEL, logLevel))
		return nil
	}
}

// WithHttpConfig sets client-level timeout and retry defaults. A value set on a
// VaultConfig wins over the one set here.
func WithHttpConfig(httpConfig fvcommon.HttpConfig) Option {
	return func(s *Skyflow) *fverrors.SkyflowError {
		if err := fvhelpers.ConfigValidator().ValidateHttpConfig(httpConfig); err != nil {
			return err
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.httpConfig = copyHttpConfig(httpConfig)
		s.syncVaults()
		return nil
	}
}

// clientSettings returns the settings vaults fall back to. The caller holds s.mu.
func (s *Skyflow) clientSettings() controller.ClientSettings {
	return controller.ClientSettings{Credentials: s.credentials, HttpConfig: s.httpConfig}
}

// syncVaults pushes client-level settings to every registered vault. The caller holds s.mu.
func (s *Skyflow) syncVaults() {
	settings := s.clientSettings()
	for _, svc := range s.vaultServices {
		svc.controller.SetClientSettings(settings)
	}
}

// copyHttpConfig copies MaxRetries so later changes to the caller's value don't leak in.
func copyHttpConfig(config fvcommon.HttpConfig) fvcommon.HttpConfig {
	if config.MaxRetries != nil {
		retries := *config.MaxRetries
		config.MaxRetries = &retries
	}
	return config
}

// copyVaultConfig copies MaxRetries and the credentials so the copy shares nothing mutable.
func copyVaultConfig(config fvcommon.VaultConfig) fvcommon.VaultConfig {
	if config.MaxRetries != nil {
		retries := *config.MaxRetries
		config.MaxRetries = &retries
	}
	config.Credentials = copyCredentials(config.Credentials)
	return config
}

// copyCredentials copies Roles and a map or slice Context, so a change to either through
// one copy doesn't reach the other.
func copyCredentials(credentials fvcommon.Credentials) fvcommon.Credentials {
	if credentials.Roles != nil {
		credentials.Roles = append([]string(nil), credentials.Roles...)
	}
	credentials.Context = copyValue(credentials.Context)
	return credentials
}

// copyValue deep-copies JSON-like maps and slices; other values are returned as they are.
func copyValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		copied := make(map[string]interface{}, len(v))
		for key, item := range v {
			copied[key] = copyValue(item)
		}
		return copied
	case []interface{}:
		copied := make([]interface{}, len(v))
		for i, item := range v {
			copied[i] = copyValue(item)
		}
		return copied
	}
	return value
}

// Vault returns the vault service registered under vaultId. With no vaultId, or an empty
// one, it returns the default vault, the first one added; only the first vaultId is used.
func (s *Skyflow) Vault(vaultId ...string) (*VaultService, *fverrors.SkyflowError) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id := ""
	if len(vaultId) > 0 {
		id = vaultId[0]
	}
	svc, ok := s.lookup(id)
	if !ok {
		fvlogger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, vaultIdForLog(id)))
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	return svc, nil
}

// Detect returns the Detect service of the vault registered under vaultId, or of the
// default vault when vaultId is omitted or empty. It uses the same vault config,
// credentials and bearer token as Vault(vaultId).
func (s *Skyflow) Detect(vaultId ...string) (*DetectService, *fverrors.SkyflowError) {
	svc, err := s.Vault(vaultId...)
	if err != nil {
		return nil, err
	}
	return svc.detect, nil
}

// lookup finds the vault registered under vaultId, or the default vault when vaultId is
// empty. The caller holds s.mu.
func (s *Skyflow) lookup(vaultId string) (*VaultService, bool) {
	if vaultId == "" {
		if len(s.vaultOrder) == 0 {
			return nil, false
		}
		vaultId = s.vaultOrder[0]
	}
	svc, ok := s.vaultServices[vaultId]
	return svc, ok
}

// vaultIdForLog names the vault ID in a log line; an empty one asked for the default vault.
func vaultIdForLog(vaultId string) string {
	if vaultId == "" {
		return logs.DEFAULT_VAULT
	}
	return vaultId
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
	if err := validateVaultCredentials(cfg.Credentials); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.vaultServices[cfg.VaultId]; exists {
		fvlogger.Error(fmt.Sprintf(logs.VAULT_CONFIG_EXISTS, cfg.VaultId))
		return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_ALREADY_IN_CONFIG_LIST)
	}
	s.vaultServices[cfg.VaultId] = newVaultService(controller.NewVaultController(copyVaultConfig(cfg), s.clientSettings()))
	s.vaultOrder = append(s.vaultOrder, cfg.VaultId)
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
	if err := fvhelpers.ConfigValidator().ValidateHttpConfig(updatedConfig.HttpConfig()); err != nil {
		return err
	}
	if err := validateVaultCredentials(updatedConfig.Credentials); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	svc, exists := s.vaultServices[updatedConfig.VaultId]
	if !exists {
		fvlogger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, updatedConfig.VaultId))
		return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	// Only the fields set on updatedConfig change. The merged config replaces the old one in
	// a single step, so a call that is running keeps the config it started with.
	config := svc.controller.Config()
	credentialsChanged := !fvcommon.IsCredentialsEmpty(updatedConfig.Credentials)
	if credentialsChanged {
		config.Credentials = copyCredentials(updatedConfig.Credentials)
	}
	if updatedConfig.ClusterId != "" {
		config.ClusterId = updatedConfig.ClusterId
	}
	// A URL set through either field replaces both, so the deprecated BaseVaultURL is
	// honored and an older BaseVaultUrl can't take precedence over it (as in v2). A cluster
	// ID without a URL clears both too: a stored URL would take precedence over it.
	if updatedConfig.BaseVaultUrl != "" || updatedConfig.BaseVaultURL != "" || updatedConfig.ClusterId != "" {
		config.BaseVaultUrl = updatedConfig.BaseVaultUrl
		config.BaseVaultURL = updatedConfig.BaseVaultURL
	}
	// Env is always applied: it defaults to PROD, so an update that omits it
	// moves the vault to PROD.
	config.Env = updatedConfig.Env
	if updatedConfig.Timeout != 0 {
		config.Timeout = updatedConfig.Timeout
	}
	if updatedConfig.ConnectTimeout != 0 {
		config.ConnectTimeout = updatedConfig.ConnectTimeout
	}
	if updatedConfig.ReadTimeout != 0 {
		config.ReadTimeout = updatedConfig.ReadTimeout
	}
	if updatedConfig.WriteTimeout != 0 {
		config.WriteTimeout = updatedConfig.WriteTimeout
	}
	if updatedConfig.MaxRetries != nil {
		retries := *updatedConfig.MaxRetries
		config.MaxRetries = &retries
	}
	if updatedConfig.InitialRetryDelay != 0 {
		config.InitialRetryDelay = updatedConfig.InitialRetryDelay
	}
	if updatedConfig.MaxRetryDelay != 0 {
		config.MaxRetryDelay = updatedConfig.MaxRetryDelay
	}
	svc.controller.SetConfig(config, credentialsChanged)
	return nil
}

// validateVaultCredentials checks vault-level credentials when they are set. Unset
// credentials are allowed: the client-level credentials are used instead.
func validateVaultCredentials(credentials fvcommon.Credentials) *fverrors.SkyflowError {
	if fvcommon.IsCredentialsEmpty(credentials) {
		return nil
	}
	return fvvalidation.ValidateCredentials(credentials)
}

// RemoveVaultConfig removes the vault config registered under vaultId.
// Returns an error if the vault ID is not registered.
func (s *Skyflow) RemoveVaultConfig(vaultId string) *fverrors.SkyflowError {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.vaultServices[vaultId]; !exists {
		fvlogger.Error(fmt.Sprintf(logs.VAULT_CONFIG_DOES_NOT_EXIST, vaultId))
		return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	delete(s.vaultServices, vaultId)
	s.vaultOrder = slices.DeleteFunc(s.vaultOrder, func(id string) bool { return id == vaultId })
	return nil
}

// GetVaultConfig returns a copy of the VaultConfig registered under vaultId, or an error if
// it is not found. An empty vaultId means the default vault, the first one added. Use
// UpdateVaultConfig to change it.
func (s *Skyflow) GetVaultConfig(vaultId string) (*fvcommon.VaultConfig, *fverrors.SkyflowError) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, exists := s.lookup(vaultId)
	if !exists {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fverrors.VAULT_ID_NOT_IN_CONFIG_LIST)
	}
	config := copyVaultConfig(svc.controller.Config())
	return &config, nil
}

// ---------------------------------------------------------------------------
// Credentials
// ---------------------------------------------------------------------------

// AddSkyflowCredentials sets or replaces client-level credentials and propagates them to all vaults.
func (s *Skyflow) AddSkyflowCredentials(credentials fvcommon.Credentials) *fverrors.SkyflowError {
	fvlogger.Info(logs.VALIDATING_CRED)
	if err := fvvalidation.ValidateCredentials(credentials); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// A new pointer, so every vault drops its cached token.
	credentials = copyCredentials(credentials)
	s.credentials = &credentials
	s.syncVaults()
	return nil
}

// UpdateSkyflowCredentials replaces client-level credentials and propagates them to all vaults.
func (s *Skyflow) UpdateSkyflowCredentials(credentials fvcommon.Credentials) *fverrors.SkyflowError {
	return s.AddSkyflowCredentials(credentials)
}

// GetSkyflowCredentials returns the client-level credentials, or nil if none are set.
func (s *Skyflow) GetSkyflowCredentials() *fvcommon.Credentials {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.credentials == nil {
		return nil
	}
	credentials := copyCredentials(*s.credentials)
	return &credentials
}

// ---------------------------------------------------------------------------
// Logging
// ---------------------------------------------------------------------------

// SetLogLevel changes the log level for the client. The level is process-wide: it is shared by
// every client of the v2 and flowvault SDKs.
func (s *Skyflow) SetLogLevel(logLevel fvlogger.LogLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logLevel = logLevel
	fvlogger.SetLogLevel(logLevel)
	fvlogger.Info(fmt.Sprintf(logs.CURRENT_LOG_LEVEL, logLevel))
}

// GetLogLevel returns the level last set on this client. Another client, in either SDK,
// may have changed the process-wide level since.
func (s *Skyflow) GetLogLevel() fvlogger.LogLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logLevel
}

// UpdateLogLevel changes the log level for the client.
//
// Deprecated: Use SetLogLevel instead.
func (s *Skyflow) UpdateLogLevel(logLevel fvlogger.LogLevel) {
	s.SetLogLevel(logLevel)
}

// GetLoglevel returns the level last set on this client.
//
// Deprecated: Use GetLogLevel instead, which returns the level rather than a pointer to it.
func (s *Skyflow) GetLoglevel() *fvlogger.LogLevel {
	logLevel := s.GetLogLevel()
	return &logLevel
}
