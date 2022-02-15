package client

import (
	"encoding/json"
	"fmt"

	vaultapi "github.com/skyflowapi/skyflow-go/skyflow/vault-api"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type Client struct {
	configuration common.Configuration
}

var clientTag = "Client"

var tokenUtils TokenUtils

func (client *Client) Insert(records map[string]interface{}, options ...common.InsertOptions) (common.InsertRecords, *errors.SkyflowError) {
	var tempOptions common.InsertOptions
	if len(options) == 0 {
		tempOptions = common.InsertOptions{Tokens: true}
	} else {
		tempOptions = options[0]
	}
	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return common.InsertRecords{}, err
	}

	insertApi := vaultapi.InsertApi{Configuration: client.configuration, Records: records, Options: tempOptions}

	res, err := insertApi.Post(token)

	if err != nil {
		return common.InsertRecords{}, err
	}

	jsonResponse, _ := json.Marshal(res)
	var response common.InsertRecords
	err1 := json.Unmarshal(jsonResponse, &response)
	if err1 != nil {
		return common.InsertRecords{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, "Insert", err1))
	}
	return response, nil
}

func (client *Client) Detokenize(records map[string]interface{}) (common.DetokenizeRecords, *errors.SkyflowError) {

	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return common.DetokenizeRecords{}, err
	}
	detokenizeApi := vaultapi.DetokenizeApi{Configuration: client.configuration, Records: records, Token: token}

	//return detokenizeApi.Get()
	res, err := detokenizeApi.Get()

	if err != nil {
		return common.DetokenizeRecords{}, err
	}

	jsonResponse, _ := json.Marshal(res)
	var response common.DetokenizeRecords
	err1 := json.Unmarshal(jsonResponse, &response)
	if err1 != nil {
		return common.DetokenizeRecords{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, "Detokenize", err1))
	}
	return response, nil
}

func (client *Client) GetById(records map[string]interface{}) (common.GetByIdRecords, *errors.SkyflowError) {

	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return common.GetByIdRecords{}, err
	}
	getByIdApi := vaultapi.GetByIdApi{Configuration: client.configuration, Records: records, Token: token}

	res, err := getByIdApi.Get()

	if err != nil {
		return common.GetByIdRecords{}, err
	}

	jsonResponse, _ := json.Marshal(res)
	var response common.GetByIdRecords
	err1 := json.Unmarshal(jsonResponse, &response)
	if err1 != nil {
		return common.GetByIdRecords{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.UNKNOWN_ERROR, "GetById", err1))
	}
	return response, nil
}

func (client *Client) InvokeConnection(connectionConfig common.ConnectionConfig) (common.ResponseBody, *errors.SkyflowError) {

	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return nil, err
	}
	invokeConnectionApi := vaultapi.InvokeConnectionApi{ConnectionConfig: connectionConfig, Token: token}
	return invokeConnectionApi.Post()
}
