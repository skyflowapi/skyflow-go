package helpers

import (
	"time"

	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
)

// profile describes v2 to the shared per-vault request plumbing used by the Detect V2
// operations. The HTTP defaults keep v2's behaviour: no overall timeout and no retries,
// unless the vault config sets them.
var profile = vaultclient.Profile{
	SDKPrefix:   constants.SDK_PREFIX,
	LogPrefix:   constants.SDK_LOG_PREFIX,
	Auth:        authSvc,
	ClusterURL:  GetURLWithEnv,
	SDKMetadata: CreateJsonMetadata,
	Defaults: vaultclient.HttpDefaults{
		InitialRetryDelay: 500 * time.Millisecond,
		MaxRetryDelay:     2 * time.Second,
	},
}

// Profile returns v2's settings for the shared common/vaultclient plumbing.
func Profile() vaultclient.Profile { return profile }
