package helpers

import (
	"net/http"
	"time"

	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

// ResolvedHttpConfig is the effective timeout and retry policy for a call.
type ResolvedHttpConfig = vaultclient.ResolvedHttpConfig

// httpDefaults apply when neither the vault nor the client sets a value.
var httpDefaults = vaultclient.HttpDefaults{
	Timeout:           fvconsts.DEFAULT_TIMEOUT_SECONDS * time.Second,
	ConnectTimeout:    fvconsts.DEFAULT_CONNECT_TIMEOUT_SECONDS * time.Second,
	ReadTimeout:       fvconsts.DEFAULT_READ_TIMEOUT_SECONDS * time.Second,
	WriteTimeout:      fvconsts.DEFAULT_WRITE_TIMEOUT_SECONDS * time.Second,
	MaxRetries:        fvconsts.DEFAULT_MAX_RETRIES,
	InitialRetryDelay: fvconsts.DEFAULT_INITIAL_RETRY_DELAY_MILLIS * time.Millisecond,
	MaxRetryDelay:     fvconsts.DEFAULT_MAX_RETRY_DELAY_MILLIS * time.Millisecond,
}

// ResolveHttpConfig picks each setting from the vault config, then the client config,
// then the SDK default.
func ResolveHttpConfig(vault, client fvcommon.HttpConfig) ResolvedHttpConfig {
	return vaultclient.ResolveHttpConfig(vault, client, httpDefaults)
}

// NewHTTPClient builds the client used for API calls. Timeout covers the whole call,
// including retries.
func NewHTTPClient(config ResolvedHttpConfig) *http.Client {
	return vaultclient.NewHTTPClient(config)
}
