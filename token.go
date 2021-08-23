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

func GetToken(filePath string) (*ResponseToken, error) {
	var key map[string]interface{}

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &key)
	if err != nil {
		return nil, err
	}

	token, err := getSATokenFromCredsFile(key)
	if err != nil {
		return nil, err
	}
	return token, nil;
}

// GetSATokenFromCredsFile gets bearer token from service account endpoint
func getSATokenFromCredsFile(
	key map[string]interface{}) (*ResponseToken, error) {
	
	pvtKey, err := getPrivateKeyFromPem(key["privateKey"].(string))
	if err != nil {
		return nil, err
	}

	clientID, _ := key["clientID"].(string)
	keyID, _ := key["keyID"].(string)
	tokenURI, _ := key["tokenURI"].(string)
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
		return nil, err
	}
	payload := strings.NewReader(string(reqBody))
	client := &http.Client{}
	req, err := http.NewRequest("POST", tokenURI, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var responseToken ResponseToken
	json.Unmarshal([]byte(body), &responseToken)
	
	return &responseToken, nil
}

func getPrivateKeyFromPem(pemKey string) (*rsa.PrivateKey, error) {
	var err error
	privPem, _ := pem.Decode([]byte(pemKey))

	if privPem.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("RSA private key is of the wrong type Pem Type: %v", privPem.Type)
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPem.Bytes); err != nil {
			return nil,
				fmt.Errorf("unable to parse RSA private key. ERR: %v", err)
		}
	}

	var privateKey *rsa.PrivateKey
	var ok bool
	privateKey, ok = parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil,
			fmt.Errorf("unable to parse RSA private key, generating a temp one, ERR: %v", err)
	}
	return privateKey, nil
}

func getSignedUserToken(
	clientID, keyID, tokenURI string,
	pvtKey *rsa.PrivateKey) (string, error) {

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
		return "", fmt.Errorf("unable to parse jwt payload, err : %v", err)
	}
	return signedToken, nil
}
