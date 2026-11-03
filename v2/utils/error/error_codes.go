package errors

import ce "github.com/skyflowapi/skyflow-go/common/errors"

// ErrorCodesEnum - Enum defines the list of error codes/categorization of errors in Skyflow.
type ErrorCodesEnum = ce.ErrorCodesEnum

// Defining the values of Error code Enum
const (
	// SERVER - Represents server side error
	SERVER ErrorCodesEnum = ce.SERVER
	// INVALID_INPUT - Input passed was not invalid format
	INVALID_INPUT_CODE ErrorCodesEnum = ce.INVALID_INPUT_CODE
	INVALID_INDEX      ErrorCodesEnum = ce.INVALID_INDEX
	PARTIAL            ErrorCodesEnum = ce.PARTIAL
)
