package vaultapi

import (
	"bytes"
	errors1 "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/commonutils/mocks"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func TestNoRecordsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyRecordsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record []interface{}
	records["records"] = record
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoTableForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TABLE, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTableForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = ""
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_NAME, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoRedactionForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
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
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_REDACTION, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoIdsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_KEY_IDS, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyIdsForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["ids"] = ""
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyIdsForGetById1(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var ids []interface{}
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTokenForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var ids []interface{}
	ids = append(ids, "")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getByIdApi.Get()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKEN_ID, getByIdTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
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

func TestInValidRequestForGetById(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var ids []interface{}
	ids = append(ids, "id1")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	resJson := `{
		"error": {
				"grpc_code": 5,
				"http_code": 404,
				"http_status": "Not Found",
				"message": "Token not found for 1234"
			}
	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(resJson)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	resp, _ := getByIdApi.Get()
	if resp["errors"] == nil {
		t.Errorf("got nil, wanted skyflow error")
	}
}

func TestInValidRequestForGetByIdWithErrors(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var ids []interface{}
	ids = append(ids, "id1")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	resJson := `{
		"errors": [
		"error": {
				"grpc_code": 5,
				"http_code": 404,
				"http_status": "Not Found",
				"message": "Token not found for 1234"
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
	resp, _ := getByIdApi.Get()
	if resp["errors"] == nil {
		t.Errorf("got nil, wanted skyflow error")
	}
}

func TestInValidRequestForGetByIdWithErr(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var ids []interface{}
	ids = append(ids, "id1")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getByIdApi := GetByIdApi{Configuration: configuration, Records: records, Token: ""}
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return nil, errors1.New("")
	}
	resp, _ := getByIdApi.Get()
	if resp["errors"] == nil {
		t.Errorf("got nil, wanted skyflow error")
	}
}
