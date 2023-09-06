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
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
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
	json := `{
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
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
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
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
	jsonResp := `{
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
			{
				"records": [
					{
						"skyflow_id": "id1"
					}
				]
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
	insertApi := InsertApi{Configuration: configuration, Records: records, Options: common.InsertOptions{Tokens: false}}
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
	}`
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
func TestValidRequestWithContinueOnErrorErrorCase(t *testing.T) {
	configuration := common.Configuration{VaultID: "123", VaultURL: "https://www.google.com", TokenProvider: GetToken}
	records := map[string]interface{}{"records": []interface{}{
		map[string]interface{}{"fields": map[string]interface{}{"cvv": "123"}, "table": "cards"},
		map[string]interface{}{"fields": map[string]interface{}{"cvv": "abc"}, "table": "cards"},
	}}
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
	var data map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	expectedResponse := map[string]interface{}{
		"errors": []interface{}{nil},
		"records": []interface{}{
			map[string]interface{}{
				"table": "cards",
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
	records := map[string]interface{}{"records": []interface{}{
		map[string]interface{}{"fields": map[string]interface{}{"cvv": "123"}, "table": "cards"},
		map[string]interface{}{"fields": map[string]interface{}{"cvv": "abc"}, "table": "cards"},
	}}
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
		"errors": []interface{}{
			nil,
			map[string]interface{}{
				"error": map[string]interface{}{
					"code":        "400",
					"description": "[skyflow] Interface: Insert - Server error Object Name credit_card was not found for Vault - requestId : reqId-123",
				},
			},
		},
		"records": []interface{}{
			map[string]interface{}{
				"table": "cards",
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
			nil,
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
		"Header" : {
			"x-request-id": "reqId-123"
		},
		"StatusCode": "200",
		"vaultID": "123",
		"responses": [
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
				"table": "cards",
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
			actualResponse := insertApi.buildResponseWithoutContinueOnErr(responses, insertRecord)
			expectedJSON, _ := json.Marshal(expectedResponse)
			actualJSON, _ := json.Marshal(actualResponse)
			check(string(expectedJSON), string(actualJSON), t)
		}
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
