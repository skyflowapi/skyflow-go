package controller

import (
	"context"
	"fmt"
	constants "skyflow-go/v2/internal/constants"
	skyflowclient "skyflow-go/v2/internal/generated"
	"skyflow-go/v2/utils/common"
	"skyflow-go/v2/utils/logger"
)

type VaultController struct {
	Config   common.VaultConfig
	Loglevel *logger.LogLevel
}

func geURLWithEnv(env common.Env, clusterId string) string {
	var url = constants.SECURE_PROTOCOL + clusterId
	switch env {
	case common.DEV:
		url = url + constants.DEV_DOMAIN
	case common.PROD:
		url = url + constants.PROD_DOMAIN
	case common.STAGE:
		url = url + constants.STAGE_DOMAIN
	case common.SANDBOX:
		url = url + constants.SANDBOX_DOMAIN
	default:
		url = url + constants.PROD_DOMAIN
	}
	return url
}

func createRequestClient(v *VaultController) *skyflowclient.APIClient {
	configuration := skyflowclient.NewConfiguration()
	configuration.AddDefaultHeader("Authorization", "Bearer TOKEN")
	configuration.Servers[0].URL = geURLWithEnv(v.Config.Env, v.Config.ClusterId)
	apiClient := skyflowclient.NewAPIClient(configuration)
	return apiClient
}

func createInsertBodyRequest(request common.InsertRequest, options common.InsertOptions, ctx context.Context) *skyflowclient.RecordServiceInsertRecordBody {
	records := skyflowclient.V1FieldRecords{
		Fields: map[string]interface{}{},
	}
	body := skyflowclient.NewRecordServiceInsertRecordBody()
	body.Records = append(body.Records, records)
	return body
}

func (v *VaultController) Insert(ctx context.Context, request common.InsertRequest, options common.InsertOptions) (*common.InsertResponse, error) {
	fmt.Println("insert cred", v.Config.Credentials, "vaultid", v.Config.VaultId, "loglevel is", *v.Loglevel)
	//insert request body
	return &common.InsertResponse{}, nil
}

func (v *VaultController) Detokenize(ctx context.Context, request common.DetokenizeRequest, options common.DetokenizeOptions) (*common.DetokenizeResponse, error) {
	fmt.Println("Detokenize cred", v.Config.Credentials, "vaultid", v.Config.VaultId, "loglevel is", *v.Loglevel)
	//detokenize request body
	return &common.DetokenizeResponse{Tokens: "detokenized"}, nil
}

func (v *VaultController) Delete(ctx context.Context, request common.DeleteRequest) (*common.DeleteResponse, error) {
	// Delete logic here
	return &common.DeleteResponse{}, nil
}

func (v *VaultController) Update(ctx context.Context, request common.UpdateRequest, options common.UpdateOptions) (*common.UpdateResponse, error) {
	// Update logic here
	return &common.UpdateResponse{}, nil
}

func (v *VaultController) Get(ctx context.Context, request common.GetRequest) (*common.GetResponse, error) {
	// Get logic here
	return &common.GetResponse{}, nil
}

func (v *VaultController) UploadFile(ctx context.Context, request common.UploadFileRequest) (*common.UploadFileResponse, error) {
	// UploadFile logic here
	return &common.UploadFileResponse{}, nil
}