package vaultclient

import (
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/skyflowapi/skyflow-go/common/config"
)

// retryJitterFactor spreads each retry delay by +/-20%.
const retryJitterFactor = 0.2

// HttpDefaults are an SDK's timeout and retry settings, used when neither the vault nor
// the client sets a value.
type HttpDefaults struct {
	Timeout           time.Duration
	MaxRetries        int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
}

// ResolvedHttpConfig is the effective timeout and retry policy for a call.
type ResolvedHttpConfig struct {
	Timeout           time.Duration
	MaxRetries        int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
}

// ResolveHttpConfig picks each setting from the vault config, then the client config,
// then the SDK default.
func ResolveHttpConfig(vault, client config.HttpConfig, defaults HttpDefaults) ResolvedHttpConfig {
	resolved := ResolvedHttpConfig{
		Timeout:           firstDuration(vault.Timeout, client.Timeout, defaults.Timeout),
		MaxRetries:        defaults.MaxRetries,
		InitialRetryDelay: firstDuration(vault.InitialRetryDelay, client.InitialRetryDelay, defaults.InitialRetryDelay),
		MaxRetryDelay:     firstDuration(vault.MaxRetryDelay, client.MaxRetryDelay, defaults.MaxRetryDelay),
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
func NewHTTPClient(resolved ResolvedHttpConfig) *http.Client {
	return &http.Client{
		Timeout: resolved.Timeout,
		Transport: &retryTransport{
			base:   http.DefaultTransport,
			config: resolved,
		},
	}
}

// retryTransport retries 408, 429 and 5xx responses with exponential backoff and jitter.
type retryTransport struct {
	base   http.RoundTripper
	config ResolvedHttpConfig
	sleep  func(time.Duration) // overridden in tests
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	for retry := 0; err == nil && shouldRetry(resp.StatusCode) && retry < t.config.MaxRetries; retry++ {
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

func shouldRetry(status int) bool {
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
	jitter := time.Duration(float64(delay) * retryJitterFactor)
	if jitter <= 0 {
		return delay
	}
	return delay - jitter + time.Duration(rand.Int63n(int64(2*jitter)+1)) //nolint:gosec // jitter, not security
}
