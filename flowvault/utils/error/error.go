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
	INVALID_INDEX      ErrorCodesEnum = v2errors.INVALID_INDEX
	PARTIAL            ErrorCodesEnum = v2errors.PARTIAL
)

func NewSkyflowError(code ErrorCodesEnum, message string) *SkyflowError {
	return v2errors.NewSkyflowError(code, message)
}

func SkyflowApiError(responseHeaders http.Response) *SkyflowError {
	return v2errors.SkyflowApiError(responseHeaders)
}

func SkyflowErrorApi(err error, header http.Header) *SkyflowError {
	return v2errors.SkyflowErrorApi(err, header)
}
