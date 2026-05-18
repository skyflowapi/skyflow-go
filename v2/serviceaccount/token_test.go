package serviceaccount_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"

	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"

	"github.com/golang-jwt/jwt/v4"
)

func TestServiceAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceAccount Suite")
}

func mockserver(res string) *httptest.Server {
	// Mock server for simulating the HTTP request/response
	mockServers := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulating a successful response
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/v1/auth/sa/oauth/token" && res == "ok" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
					"accessToken": "mockAccessToken",
					"tokenType": "bearer"
				}`))
			return
		}

		// Simulate an error response for other endpoints
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid_request"}`))
	}))
	return mockServers
}

var _ = Describe("ServiceAccount Test Suite", func() {
	var (
		options          common.BearerTokenOptions
		mockServer       *httptest.Server
		dataTokenOptions common.SignedDataTokensOptions
	)
	BeforeEach(func() {
		options = common.BearerTokenOptions{
			Ctx:     "testContext",
			RoleIds: []string{"roleid1", "roleid2"},
		}
		dataTokenOptions = common.SignedDataTokensOptions{
			DataTokens: []string{"datatoken1", "datatoken2"},
			TimeToLive: 0,
			Ctx:        "testContext",
			LogLevel:   0,
		}
	})
	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
			mockServer = nil
		}
	})

	Context("GenerateBearerToken success/error response", func() {
		It("should return a valid token when credentials are valid", func() {
			mockServer = mockserver("ok")
			// Prepare valid BearerTokenOptions
			options = common.BearerTokenOptions{
				RoleIds: []string{"role1"},
			}
			originalGetBaseURLHelper := helpers.GetBaseURLHelper

			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(urlStr string) (string, *skyflowError.SkyflowError) {
				return mockServer.URL, nil
			}
			if os.Getenv("CRED_FILE_PATH") != "" {
				var file = os.Getenv("CRED_FILE_PATH")
				// Call the function under test
				tokenResp, err := serviceaccount.GenerateBearerToken(file, options)
				// Assert the token response
				Expect(err).To(BeNil())
				Expect(tokenResp.AccessToken).To(Equal("mockAccessToken"))
				Expect(tokenResp.TokenType).To(Equal("bearer"))
			} else {
				fmt.Println("file path is not found")
			}
		})
		It("should return a valid token when credentials are valid", func() {
			mockServer = mockserver("err")
			// Prepare valid BearerTokenOptions
			options = common.BearerTokenOptions{
				RoleIds: []string{"role1"},
			}
			originalGetBaseURLHelper := helpers.GetBaseURLHelper

			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(urlStr string) (string, *skyflowError.SkyflowError) {
				return mockServer.URL, nil
			}
			// Call the function under test
			tokenResp, err := serviceaccount.GenerateBearerToken(os.Getenv("CRED_FILE_PATH"), options)
			// Assert the token response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
		It("should return an error if the credential file is missing a private key", func() {

			// Prepare BearerTokenOptions
			options := common.BearerTokenOptions{
				RoleIds: []string{"role1"},
			}

			// Call the function under test
			tokenResp, err := serviceaccount.GenerateBearerToken("credentials.json", options)

			// Assert the error response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(skyflowError.FILE_NOT_FOUND, "credentials.json")))
		})
		It("should return error for empty credentials file path", func() {
			tokenResp, err := serviceaccount.GenerateBearerToken("", options)
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.EMPTY_CREDENTIAL_FILE_PATH))
		})
		It("should return a valid token from a credential file using a local RSA key", func() {
			credsJSON, srv := makeTestCredsJSONAndServer("ok")
			mockServer = srv

			tmpFile, fileErr := os.CreateTemp("", "creds_*.json")
			Expect(fileErr).To(BeNil())
			_, _ = tmpFile.WriteString(credsJSON)
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())

			originalGetBaseURLHelper := helpers.GetBaseURLHelper
			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(_ string) (string, *skyflowError.SkyflowError) {
				return srv.URL, nil
			}

			opts := common.BearerTokenOptions{RoleIds: []string{"role1"}}
			tokenResp, err := serviceaccount.GenerateBearerToken(tmpFile.Name(), opts)
			Expect(err).To(BeNil())
			Expect(tokenResp).ToNot(BeNil())
			Expect(tokenResp.AccessToken).To(Equal("mockAccessToken"))
		})
	})
	Context("GenerateBearerTokenCreds success/error response", func() {
		It("should return a valid token when credentials are valid", func() {
			mockServer = mockserver("ok")
			// Prepare valid BearerTokenOptions
			options = common.BearerTokenOptions{
				RoleIds: []string{"role1"},
			}
			originalGetBaseURLHelper := helpers.GetBaseURLHelper

			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(urlStr string) (string, *skyflowError.SkyflowError) {
				return mockServer.URL, nil
			}

			// Call the function under test
			tokenResp, err := serviceaccount.GenerateBearerTokenFromCreds(os.Getenv("VALID_CREDS_PVT_KEY"), options)
			// Assert the token response
			Expect(err).To(BeNil())
			Expect(tokenResp.AccessToken).To(Equal("mockAccessToken"))
			Expect(tokenResp.TokenType).To(Equal("bearer"))
		})
		It("should return a valid token when credentials are valid", func() {
			mockServer = mockserver("err")
			// Prepare valid BearerTokenOptions
			options = common.BearerTokenOptions{
				RoleIds: []string{"role1"},
			}
			originalGetBaseURLHelper := helpers.GetBaseURLHelper

			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(urlStr string) (string, *skyflowError.SkyflowError) {
				return mockServer.URL, nil
			}
			// Call the function under test
			tokenResp, err := serviceaccount.GenerateBearerTokenFromCreds(os.Getenv("VALID_CREDS_PVT_KEY"), options)
			// Assert the token response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
		It("should return an error if the credential file is missing a private key", func() {

			// Prepare BearerTokenOptions
			options := common.BearerTokenOptions{
				RoleIds: []string{"role1"},
			}

			// Call the function under test
			tokenResp, err := serviceaccount.GenerateBearerTokenFromCreds("{", options)

			// Assert the error response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.INVALID_CREDENTIALS))
		})
	})
	Context("GenerateSignedTokenCreds success/error response", func() {
		It("should return a valid token when credentials are valid", func() {
			// Call the function under test
			tokenResp, err := serviceaccount.GenerateSignedDataTokensFromCreds(os.Getenv("VALID_CREDS_PVT_KEY"), dataTokenOptions)
			// Assert the token response
			Expect(err).To(BeNil())
			Expect(len(tokenResp)).To(Equal(2))
			Expect(tokenResp[0].Token).To(Equal("datatoken1"))
			Expect(tokenResp[1].Token).To(Equal("datatoken2"))
			Expect(tokenResp[0].SignedToken).To(ContainSubstring("signed_token_"))
			Expect(tokenResp[1].SignedToken).To(ContainSubstring("signed_token_"))
		})
		It("should return a error when credentials are invalid", func() {
			// Call the function under test
			tokenResp, err := serviceaccount.GenerateSignedDataTokensFromCreds("invalid", dataTokenOptions)
			// Assert the token response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
		It("should return error when credentials string is empty", func() {
			tokenResp, err := serviceaccount.GenerateSignedDataTokensFromCreds("", dataTokenOptions)
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.EMPTY_CREDENTIALS_STRING))
		})
	})
	Context("GenerateSignedToken success/error response", func() {
		It("should return a valid token when credentials are valid", func() {
			// Call the function under test
			tokenResp, err := serviceaccount.GenerateSignedDataTokens(os.Getenv("CRED_FILE_PATH"), dataTokenOptions)
			// Assert the token response
			Expect(err).To(BeNil())
			Expect(len(tokenResp)).To(Equal(2))
			Expect(tokenResp[0].Token).To(Equal("datatoken1"))
			Expect(tokenResp[1].Token).To(Equal("datatoken2"))
			Expect(tokenResp[0].SignedToken).To(ContainSubstring("signed_token_"))
			Expect(tokenResp[1].SignedToken).To(ContainSubstring("signed_token_"))
		})
		It("should return a error when credentials are invalid", func() {
			// Call the function under test
			tokenResp, err := serviceaccount.GenerateSignedDataTokens("invalid.json", dataTokenOptions)
			// Assert the token response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
		It("should return a error when credentials are empty", func() {
			// Call the function under test
			tokenResp, err := serviceaccount.GenerateSignedDataTokens("", dataTokenOptions)
			// Assert the token response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
		It("should return a error when datatokens are empty", func() {
			// Call the function under test
			dataTokenOptions.DataTokens = nil
			tokenResp, err := serviceaccount.GenerateSignedDataTokens(os.Getenv("CRED_FILE_PATH"), dataTokenOptions)
			// Assert the token response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
		It("should return a error when datatokens are empty but file path is valid", func() {
			// Create a temp credentials file so we get past the path check
			tmpFile, fileErr := os.CreateTemp("", "creds_*.json")
			Expect(fileErr).To(BeNil())
			tmpFile.WriteString(`{"clientId":"x","privateKey":"y","tokenUri":"z","keyId":"k"}`)
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())

			opts := common.SignedDataTokensOptions{DataTokens: nil}
			tokenResp, err := serviceaccount.GenerateSignedDataTokens(tmpFile.Name(), opts)
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
		It("should return signed tokens from a credential file using a local RSA key", func() {
			credsJSON, srv := makeTestCredsJSONAndServer("ok")
			mockServer = srv

			tmpFile, fileErr := os.CreateTemp("", "creds_*.json")
			Expect(fileErr).To(BeNil())
			_, _ = tmpFile.WriteString(credsJSON)
			tmpFile.Close()
			defer os.Remove(tmpFile.Name())

			opts := common.SignedDataTokensOptions{
				DataTokens: []string{"tok1", "tok2"},
				TimeToLive: 60,
			}
			tokenResp, err := serviceaccount.GenerateSignedDataTokens(tmpFile.Name(), opts)
			Expect(err).To(BeNil())
			Expect(tokenResp).ToNot(BeNil())
			Expect(len(tokenResp)).To(Equal(2))
		})

	})
	Describe("IsExpired Function tests", func() {
		Context("when the token string is empty", func() {
			It("should return true", func() {
				result := serviceaccount.IsExpired("")
				Expect(result).To(BeTrue())
			})
		})

		Context("when the token string is malformed", func() {
			It("should return true", func() {
				result := serviceaccount.IsExpired("malformed.token")
				Expect(result).To(BeTrue())
			})
		})

		Context("when the token is valid and not expired", func() {
			It("should return false", func() {
				expirationTime := time.Now().Add(1 * time.Hour).Unix()
				claims := jwt.MapClaims{
					"exp": expirationTime,
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("your-secret-key"))

				result := serviceaccount.IsExpired(tokenString)
				Expect(result).To(BeFalse())
			})
		})

		Context("when the token is valid but expired", func() {
			It("should return true", func() {
				expirationTime := time.Now().Add(-1 * time.Hour).Unix()
				claims := jwt.MapClaims{
					"exp": expirationTime,
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("your-secret-key"))

				result := serviceaccount.IsExpired(tokenString)
				Expect(result).To(BeTrue())
			})
		})

		Context("when the token has an unexpected exp format", func() {
			It("should return true", func() {
				claims := jwt.MapClaims{
					"exp": "invalid-format",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("your-secret-key"))

				result := serviceaccount.IsExpired(tokenString)
				Expect(result).To(BeTrue())
			})
		})
	})

	Describe("GenerateBearerTokenFromCreds success path (local RSA key)", func() {
		It("should return an access token when credentials are valid JSON with a real RSA key", func() {
			credsJSON, srv := makeTestCredsJSONAndServer("ok")
			defer srv.Close()

			originalGetBaseURLHelper := helpers.GetBaseURLHelper
			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(_ string) (string, *skyflowError.SkyflowError) {
				return srv.URL, nil
			}

			opts := common.BearerTokenOptions{RoleIds: []string{"role1"}}
			tokenResp, err := serviceaccount.GenerateBearerTokenFromCreds(credsJSON, opts)
			Expect(err).To(BeNil())
			Expect(tokenResp).ToNot(BeNil())
			Expect(tokenResp.AccessToken).To(Equal("mockAccessToken"))
		})

		It("should return error when HTTP call fails", func() {
			credsJSON, srv := makeTestCredsJSONAndServer("err")
			defer srv.Close()

			originalGetBaseURLHelper := helpers.GetBaseURLHelper
			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(_ string) (string, *skyflowError.SkyflowError) {
				return srv.URL, nil
			}

			opts := common.BearerTokenOptions{RoleIds: []string{"role1"}}
			tokenResp, err := serviceaccount.GenerateBearerTokenFromCreds(credsJSON, opts)
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
		})
	})

	Describe("GenerateSignedDataTokensFromCreds success path (local RSA key)", func() {
		It("should return signed tokens when credentials are valid JSON with a real RSA key", func() {
			credsJSON, srv := makeTestCredsJSONAndServer("ok")
			defer srv.Close()

			opts := common.SignedDataTokensOptions{
				DataTokens: []string{"tok1", "tok2"},
				TimeToLive: 60,
			}
			tokenResp, err := serviceaccount.GenerateSignedDataTokensFromCreds(credsJSON, opts)
			Expect(err).To(BeNil())
			Expect(tokenResp).ToNot(BeNil())
			Expect(len(tokenResp)).To(Equal(2))
		})
	})
})

// makeTestCredsJSONAndServer generates a local RSA key, encodes it as a credentials JSON
// string, and starts an httptest server responding as specified by res ("ok" or "err").
func makeTestCredsJSONAndServer(res string) (string, *httptest.Server) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	pkcs1Bytes := x509.MarshalPKCS1PrivateKey(rsaKey)
	pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs1Bytes}))

	srv := mockserver(res)
	credsMap := map[string]interface{}{
		"clientId":   "test-client",
		"keyId":      "test-key",
		"tokenUri":   srv.URL + "/v1/auth/sa/oauth/token",
		"privateKey": pemKey,
	}
	b, _ := json.Marshal(credsMap)
	return string(b), srv
}
