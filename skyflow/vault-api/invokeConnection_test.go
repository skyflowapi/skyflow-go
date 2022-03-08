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
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_CONNECTION_URL, clientTag))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}
func TestInvalidConnectionUrl(t *testing.T) {
	configuration := common.ConnectionConfig{ConnectionURL: "url"}
	invokeApi := InvokeConnectionApi{configuration, ""}
	_, err := invokeApi.Post()
	skyflowError := errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_CONNECTION_URL, clientTag, configuration.ConnectionURL))
	check(err.GetMessage(), skyflowError.GetMessage(), t)
}

func TestValidRequestForConnections(t *testing.T) {

	path := make(map[string]string)
	path["card"] = "1234"
	query := make(map[string]interface{})
	query["cvv"] = 456
	query["a"] = "23"
	query["s"] = true
	query["float"] = 4.345
	req := make(map[string]interface{})
	req["sam"] = 123
	req["xx"] = 456

	configuration := common.ConnectionConfig{ConnectionURL: "https://www.google.com/card", MethodName: common.POST, PathParams: path, QueryParams: query, RequestBody: req}
	invokeApi := InvokeConnectionApi{configuration, ""}
	invokeApi.Post()

}
