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
	// REQUEST_CANCELLED - The call's context was cancelled or its deadline passed before the
	// call finished ("client closed request").
	REQUEST_CANCELLED ErrorCodesEnum = "499"
	// TRANSPORT_ERROR - The call got no usable HTTP response: a network failure, a timeout,
	// or an empty or unreadable body. No HTTP status applies, so the code is "0" and the
	// status text is empty.
	TRANSPORT_ERROR ErrorCodesEnum = "0"
)
