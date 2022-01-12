package vaultapi

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/skyflowapi/skyflow-go/errors"
)

type Client struct {
	configuration Configuration
}

var tokenUtils TokenUtils

func (client *Client) Insert(records map[string]interface{}, options InsertOptions) (responseBody, *errors.SkyflowError) {

	var err = client.isValidVaultDetails()
	if err != nil {
		return nil, err
	}
	var totalRecords = records["records"]
	if totalRecords == nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.RECORDS_KEY_NOT_FOUND)
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORDS)
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var table = singleRecord["table"]
		var fields = singleRecord["fields"]
		if table == nil {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_TABLE)
		} else if table == "" {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TABLE_NAME)
		} else if fields == nil {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.FIELDS_KEY_ERROR)
		} else if fields == "" {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_FIELDS)
		}
		field := (singleRecord["fields"]).(map[string]interface{})
		if len(field) == 0 {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_FIELDS)
		}
		for index := range field {
			if index == "" {
				return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_COLUMN_NAME)
			}
		}
	}
	jsonRecord, _ := json.Marshal(records)
	var insertRecord InsertRecord
	if err := json.Unmarshal(jsonRecord, &insertRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.INVALID_RECORDS)
	}
	token := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	insertApi := insertApi{client.configuration, insertRecord, options, token}
	return insertApi.post()
}

func (client *Client) Detokenize(records map[string]interface{}) (responseBody, *errors.SkyflowError) {

	var err = client.isValidVaultDetails()
	if err != nil {
		return nil, err
	}
	var totalRecords = records["records"]
	if totalRecords == nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.RECORDS_KEY_NOT_FOUND)
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORDS)
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var token = singleRecord["token"]
		if token == nil {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_TOKEN)
		} else if token == "" {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TOKEN_ID)
		}
	}

	jsonRecord, _ := json.Marshal(records)
	var detokenizeRecord DetokenizeInput
	if err := json.Unmarshal(jsonRecord, &detokenizeRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.INVALID_RECORDS)
	}
	token := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	detokenizeApi := detokenizeApi{client.configuration, detokenizeRecord, token}
	return detokenizeApi.get()
}

func (client *Client) GetById(records map[string]interface{}) (responseBody, *errors.SkyflowError) {
	var err = client.isValidVaultDetails()
	if err != nil {
		return nil, err
	}
	var totalRecords = records["records"]
	if totalRecords == nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.RECORDS_KEY_NOT_FOUND)
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORDS)
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var table = singleRecord["table"]
		var ids = singleRecord["ids"]
		var redaction = singleRecord["redaction"]
		//var redactionInRecord = (redaction).(string)
		if table == nil {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_TABLE)
		} else if table == "" {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TABLE_NAME)
		} else if ids == nil {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_KEY_IDS)
		} else if ids == "" {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_RECORD_IDS)
		} else if redaction == nil {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.MISSING_REDACTION)
		}
		// else if redactionInRecord != RedactionType.PLAIN_TEXT || redactionInRecord != DEFAULT || redactionInRecord != REDACTED || redactionInRecord != MASKED {
		// 	return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.Default), errors.INVALID_REDACTION_TYPE)
		// }
		idArray := (ids).([]interface{})
		if len(idArray) == 0 {
			return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_FIELDS)
		}
		for index := range idArray {
			if idArray[index] == "" {
				return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_TOKEN_ID)
			}
		}
	}
	jsonRecord, _ := json.Marshal(records)
	var getByIdRecord GetByIdInput
	if err := json.Unmarshal(jsonRecord, &getByIdRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.INVALID_RECORDS)
	}
	token := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	getByIdApi := getByIdApi{client.configuration, getByIdRecord, token}
	return getByIdApi.get()
}

func (client *Client) InvokeConnection(connectionConfig ConnectionConfig) (responseBody, *errors.SkyflowError) {

	if connectionConfig.connectionURL == "" {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_CONNECTION_URL)
	} else if !isValidUrl(connectionConfig.connectionURL) {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.INVALID_CONNECTION_URL)
	}
	token := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	invokeConnectionApi := invokeConnectionApi{connectionConfig, token}
	return invokeConnectionApi.post()
}

func (client *Client) isValidVaultDetails() *errors.SkyflowError {

	if client.configuration.VaultID == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_VAULT_ID)

	} else if client.configuration.VaultURL == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_VAULT_URL)

	} else if !isValidUrl(client.configuration.VaultURL) {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(errors.INVALID_VAULT_URL, client.configuration.VaultURL))

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
