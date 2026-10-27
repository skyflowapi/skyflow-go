package detect

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	gen "github.com/skyflowapi/skyflow-go/common/internal/generated"
	apiclient "github.com/skyflowapi/skyflow-go/common/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/common/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/common/internal/generated/option"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
)

// The Detect V2 endpoints are called through the Fern-generated client in
// common/internal/generated. convert.go maps between the public types and the generated ones.

// newAPI builds the generated client for one call. Retries happen in call.HTTPClient, so the
// generated client makes a single attempt.
func newAPI(call *vaultclient.Call) *apiclient.Client {
	return apiclient.NewClient(
		option.WithBaseURL(call.BaseURL),
		option.WithHTTPClient(call.HTTPClient),
		option.WithHTTPHeader(call.Headers),
		option.WithToken(call.AuthToken),
		option.WithMaxAttempts(1),
	)
}

// apiError turns a generated-client error into a SkyflowError. API errors carry the response
// header, so the request ID is kept.
func apiError(err error) *skyflowError.SkyflowError {
	var apiErr *core.APIError
	if errors.As(err, &apiErr) {
		header := apiErr.Header
		if header == nil {
			header = http.Header{}
		}
		return skyflowError.SkyflowErrorApi(apiErr, header)
	}
	return skyflowError.NewSkyflowError(skyflowError.SERVER, err.Error())
}

func (c *Controller) emptyResponse(method string) *skyflowError.SkyflowError {
	return skyflowError.NewSkyflowError(skyflowError.SERVER, fmt.Sprintf(c.errMsg(errEmptyResponse), method))
}

func (c *Controller) callDeidentifyString(ctx context.Context, call *vaultclient.Call, request DeidentifyStringRequest) (*DeidentifyStringResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).StringsV2.DeidentifyStringV2(ctx, &gen.DeidentifyStringRequestV2{
		Text:            request.Text,
		ConfigurationID: optString(request.ConfigurationId),
		Configuration:   toGenConfiguration(request.Configuration, call.VaultID),
	})
	if err != nil {
		return nil, apiError(err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodDeidentifyString)
	}
	return fromGenDeidentifyString(resp), nil
}

func (c *Controller) callDeidentifyFile(ctx context.Context, call *vaultclient.Call, request DeidentifyFileRequest) (string, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).FilesV2.DeidentifyFileV2(ctx, &gen.DeidentifyFileRequestV2{
		DataSource:      gen.DeidentifyFileRequestV2DataSource(request.DataSource),
		Value:           request.Value,
		DataFormat:      optEnum[gen.DeidentifyFileRequestV2DataFormat](request.DataFormat),
		ConfigurationID: optString(request.ConfigurationId),
		Configuration:   toGenConfiguration(request.Configuration, call.VaultID),
	})
	if err != nil {
		return "", apiError(err)
	}
	if resp == nil || resp.RunID == nil {
		return "", c.emptyResponse(methodDeidentifyFile)
	}
	return *resp.RunID, nil
}

func (c *Controller) callGetRun(ctx context.Context, call *vaultclient.Call, runId string) (*GetRunResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).FilesV2.GetRunV2(ctx, &gen.GetRunV2Request{RunID: runId, VaultID: call.VaultID})
	if err != nil {
		return nil, apiError(err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodGetRun)
	}
	return fromGenRun(runId, resp), nil
}

func (c *Controller) callReidentifyString(ctx context.Context, call *vaultclient.Call, request ReidentifyStringRequest) (*ReidentifyStringResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).StringsV2.ReidentifyStringV2(ctx, &gen.ReidentifyStringRequestV2{
		VaultID:        call.VaultID,
		Text:           request.Text,
		RedactionLevel: toGenRedactionLevels(request.RedactionLevel),
	})
	if err != nil {
		return nil, apiError(err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodReidentifyString)
	}
	return &ReidentifyStringResponse{ProcessedText: deref(resp.ProcessedText), Metrics: fromGenMetrics(resp.Metrics)}, nil
}

func (c *Controller) callReidentifyFile(ctx context.Context, call *vaultclient.Call, request ReidentifyFileRequest) (*ReidentifyFileResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).FilesV2.ReidentifyFileV2(ctx, &gen.ReidentifyFileRequestV2{
		VaultID:        call.VaultID,
		DataSource:     gen.ReidentifyFileRequestV2DataSource(request.DataSource),
		Value:          request.Value,
		DataFormat:     optEnum[gen.ReidentifyFileRequestV2DataFormat](request.DataFormat),
		RedactionLevel: toGenRedactionLevels(request.RedactionLevel),
	})
	if err != nil {
		return nil, apiError(err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodReidentifyFile)
	}
	return &ReidentifyFileResponse{
		Status:     DetectRunStatus(derefEnum(resp.Status)),
		OutputType: DataSourceType(derefEnum(resp.OutputType)),
		Output:     fromGenOutputs(resp.Output),
		Metrics:    fromGenMetrics(resp.Metrics),
	}, nil
}

func (c *Controller) callCheckGuardrails(ctx context.Context, call *vaultclient.Call, request CheckGuardrailsRequest) (*CheckGuardrailsResponse, *skyflowError.SkyflowError) {
	resp, err := newAPI(call).GuardrailsV2.CheckGuardrailsV2(ctx, &gen.DetectGuardrailsRequestV2{
		VaultID:       call.VaultID,
		Text:          request.Text,
		CheckToxicity: optBool(request.CheckToxicity),
		DenyTopics:    request.DenyTopics,
	})
	if err != nil {
		return nil, apiError(err)
	}
	if resp == nil {
		return nil, c.emptyResponse(methodCheckGuardrails)
	}
	return &CheckGuardrailsResponse{
		Text:        resp.Text,
		Validation:  GuardrailsValidation(resp.Validation),
		Toxic:       deref(resp.Toxic),
		DeniedTopic: deref(resp.DeniedTopic),
	}, nil
}
