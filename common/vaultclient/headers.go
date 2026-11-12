package vaultclient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/skyflowapi/skyflow-go/common/config"
	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

// MetricsHeader carries the SDK name, version and runtime with every call.
const MetricsHeader = "sky-metadata"

// NotBatched is RequestContext's BatchIndex and TotalBatches for an operation that was not
// split into batches, as in the Java and Python SDKs. Every call made through Base sends a
// single request, so it is the only value today.
const NotBatched = -1

// RequestContext is passed to an options Interceptor before each API call.
type RequestContext struct {
	Operation string
	// BatchIndex is the zero-based position of this batch within the caller's request, and
	// TotalBatches the number of batches; both are NotBatched (-1) for an unbatched call.
	BatchIndex   int
	TotalBatches int
	headers      map[config.CustomHeaderKey]string
}

// AddHeader adds a custom header to the outgoing request.
func (rc *RequestContext) AddHeader(key config.CustomHeaderKey, value string) {
	if rc.headers == nil {
		rc.headers = make(map[config.CustomHeaderKey]string)
	}
	rc.headers[key] = value
}

// Headers returns a copy of the headers added so far.
func (rc *RequestContext) Headers() map[config.CustomHeaderKey]string {
	headers := make(map[config.CustomHeaderKey]string, len(rc.headers))
	for key, value := range rc.headers {
		headers[key] = value
	}
	return headers
}

// Interceptor customises a single call, e.g. by adding headers.
type Interceptor func(rc *RequestContext)

// buildHeaders merges the client-level headers, the interceptor's per-call headers and the
// SDK metadata header, in that order.
func (b *Base) buildHeaders(operation string, clientHeaders map[config.CustomHeaderKey]string, interceptor Interceptor) (http.Header, *skyflowError.SkyflowError) {
	if clientHeaders != nil && b.profile.ClientHeadersTag != "" {
		if err := b.profile.validateHeaders(clientHeaders, b.profile.ClientHeadersTag); err != nil {
			return nil, err
		}
	}
	headers := http.Header{}
	for key, value := range clientHeaders {
		headers.Set(string(key), value)
	}
	if interceptor != nil {
		rc, err := b.runInterceptor(operation, interceptor)
		if err != nil {
			return nil, err
		}
		if added := rc.Headers(); len(added) > 0 {
			if err := b.profile.validateHeaders(added, operation); err != nil {
				return nil, err
			}
			for key, value := range added {
				headers.Set(string(key), value)
			}
		}
	}
	if b.profile.SDKMetadata != nil {
		headers.Set(MetricsHeader, b.profile.SDKMetadata())
	}
	return headers, nil
}

// runInterceptor calls the caller's interceptor. A panic in it fails the call with a
// SkyflowError instead of crashing the caller, as the Java SDK wraps an interceptor's
// exception; a non-nil *SkyflowError panic value is returned as is. The message is the
// interceptor's own (a panic with an error or a string), else errInterceptorFailed. No HTTP
// status applies, so the code is TRANSPORT_ERROR ("0", empty status).
func (b *Base) runInterceptor(operation string, interceptor Interceptor) (rc *RequestContext, err *skyflowError.SkyflowError) {
	rc = &RequestContext{Operation: operation, BatchIndex: NotBatched, TotalBatches: NotBatched}
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		rc = nil
		if skyflowErr, ok := recovered.(*skyflowError.SkyflowError); ok && skyflowErr != nil {
			err = skyflowErr
			return
		}
		logger.Error(b.profile.logMsg(logInterceptorFailed))
		message := b.profile.errMsg(errInterceptorFailed)
		cause, _ := recovered.(error)
		if cause != nil && cause.Error() != "" {
			message = cause.Error()
		} else if text, isText := recovered.(string); isText && text != "" {
			message = text
		}
		err = b.profile.NewErrorWithCause(skyflowError.TRANSPORT_ERROR, message, cause)
	}()
	interceptor(rc)
	return rc, nil
}

// ValidateCustomHeaders allows only the supported custom headers, each with a value. tag
// names the call in the error message.
func (p Profile) ValidateCustomHeaders(headers map[config.CustomHeaderKey]string, tag string) *skyflowError.SkyflowError {
	for key, value := range headers {
		switch key {
		case config.SkyflowAccountId, config.SkyflowAccountName, config.RequestIdHeader:
		default:
			return p.invalidHeader(errInvalidHeaderKey, tag, string(key))
		}
		if strings.TrimSpace(value) == "" {
			return p.invalidHeader(errEmptyHeaderValue, tag, string(key))
		}
	}
	return nil
}

func (p Profile) validateHeaders(headers map[config.CustomHeaderKey]string, tag string) *skyflowError.SkyflowError {
	if p.ValidateHeaders != nil {
		return p.ValidateHeaders(headers, tag)
	}
	return p.ValidateCustomHeaders(headers, tag)
}

func (p Profile) invalidHeader(format, tag, key string) *skyflowError.SkyflowError {
	return p.NewError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(p.errMsg(format), tag, key))
}
