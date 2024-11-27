package client

import (
	"context"
	. "skyflow-go/v2/utils/common"
	skyflowError "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
	"skyflow-go/v2/vault/controller"
)

type vaultService struct {
	config     VaultConfig
	logLevel   *logger.LogLevel
	controller controller.VaultController
}

func (v *vaultService) Insert(ctx *context.Context, request *InsertRequest, options *InsertOptions) (*InsertResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Insert(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Detokenize(ctx *context.Context, request DetokenizeRequest, options DetokenizeOptions) (*DetokenizeResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Detokenize(ctx, &request, &options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Get(ctx context.Context, request GetRequest, options GetOptions) (*GetResponse, *skyflowError.SkyflowError) {
	// Get logic here
	res, err := v.controller.Get(&ctx, &request, &options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Delete(ctx context.Context, request DeleteRequest) (*DeleteResponse, *skyflowError.SkyflowError) {
	// Delete logic here
	res, err := v.controller.Delete(&ctx, &request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Query(ctx context.Context, request QueryRequest) (*QueryResponse, *skyflowError.SkyflowError) {
	// Query logic here
	res, err := v.controller.Query(&ctx, &request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Update(ctx context.Context, request UpdateRequest, options UpdateOptions) (*UpdateResponse, error) {
	// Update logic here
	res, err := v.controller.Update(&ctx, &request, &options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Tokenize(ctx context.Context, request []TokenizeRequest) (*TokenizeResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Tokenize(&ctx, &request)
	if err != nil {
		return nil, err
	}
	return res, nil
}
