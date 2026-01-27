package helpers_test

import (
	"encoding/json"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/core"
	. "github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	. "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Account Bearer Token Generation Helper Suite")
}

var _ = Describe("Helpers", func() {
	Context("ParseCredentialsFile", func() {
		It("should parse a valid credentials file successfully", func() {
			credentialsContent := `{"clientID":"test-client-id", "privateKey":"test-private-key"}`
			filePath := "test_credentials.json"
			ioutil.WriteFile(filePath, []byte(credentialsContent), 0644)
			defer os.Remove(filePath)

			credKeys, err := ParseCredentialsFile(filePath)

			Expect(err).To(BeNil())
			Expect(credKeys).To(HaveKeyWithValue("clientID", "test-client-id"))
			Expect(credKeys).To(HaveKeyWithValue("privateKey", "test-private-key"))
		})
		It("should fail when invalid type of private key is passes", func() {
			pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
			invalidKeyType := strings.Replace(pvtKey, "PRIVATE KEY", "PUBLIC KEY", 2)
			var credMap = map[string]interface{}{}
			_ = json.Unmarshal([]byte(invalidKeyType), &credMap)

			credKeys, err1 := ParsePrivateKey(credMap["privateKey"].(string))

			Expect(err1).ToNot(BeNil())
			Expect(credKeys).To(BeNil())
		})
		It("should return an error for an invalid file path", func() {
			_, err := ParseCredentialsFile("invalid_path.txt")

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})

		It("should return an error for an empty file", func() {
			filePath := "empty_credentials.json"
			ioutil.WriteFile(filePath, []byte(""), 0644)
			defer os.Remove(filePath)

			_, err := ParseCredentialsFile(filePath)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})
	})
	Context("GetPrivateKey", func() {
		It("should parse a valid private key successfully", func() {
			pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
			credMap := map[string]interface{}{}
			err := json.Unmarshal([]byte(pvtKey), &credMap)

			privateKey, err := GetPrivateKey(credMap)

			Expect(err).To(BeNil())
			Expect(privateKey).ToNot(BeNil())
			Expect(privateKey).To(BeAssignableToTypeOf(&rsa.PrivateKey{}))
		})

		It("should return an error for a missing private key", func() {
			credKeys := map[string]interface{}{}

			_, err := GetPrivateKey(credKeys)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})

		It("should return an error for an invalid key format", func() {
			pemKey := `INVALID PRIVATE KEY FORMAT`
			credKeys := map[string]interface{}{"privateKey": pemKey}

			_, err := GetPrivateKey(credKeys)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})
	})
	Context("GetBaseURL", func() {
		It("should return a valid base URL for a valid URL string", func() {
			urlStr := "https://example.com/some/path"

			baseURL, err := GetBaseURL(urlStr)

			Expect(err).To(BeNil())
			Expect(baseURL).To(Equal("https://example.com"))
		})

		It("should return an error for an invalid URL string", func() {
			urlStr := "invalid_url"

			_, err := GetBaseURL(urlStr)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})

		It("should return an error for a URL without protocol", func() {
			urlStr := "www.example.com"

			_, err := GetBaseURL(urlStr)

			Expect(err).NotTo(BeNil())
			Expect(err.GetCode()).To(Equal("Code: 400"))
		})
	})
	Context("ParsePrivateKey", func() {
		It("should fail a invalid PKCS1 private key successfully", func() {
			pemKey := `-----BEGIN PRIVATE KEY-----
MIIBAAIBADANINVALIDKEY==
-----END PRIVATE KEY-----`
			// Act
			privateKey, err := ParsePrivateKey(pemKey)

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
			_, parseErr := ParsePrivateKey(string(pemKey))

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
			parsedKey, parseErr := ParsePrivateKey(string(pemKey))

			// Assert
			Expect(parseErr).To(BeNil())
			Expect(parsedKey).NotTo(BeNil())
			Expect(parsedKey.Equal(rsaKey)).To(BeTrue())
		})
	})
	Context("GetCredentialParams", func() {

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
				clientID, tokenURI, keyID, err := GetCredentialParams(validCredKeys)

				Expect(clientID).To(Equal("validClientID"))
				Expect(tokenURI).To(Equal("validTokenURI"))
				Expect(keyID).To(Equal("validKeyID"))
				Expect(err).To(BeNil())
			})
		})

		Context("When one or more credential parameters are missing", func() {
			It("should return an error", func() {
				clientID, tokenURI, keyID, err := GetCredentialParams(invalidCredKeys)

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
				clientID, tokenURI, keyID, err := GetCredentialParams(emptyCredKeys)

				Expect(clientID).To(BeEmpty())
				Expect(tokenURI).To(BeEmpty())
				Expect(keyID).To(BeEmpty())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring(INVALID_CREDENTIALS))
			})
		})

	})
	Context("GetSignedDataTokens", func() {
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
				response, err = GetSignedDataTokens(credKeys, options)
				Expect(err).Should(BeNil())
				Expect(response).Should(HaveLen(2))
				Expect(response[0].Token).Should(Equal("testToken1"))
				Expect(response[0].SignedToken).Should(ContainSubstring("signed_token_"))
			})
			It("should return signed data tokens successfully when timeToLive not passed", func() {
				credKeys = getValidCreds()
				options.TimeToLive = 0
				response, err = GetSignedDataTokens(credKeys, options)
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
				response, err = GetSignedDataTokens(credMap, options)
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
				response, err = GetSignedDataTokens(credMap, options)
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
				response, err = GetSignedDataTokens(credKeys, options)
				Expect(response).Should(BeNil())
				Expect(err).ShouldNot(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400")) // Assuming 400 error for signing failure
				Expect(err.GetMessage()).Should(ContainSubstring(JWT_INVALID_FORMAT))
			})

		})
		Context("GenerateSignedDataTokensHelper", func() {
			var (
				credKeys map[string]interface{}
				options  common.SignedDataTokensOptions
				response []common.SignedDataTokensResponse
				err      *SkyflowError
			)

			BeforeEach(func() {
				credKeys = getValidCreds()
				options = common.SignedDataTokensOptions{
					DataTokens: []string{"testToken1", "testToken2"},
					TimeToLive: 3600,
					Ctx:        "testContext",
				}
			})

			Context("When tokenUri is provided in options", func() {
				It("should use the tokenUri from options if valid", func() {
					credKeys = getValidCreds()
					options.TokenURI = "https://valid-token-uri.com"
					response, err = GenerateSignedDataTokensHelper(
						credKeys["clientID"].(string),
						credKeys["keyID"].(string),
						getValidPrivateKey(),
						options,
						"https://default-uri.com",
					)
					Expect(err).Should(BeNil())
					Expect(response).Should(HaveLen(2))
				})

				It("should return error if tokenUri in options is invalid", func() {
					options.TokenURI = "http://invalid-uri.com"
					response, err = GenerateSignedDataTokensHelper("client123", "key456", nil, options, "https://default-uri.com")
					Expect(err).ShouldNot(BeNil())
					Expect(err.GetCode()).Should(Equal("Code: 400"))
					Expect(err.GetMessage()).Should(ContainSubstring(INVALID_TOKEN_URI))
					Expect(response).Should(BeNil())
				})

				It("should return error if tokenUri in options is malformed", func() {
					options.TokenURI = "not-a-valid-url"
					response, err = GenerateSignedDataTokensHelper("client123", "key456", nil, options, "https://default-uri.com")
					Expect(err).ShouldNot(BeNil())
					Expect(err.GetCode()).Should(Equal("Code: 400"))
					Expect(err.GetMessage()).Should(ContainSubstring(INVALID_TOKEN_URI))
					Expect(response).Should(BeNil())
				})

				It("should use default tokenUri if options.TokenURI is empty", func() {
					credKeys = getValidCreds()
					options.TokenURI = "" // Empty tokenUri
					response, err = GenerateSignedDataTokensHelper(
						credKeys["clientID"].(string),
						credKeys["keyID"].(string),
						getValidPrivateKey(),
						options,
						"https://default-uri.com",
					)
					Expect(err).Should(BeNil())
					Expect(response).Should(HaveLen(2))
				})
			})

			Context("When tokenUri is provided and valid", func() {
				It("should return signed data tokens successfully with valid tokenUri", func() {
					credKeys = getValidCreds()
					options.TokenURI = "https://valid-token-uri.com"
					response, err = GenerateSignedDataTokensHelper(
						credKeys["clientID"].(string),
						credKeys["keyID"].(string),
						getValidPrivateKey(),
						options,
						"https://ignored-default-uri.com",
					)
					Expect(err).Should(BeNil())
					Expect(response).Should(HaveLen(2))
					Expect(response[0].Token).Should(Equal("testToken1"))
					Expect(response[0].SignedToken).Should(ContainSubstring("signed_token_"))
				})
			})
		})
		Context("GetScopeUsingRoles", func() {
			// Test case 1: roles is nil
			It("should return an empty string when roles is nil", func() {
				result := GetScopeUsingRoles(nil)
				Expect(result).To(Equal(""))
			})

			// Test case 2: roles is an empty slice
			It("should return an empty string when roles is an empty slice", func() {
				result := GetScopeUsingRoles([]*string{})
				Expect(result).To(Equal(""))
			})

			// Test case 3: roles contains multiple roles
			It("should return a string with roles prefixed by ' role:'", func() {
				role1 := "admin"
				role2 := "user"
				role3 := "editor"
				roles := []*string{}
				roles = append(roles, &role1)
				roles = append(roles, &role2)
				roles = append(roles, &role3)
				result := GetScopeUsingRoles(roles)
				expected := " role:admin role:user role:editor"
				Expect(result).To(Equal(expected))
			})

			// Test case 4: roles contains one role
			It("should return a string with a single role", func() {
				role1 := "admin"
				roles := []*string{}
				roles = append(roles, &role1)
				result := GetScopeUsingRoles(roles)
				expected := " role:admin"
				Expect(result).To(Equal(expected))
			})

			// Test case 5: roles contains one empty string role
			It("should handle empty role string correctly", func() {
				role := ""
				roles := []*string{}
				roles = append(roles, &role)
				result := GetScopeUsingRoles(roles)
				expected := " role:"
				Expect(result).To(Equal(expected))
			})
		})
		Context("GenerateBearerTokenHelper", func() {
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
					originalGetBaseURLHelper := GetBaseURLHelper

					defer func() { GetBaseURLHelper = originalGetBaseURLHelper }()
					GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) {
						return mockServer.URL, nil
					}

					// Call the function under test
					response, err := GenerateBearerTokenHelper(credKeys, options)

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
					originalGetBaseURLHelper := GetBaseURLHelper

					defer func() { GetBaseURLHelper = originalGetBaseURLHelper }()

					GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) {
						return mockServer.URL, nil
					}

					// Call the function under test
					response, err := GenerateBearerTokenHelper(credKeys, options)

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
					response, err := GenerateBearerTokenHelper(credKeys, options)

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
					response, err := GenerateBearerTokenHelper(credKeys, options)

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
					response, err := GenerateBearerTokenHelper(credKeys, options)

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
					response, err := GenerateBearerTokenHelper(credKeys, options)

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
					response, err := GenerateBearerTokenHelper(credKeys, options)

					// Assertions
					Expect(err).ShouldNot(BeNil())
					Expect(response).Should(BeNil())
					Expect(err.GetCode()).Should(Equal("Code: 400"))
					Expect(err.GetMessage()).Should(ContainSubstring(INVALID_TOKEN_URI))
				})

			})

		})
		Context("GenerateBearerTokenHelper with tokenUri in options", func() {
			var (
				credKeys map[string]interface{}
				options  common.BearerTokenOptions
			)

			BeforeEach(func() {
				credKeys = getValidCreds()
				options = common.BearerTokenOptions{
					Ctx:     "testContext",
					RoleIDs: []string{"roleid1"},
				}
			})

			It("should use the tokenUri from options if valid", func() {
				options.TokenURI = "https://valid-token-uri.com"
				originalGetBaseURLHelper := GetBaseURLHelper

				defer func() { GetBaseURLHelper = originalGetBaseURLHelper }()

				GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) {
					return "https://valid-token-uri.com", nil
				}

				_, err := GenerateBearerTokenHelper(credKeys, options)
				Expect(err).ShouldNot(BeNil())
			})

			It("should return error if tokenUri in options is invalid (http instead of https)", func() {
				options.TokenURI = "http://invalid-uri.com"
				response, err := GenerateBearerTokenHelper(credKeys, options)

				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400"))
				Expect(err.GetMessage()).Should(ContainSubstring(INVALID_TOKEN_URI))
			})

			It("should return error if tokenUri in options is malformed", func() {
				options.TokenURI = "not-a-valid-url"
				response, err := GenerateBearerTokenHelper(credKeys, options)

				Expect(err).ShouldNot(BeNil())
				Expect(response).Should(BeNil())
				Expect(err.GetCode()).Should(Equal("Code: 400"))
				Expect(err.GetMessage()).Should(ContainSubstring(INVALID_TOKEN_URI))
			})

			It("should use credKeys tokenUri when options.TokenURI is empty", func() {
				options.TokenURI = ""
				originalGetBaseURLHelper := GetBaseURLHelper

				defer func() { GetBaseURLHelper = originalGetBaseURLHelper }()

				GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) {
					Expect(urlStr).Should(ContainSubstring(credKeys["tokenURI"].(string)))
					return "https://valid-uri.com", nil
				}

				_, err := GenerateBearerTokenHelper(credKeys, options)
				Expect(err).ShouldNot(BeNil())
			})
		})
		Context("GetHeader", func() {

			It("should return empty header and false when error is nil", func() {
				h, ok := GetHeader(nil)
				Expect(ok).To(BeFalse())
				Expect(h).To(Equal(http.Header{}))
			})

			It("should return header and true when error is core.APIError", func() {
				headers := http.Header{}
				headers.Set("X-Request-Id", "value")
				err := &core.APIError{Header: headers}
				_, ok := GetHeader(err)
				Expect(ok).To(BeTrue())
			})
			It("should return empty header and false for non-APIError", func() {
				h, ok := GetHeader(os.ErrNotExist)
				Expect(ok).To(BeFalse())
				Expect(h).To(Equal(http.Header{}))
			})
		})
		Context("Float64Ptr", func() {
			It("should return pointer to float64 value", func() {
				val := 3.14
				ptr := Float64Ptr(val)
				Expect(ptr).ToNot(BeNil())
				Expect(*ptr).To(Equal(val))
			})
		})
		Context("GetSkyflowID", func() {
			It("should return skyflow_id and true if present", func() {
				m := map[string]interface{}{"skyflow_id": "id123"}
				id, ok := GetSkyflowID(m)
				Expect(ok).To(BeTrue())
				Expect(id).To(Equal("id123"))
			})
			It("should return empty string and false if skyflow_id not present", func() {
				m := map[string]interface{}{"other": "val"}
				id, ok := GetSkyflowID(m)
				Expect(ok).To(BeFalse())
				Expect(id).To(Equal(""))
			})
			It("should return empty string and false if skyflow_id is not a string", func() {
				m := map[string]interface{}{"skyflow_id": 123}
				id, ok := GetSkyflowID(m)
				Expect(ok).To(BeFalse())
				Expect(id).To(Equal(""))
			})
		})
		Context("CreateJsonMetadata", func() {
			It("should return valid JSON string with expected keys", func() {
				jsonStr := CreateJsonMetadata()
				var m map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &m)
				Expect(err).To(BeNil())
				Expect(m).To(HaveKey("sdk_name_version"))
				Expect(m).To(HaveKey("sdk_client_device_model"))
				Expect(m).To(HaveKey("sdk_client_os_details"))
				Expect(m).To(HaveKey("sdk_runtime_details"))
			})
		})
		Context("GetDetokenizePayload", func() {
			It("should build payload with custom redaction types", func() {
				request := common.DetokenizeRequest{
					DetokenizeData: []common.DetokenizeData{
						{Token: "token1", RedactionType: "PLAIN_TEXT"},
						{Token: "token2", RedactionType: "MASKED"},
					},
				}
				options := common.DetokenizeOptions{ContinueOnError: true}
				payload := GetDetokenizePayload(request, options)
				Expect(payload.ContinueOnError).ToNot(BeNil())
				Expect(*payload.ContinueOnError).To(BeTrue())
				Expect(payload.DetokenizationParameters).To(HaveLen(2))
				Expect(payload.DetokenizationParameters[0].Token).ToNot(BeNil())
				Expect(*payload.DetokenizationParameters[0].Token).To(Equal("token1"))
				Expect(payload.DetokenizationParameters[1].Token).ToNot(BeNil())
				Expect(*payload.DetokenizationParameters[1].Token).To(Equal("token2"))
				Expect(payload.DetokenizationParameters[0].Redaction).ToNot(BeNil())
				Expect(payload.DetokenizationParameters[1].Redaction).ToNot(BeNil())
			})

			It("should use default redaction when RedactionType is empty", func() {
				request := common.DetokenizeRequest{
					DetokenizeData: []common.DetokenizeData{
						{Token: "token3", RedactionType: ""},
					},
				}
				options := common.DetokenizeOptions{ContinueOnError: false}
				payload := GetDetokenizePayload(request, options)
				Expect(payload.ContinueOnError).ToNot(BeNil())
				Expect(*payload.ContinueOnError).To(BeFalse())
				Expect(payload.DetokenizationParameters).To(HaveLen(1))
				Expect(payload.DetokenizationParameters[0].Token).ToNot(BeNil())
				Expect(*payload.DetokenizationParameters[0].Token).To(Equal("token3"))
				Expect(payload.DetokenizationParameters[0].Redaction).ToNot(BeNil())
			})

			It("should return empty parameters if DetokenizeData is empty", func() {
				request := common.DetokenizeRequest{DetokenizeData: []common.DetokenizeData{}}
				options := common.DetokenizeOptions{ContinueOnError: true}
				payload := GetDetokenizePayload(request, options)
				Expect(payload.DetokenizationParameters).To(BeNil())
			})
		})
		Context("SetTokenMode", func() {
			It("should return ENABLE_STRICT mode", func() {
				byot, err := SetTokenMode(common.ENABLE_STRICT)
				Expect(err).To(BeNil())
				Expect(byot).ToNot(BeNil())
				Expect(string(*byot)).To(Equal(string(common.ENABLE_STRICT)))
			})

			It("should return ENABLE mode", func() {
				byot, err := SetTokenMode(common.ENABLE)
				Expect(err).To(BeNil())
				Expect(byot).ToNot(BeNil())
				Expect(string(*byot)).To(Equal(string(common.ENABLE)))
			})

			It("should return DISABLE mode for unknown input", func() {
				byot, err := SetTokenMode("UNKNOWN_MODE")
				Expect(err).To(BeNil())
				Expect(byot).ToNot(BeNil())
				Expect(string(*byot)).To(Equal(string(common.DISABLE)))
			})
		})
		Context("GetFormattedGetRecord", func() {
			It("should return tokens if present", func() {
				record := vaultapis.V1FieldRecords{
					Tokens: map[string]interface{}{"field1": "token1", "field2": 123},
					Fields: map[string]interface{}{"field1": "value1", "field2": "value2"},
				}
				result := GetFormattedGetRecord(record)
				Expect(result).To(HaveKeyWithValue("field1", "token1"))
				Expect(result).To(HaveKeyWithValue("field2", 123))
				Expect(result).ToNot(HaveKeyWithValue("field1", "value1"))
			})

			It("should return fields if tokens is nil", func() {
				record := vaultapis.V1FieldRecords{
					Tokens: nil,
					Fields: map[string]interface{}{"field1": "value1", "field2": "value2"},
				}
				result := GetFormattedGetRecord(record)
				Expect(result).To(HaveKeyWithValue("field1", "value1"))
				Expect(result).To(HaveKeyWithValue("field2", "value2"))
			})

			It("should return empty map if both tokens and fields are nil", func() {
				record := vaultapis.V1FieldRecords{
					Tokens: nil,
					Fields: nil,
				}
				result := GetFormattedGetRecord(record)
				Expect(result).To(BeEmpty())
			})
		})
		Context("GetFormattedBatchInsertRecord", func() {
			It("should extract skyflow_id and tokens from valid record", func() {
				record := map[string]interface{}{
					"Body": map[string]interface{}{
						"records": []interface{}{
							map[string]interface{}{
								"skyflow_id": "id123",
								"tokens":     map[string]interface{}{"field1": "token1"},
							},
						},
					},
				}
				result, err := GetFormattedBatchInsertRecord(record, 0)
				Expect(err).To(BeNil())
				Expect(result).To(HaveKeyWithValue("skyflow_id", "id123"))
				Expect(result).To(HaveKeyWithValue("field1", "token1"))
				Expect(result).To(HaveKeyWithValue("request_index", 0))
			})

			It("should extract error field if present", func() {
				record := map[string]interface{}{
					"Body": map[string]interface{}{
						"error":   "some error",
						"records": []interface{}{},
					},
				}
				result, err := GetFormattedBatchInsertRecord(record, 2)
				Expect(err).To(BeNil())
				Expect(result).To(HaveKeyWithValue("error", "some error"))
				Expect(result).To(HaveKeyWithValue("request_index", 2))
			})

			It("should return error if Body is missing", func() {
				record := map[string]interface{}{}
				result, err := GetFormattedBatchInsertRecord(record, 1)
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return error if record is not valid JSON", func() {
				ch := make(chan int) // not serializable
				result, err := GetFormattedBatchInsertRecord(ch, 1)
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
			})
		})
		Context("GetFormattedQueryRecord", func() {
			It("should return fields map", func() {
				record := vaultapis.V1FieldRecords{
					Fields: map[string]interface{}{"f1": "v1", "f2": 2},
				}
				result := GetFormattedQueryRecord(record)
				Expect(result).To(HaveKeyWithValue("f1", "v1"))
				Expect(result).To(HaveKeyWithValue("f2", 2))
			})
			It("should return empty map if fields is nil", func() {
				record := vaultapis.V1FieldRecords{}
				result := GetFormattedQueryRecord(record)
				Expect(result).To(BeEmpty())
			})
		})
		Context("GetFormattedUpdateRecord", func() {
			It("should return tokens map", func() {
				record := vaultapis.V1UpdateRecordResponse{
					Tokens: map[string]interface{}{"f": "t"},
				}
				result := GetFormattedUpdateRecord(record)
				Expect(result).To(HaveKeyWithValue("f", "t"))
			})
			It("should return empty map if tokens is nil", func() {
				record := vaultapis.V1UpdateRecordResponse{}
				result := GetFormattedUpdateRecord(record)
				Expect(result).To(BeEmpty())
			})
		})
		Context("CreateInsertBulkBodyRequest", func() {
			It("should create valid insert body", func() {
				request := &common.InsertRequest{Values: []map[string]interface{}{{"a": 1}}}
				options := &common.InsertOptions{Upsert: "true", ReturnTokens: true, TokenMode: common.ENABLE}
				body, err := CreateInsertBulkBodyRequest(request, options)
				Expect(err).To(BeNil())
				Expect(body).ToNot(BeNil())
				Expect(body.Records).To(HaveLen(1))
			})
		})
		Context("CreateInsertBatchBodyRequest", func() {
			It("should create valid batch body", func() {
				request := &common.InsertRequest{Table: "t", Values: []map[string]interface{}{{"a": 1}}}
				options := &common.InsertOptions{Upsert: "true", ReturnTokens: true, TokenMode: common.ENABLE}
				body, err := CreateInsertBatchBodyRequest(request, options)
				Expect(err).To(BeNil())
				Expect(body).ToNot(BeNil())
				Expect(body.Records).To(HaveLen(1))
			})
		})
		Context("GetTokenizePayload", func() {
			It("should build payload from requests", func() {
				requests := []common.TokenizeRequest{{Value: "v", ColumnGroup: "cg"}}
				payload := GetTokenizePayload(requests)
				Expect(payload.TokenizationParameters).To(HaveLen(1))
				Expect(payload.TokenizationParameters[0].Value).ToNot(BeNil())
				Expect(*payload.TokenizationParameters[0].Value).To(Equal("v"))
				Expect(payload.TokenizationParameters[0].ColumnGroup).ToNot(BeNil())
				Expect(*payload.TokenizationParameters[0].ColumnGroup).To(Equal("cg"))
			})
		})
		Context("GetURLWithEnv", func() {
			It("should return correct URL for each env", func() {
				Expect(GetURLWithEnv(common.DEV, "cid")).To(ContainSubstring("cid"))
				Expect(GetURLWithEnv(common.PROD, "cid")).To(ContainSubstring("cid"))
				Expect(GetURLWithEnv(common.STAGE, "cid")).To(ContainSubstring("cid"))
				Expect(GetURLWithEnv(common.SANDBOX, "cid")).To(ContainSubstring("cid"))
			})
		})
		Context("ParseTokenizeResponse", func() {
			It("should parse tokens from response", func() {
				resp := vaultapis.V1TokenizeResponse{}
				token := "token"
				record := vaultapis.V1TokenizeRecordResponse{Token: &token}
				records := []*vaultapis.V1TokenizeRecordResponse{&record}
				resp.Records = records
				result := ParseTokenizeResponse(resp)
				Expect(result.Tokens).To(ContainElement("token"))
			})
		})
		Context("GetFileForFileUpload", func() {
			It("should not return error for empty input", func() {
				_, err := GetFileForFileUpload(common.FileUploadRequest{})
				Expect(err).To(BeNil())
			})
			It("should return error for invalid file path", func() {
				_, err := GetFileForFileUpload(common.FileUploadRequest{FilePath: "invalid_path.txt"})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("open invalid_path.txt: no such file or directory"))
			})
			It("should not return error for valid file path", func() {
				_, err := GetFileForFileUpload(common.FileUploadRequest{FilePath: "../../../credentials.json"})
				Expect(err).To(BeNil())
			})
			It("should return error for invalid base64 data", func() {
				_, err := GetFileForFileUpload(common.FileUploadRequest{Base64: "invalid_base64"})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("illegal base64 data at input byte"))
			})
			// create file object
			It("should not return error for valid base64 data", func() {
				data := "SGVsbG8sIFdvcmxkIQ==" // base64 for "Hello, World!"
				file, err := GetFileForFileUpload(common.FileUploadRequest{Base64: data, FileName: "hello.txt"})
				Expect(err).To(BeNil())
				Expect(file).ToNot(BeNil())
			})
			It("should return error for valid base64 data when file name is not passed", func() {
				data := "SGVsbG8sIFdvcmxkIQ==" // base64 for "Hello, World!"
				file, err := GetFileForFileUpload(common.FileUploadRequest{Base64: data})
				Expect(err).ToNot(BeNil())
				Expect(file).To(BeNil())
			})
			It("should not return error for valid file object", func() {
				tmpfile, err := os.Open("../../../credentials.json")
				Expect(err).To(BeNil())
				file, err := GetFileForFileUpload(common.FileUploadRequest{FileObject: *tmpfile})
				Expect(err).To(BeNil())
				Expect(file).ToNot(BeNil())
				defer tmpfile.Close()
			})
			It("should return error for nil file object", func() {
				// empty file object
				var tmpfile os.File = os.File{}
				file, err := GetFileForFileUpload(common.FileUploadRequest{FileObject: tmpfile})
				Expect(err).To(BeNil())
				Expect(file).To(BeNil())
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

func getValidPrivateKey() *rsa.PrivateKey {
	credKeys := getValidCreds()
	pvtKey, _ := GetPrivateKey(credKeys)
	return pvtKey
}
