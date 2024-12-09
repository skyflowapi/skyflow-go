package helpers_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"skyflow-go/v2/internal/helpers"
	"skyflow-go/v2/utils/common"
	. "skyflow-go/v2/utils/error"
	"strings"
	"testing"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Account Bearer Token Generation Helper Suite")
}

var _ = Describe("Helpers", func() {
	Describe("ParseCredentialsFile", func() {
		It("should parse a valid credentials file successfully", func() {
			credentialsContent := `{"clientID":"test-client-id", "privateKey":"test-private-key"}`
			filePath := "test_credentials.json"
			ioutil.WriteFile(filePath, []byte(credentialsContent), 0644)
			defer os.Remove(filePath)

			credKeys, err := helpers.ParseCredentialsFile(filePath)

			Expect(err).To(BeNil())
			Expect(credKeys).To(HaveKeyWithValue("clientID", "test-client-id"))
			Expect(credKeys).To(HaveKeyWithValue("privateKey", "test-private-key"))
		})
		It("should fail when invalid type of private key is passes", func() {
			pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
			invalidKeyType := strings.Replace(pvtKey, "PRIVATE KEY", "PUBLIC KEY", 2)
			var credMap = map[string]interface{}{}
			_ = json.Unmarshal([]byte(invalidKeyType), &credMap)

			credKeys, err1 := helpers.ParsePrivateKey(credMap["privateKey"].(string))

			Expect(err1).ToNot(BeNil())
			Expect(credKeys).To(BeNil())
		})
		It("should return an error for an invalid file path", func() {
			_, err := helpers.ParseCredentialsFile("invalid_path.txt")

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})

		It("should return an error for an empty file", func() {
			filePath := "empty_credentials.json"
			ioutil.WriteFile(filePath, []byte(""), 0644)
			defer os.Remove(filePath)

			_, err := helpers.ParseCredentialsFile(filePath)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})
	})
	Describe("GetPrivateKey", func() {
		It("should parse a valid private key successfully", func() {
			pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
			credMap := map[string]interface{}{}
			err := json.Unmarshal([]byte(pvtKey), &credMap)

			privateKey, err := helpers.GetPrivateKey(credMap)

			Expect(err).To(BeNil())
			Expect(privateKey).ToNot(BeNil())
			Expect(privateKey).To(BeAssignableToTypeOf(&rsa.PrivateKey{}))
		})

		It("should return an error for a missing private key", func() {
			credKeys := map[string]interface{}{}

			_, err := helpers.GetPrivateKey(credKeys)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})

		It("should return an error for an invalid key format", func() {
			pemKey := `INVALID PRIVATE KEY FORMAT`
			credKeys := map[string]interface{}{"privateKey": pemKey}

			_, err := helpers.GetPrivateKey(credKeys)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})
	})
	Describe("GetBaseURL", func() {
		It("should return a valid base URL for a valid URL string", func() {
			urlStr := "https://example.com/some/path"

			baseURL, err := helpers.GetBaseURL(urlStr)

			Expect(err).To(BeNil())
			Expect(baseURL).To(Equal("https://example.com"))
		})

		It("should return an error for an invalid URL string", func() {
			urlStr := "invalid_url"

			_, err := helpers.GetBaseURL(urlStr)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})

		It("should return an error for a URL without protocol", func() {
			urlStr := "www.example.com"

			_, err := helpers.GetBaseURL(urlStr)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})
	})
	Describe("ParsePrivateKey", func() {
		It("should fail a invalid PKCS1 private key successfully", func() {
			pemKey := `-----BEGIN PRIVATE KEY-----
MIIBAAIBADANINVALIDKEY==
-----END PRIVATE KEY-----`
			// Act
			privateKey, err := helpers.ParsePrivateKey(pemKey)

			// Assert
			Expect(err).ToNot(BeNil())
			Expect(privateKey).To(BeNil())
		})
		It("should return an error for a valid PKCS8 key but invalid key type", func() {
			// Arrange
			// Generate an ECDSA key, which will not be of type *rsa.PrivateKey
			ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			Expect(err).To(BeNil())

			// Convert the ECDSA key to PKCS8 format
			pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(ecdsaKey)
			Expect(err).To(BeNil())

			// Encode the PKCS8 key into PEM format
			pemKey := pem.EncodeToMemory(&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: pkcs8Bytes,
			})

			// Act
			_, parseErr := helpers.ParsePrivateKey(string(pemKey))

			// Assert
			Expect(parseErr).NotTo(BeNil())
			Expect(parseErr.GetCode()).To(Equal("Code: 400"))
			Expect(parseErr.GetMessage()).To(ContainSubstring(INVALID_KEY_SPEC))
		})
		It("should successfully parse a valid PKCS1 private key", func() {
			// Arrange
			// Generate a valid RSA private key
			rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
			Expect(err).To(BeNil())

			// Convert the RSA key to PKCS1 format
			pkcs1Bytes := x509.MarshalPKCS1PrivateKey(rsaKey)

			// Encode the PKCS1 key into PEM format
			pemKey := pem.EncodeToMemory(&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: pkcs1Bytes,
			})

			// Act
			parsedKey, parseErr := helpers.ParsePrivateKey(string(pemKey))

			// Assert
			Expect(parseErr).To(BeNil())
			Expect(parsedKey).NotTo(BeNil())
			Expect(parsedKey.Equal(rsaKey)).To(BeTrue())
		})
	})
	Describe("GetCredentialParams", func() {

		var validCredKeys map[string]interface{}
		var invalidCredKeys map[string]interface{}

		BeforeEach(func() {
			// Setting up valid and invalid credential maps before each test
			validCredKeys = map[string]interface{}{
				"clientID": "validClientID",
				"tokenURI": "validTokenURI",
				"keyID":    "validKeyID",
			}
			invalidCredKeys = map[string]interface{}{
				"clientID": "validClientID",
				// Missing tokenURI
				"keyID": "validKeyID",
			}
		})

		Context("When all credential parameters are valid", func() {
			It("should return clientID, tokenURI, keyID and no error", func() {
				clientID, tokenURI, keyID, err := helpers.GetCredentialParams(validCredKeys)

				Expect(clientID).To(Equal("validClientID"))
				Expect(tokenURI).To(Equal("validTokenURI"))
				Expect(keyID).To(Equal("validKeyID"))
				Expect(err).To(BeNil())
			})
		})

		Context("When one or more credential parameters are missing", func() {
			It("should return an error", func() {
				clientID, tokenURI, keyID, err := helpers.GetCredentialParams(invalidCredKeys)

				Expect(clientID).To(BeEmpty())
				Expect(tokenURI).To(BeEmpty())
				Expect(keyID).To(BeEmpty())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring(INVALID_CREDENTIALS))
			})
		})

		Context("When all credential parameters are missing", func() {
			It("should return an error", func() {
				emptyCredKeys := make(map[string]interface{})
				clientID, tokenURI, keyID, err := helpers.GetCredentialParams(emptyCredKeys)

				Expect(clientID).To(BeEmpty())
				Expect(tokenURI).To(BeEmpty())
				Expect(keyID).To(BeEmpty())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring(INVALID_CREDENTIALS))
			})
		})

	})
	Describe("GetSignedDataTokens", func() {
		var (
			credKeys map[string]interface{}
			options  common.SignedDataTokensOptions
			response []common.SignedDataTokensResponse
			err      *SkyflowError
		)

		BeforeEach(func() {
			// Prepare the mock credentials map
			credKeys = map[string]interface{}{
				"clientID":   "client_123",
				"keyID":      "key_456",
				"tokenURI":   "http://example.com",
				"privateKey": "mockPrivateKey", // This should be a mock or a valid private key
			}

			options = common.SignedDataTokensOptions{
				DataTokens: []string{"testToken1", "testToken2"},
				TimeToLive: 3600, // 1 hour TTL
				Ctx:        "testContext",
			}
		})

		Context("When all credentials and options are valid", func() {
			It("should return signed data tokens successfully", func() {
				credKeys = getValidCreds()
				response, err = helpers.GetSignedDataTokens(credKeys, options)
				Expect(err).Should(BeNil())
				Expect(response).Should(HaveLen(2))
				Expect(response[0].Token).Should(Equal("testToken1"))
				Expect(response[0].SignedToken).Should(ContainSubstring("signed_token_"))
			})
			It("should return signed data tokens successfully when timeToLive not passed", func() {
				credKeys = getValidCreds()
				options.TimeToLive = 0
				response, err = helpers.GetSignedDataTokens(credKeys, options)
				Expect(err).Should(BeNil())
				Expect(response).Should(HaveLen(2))
				Expect(response[0].Token).Should(Equal("testToken1"))
				Expect(response[0].SignedToken).Should(ContainSubstring("signed_token_"))
			})
		})

		Context("When private key retrieval fails", func() {
			It("should return an error", func() {
				// Simulate an error in GetPrivateKey
				pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
				credMap := map[string]interface{}{}
				_ = json.Unmarshal([]byte(pvtKey), &credMap)
				credMap["privateKey"] = nil // Invalidate the private key
				response, err = helpers.GetSignedDataTokens(credMap, options)
				Expect(response).Should(BeNil())
				Expect(err).ShouldNot(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400")) // Assuming a 400 error code for this case
				Expect(err.GetMessage()).Should(ContainSubstring(MISSING_PRIVATE_KEY))
			})
		})

		Context("When credential parameters retrieval fails", func() {
			It("should return an error", func() {
				// Simulate an error in GetCredentialParams
				pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
				credMap := map[string]interface{}{}
				_ = json.Unmarshal([]byte(pvtKey), &credMap)
				credMap["clientID"] = nil // Invalidate the clientID
				response, err = helpers.GetSignedDataTokens(credMap, options)
				Expect(response).Should(BeNil())
				Expect(err).ShouldNot(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400")) // Assuming a 400 error code for this case
				Expect(err.GetMessage()).Should(ContainSubstring(INVALID_CREDENTIALS))
			})
		})

		Context("When GenerateSignedDataTokensHelper returns an error", func() {
			It("should propagate the error", func() {
				invalidPrivateKey := "invalidKey"
				credKeys["privateKey"] = invalidPrivateKey
				response, err = helpers.GetSignedDataTokens(credKeys, options)
				Expect(response).Should(BeNil())
				Expect(err).ShouldNot(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400")) // Assuming 400 error for signing failure
				Expect(err.GetMessage()).Should(ContainSubstring(JWT_INVALID_FORMAT))
			})

		})
		Context("GetScopeUsingRoles", func() {
			// Test case 1: roles is nil
			It("should return an empty string when roles is nil", func() {
				result := helpers.GetScopeUsingRoles(nil)
				Expect(result).To(Equal(""))
			})

			// Test case 2: roles is an empty slice
			It("should return an empty string when roles is an empty slice", func() {
				result := helpers.GetScopeUsingRoles([]string{})
				Expect(result).To(Equal(""))
			})

			// Test case 3: roles contains multiple roles
			It("should return a string with roles prefixed by ' role:'", func() {
				roles := []string{"admin", "user", "editor"}
				result := helpers.GetScopeUsingRoles(roles)
				expected := " role:admin role:user role:editor"
				Expect(result).To(Equal(expected))
			})

			// Test case 4: roles contains one role
			It("should return a string with a single role", func() {
				roles := []string{"admin"}
				result := helpers.GetScopeUsingRoles(roles)
				expected := " role:admin"
				Expect(result).To(Equal(expected))
			})

			// Test case 5: roles contains one empty string role
			It("should handle empty role string correctly", func() {
				roles := []string{""}
				result := helpers.GetScopeUsingRoles(roles)
				expected := " role:"
				Expect(result).To(Equal(expected))
			})
		})
	})
	Describe("GenerateBearerTokenHelper", func() {
		var (
			credKeys   map[string]interface{}
			options    common.BearerTokenOptions
			mockServer *httptest.Server
		)

		BeforeEach(func() {
			credKeys = map[string]interface{}{
				"privateKey": "dummyPrivateKey",
				"clientID":   "client_123",
				"tokenURI":   "http://mock-api.com/token",
				"keyID":      "key_456",
			}
			options = common.BearerTokenOptions{
				Ctx:     "testContext",
				RoleIDs: []string{"roleid1", "roleid2"},
			}
		})

		AfterEach(func() {
			mockServer.Close()
		})

		Context("When the API call is successful", func() {
			It("should return a valid access token", func() {
				// Set the base URL for the mock server
				credKeys = getValidCreds()
				mockServer = mockserver("ok")
				credKeys["tokenURI"] = mockServer.URL
				originalGetBaseURLHelper := helpers.GetBaseURLHelper

				defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()
				helpers.GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) {
					return mockServer.URL, nil
				}

				// Call the function under test
				response, err := helpers.GenerateBearerTokenHelper(credKeys, options)

				// Assertions
				Expect(err).Should(BeNil())
				Expect(response).ShouldNot(BeNil())
				Expect(*response.AccessToken).Should(Equal("mockAccessToken"))
			})
			It("should return a error", func() {
				// Set the base URL for the mock server
				credKeys = getValidCreds()
				credKeys["tokenURI"] = mockServer.URL
				mockServer = mockserver("err")
				originalGetBaseURLHelper := helpers.GetBaseURLHelper

				defer func() { helpers.GetBaseURLHelper = originalGetBaseURLHelper }()

				helpers.GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) {
					return mockServer.URL, nil
				}

				// Call the function under test
				response, err := helpers.GenerateBearerTokenHelper(credKeys, options)

				// Assertions
				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
			})

		})

		Context("When the keys are missing", func() {
			It("should return an error when privateKey is missing", func() {
				// Remove privateKey from credKeys to simulate missing key
				credKeys = getValidCreds()
				delete(credKeys, "privateKey")

				// Call the function under test
				response, err := helpers.GenerateBearerTokenHelper(credKeys, options)

				// Assertions
				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400"))
				Expect(err.GetMessage()).Should(ContainSubstring(MISSING_PRIVATE_KEY))
			})
			It("should return an error when clientID is missing", func() {
				// Remove privateKey from credKeys to simulate missing key
				credKeys = getValidCreds()
				delete(credKeys, "clientID")
				// Call the function under test
				response, err := helpers.GenerateBearerTokenHelper(credKeys, options)

				// Assertions
				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400"))
				Expect(err.GetMessage()).Should(ContainSubstring(MISSING_CLIENT_ID))
			})
			It("should return an error when tokenURI is missing", func() {
				// Remove privateKey from credKeys to simulate missing key
				credKeys = getValidCreds()
				delete(credKeys, "tokenURI")
				// Call the function under test
				response, err := helpers.GenerateBearerTokenHelper(credKeys, options)

				// Assertions
				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400"))
				Expect(err.GetMessage()).Should(ContainSubstring(MISSING_TOKEN_URI))
			})
			It("should return an error when keyID is missing", func() {
				// Remove privateKey from credKeys to simulate missing key
				credKeys = getValidCreds()
				delete(credKeys, "keyID")
				// Call the function under test
				response, err := helpers.GenerateBearerTokenHelper(credKeys, options)

				// Assertions
				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400"))
				Expect(err.GetMessage()).Should(ContainSubstring(MISSING_KEY_ID))
			})
			It("should return an error when invalid token uri passed", func() {
				// Remove privateKey from credKeys to simulate missing key
				credKeys = getValidCreds()
				credKeys["tokenURI"] = ""
				// Call the function under test
				response, err := helpers.GenerateBearerTokenHelper(credKeys, options)

				// Assertions
				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400"))
				Expect(err.GetMessage()).Should(ContainSubstring(INVALID_TOKEN_URI))
			})

		})
	})
})

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
		w.Write([]byte(`{"error":{"grpc_code":16,"http_code":401,"message":"invalid_grant: Invalid Audience https://demo.com","http_status":"Unauthorized","details":[]}} `))
	}))
	return mockServers
}
func getValidCreds() map[string]interface{} {
	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	credMap := map[string]interface{}{}
	_ = json.Unmarshal([]byte(pvtKey), &credMap)
	return credMap
}
