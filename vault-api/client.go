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
	jsonRecord, _ := json.Marshal(records)
	var insertRecord InsertRecord
	if err := json.Unmarshal(jsonRecord, &insertRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.Default), errors.INVALID_RECORDS)
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

	jsonRecord, _ := json.Marshal(records)
	var detokenizeRecord DetokenizeInput
	if err := json.Unmarshal(jsonRecord, &detokenizeRecord); err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.INVALID_RECORDS)
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
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.EMPTY_CONNECTION_URL)
	} else if !isValidUrl(connectionConfig.connectionURL) {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.INVALID_CONNECTION_URL)
	}
	token := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	invokeConnectionApi := invokeConnectionApi{connectionConfig, token}
	return invokeConnectionApi.post()
}

func (client *Client) isValidVaultDetails() *errors.SkyflowError {

	if client.configuration.VaultID == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.Default), errors.EMPTY_VAULT_ID)

	} else if client.configuration.VaultURL == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.EMPTY_VAULT_URL)

	} else if !isValidUrl(client.configuration.VaultURL) {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), fmt.Sprintf(errors.INVALID_VAULT_URL, client.configuration.VaultURL))

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
