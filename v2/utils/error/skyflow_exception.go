package errors

import (
	"encoding/json"
	"fmt"
	"strings"

	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"

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
	details        []interface{}
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
func (se *SkyflowError) GetDetails() []interface{} {
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
		requestId: responseHeaders.Header.Get(constants.REQUEST_KEY),
	}
	if responseHeaders.Header.Get("Content-Type") == "application/json" {
		bodyBytes, _ := io.ReadAll(responseHeaders.Body)
		// Parse JSON into a struct
		var apiError map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &apiError); err != nil {
			return NewSkyflowError(INVALID_INPUT_CODE, "Failed to unmarhsal error")
		}
		if errorBody, ok := apiError["error"].(map[string]interface{}); ok {
			if httpCode, exists := errorBody["http_code"].(float64); exists {
				skyflowError.httpCode = strconv.FormatFloat(httpCode, 'f', 0, 64)
			} else {
				skyflowError.httpCode = strconv.Itoa(responseHeaders.StatusCode)
			}
			if message, exists := errorBody["message"].(string); exists {
				skyflowError.message = message
			} else {
				skyflowError.message = "Unknown error"
			}
			if grpcCode, exists := errorBody["grpc_code"].(float64); exists {
				skyflowError.grpcCode = strconv.FormatFloat(grpcCode, 'f', 0, 64)
			}
			if httpStatus, exists := errorBody["http_status"].(string); exists {
				skyflowError.httpStatusCode = httpStatus
			}
			if details, exists := errorBody["details"].([]interface{}); exists {
				// initalize details if nil
				if skyflowError.details == nil {
					skyflowError.details = make([]interface{}, 0)
				}
				skyflowError.details = details
			}
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
	} else if responseHeaders.Header.Get("Content-Type") == "text/plain; charset=utf-8" {
		bodyBytes, errs := io.ReadAll(responseHeaders.Body)
		if errs != nil {
			return NewSkyflowError(INVALID_INPUT_CODE, "Failed to read error")
		}
		// Parse JSON into a struct
		var apiError map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &apiError); err != nil {
			skyflowError.message = string(bodyBytes)
			skyflowError.httpStatusCode = responseHeaders.Status
		}
		if apiError != nil {
			if errorBody, ok := apiError["error"].(map[string]interface{}); ok {
				if httpCode, exists := errorBody["http_code"].(float64); exists {
					skyflowError.httpCode = strconv.FormatFloat(httpCode, 'f', 0, 64)
				} else {
					skyflowError.httpCode = strconv.Itoa(responseHeaders.StatusCode)
				}
				if message, exists := errorBody["message"].(string); exists {
					skyflowError.message = message
				} else {
					skyflowError.message = "Unknown error"
				}
				if grpcCode, exists := errorBody["grpc_code"].(float64); exists {
					skyflowError.grpcCode = strconv.FormatFloat(grpcCode, 'f', 0, 64)
				}
				if httpStatus, exists := errorBody["http_status"].(string); exists {
					skyflowError.httpStatusCode = httpStatus
				}
				if details, exists := errorBody["details"].([]interface{}); exists {
				   if skyflowError.details == nil {
					skyflowError.details = make([]interface{}, 0)
				   }
					skyflowError.details = details
				}
			} else if errBody, ok := apiError["error"].(string); ok {
				skyflowError.message = errBody
				skyflowError.httpStatusCode = responseHeaders.Status
			} else {
				skyflowError.message = string(bodyBytes)
				skyflowError.httpStatusCode = responseHeaders.Status
			}
		} else {
			skyflowError.message = string(bodyBytes)
			skyflowError.httpStatusCode = responseHeaders.Status
		}
	}

	if responseHeaders.Header.Get(constants.ERROR_FROM_CLIENT) != "" {
		if skyflowError.details == nil {
			skyflowError.details = make([]interface{}, 0)
		}
		// create a map to hold the error detail
		errorDetail := make(map[string]interface{})
		// convert the header value to boolean string
		boolValue, err := strconv.ParseBool(responseHeaders.Header.Get(constants.ERROR_FROM_CLIENT))
		if err != nil {
			boolValue = false
		}
		// set the error detail
		errorDetail["errorFromClient"] = boolValue
		skyflowError.details = append(skyflowError.details, errorDetail)
	}
	return &skyflowError
}
func SkyflowErrorApi(error error, header http.Header) *SkyflowError {
	skyflowError := SkyflowError{}
	var apiError map[string]interface{}
	// Set the request ID from the header
	skyflowError = SkyflowError{
		requestId: header.Get(constants.REQUEST_KEY),
	}
	parts := strings.SplitN(error.Error(), ": ", 2)
	if len(parts) < 2 {
		return NewSkyflowError(INVALID_INPUT_CODE, error.Error())
	}
	err := json.Unmarshal([]byte(parts[1]), &apiError)
	if err != nil {
		return NewSkyflowError(INVALID_INPUT_CODE, error.Error())
	}
	if errorBody, ok := apiError["error"].(map[string]interface{}); ok {
		if httpCode, exists := errorBody["http_code"].(float64); exists {
			skyflowError.httpCode = strconv.FormatFloat(httpCode, 'f', 0, 64)
		}
		if message, exists := errorBody["message"].(string); exists {
			skyflowError.message = message
		} else {
			skyflowError.message = "Unknown error"
		}
		if grpcCode, exists := errorBody["grpc_code"].(float64); exists {
			skyflowError.grpcCode = strconv.FormatFloat(grpcCode, 'f', 0, 64)
		}
		if httpStatus, exists := errorBody["http_status"].(string); exists {
			skyflowError.httpStatusCode = httpStatus
		}
		if details, exists := errorBody["details"].([]interface{}); exists {
			if skyflowError.details == nil {
				skyflowError.details = make([]interface{}, 0)
			}
			skyflowError.details = details
		}
	} else if errBody, ok := apiError["error"].(string); ok {
		skyflowError.message = errBody
	} else {
		skyflowError.message = error.Error()
	}
	return &skyflowError
}
