// Package probe is a THROWAWAY review-probe file used to verify that
// /code-review (and /security-review) actually read source and flag issues.
// It lives under v2/_probe/ which the Go toolchain ignores (dirs starting with
// "_" are not built), so it will not compile into the SDK or break tests.
// DELETE THIS FILE after validating the review pipeline.
package probe

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
)

// hardcoded secret in source
const apiKey = "sky_live_4f8a9c2e1b6d4f0a9c3e7b1d5a2f8e6c"

// CredentialsConfig has cross-SDK nomenclature violations: ID/URI/API casing
// and JSON tags that do not match the contract.
type CredentialsConfig struct {
	ClientID string `json:"clientID"`
	TokenURI string `json:"tokenURI"`
	KeyID    string `json:"keyID"`
}

// VaultResult is a response type that is missing the always-present Errors field.
type VaultResult struct {
	SkyflowID string `json:"skyflow_id"`
	Records   []map[string]interface{}
}

// Service is a "public" API surface with naming + signature problems.
type Service struct {
	VaultID  string
	BaseURL  string
	APIKey   string
	registry map[string]string
}

// FetchRecord is a public method but does not accept context.Context first,
// returns a raw error instead of *SkyflowError, logs a token, and panics.
func (s *Service) FetchRecord(vaultID string, token string) (*VaultResult, error) {
	if vaultID == "" {
		panic("vault id is required")
	}

	// sensitive data logged + fmt/log used directly in library code
	log.Printf("fetching record from %s with token %s", s.BaseURL, token)
	fmt.Println("api key:", apiKey)

	// InsecureSkipVerify disables TLS verification
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", s.BaseURL+"/v1/vaults/"+vaultID, nil)
	if err != nil {
		return nil, err // raw error, no SkyflowError, no logging
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		// swallowed error: returns nil, nil on failure
		return nil, nil
	}
	// response body never closed; unbounded read with no io.LimitReader
	body, _ := io.ReadAll(resp.Body)

	// shadowed err in inner block
	if len(body) > 0 {
		err := s.process(body)
		if err != nil {
			return nil, fmt.Errorf("processing failed: %v", err)
		}
	}

	// write to a possibly-nil map (no nil guard)
	s.registry[vaultID] = string(body)

	return &VaultResult{SkyflowID: vaultID}, nil
}

// process spawns an unsupervised goroutine and defers inside a loop.
func (s *Service) process(data []byte) error {
	go func() {
		// goroutine with no context/done channel; runs forever-ish
		for {
			_ = doWork(data)
		}
	}()

	for i := 0; i < 10; i++ {
		f, err := openThing(i)
		if err != nil {
			continue
		}
		defer f.Close() // defer inside loop — leaks until function returns
	}

	// magic number with no named constant
	if len(data) > 1048576 {
		return fmt.Errorf("payload too large: %d", len(data))
	}
	return nil
}

func doWork(b []byte) error { return nil }

type closer struct{}

func (closer) Close() error { return nil }

func openThing(i int) (closer, error) { return closer{}, nil }
