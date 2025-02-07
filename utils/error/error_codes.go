package errors

// ErrorCodesEnum - Enum defines the list of error codes/categorization of errors in Skyflow.
type ErrorCodesEnum string

// Defining the values of Error code Enum
const (
	// SERVER - Represents server side error
	SERVER ErrorCodesEnum = "500"
	// INVALID_INPUT - Input passed was not invalid format
	INVALID_INPUT_CODE ErrorCodesEnum = "400"
	INVALID_INDEX      ErrorCodesEnum = "404"
	PARTIAL            ErrorCodesEnum = "404"
)
