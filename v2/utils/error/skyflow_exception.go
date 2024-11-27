package errors

import (
	"errors"
	"fmt"
)

type ErrorCodesEnum string

const (
	INVALID_INPUT ErrorCodesEnum = "400"
	InvalidInput                 = "InvalidInput"
	SdkErrorCode                 = "400"
)

type SkyflowError struct {
	code          ErrorCodesEnum
	message       string
	originalError error
}

func (se *SkyflowError) Error() string {
	return fmt.Sprintf("Message: %s, Original Error (if any): %s", se.message, se.originalError.Error())
}
func (se *SkyflowError) GetMessage() string {
	return fmt.Sprintf("Message: %s", se.message)
}

func (se *SkyflowError) GetCode() string {
	return fmt.Sprintf("Code: %s", se.code)
}
func NewSkyflowError(code ErrorCodesEnum, message string) *SkyflowError {
	return &SkyflowError{code: code, message: message, originalError: errors.New("<nil>")}
}
func NewSkyflowErrorf(code ErrorCodesEnum, format string, a ...interface{}) *SkyflowError {
	return NewSkyflowError(code, fmt.Sprintf(format, a...))
}
