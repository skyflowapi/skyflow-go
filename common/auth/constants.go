package auth

import "time"

const (
	grantType           = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	httpsProtocol       = "https"
	credKeyPrivateKey   = "privateKey"
	privateKeyPEMType   = "PRIVATE KEY"
	sdkIssuer           = "sdk"
	signedTokenPrefix   = "signed_token_"
	ctxKeyRegex         = `^[a-zA-Z0-9_]+$`
	bearerTokenExpiry   = 60 * time.Minute
	defaultDataTokenTTL = 60 * time.Second

	jwtClaimExp = "exp"
	jwtClaimCtx = "ctx"
	jwtClaimIss = "iss"
	jwtClaimAud = "aud"
	jwtClaimKey = "key"
	jwtClaimIat = "iat"
	jwtClaimSub = "sub"
	jwtClaimTok = "tok"
)
