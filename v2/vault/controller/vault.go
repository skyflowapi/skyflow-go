package controller

import (
	"context"
	"fmt"
	"skyflow-go/v2/vault/utils"
)

type VaultService struct {
	Config utils.VaultConfig
}

func (v *VaultService) Insert(ctx context.Context, request utils.InsertRequest, options utils.InsertOptions) (*utils.InsertResponse, error) {
	// Insert logic here, using the vault-specific config
	return &utils.InsertResponse{}, nil
}

func (v *VaultService) Detokenize(ctx context.Context, request utils.DetokenizeRequest, options utils.DetokenizeOptions) (*utils.DetokenizeResponse, error) {
	fmt.Println("Detokenize", v.Config.ID)
	// Detokenize logic here
	return &utils.DetokenizeResponse{Tokens: "detokenized"}, nil
}

func (v *VaultService) Delete(ctx context.Context, request utils.DeleteRequest) (*utils.DeleteResponse, error) {
	// Delete logic here
	return &utils.DeleteResponse{}, nil
}

func (v *VaultService) Update(ctx context.Context, request utils.UpdateRequest, options utils.UpdateOptions) (*utils.UpdateResponse, error) {
	// Update logic here
	return &utils.UpdateResponse{}, nil
}

func (v *VaultService) Get(ctx context.Context, request utils.GetRequest) (*utils.GetResponse, error) {
	// Get logic here
	return &utils.GetResponse{}, nil
}

func (v *VaultService) UploadFile(ctx context.Context, request utils.UploadFileRequest) (*utils.UploadFileResponse, error) {
	// UploadFile logic here
	return &utils.UploadFileResponse{}, nil
}
