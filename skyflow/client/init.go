/*
	Copyright (c) 2022 Skyflow, Inc. 
*/
package client

import (
	"fmt"

	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func Init(configuration common.Configuration) Client {
	logger.Info(fmt.Sprintf(messages.INITIALIZING_SKYFLOW_CLIENT, clientTag))
	return Client{configuration}
}
