package vaultclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/skyflowapi/skyflow-go/common/config"
)

func tokenExpiringIn(t *testing.T, d time.Duration) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"exp": time.Now().Add(d).Unix()}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// vaultRejecting answers 401 to a request carrying rejected and 200 to any other, recording
// each request's token and body.
type vaultRejecting struct {
	server *httptest.Server
	tokens []string
	bodies []string
}

func newVaultRejecting(t *testing.T, rejected string) *vaultRejecting {
	v := &vaultRejecting{}
	v.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		token := strings.TrimPrefix(r.Header.Get(headerAuthorization), bearerPrefix)
		v.tokens, v.bodies = append(v.tokens, token), append(v.bodies, string(body))
		if token == rejected {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(v.server.Close)
	return v
}

// baseWithCachedToken is a Base on service-account credentials whose cached token is cached.
func baseWithCachedToken(f *fakeAuth, cached string) *Base {
	b := New(f.profile(), config.VaultConfig{VaultId: "v", Credentials: config.Credentials{CredentialsString: f.creds}}, ClientSettings{})
	b.storeToken(cached, b.snapshot().authVersion)
	return b
}

func post(t *testing.T, client *http.Client, url, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(headerAuthorization, bearerPrefix+token)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	return resp
}

// As in Java: a 401 for a token at or near expiry is resent once, body included, with a new
// token, which is then cached for later calls.
func TestUnauthorizedIsResentWithANewTokenWhenTheTokenExpired(t *testing.T) {
	f := newFakeAuth(t, nil)
	rejected := tokenExpiringIn(t, 30*time.Second)
	vault := newVaultRejecting(t, rejected)
	b := baseWithCachedToken(f, rejected)

	resp := post(t, b.newCallClient(ResolvedHttpConfig{}), vault.server.URL, rejected)
	if resp.StatusCode != http.StatusOK || len(vault.tokens) != 2 || vault.tokens[1] == rejected || vault.bodies[1] != `{"a":1}` {
		t.Fatalf("status = %d, tokens = %d, bodies = %q", resp.StatusCode, len(vault.tokens), vault.bodies)
	}
	if b.token != vault.tokens[1] || f.calls.Load() != 1 {
		t.Fatalf("the new token should be cached; generations = %d", f.calls.Load())
	}
	// A call that isn't rejected goes through once.
	if resp := post(t, b.newCallClient(ResolvedHttpConfig{}), vault.server.URL, b.token); resp.StatusCode != http.StatusOK || len(vault.tokens) != 3 {
		t.Fatalf("status = %d, requests = %d", resp.StatusCode, len(vault.tokens))
	}
}

func TestUnauthorizedIsReturnedWhenRetryingCannotHelp(t *testing.T) {
	f := newFakeAuth(t, nil)
	notDue := tokenExpiringIn(t, time.Hour)
	vault := newVaultRejecting(t, notDue)

	// Not near expiry: the 401 has another cause.
	b := baseWithCachedToken(f, notDue)
	if resp := post(t, b.newCallClient(ResolvedHttpConfig{}), vault.server.URL, notDue); resp.StatusCode != http.StatusUnauthorized || len(vault.tokens) != 1 {
		t.Fatalf("not due: status = %d, requests = %d", resp.StatusCode, len(vault.tokens))
	}

	// An API key or a static token can't be replaced.
	for _, creds := range []config.Credentials{{ApiKey: testApiKey}, {Token: tokenExpiringIn(t, 30*time.Second)}} {
		b := New(f.profile(), config.VaultConfig{VaultId: "v", Credentials: creds}, ClientSettings{})
		if got := b.refreshAfterUnauthorized(context.Background(), tokenExpiringIn(t, 30*time.Second)); got != "" {
			t.Fatalf("%+v: got %q", creds, got)
		}
	}
	// No credentials at all.
	t.Setenv(CredentialsEnv, "")
	if got := New(f.profile(), config.VaultConfig{VaultId: "v"}, ClientSettings{}).refreshAfterUnauthorized(context.Background(), "x"); got != "" {
		t.Fatalf("no credentials: got %q", got)
	}
	if f.calls.Load() != 0 {
		t.Fatalf("no token should be generated; got %d", f.calls.Load())
	}
}

func TestUnauthorizedReusesATokenAnotherCallAlreadyReplaced(t *testing.T) {
	f := newFakeAuth(t, nil)
	rejected := tokenExpiringIn(t, 30*time.Second)
	b := baseWithCachedToken(f, "replaced")
	if got := b.refreshAfterUnauthorized(context.Background(), rejected); got != "replaced" || f.calls.Load() != 0 {
		t.Fatalf("got %q, generations = %d", got, f.calls.Load())
	}
}

// A failed regeneration is not repeated for the same token; caching a new token clears that.
func TestUnauthorizedRegenerationFailureIsShared(t *testing.T) {
	f := newFakeAuth(t, nil)
	rejected := tokenExpiringIn(t, 30*time.Second)
	b := New(f.profile(), config.VaultConfig{VaultId: "v", Credentials: config.Credentials{CredentialsString: "not json"}}, ClientSettings{})
	b.storeToken(rejected, b.snapshot().authVersion)
	if got := b.refreshAfterUnauthorized(context.Background(), rejected); got != "" || b.refreshFailedFor != rejected {
		t.Fatalf("got %q, failed for %q", got, b.refreshFailedFor)
	}
	if got := b.refreshAfterUnauthorized(context.Background(), rejected); got != "" {
		t.Fatalf("second try: got %q", got)
	}
	b.storeToken("next", b.snapshot().authVersion)
	if b.refreshFailedFor != "" {
		t.Fatal("caching a token should clear the failure")
	}
}

func TestUnauthorizedStopsWhenTheContextEnds(t *testing.T) {
	f := newFakeAuth(t, nil)
	rejected := tokenExpiringIn(t, 30*time.Second)
	b := baseWithCachedToken(f, rejected)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := b.refreshAfterUnauthorized(ctx, rejected); got != "" {
		t.Fatalf("ended context: got %q", got)
	}

	// The context ends while another call holds the token lock.
	b.tokenLock <- struct{}{}
	waiting, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	if got := b.refreshAfterUnauthorized(waiting, rejected); got != "" {
		t.Fatalf("cancelled wait: got %q", got)
	}
	<-b.tokenLock
}

type failingRoundTripper struct{ calls atomic.Int32 }

func (f *failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	f.calls.Add(1)
	return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(strings.NewReader(""))}, nil
}

// A request whose body can't be sent again is not resent.
func TestUnauthorizedTransportNeedsAReplayableBody(t *testing.T) {
	base := &failingRoundTripper{}
	transport := &unauthorizedTransport{base: base, refresh: func(context.Context, string) string { return "fresh" }}

	req, _ := http.NewRequest(http.MethodPost, "http://vault.test", io.NopCloser(strings.NewReader("x")))
	if resp, err := transport.RoundTrip(req); err != nil || resp.StatusCode != http.StatusUnauthorized || base.calls.Load() != 1 {
		t.Fatalf("no GetBody: resp = %v, err = %v, calls = %d", resp, err, base.calls.Load())
	}

	req, _ = http.NewRequest(http.MethodPost, "http://vault.test", strings.NewReader("x"))
	req.GetBody = func() (io.ReadCloser, error) { return nil, errors.New("gone") }
	if resp, _ := transport.RoundTrip(req); resp.StatusCode != http.StatusUnauthorized || base.calls.Load() != 2 {
		t.Fatalf("GetBody error: calls = %d", base.calls.Load())
	}

	// Without a body it is resent; the refresh declining returns the 401.
	req, _ = http.NewRequest(http.MethodGet, "http://vault.test", nil)
	if resp, _ := transport.RoundTrip(req); resp.StatusCode != http.StatusUnauthorized || base.calls.Load() != 4 {
		t.Fatalf("no body: calls = %d", base.calls.Load())
	}
	declining := &unauthorizedTransport{base: base, refresh: func(context.Context, string) string { return "" }}
	if resp, _ := declining.RoundTrip(req); resp.StatusCode != http.StatusUnauthorized || base.calls.Load() != 5 {
		t.Fatalf("declined: calls = %d", base.calls.Load())
	}
}

func TestExpiringWithin(t *testing.T) {
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "x"}).SignedString([]byte("secret"))
	switch {
	case expiringWithin("not a jwt", time.Minute), expiringWithin(noExp, time.Minute):
		t.Fatal("a token without a readable expiry is not expiring")
	case !expiringWithin(tokenExpiringIn(t, -time.Minute), time.Minute):
		t.Fatal("an expired token is expiring")
	case expiringWithin(tokenExpiringIn(t, time.Hour), time.Minute):
		t.Fatal("a token with an hour left is not expiring")
	}
}

func TestSignedURLClientSendsNoVaultTokenRetry(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v"}, ClientSettings{})
	upload := SignedURLClient(b.newCallClient(ResolvedHttpConfig{ReadTimeout: time.Second}))
	if upload.Transport != transportFor(ResolvedHttpConfig{ReadTimeout: time.Second}) {
		t.Fatalf("transport = %T", upload.Transport)
	}
}
