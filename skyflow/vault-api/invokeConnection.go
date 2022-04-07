package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hetiansu5/urlquery"
	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type InvokeConnectionApi struct {
	ConnectionConfig common.ConnectionConfig
	Token            string
}

var connectionTag = "InvokeConnection"

func (InvokeConnectionApi *InvokeConnectionApi) doValidations() *errors.SkyflowError {

	logger.Info(fmt.Sprintf(messages.VALIDATE_CONNECTION_CONFIG, connectionTag))

	if InvokeConnectionApi.ConnectionConfig.ConnectionURL == "" {
		logger.Error(fmt.Sprintf(messages.EMPTY_CONNECTION_URL, connectionTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_CONNECTION_URL, connectionTag))
	} else if !isValidUrl(InvokeConnectionApi.ConnectionConfig.ConnectionURL) {
		logger.Error(fmt.Sprintf(messages.INVALID_CONNECTION_URL, connectionTag, InvokeConnectionApi.ConnectionConfig.ConnectionURL))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_CONNECTION_URL, connectionTag, InvokeConnectionApi.ConnectionConfig.ConnectionURL))
	}
	return nil
}

func (InvokeConnectionApi *InvokeConnectionApi) Post() (map[string]interface{}, *errors.SkyflowError) {

	validationError := InvokeConnectionApi.doValidations()
	if validationError != nil {
		return nil, validationError
	}
	requestUrl := InvokeConnectionApi.ConnectionConfig.ConnectionURL
	for index, value := range InvokeConnectionApi.ConnectionConfig.PathParams {
		requestUrl = strings.Replace(requestUrl, fmt.Sprintf("{%s}", index), value, -1)
	}
	var requestBody []byte
	var err error
	if InvokeConnectionApi.ConnectionConfig.RequestHeader["content-type"] == "application/x-www-form-urlencoded" {
		requestBody, err = urlquery.Marshal(InvokeConnectionApi.ConnectionConfig.RequestBody)
	} else {
		requestBody, err = json.Marshal(InvokeConnectionApi.ConnectionConfig.RequestBody)
	}

	if err != nil {
		logger.Error(fmt.Sprintf(messages.UNKNOWN_ERROR, connectionTag, err))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, connectionTag, err))
	}
	request, _ := http.NewRequest(
		InvokeConnectionApi.ConnectionConfig.MethodName.String(),
		requestUrl,
		strings.NewReader(string(requestBody)),
	)
	query := request.URL.Query()
	for index, value := range InvokeConnectionApi.ConnectionConfig.QueryParams {
		switch v := value.(type) {
		case int:
			query.Set(index, strconv.Itoa(v))
		case float64:
			query.Set(index, fmt.Sprintf("%f", v))
		case string:
			query.Set(index, v)
		case bool:
			query.Set(index, strconv.FormatBool(v))
		default:
			logger.Error(fmt.Sprintf(messages.INVALID_FIELD_IN_QUERY_PARAMS, connectionTag, index))
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_FIELD_IN_QUERY_PARAMS, connectionTag, index))
		}
	}
	request.URL.RawQuery = query.Encode()
	request.Header.Set("x-skyflow-authorization", InvokeConnectionApi.Token)
	request.Header.Set("content-type", "application/json")
	for index, value := range InvokeConnectionApi.ConnectionConfig.RequestHeader {
		request.Header.Set(strings.ToLower(index), value)
	}

	logger.Info(fmt.Sprintf(messages.INVOKE_CONNECTION_CALLED, connectionTag))

	res, err := http.DefaultClient.Do(request)
	var requestId = ""
	if res != nil {
		requestId = res.Header.Get("x-request-id")
	}
	if err != nil {
		logger.Error(common.AppendRequestId(fmt.Sprintf(messages.INVOKE_CONNECTION_FAILED, connectionTag), requestId))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, connectionTag, err), requestId))
	}
	data, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		logger.Error(common.AppendRequestId(fmt.Sprintf(messages.INVOKE_CONNECTION_FAILED, connectionTag), requestId))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), common.AppendRequestId(fmt.Sprintf(messages.UNKNOWN_ERROR, connectionTag, string(data)), requestId))
	}
	logger.Info(fmt.Sprintf(messages.INVOKE_CONNECTION_SUCCESS, connectionTag))
	return result, nil
}
