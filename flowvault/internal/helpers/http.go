package helpers

import (
	"io"
	"math/rand"
	"net"
	"net/http"
	"time"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

// ResolvedHttpConfig is the effective timeout and retry policy for a call.
type ResolvedHttpConfig struct {
	Timeout           time.Duration
	MaxRetries        int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
}

// ResolveHttpConfig picks each setting from the vault config, then the client config,
// then the SDK default.
func ResolveHttpConfig(vault, client fvcommon.HttpConfig) ResolvedHttpConfig {
	resolved := ResolvedHttpConfig{
		Timeout:           firstDuration(vault.Timeout, client.Timeout, fvconsts.DEFAULT_TIMEOUT_SECONDS*time.Second),
		MaxRetries:        fvconsts.DEFAULT_MAX_RETRIES,
		InitialRetryDelay: firstDuration(vault.InitialRetryDelay, client.InitialRetryDelay, fvconsts.DEFAULT_INITIAL_RETRY_DELAY_MILLIS*time.Millisecond),
		MaxRetryDelay:     firstDuration(vault.MaxRetryDelay, client.MaxRetryDelay, fvconsts.DEFAULT_MAX_RETRY_DELAY_MILLIS*time.Millisecond),
	}
	if vault.MaxRetries != nil {
		resolved.MaxRetries = *vault.MaxRetries
	} else if client.MaxRetries != nil {
		resolved.MaxRetries = *client.MaxRetries
	}
	return resolved
}

func firstDuration(values ...time.Duration) time.Duration {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

// NewHTTPClient builds the client used for API calls. Timeout covers the whole call,
// including retries.
func NewHTTPClient(config ResolvedHttpConfig) *http.Client {
	return &http.Client{
		Timeout: config.Timeout,
		Transport: &retryTransport{
			base:   http.DefaultTransport,
			config: config,
		},
	}
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

// retryTransport retries 408, 429 and 5xx responses with exponential backoff and jitter.
type retryTransport struct {
	base   http.RoundTripper
	config ResolvedHttpConfig
	sleep  func(time.Duration) // overridden in tests
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	for retry := 0; err == nil && ShouldRetry(resp.StatusCode) && retry < t.config.MaxRetries; retry++ {
		if req.Body != nil && req.GetBody == nil {
			break
		}
		retryReq := req.Clone(req.Context())
		if req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				break
			}
			retryReq.Body = body
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if waitErr := t.wait(req, backoff(retry+1, t.config.InitialRetryDelay, t.config.MaxRetryDelay)); waitErr != nil {
			return nil, waitErr
		}
		resp, err = t.base.RoundTrip(retryReq)
	}
	return resp, err
}

func (t *retryTransport) wait(req *http.Request, delay time.Duration) error {
	if t.sleep != nil {
		t.sleep(delay)
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-req.Context().Done():
		return req.Context().Err()
	case <-timer.C:
		return nil
	}
}

// ShouldRetry reports whether a response with this status is retried: 408, 429 and 5xx.
func ShouldRetry(status int) bool {
	return status == http.StatusRequestTimeout || status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

// backoff doubles the delay per attempt up to max, then applies +/-20% jitter.
func backoff(attempt int, initial, max time.Duration) time.Duration {
	delay := initial
	for step := 1; step < attempt && delay < max; step++ {
		delay *= 2
	}
	if delay > max {
		delay = max
	}
	jitter := time.Duration(float64(delay) * fvconsts.RETRY_JITTER_FACTOR)
	if jitter <= 0 {
		return delay
	}
	return delay - jitter + time.Duration(rand.Int63n(int64(2*jitter)+1)) //nolint:gosec // jitter, not security
}
