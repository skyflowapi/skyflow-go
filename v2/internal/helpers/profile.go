package helpers

import (
	"time"

	cfg "github.com/skyflowapi/skyflow-go/common/config"
	sa "github.com/skyflowapi/skyflow-go/common/serviceaccount"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	common "github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
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

// The shared option types in common/serviceaccount hold only the current API. v2 keeps its
// v2.1.0 fields, including the deprecated ones, on its own types and converts them here, so
// the deprecated fields exist in v2 alone.

// CommonBearerTokenOptions converts v2's bearer token options to the shared ones. The
// deprecated RoleIDs are used, with v2's deprecation warning, when RoleIds is empty.
func CommonBearerTokenOptions(o common.BearerTokenOptions) sa.BearerTokenOptions {
	roleIds := o.RoleIds
	if len(o.RoleIDs) > 0 { //nolint:staticcheck // the deprecated field is still honoured
		logger.Warn(logs.DEPRECATED_FIELD_ROLE_IDS)
		if len(roleIds) == 0 {
			roleIds = o.RoleIDs //nolint:staticcheck // the deprecated field is still honoured
		}
	}
	return sa.BearerTokenOptions{Ctx: o.Ctx, RoleIds: roleIds}
}

// CommonSignedDataTokensOptions converts v2's signed data token options to the shared ones.
func CommonSignedDataTokensOptions(o common.SignedDataTokensOptions) sa.SignedDataTokensOptions {
	return sa.SignedDataTokensOptions{DataTokens: o.DataTokens, TimeToLive: o.TimeToLive, Ctx: o.Ctx}
}
