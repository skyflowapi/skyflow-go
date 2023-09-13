/*
Copyright (c) 2022 Skyflow, Inc.
*/
package vaultapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
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
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)

}

func TestEmptyVaultUrl(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_URL, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvalidVaultUrl(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "url", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_VAULT_URL, clientTag, configuration.VaultURL))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestInvalidVaultUrl1(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "http://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_VAULT_URL, clientTag, configuration.VaultURL))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestNoRecords(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyRecords(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record []interface{}
	records["records"] = record
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyRecordsWithContinueOnError(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var record []interface{}
	records["records"] = record
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false, ContinueOnError: true}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
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
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
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
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
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
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
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
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
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
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
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
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_COLUMN_NAME, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestTokensAndFieldMismatch(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields["columnName"] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	var tokens = make(map[string]interface{})
	tokens["card_number"] = "3388-5335-5239-3794"
	record["tokens"] = tokens
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "token")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISMATCH_OF_FIELDS_AND_TOKENS, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestTokensInvalidType(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields["columnName"] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	var tokens = "DEMO"
	record["tokens"] = tokens
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "token")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_TOKENS_IN_INSERT_RECORD, insertTag, reflect.TypeOf(tokens)))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestEmptyTokens(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields["columnName"] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	var tokens = make(map[string]interface{})
	record["tokens"] = tokens
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "token")
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKENS_IN_INSERT, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestWithTokens(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecordsWithTokens()
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true}}
	json := `{
				"records": [
					{
						"skyflow_id": "id1",
						"tokens": {
							"first_name": "token1",
							"primary_card": {
								"*": "id2",
								"card_number": "token2",
								"cvv": "token3",
								"expiry_date": "token4"
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
	ctx := context.TODO()
	insertApi.Post(ctx, "")
}

func TestEmptyColumnInUpsertOptions(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields["card_number"] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray

	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false, Upsert: upsertArray}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")

	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_COLUMN_IN_UPSERT_OPTIONS, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestEmptyTableInUpsertOptions(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.url.com", TokenProvider: GetToken}
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields["card_number"] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray

	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false, Upsert: upsertArray}}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")

	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_IN_UPSERT_OPTIONS, insertTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestWithContext(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray}}
	json := `
			{
				"records": [
					{
						"skyflow_id": "id1",
						"tokens": {
							"first_name": "token1",
							"primary_card": {
								"*": "id2",
								"card_number": "token2",
								"cvv": "token3",
								"expiry_date": "token4"
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
	ctx := context.TODO()
	insertApi.Post(ctx, "")
}

func TestValidRequest(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray}}
	json := `{
		"records": [
			{
				"skyflow_id": "id1",
				"tokens": {
					"first_name": "token1",
					"primary_card": {
						"*": "id2",
						"card_number": "token2",
						"cvv": "token3",
						"expiry_date": "token4"
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
	var ctx context.Context
	insertApi.Post(ctx, "")
}

func TestValidRequestWithTokensFalse(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
	jsonResp := `
			{
				"records": [
					{
						"skyflow_id": "id1"
					}
				]
			}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResp)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	ctx := context.TODO()
	res, _ := insertApi.Post(ctx, "")
	jsonResponse, _ := json.Marshal(res)
	var response common.InsertRecords
	err1 := json.Unmarshal(jsonResponse, &response)
	if err1 != nil {
		check(response.Records[0].Table, "cards", t)
	}
}

func TestInsertFailure(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false, ContinueOnError: true}}
	jsonResp := `{
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "400",
		"vaultID": "123",
		"error": {
			"grpc_code": "3",
			"http_code": "400",
			"http_status": "Bad Request",
			"message": "Object Name cards was not found for Vault 123"
	}
	`
	r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResp)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	ctx := context.TODO()
	_, err := insertApi.Post(ctx, "")
	if err == nil {
		t.Errorf("got nil, wanted skyflow error")
	}
}
func TestValidRequestWithContinueOnError(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray, ContinueOnError: true}}
	jsonStr := `{
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
			{
				"Body": {
			    "records": [
					{
						"skyflow_id": "id1",
						"tokens": {
							"first_name": "token1",
							"primary_card": {
								"*": "id2",
								"card_number": "token2",
								"cvv": "token3",
								"expiry_date": "token4"
							}
						}
					}
				]
				}
			}
		]
	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(jsonStr)))
	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}
	var ctx context.Context
	insertApi.Post(ctx, "")
}
func TestValidRequestWithContinueOnErrorErrorCase(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray, ContinueOnError: true}}
	json := `{
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
			{
			"Body": {
				"records": [
					{
						"skyflow_id": "id1",
						"tokens": {
							"first_name": "token1",
							"primary_card": {
								"*": "id2",
								"card_number": "token2",
								"cvv": "token3",
								"expiry_date": "token4"
							}
						}
					}
				]
			}
		},
		{
			"Body": {
                "error": "Object Name credit_card was not found for Vault u51547c101554f3489cf2ebadc72a858"
            },
            "Status": 400
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
	var ctx context.Context
	insertApi.Post(ctx, "")
}
func TestBuildResponseWithContinueOnErrorCase(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray, ContinueOnError: true}}
	jsonStr := `{
		"records": [
			{
				"skyflow_id": "id1",
				"tokens": {
					"first_name": "token1",
					"primary_card": {
						"*": "id2",
						"card_number": "token2",
						"cvv": "token3",
						"expiry_date": "token4"
					}
				}
			}
		]
	}`
	var data map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	expectedResponse := map[string]interface{}{
		"errors": []interface{}{},
		"records": []interface{}{
			map[string]interface{}{
				"request_index": 0,
				"table":         "cards",
				"fields": map[string]interface{}{
					"first_name": "token1",
					"primary_card": map[string]interface{}{
						"*":           "id2",
						"card_number": "token2",
						"cvv":         "token3",
						"expiry_date": "token4",
					},
					"skyflow_id": "id1",
				},
			},
		},
	}
	jsonRecord, _ := json.Marshal(records)
	var insertRecord common.InsertRecords
	if err := json.Unmarshal(jsonRecord, &insertRecord); err == nil {
		if responses, ok := data["responses"].([]interface{}); ok {
			actualResponse, _ := insertApi.buildResponseWithContinueOnErr(responses, insertRecord, "reqId-123")
			expectedJSON, _ := json.Marshal(expectedResponse)
			actualJSON, _ := json.Marshal(actualResponse)
			check(string(expectedJSON), string(actualJSON), t)
		}
	}

}
func TestBuildResponseWithContinueOnErrorErrorsCase(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray, ContinueOnError: true}}
	jsonStr := `{
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
			{
			"Body": {
				"records": [
					{
						"skyflow_id": "id1",
						"tokens": {
							"first_name": "token1",
							"primary_card": {
								"*": "id2",
								"card_number": "token2",
								"cvv": "token3",
								"expiry_date": "token4"
							}
						}
					}
				]
			}
		}, {
            "Body": {
                "error": "Object Name credit_card was not found for Vault"
            },
            "Status": 400
        }
		]
	}`
	var data map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	expectedResponse := map[string]interface{}{
		"errors": []interface{}{map[string]interface{}{
			"error": map[string]interface{}{
				"request_index": 1,
				"code":          "400",
				"description":   "[skyflow] Interface: Insert - Server error Object Name credit_card was not found for Vault - requestId : reqId-123",
			},
		}},
		"records": []interface{}{
			map[string]interface{}{
				"request_index": 0,
				"table":         "cards",
				"fields": map[string]interface{}{
					"first_name": "token1",
					"primary_card": map[string]interface{}{
						"*":           "id2",
						"card_number": "token2",
						"cvv":         "token3",
						"expiry_date": "token4",
					},
					"skyflow_id": "id1",
				},
			},
		},
	}
	jsonRecord, _ := json.Marshal(records)
	var insertRecord common.InsertRecords
	if err := json.Unmarshal(jsonRecord, &insertRecord); err == nil {
		if responses, ok := data["responses"].([]interface{}); ok {
			actualResponse, _ := insertApi.buildResponseWithContinueOnErr(responses, insertRecord, "reqId-123")
			expectedJSON, _ := json.Marshal(expectedResponse)
			actualJSON, _ := json.Marshal(actualResponse)
			check(string(expectedJSON), string(actualJSON), t)
		}
	}

}
func TestBuildResponseWithoutContinueOnErrorCase(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray, ContinueOnError: false}}
	jsonStr := `{
				"records": [
					{
						"skyflow_id": "id1",
						"tokens": {
							"first_name": "token1",
							"primary_card": {
								"*": "id2",
								"card_number": "token2",
								"cvv": "token3",
								"expiry_date": "token4"
							}
						}
					}
				]
			}`
	var data map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	expectedResponse := map[string]interface{}{
		"records": []interface{}{
			map[string]interface{}{
				"request_index": 0,
				"table":         "cards",
				"fields": map[string]interface{}{
					"first_name": "token1",
					"primary_card": map[string]interface{}{
						"*":           "id2",
						"card_number": "token2",
						"cvv":         "token3",
						"expiry_date": "token4",
					},
					"skyflow_id": "id1",
				},
			},
		},
	}
	jsonRecord, _ := json.Marshal(records)
	var insertRecord common.InsertRecords
	if err := json.Unmarshal(jsonRecord, &insertRecord); err == nil {
		if responses, ok := data["responses"].(map[string]interface{}); ok {
			actualResponse := insertApi.buildResponseWithoutContinueOnErr(responses, "table", records["records"].([]map[string]interface{}))
			expectedJSON, _ := json.Marshal(expectedResponse)
			actualJSON, _ := json.Marshal(actualResponse)
			check(string(expectedJSON), string(actualJSON), t)
		}
	}
}
func TestArrangeRecords(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := constructInsertRecords()
	var upsertArray []common.UpsertOptions
	var upsertOption = common.UpsertOptions{Table: "table1", Column: "column"}
	upsertArray = append(upsertArray, upsertOption)
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: true, Upsert: upsertArray, ContinueOnError: false}}

	testCase := struct {
		recordsArray []interface{}
		expected     map[string]interface{}
	}{
		recordsArray: []interface{}{
			map[string]interface{}{
				"table": "credit_card",
				"fields": map[string]interface{}{
					"card_number": "4111111111111142",
				},
				"tokens": map[string]interface{}{
					"card_number": "9991-3466-6577-4760",
				},
			},
			map[string]interface{}{
				"table": "credit_cards",
				"fields": map[string]interface{}{
					"card_number": "4111011111111114",
				},
			},
			map[string]interface{}{
				"table": "credit_cardss",
				"fields": map[string]interface{}{
					"cvv": "1234",
				},
			},
			map[string]interface{}{
				"table": "table3",
				"fields": map[string]interface{}{
					"card_pin":    "3888",
					"card_number": "4101111111111164",
				},
			},
			map[string]interface{}{
				"table": "credit_cards",
				"fields": map[string]interface{}{
					"card_number": "4111011111111114",
				},
			},
		},
		expected: map[string]interface{}{
			"RECORDS": map[string]interface{}{
				"credit_card": []interface{}{
					map[string]interface{}{
						"fields": map[string]interface{}{
							"card_number": "4111111111111142",
						},
						"tokens": map[string]interface{}{
							"card_number": "9991-3466-6577-4760",
						},
						"request_index": 0,
					},
				},
				"credit_cards": []interface{}{
					map[string]interface{}{
						"fields": map[string]interface{}{
							"card_number": "4111011111111114",
						},
						"request_index": 1,
					},
					map[string]interface{}{
						"fields": map[string]interface{}{
							"card_number": "4111011111111114",
						},
						"request_index": 4,
					},
				},
				"credit_cardss": []interface{}{
					map[string]interface{}{
						"fields": map[string]interface{}{
							"cvv": "1234",
						},
						"request_index": 2,
					},
				},
				"table3": []interface{}{
					map[string]interface{}{
						"fields": map[string]interface{}{
							"card_pin":    "3888",
							"card_number": "4101111111111164",
						},
						"request_index": 3,
					},
				},
			},
		},
	}

	result := insertApi.arrangeRecords(testCase.recordsArray)

	expectedJSON, _ := json.Marshal(testCase.expected)
	actualJSON, _ := json.Marshal(result)
	check(string(expectedJSON), string(actualJSON), t)

}
func TestBuildResponseWithoutContinueOnErrWithTokensAsFalse(t *testing.T) {

	// Test case : Records without tokens
	testCase := struct {
		responseJson map[string]interface{}
		tableName    string
		insertRecord []map[string]interface{}
		expected     common.ResponseBody
	}{
		responseJson: map[string]interface{}{
			"records": []interface{}{
				map[string]interface{}{
					"skyflow_id": "id1",
				},
				map[string]interface{}{
					"skyflow_id": "id2",
				},
			},
		},
		tableName: "credit_cards",
		insertRecord: []map[string]interface{}{
			{
				"fields":        map[string]interface{}{"card_number": "4111011111111114"},
				"request_index": 1,
			},
			{
				"fields":        map[string]interface{}{"card_number": "4111011111111114"},
				"request_index": 4,
			},
		},
		expected: common.ResponseBody{
			"records": []interface{}{
				map[string]interface{}{
					"request_index": 1,
					"fields": map[string]interface{}{
						"skyflow_id": "id1",
					},
					"table": "credit_cards",
				},
				map[string]interface{}{
					"request_index": 4,
					"fields": map[string]interface{}{
						"skyflow_id": "id2",
					},
					"table": "credit_cards",
				},
			},
		},
	}
	insertApi := &InsertApi{
		Options: common.InsertOptions{
			Tokens:          false,
			ContinueOnError: false,
		},
	}

	result := insertApi.buildResponseWithoutContinueOnErr(testCase.responseJson, testCase.tableName, testCase.insertRecord)

	expectedJSON, _ := json.Marshal(testCase.expected)
	actualJSON, _ := json.Marshal(result)
	check(string(expectedJSON), string(actualJSON), t)

}
func TestBuildResponseWithoutContinueOnErr(t *testing.T) {
	// Test case : Records with tokens
	testCase1 := struct {
		responseJson map[string]interface{}
		tableName    string
		insertRecord []map[string]interface{}
		expected     common.ResponseBody
	}{
		responseJson: map[string]interface{}{
			"records": []interface{}{
				map[string]interface{}{
					"skyflow_id": "id1",
					"tokens": map[string]interface{}{
						"card_number": "token1",
					},
				},
				map[string]interface{}{
					"skyflow_id": "id2",
					"tokens": map[string]interface{}{
						"card_number": "token2",
					},
				},
			},
		},
		tableName: "credit_cards",
		insertRecord: []map[string]interface{}{
			{
				"fields":        map[string]interface{}{"card_number": "4111011111111114"},
				"request_index": 1,
			},
			{
				"fields":        map[string]interface{}{"card_number": "4111011111111114"},
				"request_index": 4,
			},
		},
		expected: common.ResponseBody{
			"records": []interface{}{
				map[string]interface{}{
					"request_index": 1,
					"fields": map[string]interface{}{
						"card_number": "token1",
						"skyflow_id":  "id1",
					},
					"table": "credit_cards",
				},
				map[string]interface{}{
					"request_index": 4,
					"fields": map[string]interface{}{
						"card_number": "token2",
						"skyflow_id":  "id2",
					},
					"table": "credit_cards",
				},
			},
		},
	}

	runTestCase(t, testCase1)
}

func runTestCase(t *testing.T, testCase struct {
	responseJson map[string]interface{}
	tableName    string
	insertRecord []map[string]interface{}
	expected     common.ResponseBody
}) {
	insertApi := &InsertApi{
		Options: common.InsertOptions{
			Tokens:          true,
			ContinueOnError: false,
		},
	}

	result := insertApi.buildResponseWithoutContinueOnErr(testCase.responseJson, testCase.tableName, testCase.insertRecord)

	expectedJSON, _ := json.Marshal(testCase.expected)
	actualJSON, _ := json.Marshal(result)
	check(string(expectedJSON), string(actualJSON), t)
}

func TestAddIndexInErrorObject(t *testing.T) {
	// Test case 1: Error object with multiple errors
	testCase1 := struct {
		error        map[string]interface{}
		insertRecord []map[string]interface{}
		expected     common.ResponseBody
	}{
		error: map[string]interface{}{
			"error": map[string]interface{}{
				"code":        400,
				"description": "Object Name credit_cardss was not found for Vault s41b985164cf4145bbfc2a136f968186 - requestId : a3b8fd62-8d61-9b16-9285-837f80c9c625",
			},
		},
		insertRecord: []map[string]interface{}{
			{
				"fields":        map[string]interface{}{"card_number": "4111011111111114"},
				"request_index": 1,
			},
			{
				"fields":        map[string]interface{}{"card_number": "4111011111111114"},
				"request_index": 4,
			},
		},
		expected: common.ResponseBody{
			"errors": []interface{}{
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":          400,
						"description":   "Object Name credit_cardss was not found for Vault s41b985164cf4145bbfc2a136f968186 - requestId : a3b8fd62-8d61-9b16-9285-837f80c9c625",
						"request_index": 1,
					},
				},
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":          400,
						"description":   "Object Name credit_cardss was not found for Vault s41b985164cf4145bbfc2a136f968186 - requestId : a3b8fd62-8d61-9b16-9285-837f80c9c625",
						"request_index": 4,
					},
				},
			},
		},
	}

	// Test case 2: Error object with a single error
	testCase2 := struct {
		error        map[string]interface{}
		insertRecord []map[string]interface{}
		expected     common.ResponseBody
	}{
		error: map[string]interface{}{
			"error": map[string]interface{}{
				"code":        404,
				"description": "Object Name user_not_found was not found for Vault s41b985164cf4145bbfc2a136f968186 - requestId : a3b8fd62-8d61-9b16-9285-837f80c9c625",
			},
		},
		insertRecord: []map[string]interface{}{
			{
				"fields":        map[string]interface{}{"card_number": "4111011111111114"},
				"request_index": 1,
			},
		},
		expected: common.ResponseBody{
			"errors": []interface{}{
				map[string]interface{}{
					"error": map[string]interface{}{
						"code":          404,
						"description":   "Object Name user_not_found was not found for Vault s41b985164cf4145bbfc2a136f968186 - requestId : a3b8fd62-8d61-9b16-9285-837f80c9c625",
						"request_index": 1,
					},
				},
			},
		},
	}

	// Run the test cases
	runTestCase2(t, testCase1)
	runTestCase2(t, testCase2)
}

func runTestCase2(t *testing.T, testCase struct {
	error        map[string]interface{}
	insertRecord []map[string]interface{}
	expected     common.ResponseBody
}) {
	insertApi := &InsertApi{}

	result := insertApi.addIndexInErrorObject(testCase.error, testCase.insertRecord)

	if !reflect.DeepEqual(result, testCase.expected) {
		t.Errorf("AddIndexInErrorObject result does not match the expected output.\nExpected: %v\nActual: %v", testCase.expected, result)
	}
}
func constructInsertRecords() map[string]interface{} {
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	fields["cvv"] = "1234"
	record["table"] = "cards"
	record["fields"] = fields
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	return records
}
func constructInsertRecordsWithTokens() map[string]interface{} {
	records := make(map[string]interface{})
	var recordsArray []interface{}
	var record = make(map[string]interface{})
	var fields = make(map[string]interface{})
	var tokens = make(map[string]interface{})
	fields["first_name"] = "name"
	record["table"] = "cards"
	record["fields"] = fields
	tokens["first_name"] = "token1"
	record["tokens"] = tokens
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	return records
}

func check(got string, wanted string, t *testing.T) {
	if got != wanted {
		t.Errorf("got  %s, wanted %s", got, wanted)
	}
}
