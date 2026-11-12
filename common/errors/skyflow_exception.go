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
	// rawMessage is set by RawMessage: Error and GetMessage return the message as it is.
	rawMessage bool
}

// causeSeparator joins a raw message and its cause in Error(), the Go way: "message: cause".
const causeSeparator = ": "

// ErrorOption configures a SkyflowError as it is built. An SDK passes its options to every
// error it creates, so its errors behave as its users expect without changing the other's.
type ErrorOption func(*SkyflowError)

// RawMessage makes GetMessage return the message as it is, without the "Message: " prefix,
// and Error return it followed by its cause, as the Java SDK's getMessage() and the Python
// SDK's str(e) do. The flowvault SDK builds every error with it; v2's errors keep the prefix.
func RawMessage() ErrorOption {
	return func(se *SkyflowError) { se.rawMessage = true }
}

// WithResponseBody makes GetResponseBody return body: the server's error body as it was sent,
// as in the Java SDK, rather than the body the error was built from.
func WithResponseBody(body map[string]interface{}) ErrorOption {
	return func(se *SkyflowError) { se.responseBody = body }
}

func (se *SkyflowError) apply(opts []ErrorOption) *SkyflowError {
	for _, opt := range opts {
		opt(se)
	}
	return se
}

func (se *SkyflowError) Error() string {
	if se == nil {
		return ""
	}
	hasCause := se.originalError != nil && se.originalError.Error() != se.message
	if se.rawMessage {
		if hasCause {
			return se.message + causeSeparator + se.originalError.Error()
		}
		return se.message
	}
	if hasCause {
		return fmt.Sprintf("Message: %s, Original Error (if any): %s", se.message, se.originalError.Error())
	}
	return fmt.Sprintf("Message: %s", se.message) //nolint:revive
}
func (se *SkyflowError) GetMessage() string {
	if se == nil {
		return ""
	}
	if se.rawMessage {
		return se.message
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

// GetHttpCodeInt is GetHttpCode as a number, as the Java SDK's getHttpCode() returns it, or 0
// when the error has no HTTP code (for example a network failure).
func (se *SkyflowError) GetHttpCodeInt() int {
	return codeInt(se.GetHttpCode())
}

// GetGrpcCodeInt is GetGrpcCode as a number, as the Java SDK's getGrpcCode() returns it, or 0
// when the server sent none.
func (se *SkyflowError) GetGrpcCodeInt() int {
	return codeInt(se.GetGrpcCode())
}

func codeInt(code string) int {
	value, err := strconv.Atoi(code)
	if err != nil {
		return 0
	}
	return value
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

// GetResponseBody returns the JSON error body of an API error, decoded, or nil when the error
// didn't come from a JSON API response (validation, network and non-JSON errors).
func (se *SkyflowError) GetResponseBody() map[string]interface{} {
	if se == nil {
		return nil
	}
	return se.responseBody
}
func NewSkyflowError(code ErrorCodesEnum, message string, opts ...ErrorOption) *SkyflowError {
	skyflowError := &SkyflowError{
		httpCode:       string(code),
		message:        message,
		httpStatusCode: constants.HTTP_STATUS_BAD_REQUEST,
	}
	return skyflowError.apply(opts)
}

// NewSkyflowErrorWithCause is NewSkyflowError keeping the error that caused it, so callers
// can match it with errors.Is or errors.As, e.g. context.DeadlineExceeded. Unlike
// NewSkyflowError, whose status is always "Bad Request", SERVER gets "Internal Server Error"
// and REQUEST_CANCELLED gets "Client Closed Request"; other codes keep "Bad Request".
func NewSkyflowErrorWithCause(code ErrorCodesEnum, message string, cause error, opts ...ErrorOption) *SkyflowError {
	skyflowError := NewSkyflowError(code, message)
	skyflowError.httpStatusCode = statusFor(code)
	skyflowError.originalError = cause
	return skyflowError.apply(opts)
}

// statusFor is the HTTP status text for an SDK error code.
func statusFor(code ErrorCodesEnum) string {
	switch code {
	case SERVER:
		return http.StatusText(http.StatusInternalServerError)
	case REQUEST_CANCELLED:
		return constants.HTTP_STATUS_CLIENT_CLOSED_REQUEST
	case TRANSPORT_ERROR:
		return ""
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
		skyflowError.responseBody = apiError
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
			skyflowError.responseBody = apiError
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
func SkyflowErrorApi(error error, header http.Header, opts ...ErrorOption) *SkyflowError {
	skyflowError := SkyflowError{}
	var apiError map[string]interface{}
	// Set the request ID from the header
	skyflowError = SkyflowError{
		requestId: header.Get(constants.REQUEST_KEY),
	}
	parts := strings.SplitN(error.Error(), ": ", 2)
	if len(parts) < 2 {
		return NewSkyflowError(INVALID_INPUT_CODE, rawBodyMessage(error.Error()), opts...)
	}
	err := json.Unmarshal([]byte(parts[1]), &apiError)
	if err != nil {
		return NewSkyflowError(INVALID_INPUT_CODE, rawBodyMessage(error.Error()), opts...)
	}
	skyflowError.responseBody = apiError
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
		skyflowError.message = rawBodyMessage(error.Error())
	}
	return skyflowError.apply(opts)
}

// rawBodyMessage is a response body that is not a Skyflow JSON error, used as the message.
// A gateway or proxy can echo the request in such a body, and requests carry sensitive
// data, so only its first constants.MAX_RAW_ERROR_BODY bytes are kept.
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
