package helpers

import (
	"time"

	cfg "github.com/skyflowapi/skyflow-go/common/config"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	common "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

// profile describes v2 to the shared per-vault request plumbing used by the Detect
// operations. The HTTP defaults are a 60s overall timeout, 2 retries and no per-attempt
// connect, read or write limit. A SkyVault vault config can't change them.
var profile = vaultclient.Profile{
	SDKPrefix:   constants.SDK_PREFIX,
	LogPrefix:   constants.SDK_LOG_PREFIX,
	Auth:        authSvc,
	ClusterURL:  GetURLWithEnv,
	SDKMetadata: CreateJsonMetadata,
	Defaults: vaultclient.HttpDefaults{
		Timeout:           60 * time.Second,
		MaxRetries:        2,
		InitialRetryDelay: 500 * time.Millisecond,
		MaxRetryDelay:     2 * time.Second,
	},
}

// Profile returns v2's settings for the shared common/vaultclient plumbing.
func Profile() vaultclient.Profile { return profile }

// CommonVaultConfig is v2's vault config as the shared common/config type, which the shared
// validator and common/vaultclient take. Its VaultUrl is BaseVaultUrl, falling back to the
// deprecated BaseVaultURL. Its timeout and retry fields stay unset, so the profile's
// defaults apply.
func CommonVaultConfig(c common.VaultConfig) cfg.VaultConfig {
	vaultUrl := c.BaseVaultUrl
	if vaultUrl == "" {
		vaultUrl = c.BaseVaultURL //nolint:staticcheck // the deprecated field is still honoured
	}
	return cfg.VaultConfig{
		VaultId:     c.VaultId,
		ClusterId:   c.ClusterId,
		VaultUrl:    vaultUrl,
		Env:         c.Env,
		Credentials: c.Credentials,
	}
}
