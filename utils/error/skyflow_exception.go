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
			return NewSkyflowError(INVALID_INPUT_CODE, "Failed to unmarhsal error")
		}
		skyflowError.details = apiError
		if errorBody, ok := apiError["error"].(map[string]interface{}); ok {
			skyflowError.httpCode = strconv.FormatFloat(errorBody["http_code"].(float64), 'f', 0, 64)
			skyflowError.message = errorBody["message"].(string)
			skyflowError.grpcCode = strconv.FormatFloat(errorBody["grpc_code"].(float64), 'f', 0, 64)
			skyflowError.httpStatusCode = errorBody["http_status"].(string)
		} else if errBody, ok := apiError["error"].(string); ok {
			skyflowError.message = errBody
		} else {
			skyflowError.message = string(bodyBytes)
		}

	} else if responseHeaders.Header.Get("Content-Type") == "text/plain" {
		bodyBytes, err := io.ReadAll(responseHeaders.Body)
		if err != nil {
			return NewSkyflowError(INVALID_INPUT_CODE, "Failed to read error")
		} else {
			skyflowError.message = string(bodyBytes)
			skyflowError.httpStatusCode = responseHeaders.Status
		}
	}
	return &skyflowError
}
