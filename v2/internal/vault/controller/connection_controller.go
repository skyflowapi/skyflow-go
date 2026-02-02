package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	"github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	errors "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"

	"github.com/hetiansu5/urlquery"
)

type ConnectionController struct {
	Config      *common.ConnectionConfig
	CommonCreds *common.Credentials
	Loglevel    *logger.LogLevel
	Token       string
	ApiKey      string
}

var SetBearerTokenForConnectionControllerFunc = setBearerTokenForConnectionController

// SetBearerTokenForConnectionController checks and updates the token if necessary.
func setBearerTokenForConnectionController(v *ConnectionController) *errors.SkyflowError {
	// Validate token or generate a new one if expired or not set.
	// check if apikey or token already initalised
	credToUse, err := setConnectionCredentials(v.Config, v.CommonCreds)
	if err != nil {
		return err
	}
	if v.Token == "" || serviceaccount.IsExpired(v.Token) {
		logger.Info(logs.GENERATE_BEARER_TOKEN_TRIGGERED)
		token, err := GenerateToken(*credToUse)
		if err != nil {
			return err
		}
		v.Token = *token
	} else {
		logger.Info(logs.REUSE_BEARER_TOKEN)
	}
	return nil
}
func setConnectionCredentials(config *common.ConnectionConfig, builderCreds *common.Credentials) (*common.Credentials, *errors.SkyflowError) {
	// here if credentials are empty in the vaultapi config
	creds := common.Credentials{}
	if config == nil || isCredentialsEmpty(config.Credentials) {
		// here if builder credentials are available
		if builderCreds != nil && !isCredentialsEmpty(*builderCreds) {
			creds = *builderCreds
		} else if envCreds := os.Getenv(constants.SKYFLOW_CREDENTIALS_ENV); envCreds != "" {
			creds.CredentialsString = os.Getenv(constants.SKYFLOW_CREDENTIALS_ENV)
		} else {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.EMPTY_CREDENTIALS)
		}
	} else {
		creds = config.Credentials
	}
	return &creds, nil
}

func (v *ConnectionController) Invoke(ctx context.Context, request common.InvokeConnectionRequest) (*common.InvokeConnectionResponse, *errors.SkyflowError) {
	tag := constants.REQUEST_INVOKE_CONN
	logger.Info(logs.INVOKE_CONNECTION_TRIGGERED)
	// Step 1: Validate Configuration
	logger.Info(logs.VALIDATING_INVOKE_CONNECTION_REQUEST)
	er := validation.ValidateInvokeConnectionRequest(request)
	if er != nil {
		return nil, er
	}
	err := SetBearerTokenForConnectionControllerFunc(v)
	if err != nil {
		return nil, err
	}

	// Step 2: Build Request URL
	requestUrl := buildRequestURL(
		v.Config.ConnectionUrl,
		request.PathParams,
	)

	// Step 3: Prepare Request Body
	requestBody, err1 := prepareRequest(
		request,
		requestUrl,
	)
	if err1 != nil {
		logger.Error(fmt.Sprintf(logs.INVALID_REQUEST_HEADERS, tag))
		return nil, errors.NewSkyflowError(errors.INVALID_INPUT_CODE, fmt.Sprintf(errors.UNKNOWN_ERROR, err1.Error()))
	}

	// Step 4: Set Query Params
	err2 := setQueryParams(requestBody, request.QueryParams)
	if err2 != nil {
		logger.Error(fmt.Sprintf(logs.INVALID_QUERY_PARAM, tag))
		return nil, err2
	}

	// Step 5: Set Headers
	setHeaders(requestBody, *v, request)

	// Step 6: Send Request
	res, requestId, invokeErr := sendRequest(requestBody)
	if invokeErr != nil {
		logger.Error(logs.INVOKE_CONNECTION_REQUEST_REJECTED)
		return nil, errors.NewSkyflowError(errors.INVALID_INPUT_CODE, fmt.Sprintf(errors.UNKNOWN_ERROR, invokeErr.Error()))
	}
	metaData := map[string]interface{}{
		constants.REQUEST_ID_KEY: requestId,
	}

	logger.Info(logs.INVOKE_CONNECTION_REQUEST_RESOLVED)
	// Step 7: Parse Response
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		parseRes, parseErr := parseResponse(res)
		if parseErr != nil {
			return nil, parseErr
		}
		return &common.InvokeConnectionResponse{Data: parseRes, Metadata: metaData}, nil
	}
	return nil, errors.SkyflowApiError(*res)
}

// Utility Functions
func buildRequestURL(baseURL string, pathParams map[string]string) string {
	for key, value := range pathParams {
		baseURL = strings.Replace(baseURL, fmt.Sprintf("{%s}", key), value, -1)
	}
	return baseURL
}
func prepareRequest(request common.InvokeConnectionRequest, url string) (*http.Request, error) {
	var body io.Reader
	var writer *multipart.Writer
	contentType := detectContentType(request.Headers)

	switch contentType {
	case string(common.FORMURLENCODED):
		data, err := urlquery.Marshal(request.Body)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(data))

	case string(common.FORMDATA):
		buffer := new(bytes.Buffer)
		writer = multipart.NewWriter(buffer)
		if err := writeFormData(writer, request.Body); err != nil {
			return nil, err
		}
		writer.Close()
		body = buffer

	default:
		data, err := json.Marshal(request.Body)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(data))
	}
	if request.Method == "" {
		request.Method = common.POST
	}

	request1, err := http.NewRequest(string(request.Method), url, body)
	if err == nil && writer != nil {
		request1.Header.Set(constants.HEADER_CONTENT_TYPE, writer.FormDataContentType())
	}
	return request1, err
}
func writeFormData(writer *multipart.Writer, requestBody interface{}) error {
	formData := rUrlencode(make([]interface{}, 0), make(map[string]string), requestBody)
	for key, value := range formData {
		if err := writer.WriteField(key, value); err != nil {
			return err
		}
	}
	return nil
}
func rUrlencode(parents []interface{}, pairs map[string]string, data interface{}) map[string]string {

	switch reflect.TypeOf(data).Kind() {
	case reflect.Int:
		pairs[renderKey(parents)] = fmt.Sprintf("%d", data)
	case reflect.Float32:
		pairs[renderKey(parents)] = fmt.Sprintf("%f", data) //nolint:revive
	case reflect.Float64:
		pairs[renderKey(parents)] = fmt.Sprintf("%f", data) //nolint:revive
	case reflect.Bool:
		pairs[renderKey(parents)] = fmt.Sprintf("%t", data)
	case reflect.Map:
		var mapOfdata = (data).(map[string]interface{})
		for index, value := range mapOfdata {
			parents = append(parents, index)
			rUrlencode(parents, pairs, value)
			parents = parents[:len(parents)-1]
		}
	default:
		pairs[renderKey(parents)] = fmt.Sprintf("%s", data)
	}
	return pairs
}
func renderKey(parents []interface{}) string {
	var depth = 0
	var outputString = ""
	for index := range parents {
		var typeOfindex = reflect.TypeOf(parents[index]).Kind()
		if depth > 0 || typeOfindex == reflect.Int {
			outputString = outputString + fmt.Sprintf("[%v]", parents[index])
		} else {
			outputString = outputString + (parents[index]).(string)
		}
		depth = depth + 1
	}
	return outputString
}
func detectContentType(headers map[string]string) string {
	for key, value := range headers {
		if strings.ToLower(key) == constants.HEADER_CONTENT_TYPE {
			return value
		}
	}
	return string(common.APPLICATIONORJSON)
}
func setQueryParams(request *http.Request, queryParams map[string]interface{}) *errors.SkyflowError {
	query := request.URL.Query()
	for key, value := range queryParams {
		switch v := value.(type) {
		case int:
			query.Set(key, strconv.Itoa(v))
		case float64:
			query.Set(key, fmt.Sprintf("%f", v))
		case string:
			query.Set(key, v)
		case bool:
			query.Set(key, strconv.FormatBool(v))
		default:
			return errors.NewSkyflowError(errors.INVALID_INPUT_CODE, errors.INVALID_QUERY_PARAM)
		}
	}
	request.URL.RawQuery = query.Encode()
	return nil
}
func setHeaders(request *http.Request, api ConnectionController, invokeRequest common.InvokeConnectionRequest) {
	if api.ApiKey != "" {
		request.Header.Set(constants.HEADER_AUTHORIZATION, api.ApiKey)
	} else {
		request.Header.Set(constants.HEADER_AUTHORIZATION, api.Token)
	}
	request.Header.Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_JSON)

	for key, value := range invokeRequest.Headers {
		request.Header.Set(key, value)
	}
}
func sendRequest(request *http.Request) (*http.Response, string, error) {
	response, err := http.DefaultClient.Do(request)
	requestId := ""
	if response != nil {
		requestId = response.Header.Get(constants.RESPONSE_HEADER_REQUEST_ID)
	}
	if err != nil {
		return nil, requestId, err
	}
	return response, requestId, nil
}
func parseResponse(response *http.Response) (map[string]interface{}, *errors.SkyflowError) {
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.NewSkyflowError(errors.INVALID_INPUT_CODE, errors.INVALID_RESPONSE)
	}
	var result map[string]interface{}
	if err1 := json.Unmarshal(data, &result); err1 != nil {
		return nil, errors.NewSkyflowError(errors.INVALID_INPUT_CODE, errors.INVALID_RESPONSE)
	}
	return result, nil
}
