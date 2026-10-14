package controller

import (
	"net/http"
	"os"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvhelpers "github.com/skyflowapi/skyflow-go/flowvault/internal/helpers"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

// BaseController holds shared auth state and HTTP client logic for flowvault vault operations.
// flowvault's VaultController embeds this — equivalent to Java's AbstractBaseVaultController.
//
// Can only use v2's EXPORTED packages (serviceaccount, utils/*).
// v2's internal/* are module-private and cannot be imported from flowvault.
type BaseController struct {
	Config        *fvcommon.VaultConfig
	CommonCreds   *fvcommon.Credentials
	Token         string
	ApiKey        string
	HTTPClient    *http.Client
	BaseURL       string
	AuthHeader    string
	CustomHeaders map[fvcommon.CustomHeaderKey]string
}

// GenerateToken resolves a bearer token from credentials.
// Uses the shared skyflow-go/common/auth implementation.
func GenerateToken(creds fvcommon.Credentials) (*string, *fverrors.SkyflowError) {
	opts := fvcommon.BearerTokenOptions{}
	if creds.Roles != nil {
		opts.RoleIds = creds.Roles
	}
	if creds.Context != nil {
		opts.Ctx = creds.Context
	}
	var token string
	switch {
	case creds.Token != "":
		token = creds.Token
	case creds.ApiKey != "":
		token = creds.ApiKey
	case creds.Path != "":
		resp, err := fvhelpers.Auth().GenerateBearerToken(creds.Path, opts)
		if err != nil {
			return nil, err
		}
		token = resp.AccessToken
	case creds.CredentialsString != "":
		resp, err := fvhelpers.Auth().GenerateBearerTokenFromCreds(creds.CredentialsString, opts)
		if err != nil {
			return nil, err
		}
		token = resp.AccessToken
	default:
		return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "no valid credentials provided")
	}
	return &token, nil
}

// RefreshToken checks token expiry and regenerates if needed.
// fvhelpers.Auth().IsExpired is exported — reused directly.
func (b *BaseController) RefreshToken() *fverrors.SkyflowError {
	creds, err := b.resolveCredentials()
	if err != nil {
		return err
	}
	if b.Token == "" || fvhelpers.Auth().IsExpired(b.Token) {
		fvlogger.Info("generating new bearer token")
		token, err := GenerateToken(*creds)
		if err != nil {
			return err
		}
		b.Token = *token
	} else {
		fvlogger.Info("reusing existing bearer token")
	}
	return nil
}

// SetupClient prepares the authenticated HTTP client for a vault request.
func (b *BaseController) SetupClient(customHeaders map[string]string) *fverrors.SkyflowError {
	// 1. resolve auth token
	if b.Config.Credentials.ApiKey != "" {
		b.ApiKey = b.Config.Credentials.ApiKey
	} else if b.Config.Credentials.Token != "" {
		if fvhelpers.Auth().IsExpired(b.Config.Credentials.Token) {
			return fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "bearer token is expired")
		}
		b.Token = b.Config.Credentials.Token
	} else {
		if err := b.RefreshToken(); err != nil {
			return err
		}
	}

	b.AuthHeader = b.Token
	if b.ApiKey != "" {
		b.AuthHeader = b.ApiKey
	}

	// 2. build base URL
	if b.Config.BaseVaultUrl != "" {
		b.BaseURL = b.Config.BaseVaultUrl
	} else {
		b.BaseURL = fvhelpers.GetURLWithEnv(b.Config.Env, b.Config.ClusterId)
	}

	// 3. build default headers applied to every request
	b.HTTPClient = &http.Client{}
	return nil
}

// NewRequest builds an authenticated *http.Request for a vault API call.
func (b *BaseController) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	req, err := http.NewRequest(method, b.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.AuthHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(fvconsts.SDK_METRICS_HEADER_KEY, fvhelpers.CreateSDKMetadata())
	for k, val := range b.CustomHeaders {
		req.Header.Set(string(k), val)
	}
	return req, nil
}

func (b *BaseController) resolveCredentials() (*fvcommon.Credentials, *fverrors.SkyflowError) {
	if b.Config != nil && !isEmpty(b.Config.Credentials) {
		return &b.Config.Credentials, nil
	}
	if b.CommonCreds != nil && !isEmpty(*b.CommonCreds) {
		return b.CommonCreds, nil
	}
	if envCreds := os.Getenv("SKYFLOW_CREDENTIALS"); envCreds != "" {
		c := fvcommon.Credentials{CredentialsString: envCreds}
		return &c, nil
	}
	return nil, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, "no credentials found")
}

func isEmpty(c fvcommon.Credentials) bool {
	return c.Path == "" && c.CredentialsString == "" && c.Token == "" && c.ApiKey == ""
}
