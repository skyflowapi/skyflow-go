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

func init() {
	Client = &mocks.MockClient{}
}

func GetToken() (string, error) {
	return "", nil
}
func TestEmptyVaultId(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)

}

func TestEmptyVaultUrl(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_URL, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvalidVaultUrl(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "url", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_VAULT_URL, clientTag, configuration.VaultURL))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvalidVaultUrl1(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "http://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_VAULT_URL, clientTag, configuration.VaultURL))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestNoRecords(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyRecords(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record []interface{}
	records["records"] = record
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestMissingTable(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record map[string]interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TABLE, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTable(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	record["table"] = ""
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_NAME, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestMissingFields(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	record["table"] = "cards"
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.FIELDS_KEY_ERROR, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyFields(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	record["table"] = "cards"
	var fields map[string]interface{}
	record["fields"] = fields
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_FIELDS, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyFields1(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	record["table"] = "cards"
	record["fields"] = ""
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_FIELDS, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyColumn(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields[""] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	_, err := insertApi.Post("")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_COLUMN_NAME, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequest(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields["cvv"] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true}}
	json := `{
		"vaultID": "123",
		"responses": [
			{
				"records": [
					{
						"skyflow_id": "id1"
					}
				]
			},
			{
				"fields": {
					"first_name": "token1",
					"primary_card": {
						"card_number": "token2"
					}
				}
			}
		]
	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	insertApi.Post("")
}

func check(got string, wanted string, t *testing.T) {
	if got != wanted {
		t.Errorf("got  %s, wanted %s", got, wanted)
	}
}
