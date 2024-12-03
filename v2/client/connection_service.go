package client

import (
	"context"
	"skyflow-go/v2/internal/vault/controller"
	. "skyflow-go/v2/utils/common"
	skyflowError "skyflow-go/v2/utils/error"
	"skyflow-go/v2/utils/logger"
)

type connectionService struct {
	config     ConnectionConfig
	logLevel   *logger.LogLevel
	controller controller.ConnectionController
}

var connectionTag = "InvokeConnection"

func (c *connectionService) Invoke(ctx context.Context, request InvokeConnectionRequest) (*InvokeConnectionResponse, *skyflowError.SkyflowError) {
	res, err := c.controller.Invoke(&ctx, &request)
	if err != nil {
		return nil, err
	}
	return res, nil
}
