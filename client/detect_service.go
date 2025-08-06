package client

import (
	"github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

type detectService struct {
	config     *common.DetectConfig
	logLevel   *logger.LogLevel
	controller *controller.DetectController
}

func (d *detectService) DeidentifyText(request common.DeidentifyTextRequest) (*common.DeidentifyTextResponse, *skyflowError.SkyflowError) {
	res, err := d.controller.DeidentifyText(request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// func (d *detectService) ReidentifyText(request common.ReidentifyTextRequest) (*common.ReidentifyTextResponse, *skyflowError.SkyflowError) {
// }

// func (d *detectService) DeidentifyFile(request common.DeidentifyFileRequest) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
// }

// func (d *detectService) GetDetectRun(request common.GetDetectRunRequest) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
// }








