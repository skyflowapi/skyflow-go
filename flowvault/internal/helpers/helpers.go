package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/skyflowapi/skyflow-go/common/auth"
	"github.com/skyflowapi/skyflow-go/common/config"
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
		"sdk_name_version": fmt.Sprintf("%s@%s", fvconsts.METRICS_SDK_NAME, fvconsts.SDK_VERSION),
		"device_model":     runtime.GOOS,
		"os_details":       fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		"runtime_details":  runtime.Version(),
	}
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(b)
}

// authSvc is the shared service-account authenticator configured for flowvault. Its token
// requests time out like Java's, so a token endpoint that stops answering can't block calls.
var authSvc = auth.New(auth.Config{
	SDKPrefix:  fvconsts.SDK_PREFIX,
	LogPrefix:  fvconsts.SDK_LOG_PREFIX,
	HTTPClient: &http.Client{Timeout: fvconsts.AUTH_TIMEOUT_SECONDS * time.Second},
})

// Auth returns the shared service-account authenticator configured for flowvault.
func Auth() *auth.Service { return authSvc }

// configValidator is the shared vault-config/credential validator configured for flowvault.
var configValidator = config.NewValidator(fvconsts.SDK_PREFIX, fvconsts.SDK_LOG_PREFIX)

// ConfigValidator returns the shared validator configured for flowvault.
func ConfigValidator() config.Validator { return configValidator }
