package vaultclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/skyflowapi/skyflow-go/common/auth"
	"github.com/skyflowapi/skyflow-go/common/config"
	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

const testApiKey = "sky-abcde-1234567890abcdef1234567890abcdef"

func testProfile() Profile {
	return Profile{
		SDKPrefix:   "Test SDK",
		LogPrefix:   "[ Test SDK ] ",
		Auth:        auth.New(auth.Config{SDKPrefix: "Test SDK", LogPrefix: "[ Test SDK ] "}),
		ClusterURL:  func(_ config.Env, clusterId string) string { return "https://" + clusterId + ".example.com" },
		SDKMetadata: func() string { return "meta" },
		Defaults:    HttpDefaults{Timeout: 60 * time.Second, InitialRetryDelay: 500 * time.Millisecond, MaxRetryDelay: 2 * time.Second},
	}
}

func TestResolveHttpConfigPrecedence(t *testing.T) {
	vaultRetries, clientRetries := 2, 5
	got := ResolveHttpConfig(
		config.HttpConfig{Timeout: 5 * time.Second, MaxRetries: &vaultRetries},
		config.HttpConfig{Timeout: 9 * time.Second, MaxRetries: &clientRetries, MaxRetryDelay: time.Second},
		testProfile().Defaults,
	)
	want := ResolvedHttpConfig{Timeout: 5 * time.Second, MaxRetries: 2, InitialRetryDelay: 500 * time.Millisecond, MaxRetryDelay: time.Second}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestBackoffStaysWithinJitterBounds(t *testing.T) {
	for attempt := 1; attempt <= 5; attempt++ {
		delay := backoff(attempt, 500*time.Millisecond, 2*time.Second)
		if delay < 400*time.Millisecond || delay > 2400*time.Millisecond {
			t.Fatalf("attempt %d: delay %v out of bounds", attempt, delay)
		}
	}
}

func TestRetryTransportResendsBody(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "payload" {
			t.Errorf("body = %q", body)
		}
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &retryTransport{
		base:   http.DefaultTransport,
		config: ResolvedHttpConfig{MaxRetries: 3, InitialRetryDelay: time.Millisecond, MaxRetryDelay: time.Millisecond},
		sleep:  func(time.Duration) {},
	}
	req, _ := http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	resp, err := transport.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("resp=%v err=%v", resp, err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestPrepareUsesApiKeyHeadersAndClusterURL(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v1", ClusterId: "c1"}, ClientSettings{Credentials: &config.Credentials{ApiKey: testApiKey}})
	call, err := b.Prepare("DEIDENTIFY_STRING", func(rc *RequestContext) {
		if rc.Operation != "DEIDENTIFY_STRING" {
			t.Errorf("operation = %q", rc.Operation)
		}
		rc.AddHeader(config.RequestIdHeader, "req-1")
	})
	if err != nil {
		t.Fatal(err)
	}
	if call.VaultID != "v1" || call.AuthToken != testApiKey || call.BaseURL != "https://c1.example.com" {
		t.Fatalf("unexpected call %+v", call)
	}
	if call.Headers.Get("x-request-id") != "req-1" || call.Headers.Get(MetricsHeader) != "meta" {
		t.Fatalf("headers = %v", call.Headers)
	}
}

func TestPrepareRejectsUnsupportedHeader(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v1", ClusterId: "c1"}, ClientSettings{Credentials: &config.Credentials{ApiKey: testApiKey}})
	_, err := b.Prepare("GET", func(rc *RequestContext) { rc.AddHeader("x-other", "v") })
	if err == nil || !strings.Contains(err.GetMessage(), "is not supported") {
		t.Fatalf("err = %v", err)
	}
}

func TestCredentialPrecedenceAndEnvFallback(t *testing.T) {
	vaultKey := "sky-vault-1234567890abcdef1234567890abcdefgh"
	b := New(testProfile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{ApiKey: vaultKey}},
		ClientSettings{Credentials: &config.Credentials{ApiKey: testApiKey}})
	if call, _ := b.Prepare("GET", nil); call.AuthToken != vaultKey {
		t.Fatalf("vault credentials should win, got %q", call.AuthToken)
	}

	t.Setenv(CredentialsEnv, "")
	_, err := New(testProfile(), config.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{}).Prepare("GET", nil)
	if err == nil || !strings.HasPrefix(err.GetMessage(), "Message: Test SDK Validation error. Invalid credentials.") {
		t.Fatalf("err = %v", err)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	// A JWT whose exp claim is in 2001.
	expired := "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjEwMDAwMDAwMDB9.c2ln"
	b := New(testProfile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{Token: expired}}, ClientSettings{})
	_, err := b.Prepare("GET", nil)
	if err == nil || !strings.Contains(err.GetMessage(), "Bearer token is expired") {
		t.Fatalf("err = %v", err)
	}
}

func TestAuthVersionChangesWithCredentials(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v"}, ClientSettings{})
	version := b.snapshot().authVersion
	b.SetClientSettings(ClientSettings{Credentials: &config.Credentials{ApiKey: testApiKey}})
	b.SetClientSettings(ClientSettings{Credentials: b.ClientSettings().Credentials, HttpConfig: config.HttpConfig{Timeout: time.Second}})
	if got := b.snapshot().authVersion; got != version+1 {
		t.Fatalf("after client credentials change: %d, want %d", got, version+1)
	}
	b.SetConfig(b.Config(), true)
	b.ResetAuth()
	if got := b.snapshot().authVersion; got != version+3 {
		t.Fatalf("after SetConfig and ResetAuth: %d, want %d", got, version+3)
	}
}

func TestClientHeadersComeBeforeInterceptorHeaders(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v1", ClusterId: "c1"}, ClientSettings{
		Credentials: &config.Credentials{ApiKey: testApiKey},
		Headers:     map[config.CustomHeaderKey]string{config.SkyflowAccountId: "acc", config.RequestIdHeader: "client"},
	})
	call, err := b.Prepare("GET", func(rc *RequestContext) { rc.AddHeader(config.RequestIdHeader, "call") })
	if err != nil {
		t.Fatal(err)
	}
	if call.Headers.Get("x-skyflow-account-id") != "acc" || call.Headers.Get("x-request-id") != "call" {
		t.Fatalf("headers = %v", call.Headers)
	}
}

func TestProfileHooksValidateClientHeadersAndReplaceTokenExpired(t *testing.T) {
	profile := testProfile()
	profile.ClientHeadersTag = "Client headers in"
	b := New(profile, config.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{
		Credentials: &config.Credentials{ApiKey: testApiKey},
		Headers:     map[config.CustomHeaderKey]string{"x-unknown": "v"},
	})
	if _, err := b.Prepare("GET", nil); err == nil || !strings.Contains(err.GetMessage(), "Client headers in") {
		t.Fatalf("client headers: err = %v", err)
	}

	var tags []string
	profile.ValidateHeaders = func(_ map[config.CustomHeaderKey]string, tag string) *skyflowError.SkyflowError {
		tags = append(tags, tag)
		return nil
	}
	b = New(profile, config.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{
		Credentials: &config.Credentials{ApiKey: testApiKey},
		Headers:     map[config.CustomHeaderKey]string{config.SkyflowAccountId: "acc"},
	})
	if _, err := b.Prepare("GET", func(rc *RequestContext) { rc.AddHeader(config.RequestIdHeader, "r") }); err != nil {
		t.Fatal(err)
	}
	if strings.Join(tags, ",") != "Client headers in,GET" {
		t.Fatalf("validator tags = %v", tags)
	}

	profile.ErrTokenExpired = "custom expired"
	expired := "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjEwMDAwMDAwMDB9.c2ln"
	b = New(profile, config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{Token: expired}}, ClientSettings{})
	if _, err := b.Prepare("GET", nil); err == nil || !strings.Contains(err.GetMessage(), "custom expired") {
		t.Fatalf("expired token: err = %v", err)
	}
}

func TestProfileLogsReplaceAndAddLines(t *testing.T) {
	var out bytes.Buffer
	logger.SetOutput(&out)
	logger.SetLogLevel(logger.INFO)
	defer func() {
		logger.SetOutput(os.Stderr)
		logger.SetLogLevel(logger.ERROR)
	}()

	profile := testProfile()
	profile.Logs = Logs{UsingApiKey: "own: using api key", DeprecatedBaseVaultURL: "own: deprecated"}
	b := New(profile, config.VaultConfig{VaultId: "v", BaseVaultURL: "https://x.example.com", Credentials: config.Credentials{ApiKey: testApiKey}}, ClientSettings{})
	if _, err := b.Prepare("GET", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "own: using api key") || !strings.Contains(out.String(), "own: deprecated") {
		t.Fatalf("logs = %q", out.String())
	}

	// Without own lines the optional ones are not logged.
	out.Reset()
	b = New(testProfile(), config.VaultConfig{VaultId: "v", BaseVaultURL: "https://x.example.com", Credentials: config.Credentials{ApiKey: testApiKey}}, ClientSettings{})
	if _, err := b.Prepare("GET", nil); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("logs = %q", out.String())
	}
}

func TestFirstDurationAndBackoffEdgeCases(t *testing.T) {
	if firstDuration() != 0 || firstDuration(0, -time.Second) != 0 {
		t.Fatal("firstDuration should be zero when no positive duration is set")
	}
	if delay := backoff(1, 10*time.Second, time.Second); delay < 800*time.Millisecond || delay > 1200*time.Millisecond {
		t.Fatalf("backoff not capped at max: %v", delay)
	}
	if backoff(3, 0, 0) != 0 || backoff(1, time.Nanosecond, time.Nanosecond) != time.Nanosecond {
		t.Fatal("backoff should skip jitter that rounds to zero")
	}
}

func newUnavailableServer(calls *int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
}

func TestRetryTransportDoesNotRetryABodyItCannotResend(t *testing.T) {
	var calls int32
	server := newUnavailableServer(&calls)
	defer server.Close()
	transport := &retryTransport{base: http.DefaultTransport, config: ResolvedHttpConfig{MaxRetries: 3}, sleep: func(time.Duration) {}}

	// A body that isn't a bytes/strings reader gets no GetBody.
	req, _ := http.NewRequest(http.MethodPost, server.URL, io.NopCloser(strings.NewReader("payload")))
	resp, err := transport.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusServiceUnavailable || atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("resp=%v err=%v calls=%d", resp, err, calls)
	}
	_ = resp.Body.Close()

	// GetBody that fails stops the retries too.
	req, _ = http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	req.GetBody = func() (io.ReadCloser, error) { return nil, errors.New("no body") }
	resp, err = transport.RoundTrip(req)
	if err != nil || atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	_ = resp.Body.Close()
}

func TestRetryTransportWaitsAndStopsWhenCancelled(t *testing.T) {
	var calls int32
	server := newUnavailableServer(&calls)
	defer server.Close()

	transport := &retryTransport{base: http.DefaultTransport, config: ResolvedHttpConfig{MaxRetries: 1, InitialRetryDelay: time.Millisecond, MaxRetryDelay: time.Millisecond}}
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil || atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	_ = resp.Body.Close()

	ctx, cancel := context.WithCancel(context.Background())
	transport.config = ResolvedHttpConfig{MaxRetries: 1, InitialRetryDelay: time.Hour, MaxRetryDelay: time.Hour}
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	done := make(chan error, 1)
	go func() {
		_, err := transport.RoundTrip(req)
		done <- err
	}()
	for atomic.LoadInt32(&calls) < 3 {
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestShouldRetry(t *testing.T) {
	for status, want := range map[int]bool{http.StatusOK: false, http.StatusBadRequest: false, http.StatusRequestTimeout: true, http.StatusTooManyRequests: true, http.StatusInternalServerError: true, http.StatusServiceUnavailable: true} {
		if ShouldRetry(status) != want {
			t.Errorf("ShouldRetry(%d) = %v", status, !want)
		}
	}
}

// Credentials are compared by value, so a copy with the same values keeps the token, and a
// change made to a struct the caller still holds is seen when it is set again.
func TestAuthVersionComparesCredentialsByValue(t *testing.T) {
	creds := &config.Credentials{ApiKey: testApiKey, Roles: []string{"r1"}}
	b := New(testProfile(), config.VaultConfig{VaultId: "v"}, ClientSettings{Credentials: creds})
	version := b.snapshot().authVersion

	b.SetClientSettings(ClientSettings{Credentials: &config.Credentials{ApiKey: testApiKey, Roles: []string{"r1"}}})
	b.SetConfig(b.Config(), false)
	if got := b.snapshot().authVersion; got != version {
		t.Fatalf("same values: %d, want %d", got, version)
	}
	creds.Roles[0] = "r2"
	b.SetClientSettings(ClientSettings{Credentials: creds})
	vaultConfig := b.Config()
	vaultConfig.Credentials = config.Credentials{Token: "t"}
	b.SetConfig(vaultConfig, false)
	if got := b.snapshot().authVersion; got != version+2 {
		t.Fatalf("changed client and vault credentials: %d, want %d", got, version+2)
	}
}

// Settings are copied on the way in and out: changing what was passed in or returned does
// not change the Base.
func TestSettingsAreCopiedInAndOut(t *testing.T) {
	retries := 2
	creds := &config.Credentials{ApiKey: testApiKey, Roles: []string{"r1"}}
	headers := map[config.CustomHeaderKey]string{config.SkyflowAccountId: "acc"}
	vaultConfig := config.VaultConfig{VaultId: "v", ClusterId: "c", MaxRetries: &retries, Credentials: config.Credentials{Roles: []string{"v1"}}}
	b := New(testProfile(), vaultConfig, ClientSettings{Credentials: creds, Headers: headers, HttpConfig: config.HttpConfig{MaxRetries: &retries}})

	creds.ApiKey, creds.Roles[0] = "changed", "changed"
	headers[config.SkyflowAccountId] = "changed"
	retries = 9
	vaultConfig.Credentials.Roles[0] = "changed"

	got := b.ClientSettings()
	if got.Credentials.ApiKey != testApiKey || got.Credentials.Roles[0] != "r1" || got.Headers[config.SkyflowAccountId] != "acc" || *got.HttpConfig.MaxRetries != 2 {
		t.Fatalf("client settings changed through the caller's values: %+v", got)
	}
	gotConfig := b.Config()
	if *gotConfig.MaxRetries != 2 || gotConfig.Credentials.Roles[0] != "v1" {
		t.Fatalf("vault config changed through the caller's values: %+v", gotConfig)
	}

	got.Credentials.ApiKey, got.Headers[config.SkyflowAccountId] = "changed", "changed"
	gotConfig.Credentials.Roles[0], *gotConfig.MaxRetries = "changed", 9
	if again := b.ClientSettings(); again.Credentials.ApiKey != testApiKey || again.Headers[config.SkyflowAccountId] != "acc" {
		t.Fatalf("client settings changed through a returned copy: %+v", again)
	}
	if again := b.Config(); again.Credentials.Roles[0] != "v1" || *again.MaxRetries != 2 {
		t.Fatalf("vault config changed through a returned copy: %+v", again)
	}
}

func TestRefreshAuthUsesTheCurrentCredentials(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{Credentials: &config.Credentials{ApiKey: testApiKey}})
	call, err := b.Prepare("DEIDENTIFY_FILE", nil)
	if err != nil {
		t.Fatal(err)
	}
	newKey := "sky-zzzzz-1234567890abcdef1234567890abcdef"
	b.SetClientSettings(ClientSettings{Credentials: &config.Credentials{ApiKey: newKey}})
	refreshed, err := b.RefreshAuth(context.Background(), call)
	if err != nil || refreshed.AuthToken != newKey || refreshed.BaseURL != call.BaseURL || call.AuthToken != testApiKey {
		t.Fatalf("refreshed = %+v, err = %v, original = %+v", refreshed, err, call)
	}

	t.Setenv(CredentialsEnv, "")
	b.SetClientSettings(ClientSettings{})
	if _, err := b.RefreshAuth(context.Background(), call); err == nil {
		t.Fatal("no credentials should fail")
	}
}
