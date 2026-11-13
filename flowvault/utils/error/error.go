package errors

import (
	"net/http"

	v2errors "github.com/skyflowapi/skyflow-go/common/errors"
)

// Shared SkyflowError from skyflow-go/common/errors, re-exported under flowvault's path. It is the
// same type, so errors from either SDK can be handled uniformly.
type SkyflowError = v2errors.SkyflowError
type ErrorCodesEnum = v2errors.ErrorCodesEnum

const (
	SERVER             ErrorCodesEnum = v2errors.SERVER
	INVALID_INPUT_CODE ErrorCodesEnum = v2errors.INVALID_INPUT_CODE
	// REQUEST_CANCELLED - The call's context ended before the call finished.
	REQUEST_CANCELLED ErrorCodesEnum = v2errors.REQUEST_CANCELLED
	// TRANSPORT_ERROR - The call got no usable HTTP response (code "0", empty status).
	TRANSPORT_ERROR ErrorCodesEnum = v2errors.TRANSPORT_ERROR
)

// Options are the options every flowvault error is built with: RawMessage, so GetMessage
// returns the message as it is, and EmptyValidationDetails, so a validation error has empty,
// non-nil details. The v2 SDK's errors, the same type, keep their "Message: " prefix and nil
// details.
func Options() []v2errors.ErrorOption {
	return []v2errors.ErrorOption{v2errors.RawMessage(), v2errors.EmptyValidationDetails()}
}

func NewSkyflowError(code ErrorCodesEnum, message string) *SkyflowError {
	return v2errors.NewSkyflowError(code, message, Options()...)
}

// NewSkyflowErrorWithCause is NewSkyflowError keeping the error that caused it, so callers
// can match it with errors.Is or errors.As, e.g. context.DeadlineExceeded.
func NewSkyflowErrorWithCause(code ErrorCodesEnum, message string, cause error) *SkyflowError {
	return v2errors.NewSkyflowErrorWithCause(code, message, cause, Options()...)
}

func SkyflowErrorApi(err error, header http.Header) *SkyflowError {
	return v2errors.SkyflowErrorApi(err, header, Options()...)
}
