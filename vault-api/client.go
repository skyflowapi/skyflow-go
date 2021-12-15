package vaultapi

import (
	"encoding/json"
	"net/url"

	"github.com/skyflowapi/skyflow-go/errors"
)

type Client struct {
	configuration Configuration
}

var token = ""

func (client *Client) Insert(records map[string]interface{}, options InsertOptions) (responseBody, *errors.SkyflowError) {

	var err = client.isValidVaultDetails()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(records)
	var insertRecord InsertRecord
	if err := json.Unmarshal(jsonRecord, &insertRecord); err != nil {
		panic(err) //to do
	}
	tokenUtils := TokenUtils{token}
	token = tokenUtils.getBearerToken(client.configuration.TokenProvider)
	insertApi := insertApi{client.configuration, insertRecord, options, token}
	return insertApi.post()
}

func (client *Client) Detokenize(records map[string]interface{}) (responseBody, error) {

	var err = client.isValidVaultDetails()
	if err != nil {
		return nil, err
	}

	jsonRecord, _ := json.Marshal(records)
	var detokenizeRecord DetokenizeInput
	if err := json.Unmarshal(jsonRecord, &detokenizeRecord); err != nil {
		panic(err) //to do
	}
	tokenUtils := TokenUtils{token}
	token = tokenUtils.getBearerToken(client.configuration.TokenProvider)
	detokenizeApi := detokenizeApi{client.configuration, detokenizeRecord, token}
	return detokenizeApi.get()
}

func (client *Client) GetById(records map[string]interface{}) (responseBody, error) {

	var err = client.isValidVaultDetails()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(records)
	var getByIdRecord GetByIdInput
	if err := json.Unmarshal(jsonRecord, &getByIdRecord); err != nil {
		panic(err) //to do
	}
	tokenUtils := TokenUtils{token}
	token = tokenUtils.getBearerToken(client.configuration.TokenProvider)
	getByIdApi := getByIdApi{client.configuration, getByIdRecord, token}
	return getByIdApi.get()
}

func (client *Client) InvokeConnection(connectionConfig ConnectionConfig) (responseBody, error) {

	var err = client.isValidVaultDetails()
	if err != nil {
		return nil, err
	}
	tokenUtils := TokenUtils{token}
	token = tokenUtils.getBearerToken(client.configuration.TokenProvider)
	return nil, nil
}

func (client *Client) isValidVaultDetails() *errors.SkyflowError {

	if client.configuration.VaultID == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.EMPTY_VAULT_ID)

	} else if client.configuration.VaultURL == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.EMPTY_VAULT_URL)

	} else if !isValidUrl(client.configuration.VaultURL) {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(DEFAULT), errors.INVALID_VAULT_URL)

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
