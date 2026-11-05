package vaultclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

// unauthorizedRetryWindow is how close to expiry a bearer token the vault rejected with 401
// must be for the SDK to replace it and resend, as in the Java SDK. A 401 for a token with
// longer left was not caused by expiry, so a new token would not help.
const unauthorizedRetryWindow = 60 * time.Second

const (
	headerAuthorization = "Authorization"
	bearerPrefix        = "Bearer "
)

// unauthorizedTransport resends a request once with a new bearer token when the vault
// answers 401 and refresh supplies one. It runs inside retryTransport, so each attempt gets
// the chance, as in the Java SDK.
type unauthorizedTransport struct {
	base http.RoundTripper
	// refresh returns the token to resend with, or "" to return the 401 as it is.
	refresh func(ctx context.Context, rejected string) string
}

func (t *unauthorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}
	if req.Body != nil && req.GetBody == nil {
		return resp, nil // the body can't be sent again
	}
	sent := strings.TrimPrefix(req.Header.Get(headerAuthorization), bearerPrefix)
	fresh := t.refresh(req.Context(), sent)
	if fresh == "" {
		return resp, nil
	}
	resend := req.Clone(req.Context())
	if req.GetBody != nil {
		body, bodyErr := req.GetBody()
		if bodyErr != nil {
			return resp, nil
		}
		resend.Body = body
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	resend.Header.Set(headerAuthorization, bearerPrefix+fresh)
	return t.base.RoundTrip(resend)
}

// refreshAfterUnauthorized is called when the vault rejected the bearer token rejected with
// 401. It returns the token to resend with, or "" when the request must fail as it is:
//   - the credentials are an API key or a static token, which the SDK cannot replace;
//   - rejected is not within unauthorizedRetryWindow of expiry (or has no readable expiry),
//     so the 401 was not caused by expiry;
//   - generating a replacement failed or gave back the rejected token, now or earlier for
//     this same token (a token the SDK caches later clears that record).
//
// Concurrent 401s on one token share a single regeneration: a call that finds the cached
// token already replaced reuses it.
func (b *Base) refreshAfterUnauthorized(ctx context.Context, rejected string) string {
	current := b.snapshot()
	creds, err := b.resolveCredentials(current)
	if err != nil || creds.ApiKey != "" || creds.Token != "" {
		return ""
	}
	if !expiringWithin(rejected, unauthorizedRetryWindow) {
		logger.Info(b.profile.logMsg(logUnauthorizedNotDue))
		return ""
	}
	if ctx.Err() != nil {
		return ""
	}
	select {
	case b.tokenLock <- struct{}{}:
	case <-ctx.Done():
		return ""
	}
	defer func() { <-b.tokenLock }()
	if b.token != "" && b.token != rejected && b.tokenVersion == current.authVersion {
		logger.Info(b.profile.logMsg(logUnauthorizedReusing))
		return b.token
	}
	if rejected == b.refreshFailedFor {
		logger.Info(b.profile.logMsg(logUnauthorizedAlreadyFailed))
		return ""
	}
	logger.Info(b.profile.logMsg(logUnauthorizedRegenerating))
	fresh, genErr := b.profile.GenerateToken(*creds)
	if genErr != nil || fresh == "" || fresh == rejected {
		logger.Warn(b.profile.logMsg(logUnauthorizedRegenerationFailed))
		b.refreshFailedFor = rejected
		return ""
	}
	// A call that started before a credentials change must not replace a newer token.
	if current.authVersion >= b.tokenVersion {
		b.storeToken(fresh, current.authVersion)
	}
	return fresh
}

// storeToken caches a token the SDK obtained and forgets any failed regeneration, so a
// later 401 can be retried again. The caller holds tokenLock.
func (b *Base) storeToken(token string, version uint64) {
	b.token, b.tokenVersion, b.refreshFailedFor = token, version, ""
}

// expiringWithin reports whether a JWT expires within window. A token with no readable
// expiry is not, as in the Java SDK, so its 401 is not retried.
func expiringWithin(token string, window time.Duration) bool {
	claims := jwt.MapClaims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(token, claims); err != nil {
		return false
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}
	return !time.Unix(int64(exp), 0).After(time.Now().Add(window))
}
