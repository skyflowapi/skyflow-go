package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	errors1 "github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	saUtil "github.com/skyflowapi/skyflow-go/service-account/util"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func init() {
	err := godotenv.Load("../../.env")

	if err != nil {
		fmt.Println("Error loading .env file")
	}

}

func invalidToken() (string, error) {
	validFilePath := os.Getenv("CREDENTIALS_FILE_PATH")

	token, err := saUtil.GenerateBearerToken(validFilePath)
	if err != nil {
		return "", err
	} else {
		return token.AccessToken, nil
	}
}
func validToken() (string, error) {

	return "token", nil
}
func TestInsertInvalidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: invalidToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.Insert(record)
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.INVALID_BEARER_TOKEN, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)

}
func TestDetokenizeInvalidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: invalidToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.Detokenize(record)
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.INVALID_BEARER_TOKEN, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestGetByIdInvalidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: invalidToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.GetById(record)
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.INVALID_BEARER_TOKEN, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvokeConnectionInvalidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: invalidToken}
	var client = Init(configuration)
	var record = common.ConnectionConfig{}
	_, err := client.InvokeConnection(record)
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.INVALID_BEARER_TOKEN, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInsertValidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.Insert(record, common.InsertOptions{Tokens: true})
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestDetokenizeValidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.Detokenize(record)
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestGetByIdValidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.GetById(record)
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvokeConnectionValidToken(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = common.ConnectionConfig{}
	_, err := client.InvokeConnection(record)
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.EMPTY_CONNECTION_URL, "InvokeConnection"))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func check(got string, wanted string, t *testing.T) {
	if got != wanted {
		t.Errorf("got  %s, wanted %s", got, wanted)
	}
}
