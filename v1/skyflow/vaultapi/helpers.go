/*
Copyright (c) 2022 Skyflow, Inc.
*/
package vaultapi

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/skyflowapi/skyflow-go/v1/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/v1/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/v1/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/v1/skyflow/common"
)

var clientTag = "Client"

func isValidVaultDetails(configuration common.Configuration) *errors.SkyflowError {
	logger.Info(fmt.Sprintf(messages.VALIDATE_INIT_CONFIG, clientTag))
	if configuration.VaultID == "" {
		logger.Error(fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_ID, clientTag))

	} else if configuration.VaultURL == "" {
		logger.Error(fmt.Sprintf(messages.EMPTY_VAULT_URL, clientTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_VAULT_URL, clientTag))

	} else if !isValidUrl(configuration.VaultURL) {
		logger.Error(fmt.Sprintf(messages.INVALID_VAULT_URL, clientTag, configuration.VaultURL))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_VAULT_URL, clientTag, configuration.VaultURL))

	}
	return nil
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Scheme != "https" || u.Host == "" {
		return false
	}

	return true
}

func r_urlEncode(parents []interface{}, pairs map[string]string, data interface{}) map[string]string {

	switch reflect.TypeOf(data).Kind() {
	case reflect.Int:
		pairs[renderKey(parents)] = fmt.Sprintf("%d", data)
	case reflect.Float32:
		pairs[renderKey(parents)] = fmt.Sprintf("%f", data)
	case reflect.Float64:
		pairs[renderKey(parents)] = fmt.Sprintf("%f", data)
	case reflect.Bool:
		pairs[renderKey(parents)] = fmt.Sprintf("%t", data)
	case reflect.Map:
		var mapOfdata = (data).(map[string]interface{})
		for index, value := range mapOfdata {
			parents = append(parents, index)
			r_urlEncode(parents, pairs, value)
			parents = parents[:len(parents)-1]
		}
	default:
		pairs[renderKey(parents)] = fmt.Sprintf("%s", data)
	}
	return pairs
}

func renderKey(parents []interface{}) string {
	var depth = 0
	var outputString = ""
	for index := range parents {
		var typeOfindex = reflect.TypeOf(parents[index]).Kind()
		if depth > 0 || typeOfindex == reflect.Int {
			outputString = outputString + fmt.Sprintf("[%v]", parents[index])
		} else {
			outputString = outputString + (parents[index]).(string)
		}
		depth = depth + 1
	}
	return outputString
}
