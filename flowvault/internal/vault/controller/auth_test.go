package controller

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// unsignedJWT builds a JWT whose exp claim is at exp. The SDK only reads the claims to
// check expiry, so the signature is a placeholder.
func unsignedJWT(exp time.Time) string {
	encode := func(v interface{}) string {
		raw, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	return encode(map[string]string{"alg": "RS256", "typ": "JWT"}) + "." + encode(map[string]int64{"exp": exp.Unix()}) + ".c2ln"
}

// fakeAuthServer issues bearer tokens. The shared authenticator is pointed at it for the
// spec, so service-account credentials never reach a real token endpoint.
type fakeAuthServer struct {
	server *httptest.Server
	calls  int32
	scope  atomic.Value
	status int32
}

func newFakeAuthServer() *fakeAuthServer {
	f := &fakeAuthServer{}
	f.scope.Store("")
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&f.calls, 1)
		raw, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		_ = json.Unmarshal(raw, &body)
		if scope, ok := body["scope"].(string); ok {
			f.scope.Store(scope)
		}
		w.Header().Set("Content-Type", "application/json")
		if status := atomic.LoadInt32(&f.status); status != 0 {
			w.WriteHeader(int(status))
			_, _ = io.WriteString(w, `{"error":{"http_code":401,"message":"bad assertion"}}`)
			return
		}
		_, _ = fmt.Fprintf(w, `{"accessToken":%q,"tokenType":"Bearer"}`, unsignedJWT(time.Now().Add(time.Hour)))
	}))
	fvhelpers.Auth().SetBaseURLResolver(func(string) (string, *fverrors.SkyflowError) { return f.server.URL, nil })
	DeferCleanup(func() {
		fvhelpers.Auth().SetBaseURLResolver(nil)
		f.server.Close()
	})
	return f
}

// fakeCredentials returns service-account credentials signed with a key made for the test,
// so no key material is checked in.
func fakeCredentials() string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).To(BeNil())
	der, err := x509.MarshalPKCS8PrivateKey(key)
	Expect(err).To(BeNil())
	raw, _ := json.Marshal(map[string]string{
		"clientID":   "client-id",
		"keyID":      "key-id",
		"tokenURI":   "https://auth.example.invalid/v1/auth/sa/oauth/token",
		"privateKey": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
	})
	return string(raw)
}

var _ = Describe("GenerateToken", func() {
	It("passes an API key or bearer token through", func() {
		token, err := GenerateToken(fvcommon.Credentials{Token: "bearer"})
		Expect(err).To(BeNil())
		Expect(*token).To(Equal("bearer"))
		token, err = GenerateToken(fvcommon.Credentials{ApiKey: testApiKey})
		Expect(err).To(BeNil())
		Expect(*token).To(Equal(testApiKey))
	})

	It("fails without credentials or with bad service-account credentials", func() {
		_, err := GenerateToken(fvcommon.Credentials{})
		Expect(err.GetMessage()).To(ContainSubstring("Invalid credentials"))
		_, err = GenerateToken(fvcommon.Credentials{Path: filepath.Join(GinkgoT().TempDir(), "missing.json")})
		Expect(err).ToNot(BeNil())
		_, err = GenerateToken(fvcommon.Credentials{CredentialsString: "not json"})
		Expect(err).ToNot(BeNil())
	})

	It("exchanges a credentials file or string for a bearer token, with roles and context", func() {
		auth := newFakeAuthServer()
		creds := fakeCredentials()
		path := filepath.Join(GinkgoT().TempDir(), "credentials.json")
		Expect(os.WriteFile(path, []byte(creds), 0o600)).To(Succeed())

		token, err := GenerateToken(fvcommon.Credentials{Path: path, Roles: []string{"r1", "r2"}, Context: "ctx"})
		Expect(err).To(BeNil())
		Expect(fvhelpers.Auth().IsExpired(*token)).To(BeFalse())
		Expect(auth.scope.Load()).To(Equal(" role:r1 role:r2"))

		token, err = GenerateToken(fvcommon.Credentials{CredentialsString: creds})
		Expect(err).To(BeNil())
		Expect(*token).ToNot(BeEmpty())
		Expect(atomic.LoadInt32(&auth.calls)).To(Equal(int32(2)))
	})
})

var _ = Describe("Bearer token handling", func() {
	It("sends a bearer token that hasn't expired", func() {
		bearer := unsignedJWT(time.Now().Add(time.Hour))
		config := fvcommon.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: fvcommon.Credentials{Token: bearer}}
		setup, err := NewVaultController(config, ClientSettings{}).SetupRequest("GET", nil)
		Expect(err).To(BeNil())
		Expect(setup.AuthToken).To(Equal(bearer))
	})

	It("generates a token once, reuses it, and generates again after the credentials change", func() {
		auth := newFakeAuthServer()
		config := fvcommon.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: fvcommon.Credentials{CredentialsString: fakeCredentials()}}
		v := NewVaultController(config, ClientSettings{})

		first, err := v.SetupRequest("GET", nil)
		Expect(err).To(BeNil())
		second, err := v.SetupRequest("GET", nil)
		Expect(err).To(BeNil())
		Expect(second.AuthToken).To(Equal(first.AuthToken))
		Expect(atomic.LoadInt32(&auth.calls)).To(Equal(int32(1)))

		v.ResetAuth()
		_, err = v.SetupRequest("GET", nil)
		Expect(err).To(BeNil())
		Expect(atomic.LoadInt32(&auth.calls)).To(Equal(int32(2)))
	})

	It("fails the call when the token endpoint rejects the credentials", func() {
		auth := newFakeAuthServer()
		atomic.StoreInt32(&auth.status, http.StatusUnauthorized)
		config := fvcommon.VaultConfig{VaultId: "v", ClusterId: "c", Credentials: fvcommon.Credentials{CredentialsString: fakeCredentials()}}
		_, err := NewVaultController(config, ClientSettings{}).SetupRequest("GET", nil)
		Expect(err).ToNot(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring("bad assertion"))
	})

	It("falls back to SKYFLOW_CREDENTIALS", func() {
		auth := newFakeAuthServer()
		previous, had := os.LookupEnv("SKYFLOW_CREDENTIALS")
		Expect(os.Setenv("SKYFLOW_CREDENTIALS", fakeCredentials())).To(Succeed())
		DeferCleanup(func() {
			if had {
				_ = os.Setenv("SKYFLOW_CREDENTIALS", previous)
			} else {
				_ = os.Unsetenv("SKYFLOW_CREDENTIALS")
			}
		})
		setup, err := NewVaultController(fvcommon.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{}).SetupRequest("GET", nil)
		Expect(err).To(BeNil())
		Expect(setup.AuthToken).ToNot(BeEmpty())
		Expect(atomic.LoadInt32(&auth.calls)).To(Equal(int32(1)))
	})
})

var _ = Describe("SetupRequest, edge cases", func() {
	It("uses the deprecated BaseVaultURL field", func() {
		config := fvcommon.VaultConfig{VaultId: "v", BaseVaultURL: "https://legacy.example.com"}
		setup, err := NewVaultController(config, ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}}).SetupRequest("GET", nil)
		Expect(err).To(BeNil())
		Expect(setup.BaseURL).To(Equal("https://legacy.example.com"))
	})

	It("rejects an unsupported interceptor header before authenticating", func() {
		_, err := NewVaultController(fvcommon.VaultConfig{VaultId: "v", ClusterId: "c"}, ClientSettings{}).
			SetupRequest("GET", func(rc *fvcommon.RequestContext) { rc.AddHeader("x-other", "v") })
		Expect(err.GetMessage()).To(ContainSubstring("is not supported"))
	})

	It("reports a retryable response whose body can't be read", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, buf, err := w.(http.Hijacker).Hijack()
			Expect(err).To(BeNil())
			defer conn.Close()
			_, _ = buf.WriteString("HTTP/1.1 503 Service Unavailable\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\nshort")
			_ = buf.Flush()
		}))
		defer server.Close()
		v := NewVaultController(fvcommon.VaultConfig{VaultId: "v", BaseVaultUrl: server.URL}, ClientSettings{Credentials: &fvcommon.Credentials{ApiKey: testApiKey}})
		setup, err := v.SetupRequest("QUERY", nil)
		Expect(err).To(BeNil())

		req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
		resp, doErr := finalResponseClient{setup.HTTPClient}.Do(req)
		Expect(resp).To(BeNil())
		Expect(doErr).To(MatchError(io.ErrUnexpectedEOF))
	})
})
