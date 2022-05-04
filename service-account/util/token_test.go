package util

import (
	"bytes"
	"errors"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"

	sErrors "github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/commonutils/mocks"
)

func init() {
	err := godotenv.Load("../../.env")

	if err != nil {
		fmt.Println("Error loading .env file")
	}

}

func TestGetToken(t *testing.T) {
	_, err := GenerateToken("")
	var apiErr *sErrors.SkyflowError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expect error to be Skyflow error, was not, %v", err)
	}
}

type generateBearerTokenTest struct {
	filePath string
	expected string
}

type generateBearerTokenFromCredsTest struct {
	creds    string
	expected string
}

type isValidTokenTest struct {
	token    string
	expected bool
}

func setUpGenerateBearerTokenTests() []generateBearerTokenTest {

	invalidJsonFilePath := "../../test/invalidJson.json"
	validFilePath := os.Getenv("CREDENTIALS_FILE_PATH")

	generateBearerTokenTests := []generateBearerTokenTest{
		{invalidJsonFilePath, "Provided json file is in wrong format"},
		{validFilePath, "Bearer"},
	}
	return generateBearerTokenTests
}

func setUpGenerateBearerTokenFromCredsTests() []generateBearerTokenFromCredsTest {

	pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")
	invalidKeyType := strings.Replace(pvtKey, "PRIVATE KEY", "PUBLIC KEY", 2)

	invalidPvtKeyCreds := fmt.Sprintf("{\"privateKey\" : \"%s\"}", "invalidKey")
	invalidKeyTypeCreds := fmt.Sprintf("{\"privateKey\" : \"%s\"}", invalidKeyType)
	invalidClientIdCreds := fmt.Sprintf("{\"privateKey\" : \"%s\"}", pvtKey)
	invalidKeyIdCreds := fmt.Sprintf("{\"privateKey\" : \"%s\", \"clientID\": \"cId\"}", pvtKey)
	invalidtokenURICreds := fmt.Sprintf("{\"privateKey\" : \"%s\", \"clientID\": \"cId\", \"keyID\": \"kId\"}", pvtKey)
	invalidCreds := fmt.Sprintf("{\"privateKey\" : \"%s\", \"clientID\": \"cId\", \"keyID\": \"kId\", \"tokenURI\": \"tokenURI\" }", pvtKey)
	invalidCreds2 := fmt.Sprintf("{\"privateKey\" : \"%s\", \"clientID\": \"cId\", \"keyID\": \"kId\", \"tokenURI\": \"https://manage.skyflowapis.com/v1/auth/sa/oauth/token\" }", pvtKey)
	generateBearerTokenFromCredsTests := []generateBearerTokenFromCredsTest{
		{"", "credentials string is not a valid json string format"},
		{"{}", "Unable to read privateKey"},
		{invalidPvtKeyCreds, "Unable to decode the RSA private PEM"},
		{invalidKeyTypeCreds, "RSA private key is of the wrong type Pem Type"},
		{invalidClientIdCreds, "Unable to read clientID"},
		{invalidKeyIdCreds, "Unable to read keyID"},
		{invalidtokenURICreds, "Unable to read tokenURI"},
		{invalidCreds, "Internal server error"},
		{invalidCreds2, "Error Occured"},
	}

	return generateBearerTokenFromCredsTests
}

func setUpIsValidTests() []isValidTokenTest {

	expiredToken := os.Getenv("EXPIRED_TOKEN")

	isValidTokenTests := []isValidTokenTest{
		{"", false},
		{"invalidToken", false},
		{expiredToken, false},
	}

	return isValidTokenTests
}

func TestGenerateBearerToken(t *testing.T) {

	mockApi()
	generateBearerTokenFromCredsTests := setUpGenerateBearerTokenTests()

	for _, test := range generateBearerTokenFromCredsTests {
		resp, err := GenerateBearerToken(test.filePath)
		check(resp, err, test.expected, t)
	}
}

func TestGenerateBearerTokenFromCreds(t *testing.T) {

	mockApi()
	generateBearerTokenFromCredsTests := setUpGenerateBearerTokenFromCredsTests()

	for _, test := range generateBearerTokenFromCredsTests {

		resp, err := GenerateBearerTokenFromCreds(test.creds)
		check(resp, err, test.expected, t)
	}
}

func TestIsValid(t *testing.T) {

	isValidTokenTests := setUpIsValidTests()

	for _, test := range isValidTokenTests {
		if output := IsValid(test.token); output != test.expected {
			t.Errorf("Output %t not equal to expected %t", output, test.expected)
		}
	}
}

func check(resp *ResponseToken, err *sErrors.SkyflowError, expected string, t *testing.T) {
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
func TestAppendRequestId(t *testing.T) {
	var message = appendRequestId("message", "1234")
	checkErrorMessage(message, "message - requestId : 1234", t)
}
func TestAppendRequestIdWithEmpty(t *testing.T) {
	var message = appendRequestId("message", "")
	checkErrorMessage(message, "message", t)
}

func checkErrorMessage(got string, wanted string, t *testing.T) {
	if got != wanted {
		t.Errorf("got  %s, wanted %s", got, wanted)
	}
}

func mockApi() {
	resJson := `{
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "400",
		"AccessToken":"token",
		"TokenType":"string"

	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(resJson)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
}
