package config

import (
	"fmt"
	"net/url"
	"strings"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

const (
	apiKeyPrefix  = "sky-"
	apiKeyLength  = 42
	httpProtocol  = "http"
	httpsProtocol = "https"
)

// Validator validates the shared configuration types. The prefixes identify
// the SDK on whose behalf it validates, so messages carry that SDK's
// name/version.
type Validator struct {
	SDKPrefix string // prepended to error messages, e.g. "Skyflow Go SDK v2.1.0"
	LogPrefix string // prepended to log lines, e.g. "[ Skyflow Go SDK v2.1.0 ] "
}

func NewValidator(sdkPrefix, logPrefix string) Validator {
	return Validator{SDKPrefix: sdkPrefix, LogPrefix: logPrefix}
}

func (v Validator) errMsg(base string) string { return v.SDKPrefix + base }
func (v Validator) logMsg(base string) string { return v.LogPrefix + base }

func (v Validator) invalid(logBase, errBase string) *skyflowError.SkyflowError {
	logger.Error(v.logMsg(logBase))
	return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, v.errMsg(errBase))
}

// ValidateVaultConfig checks a vault config used to register a vault.
func (v Validator) ValidateVaultConfig(vaultConfig VaultConfig) *skyflowError.SkyflowError {
	if vaultConfig.VaultId == "" {
		return v.invalid(logVaultIdIsRequired, errInvalidVaultId)
	}
	baseVaultUrl := vaultConfig.BaseVaultUrl
	if vaultConfig.BaseVaultURL != "" {
		logger.Warn(v.logMsg(logDeprecatedFieldBaseVaultUrl))
		if baseVaultUrl == "" {
			baseVaultUrl = vaultConfig.BaseVaultURL
		}
	}
	if baseVaultUrl == "" {
		if vaultConfig.ClusterId == "" {
			return v.invalid(logClusterIdIsRequired, errInvalidClusterId)
		}
	} else if !isValidHTTPURL(baseVaultUrl) {
		return v.invalid(logVaultUrlIsInvalid, errInvalidVaultUrl)
	}
	return v.ValidateHttpConfig(vaultConfig.HttpConfig())
}

// ValidateHttpConfig checks that timeout, retry count and retry delays are not negative.
func (v Validator) ValidateHttpConfig(httpConfig HttpConfig) *skyflowError.SkyflowError {
	if httpConfig.Timeout < 0 || httpConfig.InitialRetryDelay < 0 || httpConfig.MaxRetryDelay < 0 ||
		(httpConfig.MaxRetries != nil && *httpConfig.MaxRetries < 0) {
		return v.invalid(logNegativeHttpConfig, errNegativeHttpConfig)
	}
	return nil
}

// ValidateUpdateVaultConfig checks a vault config used to update a registered vault.
func (v Validator) ValidateUpdateVaultConfig(vaultConfig VaultConfig) *skyflowError.SkyflowError {
	if vaultConfig.VaultId == "" {
		return v.invalid(logVaultIdIsRequired, errInvalidVaultId)
	}
	return nil
}

// ValidateCredentials checks that exactly one credential source is set and
// that api key and roles, when present, are well formed.
func (v Validator) ValidateCredentials(credentials Credentials) *skyflowError.SkyflowError {
	credPresent := 0
	for _, set := range []bool{
		credentials.Path != "",
		credentials.CredentialsString != "",
		credentials.Token != "",
		credentials.ApiKey != "",
	} {
		if set {
			credPresent++
		}
	}
	if credPresent > 1 {
		return v.invalid(logMultipleTokenGenerationMeansPassed, errMultipleTokenGenerationMeansPassed)
	} else if credPresent < 1 {
		return v.invalid(logNoTokenGenerationMeansPassed, errNoTokenGenerationMeansPassed)
	}

	if credentials.ApiKey != "" {
		if len(credentials.ApiKey) != apiKeyLength || !strings.Contains(credentials.ApiKey, apiKeyPrefix) {
			return v.invalid(logInvalidApiKey, errInvalidApiKey)
		}
	}

	if credentials.Roles != nil {
		if len(credentials.Roles) == 0 {
			return v.invalid(logEmptyRoles, errEmptyRoles)
		}
		for index, role := range credentials.Roles {
			if role == "" {
				logger.Error(fmt.Sprintf(v.logMsg(logEmptyOrNullRoleInRoles), index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, v.errMsg(errEmptyRoleInRoles))
			}
		}
	}
	return nil
}

// IsCredentialsEmpty reports whether no credential source is set.
func IsCredentialsEmpty(c Credentials) bool {
	return c.Path == "" && c.CredentialsString == "" && c.Token == "" && c.ApiKey == ""
}

func isValidHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != httpProtocol && u.Scheme != httpsProtocol {
		return false
	}
	return u.Host != ""
}
