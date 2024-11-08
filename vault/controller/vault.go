package controller

import (
	"context"
	"fmt"
	"os"
	constants "skyflow-go/internal/constants"
	vaultmethods "skyflow-go/internal/generated/vaultapi"
	"skyflow-go/utils/common"
	"skyflow-go/utils/logger"
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

func createRequestClient(v *VaultController) *vaultmethods.APIClient {
	configuration := vaultmethods.NewConfiguration()
	configuration.AddDefaultHeader("Authorization", "Bearer TOKEN")
	configuration.Servers[0].URL = geURLWithEnv(v.Config.Env, v.Config.ClusterId)
	apiClient := vaultmethods.NewAPIClient(configuration)
	return apiClient
}

func createInsertBodyRequest(request common.InsertRequest, options common.InsertOptions, ctx context.Context) *vaultmethods.RecordServiceInsertRecordBody {
	records := vaultmethods.V1FieldRecords{
		Fields: map[string]interface{}{},
	}
	body := vaultmethods.NewRecordServiceInsertRecordBody()
	body.Records = append(body.Records, records)
	return body
}

func (v *VaultController) Insert(ctx context.Context, request common.InsertRequest, options common.InsertOptions) (*common.InsertResponse, error) {
	fmt.Println("insert cred", v.Config.Credentials, "vaultid", v.Config.VaultId, "loglevel is", *v.Loglevel)
	//insert request body
	fmt.Println("insert cred", v.Config.Credentials, "vaultid", v.Config.VaultId)
	objectName := "consumers"
	apiClient := createRequestClient(v)
	body := createInsertBodyRequest(request, options, ctx)
	resp, r, err := apiClient.RecordsAPI.RecordServiceInsertRecord(context.Background(), v.Config.VaultId, objectName).Body(*body).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RecordsAPI.RecordServiceInsertRecord`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, nil
	}
	fmt.Println("res", resp.Records)
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
