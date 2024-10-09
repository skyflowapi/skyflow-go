/*
Copyright (c) 2022 Skyflow, Inc.
*/
package vaultapi

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/skyflowapi/skyflow-go/v1/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/v1/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/v1/commonutils/mocks"
	"github.com/skyflowapi/skyflow-go/v1/skyflow/common"
)

func TestNoRecordsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyRecordsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record []interface{}
	records["records"] = record
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoTableForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TABLE, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInvalidTableForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = make(map[string]interface{})
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_TABLE_NAME_TYPE, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTableForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = ""
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_NAME, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestNoRedactionForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	var ids []interface{}
	ids = append(ids, "123")
	record1["ids"] = ids

	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_REDACTION, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvalidRedactionForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = 123
	var ids []interface{}
	ids = append(ids, "123")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_REDACTION_TYPE, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvalidIdsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	record1["ids"] = "ids"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_IDS_TYPE, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInvalidColumnValuesForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	record1["columnValues"] = "values"
	record1["columnName"] = "name"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_COLUMN_VALUES_IN_GET, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInvalidColumnNameForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var columnValues []interface{}
	columnValues = append(columnValues, "123")
	record1["columnValues"] = columnValues
	record1["columnName"] = columnValues
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_COLUMN_NAME, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestNoIdsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_IDS_OR_COLUMN_VALUES_IN_GET, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyIdsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	record1["ids"] = ""
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyIdsForGet1(t *testing.T) {
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
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTokenForGet(t *testing.T) {
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
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKEN_ID, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestIdsRedactionWithOptionsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var ids []interface{}
	ids = append(ids, "123")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Options: common.GetOptions{Tokens: true}, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.REDACTION_WITH_TOKEN_NOT_SUPPORTED, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestColumnValuesWithOptionsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var columnValues []interface{}
	columnValues = append(columnValues, "123")
	record1["columnValues"] = columnValues
	record1["columnName"] = "name"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Options: common.GetOptions{Tokens: true}, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.TOKENS_GET_COLUMN_NOT_SUPPORTED, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestColumnNameMissingForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var columnValues []interface{}
	columnValues = append(columnValues, "123")
	record1["columnValues"] = columnValues
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Options: common.GetOptions{Tokens: false}, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_COLUMN_NAME, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestColumnValuesMissingForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	record1["columnName"] = "columnName"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Options: common.GetOptions{Tokens: false}, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_RECORD_COLUMN_VALUE, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestColumnValuesWithIdsForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var ids []interface{}
	ids = append(ids, "123")
	record1["ids"] = ids
	var columnValues []interface{}
	columnValues = append(columnValues, "123")
	record1["columnValues"] = columnValues
	record1["columnName"] = "name"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Options: common.GetOptions{Tokens: false}, Token: ""}
	_, err := getApi.GetRecords(context.TODO())
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.SKYFLOW_IDS_AND_COLUMN_NAME_BOTH_SPECIFIED, getTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestForGet(t *testing.T) {
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
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}

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
	getApi.GetRecords(context.TODO())

}
func TestValidRequestWithTokensForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	var ids []interface{}
	ids = append(ids, "id1")
	record1["ids"] = ids
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Options: common.GetOptions{Tokens: true}, Token: ""}

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
	getApi.GetRecords(context.TODO())

}

func TestValidRequestWithColumnValuesForGet(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["table"] = "cards"
	record1["redaction"] = common.PLAIN_TEXT
	var columnValues []interface{}
	columnValues = append(columnValues, "123")
	record1["columnValues"] = columnValues
	record1["columnName"] = "card_pin"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	getApi := GetApi{Configuration: configuration, Records: records, Token: ""}

	resJson := `{
		"records": [
			{
				"fields": {
					"first_name": "rach",
					"middle_name": "green",
					"card_pin": "123"
					"skyflow_id": "nil"
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
	getApi.GetRecords(context.TODO())

}
