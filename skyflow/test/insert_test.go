package test

import (
	"fmt"
	"testing"

	"github.com/skyflowapi/skyflow-go/errors"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
	vaultapi "github.com/skyflowapi/skyflow-go/skyflow/vault-api"
)

func GetToken() (string, error) {
	return "", nil
}
func TestEmptyVaultId(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	insertApi := vaultapi.InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_VAULT_ID)
	check(err.GetMessage(), skyflowError.GetMessage(), t)

}

func TestEmptyVaultUrl(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	insertApi := vaultapi.InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_VAULT_URL)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvalidVaultUrl(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "url", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	insertApi := vaultapi.InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(errors.INVALID_VAULT_URL, configuration.VaultURL))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func check(got string, wanted string, t *testing.T) {
	if got != wanted {
		t.Errorf("got  %s, wanted %s", got, wanted)
	}
}
