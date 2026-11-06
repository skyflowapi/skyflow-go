package client

import (
	"context"

	"github.com/skyflowapi/skyflow-go/flowvault/internal/vault/controller"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// VaultService handles all vault operations for a single vault. Every method is unary:
// one request, one response.
//
// The client updates the controller's settings in place, so a VaultService obtained
// earlier sees UpdateVaultConfig and credential changes on its next call.
type VaultService struct {
	controller *controller.VaultController
}

func newVaultService(controller *controller.VaultController) *VaultService {
	return &VaultService{controller: controller}
}

// Insert inserts records into the vault, optionally upserting on unique columns.
func (v *VaultService) Insert(ctx context.Context, request fvcommon.InsertRequest, options fvcommon.InsertOptions) (*fvcommon.InsertResponse, *fverrors.SkyflowError) {
	return v.controller.Insert(ctx, request, options)
}

// Update updates existing records by Skyflow ID.
func (v *VaultService) Update(ctx context.Context, request fvcommon.UpdateRequest, options fvcommon.UpdateOptions) (*fvcommon.UpdateResponse, *fverrors.SkyflowError) {
	return v.controller.Update(ctx, request, options)
}

// Get retrieves records by Skyflow IDs or unique values.
func (v *VaultService) Get(ctx context.Context, request fvcommon.GetRequest, options fvcommon.GetOptions) (*fvcommon.GetResponse, *fverrors.SkyflowError) {
	return v.controller.Get(ctx, request, options)
}

// Delete deletes records by Skyflow IDs or unique values.
func (v *VaultService) Delete(ctx context.Context, request fvcommon.DeleteRequest, options fvcommon.DeleteOptions) (*fvcommon.DeleteResponse, *fverrors.SkyflowError) {
	return v.controller.Delete(ctx, request, options)
}

// Detokenize returns the values behind tokens.
func (v *VaultService) Detokenize(ctx context.Context, request fvcommon.DetokenizeRequest, options fvcommon.DetokenizeOptions) (*fvcommon.DetokenizeResponse, *fverrors.SkyflowError) {
	return v.controller.Detokenize(ctx, request, options)
}

// GetTokens returns existing tokens for values in deterministic token groups.
func (v *VaultService) GetTokens(ctx context.Context, request fvcommon.GetTokensRequest, options fvcommon.GetTokensOptions) (*fvcommon.GetTokensResponse, *fverrors.SkyflowError) {
	return v.controller.GetTokens(ctx, request, options)
}

// Query runs a SELECT query against the vault.
func (v *VaultService) Query(ctx context.Context, request fvcommon.QueryRequest, options fvcommon.QueryOptions) (*fvcommon.QueryResponse, *fverrors.SkyflowError) {
	return v.controller.Query(ctx, request, options)
}

// UploadFiles uploads files into file columns.
func (v *VaultService) UploadFiles(ctx context.Context, request fvcommon.UploadFilesRequest, options fvcommon.UploadFilesOptions) (*fvcommon.UploadFilesResponse, *fverrors.SkyflowError) {
	return v.controller.UploadFiles(ctx, request, options)
}

// DeleteFiles deletes files from file columns of existing records.
func (v *VaultService) DeleteFiles(ctx context.Context, request fvcommon.DeleteFilesRequest, options fvcommon.DeleteFilesOptions) (*fvcommon.DeleteFilesResponse, *fverrors.SkyflowError) {
	return v.controller.DeleteFiles(ctx, request, options)
}
