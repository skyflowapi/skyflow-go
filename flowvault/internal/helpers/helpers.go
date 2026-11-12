package helpers

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/skyflowapi/skyflow-go/common/auth"
	"github.com/skyflowapi/skyflow-go/common/config"
	commonerrors "github.com/skyflowapi/skyflow-go/common/errors"
	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

// GetURLWithEnv builds the vault base URL from environment and cluster ID.
// Mirrors v2's helper but lives here since v2/internal is module-private.
func GetURLWithEnv(env fvcommon.Env, clusterId string) string {
	base := fvconsts.SECURE_PROTOCOL + clusterId
	switch env {
	case fvcommon.DEV:
		return base + fvconsts.DEV_DOMAIN
	case fvcommon.STAGE:
		return base + fvconsts.STAGE_DOMAIN
	case fvcommon.SANDBOX:
		return base + fvconsts.SANDBOX_DOMAIN
	default:
		return base + fvconsts.PROD_DOMAIN
	}
}

// CreateSDKMetadata builds the sky-metadata header value.
func CreateSDKMetadata() string {
	data := map[string]string{
		fvconsts.SDK_METADATA_KEY_NAME_VERSION:    fmt.Sprintf("%s@%s", fvconsts.METRICS_SDK_NAME, fvconsts.SDK_VERSION),
		fvconsts.SDK_METADATA_KEY_DEVICE_MODEL:    runtime.GOOS,
		fvconsts.SDK_METADATA_KEY_OS_DETAILS:      fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		fvconsts.SDK_METADATA_KEY_RUNTIME_DETAILS: runtime.Version(),
	}
	// Marshaling a map[string]string can't fail.
	b, _ := json.Marshal(data)
	return string(b)
}

// authSvc is the shared service-account authenticator configured for flowvault. Its token
// requests time out after auth.DefaultTimeout, like Java's.
var authSvc = auth.New(auth.Config{
	SDKPrefix:    fvconsts.SDK_PREFIX,
	LogPrefix:    fvconsts.SDK_LOG_PREFIX,
	ErrorOptions: []commonerrors.ErrorOption{commonerrors.RawMessage()},
})

// Auth returns the shared service-account authenticator configured for flowvault.
func Auth() *auth.Service { return authSvc }

// Vault-config messages in the Java SDK's wording, which the shared validator uses for flowvault
// instead of its defaults (the v2 SDK keeps those).
const (
	javaEmptyVaultId          = " Initialization failed. Invalid vault ID. Vault ID must not be empty."
	javaClusterIdOrVaultUrl   = " Initialization failed. Specify either 'clusterId' or 'vaultURL'."
	javaInvalidVaultUrlFormat = " Initialization failed. Vault URL must start with 'https://'."
)

// configValidator is the shared vault-config/credential validator configured for flowvault:
// Java's messages, and raw-message errors like every other flowvault error.
var configValidator = config.NewValidator(fvconsts.SDK_PREFIX, fvconsts.SDK_LOG_PREFIX,
	config.WithMessages(config.Messages{
		InvalidVaultId:    javaEmptyVaultId,
		ClusterIdRequired: javaClusterIdOrVaultUrl,
		InvalidVaultUrl:   javaInvalidVaultUrlFormat,
		InsecureVaultUrl:  javaInvalidVaultUrlFormat,
	}),
	config.WithErrorOptions(commonerrors.RawMessage()),
)

// ConfigValidator returns the shared validator configured for flowvault.
func ConfigValidator() config.Validator { return configValidator }
