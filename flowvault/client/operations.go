package client

import (
	"github.com/skyflowapi/skyflow-go/common/vault"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

// Compile-time checks that the flowvault vault service implements the operations shared
// with the other Skyflow Go SDKs.
var (
	_ vault.Inserter[fvcommon.InsertRequest, fvcommon.InsertOptions, fvcommon.InsertResponse]                     = (*VaultService)(nil)
	_ vault.Updater[fvcommon.UpdateRequest, fvcommon.UpdateOptions, fvcommon.UpdateResponse]                      = (*VaultService)(nil)
	_ vault.Getter[fvcommon.GetRequest, fvcommon.GetOptions, fvcommon.GetResponse]                                = (*VaultService)(nil)
	_ vault.Deleter[fvcommon.DeleteRequest, fvcommon.DeleteOptions, fvcommon.DeleteResponse]                      = (*VaultService)(nil)
	_ vault.Querier[fvcommon.QueryRequest, fvcommon.QueryOptions, fvcommon.QueryResponse]                         = (*VaultService)(nil)
	_ vault.Detokenizer[fvcommon.DetokenizeRequest, fvcommon.DetokenizeOptions, fvcommon.DetokenizeResponse]      = (*VaultService)(nil)
	_ vault.TokenGetter[fvcommon.GetTokensRequest, fvcommon.GetTokensOptions, fvcommon.GetTokensResponse]         = (*VaultService)(nil)
	_ vault.FilesUploader[fvcommon.UploadFilesRequest, fvcommon.UploadFilesOptions, fvcommon.UploadFilesResponse] = (*VaultService)(nil)
	_ vault.FilesDeleter[fvcommon.DeleteFilesRequest, fvcommon.DeleteFilesOptions, fvcommon.DeleteFilesResponse]  = (*VaultService)(nil)
)
