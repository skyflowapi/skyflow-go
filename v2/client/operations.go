package client

import (
	"github.com/skyflowapi/skyflow-go/common/vault"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
)

// Compile-time checks that the v2 vault service implements the operations shared with
// the other Skyflow Go SDKs.
var (
	_ vault.Inserter[common.InsertRequest, common.InsertOptions, common.InsertResponse]                = (*vaultService)(nil)
	_ vault.Updater[common.UpdateRequest, common.UpdateOptions, common.UpdateResponse]                 = (*vaultService)(nil)
	_ vault.Getter[common.GetRequest, common.GetOptions, common.GetResponse]                           = (*vaultService)(nil)
	_ vault.Deleter[common.DeleteRequest, common.DeleteOptions, common.DeleteResponse]                 = (*vaultService)(nil)
	_ vault.Querier[common.QueryRequest, common.QueryOptions, common.QueryResponse]                    = (*vaultService)(nil)
	_ vault.Detokenizer[common.DetokenizeRequest, common.DetokenizeOptions, common.DetokenizeResponse] = (*vaultService)(nil)
)
