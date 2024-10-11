package controller

import (
	"context"
	"fmt"
	"os"
	"skyflow-go/v2/common/logger"
	vaultmethods "skyflow-go/v2/generated"
	"skyflow-go/v2/vault/utils"
)

type VaultService struct {
	Config   utils.VaultConfig
	Loglevel logger.LogLevel
}

func (v *VaultService) Insert(ctx context.Context, request utils.InsertRequest, options utils.InsertOptions) (*utils.InsertResponse, error) {
	fmt.Println("insert cred", v.Config.Credentials, "vaultid", v.Config.VaultId)
	objectName := "consumers"
	configuration := vaultmethods.NewConfiguration()
	configuration.AddDefaultHeader("Authorization", "Bearer TOKEN")
	configuration.Servers[0].URL = "VaultURL"
	apiClient := vaultmethods.NewAPIClient(configuration)
	records := vaultmethods.V1FieldRecords{
		Fields: map[string]interface{}{
			"ssn": "123-12-1234",
		},
	}
	body := vaultmethods.NewRecordServiceInsertRecordBody()
	body.Records = append(body.Records, records)
	resp, r, err := apiClient.RecordsAPI.RecordServiceInsertRecord(context.Background(), v.Config.VaultId, objectName).Body(*body).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RecordsAPI.RecordServiceInsertRecord`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return nil, nil
	}
	fmt.Println("res", resp.Records)
	return &utils.InsertResponse{}, nil
}

func (v *VaultService) Detokenize(ctx context.Context, request utils.DetokenizeRequest, options utils.DetokenizeOptions) (*utils.DetokenizeResponse, error) {
	fmt.Println("Detokenize cred", v.Config.Credentials, "vaultid", v.Config.VaultId)
	fmt.Println("loglevel is", v.Loglevel)
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
