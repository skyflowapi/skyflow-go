package client

import (
	"context"

	"github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

type detectService struct {
	config     *common.VaultConfig
	logLevel   *logger.LogLevel
	controller *controller.DetectController
}

func (d *detectService) DeidentifyText(ctx context.Context, request common.DeidentifyTextRequest) (*common.DeidentifyTextResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.DeidentifyText(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *detectService) ReidentifyText(ctx context.Context, request common.ReidentifyTextRequest) (*common.ReidentifyTextResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.ReidentifyText(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *detectService) DeidentifyFile(ctx context.Context, request common.DeidentifyFileRequest) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.DeidentifyFile(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *detectService) GetDetectRun(ctx context.Context, request common.GetDetectRunRequest) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.GetDetectRun(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}
