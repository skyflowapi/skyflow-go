package helpers_test

import (
	"encoding/json"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
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
	"github.com/skyflowapi/skyflow-go/v2/internal/constants"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Account Bearer Token Generation Helper Suite")
}

var _ = Describe("Helpers", func() {
	Context("ParseCredentialsFile", func() {
		It("should parse a valid credentials file successfully", func() {
			credentialsContent := `{"clientId":"test-client-id", "privateKey":"test-private-key"}`
			filePath := "test_credentials.json"
			ioutil.WriteFile(filePath, []byte(credentialsContent), 0644)
			defer os.Remove(filePath)

			credKeys, err := ParseCredentialsFile(filePath)

			Expect(err).To(BeNil())
			Expect(credKeys).To(HaveKeyWithValue("clientId", "test-client-id"))
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
				"clientId": "validclientId",
				"tokenUri": "validtokenUri",
				"keyId":    "validkeyId",
			}
			invalidCredKeys = map[string]interface{}{
				"clientId": "validclientId",
				// Missing tokenUri
				"keyId": "validkeyId",
			}
		})

		Context("When all credential parameters are valid", func() {
			It("should return clientId, tokenUri, keyId and no error", func() {
				clientId, tokenUri, keyId, err := GetCredentialParams(validCredKeys)

				Expect(clientId).To(Equal("validclientId"))
				Expect(tokenUri).To(Equal("validtokenUri"))
				Expect(keyId).To(Equal("validkeyId"))
				Expect(err).To(BeNil())
			})
		})

		Context("When one or more credential parameters are missing", func() {
			It("should return an error", func() {
				clientId, tokenUri, keyId, err := GetCredentialParams(invalidCredKeys)

				Expect(clientId).To(BeEmpty())
				Expect(tokenUri).To(BeEmpty())
				Expect(keyId).To(BeEmpty())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring(MISSING_TOKEN_URI))
			})
		})

		Context("When all credential parameters are missing", func() {
			It("should return an error", func() {
				emptyCredKeys := make(map[string]interface{})
				clientId, tokenUri, keyId, err := GetCredentialParams(emptyCredKeys)

				Expect(clientId).To(BeEmpty())
				Expect(tokenUri).To(BeEmpty())
				Expect(keyId).To(BeEmpty())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring(MISSING_CLIENT_ID))
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
				"clientId":   "client_123",
				"keyId":      "key_456",
				"tokenUri":   "http://example.com",
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
				Expect(err.GetMessage()).Should(ContainSubstring(MISSING_CLIENT_ID))
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
					"clientId":   "client_123",
					"tokenUri":   "http://mock-api.com/token",
					"keyId":      "key_456",
				}
				options = common.BearerTokenOptions{
					Ctx:     "testContext",
					RoleIds: []string{"roleid1", "roleid2"},
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
					credKeys["tokenUri"] = mockServer.URL
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
				It("should return an error when clientId is missing", func() {
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
				It("should return an error when tokenUri is missing", func() {
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
				It("should return an error when keyId is missing", func() {
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
				m := map[string]interface{}{"SkyflowId": "id123"}
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
				m := map[string]interface{}{"SkyflowId": 123}
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
				Expect(result).To(HaveKeyWithValue("SkyflowId", "id123"))
				Expect(result).To(HaveKeyWithValue("field1", "token1"))
				Expect(result).To(HaveKeyWithValue(internal.JSON_KEY_REQUEST_INDEX, 0))
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
				Expect(result).To(HaveKeyWithValue(internal.JSON_KEY_REQUEST_INDEX, 2))
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
			It("should not return error for valid base64 data when file name is not passed", func() {
				data := "SGVsbG8sIFdvcmxkIQ==" // base64 for "Hello, World!"
				file, err := GetFileForFileUpload(common.FileUploadRequest{Base64: data})
				Expect(err).To(BeNil())
				Expect(file).ToNot(BeNil())
				type namer interface{ Name() string }
				named, ok := file.(namer)
				Expect(ok).To(BeTrue())
				Expect(named.Name()).To(Equal(""))
			})
			It("should return namedReader with correct Name() for valid base64 data", func() {
				data := "SGVsbG8sIFdvcmxkIQ==" // base64 for "Hello, World!"
				file, err := GetFileForFileUpload(common.FileUploadRequest{Base64: data, FileName: "hello.txt"})
				Expect(err).To(BeNil())
				Expect(file).ToNot(BeNil())
				type namer interface{ Name() string }
				named, ok := file.(namer)
				Expect(ok).To(BeTrue())
				Expect(named.Name()).To(Equal("hello.txt"))
			})
			It("should return namedReader with correct content for valid base64 data", func() {
				data := "SGVsbG8sIFdvcmxkIQ==" // base64 for "Hello, World!"
				file, err := GetFileForFileUpload(common.FileUploadRequest{Base64: data, FileName: "hello.txt"})
				Expect(err).To(BeNil())
				Expect(file).ToNot(BeNil())
				content, readErr := io.ReadAll(file)
				Expect(readErr).To(BeNil())
				Expect(string(content)).To(Equal("Hello, World!"))
			})
			It("should close namedReader without error", func() {
				data := "SGVsbG8sIFdvcmxkIQ==" // base64 for "Hello, World!"
				file, err := GetFileForFileUpload(common.FileUploadRequest{Base64: data, FileName: "hello.txt"})
				Expect(err).To(BeNil())
				Expect(file).ToNot(BeNil())
				Expect(file.Close()).To(BeNil())
			})
			It("should return error containing failed to decode for invalid base64 data", func() {
				_, err := GetFileForFileUpload(common.FileUploadRequest{Base64: "!!!invalid!!!", FileName: "test.txt"})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("Failed to decode base64"))
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

		Context("ValidateAndResolveCtx", func() {
			It("should return nil, nil when input is nil", func() {
				result, err := ValidateAndResolveCtx(nil)
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return nil, nil when input is an empty string", func() {
				result, err := ValidateAndResolveCtx("")
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return the string when input is a valid non-empty string", func() {
				result, err := ValidateAndResolveCtx("testContext")
				Expect(err).To(BeNil())
				Expect(result).To(Equal("testContext"))
			})

			It("should return nil, nil when input is an empty map", func() {
				result, err := ValidateAndResolveCtx(map[string]interface{}{})
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return the map when input is a valid map with simple keys", func() {
				input := map[string]interface{}{"key1": "value1", "key2": "value2"}
				result, err := ValidateAndResolveCtx(input)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(input))
			})

			It("should return the map when keys contain alphanumeric characters and underscores", func() {
				input := map[string]interface{}{"abc_123": "val", "ABC_XYZ": "val2", "a1_B2_c3": "val3"}
				result, err := ValidateAndResolveCtx(input)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(input))
			})

			It("should return error when map key contains a hyphen", func() {
				input := map[string]interface{}{"invalid-key": "value"}
				result, err := ValidateAndResolveCtx(input)
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring("invalid-key"))
			})

			It("should return error when map key contains a space", func() {
				input := map[string]interface{}{"invalid key": "value"}
				result, err := ValidateAndResolveCtx(input)
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring("invalid key"))
			})

			It("should return error when map key contains a dot", func() {
				input := map[string]interface{}{"invalid.key": "value"}
				result, err := ValidateAndResolveCtx(input)
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring("invalid.key"))
			})

			It("should return float64 when input is an int", func() {
				result, err := ValidateAndResolveCtx(42)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(float64(42)))
			})

			It("should return float64 when input is a float64", func() {
				result, err := ValidateAndResolveCtx(3.14)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(3.14))
			})

			It("should return bool when input is true", func() {
				result, err := ValidateAndResolveCtx(true)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(true))
			})

			It("should return bool when input is false", func() {
				result, err := ValidateAndResolveCtx(false)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(false))
			})

			It("should return error when input is a slice (invalid type)", func() {
				result, err := ValidateAndResolveCtx([]string{"a", "b"})
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(err.GetMessage()).To(ContainSubstring(INVALID_CTX_TYPE))
			})

			It("should return the map when values are mixed types (string, int, bool)", func() {
				input := map[string]interface{}{"name": "test", "count": 5, "active": true}
				result, err := ValidateAndResolveCtx(input)
				Expect(err).To(BeNil())
				resultMap := result.(map[string]interface{})
				Expect(resultMap["name"]).To(Equal("test"))
				Expect(resultMap["count"]).To(Equal(5))
				Expect(resultMap["active"]).To(Equal(true))
			})

			It("should return the map when values contain nested objects", func() {
				input := map[string]interface{}{
					"outer": map[string]interface{}{"inner": "value"},
					"list":  []interface{}{1, 2, 3},
				}
				result, err := ValidateAndResolveCtx(input)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(input))
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

var _ = Describe("GetFormattedBulkInsertRecord", func() {
	Context("when SkyflowId is nil", func() {
		It("should return a map without SkyflowId key", func() {
			record := vaultapis.V1RecordMetaProperties{}
			result := GetFormattedBulkInsertRecord(record)
			Expect(result).ToNot(HaveKey("SkyflowId"))
		})
	})

	Context("when SkyflowId is set", func() {
		It("should include SkyflowId in the map", func() {
			id := "sky-123"
			record := vaultapis.V1RecordMetaProperties{SkyflowId: &id}
			result := GetFormattedBulkInsertRecord(record)
			Expect(result).To(HaveKeyWithValue("SkyflowId", "sky-123"))
		})

		It("should include tokens alongside SkyflowId", func() {
			id := "sky-456"
			record := vaultapis.V1RecordMetaProperties{
				SkyflowId: &id,
				Tokens:    map[string]interface{}{"card_number": "tok_abc", "cvv": "tok_xyz"},
			}
			result := GetFormattedBulkInsertRecord(record)
			Expect(result).To(HaveKeyWithValue("SkyflowId", "sky-456"))
			Expect(result).To(HaveKeyWithValue("card_number", "tok_abc"))
			Expect(result).To(HaveKeyWithValue("cvv", "tok_xyz"))
		})
	})

	Context("when tokens are empty", func() {
		It("should return a map with only SkyflowId", func() {
			id := "sky-789"
			record := vaultapis.V1RecordMetaProperties{SkyflowId: &id}
			result := GetFormattedBulkInsertRecord(record)
			Expect(result).To(HaveLen(1))
			Expect(result).To(HaveKey("SkyflowId"))
		})
	})
})

var _ = Describe("GetFormattedQueryRecord — additional paths", func() {
	Context("when fields contain skyflow_id wire key", func() {
		It("should remap skyflow_id to SkyflowId", func() {
			record := vaultapis.V1FieldRecords{
				Fields: map[string]interface{}{"skyflow_id": "rec-001", "name": "alice"},
			}
			result := GetFormattedQueryRecord(record)
			Expect(result).To(HaveKeyWithValue("SkyflowId", "rec-001"))
			Expect(result).To(HaveKeyWithValue("name", "alice"))
			Expect(result).ToNot(HaveKey("skyflow_id"))
		})
	})

	Context("when both fields and tokens are set", func() {
		It("should include TokenizedData map alongside fields", func() {
			record := vaultapis.V1FieldRecords{
				Fields: map[string]interface{}{"name": "bob"},
				Tokens: map[string]interface{}{"card": "tok_card"},
			}
			result := GetFormattedQueryRecord(record)
			Expect(result).To(HaveKeyWithValue("name", "bob"))
			tokenizedData, ok := result["TokenizedData"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(tokenizedData).To(HaveKeyWithValue("card", "tok_card"))
		})
	})
})

// ---------------------------------------------------------------------------
// Additional branch-coverage tests
// ---------------------------------------------------------------------------

var _ = Describe("GetCredentialParams — alternate key names", func() {
	It("should accept clientID (capital D) as fallback", func() {
		credKeys := map[string]interface{}{
			"clientID": "client-id-value",
			"tokenUri": "https://token.example.com",
			"keyId":    "key-id-value",
		}
		clientId, tokenUri, keyId, err := GetCredentialParams(credKeys)
		Expect(err).To(BeNil())
		Expect(clientId).To(Equal("client-id-value"))
		Expect(tokenUri).To(Equal("https://token.example.com"))
		Expect(keyId).To(Equal("key-id-value"))
	})

	It("should accept tokenURI (capital URI) as fallback", func() {
		credKeys := map[string]interface{}{
			"clientId": "cid",
			"tokenURI": "https://token2.example.com",
			"keyId":    "kid",
		}
		_, tokenUri, _, err := GetCredentialParams(credKeys)
		Expect(err).To(BeNil())
		Expect(tokenUri).To(Equal("https://token2.example.com"))
	})

	It("should accept keyID (capital D) as fallback", func() {
		credKeys := map[string]interface{}{
			"clientId": "cid",
			"tokenUri": "https://token3.example.com",
			"keyID":    "key-id-capital",
		}
		_, _, keyId, err := GetCredentialParams(credKeys)
		Expect(err).To(BeNil())
		Expect(keyId).To(Equal("key-id-capital"))
	})
})

var _ = Describe("ParsePrivateKey — wrong PEM type", func() {
	It("should return error when PEM type is not PRIVATE KEY", func() {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		Expect(err).To(BeNil())
		pkcs1Bytes := x509.MarshalPKCS1PrivateKey(rsaKey)
		pemData := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkcs1Bytes})
		_, parseErr := ParsePrivateKey(string(pemData))
		Expect(parseErr).ToNot(BeNil())
		Expect(parseErr.GetMessage()).To(ContainSubstring(JWT_INVALID_FORMAT))
	})
})

var _ = Describe("GetPrivateKeyFromPem and GetSignedBearerUserToken", func() {
	var rsaPEM string

	BeforeEach(func() {
		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		Expect(err).To(BeNil())
		pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(rsaKey)
		Expect(err).To(BeNil())
		rsaPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes}))
	})

	It("GetPrivateKeyFromPem should parse a valid RSA PKCS8 PEM", func() {
		key, err := GetPrivateKeyFromPem(rsaPEM)
		Expect(err).To(BeNil())
		Expect(key).ToNot(BeNil())
	})

	It("GetPrivateKeyFromPem should return error for invalid PEM", func() {
		_, err := GetPrivateKeyFromPem("not-a-pem")
		Expect(err).ToNot(BeNil())
	})

	It("GetPrivateKeyFromPem should return error for wrong PEM type", func() {
		rsaKey, goErr := rsa.GenerateKey(rand.Reader, 2048)
		Expect(goErr).To(BeNil())
		pkcs1Bytes := x509.MarshalPKCS1PrivateKey(rsaKey)
		wrongTypePEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: pkcs1Bytes}))
		_, err := GetPrivateKeyFromPem(wrongTypePEM)
		Expect(err).ToNot(BeNil())
	})

	It("GetSignedBearerUserToken should succeed with valid RSA key", func() {
		key, err := GetPrivateKeyFromPem(rsaPEM)
		Expect(err).To(BeNil())
		token, skyErr := GetSignedBearerUserToken("client-id", "key-id", "https://token.example.com", key, common.BearerTokenOptions{})
		Expect(skyErr).To(BeNil())
		Expect(token).ToNot(BeEmpty())
	})

	It("GetSignedBearerUserToken should include ctx claim when ctx is set", func() {
		key, err := GetPrivateKeyFromPem(rsaPEM)
		Expect(err).To(BeNil())
		token, skyErr := GetSignedBearerUserToken("cid", "kid", "https://tok.example.com", key, common.BearerTokenOptions{Ctx: "myContext"})
		Expect(skyErr).To(BeNil())
		Expect(token).ToNot(BeEmpty())
	})

	It("GenerateSignedDataTokensHelper should succeed with valid RSA key", func() {
		key, err := GetPrivateKeyFromPem(rsaPEM)
		Expect(err).To(BeNil())
		options := common.SignedDataTokensOptions{
			DataTokens: []string{"tok1", "tok2"},
			TimeToLive: 60,
			Ctx:        "ctx-value",
		}
		resp, skyErr := GenerateSignedDataTokensHelper("client-id", "key-id", key, options, "https://token.example.com")
		Expect(skyErr).To(BeNil())
		Expect(resp).To(HaveLen(2))
		Expect(resp[0].Token).To(Equal("tok1"))
		Expect(resp[0].SignedToken).To(ContainSubstring("signed_token_"))
	})

	It("GenerateSignedDataTokensHelper should succeed when TimeToLive is 0 (uses default 60s)", func() {
		key, err := GetPrivateKeyFromPem(rsaPEM)
		Expect(err).To(BeNil())
		options := common.SignedDataTokensOptions{
			DataTokens: []string{"tok3"},
			TimeToLive: 0,
		}
		resp, skyErr := GenerateSignedDataTokensHelper("cid", "kid", key, options, "https://tok.example.com")
		Expect(skyErr).To(BeNil())
		Expect(resp).To(HaveLen(1))
	})
})

var _ = Describe("GetURLWithEnv — default branch", func() {
	It("should return a URL for an unknown Env using the default case", func() {
		url := GetURLWithEnv(common.Env(99), "mycluster")
		Expect(url).To(ContainSubstring("mycluster"))
	})
})

// ---------------------------------------------------------------------------
// Branch-coverage batch 2 — uncovered lines
// ---------------------------------------------------------------------------

var _ = Describe("GetFormattedGetRecord — skyflow_id remapping", func() {
	It("should remap skyflow_id key to SkyflowId in Fields", func() {
		record := vaultapis.V1FieldRecords{
			Fields: map[string]interface{}{
				"skyflow_id": "rec-001",
				"name":       "alice",
			},
		}
		result := GetFormattedGetRecord(record)
		Expect(result).To(HaveKey("SkyflowId"))
		Expect(result["SkyflowId"]).To(Equal("rec-001"))
		Expect(result).To(HaveKeyWithValue("name", "alice"))
	})
})

var _ = Describe("GetFormattedBatchInsertRecord — non-map element in records", func() {
	It("should skip non-map elements (continues) without error", func() {
		type fakeRecords struct {
			Records []interface{} `json:"records"`
		}
		type fakeOuter struct {
			Body fakeRecords `json:"Body"`
		}
		outer := fakeOuter{Body: fakeRecords{Records: []interface{}{42, "not-a-map"}}}
		result, err := GetFormattedBatchInsertRecord(outer, 0)
		Expect(err).To(BeNil())
		Expect(result).To(HaveKeyWithValue(internal.JSON_KEY_REQUEST_INDEX, 0))
	})
})

var _ = Describe("CreateInsertBulkBodyRequest — tokens branch", func() {
	It("should assign tokens to fields when Tokens slice is provided", func() {
		req := &common.InsertRequest{
			Table:  "test_table",
			Values: []map[string]interface{}{{"name": "alice"}, {"name": "bob"}},
		}
		opts := &common.InsertOptions{
			TokenMode: common.DISABLE,
			Tokens:    []map[string]interface{}{{"name": "tok_alice"}, {"name": "tok_bob"}},
		}
		body, skyErr := CreateInsertBulkBodyRequest(req, opts)
		Expect(skyErr).To(BeNil())
		Expect(body).ToNot(BeNil())
	})
})

var _ = Describe("CreateInsertBatchBodyRequest — tokens branch", func() {
	It("should assign tokens to batch records when Tokens slice is provided", func() {
		req := &common.InsertRequest{
			Table:  "test_table",
			Values: []map[string]interface{}{{"name": "alice"}, {"name": "bob"}},
		}
		opts := &common.InsertOptions{
			TokenMode: common.DISABLE,
			Tokens:    []map[string]interface{}{{"name": "tok_alice"}, {"name": "tok_bob"}},
		}
		body, err := CreateInsertBatchBodyRequest(req, opts)
		Expect(err).To(BeNil())
		Expect(body).ToNot(BeNil())
	})
})

var _ = Describe("GetCredentialParams — missing both keyId and keyID", func() {
	It("should return error when neither keyId nor keyID is present", func() {
		credKeys := map[string]interface{}{
			"clientId": "cid",
			"tokenUri": "https://token.example.com",
			// keyId / keyID absent deliberately
		}
		_, _, _, err := GetCredentialParams(credKeys)
		Expect(err).ToNot(BeNil())
	})
})

var _ = Describe("ParsePrivateKey — PKCS8 RSA key success path", func() {
	It("should parse a PKCS8 RSA PEM and return the private key (covers ok branch)", func() {
		rsaKey, goErr := rsa.GenerateKey(rand.Reader, 2048)
		Expect(goErr).To(BeNil())
		pkcs8Bytes, goErr := x509.MarshalPKCS8PrivateKey(rsaKey)
		Expect(goErr).To(BeNil())
		pemData := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes}))
		key, err := ParsePrivateKey(pemData)
		Expect(err).To(BeNil())
		Expect(key).ToNot(BeNil())
	})
})

var _ = Describe("GenerateSignedDataTokensHelper — invalid ctx type", func() {
	It("should return error when Ctx is an unsupported type", func() {
		rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(rsaKey)
		rsaPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes}))
		key, skyErr := GetPrivateKeyFromPem(rsaPEM)
		Expect(skyErr).To(BeNil())
		opts := common.SignedDataTokensOptions{
			DataTokens: []string{"tok1"},
			Ctx:        []int{1, 2, 3}, // unsupported type → ValidateAndResolveCtx error
		}
		_, err := GenerateSignedDataTokensHelper("cid", "kid", key, opts, "https://tok.example.com")
		Expect(err).ToNot(BeNil())
	})
})

var _ = Describe("GetSignedDataTokens — full credKeys path", func() {
	var rsaPEM string

	BeforeEach(func() {
		rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(rsaKey)
		rsaPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes}))
	})

	It("should succeed with a valid credKeys map including a PEM private key", func() {
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			"clientId":   "client-123",
			"tokenUri":   "https://token.example.com",
			"keyId":      "key-123",
		}
		opts := common.SignedDataTokensOptions{DataTokens: []string{"tok1", "tok2"}, TimeToLive: 60}
		resp, err := GetSignedDataTokens(credKeys, opts)
		Expect(err).To(BeNil())
		Expect(resp).To(HaveLen(2))
	})

	It("should return error from GetCredentialParams when clientId is missing after valid private key", func() {
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			// clientId intentionally missing
		}
		opts := common.SignedDataTokensOptions{DataTokens: []string{"tok1"}}
		_, err := GetSignedDataTokens(credKeys, opts)
		Expect(err).ToNot(BeNil())
	})
})

var _ = Describe("GetSignedBearerUserToken — ctx error path", func() {
	It("should return error when Ctx is an unsupported type", func() {
		rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(rsaKey)
		rsaPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes}))
		key, skyErr := GetPrivateKeyFromPem(rsaPEM)
		Expect(skyErr).To(BeNil())
		_, err := GetSignedBearerUserToken("cid", "kid", "https://tok.example.com", key, common.BearerTokenOptions{
			Ctx: []int{1, 2}, // unsupported type
		})
		Expect(err).ToNot(BeNil())
	})
})

var _ = Describe("GetPrivateKeyFromPem — additional error branches", func() {
	It("should return error when PKCS8 bytes are corrupt (ParsePKCS8 fails after PKCS1 fails)", func() {
		corruptPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("corrupt-bytes")}))
		_, err := GetPrivateKeyFromPem(corruptPEM)
		Expect(err).ToNot(BeNil())
	})

	It("should return error when PKCS8 key is an EC key (not *rsa.PrivateKey)", func() {
		ecKey, goErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		Expect(goErr).To(BeNil())
		pkcs8Bytes, goErr := x509.MarshalPKCS8PrivateKey(ecKey)
		Expect(goErr).To(BeNil())
		ecPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes}))
		_, err := GetPrivateKeyFromPem(ecPEM)
		Expect(err).ToNot(BeNil())
	})
})

var _ = Describe("GenerateBearerTokenHelper — all branches", func() {
	var rsaPEM string
	var savedHelper func(string) (string, *SkyflowError)

	BeforeEach(func() {
		rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(rsaKey)
		rsaPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes}))
		savedHelper = GetBaseURLHelper
	})

	AfterEach(func() {
		GetBaseURLHelper = savedHelper
	})

	It("should return error when privateKey is absent from credKeys", func() {
		credKeys := map[string]interface{}{"clientId": "cid", "tokenUri": "https://t.example.com", "keyId": "kid"}
		_, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{})
		Expect(err).ToNot(BeNil())
	})

	It("should return error when privateKey is an invalid PEM string", func() {
		credKeys := map[string]interface{}{
			"privateKey": "not-a-pem",
			"clientId":   "cid",
			"tokenUri":   "https://t.example.com",
			"keyId":      "kid",
		}
		_, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{})
		Expect(err).ToNot(BeNil())
	})

	It("should return error when clientId is missing from credKeys (GetCredentialParams fails)", func() {
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			"tokenUri":   "https://t.example.com",
			"keyId":      "kid",
		}
		_, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{})
		Expect(err).ToNot(BeNil())
	})

	It("should return error when Ctx is an invalid type (GetSignedBearerUserToken fails)", func() {
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			"clientId":   "cid",
			"tokenUri":   "https://t.example.com",
			"keyId":      "kid",
		}
		_, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{Ctx: []int{1}})
		Expect(err).ToNot(BeNil())
	})

	It("should return error when GetBaseURLHelper returns an error", func() {
		GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) {
			return "", NewSkyflowError(INVALID_INPUT_CODE, "mock base url error")
		}
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			"clientId":   "cid",
			"tokenUri":   "https://t.example.com",
			"keyId":      "kid",
		}
		_, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{})
		Expect(err).ToNot(BeNil())
	})

	It("should return access token when server responds with 200", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"accessToken":"mock-access-token","tokenType":"Bearer"}`)
		}))
		defer srv.Close()
		GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) { return srv.URL, nil }
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			"clientId":   "cid",
			"tokenUri":   "https://t.example.com",
			"keyId":      "kid",
		}
		resp, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{})
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())
	})

	It("should return error when server responds with 401", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":"unauthorized"}`)
		}))
		defer srv.Close()
		GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) { return srv.URL, nil }
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			"clientId":   "cid",
			"tokenUri":   "https://t.example.com",
			"keyId":      "kid",
		}
		_, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{})
		Expect(err).ToNot(BeNil())
	})

	It("should set scope when RoleIds is provided and server returns 200", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"accessToken":"scoped-token","tokenType":"Bearer"}`)
		}))
		defer srv.Close()
		GetBaseURLHelper = func(urlStr string) (string, *SkyflowError) { return srv.URL, nil }
		credKeys := map[string]interface{}{
			"privateKey": rsaPEM,
			"clientId":   "cid",
			"tokenUri":   "https://t.example.com",
			"keyId":      "kid",
		}
		resp, err := GenerateBearerTokenHelper(credKeys, common.BearerTokenOptions{RoleIds: []string{"role1", "role2"}})
		Expect(err).To(BeNil())
		Expect(resp).ToNot(BeNil())
	})
})
