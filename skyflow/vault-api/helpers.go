package vaultapi

import (
	"fmt"
	"net/url"

	"github.com/skyflowapi/skyflow-go/errors"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func isValidVaultDetails(configuration common.Configuration) *errors.SkyflowError {

	if configuration.VaultID == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_VAULT_ID)

	} else if configuration.VaultURL == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_VAULT_URL)

	} else if !isValidUrl(configuration.VaultURL) {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(errors.INVALID_VAULT_URL, configuration.VaultURL))

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
