package vaultapi

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/skyflowapi/skyflow-go/errors"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
	mocks "github.com/skyflowapi/skyflow-go/skyflow/utils/mocks"
)

func TestNoRecordsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.RECORDS_KEY_NOT_FOUND)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyRecordsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record []interface{}
	records["records"] = record
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORDS)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoTableForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_TABLE)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTableForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = ""
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TABLE_NAME)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoRedactionForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	var ids []interface{}
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_REDACTION)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoIdsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_KEY_IDS)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyIdsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["ids"] = ""
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORD_IDS)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyIdsForGetById1(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = "PLAIN_TEXT"
	var ids []interface{}
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORD_IDS)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTokenForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = "PLAIN_TEXT"
	var ids []interface{}
	ids = append(ids, "")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TOKEN_ID)
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken, Options: common.Options{LogLevel: common.WARN}}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = "PLAIN_TEXT"
	var ids []interface{}
	ids = append(ids, "id1")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}

	resJson := `{
		"records": [
			{
				"fields": {
					"first_name": "rach",
					"middle_name": "green",
					"skyflow_id": "id1"
				}
			}]
	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(resJson)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	getByIdApi.Get()

}
