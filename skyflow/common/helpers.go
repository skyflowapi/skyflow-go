/*
	Copyright (c) 2022 Skyflow, Inc.
*/
package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"

	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	yaml "gopkg.in/yaml.v2"
)

type Data struct {
	Name string `yaml:"name"`
	Version string    `yaml:"version"`
	// Add more fields as needed
}

func AppendRequestId(message string, requestId string) string {
	if requestId == "" {
		return message
	}

	return message + " - requestId : " + requestId
}

func CreateJsonMetadata() (string, error) {
	sdkData, err := ioutil.ReadFile("config.yml")
	if err != nil {
		logger.Debug("failed for reading config Yaml")
	}

	var config Data
	// Unmarshal the YAML data
	err = yaml.Unmarshal(sdkData, &config)
	if err != nil {
		logger.Debug("failed for unmarshalling Yaml data")
	}
	// Create a map to hold the key-value pairs
	data := map[string]string{
			"sdk_name_version": fmt.Sprintf("%s %s",config.Name,config.Version),
			"sdk_client_device_model": string(runtime.GOOS),
			"sdk_client_os_details": fmt.Sprintf("%s %s",runtime.GOOS,runtime.GOARCH),
			"sdk_runtime_details": runtime.Version(),
	}

	// Marshal the map into JSON format
	jsonData, err := json.Marshal(data)
	if err != nil {
			return "", err
	}

	return string(jsonData), nil
}
