// Package auth is the single implementation of Skyflow service-account
// authentication: bearer-token generation, signed data tokens and token
// expiry checks. skyflow-go/v2 and skyflow-go/flowvault both delegate to it.
//
// This package exists to be shared by the SDKs in the skyflow-go repository.
// It is not part of the public SDK API: use the serviceaccount package of the
// SDK you depend on instead. Its API may change without notice.
package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v4"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	gen "github.com/skyflowapi/skyflow-go/common/internal/generated"
	genauth "github.com/skyflowapi/skyflow-go/common/internal/generated/authentication"
	"github.com/skyflowapi/skyflow-go/common/internal/generated/core"
	"github.com/skyflowapi/skyflow-go/common/internal/generated/option"
	"github.com/skyflowapi/skyflow-go/common/logger"
	sa "github.com/skyflowapi/skyflow-go/common/serviceaccount"
)

// BaseURLResolver derives the auth service base URL from a credentials tokenUri.
type BaseURLResolver func(tokenUri string) (string, *skyflowError.SkyflowError)

// Config identifies the SDK that owns a Service. The prefixes are prepended to
// every error message and log line so each SDK reports its own name/version.
type Config struct {
	// SDKPrefix is prepended to error messages, e.g. "Skyflow Go SDK v2.1.0".
	SDKPrefix string
	// LogPrefix is prepended to log lines, e.g. "[ Skyflow Go SDK v2.1.0 ] ".
	LogPrefix string
	// HTTPClient used for the token endpoint. Nil uses the generated client's default.
	HTTPClient *http.Client
}

// Service performs service-account authentication for one SDK.
type Service struct {
	cfg            Config
	resolveBaseURL BaseURLResolver
}

// New returns a Service for the SDK described by cfg.
func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// SetBaseURLResolver overrides how the token endpoint base URL is derived.
// Passing nil restores the default, GetBaseURL.
func (s *Service) SetBaseURLResolver(r BaseURLResolver) {
	s.resolveBaseURL = r
}

// TokenResult is the raw token endpoint response. Pointer fields distinguish a
// missing field from an empty one, which callers treat as a rejected request.
type TokenResult struct {
	AccessToken *string `json:"accessToken,omitempty"`
	TokenType   *string `json:"tokenType,omitempty"`
}

func (s *Service) errMsg(base string) string { return s.cfg.SDKPrefix + base }
func (s *Service) logMsg(base string) string { return s.cfg.LogPrefix + base }

func (s *Service) baseURLResolver() BaseURLResolver {
	if s.resolveBaseURL != nil {
		return s.resolveBaseURL
	}
	return s.GetBaseURL
}

// ---------------------------------------------------------------------------
// Public SDK operations. Each SDK's serviceaccount package forwards to these.
// ---------------------------------------------------------------------------

// GenerateBearerToken exchanges the credentials file for a bearer token.
func (s *Service) GenerateBearerToken(credentialsFilePath string, options sa.BearerTokenOptions) (*sa.TokenResponse, *skyflowError.SkyflowError) {
	logger.Info(s.logMsg(logGenerateBearerTokenTriggered))
	if credentialsFilePath == "" {
		logger.Error(s.logMsg(logEmptyCredentialsPath))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errEmptyCredentialFilePath))
	}
	credKeys, err := s.ParseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}
	return s.bearerTokenResponse(credKeys, options)
}

// GenerateBearerTokenFromCreds exchanges a credentials JSON string for a bearer token.
func (s *Service) GenerateBearerTokenFromCreds(credentials string, options sa.BearerTokenOptions) (*sa.TokenResponse, *skyflowError.SkyflowError) {
	logger.Info(s.logMsg(logGenerateBearerTokenTriggered))
	var credKeys map[string]interface{}
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		logger.Error(s.logMsg(logInvalidCredentialsStringFormat))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errInvalidCredentials))
	}
	return s.bearerTokenResponse(credKeys, options)
}

func (s *Service) bearerTokenResponse(credKeys map[string]interface{}, options sa.BearerTokenOptions) (*sa.TokenResponse, *skyflowError.SkyflowError) {
	token, err := s.GenerateBearerTokenHelper(credKeys, options)
	if err != nil {
		return nil, err
	}
	if token.AccessToken == nil || token.TokenType == nil {
		logger.Error(s.logMsg(logBearerTokenRejected))
		return nil, skyflowError.NewSkyflowError(skyflowError.SERVER, s.logMsg(logBearerTokenRejected))
	}
	logger.Info(s.logMsg(logGenerateBearerTokenSuccess))
	return &sa.TokenResponse{
		AccessToken: *token.AccessToken,
		TokenType:   *token.TokenType,
	}, nil
}

// GenerateSignedDataTokens signs data tokens with the key in the credentials file.
func (s *Service) GenerateSignedDataTokens(credentialsFilePath string, options sa.SignedDataTokensOptions) ([]sa.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	logger.Info(s.logMsg(logGenerateSignedDataTokensTriggered))
	if credentialsFilePath == "" {
		logger.Error(s.logMsg(logEmptyCredentialsPath))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errEmptyCredentialFilePath))
	}
	if len(options.DataTokens) == 0 {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errEmptyTokensDetokenize))
	}
	credKeys, err := s.ParseCredentialsFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}
	return s.GetSignedDataTokens(credKeys, options)
}

// GenerateSignedDataTokensFromCreds signs data tokens with the key in a credentials JSON string.
func (s *Service) GenerateSignedDataTokensFromCreds(credentials string, options sa.SignedDataTokensOptions) ([]sa.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	if credentials == "" {
		logger.Error(s.logMsg(logInvalidCredentialsStringFormat))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errEmptyCredentialsString))
	}
	logger.Info(s.logMsg(logGenerateSignedDataTokensTriggered))
	var credKeys map[string]interface{}
	if err := json.Unmarshal([]byte(credentials), &credKeys); err != nil {
		logger.Error(s.logMsg(logInvalidCredentialsStringFormat))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errInvalidCredentials))
	}
	return s.GetSignedDataTokens(credKeys, options)
}

// IsExpired reports whether a JWT's exp claim is in the past. The signature is
// not verified; this is a local pre-check before hitting the API.
func (s *Service) IsExpired(tokenString string) bool {
	if tokenString == "" {
		logger.Info(s.logMsg(logEmptyBearerToken))
		return true
	}
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return true
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return true
	}
	var expiryTime time.Time
	switch exp := claims[jwtClaimExp].(type) {
	case float64:
		expiryTime = time.Unix(int64(exp), 0)
	case json.Number:
		v, _ := exp.Int64()
		expiryTime = time.Unix(v, 0)
	}
	expired := expiryTime.Before(time.Now())
	if expired {
		logger.Info(s.logMsg(logBearerTokenExpired))
	}
	return expired
}

// ---------------------------------------------------------------------------
// Building blocks. Exported so v2's internal/helpers can keep delegating to
// them under its existing names.
// ---------------------------------------------------------------------------

// ParseCredentialsFile reads and unmarshals a credentials JSON file.
func (s *Service) ParseCredentialsFile(credentialsFilePath string) (map[string]interface{}, *skyflowError.SkyflowError) {
	file, err := os.Open(credentialsFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf(s.logMsg(logFileNotFound), credentialsFilePath))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(s.errMsg(errFileNotFound), credentialsFilePath))
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		logger.Error(fmt.Sprintf(s.logMsg(logInvalidInputFile), credentialsFilePath))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(s.logMsg(logInvalidInputFile), credentialsFilePath))
	}
	var credKeys map[string]interface{}
	if err := json.Unmarshal(raw, &credKeys); err != nil {
		logger.Error(s.logMsg(logNotAValidJSON))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errCredentialsStringInvalidJSON))
	}
	return credKeys, nil
}

// GetCredentialParams extracts clientId, tokenUri and keyId, accepting the
// legacy clientID/tokenURI/keyID spellings as a fallback.
func (s *Service) GetCredentialParams(credKeys map[string]interface{}) (string, string, string, *skyflowError.SkyflowError) {
	clientId, ok := credKeys["clientId"].(string)
	if oldVal, hasOld := credKeys["clientID"].(string); hasOld && !ok {
		clientId, ok = oldVal, true
	}
	if !ok {
		logger.Error(s.logMsg(logClientIdNotFound))
		return "", "", "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errMissingClientId))
	}

	tokenUri, ok := credKeys["tokenUri"].(string)
	if oldVal, hasOld := credKeys["tokenURI"].(string); hasOld && !ok {
		tokenUri, ok = oldVal, true
	}
	if !ok {
		logger.Error(s.logMsg(logTokenUriNotFound))
		return "", "", "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errMissingTokenUri))
	}

	keyId, ok := credKeys["keyId"].(string)
	if oldVal, hasOld := credKeys["keyID"].(string); hasOld && !ok {
		keyId, ok = oldVal, true
	}
	if !ok {
		logger.Error(s.logMsg(logKeyIdNotFound))
		return "", "", "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errMissingKeyId))
	}
	return clientId, tokenUri, keyId, nil
}

// GetPrivateKey reads and parses the privateKey entry of the credentials.
func (s *Service) GetPrivateKey(credKeys map[string]interface{}) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privateKeyStr, ok := credKeys[credKeyPrivateKey].(string)
	if !ok {
		logger.Error(s.logMsg(logPrivateKeyNotFound))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errMissingPrivateKey))
	}
	return s.ParsePrivateKey(privateKeyStr)
}

// ParsePrivateKey parses a PEM-encoded RSA private key in PKCS1 or PKCS8 form.
func (s *Service) ParsePrivateKey(pemKey string) (*rsa.PrivateKey, *skyflowError.SkyflowError) {
	privPem, _ := pem.Decode([]byte(pemKey))
	if privPem == nil {
		logger.Error(s.logMsg(logJwtInvalidFormat))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errJwtInvalidFormat))
	}
	if privPem.Type != privateKeyPEMType {
		logger.Error(fmt.Sprintf(s.logMsg(logPrivateKeyType), privPem.Type))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errJwtInvalidFormat))
	}
	if key, err := x509.ParsePKCS1PrivateKey(privPem.Bytes); err == nil {
		return key, nil
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(privPem.Bytes)
	if err != nil {
		logger.Error(s.logMsg(logInvalidAlgorithm))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errInvalidAlgorithm))
	}
	if privateKey, ok := parsedKey.(*rsa.PrivateKey); ok {
		return privateKey, nil
	}
	logger.Error(s.logMsg(logInvalidKeySpec))
	return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errInvalidKeySpec))
}

// GetSignedDataTokens signs options.DataTokens with the key in credKeys.
func (s *Service) GetSignedDataTokens(credKeys map[string]interface{}, options sa.SignedDataTokensOptions) ([]sa.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	pvtKey, err := s.GetPrivateKey(credKeys)
	if err != nil {
		return nil, err
	}
	clientId, tokenUri, keyId, err := s.GetCredentialParams(credKeys)
	if err != nil {
		return nil, err
	}
	return s.GenerateSignedDataTokensHelper(clientId, keyId, pvtKey, options, tokenUri)
}

// GenerateSignedDataTokensHelper builds one RS256 JWT per data token.
func (s *Service) GenerateSignedDataTokensHelper(clientId, keyId string, pvtKey *rsa.PrivateKey, options sa.SignedDataTokensOptions, tokenUri string) ([]sa.SignedDataTokensResponse, *skyflowError.SkyflowError) {
	resolvedCtx, ctxErr := s.ValidateAndResolveCtx(options.Ctx)
	if ctxErr != nil {
		return nil, ctxErr
	}
	ttl := defaultDataTokenTTL
	if options.TimeToLive > 0 {
		ttl = time.Duration(options.TimeToLive) * time.Second
	}

	var responseArray []sa.SignedDataTokensResponse
	for _, token := range options.DataTokens {
		claims := jwt.MapClaims{
			jwtClaimIss: sdkIssuer,
			jwtClaimKey: keyId,
			jwtClaimAud: tokenUri,
			jwtClaimIat: time.Now().Unix(),
			jwtClaimSub: clientId,
			jwtClaimTok: token,
			jwtClaimExp: time.Now().Add(ttl).Unix(),
		}
		if resolvedCtx != "" {
			claims[jwtClaimCtx] = resolvedCtx
		}
		tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(pvtKey)
		if err != nil {
			logger.Error(s.logMsg(logParseJwtPayload))
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(s.errMsg(errErrorOccurred)+"%v", err))
		}
		responseArray = append(responseArray, sa.SignedDataTokensResponse{Token: token, SignedToken: signedTokenPrefix + tokenString})
	}
	logger.Info(s.logMsg(logGenerateSignedDataTokenSuccess))
	return responseArray, nil
}

// GetSignedBearerUserToken builds the RS256 assertion sent to the token endpoint.
func (s *Service) GetSignedBearerUserToken(clientId, keyId, tokenUri string, pvtKey *rsa.PrivateKey, options sa.BearerTokenOptions) (string, *skyflowError.SkyflowError) {
	resolvedCtx, ctxErr := s.ValidateAndResolveCtx(options.Ctx)
	if ctxErr != nil {
		return "", ctxErr
	}
	claims := jwt.MapClaims{
		jwtClaimIss: clientId,
		jwtClaimKey: keyId,
		jwtClaimAud: tokenUri,
		jwtClaimSub: clientId,
		jwtClaimExp: time.Now().Add(bearerTokenExpiry).Unix(),
	}
	if resolvedCtx != "" {
		claims[jwtClaimCtx] = resolvedCtx
	}
	signedToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(pvtKey)
	if err != nil {
		logger.Error(s.logMsg(logParseJwtPayload))
		return "", skyflowError.NewSkyflowError(skyflowError.SERVER, fmt.Sprintf(s.errMsg(errUnknownError), err))
	}
	return signedToken, nil
}

// GenerateBearerTokenHelper signs the assertion and calls the token endpoint.
// It returns the raw response; callers decide how to treat a missing token.
func (s *Service) GenerateBearerTokenHelper(credKeys map[string]interface{}, options sa.BearerTokenOptions) (*TokenResult, *skyflowError.SkyflowError) {
	privateKey, ok := credKeys[credKeyPrivateKey].(string)
	if !ok {
		logger.Error(s.logMsg(logPrivateKeyNotFound))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errMissingPrivateKey))
	}
	pvtKey, err := s.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	clientId, tokenUri, keyId, err := s.GetCredentialParams(credKeys)
	if err != nil {
		return nil, err
	}
	signedUserJWT, err := s.GetSignedBearerUserToken(clientId, keyId, tokenUri, pvtKey, options)
	if err != nil {
		return nil, err
	}
	baseURL, err := s.baseURLResolver()(tokenUri)
	if err != nil {
		return nil, err
	}

	body := &gen.GetAuthTokenRequest{
		GrantType: grantType,
		Assertion: signedUserJWT,
	}
	roleIds := options.RoleIds
	if len(options.RoleIDs) > 0 {
		logger.Warn(s.logMsg(logDeprecatedFieldRoleIds))
		if len(roleIds) == 0 {
			roleIds = options.RoleIDs
		}
	}
	if len(roleIds) > 0 {
		scope := GetScopeUsingRoles(roleIds)
		body.Scope = &scope
	}
	return s.postAuthToken(baseURL, body)
}

// postAuthToken calls the token endpoint through the generated authentication
// client (common/internal/generated), which owns the endpoint path, retries
// and error decoding.
func (s *Service) postAuthToken(baseURL string, body *gen.GetAuthTokenRequest) (*TokenResult, *skyflowError.SkyflowError) {
	opts := []option.RequestOption{option.WithBaseURL(baseURL)}
	if s.cfg.HTTPClient != nil {
		opts = append(opts, option.WithHTTPClient(s.cfg.HTTPClient))
	}
	client := genauth.NewClient(opts...)
	resp, err := client.WithRawResponse.GetBearerToken(context.Background(), body)
	if err != nil {
		return nil, skyflowError.SkyflowErrorApi(err, responseHeader(err))
	}
	return &TokenResult{AccessToken: resp.Body.AccessToken, TokenType: resp.Body.TokenType}, nil
}

// responseHeader extracts the HTTP response header carried by a generated
// client error, so the request ID reaches the SkyflowError.
func responseHeader(err error) http.Header {
	var apiError *core.APIError
	if errors.As(err, &apiError) {
		return apiError.Header
	}
	return http.Header{}
}

// GetScopeUsingRoles formats role IDs as the scope string the token endpoint expects.
func GetScopeUsingRoles(roles []string) string {
	scope := ""
	for _, role := range roles {
		scope += " role:" + role
	}
	return scope
}

// GetBaseURL reduces a tokenUri to scheme://host, requiring https.
func (s *Service) GetBaseURL(urlStr string) (string, *skyflowError.SkyflowError) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil || parsedUrl.Scheme != httpsProtocol || parsedUrl.Host == "" {
		logger.Error(s.logMsg(logInvalidTokenUri))
		return "", skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errInvalidTokenUri))
	}
	return fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host), nil
}

var ctxKeyPattern = regexp.MustCompile(ctxKeyRegex)

// ValidateAndResolveCtx normalises the ctx option into a JWT claim value.
func (s *Service) ValidateAndResolveCtx(ctx interface{}) (interface{}, *skyflowError.SkyflowError) {
	if ctx == nil {
		return nil, nil
	}
	switch v := ctx.(type) {
	case string:
		if v == "" {
			return nil, nil
		}
		return v, nil
	case bool:
		return v, nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case map[string]interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		for key := range v {
			if !ctxKeyPattern.MatchString(key) {
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(s.errMsg(errInvalidCtxMapKey), key))
			}
		}
		return v, nil
	default:
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, s.errMsg(errInvalidCtxType))
	}
}
