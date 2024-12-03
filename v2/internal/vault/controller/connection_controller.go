package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hetiansu5/urlquery"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"reflect"
	"skyflow-go/v2/internal/validation"
	"skyflow-go/v2/serviceaccount"
	. "skyflow-go/v2/utils/common"
	. "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
	"strconv"
	"strings"
)

type ConnectionController struct {
	Config   ConnectionConfig
	Loglevel *logger.LogLevel
	Token    string
	ApiKey   string
}

var SetBearerTokenForConnectionControllerFunc = setBearerTokenForConnectionController

// SetBearerTokenForConnectionController checks and updates the token if necessary.
func setBearerTokenForConnectionController(v *ConnectionController) *SkyflowError {
	// Validate token or generate a new one if expired or not set.
	if v.Token == "" || serviceaccount.IsExpired(v.Token) {
		token, err := GenerateToken(v.Config.Credentials)
		if err != nil {
			return err
		}
		v.Token = *token
	}
	return nil
}

func (v *ConnectionController) Invoke(ctx *context.Context, request *InvokeConnectionRequest) (*InvokeConnectionResponse, *SkyflowError) {
	// Step 1: Validate Configuration
	er := validation.ValidateInvokeConnectionRequest(*request)
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
		*request,
		requestUrl,
	)
	if err1 != nil {
		return nil, NewSkyflowError("400", "request not prepared")
	}

	// Step 4: Set Query Params
	err2 := setQueryParams(requestBody, request.QueryParams)
	if err2 != nil {
		return nil, NewSkyflowError("400", "query param not added")
	}

	// Step 5: Set Headers
	setHeaders(requestBody, *v, *request)

	// Step 6: Send Request
	res, requestId, invokeErr := sendRequest(requestBody)
	if invokeErr != nil {
		return nil, NewSkyflowError("400", "request not prepared")
	}

	// Step 7: Parse Response
	parseRes, parseErr := parseResponse(res, requestId)
	if parseErr != nil {
		return nil, NewSkyflowError("400", "response parse error")
	}
	return &InvokeConnectionResponse{Response: parseRes}, nil
}

// Utility Functions
func buildRequestURL(baseURL string, pathParams map[string]string) string {
	for key, value := range pathParams {
		baseURL = strings.Replace(baseURL, fmt.Sprintf("{%s}", key), value, -1)
	}
	return baseURL
}
func prepareRequest(request InvokeConnectionRequest, url string) (*http.Request, error) {
	var body io.Reader
	var writer *multipart.Writer
	contentType := detectContentType(request.Headers)

	switch contentType {
	case string(FORMURLENCODED):
		data, err := urlquery.Marshal(request.Body)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(data))

	case string(FORMDATA):
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

	request1, err := http.NewRequest(string(request.Method), url, body)
	if err == nil && writer != nil {
		request1.Header.Set("content-type", writer.FormDataContentType())
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
		pairs[renderKey(parents)] = fmt.Sprintf("%f", data)
	case reflect.Float64:
		pairs[renderKey(parents)] = fmt.Sprintf("%f", data)
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
		if strings.ToLower(key) == "content-type" {
			return value
		}
	}
	return string(APPLICATIONORJSON)
}
func setQueryParams(request *http.Request, queryParams map[string]interface{}) error {
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
			return NewSkyflowError("400", "invalid query param")
		}
	}
	request.URL.RawQuery = query.Encode()
	return nil
}
func setHeaders(request *http.Request, api ConnectionController, invokeRequest InvokeConnectionRequest) {
	request.Header.Set("x-skyflow-authorization", api.Token)
	request.Header.Set("content-type", "application/json")

	for key, value := range invokeRequest.Headers {
		request.Header.Set(key, value)
	}
}
func sendRequest(request *http.Request) (*http.Response, string, error) {
	response, err := http.DefaultClient.Do(request)
	requestId := ""
	if response != nil {
		requestId = response.Header.Get("x-request-id")
	}
	if err != nil {
		return nil, requestId, err
	}
	return response, requestId, nil
}
func parseResponse(response *http.Response, requestId string) (map[string]interface{}, *SkyflowError) {
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, NewSkyflowError("400", "error in parsing")
	}
	var result map[string]interface{}
	if err1 := json.Unmarshal(data, &result); err1 != nil {
		return nil, NewSkyflowError("400", "error in parsing")
	}
	return result, nil
}
