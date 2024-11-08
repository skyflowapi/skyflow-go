// Package client
package client

import (
	"context"
	vaultutils "skyflow-go/utils/common"
	"skyflow-go/utils/logger"
	"skyflow-go/vault/controller"
)

type vaultService struct {
	config   vaultutils.VaultConfig
	logLevel *logger.LogLevel
}

func (v *vaultService) Insert(ctx context.Context, request vaultutils.InsertRequest, options vaultutils.InsertOptions) (*vaultutils.InsertResponse, error) {
	vaultController := controller.VaultController{
		Config:   v.config,
		Loglevel: v.logLevel,
	}

	return vaultController.Insert(ctx, request, options)
}

func (v *vaultService) Detokenize(ctx context.Context, request vaultutils.DetokenizeRequest, options vaultutils.DetokenizeOptions) (*vaultutils.DetokenizeResponse, error) {
	vaultController := controller.VaultController{
		Config:   v.config,
		Loglevel: v.logLevel,
	}
	return vaultController.Detokenize(ctx, request, options)
}

func (v *vaultService) Delete(ctx context.Context, request vaultutils.DeleteRequest) (*vaultutils.DeleteResponse, error) {
	// Delete logic here
	return &vaultutils.DeleteResponse{}, nil
}

func (v *vaultService) Update(ctx context.Context, request vaultutils.UpdateRequest, options vaultutils.UpdateOptions) (*vaultutils.UpdateResponse, error) {
	// Update logic here
	return &vaultutils.UpdateResponse{}, nil
}

func (v *vaultService) Get(ctx context.Context, request vaultutils.GetRequest) (*vaultutils.GetResponse, error) {
	// Get logic here
	return &vaultutils.GetResponse{}, nil
}

func (v *vaultService) UploadFile(ctx context.Context, request vaultutils.UploadFileRequest) (*vaultutils.UploadFileResponse, error) {
	// UploadFile logic here
	return &vaultutils.UploadFileResponse{}, nil
}
