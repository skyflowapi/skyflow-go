// Package vault defines the vault operations shared by the Skyflow Go SDKs.
//
// The SDKs (v2 for SkyVault, flowvault for FlowVault) expose the same methods but with
// their own request, options and response types, because the vault APIs differ. Each
// interface here is generic over those types, so both SDKs' vault services satisfy the
// same contract:
//
//	var _ vault.Inserter[common.InsertRequest, common.InsertOptions, common.InsertResponse] = (*vaultService)(nil)
//
// Code that works with either SDK can depend on these interfaces instead of a concrete
// vault service.
package vault

import (
	"context"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
)

// Inserter inserts records into a vault.
type Inserter[Req, Opts, Resp any] interface {
	Insert(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// Updater updates existing records in a vault.
type Updater[Req, Opts, Resp any] interface {
	Update(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// Getter retrieves records from a vault.
type Getter[Req, Opts, Resp any] interface {
	Get(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// Deleter deletes records from a vault.
type Deleter[Req, Opts, Resp any] interface {
	Delete(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// Querier runs a query against a vault.
type Querier[Req, Opts, Resp any] interface {
	Query(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// Detokenizer returns the values behind tokens.
type Detokenizer[Req, Opts, Resp any] interface {
	Detokenize(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// TokenGetter returns existing tokens for values.
type TokenGetter[Req, Opts, Resp any] interface {
	GetTokens(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// FilesUploader uploads files into file columns.
type FilesUploader[Req, Opts, Resp any] interface {
	UploadFiles(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}

// FilesDeleter deletes files from file columns.
type FilesDeleter[Req, Opts, Resp any] interface {
	DeleteFiles(ctx context.Context, request Req, options Opts) (*Resp, *skyflowError.SkyflowError)
}
