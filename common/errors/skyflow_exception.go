package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	constants "github.com/skyflowapi/skyflow-go/common/internal/constants"

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
	if se == nil {
		return ""
	}
	if se.originalError != nil && se.originalError.Error() != se.message {
		return fmt.Sprintf("Message: %s, Original Error (if any): %s", se.message, se.originalError.Error())
	}
	return fmt.Sprintf("Message: %s", se.message) //nolint:revive
}
func (se *SkyflowError) GetMessage() string {
	if se == nil {
		return ""
	}
	return fmt.Sprintf("Message: %s", se.message) //nolint:revive
}

// Deprecated: Use GetHttpCode instead.
func (se *SkyflowError) GetCode() string {
	if se == nil {
		return ""
	}
	return fmt.Sprintf("Code: %s", se.httpCode)
}
func (se *SkyflowError) GetHttpCode() string {
	if se == nil {
		return ""
	}
	return se.httpCode
}
func (se *SkyflowError) GetRequestId() string {
	if se == nil {
		return ""
	}
	return se.requestId
}
func (se *SkyflowError) GetGrpcCode() string {
	if se == nil {
		return ""
	}
	return se.grpcCode
}
func (se *SkyflowError) GetHttpStatusCode() string {
	if se == nil {
		return ""
	}
	return se.httpStatusCode
}
func (se *SkyflowError) GetDetails() []interface{} {
	if se == nil {
		return nil
	}
	return se.details
}
func (se *SkyflowError) GetResponseBody() map[string]interface{} {
	if se == nil {
		return nil
	}
	return se.responseBody
}
func NewSkyflowError(code ErrorCodesEnum, message string) *SkyflowError {
	return &SkyflowError{
		httpCode:       string(code),
		message:        message,
		httpStatusCode: constants.HTTP_STATUS_BAD_REQUEST,
	}
}

// NewSkyflowErrorWithCause is NewSkyflowError keeping the error that caused it, so callers
// can match it with errors.Is or errors.As, e.g. context.DeadlineExceeded. Unlike
// NewSkyflowError, whose status is always "Bad Request", its HTTP status matches code.
func NewSkyflowErrorWithCause(code ErrorCodesEnum, message string, cause error) *SkyflowError {
	skyflowError := NewSkyflowError(code, message)
	skyflowError.httpStatusCode = statusFor(code)
	skyflowError.originalError = cause
	return skyflowError
}

// statusFor is the HTTP status text for an SDK error code.
func statusFor(code ErrorCodesEnum) string {
	switch code {
	case SERVER:
		return http.StatusText(http.StatusInternalServerError)
	case REQUEST_CANCELLED:
		return constants.HTTP_STATUS_CLIENT_CLOSED_REQUEST
	}
	return constants.HTTP_STATUS_BAD_REQUEST
}

// Unwrap returns the error that caused this one, if any.
func (se *SkyflowError) Unwrap() error {
	if se == nil {
		return nil
	}
	return se.originalError
}

// SkyflowApiError builds an error from an API response: its request ID, and the message,
// gRPC code and details from a JSON error body (else the raw body).
func SkyflowApiError(responseHeaders http.Response) *SkyflowError {
	skyflowError := SkyflowError{
		requestId: responseHeaders.Header.Get(constants.REQUEST_KEY),
	}
	ct := responseHeaders.Header.Get(constants.HEADER_CONTENT_TYPE_CAPITAL)
	if strings.Contains(ct, constants.CONTENT_TYPE_JSON) {
		bodyBytes, _ := io.ReadAll(responseHeaders.Body)
		var apiError map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &apiError); err != nil {
			return NewSkyflowError(INVALID_INPUT_CODE, "Failed to unmarshal error")
		}
		if errorBody, ok := apiError[constants.ERROR_KEY_ERROR].(map[string]interface{}); ok {
			if httpCode, exists := errorBody[constants.ERROR_KEY_HTTP_CODE].(float64); exists {
				skyflowError.httpCode = strconv.FormatFloat(httpCode, 'f', 0, 64)
			} else {
				skyflowError.httpCode = strconv.Itoa(responseHeaders.StatusCode)
			}
			if message, exists := errorBody[constants.ERROR_KEY_MESSAGE].(string); exists {
				skyflowError.message = message
			} else {
				skyflowError.message = constants.UNKNOWN_ERROR
			}
			if grpcCode, exists := errorBody[constants.ERROR_KEY_GRPC_CODE].(float64); exists {
				skyflowError.grpcCode = strconv.FormatFloat(grpcCode, 'f', 0, 64)
			}
			if httpStatus, exists := errorBody[constants.ERROR_KEY_HTTP_STATUS].(string); exists {
				skyflowError.httpStatusCode = httpStatus
			}
			if details, exists := errorBody[constants.ERROR_KEY_DETAILS].([]interface{}); exists {
				if skyflowError.details == nil {
					skyflowError.details = make([]interface{}, 0)
				}
				skyflowError.details = details
			}
		} else if errBody, ok := apiError[constants.ERROR_KEY_ERROR].(string); ok {
			skyflowError.message = errBody
		} else {
			skyflowError.message = rawBodyMessage(string(bodyBytes))
		}
	} else if responseHeaders.Header.Get(constants.HEADER_CONTENT_TYPE_CAPITAL) == constants.CONTENT_TYPE_TEXT_PLAIN {
		bodyBytes, err := io.ReadAll(responseHeaders.Body)
		if err != nil {
			return NewSkyflowError(INVALID_INPUT_CODE, constants.ERROR_FAILED_TO_READ)
		} else {
			skyflowError.message = rawBodyMessage(string(bodyBytes))
			skyflowError.httpStatusCode = responseHeaders.Status
		}
	} else if responseHeaders.Header.Get(constants.HEADER_CONTENT_TYPE_CAPITAL) == constants.CONTENT_TYPE_TEXT_CHARSET {
		bodyBytes, errs := io.ReadAll(responseHeaders.Body)
		if errs != nil {
			return NewSkyflowError(INVALID_INPUT_CODE, constants.ERROR_FAILED_TO_READ)
		}
		// Parse JSON into a struct
		var apiError map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &apiError); err != nil {
			skyflowError.message = rawBodyMessage(string(bodyBytes))
			skyflowError.httpStatusCode = responseHeaders.Status
		}
		if apiError != nil {
			if errorBody, ok := apiError[constants.ERROR_KEY_ERROR].(map[string]interface{}); ok {
				if httpCode, exists := errorBody[constants.ERROR_KEY_HTTP_CODE].(float64); exists {
					skyflowError.httpCode = strconv.FormatFloat(httpCode, 'f', 0, 64)
				} else {
					skyflowError.httpCode = strconv.Itoa(responseHeaders.StatusCode)
				}
				if message, exists := errorBody[constants.ERROR_KEY_MESSAGE].(string); exists {
					skyflowError.message = message
				} else {
					skyflowError.message = constants.UNKNOWN_ERROR
				}
				if grpcCode, exists := errorBody[constants.ERROR_KEY_GRPC_CODE].(float64); exists {
					skyflowError.grpcCode = strconv.FormatFloat(grpcCode, 'f', 0, 64)
				}
				if httpStatus, exists := errorBody[constants.ERROR_KEY_HTTP_STATUS].(string); exists {
					skyflowError.httpStatusCode = httpStatus
				}
				if details, exists := errorBody[constants.ERROR_KEY_DETAILS].([]interface{}); exists {
					if skyflowError.details == nil {
						skyflowError.details = make([]interface{}, 0)
					}
					skyflowError.details = details
				}
			} else if errBody, ok := apiError[constants.ERROR_KEY_ERROR].(string); ok {
				skyflowError.message = errBody
				skyflowError.httpStatusCode = responseHeaders.Status
			} else {
				skyflowError.message = rawBodyMessage(string(bodyBytes))
				skyflowError.httpStatusCode = responseHeaders.Status
			}
		} else {
			skyflowError.message = rawBodyMessage(string(bodyBytes))
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
		errorDetail[constants.ERROR_KEY_FROM_CLIENT] = boolValue
		skyflowError.details = append(skyflowError.details, errorDetail)
	}
	return &skyflowError
}

// SkyflowErrorApi converts an error from a generated API client into a SkyflowError. When the
// error carries an HTTP status, the error keeps that status and the request ID even if the
// body is empty or not JSON, so a 5xx from a gateway is not reported as invalid input.
func SkyflowErrorApi(error error, header http.Header) *SkyflowError {
	var apiError map[string]interface{}
	// Set the request ID from the header
	skyflowError := SkyflowError{
		requestId: header.Get(constants.REQUEST_KEY),
	}
	raw := error.Error()
	prefix, body, found := strings.Cut(raw, constants.API_ERROR_SEPARATOR)
	status, hasStatus := parseStatusPrefix(prefix)
	if !found || json.Unmarshal([]byte(body), &apiError) != nil {
		if !hasStatus {
			return NewSkyflowError(INVALID_INPUT_CODE, rawBodyMessage(raw))
		}
		skyflowError.message = rawBodyMessage(raw)
		skyflowError.fillMissingStatus(status)
		return &skyflowError
	}
	if errorBody, ok := apiError[constants.ERROR_KEY_ERROR].(map[string]interface{}); ok {
		if httpCode, exists := errorBody[constants.ERROR_KEY_HTTP_CODE].(float64); exists {
			skyflowError.httpCode = strconv.FormatFloat(httpCode, 'f', 0, 64)
		}
		if message, exists := errorBody[constants.ERROR_KEY_MESSAGE].(string); exists {
			skyflowError.message = message
		} else {
			skyflowError.message = constants.UNKNOWN_ERROR
		}
		if grpcCode, exists := errorBody[constants.ERROR_KEY_GRPC_CODE].(float64); exists {
			skyflowError.grpcCode = strconv.FormatFloat(grpcCode, 'f', 0, 64)
		}
		if httpStatus, exists := errorBody[constants.ERROR_KEY_HTTP_STATUS].(string); exists {
			skyflowError.httpStatusCode = httpStatus
		}
		if details, exists := errorBody[constants.ERROR_KEY_DETAILS].([]interface{}); exists {
			if skyflowError.details == nil {
				skyflowError.details = make([]interface{}, 0)
			}
			skyflowError.details = details
		}
	} else if errBody, ok := apiError[constants.ERROR_KEY_ERROR].(string); ok {
		skyflowError.message = errBody
	} else {
		skyflowError.message = rawBodyMessage(raw)
	}
	if hasStatus {
		// The body's own http_code and http_status win; the response status fills the gaps.
		skyflowError.fillMissingStatus(status)
	}
	return &skyflowError
}

// rawBodyMessage is a response body that is not a Skyflow JSON error, used as the message.
// A gateway or proxy can echo the request in such a body, and Detect requests carry sensitive
// text, so only its first constants.MAX_RAW_ERROR_BODY bytes are kept.
func rawBodyMessage(body string) string {
	if len(body) <= constants.MAX_RAW_ERROR_BODY {
		return body
	}
	cut := constants.MAX_RAW_ERROR_BODY
	for cut > 0 && !utf8.RuneStart(body[cut]) {
		cut-- // don't split a multi-byte character
	}
	return body[:cut] + constants.TRUNCATED_SUFFIX
}

// parseStatusPrefix reads the HTTP status that generated clients put before the body.
func parseStatusPrefix(prefix string) (int, bool) {
	status, err := strconv.Atoi(prefix)
	if err != nil || status < constants.MIN_HTTP_STATUS || status > constants.MAX_HTTP_STATUS {
		return 0, false
	}
	return status, true
}

// fillMissingStatus sets the HTTP code from status where it is still unset, and the status
// text from the final code, so the two always agree.
func (se *SkyflowError) fillMissingStatus(status int) {
	if se.httpCode == "" {
		se.httpCode = strconv.Itoa(status)
	}
	if se.httpStatusCode == "" {
		// httpCode is numeric here: set above or formatted from the body's http_code.
		code, _ := strconv.Atoi(se.httpCode)
		se.httpStatusCode = http.StatusText(code)
	}
}
