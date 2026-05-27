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

// Deprecated: Init is part of skyflow-go v1 which is deprecated and will reach End of Life on October 31, 2026.
// Migrate to v2: https://github.com/skyflowapi/skyflow-go/blob/main/docs/migrate_to_v2.md
func Init(configuration common.Configuration) Client {
	logger.Info(fmt.Sprintf(messages.INITIALIZING_SKYFLOW_CLIENT, clientTag))
	return Client{configuration}
}
