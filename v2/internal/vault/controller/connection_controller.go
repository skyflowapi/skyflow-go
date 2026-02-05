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
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	errors "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"

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
		} else if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
			creds.CredentialsString = os.Getenv("SKYFLOW_CREDENTIALS")
		} else {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.INVALID_INPUT_CODE), errors.EMPTY_CREDENTIALS)
		}
	} else {
		creds = config.Credentials
	}
	return &creds, nil
}

func (v *ConnectionController) Invoke(ctx context.Context, request common.InvokeConnectionRequest) (*common.InvokeConnectionResponse, *errors.SkyflowError) {
	tag := "Invoke Connection"
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
		"request_id": requestId,
	}

	logger.Info(logs.INVOKE_CONNECTION_REQUEST_RESOLVED)
	// Step 7: Parse Response
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		response := common.InvokeConnectionResponse{Metadata: metaData}
		if res.Body != nil {
			contentType := res.Header.Get("Content-Type")
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
					return nil, errors.NewSkyflowError(errors.INVALID_INPUT_CODE, errors.INVALID_RESPONSE)
			}
			if strings.Contains(contentType, string(common.APPLICATIONXML)) || strings.Contains(contentType, string(common.TEXTORXML)) {
				response.Data = string(data)
				return &response, nil
			} else if strings.Contains(contentType, string(common.APPLICATIONORJSON)) || contentType == "" {
				var jsonData interface{}
				err = json.Unmarshal(data, &jsonData)
				if err != nil {
					response.Data = data
					return &response, nil
				} else {
					response.Data = jsonData
					return &response, nil
				}

			} else if strings.Contains(contentType, string(common.TEXTORPLAIN)) {
				response.Data = string(data)
				return &response, nil
			} else if strings.Contains(contentType, string(common.FORMURLENCODED)) {
				// Parse URL-encoded form data
				values, err := url.ParseQuery(string(data))
				if err != nil {
					return nil, errors.NewSkyflowError(errors.INVALID_INPUT_CODE, errors.INVALID_RESPONSE)
				}
				// Convert url.Values to map[string]interface{}
				result := make(map[string]interface{})
				for key, val := range values {
					if len(val) == 1 {
						result[key] = val[0]
					} else {
						result[key] = val
					}
				}
				response.Data = result
				return &response, nil
			} else if strings.Contains(contentType, string(common.FORMDATA)) {
				response.Data = string(data)
			} else if strings.Contains(contentType, string(common.TEXTHTML)) {
				response.Data = string(data)
				return &response, nil
			} else {
				response.Data = string(data)
				return &response, nil
			}
		}
		return &response, nil
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
	var contentType string
	var bodyContent string // For debugging
	shouldSetContentType := true
	
	contentType = detectContentType(request.Headers)

	// If no content-type and body is an object, default to JSON
	if contentType == string(common.APPLICATIONORJSON) && request.Body != nil {
		if _, ok := request.Body.(map[string]interface{}); ok {
			contentType = string(common.APPLICATIONORJSON)
		}
	}

	switch contentType {	
	case string(common.APPLICATIONORJSON):
		if strBody, ok := request.Body.(string); ok {
			body = strings.NewReader(strBody)
		} else if bodyMap, ok := request.Body.(map[string]interface{}); ok {
			data, err := json.Marshal(bodyMap)
			if err != nil {
				return nil, err
			}
			bodyContent = string(data)
			body = strings.NewReader(string(data))
		} else if request.Body != nil {
			if strBody, ok := request.Body.(string); ok {
				body = strings.NewReader(strBody)
			} else {
				body = strings.NewReader(fmt.Sprintf("%v", request.Body))
			}
		}	

	case string(common.FORMURLENCODED):
		if bodyMap, ok := request.Body.(map[string]interface{}); ok {
			urlParams := buildURLEncodedParams(bodyMap)
			bodyContent = urlParams.Encode()
			body = strings.NewReader(bodyContent)
		} else { //need to check here
			bodyContent = ""
			body = strings.NewReader("")	
		}

	case string(common.FORMDATA):
		buffer := new(bytes.Buffer)
		writer = multipart.NewWriter(buffer)
		
		if bodyMap, ok := request.Body.(map[string]interface{}); ok {
			for key, value := range bodyMap {
				if value == nil {
					continue
				}
				
				// Check if value is *os.File or io.Reader for file uploads
				if file, ok := value.(*os.File); ok {
					// Handle *os.File - create form file
					part, err := writer.CreateFormFile(key, file.Name())
					if err != nil {
						return nil, err
					}
					if _, err := io.Copy(part, file); err != nil {
						return nil, err
					}
				} else if reader, ok := value.(io.Reader); ok {
					// Handle io.Reader - create form file with generic name
					part, err := writer.CreateFormFile(key, key)
					if err != nil {
						return nil, err
					}
					if _, err := io.Copy(part, reader); err != nil {
						return nil, err
					}
				} else if nestedMap, ok := value.(map[string]interface{}); ok {
					// Check if value is a map/object - stringify it as JSON
					jsonData, err := json.Marshal(nestedMap)
					if err != nil {
						return nil, err
					}
					if err := writer.WriteField(key, string(jsonData)); err != nil {
						return nil, err
					}
				} else if arr, ok := value.([]interface{}); ok {
					// Handle arrays - stringify as JSON
					jsonData, err := json.Marshal(arr)
					if err != nil {
						return nil, err
					}
					if err := writer.WriteField(key, string(jsonData)); err != nil {
						return nil, err
					}
				} else {
					// Handle primitive values - convert to string
					if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
						return nil, err
					}
				}
			}
		} else if strBody, ok := request.Body.(string); ok {
			// If body is already a string, use it as-is (though this is unusual for multipart)
			bodyContent = strBody
			body = strings.NewReader(strBody)
			writer = nil // Don't use multipart writer for string body
			shouldSetContentType = false // Keep user's content-type
		} else if request.Body != nil {
			// For other types, convert to string
			bodyContent = fmt.Sprintf("%v", request.Body)
			body = strings.NewReader(bodyContent)
			writer = nil
			shouldSetContentType = false
		}
		
		if writer != nil {
			writer.Close()
			body = buffer
			bodyContent = buffer.String()
			contentType = writer.FormDataContentType() // set with boundary
			shouldSetContentType = true // Force set with boundary
		}

	case string(common.APPLICATIONXML), string(common.TEXTORXML):
		if strBody, ok := request.Body.(string); ok {
			// Body is already a string (raw XML)
			body = strings.NewReader(strBody)
		} else if bodyMap, ok := request.Body.(map[string]interface{}); ok {
			// Convert map to XML
			data, err := mapToXML(bodyMap)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(data)
		} else {
			// throw error for unsupported body type
		    return nil, errors.NewSkyflowError(errors.INVALID_INPUT_CODE, errors.INVALID_XML_FORMAT)
		}

	case string(common.TEXTORPLAIN):
		if strBody, ok := request.Body.(string); ok {
			bodyContent = strBody
			body = strings.NewReader(strBody)
		} else if request.Body != nil {
			body = strings.NewReader(fmt.Sprintf("%v", request.Body))
		} 
	case string(common.TEXTHTML):
		if strBody, ok := request.Body.(string); ok {
			bodyContent = strBody
			body = strings.NewReader(strBody)
		} else if bodyMap, ok := request.Body.(map[string]interface{}); ok {
			// send map as json in body
			data, err := json.Marshal(bodyMap)
			if err != nil {
				return nil, err
			}
			bodyContent = string(data)
			body = strings.NewReader(string(data))
		} else if request.Body != nil {
			bodyContent = fmt.Sprintf("%v", request.Body)
			body = strings.NewReader(bodyContent)
		} 
	
	default:
		if strBody, ok := request.Body.(string); ok {
			bodyContent = strBody
			body = strings.NewReader(strBody)
		} else if request.Body != nil {
			if bodyMap, ok := request.Body.(map[string]interface{}); ok {
				data, err := json.Marshal(bodyMap)
				if err != nil {
					return nil, err
				}
				bodyContent = string(data)
				body = strings.NewReader(string(data))
			} else {
				body = strings.NewReader(fmt.Sprintf("%v", request.Body))
			}
		}
	}
	if request.Method == "" {
		request.Method = common.POST
	}

	request1, err := http.NewRequest(string(request.Method), url, body)
	if err != nil {
		return nil, err
	}
	
	// Set content-type header
	if shouldSetContentType && contentType != "" {
		request1.Header.Set("content-type", contentType)
	}
	return request1, nil
}
func writeFormData(writer *multipart.Writer, requestBody interface{}) error {
	formData := RUrlencode(make([]interface{}, 0), make(map[string]string), requestBody)
	for key, value := range formData {
		if err := writer.WriteField(key, value); err != nil {
			return err
		}
	}
	return nil
}

// buildURLEncodedParams converts a map to URL encoded params matching Node.js URLSearchParams behavior
func buildURLEncodedParams(data map[string]interface{}) *url.Values {
	params := url.Values{}
	
	for key, value := range data {
		if value == nil {
			continue
		}
		
		// Check if value is a map (nested object)
		if nestedMap, ok := value.(map[string]interface{}); ok {
			for nestedKey, nestedValue := range nestedMap {
				paramKey := fmt.Sprintf("%s[%s]", key, nestedKey)
				params.Add(paramKey, fmt.Sprintf("%v", nestedValue))
			}
		} else if arr, ok := value.([]interface{}); ok {
			// Handle arrays
			for _, item := range arr {
				params.Add(key, fmt.Sprintf("%v", item))
			}
		} else {
			// Handle primitive values
			params.Add(key, fmt.Sprintf("%v", value))
		}
	}
	
	return &params
}

func RUrlencode(parents []interface{}, pairs map[string]string, data interface{}) map[string]string {

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
			RUrlencode(parents, pairs, value)
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
		request.Header.Set("x-skyflow-authorization", api.ApiKey)
	} else {
		request.Header.Set("x-skyflow-authorization", api.Token)
	}
	
	// Only set default content-type if not already set (preserve multipart boundary)
	if request.Header.Get("content-type") == "" {
		request.Header.Set("content-type", "application/json")
	}

	for key, value := range invokeRequest.Headers {
		// Skip content-type from user headers to preserve the one set in prepareRequest
		if strings.ToLower(key) == "content-type" {
			continue
		}
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

// mapToXML converts a map[string]interface{} to XML format
func mapToXML(data map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	buf.WriteString("<request>")
	
	for key, value := range data {
		writeXMLElement(&buf, key, value)
	}
	
	buf.WriteString("</request>")
	return buf.Bytes(), nil
}

// writeXMLElement recursively writes XML elements with proper escaping
func writeXMLElement(buf *bytes.Buffer, key string, value interface{}) {
	if value == nil {
		buf.WriteString(fmt.Sprintf("<%s/>", key))
		return
	}

	switch v := value.(type) {
	case map[string]interface{}:
		buf.WriteString(fmt.Sprintf("<%s>", key))
		for k, val := range v {
			writeXMLElement(buf, k, val)
		}
		buf.WriteString(fmt.Sprintf("</%s>", key))
	case []interface{}:
		for _, item := range v {
			writeXMLElement(buf, key, item)
		}
	default:
		// Escape special XML characters
		escapedValue := escapeXML(fmt.Sprintf("%v", v))
		buf.WriteString(fmt.Sprintf("<%s>%s</%s>", key, escapedValue, key))
	}
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
