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

func (d *detectService) DeidentifyText(ctx context.Context, request common.DeidentifyTextRequest, options common.DeidentifyTextOptions) (*common.DeidentifyTextResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.DeidentifyText(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *detectService) ReidentifyText(ctx context.Context, request common.ReidentifyTextRequest, options common.ReidentifyTextOptions) (*common.ReidentifyTextResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.ReidentifyText(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *detectService) DeidentifyFile(ctx context.Context, request common.DeidentifyFileRequest, options common.DeidentifyFileOptions) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.DeidentifyFile(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *detectService) GetDetectRun(ctx context.Context, request common.GetDetectRunRequest, options common.GetDetectRunOptions) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.GetDetectRun(ctx, request, options)
	if err != nil {
		return nil, err
	}
	return res, nil
}
