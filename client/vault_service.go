package client

import (
	"context"

	"github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

type vaultService struct {
	config     *common.VaultConfig
	logLevel   *logger.LogLevel
	controller *controller.VaultController
}

func (v *vaultService) Insert(ctx context.Context, request common.InsertRequest, options common.InsertOptions) (*common.InsertResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Insert(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Detokenize(ctx context.Context, request common.DetokenizeRequest, options common.DetokenizeOptions) (*common.DetokenizeResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Detokenize(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Get(ctx context.Context, request common.GetRequest, options common.GetOptions) (*common.GetResponse, *skyflowError.SkyflowError) {
	// Get logic here
	res, err := v.controller.Get(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Delete(ctx context.Context, request common.DeleteRequest) (*common.DeleteResponse, *skyflowError.SkyflowError) {
	// Delete logic here
	res, err := v.controller.Delete(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Query(ctx context.Context, request common.QueryRequest) (*common.QueryResponse, *skyflowError.SkyflowError) {
	// Query logic here
	res, err := v.controller.Query(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Update(ctx context.Context, request common.UpdateRequest, options common.UpdateOptions) (*common.UpdateResponse, *skyflowError.SkyflowError) {
	// Update logic here
	res, err := v.controller.Update(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Tokenize(ctx context.Context, request []common.TokenizeRequest) (*common.TokenizeResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Tokenize(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}
