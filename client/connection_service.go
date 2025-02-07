package client

import (
	"context"
	"github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

type connectionService struct {
	config     common.ConnectionConfig
	logLevel   *logger.LogLevel
	controller controller.ConnectionController
}

func (c *connectionService) Invoke(ctx context.Context, request common.InvokeConnectionRequest) (*common.InvokeConnectionResponse, *skyflowError.SkyflowError) {
	res, err := c.controller.Invoke(ctx, request)
	if err != nil {
		return nil, err
	}
	return res, nil
}
