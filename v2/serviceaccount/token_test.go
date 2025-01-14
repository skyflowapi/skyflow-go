package serviceaccount_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
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
			RoleIDs: []string{"roleid1", "roleid2"},
		}
		dataTokenOptions = common.SignedDataTokensOptions{
			DataTokens: []string{"datatoken1", "datatoken2"},
			TimeToLive: 0,
			Ctx:        "testContext",
			LogLevel:   0,
		}
	})
	AfterEach(func() {
		mockServer.Close()
	})

	Context("GenerateBearerToken success/error response", func() {
		It("should return a valid token when credentials are valid", func() {
			mockServer = mockserver("ok")
			// Prepare valid BearerTokenOptions
			options = common.BearerTokenOptions{
				RoleIDs: []string{"role1"},
			}
			originalGetBaseURLHelper := helpers.GetBaseURLHelper

			defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
			helpers.GetBaseURLHelper = func(urlStr string) (string, *skyflowError.SkyflowError) {
				return mockServer.URL, nil
			}
			if os.Getenv("CRED_FILE_PATH") != "" {
				var file = os.Getenv("CRED_FILE_PATH")
				fmt.Println("file path is ", file)
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
				RoleIDs: []string{"role1"},
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
				RoleIDs: []string{"role1"},
			}

			// Call the function under test
			tokenResp, err := serviceaccount.GenerateBearerToken("credentials.json", options)

			// Assert the error response
			Expect(err).ToNot(BeNil())
			Expect(tokenResp).To(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(skyflowError.FILE_NOT_FOUND, "credentials.json")))
		})
	})
	Context("GenerateBearerTokenCreds success/error response", func() {
		It("should return a valid token when credentials are valid", func() {
			mockServer = mockserver("ok")
			// Prepare valid BearerTokenOptions
			options = common.BearerTokenOptions{
				RoleIDs: []string{"role1"},
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
				RoleIDs: []string{"role1"},
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
				RoleIDs: []string{"role1"},
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
})
