package error

//  TO DO

import "errors"

type SkyflowError struct {
	code          string
	message       string
	originalError error
}

func NewSkyflowError(code string, message string) *SkyflowError {
	return &SkyflowError{code: code, message: message, originalError: errors.New("<nil>")}
}
