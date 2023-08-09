/*
	Copyright (c) 2022 Skyflow, Inc.
*/
package client

import (
	"encoding/json"
	"fmt"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
	vaultapi "github.com/skyflowapi/skyflow-go/skyflow/vaultapi"
)

// Represents a client instance to interact with Skyflow's vault API.
type Client struct {
	configuration common.Configuration
}

var clientTag = "Client"

var tokenUtils TokenUtils

// Inserts data into the vault.
func (client *Client) Insert(records map[string]interface{}, options ...common.InsertOptions) (common.InsertRecords, *errors.SkyflowError) {
	var tempOptions common.InsertOptions
	if len(options) == 0 {
		tempOptions = common.InsertOptions{Tokens: true}
	} else {
		tempOptions = options[0]
	}
	if client.configuration.TokenProvider == nil {
		logger.Error(fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
		return common.InsertRecords{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
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

// Returns values that correspond to the specified tokens.
func (client *Client) Detokenize(records map[string]interface{}) (common.DetokenizeRecords, *errors.SkyflowError) {

	if client.configuration.TokenProvider == nil {
		logger.Error(fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
		return common.DetokenizeRecords{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
	}
	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return common.DetokenizeRecords{}, err
	}
	detokenizeApi := vaultapi.DetokenizeApi{Configuration: client.configuration, Records: records, Token: token}

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

// Reveals records by Skyflow ID.
func (client *Client) GetById(records map[string]interface{}) (common.GetByIdRecords, *errors.SkyflowError) {

	if client.configuration.TokenProvider == nil {
		logger.Error(fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
		return common.GetByIdRecords{}, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
	}
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

// Invokes a connection to an external service using the provided connection configuration.
func (client *Client) InvokeConnection(connectionConfig common.ConnectionConfig) (common.ResponseBody, *errors.SkyflowError) {

	if client.configuration.TokenProvider == nil {
		logger.Error(fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TOKENPROVIDER, clientTag))
	}
	token, err := tokenUtils.getBearerToken(client.configuration.TokenProvider)
	if err != nil {
		return nil, err
	}
	invokeConnectionApi := vaultapi.InvokeConnectionApi{ConnectionConfig: connectionConfig, Token: token}
	return invokeConnectionApi.Post()
}
