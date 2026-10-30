package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	sa "github.com/skyflowapi/skyflow-go/common/serviceaccount"
)

func TestAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Auth Suite")
}

const (
	testSDKPrefix = "Skyflow Go SDK test"
	testLogPrefix = "[ Skyflow Go SDK test ] "
	tokenPath     = "/v1/auth/sa/oauth/token"
)

// failingTransport makes every request fail before reaching a server, which
// yields a non-APIError from the generated client.
type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("forced transport error")
}

func newRSAKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).NotTo(HaveOccurred())
	return key
}

func pkcs1PEM(key *rsa.PrivateKey) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: privateKeyPEMType, Bytes: x509.MarshalPKCS1PrivateKey(key)}))
}

func pkcs8PEM(key interface{}) string {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	Expect(err).NotTo(HaveOccurred())
	return string(pem.EncodeToMemory(&pem.Block{Type: privateKeyPEMType, Bytes: der}))
}

func credsMap(pemKey, tokenUri string) map[string]interface{} {
	return map[string]interface{}{
		"clientId":   "client-id",
		"keyId":      "key-id",
		"tokenUri":   tokenUri,
		"privateKey": pemKey,
	}
}

func credsJSON(m map[string]interface{}) string {
	b, err := json.Marshal(m)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}

func writeFile(content string) string {
	path := filepath.Join(GinkgoT().TempDir(), "creds.json")
	Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
	return path
}

// tokenServer serves the token endpoint over TLS. It records the last request
// body and replies with status and body.
func tokenServer(status int, body string, lastBody *map[string]interface{}) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		Expect(r.URL.Path).To(Equal(tokenPath))
		raw, _ := io.ReadAll(r.Body)
		if lastBody != nil {
			Expect(json.Unmarshal(raw, lastBody)).To(Succeed())
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-request-id", "req-123")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func signedJWT(claims jwt.MapClaims) string {
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	Expect(err).NotTo(HaveOccurred())
	return s
}

var _ = Describe("Auth Service", func() {
	var (
		svc *Service
		key *rsa.PrivateKey
	)

	BeforeEach(func() {
		svc = New(Config{SDKPrefix: testSDKPrefix, LogPrefix: testLogPrefix})
		key = newRSAKey()
	})

	Context("GetScopeUsingRoles", func() {
		It("should format roles as a scope string", func() {
			Expect(GetScopeUsingRoles([]string{"r1", "r2"})).To(Equal(" role:r1 role:r2"))
		})
		It("should return empty string for no roles", func() {
			Expect(GetScopeUsingRoles(nil)).To(Equal(""))
		})
	})

	Context("GetBaseURL", func() {
		It("should reduce an https URL to scheme and host", func() {
			u, err := svc.GetBaseURL("https://manage.skyflow.dev/v1/auth/sa/oauth/token")
			Expect(err).To(BeNil())
			Expect(u).To(Equal("https://manage.skyflow.dev"))
		})
		DescribeTable("should reject invalid token URIs",
			func(in string) {
				u, err := svc.GetBaseURL(in)
				Expect(u).To(Equal(""))
				Expect(err).NotTo(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring(testSDKPrefix + errInvalidTokenUri))
			},
			Entry("http scheme", "http://manage.skyflow.dev"),
			Entry("missing host", "https://"),
			Entry("unparseable", "://bad url"),
		)
	})

	Context("ValidateAndResolveCtx", func() {
		DescribeTable("should resolve valid ctx values",
			func(in interface{}, expected interface{}) {
				out, err := svc.ValidateAndResolveCtx(in)
				Expect(err).To(BeNil())
				if expected == nil {
					Expect(out).To(BeNil())
				} else {
					Expect(out).To(Equal(expected))
				}
			},
			Entry("nil", nil, nil),
			Entry("empty string", "", nil),
			Entry("string", "user_1", "user_1"),
			Entry("bool", true, true),
			Entry("float64", 1.5, 1.5),
			Entry("int", 7, float64(7)),
			Entry("empty map", map[string]interface{}{}, nil),
			Entry("valid map", map[string]interface{}{"role_1": "a"}, map[string]interface{}{"role_1": "a"}),
		)
		It("should reject a map with an invalid key", func() {
			out, err := svc.ValidateAndResolveCtx(map[string]interface{}{"bad-key": 1})
			Expect(out).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring("Invalid key 'bad-key' in ctx map"))
		})
		It("should reject an unsupported type", func() {
			out, err := svc.ValidateAndResolveCtx([]string{"x"})
			Expect(out).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(testSDKPrefix + errInvalidCtxType))
		})
	})

	Context("ParsePrivateKey", func() {
		It("should parse a PKCS1 key", func() {
			k, err := svc.ParsePrivateKey(pkcs1PEM(key))
			Expect(err).To(BeNil())
			Expect(k.Equal(key)).To(BeTrue())
		})
		It("should parse a PKCS8 RSA key", func() {
			k, err := svc.ParsePrivateKey(pkcs8PEM(key))
			Expect(err).To(BeNil())
			Expect(k.Equal(key)).To(BeTrue())
		})
		It("should reject non-PEM input", func() {
			_, err := svc.ParsePrivateKey("not a pem")
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errJwtInvalidFormat))
		})
		It("should reject a PEM block of the wrong type", func() {
			p := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
			_, err := svc.ParsePrivateKey(p)
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errJwtInvalidFormat))
		})
		It("should reject undecodable key bytes", func() {
			p := string(pem.EncodeToMemory(&pem.Block{Type: privateKeyPEMType, Bytes: []byte("garbage")}))
			_, err := svc.ParsePrivateKey(p)
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errInvalidAlgorithm))
		})
		It("should reject a non-RSA PKCS8 key", func() {
			ec, genErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			Expect(genErr).NotTo(HaveOccurred())
			_, err := svc.ParsePrivateKey(pkcs8PEM(ec))
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errInvalidKeySpec))
		})
	})

	Context("GetPrivateKey", func() {
		It("should return the parsed key", func() {
			k, err := svc.GetPrivateKey(map[string]interface{}{"privateKey": pkcs1PEM(key)})
			Expect(err).To(BeNil())
			Expect(k).NotTo(BeNil())
		})
		It("should error when privateKey is missing", func() {
			_, err := svc.GetPrivateKey(map[string]interface{}{})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errMissingPrivateKey))
		})
	})

	Context("GetCredentialParams", func() {
		It("should read the current field names", func() {
			c, t, k, err := svc.GetCredentialParams(map[string]interface{}{"clientId": "c", "tokenUri": "t", "keyId": "k"})
			Expect(err).To(BeNil())
			Expect([]string{c, t, k}).To(Equal([]string{"c", "t", "k"}))
		})
		It("should fall back to the legacy field names", func() {
			c, t, k, err := svc.GetCredentialParams(map[string]interface{}{"clientID": "c", "tokenURI": "t", "keyID": "k"})
			Expect(err).To(BeNil())
			Expect([]string{c, t, k}).To(Equal([]string{"c", "t", "k"}))
		})
		It("should prefer the current field names over legacy ones", func() {
			c, _, _, err := svc.GetCredentialParams(map[string]interface{}{"clientId": "new", "clientID": "old", "tokenUri": "t", "keyId": "k"})
			Expect(err).To(BeNil())
			Expect(c).To(Equal("new"))
		})
		DescribeTable("should report the missing field",
			func(m map[string]interface{}, msg string) {
				_, _, _, err := svc.GetCredentialParams(m)
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(testSDKPrefix + msg))
			},
			Entry("client id", map[string]interface{}{"tokenUri": "t", "keyId": "k"}, errMissingClientId),
			Entry("token uri", map[string]interface{}{"clientId": "c", "keyId": "k"}, errMissingTokenUri),
			Entry("key id", map[string]interface{}{"clientId": "c", "tokenUri": "t"}, errMissingKeyId),
		)
	})

	Context("ParseCredentialsFile", func() {
		It("should parse a valid JSON file", func() {
			m, err := svc.ParseCredentialsFile(writeFile(`{"clientId":"c"}`))
			Expect(err).To(BeNil())
			Expect(m).To(HaveKeyWithValue("clientId", "c"))
		})
		It("should error when the file does not exist", func() {
			missing := filepath.Join(GinkgoT().TempDir(), "missing.json")
			_, err := svc.ParseCredentialsFile(missing)
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring("Credential file not found at " + missing))
		})
		It("should error when the path cannot be read", func() {
			dir := GinkgoT().TempDir()
			_, err := svc.ParseCredentialsFile(dir)
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring("Unable to read credentials - file " + dir))
		})
		It("should error when the file is not valid JSON", func() {
			_, err := svc.ParseCredentialsFile(writeFile("{not json"))
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errCredentialsStringInvalidJSON))
		})
	})

	Context("IsExpired", func() {
		It("should treat an empty token as expired", func() {
			Expect(svc.IsExpired("")).To(BeTrue())
		})
		It("should treat a malformed token as expired", func() {
			Expect(svc.IsExpired("not.a.jwt")).To(BeTrue())
		})
		It("should report an expired token", func() {
			Expect(svc.IsExpired(signedJWT(jwt.MapClaims{"exp": time.Now().Add(-time.Hour).Unix()}))).To(BeTrue())
		})
		It("should report a valid token as not expired", func() {
			Expect(svc.IsExpired(signedJWT(jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}))).To(BeFalse())
		})
		It("should treat a token without exp as expired", func() {
			Expect(svc.IsExpired(signedJWT(jwt.MapClaims{"sub": "x"}))).To(BeTrue())
		})
	})

	Context("GenerateSignedDataTokensHelper", func() {
		It("should sign each token with default TTL and no ctx", func() {
			res, err := svc.GenerateSignedDataTokensHelper("cid", "kid", key, sa.SignedDataTokensOptions{DataTokens: []string{"t1", "t2"}}, "https://aud")
			Expect(err).To(BeNil())
			Expect(res).To(HaveLen(2))
			Expect(res[0].Token).To(Equal("t1"))
			Expect(res[0].SignedToken).To(HavePrefix(signedTokenPrefix))

			parsed, parseErr := jwt.Parse(strings.TrimPrefix(res[1].SignedToken, signedTokenPrefix), func(*jwt.Token) (interface{}, error) {
				return &key.PublicKey, nil
			})
			Expect(parseErr).NotTo(HaveOccurred())
			claims := parsed.Claims.(jwt.MapClaims)
			Expect(claims).To(HaveKeyWithValue(jwtClaimTok, "t2"))
			Expect(claims).To(HaveKeyWithValue(jwtClaimIss, sdkIssuer))
			Expect(claims).To(HaveKeyWithValue(jwtClaimSub, "cid"))
			Expect(claims).To(HaveKeyWithValue(jwtClaimKey, "kid"))
			Expect(claims).To(HaveKeyWithValue(jwtClaimAud, "https://aud"))
			Expect(claims).NotTo(HaveKey(jwtClaimCtx))
			Expect(claims[jwtClaimExp].(float64) - claims[jwtClaimIat].(float64)).To(BeNumerically("==", 60))
		})
		It("should honour TimeToLive and ctx", func() {
			res, err := svc.GenerateSignedDataTokensHelper("cid", "kid", key, sa.SignedDataTokensOptions{DataTokens: []string{"t1"}, TimeToLive: 300, Ctx: "c1"}, "https://aud")
			Expect(err).To(BeNil())
			parsed, _, parseErr := new(jwt.Parser).ParseUnverified(strings.TrimPrefix(res[0].SignedToken, signedTokenPrefix), jwt.MapClaims{})
			Expect(parseErr).NotTo(HaveOccurred())
			claims := parsed.Claims.(jwt.MapClaims)
			Expect(claims).To(HaveKeyWithValue(jwtClaimCtx, "c1"))
			Expect(claims[jwtClaimExp].(float64) - claims[jwtClaimIat].(float64)).To(BeNumerically("==", 300))
		})
		It("should reject an invalid ctx", func() {
			_, err := svc.GenerateSignedDataTokensHelper("cid", "kid", key, sa.SignedDataTokensOptions{DataTokens: []string{"t1"}, Ctx: 1.5i}, "https://aud")
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errInvalidCtxType))
		})
		It("should return an error when signing fails", func() {
			bad := &rsa.PrivateKey{PublicKey: rsa.PublicKey{N: big.NewInt(1), E: 65537}, D: big.NewInt(1)}
			_, err := svc.GenerateSignedDataTokensHelper("cid", "kid", bad, sa.SignedDataTokensOptions{DataTokens: []string{"t1"}}, "https://aud")
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(testSDKPrefix + errErrorOccurred))
		})
	})

	Context("GetSignedBearerUserToken", func() {
		It("should sign an assertion with ctx", func() {
			tok, err := svc.GetSignedBearerUserToken("cid", "kid", "https://aud", key, sa.BearerTokenOptions{Ctx: map[string]interface{}{"a": 1}})
			Expect(err).To(BeNil())
			parsed, parseErr := jwt.Parse(tok, func(*jwt.Token) (interface{}, error) { return &key.PublicKey, nil })
			Expect(parseErr).NotTo(HaveOccurred())
			claims := parsed.Claims.(jwt.MapClaims)
			Expect(claims).To(HaveKeyWithValue(jwtClaimIss, "cid"))
			Expect(claims).To(HaveKey(jwtClaimCtx))
		})
		It("should leave ctx out of the assertion when none is given", func() {
			tok, err := svc.GetSignedBearerUserToken("cid", "kid", "https://aud", key, sa.BearerTokenOptions{})
			Expect(err).To(BeNil())
			parsed, parseErr := jwt.Parse(tok, func(*jwt.Token) (interface{}, error) { return &key.PublicKey, nil })
			Expect(parseErr).NotTo(HaveOccurred())
			Expect(parsed.Claims.(jwt.MapClaims)).NotTo(HaveKey(jwtClaimCtx))
		})
		It("should reject an invalid ctx", func() {
			_, err := svc.GetSignedBearerUserToken("cid", "kid", "https://aud", key, sa.BearerTokenOptions{Ctx: struct{}{}})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errInvalidCtxType))
		})
		It("should return an error when signing fails", func() {
			bad := &rsa.PrivateKey{PublicKey: rsa.PublicKey{N: big.NewInt(1), E: 65537}, D: big.NewInt(1)}
			_, err := svc.GetSignedBearerUserToken("cid", "kid", "https://aud", bad, sa.BearerTokenOptions{})
			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 500"))
			Expect(err.GetMessage()).To(ContainSubstring(testSDKPrefix + " Error occurred."))
		})
	})

	Context("GetSignedDataTokens", func() {
		It("should sign tokens from credentials", func() {
			res, err := svc.GetSignedDataTokens(credsMap(pkcs1PEM(key), "https://aud"), sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
		It("should error when the private key is missing", func() {
			m := credsMap(pkcs1PEM(key), "https://aud")
			delete(m, "privateKey")
			_, err := svc.GetSignedDataTokens(m, sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errMissingPrivateKey))
		})
		It("should error when a credential param is missing", func() {
			m := credsMap(pkcs1PEM(key), "https://aud")
			delete(m, "keyId")
			_, err := svc.GetSignedDataTokens(m, sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errMissingKeyId))
		})
	})

	Context("GenerateSignedDataTokens", func() {
		It("should sign tokens from a credentials file", func() {
			path := writeFile(credsJSON(credsMap(pkcs1PEM(key), "https://aud")))
			res, err := svc.GenerateSignedDataTokens(path, sa.SignedDataTokensOptions{DataTokens: []string{"t1", "t2"}})
			Expect(err).To(BeNil())
			Expect(res).To(HaveLen(2))
		})
		It("should reject an empty path", func() {
			_, err := svc.GenerateSignedDataTokens("", sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errEmptyCredentialFilePath))
		})
		It("should reject empty data tokens", func() {
			_, err := svc.GenerateSignedDataTokens("creds.json", sa.SignedDataTokensOptions{})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errEmptyTokensDetokenize))
		})
		It("should propagate file errors", func() {
			_, err := svc.GenerateSignedDataTokens(filepath.Join(GinkgoT().TempDir(), "nope.json"), sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring("Credential file not found"))
		})
	})

	Context("GenerateSignedDataTokensFromCreds", func() {
		It("should sign tokens from a credentials string", func() {
			res, err := svc.GenerateSignedDataTokensFromCreds(credsJSON(credsMap(pkcs1PEM(key), "https://aud")), sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
		It("should reject an empty string", func() {
			_, err := svc.GenerateSignedDataTokensFromCreds("", sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errEmptyCredentialsString))
		})
		It("should reject invalid JSON", func() {
			_, err := svc.GenerateSignedDataTokensFromCreds("{bad", sa.SignedDataTokensOptions{DataTokens: []string{"t"}})
			Expect(err).NotTo(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errInvalidCredentials))
		})
	})

	Context("Bearer token generation", func() {
		var (
			srv      *httptest.Server
			lastBody map[string]interface{}
		)

		newService := func(status int, body string) {
			lastBody = nil
			srv = tokenServer(status, body, &lastBody)
			DeferCleanup(srv.Close)
			svc = New(Config{SDKPrefix: testSDKPrefix, LogPrefix: testLogPrefix, HTTPClient: srv.Client()})
		}

		Context("GenerateBearerTokenHelper", func() {
			It("should return the raw token and send roleIds as scope", func() {
				newService(http.StatusOK, `{"accessToken":"at","tokenType":"Bearer"}`)
				res, err := svc.GenerateBearerTokenHelper(credsMap(pkcs8PEM(key), srv.URL+tokenPath), sa.BearerTokenOptions{RoleIds: []string{"r1"}, RoleIDs: []string{"ignored"}})
				Expect(err).To(BeNil())
				Expect(*res.AccessToken).To(Equal("at"))
				Expect(*res.TokenType).To(Equal("Bearer"))
				Expect(lastBody).To(HaveKeyWithValue("grant_type", grantType))
				Expect(lastBody).To(HaveKeyWithValue("scope", " role:r1"))
				Expect(lastBody).To(HaveKey("assertion"))
			})
			It("should fall back to deprecated RoleIDs", func() {
				newService(http.StatusOK, `{"accessToken":"at","tokenType":"Bearer"}`)
				_, err := svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), srv.URL), sa.BearerTokenOptions{RoleIDs: []string{"old"}})
				Expect(err).To(BeNil())
				Expect(lastBody).To(HaveKeyWithValue("scope", " role:old"))
			})
			It("should omit scope when no roles are given", func() {
				newService(http.StatusOK, `{"accessToken":"at","tokenType":"Bearer"}`)
				_, err := svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), srv.URL), sa.BearerTokenOptions{})
				Expect(err).To(BeNil())
				Expect(lastBody).NotTo(HaveKey("scope"))
			})
			It("should use a custom base URL resolver and restore the default with nil", func() {
				newService(http.StatusOK, `{"accessToken":"at","tokenType":"Bearer"}`)
				var seen string
				svc.SetBaseURLResolver(func(tokenUri string) (string, *skyflowError.SkyflowError) {
					seen = tokenUri
					return srv.URL, nil
				})
				_, err := svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), "custom-uri"), sa.BearerTokenOptions{})
				Expect(err).To(BeNil())
				Expect(seen).To(Equal("custom-uri"))

				svc.SetBaseURLResolver(nil)
				_, err = svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), "custom-uri"), sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errInvalidTokenUri))
			})
			It("should error when the private key is missing", func() {
				m := credsMap(pkcs1PEM(key), "https://aud")
				delete(m, "privateKey")
				_, err := svc.GenerateBearerTokenHelper(m, sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errMissingPrivateKey))
			})
			It("should error when the private key is invalid", func() {
				_, err := svc.GenerateBearerTokenHelper(credsMap("bad", "https://aud"), sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errJwtInvalidFormat))
			})
			It("should error when a credential param is missing", func() {
				m := credsMap(pkcs1PEM(key), "https://aud")
				delete(m, "clientId")
				_, err := svc.GenerateBearerTokenHelper(m, sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errMissingClientId))
			})
			It("should error when ctx is invalid", func() {
				_, err := svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), "https://aud"), sa.BearerTokenOptions{Ctx: []int{1}})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errInvalidCtxType))
			})
			It("should map an API error and keep the request ID", func() {
				newService(http.StatusUnauthorized, `{"error":{"http_code":401,"message":"bad assertion","grpc_code":16,"http_status":"Unauthorized","details":[]}}`)
				_, err := svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), srv.URL), sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring("bad assertion"))
				Expect(err.GetRequestId()).To(Equal("req-123"))
			})
			It("should map a transport error without a response header", func() {
				svc = New(Config{SDKPrefix: testSDKPrefix, HTTPClient: &http.Client{Transport: failingTransport{}}})
				_, err := svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), "https://manage.example.com"), sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring("forced transport error"))
				Expect(err.GetRequestId()).To(Equal(""))
			})
			It("should use the default HTTP client when none is configured", func() {
				svc = New(Config{})
				svc.SetBaseURLResolver(func(string) (string, *skyflowError.SkyflowError) {
					return "http://127.0.0.1:1", nil
				})
				_, err := svc.GenerateBearerTokenHelper(credsMap(pkcs1PEM(key), "https://x"), sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
			})
		})

		Context("GenerateBearerToken", func() {
			It("should return a token from a credentials file", func() {
				newService(http.StatusOK, `{"accessToken":"at","tokenType":"Bearer"}`)
				path := writeFile(credsJSON(credsMap(pkcs1PEM(key), srv.URL)))
				res, err := svc.GenerateBearerToken(path, sa.BearerTokenOptions{})
				Expect(err).To(BeNil())
				Expect(res.AccessToken).To(Equal("at"))
				Expect(res.TokenType).To(Equal("Bearer"))
			})
			It("should reject an empty path", func() {
				_, err := svc.GenerateBearerToken("", sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errEmptyCredentialFilePath))
			})
			It("should propagate file errors", func() {
				_, err := svc.GenerateBearerToken(writeFile("{bad"), sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errCredentialsStringInvalidJSON))
			})
		})

		Context("GenerateBearerTokenFromCreds", func() {
			It("should return a token from a credentials string", func() {
				newService(http.StatusOK, `{"accessToken":"at","tokenType":"Bearer"}`)
				res, err := svc.GenerateBearerTokenFromCreds(credsJSON(credsMap(pkcs1PEM(key), srv.URL)), sa.BearerTokenOptions{})
				Expect(err).To(BeNil())
				Expect(res.AccessToken).To(Equal("at"))
			})
			It("should reject invalid JSON", func() {
				_, err := svc.GenerateBearerTokenFromCreds("{bad", sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errInvalidCredentials))
			})
			It("should propagate helper errors", func() {
				_, err := svc.GenerateBearerTokenFromCreds(`{}`, sa.BearerTokenOptions{})
				Expect(err).NotTo(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errMissingPrivateKey))
			})
			DescribeTable("should reject a response missing a token field",
				func(body string) {
					newService(http.StatusOK, body)
					_, err := svc.GenerateBearerTokenFromCreds(credsJSON(credsMap(pkcs1PEM(key), srv.URL)), sa.BearerTokenOptions{})
					Expect(err).NotTo(BeNil())
					Expect(err.GetCode()).To(Equal("Code: 500"))
					Expect(err.GetMessage()).To(ContainSubstring(testLogPrefix + logBearerTokenRejected))
				},
				Entry("no access token", `{"tokenType":"Bearer"}`),
				Entry("no token type", `{"accessToken":"at"}`),
			)
		})
	})
})
