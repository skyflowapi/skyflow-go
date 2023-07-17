/*
Copyright (c) 2022 Skyflow, Inc.
*/
package common

import (
	"encoding/json"
	"fmt"
	"runtime"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
)

// Internal
func AppendRequestId(message string, requestId string) string {
	if requestId == "" {
		return message
	}

	return message + " - requestId : " + requestId
}

// Internal
func CreateJsonMetadata() string {
	// Create a map to hold the key-value pairs
	data := map[string]string{
		"sdk_name_version":        fmt.Sprintf("%s@%s", sdk_name, sdk_version),
		"sdk_client_device_model": string(runtime.GOOS),
		"sdk_client_os_details":   fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		"sdk_runtime_details":     runtime.Version(),
	}

	// Marshal the map into JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Debug("failed for marshalling json data in createJSONMetadata()")
		return ""
	}
	return string(jsonData)
}
