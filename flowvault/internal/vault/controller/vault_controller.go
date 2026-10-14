package controller

import (
	"context"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

// VaultController is flowvault's vault controller.
// Embeds BaseController for shared auth/HTTP logic — same as Java's
// VaultController extends AbstractBaseVaultController.
type VaultController struct {
	BaseController
}

func NewVaultController(config fvcommon.VaultConfig, commonCreds *fvcommon.Credentials) *VaultController {
	return &VaultController{
		BaseController: BaseController{
			Config:        &config,
			CommonCreds:   commonCreds,
			CustomHeaders: make(map[fvcommon.CustomHeaderKey]string),
		},
	}
}

// Insert handles flowvault-specific InsertRequest (ContinueOnError, flowvault Upsert).
func (v *VaultController) Insert(ctx context.Context, request fvcommon.InsertRequest, options fvcommon.InsertOptions) (*fvcommon.InsertResponse, *fverrors.SkyflowError) {
	fvlogger.Info("flowvault Insert triggered for vault: " + v.Config.VaultId)
	fvlogger.Debug("Insert request: ", request)
	if request.Table == "" {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "table must not be empty")
	}
	if len(request.Values) == 0 {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "values must not be empty")
	}
	if err := v.SetupClient(options.CustomHeaders); err != nil {
		return nil, err
	}
	// TODO: translate flowvault InsertRequest → REST API call via v.HTTPClient
	return &fvcommon.InsertResponse{}, nil
}

// Get — BaseController handles auth, VaultController translates flowvault request type.
func (v *VaultController) Get(ctx context.Context, request fvcommon.GetRequest, options fvcommon.GetOptions) (*fvcommon.GetResponse, *fverrors.SkyflowError) {
	fvlogger.Info("flowvault Get triggered for vault: " + v.Config.VaultId)
	if err := v.SetupClient(toStringHeaders(options.CustomHeaders)); err != nil {
		return nil, err
	}
	// TODO: call Skyflow REST API via v.HTTPClient
	return &fvcommon.GetResponse{}, nil
}

// Delete — same pattern
func (v *VaultController) Delete(ctx context.Context, request fvcommon.DeleteRequest, options fvcommon.DeleteOptions) (*fvcommon.DeleteResponse, *fverrors.SkyflowError) {
	fvlogger.Info("flowvault Delete triggered for vault: " + v.Config.VaultId)
	if err := v.SetupClient(toStringHeaders(options.CustomHeaders)); err != nil {
		return nil, err
	}
	// TODO: call Skyflow REST API via v.HTTPClient
	return &fvcommon.DeleteResponse{}, nil
}

// toStringHeaders converts map[CustomHeaderKey]string to map[string]string.
// Needed because CustomHeaderKey is a named type, not a raw string.
func toStringHeaders(h map[fvcommon.CustomHeaderKey]string) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		out[string(k)] = v
	}
	return out
}
