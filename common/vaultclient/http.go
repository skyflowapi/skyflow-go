package vaultclient

import (
	"context"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/skyflowapi/skyflow-go/common/config"
)

// retryJitterFactor spreads each retry delay by +/-20%.
const retryJitterFactor = 0.2

// Connection pool settings of the transports built for per-attempt timeouts; the same as
// http.DefaultTransport's.
const (
	defaultMaxIdleConns        = 100
	defaultIdleConnTimeout     = 90 * time.Second
	defaultTLSHandshakeTimeout = 10 * time.Second
	defaultKeepAlive           = 30 * time.Second
)

// HttpDefaults are an SDK's timeout and retry settings, used when neither the vault nor
// the client sets a value.
type HttpDefaults struct {
	Timeout           time.Duration
	ConnectTimeout    time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxRetries        int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
}

// ResolvedHttpConfig is the effective timeout and retry policy for a call.
type ResolvedHttpConfig struct {
	Timeout           time.Duration
	ConnectTimeout    time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxRetries        int
	InitialRetryDelay time.Duration
	MaxRetryDelay     time.Duration
}

// ResolveHttpConfig picks each setting from the vault config, then the client config,
// then the SDK default.
func ResolveHttpConfig(vault, client config.HttpConfig, defaults HttpDefaults) ResolvedHttpConfig {
	resolved := ResolvedHttpConfig{
		Timeout:           firstDuration(vault.Timeout, client.Timeout, defaults.Timeout),
		ConnectTimeout:    firstDuration(vault.ConnectTimeout, client.ConnectTimeout, defaults.ConnectTimeout),
		ReadTimeout:       firstDuration(vault.ReadTimeout, client.ReadTimeout, defaults.ReadTimeout),
		WriteTimeout:      firstDuration(vault.WriteTimeout, client.WriteTimeout, defaults.WriteTimeout),
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
// including retries; ConnectTimeout, ReadTimeout and WriteTimeout apply to each attempt.
func NewHTTPClient(resolved ResolvedHttpConfig) *http.Client {
	return &http.Client{
		Timeout: resolved.Timeout,
		Transport: &retryTransport{
			base:   transportFor(resolved),
			config: resolved,
		},
	}
}

// newCallClient is NewHTTPClient with the 401 retry: a request the vault rejects because its
// bearer token expired is sent once more with a new token (see refreshAfterUnauthorized).
func (b *Base) newCallClient(resolved ResolvedHttpConfig) *http.Client {
	return &http.Client{
		Timeout: resolved.Timeout,
		Transport: &retryTransport{
			base:   &unauthorizedTransport{base: transportFor(resolved), refresh: b.refreshAfterUnauthorized},
			config: resolved,
		},
	}
}

// SignedURLClient returns the client for uploads straight to signed URLs, as the Java and
// Python SDKs do: it shares api's connections and per-attempt connect, read and write
// timeouts, but makes a single attempt and has no overall timeout, because a large file can
// take longer than one API call's timeout while still uploading. The caller's context can
// cancel an upload at any time.
func SignedURLClient(api *http.Client) *http.Client {
	transport := api.Transport
	if retrying, ok := transport.(*retryTransport); ok {
		transport = retrying.base
	}
	if unauthorized, ok := transport.(*unauthorizedTransport); ok {
		transport = unauthorized.base // no vault token is sent to a signed URL
	}
	return &http.Client{Transport: transport}
}

// transportKey identifies the per-attempt timeouts a shared transport applies.
type transportKey struct {
	connect, read, write time.Duration
}

// transports holds one transport per set of per-attempt timeouts, so calls with the same
// settings share connections. A client is built per call, so transports must not be.
var transports sync.Map // transportKey -> *http.Transport

// transportFor returns the shared transport for the resolved per-attempt timeouts; with
// none set it is http.DefaultTransport.
func transportFor(resolved ResolvedHttpConfig) http.RoundTripper {
	key := transportKey{connect: resolved.ConnectTimeout, read: resolved.ReadTimeout, write: resolved.WriteTimeout}
	if key == (transportKey{}) {
		return http.DefaultTransport
	}
	if cached, ok := transports.Load(key); ok {
		return cached.(*http.Transport)
	}
	cached, _ := transports.LoadOrStore(key, newTransport(key))
	return cached.(*http.Transport)
}

// newTransport is http.DefaultTransport's configuration with the per-attempt timeouts:
// connect bounds dialing, read bounds the wait for the response headers once the request
// is written, and write bounds each write to the connection.
func newTransport(key transportKey) *http.Transport {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          defaultMaxIdleConns,
		IdleConnTimeout:       defaultIdleConnTimeout,
		TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
		ExpectContinueTimeout: time.Second,
		ResponseHeaderTimeout: key.read,
	}
	dialer := &net.Dialer{Timeout: key.connect, KeepAlive: defaultKeepAlive}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := dialer.DialContext(ctx, network, address)
		if err != nil || key.write == 0 {
			return conn, err
		}
		return &writeDeadlineConn{Conn: conn, timeout: key.write}, nil
	}
	return transport
}

// writeDeadlineConn bounds each Write. Reads are left alone, so idle pooled connections
// are not closed by a deadline.
type writeDeadlineConn struct {
	net.Conn
	timeout time.Duration
}

func (c *writeDeadlineConn) Write(p []byte) (int, error) {
	if err := c.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return 0, err
	}
	return c.Conn.Write(p)
}

// retryTransport retries 408, 429 and 5xx responses with exponential backoff and jitter.
type retryTransport struct {
	base   http.RoundTripper
	config ResolvedHttpConfig
	sleep  func(time.Duration) // overridden in tests
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	for retry := 0; err == nil && ShouldRetryRequest(req, resp.StatusCode) && retry < t.config.MaxRetries; retry++ {
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

// noServerErrorRetryKey marks a context whose requests are not retried on a 5xx.
type noServerErrorRetryKey struct{}

// WithoutServerErrorRetry returns a context whose requests are retried on 408 and 429 but
// not on a 5xx, for a call that isn't safe to repeat once the server may have acted on it.
// For example, the Java SDK doesn't retry FlowVault's uploadFiles on a 5xx: it creates a
// record for every entry without a SkyflowId, so a retry could create duplicates. 408 and
// 429 mean the request was not processed, so those are still retried.
func WithoutServerErrorRetry(ctx context.Context) context.Context {
	return context.WithValue(ctx, noServerErrorRetryKey{}, true)
}

// ShouldRetryRequest is ShouldRetry, except that a request whose context comes from
// WithoutServerErrorRetry is not retried on a 5xx.
func ShouldRetryRequest(req *http.Request, status int) bool {
	if status >= http.StatusInternalServerError && req.Context().Value(noServerErrorRetryKey{}) != nil {
		return false
	}
	return ShouldRetry(status)
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
