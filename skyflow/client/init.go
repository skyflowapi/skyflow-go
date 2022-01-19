package client

import (
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

func Init(configuration common.Configuration) Client {
	return Client{configuration}
}
