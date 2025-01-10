package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type SkyflowError struct {
	httpCode       string
	message        string
	requestId      string
	grpcCode       string
	httpStatusCode string
	details        map[string]interface{}
	responseBody   map[string]interface{}
	originalError  error
}

func (se *SkyflowError) Error() string {
	return fmt.Sprintf("Message: %s, Original Error (if any): %s", se.message, se.originalError.Error())
}
func (se *SkyflowError) GetMessage() string {
	return fmt.Sprintf("Message: %s", se.message)
}
func (se *SkyflowError) GetCode() string {
	return fmt.Sprintf("Code: %s", se.httpCode)
}
func (se *SkyflowError) GetRequestId() string {
	return se.requestId
}
func (se *SkyflowError) GetGrpcCode() string {
	return se.grpcCode
}
func (se *SkyflowError) GetHttpStatusCode() string {
	return se.httpStatusCode
}
func (se *SkyflowError) GetDetails() map[string]interface{} {
	return se.details
}
func (se *SkyflowError) GetResponseBody() map[string]interface{} {
	return se.responseBody
}
func NewSkyflowError(code ErrorCodesEnum, message string) *SkyflowError {
	return &SkyflowError{
		httpCode:       string(code),
		message:        message,
		httpStatusCode: string("Bad Request"),
	}
}
func SkyflowApiError(responseHeaders http.Response) *SkyflowError {
	skyflowError := SkyflowError{
		requestId: responseHeaders.Header.Get("x-request-id"),
	}
	if responseHeaders.Header.Get("Content-Type") == "application/json" {
		bodyBytes, _ := io.ReadAll(responseHeaders.Body)

		// Parse JSON into a struct
		var apiError map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &apiError); err != nil {
		}
		skyflowError.details = apiError
		errorBody := apiError["error"].(map[string]interface{})
		skyflowError.httpCode = strconv.FormatFloat(errorBody["http_code"].(float64), 'f', 0, 64)
		skyflowError.message = errorBody["message"].(string)
		skyflowError.grpcCode = strconv.FormatFloat(errorBody["grpc_code"].(float64), 'f', 0, 64)
		skyflowError.httpStatusCode = errorBody["http_status"].(string)
	} else if responseHeaders.Header.Get("Content-Type") == "text/plain" {
		bodyBytes, err := io.ReadAll(responseHeaders.Body)
		if err == nil {
			skyflowError.message = string(bodyBytes)
			skyflowError.httpStatusCode = responseHeaders.Status
		}
	}
	return &skyflowError
}
