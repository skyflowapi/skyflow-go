package vaultclient

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/skyflowapi/skyflow-go/common/auth"
	"github.com/skyflowapi/skyflow-go/common/config"
	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
)

// fakeAuth is a service-account token endpoint. release, when set, holds every response
// until it is closed. hit receives a value for each request, once it has arrived.
type fakeAuth struct {
	server  *httptest.Server
	calls   atomic.Int32
	hit     chan struct{}
	release chan struct{}
	creds   string
}

func newFakeAuth(t *testing.T, release chan struct{}) *fakeAuth {
	t.Helper()
	f := &fakeAuth{release: release, hit: make(chan struct{}, 16)}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	f.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		f.calls.Add(1)
		f.hit <- struct{}{}
		if f.release != nil {
			<-f.release
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"accessToken": token, "tokenType": "Bearer"})
	}))
	t.Cleanup(f.server.Close)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	raw, err := json.Marshal(map[string]string{"clientId": "client", "keyId": "key", "tokenUri": f.server.URL, "privateKey": string(pemKey)})
	if err != nil {
		t.Fatal(err)
	}
	f.creds = string(raw)
	return f
}

func (f *fakeAuth) profile() Profile {
	p := testProfile()
	p.Auth = auth.New(auth.Config{SDKPrefix: p.SDKPrefix, LogPrefix: p.LogPrefix, HTTPClient: f.server.Client()})
	return p
}

func TestProfileReturnsTheProfile(t *testing.T) {
	if got := New(testProfile(), config.VaultConfig{}, ClientSettings{}).Profile(); got.SDKPrefix != "Test SDK" {
		t.Fatalf("profile = %+v", got)
	}
}

func TestGenerateTokenFromEachCredentialSource(t *testing.T) {
	f := newFakeAuth(t, nil)
	p := f.profile()
	path := filepath.Join(t.TempDir(), "creds.json")
	if err := os.WriteFile(path, []byte(f.creds), 0o600); err != nil {
		t.Fatal(err)
	}

	for name, creds := range map[string]config.Credentials{
		"token":              {Token: "tok"},
		"api key":            {ApiKey: testApiKey},
		"path":               {Path: path},
		"credentials string": {CredentialsString: f.creds},
	} {
		if token, err := p.GenerateToken(creds); err != nil || token == "" {
			t.Fatalf("%s: token = %q, err = %v", name, token, err)
		}
	}
	for name, creds := range map[string]config.Credentials{
		"missing file":     {Path: filepath.Join(t.TempDir(), "none.json")},
		"bad json":         {CredentialsString: "{"},
		"no source at all": {},
	} {
		if _, err := p.GenerateToken(creds); err == nil {
			t.Fatalf("%s: expected an error", name)
		}
	}
}

func TestBearerTokenIsGeneratedOnceAndReused(t *testing.T) {
	f := newFakeAuth(t, nil)
	b := New(f.profile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{CredentialsString: f.creds}}, ClientSettings{})
	first, err := b.Prepare("GET", nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := b.Prepare("GET", nil)
	if err != nil || second.AuthToken != first.AuthToken || f.calls.Load() != 1 {
		t.Fatalf("second = %+v, err = %v, calls = %d", second, err, f.calls.Load())
	}
}

func TestBearerTokenGenerationErrorIsReturned(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{CredentialsString: "{"}}, ClientSettings{})
	if _, err := b.Prepare("GET", nil); err == nil {
		t.Fatal("expected an error for bad credentials")
	}
}

func TestContextEndsWhileWaitingForTheTokenLock(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{CredentialsString: "{}"}}, ClientSettings{})
	b.tokenLock <- struct{}{} // another call is generating a token
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := b.PrepareContext(ctx, "GET", nil)
	if err == nil || !errors.Is(err, context.Canceled) || !strings.Contains(err.GetMessage(), "Test SDK Request cancelled") ||
		err.GetHttpCode() != string(skyflowError.REQUEST_CANCELLED) || err.GetHttpStatusCode() != "Client Closed Request" {
		t.Fatalf("err = %v", err)
	}
}

// The context is live when the call starts and ends while it waits for the lock.
func TestContextTimesOutWhileWaitingForTheTokenLock(t *testing.T) {
	b := New(testProfile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{CredentialsString: "{}"}}, ClientSettings{})
	b.tokenLock <- struct{}{} // another call is generating a token
	defer func() { <-b.tokenLock }()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := b.PrepareContext(ctx, "GET", nil)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) || err.GetHttpCode() != string(skyflowError.REQUEST_CANCELLED) {
		t.Fatalf("err = %v", err)
	}
}

func TestEndedContextNeverStartsATokenRequest(t *testing.T) {
	f := newFakeAuth(t, nil)
	b := New(f.profile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{CredentialsString: f.creds}}, ClientSettings{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range 20 { // the lock is free, so a random select would sometimes take it
		if _, err := b.PrepareContext(ctx, "GET", nil); err == nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v", err)
		}
	}
	time.Sleep(50 * time.Millisecond) // a request started in the background would land by now
	if calls := f.calls.Load(); calls != 0 {
		t.Fatalf("calls = %d, want 0", calls)
	}
}

func TestContextEndsWhileATokenIsGeneratedAndTheTokenIsStillCached(t *testing.T) {
	release := make(chan struct{})
	f := newFakeAuth(t, release)
	b := New(f.profile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{CredentialsString: f.creds}}, ClientSettings{})
	call := &Call{VaultID: "v", BaseURL: "https://c.example.com"}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := b.RefreshAuth(ctx, call)
		done <- err
	}()
	select {
	case <-f.hit: // the generation is in flight
	case <-time.After(5 * time.Second):
		t.Fatal("the token endpoint was never called")
	}
	cancel()
	if err := <-done; err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}

	close(release) // the generation started for the cancelled call finishes and is cached
	refreshed, err := b.RefreshAuth(context.Background(), call)
	if err != nil || refreshed.AuthToken == "" || f.calls.Load() != 1 {
		t.Fatalf("refreshed = %+v, err = %v, calls = %d", refreshed, err, f.calls.Load())
	}
}

func TestValidBearerTokenEnvCredentialsAndBaseVaultUrl(t *testing.T) {
	f := newFakeAuth(t, nil)
	token, err := f.profile().GenerateToken(config.Credentials{CredentialsString: f.creds})
	if err != nil {
		t.Fatal(err)
	}
	profile := f.profile()
	profile.Logs = Logs{UsingBearerToken: "own: using token", GenerateBearerToken: "own: generating"}
	b := New(profile, config.VaultConfig{VaultId: "v", BaseVaultUrl: "https://vault.example.com", Credentials: config.Credentials{Token: token}}, ClientSettings{})
	call, err := b.Prepare("GET", nil)
	if err != nil || call.AuthToken != token || call.BaseURL != "https://vault.example.com" {
		t.Fatalf("call = %+v, err = %v", call, err)
	}

	t.Setenv(CredentialsEnv, f.creds)
	b = New(profile, config.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{})
	if call, err := b.Prepare("GET", nil); err != nil || call.AuthToken == "" {
		t.Fatalf("env credentials: call = %+v, err = %v", call, err)
	}
}

func TestClientRetriesApplyWhenTheVaultSetsNone(t *testing.T) {
	retries := 4
	got := ResolveHttpConfig(config.HttpConfig{}, config.HttpConfig{MaxRetries: &retries}, testProfile().Defaults)
	if got.MaxRetries != 4 {
		t.Fatalf("got %+v", got)
	}
}

func TestCustomHeaderNeedsAValue(t *testing.T) {
	err := testProfile().ValidateCustomHeaders(map[config.CustomHeaderKey]string{config.RequestIdHeader: " "}, "Client headers")
	if err == nil || !strings.Contains(err.GetMessage(), "must have a value") {
		t.Fatalf("err = %v", err)
	}
}

func TestCachedTokenIsReplacedBeforeItExpires(t *testing.T) {
	f := newFakeAuth(t, nil)
	b := New(f.profile(), config.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: config.Credentials{CredentialsString: f.creds}}, ClientSettings{})
	if _, err := b.Prepare("GET", nil); err != nil {
		t.Fatal(err)
	}
	// A cached token inside the refresh margin is not reused.
	soon, signErr := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"exp": time.Now().Add(tokenRefreshMargin / 2).Unix()}).SignedString([]byte("secret"))
	if signErr != nil {
		t.Fatal(signErr)
	}
	b.token = soon
	call, err := b.Prepare("GET", nil)
	if err != nil || call.AuthToken == soon || f.calls.Load() != 2 {
		t.Fatalf("call = %+v, err = %v, calls = %d", call, err, f.calls.Load())
	}
}
