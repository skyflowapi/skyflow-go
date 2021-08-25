package util

import (
	"errors"
	"testing"

	sErrors "github.com/skyflowapi/skyflow-go/errors"
)

func TestGetToken(t *testing.T) {
	_, err := GenerateToken("")
	var apiErr *sErrors.SkyflowError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expect error to be Skyflow error, was not, %v", err)
	}
}
