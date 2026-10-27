package helpers

import (
	"net"
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

// uploadClient is shared by all uploads so they reuse one connection pool.
var uploadClient = newUploadHTTPClient()

// UploadHTTPClient returns the client that PUTs files to signed URLs. Like the Java and
// Python SDKs, it doesn't retry and has no overall timeout, because a large file can take
// longer than one API call's timeout while still uploading. Connecting and waiting for the
// storage service's response are bounded, so a stalled connection still fails, and the
// caller's context can cancel an upload at any time.
func UploadHTTPClient() *http.Client {
	return uploadClient
}

func newUploadHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   fvconsts.UPLOAD_CONNECT_TIMEOUT_SECONDS * time.Second,
		KeepAlive: fvconsts.UPLOAD_KEEP_ALIVE_SECONDS * time.Second,
	}).DialContext
	transport.TLSHandshakeTimeout = fvconsts.UPLOAD_CONNECT_TIMEOUT_SECONDS * time.Second
	transport.ResponseHeaderTimeout = fvconsts.UPLOAD_RESPONSE_HEADER_TIMEOUT_SECONDS * time.Second
	return &http.Client{Transport: transport}
}
