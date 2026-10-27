package vaultclient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/skyflowapi/skyflow-go/common/config"
	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
)

// MetricsHeader carries the SDK name, version and runtime with every call.
const MetricsHeader = "sky-metadata"

// RequestContext is passed to an options Interceptor before each API call.
type RequestContext struct {
	Operation string
	headers   map[config.CustomHeaderKey]string
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
		rc := &RequestContext{Operation: operation}
		interceptor(rc)
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
	return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(p.errMsg(format), tag, key))
}
