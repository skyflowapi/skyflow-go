/*
	Copyright (c) 2022 Skyflow, Inc. 
*/
package vaultapi

import (
	"fmt"
	"testing"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func TestEmptyConnectionUrl(t *testing.T) {
	configuration := common.ConnectionConfig{ConnectionURL: ""}
	invokeApi := InvokeConnectionApi{configuration, ""}
	_, err := invokeApi.Post()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_CONNECTION_URL, connectionTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInvalidConnectionUrl(t *testing.T) {
	configuration := common.ConnectionConfig{ConnectionURL: "url"}
	invokeApi := InvokeConnectionApi{configuration, ""}
	_, err := invokeApi.Post()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_CONNECTION_URL, connectionTag, configuration.ConnectionURL))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestForConnections(t *testing.T) {

	pathParams := make(map[string]string)
	pathParams["card"] = "1234"
	queryParams := make(map[string]interface{})
	queryParams["cvv"] = 456
	queryParams["a"] = "23"
	queryParams["s"] = true
	queryParams["float"] = 4.345
	requestBody := make(map[string]interface{})
	requestBody["sam"] = 123
	requestBody["xx"] = 456

	configuration := common.ConnectionConfig{ConnectionURL: "https://www.google.com/card", MethodName: common.POST, PathParams: pathParams, QueryParams: queryParams, RequestBody: requestBody}
	invokeApi := InvokeConnectionApi{configuration, ""}
	invokeApi.Post()

}

func TestValidRequestUrlEncodedForConnections(t *testing.T) {

	requestBody := make(map[string]interface{})
	requestBody["sam"] = "123"
	requestBody["xx"] = "456"

	requestHeader := make(map[string]string)
	requestHeader["content-type"] = "multipart/form-data"

	configuration := common.ConnectionConfig{ConnectionURL: "https://www.google.com/card", MethodName: common.POST, RequestBody: requestBody, RequestHeader: requestHeader}
	invokeApi := InvokeConnectionApi{configuration, ""}
	invokeApi.Post()

}

func TestValidRequestFormDataForConnections(t *testing.T) {

	requestBody := make(map[string]interface{})
	requestBody["type"] = "card"
	card := make(map[string]interface{})
	card["number"] = 23.4
	card["exp_month"] = 12
	card["exp_year"] = "2024"
	card["valid"] = true
	x := make(map[string]interface{})
	x["sample"] = "sample"
	card["x"] = x
	requestBody["card"] = card

	requestHeader := make(map[string]string)
	requestHeader["content-type"] = "application/x-www-form-urlencoded"

	configuration := common.ConnectionConfig{ConnectionURL: "https://www.google.com/card", MethodName: common.POST, RequestBody: requestBody, RequestHeader: requestHeader}
	invokeApi := InvokeConnectionApi{configuration, ""}
	invokeApi.Post()

}

func TestUrlEncodeFunction(t *testing.T) {
	requestBody := make(map[string]interface{})
	requestBody["type"] = "card"
	card := make(map[string]interface{})
	card["number"] = 23.4
	card["exp_month"] = 12
	card["exp_year"] = "2024"
	card["valid"] = true
	x := make(map[string]interface{})
	x["sample"] = "sample"
	card["x"] = x
	requestBody["card"] = card
	var got = r_urlEncode(make([]interface{}, 0), make(map[string]string), requestBody)
	var wanted = "map[card[exp_month]:12 card[exp_year]:2024 card[number]:23.400000 card[valid]:true card[x][sample]:sample type:card]"
	check(fmt.Sprintf("%s", got), wanted, t)
}

func TestUrlEncodeFunctionForOtherDataTypes(t *testing.T) {
	var got = r_urlEncode(make([]interface{}, 0), make(map[string]string), 2.14)
	var wanted = "map[:2.140000]"
	check(fmt.Sprintf("%s", got), wanted, t)

	got = r_urlEncode(make([]interface{}, 0), make(map[string]string), 2)
	wanted = "map[:2]"
	check(fmt.Sprintf("%s", got), wanted, t)

	got = r_urlEncode(make([]interface{}, 0), make(map[string]string), true)
	wanted = "map[:true]"
	check(fmt.Sprintf("%s", got), wanted, t)

	var x float32 = 2.1
	got = r_urlEncode(make([]interface{}, 0), make(map[string]string), x)
	wanted = "map[:2.100000]"
	check(fmt.Sprintf("%s", got), wanted, t)

}
