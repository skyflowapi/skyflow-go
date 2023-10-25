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

func AppendRequestId(message string, requestId string) string {
	if requestId == "" {
		return message
	}

	return message + " - requestId : " + requestId
}

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

func ConvertToMaps(data interface{}) ([]map[string]interface{}, error) {
	switch data := data.(type) {
	case []map[string]interface{}:
		return data, nil
	case map[string]interface{}:
		return []map[string]interface{}{data}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", data)
	}
}
