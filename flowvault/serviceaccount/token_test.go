package serviceaccount

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

func TestServiceAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault ServiceAccount Suite")
}

const sdkPrefix = "Skyflow FlowVault Go SDK"

// fakeCredentials returns service-account credentials with a key made for the test, so no
// key material is checked in. The token URI is never called: signed data tokens are
// signed locally and the bearer-token specs only cover input errors.
func fakeCredentials() string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	Expect(err).To(BeNil())
	der, err := x509.MarshalPKCS8PrivateKey(key)
	Expect(err).To(BeNil())
	raw, _ := json.Marshal(map[string]string{
		"clientId":   "client-id",
		"keyId":      "key-id",
		"tokenUri":   "https://auth.example.invalid/v1/auth/sa/oauth/token",
		"privateKey": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
	})
	return string(raw)
}

func jwtWithExp(exp time.Time) string {
	encode := func(v interface{}) string {
		raw, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(raw)
	}
	return encode(map[string]string{"alg": "RS256", "typ": "JWT"}) + "." + encode(map[string]int64{"exp": exp.Unix()}) + ".c2ln"
}

var _ = Describe("Service account", func() {
	It("reports bearer-token input errors with the flowvault prefix", func() {
		_, err := GenerateBearerToken("", fvcommon.BearerTokenOptions{})
		Expect(err.GetMessage()).To(ContainSubstring(sdkPrefix))
		_, err = GenerateBearerToken(filepath.Join(GinkgoT().TempDir(), "missing.json"), fvcommon.BearerTokenOptions{})
		Expect(err.GetMessage()).To(ContainSubstring(sdkPrefix))
		_, err = GenerateBearerTokenFromCreds("not json", fvcommon.BearerTokenOptions{})
		Expect(err.GetMessage()).To(ContainSubstring(sdkPrefix))
	})

	It("signs data tokens from a credentials file or string", func() {
		creds := fakeCredentials()
		path := filepath.Join(GinkgoT().TempDir(), "credentials.json")
		Expect(os.WriteFile(path, []byte(creds), 0o600)).To(Succeed())
		options := fvcommon.SignedDataTokensOptions{DataTokens: []string{"tok1", "tok2"}, TimeToLive: 60}

		fromFile, err := GenerateSignedDataTokens(path, options)
		Expect(err).To(BeNil())
		Expect(fromFile).To(HaveLen(2))
		Expect(fromFile[0].Token).To(Equal("tok1"))
		Expect(fromFile[0].SignedToken).ToNot(BeEmpty())

		fromString, err := GenerateSignedDataTokensFromCreds(creds, options)
		Expect(err).To(BeNil())
		Expect(fromString).To(HaveLen(2))
		Expect(fromString[1].Token).To(Equal("tok2"))
		jwt := fromString[1].SignedToken[strings.Index(fromString[1].SignedToken, "ey"):]
		Expect(IsExpired(jwt)).To(BeFalse())
	})

	It("rejects signed-data-token input errors", func() {
		_, err := GenerateSignedDataTokens("", fvcommon.SignedDataTokensOptions{DataTokens: []string{"t"}})
		Expect(err.GetMessage()).To(ContainSubstring(sdkPrefix))
		_, err = GenerateSignedDataTokensFromCreds("", fvcommon.SignedDataTokensOptions{DataTokens: []string{"t"}})
		Expect(err.GetMessage()).To(ContainSubstring(sdkPrefix))
	})

	It("checks token expiry", func() {
		Expect(IsExpired("")).To(BeTrue())
		Expect(IsExpired("not-a-jwt")).To(BeTrue())
		Expect(IsExpired(jwtWithExp(time.Now().Add(-time.Minute)))).To(BeTrue())
		Expect(IsExpired(jwtWithExp(time.Now().Add(time.Hour)))).To(BeFalse())
	})
})
