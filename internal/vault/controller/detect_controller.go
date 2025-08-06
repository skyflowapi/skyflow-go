package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

type DetectController struct {
	Config    common.DetectConfig
	Loglevel  *logger.LogLevel
	Token     string
	ApiKey    string
	ApiClient client.Client
}
func (d *DetectController) DeidentifyText(request common.DeidentifyTextRequest) (*common.DeidentifyTextResponse, *skyflowError.SkyflowError) {
}


