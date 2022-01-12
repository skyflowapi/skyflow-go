package errors

// ErrorCodesEnum - Enum defines the list of error codes/categorization of errors in Skyflow.
type ErrorCodesEnum string

// Defining the values of Error code Enum
const (
	// Server - Represents server side error
	Server ErrorCodesEnum = "Server"
	// InvalidInput - Input passed was not invalid format
	InvalidInput = "InvalidInput"
	SdkErrorCode = "400"
)
