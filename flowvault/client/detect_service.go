package client

import (
	"context"

	cdetect "github.com/skyflowapi/skyflow-go/common/detect"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

// DetectService runs the Detect V2 operations for a single vault. The vault ID comes from
// the vault config, and the service shares the vault's credentials, HTTP settings and
// cached bearer token with its VaultService.
//
// Like VaultService, a DetectService obtained earlier sees UpdateVaultConfig and
// credential changes on its next call.
type DetectService struct {
	controller *cdetect.Controller
}

func newDetectService(controller *cdetect.Controller) *DetectService {
	return &DetectService{controller: controller}
}

// DeidentifyString detects and de-identifies sensitive data in a string. Synchronous.
func (d *DetectService) DeidentifyString(ctx context.Context, request fvcommon.DeidentifyStringRequest, options fvcommon.DeidentifyStringOptions) (*fvcommon.DeidentifyStringResponse, *fverrors.SkyflowError) {
	return d.controller.DeidentifyString(ctx, request, options)
}

// DeidentifyFile submits a file for de-identification and returns its run ID. With
// options.Poll set it waits for the run and, for a local file, writes the outputs.
func (d *DetectService) DeidentifyFile(ctx context.Context, request fvcommon.DeidentifyFileRequest, options fvcommon.DeidentifyFileOptions) (*fvcommon.DeidentifyFileResponse, *fverrors.SkyflowError) {
	return d.controller.DeidentifyFile(ctx, request, options)
}

// GetRun returns the current state of a de-identify file run.
func (d *DetectService) GetRun(ctx context.Context, request fvcommon.GetRunRequest, options fvcommon.GetRunOptions) (*fvcommon.GetRunResponse, *fverrors.SkyflowError) {
	return d.controller.GetRun(ctx, request, options)
}

// ReidentifyString replaces tokens in a string with their original or redacted values.
func (d *DetectService) ReidentifyString(ctx context.Context, request fvcommon.ReidentifyStringRequest, options fvcommon.ReidentifyStringOptions) (*fvcommon.ReidentifyStringResponse, *fverrors.SkyflowError) {
	return d.controller.ReidentifyString(ctx, request, options)
}

// ReidentifyFile re-identifies a file. Synchronous; for a local file the outputs are written.
func (d *DetectService) ReidentifyFile(ctx context.Context, request fvcommon.ReidentifyFileRequest, options fvcommon.ReidentifyFileOptions) (*fvcommon.ReidentifyFileResponse, *fverrors.SkyflowError) {
	return d.controller.ReidentifyFile(ctx, request, options)
}

// CheckGuardrails screens text for toxicity and denied topics.
func (d *DetectService) CheckGuardrails(ctx context.Context, request fvcommon.CheckGuardrailsRequest, options fvcommon.CheckGuardrailsOptions) (*fvcommon.CheckGuardrailsResponse, *fverrors.SkyflowError) {
	return d.controller.CheckGuardrails(ctx, request, options)
}
