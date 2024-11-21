// Package client
package client

import (
	"context"
	vaultutils "skyflow-go/v2/utils/common"
	skyflowError "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
	"skyflow-go/v2/vault/controller"
)

type vaultService struct {
	config     vaultutils.VaultConfig
	logLevel   *logger.LogLevel
	controller controller.VaultController
}

func (v *vaultService) Insert(ctx *context.Context, request *vaultutils.InsertRequest, options *vaultutils.InsertOptions) (*vaultutils.InsertResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Insert(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Detokenize(ctx *context.Context, request vaultutils.DetokenizeRequest, options vaultutils.DetokenizeOptions) (*vaultutils.DetokenizeResponse, *skyflowError.SkyflowError) {
	res, err := v.controller.Detokenize(ctx, &request, &options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Get(ctx context.Context, request vaultutils.GetRequest, options vaultutils.GetOptions) (*vaultutils.GetResponse, error) {
	// Get logic here
	res, err := v.controller.Get(&ctx, &request, &options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *vaultService) Delete(ctx context.Context, request vaultutils.DeleteRequest) (*vaultutils.DeleteResponse, error) {
	// Delete logic here
	return &vaultutils.DeleteResponse{}, nil
}

func (v *vaultService) Update(ctx context.Context, request vaultutils.UpdateRequest, options vaultutils.UpdateOptions) (*vaultutils.UpdateResponse, error) {
	// Update logic here
	return &vaultutils.UpdateResponse{}, nil
}

func (v *vaultService) UploadFile(ctx context.Context, request vaultutils.UploadFileRequest) (*vaultutils.UploadFileResponse, error) {
	// UploadFile logic here
	return &vaultutils.UploadFileResponse{}, nil
}
