package client

import (
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

// VaultService handles all vault operations for a single vault.
//
// It keeps a pointer to its config and a back-reference to the client, so
// UpdateVaultConfig and credential updates on the client are visible to it
// without any copying.
type VaultService struct {
	client *Skyflow
	config *fvcommon.VaultConfig
}

func newVaultService(client *Skyflow, config *fvcommon.VaultConfig) *VaultService {
	return &VaultService{client: client, config: config}
}

// credentials returns the client-level credentials (nil if none are set).
func (v *VaultService) credentials() *fvcommon.Credentials {
	return v.client.credentials
}

// resetAuth drops any cached token so the next call re-authenticates with
// the current credentials. The vault controller is not wired yet; once it is,
// this is where its Token/ApiKey are cleared.
func (v *VaultService) resetAuth() {}

// Insert inserts records into the vault.
func (v *VaultService) Insert(request fvcommon.InsertRequest, options fvcommon.InsertOptions) (*fvcommon.InsertResponse, *fverrors.SkyflowError) {
	fvlogger.Info("Insert triggered for vault: " + v.config.VaultId + ", table: " + request.Table)
	fvlogger.Debug("Insert request: ", request)
	if request.Table == "" {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "table must not be empty")
	}
	if len(request.Values) == 0 {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "values must not be empty")
	}
	// TODO: implement actual HTTP call to Skyflow vault API
	return &fvcommon.InsertResponse{}, nil
}

// Detokenize detokenizes the given tokens.
func (v *VaultService) Detokenize(request fvcommon.DetokenizeRequest, options fvcommon.DetokenizeOptions) (*fvcommon.DetokenizeResponse, *fverrors.SkyflowError) {
	fvlogger.Info("Detokenize triggered for vault: " + v.config.VaultId)
	if len(request.DetokenizeData) == 0 {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "detokenize data must not be empty")
	}
	// TODO: implement actual HTTP call
	return &fvcommon.DetokenizeResponse{}, nil
}

// Get retrieves records from the vault.
func (v *VaultService) Get(request fvcommon.GetRequest, options fvcommon.GetOptions) (*fvcommon.GetResponse, *fverrors.SkyflowError) {
	fvlogger.Info("Get triggered for vault: " + v.config.VaultId + ", table: " + request.Table)
	if request.Table == "" {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "table must not be empty")
	}
	// TODO: implement actual HTTP call
	return &fvcommon.GetResponse{}, nil
}

// Delete deletes records from the vault.
func (v *VaultService) Delete(request fvcommon.DeleteRequest, options fvcommon.DeleteOptions) (*fvcommon.DeleteResponse, *fverrors.SkyflowError) {
	fvlogger.Info("Delete triggered for vault: " + v.config.VaultId + ", table: " + request.Table)
	if request.Table == "" {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "table must not be empty")
	}
	// TODO: implement actual HTTP call
	return &fvcommon.DeleteResponse{}, nil
}

// Update updates records in the vault.
func (v *VaultService) Update(request fvcommon.UpdateRequest, options fvcommon.UpdateOptions) (*fvcommon.UpdateResponse, *fverrors.SkyflowError) {
	fvlogger.Info("Update triggered for vault: " + v.config.VaultId + ", table: " + request.Table)
	if request.Table == "" {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "table must not be empty")
	}
	// TODO: implement actual HTTP call
	return &fvcommon.UpdateResponse{}, nil
}

// Query executes a query against the vault.
func (v *VaultService) Query(request fvcommon.QueryRequest, options fvcommon.QueryOptions) (*fvcommon.QueryResponse, *fverrors.SkyflowError) {
	fvlogger.Info("Query triggered for vault: " + v.config.VaultId)
	if request.Query == "" {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "query must not be empty")
	}
	// TODO: implement actual HTTP call
	return &fvcommon.QueryResponse{}, nil
}

// Tokenize tokenizes values using the vault.
func (v *VaultService) Tokenize(requests []fvcommon.TokenizeRequest, options fvcommon.TokenizeOptions) (*fvcommon.TokenizeResponse, *fverrors.SkyflowError) {
	fvlogger.Info("Tokenize triggered for vault: " + v.config.VaultId)
	if len(requests) == 0 {
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "tokenize requests must not be empty")
	}
	// TODO: implement actual HTTP call
	return &fvcommon.TokenizeResponse{}, nil
}
