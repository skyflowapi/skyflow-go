package vaultapi

import (
	"net/url"
)

type Client struct {
	configuration Configuration
}

var apiclient ApiClient = ApiClient{}

func (client *Client) insert(records map[string]interface{}, options map[string]interface{}) (responseBody, error) {
	//insert
	var err = client.isValidVaultDetails()
	if err != nil {
		return apiclient.insert(records, options)
	}
	return nil, nil
}

func (client *Client) detokenize(records map[string]interface{}) (responseBody, error) {
	//detokenize
	var err = client.isValidVaultDetails()
	if err != nil {
		return apiclient.detokenize(records)
	}
	return nil, nil
}

func (client *Client) getById(records map[string]interface{}) (responseBody, error) {
	//getById
	var err = client.isValidVaultDetails()
	if err != nil {
		return apiclient.getById(records)
	}
	return nil, nil
}

func (client *Client) invokeConnection(connectionConfig ConnectionConfig) (responseBody, error) {
	//invokeConnection
	var err = client.isValidVaultDetails()
	if err != nil {
		return apiclient.invokeConnection(connectionConfig)
	}
	return nil, nil
}

func (client *Client) isValidVaultDetails() error {

	if client.configuration.vaultID == "" {

	} else if client.configuration.vaultURL == "" {

	} else if !isValidUrl(client.configuration.vaultURL) {

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
