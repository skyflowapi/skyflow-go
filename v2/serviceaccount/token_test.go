package serviceaccount

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	internalAuthApi "skyflow-go/internal/generated/auth"
	skyflowError "skyflow-go/utils/error"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Sample valid RSA private key in PEM format for testing
const validPrivateKeyPEM = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASC...
-----END PRIVATE KEY-----
`

// Sample invalid RSA private key in PEM format
const invalidPrivateKeyPEM = `
-----BEGIN INVALID KEY-----
MIIBOgIBAAJBAK...
-----END INVALID KEY-----
`

// Invalid PEM format (unable to decode)
const undecodablePEM = `
-----BEGIN PRIVATE KEY-----
invalid-content
-----END PRIVATE KEY-----
`

func TestGetCredentialParams(t *testing.T) {
	var credMap = map[string]interface{}{}
	credMap["clientSecret"] = "clientSecret"
	credMap["keyID"] = "keyID"
	_, _, _, err := getCredentialParams(credMap)
	if err != nil {
		if !(strings.Contains(err.GetMessage(), "Invalid credential parameters")) {
			t.Errorf("Output %s not equal to expected %s", err.GetMessage(), "Invalid credential parameters")
		}
	}
}
func TestGetPrivateKeySuccess(t *testing.T) {
	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	var credMap = map[string]interface{}{}
	err := json.Unmarshal([]byte(pvtKey), &credMap)
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("env is missing")
	}
	res, errr := getPrivateKey(credMap)
	assert.Nil(t, errr)
	assert.NotNil(t, res)
}
func TestGetPrivateKeyFail(t *testing.T) {
	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	var credMap = map[string]interface{}{}
	err := json.Unmarshal([]byte(pvtKey), &credMap)
	assert.Nil(t, err)
	delete(credMap, "privateKey")
	res, errr := getPrivateKey(credMap)
	assert.Equal(t, errr.GetMessage(), "Message: Missing private key")
	assert.NotNil(t, errr)
	assert.Nil(t, res)
}
func TestGetPrivateKeyInvalid(t *testing.T) {
	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	var credMap = map[string]interface{}{}
	err := json.Unmarshal([]byte(pvtKey), &credMap)
	assert.Nil(t, err)
	credMap["privateKey"] = "invalid"
	res, errr := getPrivateKey(credMap)
	assert.Equal(t, errr.GetMessage(), "Message: Invalid private key format")
	assert.Nil(t, res)
}

func TestGetPrivateKeyInvalidKeyType(t *testing.T) {
	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	invalidKeyType := strings.Replace(pvtKey, "PRIVATE KEY", "PUBLIC KEY", 2)
	var credMap = map[string]interface{}{}
	err := json.Unmarshal([]byte(invalidKeyType), &credMap)
	assert.Nil(t, err)
	res, errr := getPrivateKey(credMap)
	assert.Equal(t, errr.GetMessage(), "Message: Invalid private key type")
	assert.Nil(t, res)
}

// Test for successfully decoding a valid PEM key
func TestGetPrivateKeyFromValidPEM(t *testing.T) {
	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	var credMap = map[string]interface{}{}
	err := json.Unmarshal([]byte(pvtKey), &credMap)
	assert.Nil(t, err)
	key, err := getPrivateKeyFromPem(credMap["privateKey"].(string))

	assert.NotNil(t, key, "Expected a valid private key")
	assert.IsType(t, &rsa.PrivateKey{}, key)
	assert.Nil(t, err, "Expected no error for a valid private key")
}

// Test for PEM block being nil
func TestGetPrivateKeyFromNilPEM(t *testing.T) {
	key, err := getPrivateKeyFromPem("")

	assert.Nil(t, key, "Expected nil key for undecodable PEM")
	assert.NotNil(t, err, "Expected error for nil PEM block")
	assert.Equal(t, "Code: 400", err.GetCode())
	assert.Equal(t, "Message: Unable to decode the RSA private PEM", err.GetMessage())
}

// Test for incorrect PEM type
func TestGetPrivateKeyFromInvalidTypePEM(t *testing.T) {
	key, err := getPrivateKeyFromPem(invalidPrivateKeyPEM)

	assert.Nil(t, key, "Expected nil key for invalid PEM type")
	assert.NotNil(t, err, "Expected error for invalid PEM type")
	assert.Equal(t, "Code: 400", err.GetCode())
	assert.Contains(t, err.GetMessage(), "Message: Unable to decode the RSA private PEM")
}

// Test for undecodable PEM format
func TestGetPrivateKeyFromUndecodablePEM(t *testing.T) {
	key, err := getPrivateKeyFromPem(undecodablePEM)

	assert.Nil(t, key, "Expected nil key for undecodable PEM")
	assert.NotNil(t, err, "Expected error for undecodable PEM format")
	assert.Equal(t, "Code: 400", err.GetCode())
	assert.Contains(t, err.GetMessage(), "Unable to decode the RSA private PEM")
}

// Test for PEM with invalid private key data
func TestGetPrivateKeyFromInvalidPrivateKey(t *testing.T) {
	key, err := getPrivateKeyFromPem(invalidPrivateKeyPEM)

	assert.Nil(t, key, "Expected nil key for invalid private key data")
	assert.NotNil(t, err, "Expected error for invalid private key data")
	assert.Equal(t, "Code: 400", err.GetCode())
	assert.Equal(t, "Message: Unable to decode the RSA private PEM", err.GetMessage())
}

// Test with a valid URL, expecting to return the correct base URL
func TestGetBaseURL_ValidURL(t *testing.T) {
	urlStr := "https://example.com/path/to/resource"
	expectedBaseURL := "https://example.com"

	baseURL, err := getBaseURL(urlStr)

	assert.Nil(t, err, "Expected no error for valid URL")
	assert.Equal(t, expectedBaseURL, baseURL, "Expected base URL to match")
}

// Test with an invalid URL, expecting an error
func TestGetBaseURL_InvalidURL(t *testing.T) {
	urlStr := "htp:/invalid-url"

	baseURL, err := getBaseURL(urlStr)
	assert.NotNil(t, err.GetMessage(), "Expected error for invalid URL")
	assert.Empty(t, baseURL, "Expected empty base URL for invalid URL")
}

// // Test with a URL missing scheme, expecting to return the host
func TestGetBaseURL_MissingScheme(t *testing.T) {
	urlStr := "example.com/path/to/resource"

	baseURL, err := getBaseURL(urlStr)

	assert.NotNil(t, err, "Expected error for URL without scheme")
	assert.Empty(t, baseURL, "Expected empty base URL when URL has no scheme")
}

// Test with URL containing port number, expecting the port to be included in base URL

func TestGetBaseURL_WithPort(t *testing.T) {
	urlStr := "https://example.com:8080/path/to/resource"
	expectedBaseURL := "https://example.com:8080"

	baseURL, err := getBaseURL(urlStr)

	assert.Nil(t, err, "Expected no error for valid URL with port")
	assert.Equal(t, expectedBaseURL, baseURL, "Expected base URL to include port")
}

// Test with URL having query parameters, expecting them to be ignored in base URL

func TestGetBaseURL_WithQueryParams(t *testing.T) {
	urlStr := "https://example.com/path?query=1"
	expectedBaseURL := "https://example.com"

	baseURL, err := getBaseURL(urlStr)

	assert.Nil(t, err, "Expected no error for URL with query params")
	assert.Equal(t, expectedBaseURL, baseURL, "Expected base URL to ignore query parameters")
}

// Test with URL having fragment, expecting it to be ignored in base URL

func TestGetBaseURL_WithFragment(t *testing.T) {
	urlStr := "https://example.com/path#fragment"
	expectedBaseURL := "https://example.com"

	baseURL, err := getBaseURL(urlStr)

	assert.Nil(t, err, "Expected no error for URL with fragment")
	assert.Equal(t, expectedBaseURL, baseURL, "Expected base URL to ignore fragment")
}

// Test with an empty string as URL, expecting an error
func TestGetBaseURL_EmptyURL(t *testing.T) {
	urlStr := ""

	baseURL, err := getBaseURL(urlStr)

	assert.NotNil(t, err, "Expected error for empty URL string")
	assert.Empty(t, baseURL, "Expected empty base URL for empty URL string")
}
func TestParseCredentialsFile(t *testing.T) {
	file, err := parseCredentialsFile("")
	if err != nil {
		assert.NotNil(t, err, "Expected a valid file for credentials file")
	}
	assert.Equal(t, len(file), 0)
}
func TestParseCredentialsFileInvalidPath(t *testing.T) {
	file, err := parseCredentialsFile("invalid.json")
	if err != nil {
		assert.NotNil(t, err, "Expected a valid file for credentials file")
		assert.Equal(t, err.GetMessage(), "Message: Failed to open credentials file")
	}
	assert.Equal(t, len(file), 0)
}
func TestParseCredentialsFileInvalidJSON(t *testing.T) {
	file, err := parseCredentialsFile("/Users/bharts/Desktop/skyflow-go/invalid.json")
	if err != nil {
		assert.NotNil(t, err, "Expected a valid file for credentials file")
	}
	assert.Equal(t, len(file), 0)
}

func TestParseCredentialsFileValidJSON(t *testing.T) {
	file, err := parseCredentialsFile("/Users/bharts/Desktop/skyflow-go/validcred.json")

	assert.Nil(t, err, "Expected error to be nil")
	assert.Equal(t, len(file), 8)
}

// bearer token
func TestGenerateBearerTokenInvalidPath(t *testing.T) {
	bearerToken, err := GenerateBearerToken("", BearerTokenOptions{})
	assert.NotNil(t, err, "Expected error for empty bearer token")
	assert.Nil(t, bearerToken, "Expected bearer token for empty bearer token")
}

func TestGenerateBearerTokenInvalidJSON(t *testing.T) {
	bearerToken, err := GenerateBearerToken("/Users/bharts/Desktop/skyflow-go/invalid.json", BearerTokenOptions{})
	assert.NotNil(t, err, "Expected error for empty bearer token")
	assert.Equal(t, err.GetMessage(), "Message: Failed to parse credential file")
	assert.Nil(t, bearerToken, "Expected bearer token for empty bearer token")
}

func TestGenerateBearerToken(t *testing.T) {
	pvtKey := os.Getenv("VALID_PVT_KEY")
	bearerToken, err := GenerateBearerTokenFromCreds(pvtKey, BearerTokenOptions{})
	assert.Contains(t, err.GetMessage(), "Error occurred")
	assert.NotNil(t, err, "Expected error for empty")
	assert.Nil(t, bearerToken, "Expected bearer token for empty bearer token")
}
func TestGenerateBearerTokenInvalidString(t *testing.T) {
	bearerToken, err := GenerateBearerTokenFromCreds("empty", BearerTokenOptions{})
	assert.Contains(t, err.GetMessage(), "Message: Failed to parse credential")
	assert.NotNil(t, err, "Expected error for empty")
	assert.Nil(t, bearerToken, "Expected bearer token for empty bearer token")
}
func TestGenerateSignedTokenFromCredsInvalidString(t *testing.T) {
	bearerToken, err := GenerateSignedDataTokensFromCreds("[", SignedDataTokensOptions{})
	assert.Contains(t, err.GetMessage(), "Message: Failed to parse credential")
	assert.NotNil(t, err, "Expected error for empty")
	assert.Nil(t, bearerToken, "Expected bearer token for empty bearer token")
}
func TestGenerateSignedTokenInvalidJSON(t *testing.T) {
	dataTokens, err := GenerateSignedDataTokens("/Users/bharts/Desktop/skyflow-go/invalid.json", SignedDataTokensOptions{DataTokens: make([]string, 2)})
	assert.NotNil(t, err, "Expected error for empty bearer token")
	assert.Equal(t, "Message: Failed to parse credential file", err.GetMessage())
	assert.Nil(t, dataTokens, "Expected bearer token for empty bearer token")
}
func TestGenerateSignedTokenInvalidPath(t *testing.T) {
	bearerToken, err := GenerateSignedDataTokens("", SignedDataTokensOptions{})
	assert.NotNil(t, err, "Expected error for empty bearer token")
	assert.Empty(t, bearerToken, "Expected bearer token for empty bearer token")
}
func TestGenerateSignedTokenEmptyDataTokens(t *testing.T) {
	bearerToken, err := GenerateSignedDataTokens("/Users/bharts/Desktop/skyflow-go/validcred.json", SignedDataTokensOptions{})
	assert.NotNil(t, err, "Expected error for empty bearer token")
	assert.Nil(t, bearerToken, "Expected bearer token for empty bearer token")
}
func TestGenerateSignedTokenDataTokensSuccess(t *testing.T) {
	dataTokens := []string{"token1", "token2"}
	bearerToken, err := GenerateSignedDataTokens("/Users/bharts/Desktop/skyflow-go/validcred.json", SignedDataTokensOptions{DataTokens: dataTokens, TimeToLive: 2, Ctx: "ctx"})
	assert.Nil(t, err, "Expected error as nil")
	assert.NotNil(t, bearerToken, "Expected bearer token")
	assert.Equal(t, len(bearerToken), 2, "Expected 2 bearer tokens")
	assert.Equal(t, bearerToken[0].Token, "token1", "Expected bearer token 1")
	assert.Equal(t, bearerToken[1].Token, "token2", "Expected bearer token 2")
	assert.Contains(t, bearerToken[0].SignedToken, "signed_token_")
	assert.Contains(t, bearerToken[1].SignedToken, "signed_token_")
}
func TestGenerateSignedTokenDataTokensSuccessWithCredString(t *testing.T) {
	dataTokens := []string{"token1", "token2"}
	pvtKey := os.Getenv("VALID_PVT_KEY")
	bearerToken, err := GenerateSignedDataTokensFromCreds(pvtKey, SignedDataTokensOptions{DataTokens: dataTokens})
	assert.Nil(t, err, "Expected error as nil")
	assert.NotNil(t, bearerToken, "Expected bearer token")
	assert.Equal(t, len(bearerToken), 2, "Expected 2 bearer tokens")
	assert.Equal(t, bearerToken[0].Token, "token1", "Expected bearer token 1")
	assert.Equal(t, bearerToken[1].Token, "token2", "Expected bearer token 2")
	assert.Contains(t, bearerToken[0].SignedToken, "signed_token_")
	assert.Contains(t, bearerToken[1].SignedToken, "signed_token_")
}
func TestIsExpired(t *testing.T) {
	token := os.Getenv("EXPIRED_TOKEN")
	isExpired := IsExpired(token)
	assert.True(t, isExpired)

}

// Mock TokenResponse struct from the auth package
type MockAuthTokenResponse struct {
	AccessToken string
	TokenType   string
}

func (m *MockAuthTokenResponse) GetAccessToken() string {
	return m.AccessToken
}

func (m *MockAuthTokenResponse) GetTokenType() string {
	return m.TokenType
}

// Mock function for generateBearerToken
func mockGenerateBearerToken(credKeys map[string]interface{}, options BearerTokenOptions) (*internalAuthApi.V1GetAuthTokenResponse, *skyflowError.SkyflowError) {
	var token = "mockAccessToken"
	var tokenType = "Bearer"
	return &internalAuthApi.V1GetAuthTokenResponse{
		AccessToken: &token,
		TokenType:   &tokenType,
	}, nil
}

func TestGenerateBearerToken_Success(t *testing.T) {
	generateBearerTokenFunc = mockGenerateBearerToken
	// Call GenerateBearerToken
	tokenResponse, _ := GenerateBearerToken("../validcred.json", BearerTokenOptions{})

	// Assertions
	//require.NoError(t, err)
	assert.NotNil(t, tokenResponse)
	assert.Equal(t, "mockAccessToken", tokenResponse.AccessToken)
	assert.Equal(t, "Bearer", tokenResponse.TokenType)
}
func TestGenerateBearerToken_SuccessWithString(t *testing.T) {
	content := os.Getenv("VALID_PVT_KEY")
	generateBearerTokenFunc = mockGenerateBearerToken
	tokenResponse, err := GenerateBearerTokenFromCreds(string(content), BearerTokenOptions{})
	assert.Nil(t, err, "Expected error as nil")
	assert.NotNil(t, tokenResponse)
	assert.Equal(t, "mockAccessToken", tokenResponse.AccessToken)
	assert.Equal(t, "Bearer", tokenResponse.TokenType)
}
func TestGetSignedBearerUserTokenValidCase(t *testing.T) {
	content := os.Getenv("VALID_PVT_KEY")
	var credKeys map[string]interface{}
	err := json.Unmarshal([]byte(content), &credKeys)
	if err != nil {
		assert.Fail(t, "key is empty")
	}
	key, _ := getPrivateKeyFromPem(credKeys["privateKey"].(string))
	priv, err := getSignedBearerUserToken(credKeys["clientID"].(string), credKeys["keyID"].(string), credKeys["tokenURI"].(string), key, BearerTokenOptions{})
	assert.Nil(t, err, "Expected error as nil")
	assert.NotNil(t, priv)
}
func TestGetSignedBearerUserTokenInvalid(t *testing.T) {
	content := os.Getenv("VALID_PVT_KEY")
	var credKeys map[string]interface{}
	err := json.Unmarshal([]byte(content), &credKeys)
	if err != nil {
		assert.Fail(t, "key is empty")
	}
	var privateKey *rsa.PrivateKey
	privateKey, _ = rsa.GenerateKey(rand.Reader, 242)
	priv, err := getSignedBearerUserToken(credKeys["clientID"].(string), credKeys["keyID"].(string), credKeys["tokenURI"].(string), privateKey, BearerTokenOptions{Ctx: "ctx"})
	assert.NotNil(t, err, "Expected error as nil")
	assert.Equal(t, priv, "")
}
func getValidKeys() map[string]interface{} {
	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	var credMap = map[string]interface{}{}
	_ = json.Unmarshal([]byte(pvtKey), &credMap)
	return credMap
}
func TestGenerateBearerTokenCheckForKey(t *testing.T) {
	// Helper function to run the common assertions for error message and response
	runTest := func(credMap map[string]interface{}, options BearerTokenOptions, expectedErrorMessage string, t *testing.T) {
		res, err := generateBearerToken(credMap, options)
		assert.Equal(t, err.GetMessage(), expectedErrorMessage)
		assert.NotNil(t, err)
		assert.Nil(t, res)
	}

	// Test case for missing or invalid privateKey
	credMap := getValidKeys()
	credMap["privateKey"] = ""
	runTest(credMap, BearerTokenOptions{}, "Message: Unable to decode the RSA private PEM", t)
	credMap["privateKey"] = nil
	runTest(credMap, BearerTokenOptions{}, "Message: privateKey is nil", t)

	// Test case for missing tokenURI
	credMap = getValidKeys()
	credMap["tokenURI"] = nil
	runTest(credMap, BearerTokenOptions{}, "Message: tokenURI is nil", t)

	// Test case for missing clientID
	credMap = getValidKeys()
	credMap["clientID"] = nil
	runTest(credMap, BearerTokenOptions{}, "Message: clientID is nil", t)

	// Test case for missing keyID
	credMap = getValidKeys()
	credMap["keyID"] = nil
	runTest(credMap, BearerTokenOptions{}, "Message: keyID is nil", t)

	// Test case for invalid tokenURI
	credMap = getValidKeys()
	credMap["tokenURI"] = "demo.com"
	runTest(credMap, BearerTokenOptions{}, "Message: Failed to get token URL", t)

	// Test case for missing role IDs
	credMap = getValidKeys()
	roleIDs := []string{"id1", "id2"}
	credMap["roleIDs"] = roleIDs
	runTest(credMap, BearerTokenOptions{RoleIDs: roleIDs}, "Message: Error occurred", t)

	// Test case for missing tokenURI in getSignedDataTokens
	credMap = getValidKeys()
	delete(credMap, "tokenURI")
	res1, errr1 := getSignedDataTokens(credMap, SignedDataTokensOptions{})
	assert.Equal(t, errr1.GetMessage(), "Message: Invalid credential parameters")
	assert.NotNil(t, errr1)
	assert.Nil(t, res1)

	// Test case for missing privateKey in getSignedDataTokens
	credMap = getValidKeys()
	delete(credMap, "privateKey")
	res1, errr1 = getSignedDataTokens(credMap, SignedDataTokensOptions{})
	assert.Equal(t, errr1.GetMessage(), "Message: Missing private key")
	assert.NotNil(t, errr1)
	assert.Nil(t, res1)
}

func check(resp *TokenResponse, err *skyflowError.SkyflowError, expected string, t *testing.T) {
	if resp != nil {
		if resp.TokenType != expected {
			t.Errorf("Output %s not equal to expected %s", resp.TokenType, expected)
		}
	} else if err != nil {
		if !strings.Contains(err.GetMessage(), expected) {
			t.Errorf("Output %s not equal to expected %s", err.GetMessage(), expected)
		}
	}
}
