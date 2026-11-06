// Package serviceaccount holds the service-account option and response types
// shared by every Skyflow Go SDK. Each SDK re-exports them under its own
// utils/common package through type aliases, so they are the same types.
package serviceaccount

import "github.com/skyflowapi/skyflow-go/common/logger"

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

type BearerTokenOptions struct {
	Ctx     interface{}
	RoleIds []string
	// Deprecated: Use RoleIds instead.
	RoleIDs  []string
	LogLevel logger.LogLevel
}

type SignedDataTokensOptions struct {
	DataTokens []string
	TimeToLive int
	Ctx        interface{}
	LogLevel   logger.LogLevel
}

type SignedDataTokensResponse struct {
	Token       string
	SignedToken string
}
