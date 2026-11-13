package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

// apiKeyPattern is the format a Skyflow API key must have.
var apiKeyPattern = regexp.MustCompile(`^sky-[a-zA-Z0-9]{5}-[a-fA-F0-9]{32}$`)

const (
	httpProtocol  = "http"
	httpsProtocol = "https"
	localhostHost = "localhost"
)

// Validator validates the shared configuration types. The prefixes identify
// the SDK on whose behalf it validates, so messages carry that SDK's
// name/version.
type Validator struct {
	SDKPrefix string // prepended to error messages, e.g. "Skyflow Go SDK v2.1.0"
	LogPrefix string // prepended to log lines, e.g. "[ Skyflow Go SDK v2.1.0 ] "
	// Messages replaces some error messages with the SDK's own wording.
	Messages Messages
	// ErrorOptions are applied to every error the validator returns, e.g. errors.RawMessage().
	ErrorOptions []skyflowError.ErrorOption
	// AllowInsecureVaultUrl accepts a plain http vault URL for any host, not only loopback.
	AllowInsecureVaultUrl bool
}

// Messages are replacement error message bodies (without the SDK prefix) for an SDK whose
// wording differs, e.g. flowvault's. An empty field keeps the default message.
type Messages struct {
	InvalidVaultId    string // the vault ID is empty
	ClusterIdRequired string // neither a cluster ID nor a vault URL is set
	InvalidVaultUrl   string // the vault URL is not a valid http(s) URL
	InsecureVaultUrl  string // the vault URL uses http for a host that isn't loopback
}

// ValidatorOption configures a Validator built by NewValidator.
type ValidatorOption func(*Validator)

// WithMessages sets the Validator's replacement messages.
func WithMessages(messages Messages) ValidatorOption {
	return func(v *Validator) { v.Messages = messages }
}

// WithErrorOptions sets the options applied to every error the Validator returns.
func WithErrorOptions(opts ...skyflowError.ErrorOption) ValidatorOption {
	return func(v *Validator) { v.ErrorOptions = opts }
}

// WithInsecureVaultUrl makes the Validator accept a plain http vault URL for any host. It is
// for an SDK whose released versions accepted one; without it, http is allowed only for
// loopback hosts.
func WithInsecureVaultUrl() ValidatorOption {
	return func(v *Validator) { v.AllowInsecureVaultUrl = true }
}

func NewValidator(sdkPrefix, logPrefix string, opts ...ValidatorOption) Validator {
	v := Validator{SDKPrefix: sdkPrefix, LogPrefix: logPrefix}
	for _, opt := range opts {
		opt(&v)
	}
	return v
}

func (v Validator) errMsg(base string) string { return v.SDKPrefix + v.replacement(base) }
func (v Validator) logMsg(base string) string { return v.LogPrefix + base }

// replacement is the SDK's own wording for a default message body, when it set one.
func (v Validator) replacement(base string) string {
	replacement := ""
	switch base {
	case errInvalidVaultId:
		replacement = v.Messages.InvalidVaultId
	case errInvalidClusterId:
		replacement = v.Messages.ClusterIdRequired
	case errInvalidVaultUrl:
		replacement = v.Messages.InvalidVaultUrl
	case errInsecureVaultUrl:
		replacement = v.Messages.InsecureVaultUrl
	}
	if replacement == "" {
		return base
	}
	return replacement
}

func (v Validator) newError(message string) *skyflowError.SkyflowError {
	return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, message, v.ErrorOptions...)
}

func (v Validator) invalid(logBase, errBase string) *skyflowError.SkyflowError {
	logger.Error(v.logMsg(logBase))
	return v.newError(v.errMsg(errBase))
}

// ValidateVaultConfig checks a vault config used to register a vault.
func (v Validator) ValidateVaultConfig(vaultConfig VaultConfig) *skyflowError.SkyflowError {
	if vaultConfig.VaultId == "" {
		return v.invalid(logVaultIdIsRequired, errInvalidVaultId)
	}
	if vaultConfig.BaseVaultUrl == "" {
		if vaultConfig.ClusterId == "" {
			return v.invalid(logClusterIdIsRequired, errInvalidClusterId)
		}
	} else if err := v.validateBaseVaultUrl(vaultConfig.BaseVaultUrl); err != nil {
		return err
	}
	return v.ValidateHttpConfig(vaultConfig.HttpConfig())
}

// ValidateHttpConfig checks that the timeouts, retry count and retry delays are not negative.
func (v Validator) ValidateHttpConfig(httpConfig HttpConfig) *skyflowError.SkyflowError {
	if httpConfig.Timeout < 0 || httpConfig.InitialRetryDelay < 0 || httpConfig.MaxRetryDelay < 0 ||
		httpConfig.ConnectTimeout < 0 || httpConfig.ReadTimeout < 0 || httpConfig.WriteTimeout < 0 ||
		(httpConfig.MaxRetries != nil && *httpConfig.MaxRetries < 0) {
		return v.invalid(logNegativeHttpConfig, errNegativeHttpConfig)
	}
	return nil
}

// ValidateUpdateVaultConfig checks a vault config used to update a registered vault.
// Fields left unset keep their stored values, so only the ones that are set are checked.
func (v Validator) ValidateUpdateVaultConfig(vaultConfig VaultConfig) *skyflowError.SkyflowError {
	if vaultConfig.VaultId == "" {
		return v.invalid(logVaultIdIsRequired, errInvalidVaultId)
	}
	if vaultConfig.BaseVaultUrl != "" {
		return v.validateBaseVaultUrl(vaultConfig.BaseVaultUrl)
	}
	return nil
}

// validateBaseVaultUrl checks that the vault URL is well formed and uses https.
// Bearer tokens and sensitive data travel on it, so plain http is allowed only for
// loopback hosts (local development and tests), unless AllowInsecureVaultUrl is set.
func (v Validator) validateBaseVaultUrl(raw string) *skyflowError.SkyflowError {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != httpProtocol && u.Scheme != httpsProtocol) || u.Host == "" {
		return v.invalid(logVaultUrlIsInvalid, errInvalidVaultUrl)
	}
	if u.Scheme == httpProtocol && !v.AllowInsecureVaultUrl && !isLoopbackHost(u.Hostname()) {
		return v.invalid(logVaultUrlIsInsecure, errInsecureVaultUrl)
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
		if !apiKeyPattern.MatchString(credentials.ApiKey) {
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
				return v.newError(v.errMsg(errEmptyRoleInRoles))
			}
		}
	}
	return nil
}

// IsCredentialsEmpty reports whether no credential source is set.
func IsCredentialsEmpty(c Credentials) bool {
	return c.Path == "" && c.CredentialsString == "" && c.Token == "" && c.ApiKey == ""
}

// isLoopbackHost reports whether host is localhost or a loopback IP address.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, localhostHost) {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
