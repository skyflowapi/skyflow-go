package util

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"

	sErrors "github.com/skyflowapi/skyflow-go/commonutils/errors"
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

func setUpGenerateBearerTokenTests() []generateBearerTokenTest {

	validFilePath := os.Getenv("CREDENTIALS_FILE_PATH")
	generateBearerTokenTests := []generateBearerTokenTest{
		{validFilePath, "Bearer"},
	}
	return generateBearerTokenTests
}

func TestGenerateBearerToken(t *testing.T) {

	generateBearerTokenFromCredsTests := setUpGenerateBearerTokenTests()

	for _, test := range generateBearerTokenFromCredsTests {
		if resp, err := GenerateBearerToken(test.filePath); !strings.Contains(resp.TokenType, test.expected) {
			if err != nil {
				t.Errorf("Output %s not equal to expected %s", err.GetMessage(), test.expected)
			} else {
				t.Errorf("Output %s not equal to expected %s", resp.TokenType, test.expected)
			}
		}
	}
}

type generateBearerTokenFromCredsTest struct {
	creds    string
	expected string
}

func setUpGenerateBearerTokenFromCredsTests() []generateBearerTokenFromCredsTest {

	// pvtKey := os.Getenv("VALID_CREDS_PVT_KEY")

	// fmt.Println(pvtKey)
	// invalidClientIdCreds := fmt.Sprintf("{\"privateKey\" : \"%s\"}", pvtKey)
	// invalidKeyIdCreds := fmt.Sprintf("{\"privateKey\" : \"%s\", \"clientID\": \"cId\"}", pvtKey)
	// invalidtokenURICreds := fmt.Sprintf("{\"privateKey\" : \"%s\", \"clientID\": \"cId\", \"keyID\": \"kId\"}", pvtKey)
	generateBearerTokenFromCredsTests := []generateBearerTokenFromCredsTest{
		{"", "credentials string is not a valid json string format"},
		{"{}", "Unable to read privateKey"},
		// {invalidClientIdCreds, "Unable to read clientID"},
		// {invalidKeyIdCreds, "Unable to read keyID"},
		// {invalidtokenURICreds, "Unable to read tokenURI"},
	}

	return generateBearerTokenFromCredsTests
}

func TestGenerateBearerTokenFromCreds(t *testing.T) {

	generateBearerTokenFromCredsTests := setUpGenerateBearerTokenFromCredsTests()

	for _, test := range generateBearerTokenFromCredsTests {
		if _, err := GenerateBearerTokenFromCreds(test.creds); !strings.Contains(err.GetMessage(), test.expected) {
			if err != nil {
				t.Errorf("Output %s not equal to expected %s", err.GetMessage(), test.expected)
			}
		}
	}
}

type isValidTokenTest struct {
	token    string
	expected bool
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

func TestIsValid(t *testing.T) {

	isValidTokenTests := setUpIsValidTests()

	for _, test := range isValidTokenTests {
		if output := IsValid(test.token); output != test.expected {
			t.Errorf("Output %t not equal to expected %t", output, test.expected)
		}
	}
}
