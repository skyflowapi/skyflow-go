package client

import (
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func Init(configuration common.Configuration) Client {
	logger.Info(messages.INITIALIZING_SKYFLOW_CLIENT)
	return Client{configuration}
}
