// Package detect implements the Detect APIs (/v2/detect/...) once for the Skyflow Go
// SDKs: the request and response types, validation, the API calls, polling a file run and
// writing its outputs. An SDK creates a Controller per vault on top of its vaultclient.Base
// and re-exports the types from its own detect package, e.g. flowvault/detect.
//
// This package exists to be shared by the SDKs in the skyflow-go repository. It is not part
// of the public SDK API: use the detect package of the SDK you depend on instead. Its API
// may change without notice.
package detect

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
)

// Controller runs the Detect operations for one vault. The vault ID comes from the
// vault config: it is never part of a request.
//
// It is safe for concurrent use; settings changes on the Base apply to the next call.
type Controller struct {
	base *vaultclient.Base
	// sources are the vault type's source sets; base64Source is its de-identify BASE64.
	sources      VaultSources
	base64Source DeidentifyFileSource
	// sleep waits between run lookups; it returns false when ctx ends first. Replaced in tests.
	sleep func(ctx context.Context, d time.Duration) bool
}

// NewController returns the detect controller for the vault that base serves. Share the
// Base with the SDK's vault controller so both reuse one bearer token.
//
// sources are the vault type's source sets (FlowSources or SkySources). Response OutputType
// values are resolved against them; a value not among them is nil.
func NewController(base *vaultclient.Base, sources VaultSources) *Controller {
	sources = VaultSources{Deidentify: slices.Clone(sources.Deidentify), Reidentify: slices.Clone(sources.Reidentify)}
	return &Controller{base: base, sources: sources, base64Source: resolveDeidentifySource(base64Value, sources.Deidentify), sleep: sleepContext}
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (c *Controller) errMsg(base string) string { return c.base.Profile().SDKPrefix + base }
func (c *Controller) logMsg(base string) string { return c.base.Profile().LogPrefix + base }

// invalid is the INVALID_INPUT error for a validation message, which takes no arguments.
func (c *Controller) invalid(text string) *skyflowError.SkyflowError {
	message := c.errMsg(text)
	logger.Error(message)
	return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, message)
}

// prepare logs the call, validates the request and resolves auth, headers and HTTP settings.
func (c *Controller) prepare(ctx context.Context, method, operation string, interceptor Interceptor, validate func() *skyflowError.SkyflowError) (*vaultclient.Call, *skyflowError.SkyflowError) {
	logger.Info(fmt.Sprintf(c.logMsg(logMethodTriggered), method))
	logger.Info(fmt.Sprintf(c.logMsg(logValidatingRequest), method))
	if err := validate(); err != nil {
		return nil, err
	}
	return c.base.PrepareContext(ctx, operation, interceptor)
}

// finish logs how the call ended and passes its error through.
func (c *Controller) finish(method string, err *skyflowError.SkyflowError) *skyflowError.SkyflowError {
	if err != nil {
		logger.Error(fmt.Sprintf(c.logMsg(logRequestRejected), method))
		return err
	}
	logger.Info(fmt.Sprintf(c.logMsg(logRequestResolved), method))
	return nil
}

// DeidentifyString de-identifies sensitive data in a string. Synchronous.
func (c *Controller) DeidentifyString(ctx context.Context, request DeidentifyStringRequest, options DeidentifyStringOptions) (*DeidentifyStringResponse, *skyflowError.SkyflowError) {
	call, err := c.prepare(ctx, methodDeidentifyString, OperationDeidentifyString, options.Interceptor, func() *skyflowError.SkyflowError {
		return c.validateDeidentifyString(request)
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.callDeidentifyString(ctx, call, request)
	if err := c.finish(methodDeidentifyString, err); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeidentifyFile submits a file for de-identification and returns the run ID. With
// options.Poll set it also waits for the run, returning the finished run, or the run ID and
// a non-terminal Status if the wait ran out. A successful polled run of a local file has
// its outputs written to options.OutputDirectory.
//
// If a run lookup or writing the outputs fails, the response is returned together with the
// error, so the run ID (and any outputs) are not lost.
func (c *Controller) DeidentifyFile(ctx context.Context, request DeidentifyFileRequest, options DeidentifyFileOptions) (*DeidentifyFileResponse, *skyflowError.SkyflowError) {
	var input *localFile
	call, err := c.prepare(ctx, methodDeidentifyFile, OperationDeidentifyFile, options.Interceptor, func() *skyflowError.SkyflowError {
		if err := c.validateDeidentifyFile(request, options); err != nil {
			return err
		}
		var readErr *skyflowError.SkyflowError
		input, readErr = c.readLocalFile(request.FilePath, request.FileObject)
		return readErr
	})
	if err != nil {
		return nil, err
	}
	if input != nil {
		request.DataSource, request.Value = c.base64Source, input.base64
		if request.DataFormat == "" {
			request.DataFormat = input.format
		}
	}
	runId, err := c.callDeidentifyFile(ctx, call, request)
	if err := c.finish(methodDeidentifyFile, err); err != nil {
		return nil, err
	}
	resp := &DeidentifyFileResponse{RunId: runId}
	if options.Poll == nil {
		return resp, nil
	}

	// Each lookup refreshes the token, so a wait of up to maxWaitTimeSeconds neither sends an
	// expired token nor keeps one from settings that were changed meanwhile.
	run, err := c.poll(ctx, runId, *options.Poll, func(ctx context.Context) (*GetRunResponse, *skyflowError.SkyflowError) {
		lookup, err := c.base.RefreshAuth(ctx, call)
		if err != nil {
			return nil, err
		}
		return c.callGetRun(ctx, lookup, runId)
	})
	if err != nil {
		return resp, err
	}
	resp.Status, resp.OutputType, resp.Output, resp.Metrics, resp.Message = run.Status, run.OutputType, run.Output, run.Metrics, run.Message
	if input != nil && run.Status == DetectRunStatusSuccess {
		if err := c.writeOutputs(run.Output, run.OutputType, input.name, options.OutputDirectory); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// GetRun returns the current state of a de-identify file run.
func (c *Controller) GetRun(ctx context.Context, request GetRunRequest, options GetRunOptions) (*GetRunResponse, *skyflowError.SkyflowError) {
	call, err := c.prepare(ctx, methodGetRun, OperationGetRun, options.Interceptor, func() *skyflowError.SkyflowError {
		if blank(request.RunId) {
			return c.invalid(errEmptyRunId)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	run, err := c.callGetRun(ctx, call, strings.TrimSpace(request.RunId))
	if err := c.finish(methodGetRun, err); err != nil {
		return nil, err
	}
	return run, nil
}

// ReidentifyString replaces tokens in a string with their original or redacted values.
// Synchronous.
func (c *Controller) ReidentifyString(ctx context.Context, request ReidentifyStringRequest, options ReidentifyStringOptions) (*ReidentifyStringResponse, *skyflowError.SkyflowError) {
	call, err := c.prepare(ctx, methodReidentifyString, OperationReidentifyString, options.Interceptor, func() *skyflowError.SkyflowError {
		if blank(request.Text) {
			return c.invalid(errInvalidTextReidentifyString)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.callReidentifyString(ctx, call, request)
	if err := c.finish(methodReidentifyString, err); err != nil {
		return nil, err
	}
	return resp, nil
}

// ReidentifyFile re-identifies a file. Synchronous: there is no run and no polling. The
// output comes back in the response; the SDK writes nothing to disk.
func (c *Controller) ReidentifyFile(ctx context.Context, request ReidentifyFileRequest, options ReidentifyFileOptions) (*ReidentifyFileResponse, *skyflowError.SkyflowError) {
	call, err := c.prepare(ctx, methodReidentifyFile, OperationReidentifyFile, options.Interceptor, func() *skyflowError.SkyflowError {
		return c.validateReidentifyFile(request)
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.callReidentifyFile(ctx, call, request)
	if err := c.finish(methodReidentifyFile, err); err != nil {
		return nil, err
	}
	return resp, nil
}

// CheckGuardrails screens text for toxicity and denied topics. Synchronous.
func (c *Controller) CheckGuardrails(ctx context.Context, request CheckGuardrailsRequest, options CheckGuardrailsOptions) (*CheckGuardrailsResponse, *skyflowError.SkyflowError) {
	call, err := c.prepare(ctx, methodCheckGuardrails, OperationCheckGuardrails, options.Interceptor, func() *skyflowError.SkyflowError {
		if blank(request.Text) {
			return c.invalid(errInvalidTextCheckGuardrails)
		}
		// CheckToxicity and DenyTopics are optional on the API and are sent as given.
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.callCheckGuardrails(ctx, call, request)
	if err := c.finish(methodCheckGuardrails, err); err != nil {
		return nil, err
	}
	return resp, nil
}
