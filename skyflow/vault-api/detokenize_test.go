package vaultapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/commonutils/mocks"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func TestNoRecordsForDetokenize(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	detokenizeApi := DetokenizeApi{Configuration: configuration, Records: records, Token: ""}
	_, err := detokenizeApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, detokenizeTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyRecordsForDetokenize(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record []interface{}
	records["records"] = record
	detokenizeApi := DetokenizeApi{Configuration: configuration, Records: records, Token: ""}
	_, err := detokenizeApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, detokenizeTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoTokenForDetokenize(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	detokenizeApi := DetokenizeApi{Configuration: configuration, Records: records, Token: ""}
	_, err := detokenizeApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TOKEN, detokenizeTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyEmptyTokenForDetokenize(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["token"] = ""
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	detokenizeApi := DetokenizeApi{Configuration: configuration, Records: records, Token: ""}
	_, err := detokenizeApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKEN_ID, detokenizeTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestForDetokenize(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["token"] = "1234"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	detokenizeApi := DetokenizeApi{Configuration: configuration, Records: records, Token: ""}
	resJson := `{
		"records": [
			{
				"token": "1234",
				"valueType": "STRING",
				"value": "rach"
			}
		]
	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(resJson)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	detokenizeApi.Get()
}
