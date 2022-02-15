package errors

import (
	"fmt"
)

// SkyflowError - The Error Object for Skyflow APIs. Contains :
// 1. code - Represents the type of error
// 2. message - The message contained in the error
// 3. originalError - The original error (if any) which resulted in this error
type SkyflowError struct {
	code          ErrorCodesEnum
	message       string
	originalError error
}

// NewSkyflowErrorf - Creates a new Skyflow Error Object with Parameter Substitution
func NewSkyflowErrorf(code ErrorCodesEnum, format string, a ...interface{}) *SkyflowError {
	return NewSkyflowError(code, fmt.Sprintf(format, a...))
}

// NewSkyflowError - Creates a new Skyflow Error Object with given message
func NewSkyflowError(code ErrorCodesEnum, message string) *SkyflowError {
	return &SkyflowError{code: code, message: message}
}

// NewSkyflowErrorWrap - Creates a new Skyflow Error Object using the given error
func NewSkyflowErrorWrap(code ErrorCodesEnum, err error, message string) *SkyflowError {
	return &SkyflowError{code: code, message: message, originalError: err}
}

// GetOriginalError - Returns the underlying error (if any)
func (se *SkyflowError) GetOriginalError() error {
	return se.originalError
}

// Error - Uses the Underlying go's error for providing Error() interface impl.
func (se *SkyflowError) Error() string {
	return fmt.Sprintf("Message: %s, Original Error (if any): %s", se.message, se.originalError.Error())
}

func (se *SkyflowError) GetMessage() string {
	return fmt.Sprintf("Message: %s", se.message)
}

func (se *SkyflowError) GetCode() string {
	return fmt.Sprintf("Code: %s", se.code)
}
