/*
Copyright (c) 2022 Skyflow, Inc.
*/
package client

import (
	"context"
	"errors"
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
	return "nil", errors.New("Not Valid")
}
func validToken() (string, error) {

	validFilePath := os.Getenv("CREDENTIALS_FILE_PATH")

	token, err := saUtil.GenerateBearerToken(validFilePath)
	if err != nil {
		return "", err
	} else {
		return token.AccessToken, nil
	}
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
func TestInsertInValidByot(t *testing.T) {
	configuration := common.Configuration{VaultID: "id", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.Insert(record, common.InsertOptions{Tokens: true, Byot: "demo"})
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.INVALID_BYOT_TYPE, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInsertByotTokensNotPassed(t *testing.T) {
	configuration := common.Configuration{VaultID: "id", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var records = make(map[string]interface{})
	var record = make(map[string]interface{})
	record["table"] = "credit_cards"
	var fields = make(map[string]interface{})
	fields["cardholder_name"] = "name"
	fields["card_number"] = "4111111111111112"
	record["fields"] = fields
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	_, err := client.Insert(record, common.InsertOptions{Tokens: true, Byot: common.ENABLE})
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.NO_TOKENS_IN_INSERT, clientTag, "ENABLE"))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInsertByotTokensNotAllPassed(t *testing.T) {
	configuration := common.Configuration{VaultID: "id", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var records = make(map[string]interface{})
	var record = make(map[string]interface{})
	record["table"] = "credit_cards"
	var fields = make(map[string]interface{})
	fields["cardholder_name"] = "name"
	fields["card_number"] = "4111111111111112"
	record["fields"] = fields
	var tokens = make(map[string]interface{})
	tokens["cardholder_name"] = "token1"
	record["tokens"] = tokens
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	_, err := client.Insert(record, common.InsertOptions{Tokens: true, Byot: common.ENABLE_STRICT})
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInsertByotNotPassedforTokens(t *testing.T) {
	configuration := common.Configuration{VaultID: "id", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var records = make(map[string]interface{})
	var record = make(map[string]interface{})
	record["table"] = "credit_cards"
	var fields = make(map[string]interface{})
	fields["cardholder_name"] = "name"
	fields["card_number"] = "4111111111111112"
	record["fields"] = fields
	var tokens = make(map[string]interface{})
	tokens["cardholder_name"] = "token1"
	record["tokens"] = tokens
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	_, err := client.Insert(record, common.InsertOptions{Tokens: true})
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.TOKENS_PASSED_FOR_BYOT_DISABLE, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInsertValidTokenWithContext(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	ctx := context.TODO()
	_, err := client.Insert(record, common.InsertOptions{Tokens: true, Context: ctx})
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

func TestDetokenizeValidTokenWithContext(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	ctx := context.TODO()
	_, err := client.Detokenize(record, common.DetokenizeOptions{Context: ctx})
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestDetokenizeBulkValidTokenWithContext(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	ctx := context.TODO()
	_, err := client.Detokenize(record, common.DetokenizeOptions{Context: ctx, ContinueOnError: false})
	skyflowError := errors1.NewSkyflowError(errors1.ErrorCodesEnum(errors1.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestDetokenizeBulkValidTokenWithoutContext(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	_, err := client.Detokenize(record, common.DetokenizeOptions{ContinueOnError: false})
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

func TestGetByIdValidTokenWithContext(t *testing.T) {
	configuration := common.Configuration{VaultID: "", VaultURL: "https://www.url.com", TokenProvider: validToken}
	var client = Init(configuration)
	var record = make(map[string]interface{})
	ctx := context.TODO()
	_, err := client.GetById(record, common.GetByIdOptions{Context: ctx})
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
