// Package serviceaccount holds the service-account option and response types
// shared by every Skyflow Go SDK. Each SDK re-exports them under its own
// utils/common package through type aliases, so they are the same types.
//
// The types hold only the current API. An SDK that keeps deprecated fields for its
// existing users (the v2 SDK's RoleIDs and LogLevel) declares its own option types and
// converts them to these before calling the shared auth service.
package serviceaccount

type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}

type BearerTokenOptions struct {
	Ctx     interface{}
	RoleIds []string
}

type SignedDataTokensOptions struct {
	DataTokens []string
	TimeToLive int
	Ctx        interface{}
}

type SignedDataTokensResponse struct {
	Token       string
	SignedToken string
}
