package helpers

import (
	"time"

	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
)

// profile describes v2 to the shared per-vault request plumbing used by the Detect
// operations. The HTTP defaults match the Java and Python SkyVault SDKs' Detect clients: a
// 60s overall timeout, 2 retries and no per-attempt connect, read or write limit, unless the
// vault config sets them.
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
