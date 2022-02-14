package client

import (
	vaultapi "github.com/skyflowapi/skyflow-go/skyflow/vault-api"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type Client struct {
	configuration common.Configuration
}

var tokenUtils TokenUtils

func (client *Client) Insert(records map[string]interface{}, options ...common.InsertOptions) (common.ResponseBody, *errors.SkyflowError) {
	var tempOptions common.InsertOptions
	if len(options) == 0 {
		tempOptions = common.InsertOptions{Tokens: true}
	} else {
		tempOptions = options[0]
	}
	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return nil, err
	}

	insertApi := vaultapi.InsertApi{Configuration: client.configuration, Records: records, Options: tempOptions}
	return insertApi.Post(token)
}

func (client *Client) Detokenize(records map[string]interface{}) (common.ResponseBody, *errors.SkyflowError) {

	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return nil, err
	}
	detokenizeApi := vaultapi.DetokenizeApi{Configuration: client.configuration, Records: records, Token: token}

	return detokenizeApi.Get()
}

func (client *Client) GetById(records map[string]interface{}) (common.ResponseBody, *errors.SkyflowError) {

	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return nil, err
	}
	getByIdApi := vaultapi.GetByIdApi{Configuration: client.configuration, Records: records, Token: token}
	return getByIdApi.Get()
}

func (client *Client) InvokeConnection(connectionConfig common.ConnectionConfig) (common.ResponseBody, *errors.SkyflowError) {

	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return nil, err
	}
	invokeConnectionApi := vaultapi.InvokeConnectionApi{ConnectionConfig: connectionConfig, Token: token}
	return invokeConnectionApi.Post()
}
