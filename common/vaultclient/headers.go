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

// buildHeaders merges the client-level headers, the call's own headers and the SDK metadata
// header, in that order, so a call's header replaces the client's header of the same name.
// The call's headers are validated when set; operation names the call in the error.
func (b *Base) buildHeaders(operation string, clientHeaders, callHeaders map[config.CustomHeaderKey]string) (http.Header, *skyflowError.SkyflowError) {
	if clientHeaders != nil && b.profile.ClientHeadersTag != "" {
		if err := b.profile.validateHeaders(clientHeaders, b.profile.ClientHeadersTag); err != nil {
			return nil, err
		}
	}
	if callHeaders != nil {
		if err := b.profile.validateHeaders(callHeaders, operation); err != nil {
			return nil, err
		}
	}
	headers := http.Header{}
	for key, value := range clientHeaders {
		headers.Set(string(key), value)
	}
	for key, value := range callHeaders {
		headers.Set(string(key), value)
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
	return p.NewError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(p.errMsg(format), tag, key))
}
