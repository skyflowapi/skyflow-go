/*
Copyright (c) 2022 Skyflow, Inc.
*/
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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

func CreateContextData(ctx  context.Context) map[string]interface{} {
	keys := []string{}
	rv := reflect.ValueOf(ctx)
	RecursiveFunction(rv,&keys)
	data := make(map[string]interface{})
	
	if len(keys) == 0 {
		return data
	}

	for _, key := range keys {
		data[key] = ctx.Value(key)
	}
	
	return data
}

func RecursiveFunction(rv reflect.Value,keys *[]string){
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		for i := 0; i < rv.NumField(); i++ {
			f := rv.Type().Field(i)
			if f.Name == "key" {
				value := fmt.Sprintf("%v", rv.Field(i)) 
				*keys = append(*keys,value)
			}
			if f.Name == "Context" {
				rv := rv.Field(i)
				RecursiveFunction(rv,keys)
			}
		}
	}
	return
}

