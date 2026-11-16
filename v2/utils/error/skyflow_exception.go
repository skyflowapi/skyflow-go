package errors

import (
	"net/http"

	ce "github.com/skyflowapi/skyflow-go/common/errors"
)

// SkyflowError is implemented once in skyflow-go/common/errors and shared by
// every SDK in this repository. The alias keeps it the same type under v2's
// import path, so existing v2 code and error handling are unaffected.
type SkyflowError = ce.SkyflowError

func NewSkyflowError(code ErrorCodesEnum, message string) *SkyflowError {
	return ce.NewSkyflowError(code, message)
}

func SkyflowApiError(responseHeaders http.Response) *SkyflowError {
	return ce.SkyflowApiError(responseHeaders)
}

func SkyflowErrorApi(error error, header http.Header) *SkyflowError {
	return ce.SkyflowErrorApi(error, header)
}
