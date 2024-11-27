package errors

import (
	"fmt"
)

type ErrorCodesEnum string

const (
	INVALID_INPUT ErrorCodesEnum = "400"
)

type SkyflowError struct {
	httpCode       string
	message        string
	requestId      string
	grpcCode       string
	httpStatusCode string
	details        map[string]interface{}
	responseBody   map[string]interface{}
}

func (se *SkyflowError) Error() string {
	return fmt.Sprintf("Message: %s, Original Error (if any): %s", se.message)
}
func (se *SkyflowError) GetMessage() string {
	return fmt.Sprintf("Message: %s", se.message)
}

func (se *SkyflowError) GetCode() string {
	return fmt.Sprintf("Code: %s", se.httpCode)
}
func NewSkyflowError(code string, message string) *SkyflowError {
	return &SkyflowError{httpCode: code, message: message}
}
func NewSkyflowErrorf(code string, format string, a ...interface{}) *SkyflowError {
	return NewSkyflowError(code, fmt.Sprintf(format, a...))
}

func SkyflowApiError(httpCode string, message string, requestId string, grpcCode string, httpStatusCode string) *SkyflowError {
	return &SkyflowError{
		httpCode:  httpCode,
		message:   message,
		requestId: requestId,
		grpcCode:  grpcCode,
	}
}
