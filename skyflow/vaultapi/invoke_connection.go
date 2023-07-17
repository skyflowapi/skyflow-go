/*
Copyright (c) 2022 Skyflow, Inc.
*/
package vaultapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/hetiansu5/urlquery"
	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

// This is the description for InvokeConnectionApi struct
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

// This is the description for Post function
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
	var request *http.Request
	var writer *multipart.Writer
	var contentType = string(common.APPLICATIONORJSON)
	for index, value := range InvokeConnectionApi.ConnectionConfig.RequestHeader {
		var key = strings.ToLower(index)
		if key == "content-type" {
			contentType = value
			break
		}
	}
	if contentType == string(common.FORMURLENCODED) {
		requestBody, err = urlquery.Marshal(InvokeConnectionApi.ConnectionConfig.RequestBody)
		request, _ = http.NewRequest(
			InvokeConnectionApi.ConnectionConfig.MethodName.String(),
			requestUrl,
			strings.NewReader(string(requestBody)),
		)
	} else if contentType == string(common.FORMDATA) {
		body := new(bytes.Buffer)
		writer = multipart.NewWriter(body)
		for key, val := range r_urlEncode(make([]interface{}, 0), make(map[string]string), InvokeConnectionApi.ConnectionConfig.RequestBody) {
			_ = writer.WriteField(key, val)
		}
		err = writer.Close()
		if err != nil {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, connectionTag, err))
		}
		request, _ = http.NewRequest(
			InvokeConnectionApi.ConnectionConfig.MethodName.String(),
			requestUrl,
			body,
		)
	} else {
		requestBody, err = json.Marshal(InvokeConnectionApi.ConnectionConfig.RequestBody)
		request, _ = http.NewRequest(
			InvokeConnectionApi.ConnectionConfig.MethodName.String(),
			requestUrl,
			strings.NewReader(string(requestBody)),
		)
	}
	if err != nil {
		logger.Error(fmt.Sprintf(messages.UNKNOWN_ERROR, connectionTag, err))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, connectionTag, err))
	}
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
	skyMetadata := common.CreateJsonMetadata()
	request.Header.Add("sky-metadata", skyMetadata)
	for index, value := range InvokeConnectionApi.ConnectionConfig.RequestHeader {
		var key = strings.ToLower(index)
		if key == "content-type" && value == "multipart/form-data" {
			request.Header.Set(key, writer.FormDataContentType())
		} else {
			request.Header.Set(key, value)
		}
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
