package client

import (
	"context"

	cdetect "github.com/skyflowapi/skyflow-go/common/detect"
	"github.com/skyflowapi/skyflow-go/common/vaultclient"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

// detectService runs the Detect V2 operations (/v2/detect/...) for a single vault. The vault
// ID comes from the vault config.
//
// The client updates the service's settings in place, so a service obtained earlier sees
// UpdateVaultConfig and credential changes on its next call, and reuses its bearer token
// until it expires.
type detectService struct {
	config     *common.VaultConfig
	logLevel   *logger.LogLevel
	base       *vaultclient.Base
	controller *cdetect.Controller
}

// detectProfile is v2's profile with the checks and messages v2's vault operations use:
// client headers are validated on every call, an expired token gets TOKEN_EXPIRED, and the
// log lines are v2's.
func detectProfile() vaultclient.Profile {
	profile := helpers.Profile()
	profile.ValidateHeaders = validation.ValidateCustomHeaders
	profile.ClientHeadersTag = constants.CLIENT_HEADER_MESSAGE_PREFIX
	profile.ErrTokenExpired = skyflowError.TOKEN_EXPIRED
	profile.Logs = vaultclient.Logs{
		TokenExpired:           logs.BEARER_TOKEN_EXPIRED,
		GenerateBearerToken:    logs.GENERATE_BEARER_TOKEN_TRIGGERED,
		ReuseBearerToken:       logs.REUSE_BEARER_TOKEN,
		UsingApiKey:            logs.USING_API_KEY,
		UsingBearerToken:       logs.USING_BEARER_TOKEN,
		DeprecatedBaseVaultURL: logs.DEPRECATED_FIELD_BASE_VAULT_URL,
	}
	return profile
}

func newDetectService(config common.VaultConfig, settings vaultclient.ClientSettings, logLevel *logger.LogLevel) *detectService {
	base := vaultclient.New(detectProfile(), config, settings)
	return &detectService{config: &config, logLevel: logLevel, base: base, controller: cdetect.NewController(base)}
}

// DeidentifyString detects and de-identifies sensitive data in a string. Synchronous.
func (d *detectService) DeidentifyString(ctx context.Context, request common.DeidentifyStringRequest, options common.DeidentifyStringOptions) (*common.DeidentifyStringResponse, *skyflowError.SkyflowError) {
	return d.controller.DeidentifyString(ctx, request, options)
}

// DeidentifyFile submits a file for de-identification and returns its run ID. With
// options.Poll set it waits for the run and, for a local file, writes the outputs.
func (d *detectService) DeidentifyFile(ctx context.Context, request common.DeidentifyFileRequest, options common.DeidentifyFileOptions) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
	return d.controller.DeidentifyFile(ctx, request, options)
}

// GetRun returns the current state of a de-identify file run.
func (d *detectService) GetRun(ctx context.Context, request common.GetRunRequest, options common.GetRunOptions) (*common.GetRunResponse, *skyflowError.SkyflowError) {
	return d.controller.GetRun(ctx, request, options)
}

// ReidentifyString replaces tokens in a string with their original or redacted values.
func (d *detectService) ReidentifyString(ctx context.Context, request common.ReidentifyStringRequest, options common.ReidentifyStringOptions) (*common.ReidentifyStringResponse, *skyflowError.SkyflowError) {
	return d.controller.ReidentifyString(ctx, request, options)
}

// ReidentifyFile re-identifies a file. Synchronous; for a local file the outputs are written.
func (d *detectService) ReidentifyFile(ctx context.Context, request common.ReidentifyFileRequest, options common.ReidentifyFileOptions) (*common.ReidentifyFileResponse, *skyflowError.SkyflowError) {
	return d.controller.ReidentifyFile(ctx, request, options)
}

// CheckGuardrails screens text for toxicity and denied topics.
func (d *detectService) CheckGuardrails(ctx context.Context, request common.CheckGuardrailsRequest, options common.CheckGuardrailsOptions) (*common.CheckGuardrailsResponse, *skyflowError.SkyflowError) {
	return d.controller.CheckGuardrails(ctx, request, options)
}

// detectSettings are the client-level settings every vault's Detect calls fall back to.
func (s *Skyflow) detectSettings() vaultclient.ClientSettings {
	return vaultclient.ClientSettings{Credentials: s.credentials, Headers: s.customHeaders}
}

// syncDetect pushes client-level settings to every vault. A change of client credentials
// drops the cached bearer tokens.
func (s *Skyflow) syncDetect() {
	settings := s.detectSettings()
	for _, service := range s.detectServices {
		service.base.SetClientSettings(settings)
	}
}

// updateDetect merges the fields UpdateVaultConfig changes into the vault's Detect config.
func (s *Skyflow) updateDetect(service *detectService, updated common.VaultConfig) {
	config := service.base.Config()
	credentialsChanged := !isCredentialsEmpty(updated.Credentials)
	if credentialsChanged {
		config.Credentials = updated.Credentials
	}
	if updated.ClusterId != "" {
		config.ClusterId = updated.ClusterId
	}
	if updated.Env.IsSet() {
		config.Env = updated.Env
	}
	service.base.SetConfig(config, credentialsChanged)
}
