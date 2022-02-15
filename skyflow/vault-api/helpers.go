package vaultapi

import (
	"fmt"
	"net/url"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
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
