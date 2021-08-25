package token

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"github.com/golang-jwt/jwt"
)

type ResponseToken struct {
	AccessToken string `json:"accessToken"`
	TokenType 	string `json:tokenType`
}

func GenerateToken(filePath string) (*ResponseToken, error) {
	var key map[string]interface{}

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to open credentials file")
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to read credentials file")
	}

	err = json.Unmarshal(byteValue, &key)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Provided json file is in wrong format")
	}

	token, err := getSATokenFromCredsFile(key)
	if err != nil {
		return nil, err
	}
	return token, nil;
}

// GetSATokenFromCredsFile gets bearer token from service account endpoint
func getSATokenFromCredsFile(
	key map[string]interface{}) (*ResponseToken, SkyflowError) {
	
	pvtKey, err := getPrivateKeyFromPem(key["privateKey"].(string))
	if err != nil {
		return nil, err
	}

	clientID, err := key["clientID"].(string)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to read clientID")
	}
	keyID, err := key["keyID"].(string)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to read keyID")
	}
	tokenURI, err := key["tokenURI"].(string)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to read tokenURI")
	}
	signedUserJWT, err := getSignedUserToken(
		clientID, keyID, tokenURI, pvtKey)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(map[string]string{
		"grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer",
		"assertion":  signedUserJWT,
	})
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to construct request payload")
	}
	payload := strings.NewReader(string(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("POST", tokenURI, payload)

	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to create new request with tokenURI and payload")
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Internal server error")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to read response payload")
	}

	var responseToken ResponseToken
	json.Unmarshal([]byte(body), &responseToken)
	
	return &responseToken, nil
}

func getPrivateKeyFromPem(pemKey string) (*rsa.PrivateKey, SkyflowError) {
	var err error
	privPem, err := pem.Decode([]byte(pemKey))

	if (err != nil) {
		return nil, NewSkyflowErrorWrap(http.StatusBadRequest, err, "Unable to decode the RSA private key")
	}

	if privPem.Type != "PRIVATE KEY" {
		return nil, NewSkyflowError(http.StatusBadRequest, fmt.Sprintf("RSA private key is of the wrong type Pem Type: %v", privPem.Type))
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPem.Bytes); err != nil {
			return nil, NewSkyflowError(http.StatusBadRequest, err, "unable to parse RSA private key")
		}
	}

	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, NewSkyflowError(http.StatusBadRequest, err, "unable to parse RSA private key, generating a temp one")
	}
	return privateKey, nil
}

func getSignedUserToken(
	clientID, keyID, tokenURI string,
	pvtKey *rsa.PrivateKey) (string, SkyflowError) {

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": clientID,
		"key": keyID,
		"aud": tokenURI,
		"sub": clientID,
		"exp": time.Now().Add(60 * time.Minute).Unix(),
	})

	var err error
	signedToken, err := token.SignedString(pvtKey)
	if err != nil {
		return "", NewSkyflowError(http.StatusBadRequest, err, "unable to parse jwt payload,")
	}
	return signedToken, nil
}
