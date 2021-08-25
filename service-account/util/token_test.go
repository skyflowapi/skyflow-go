package util

import (
	"testing"
)

func TestGetToken(t *testing.T) {
	_, err := GenerateToken("")
	if err != nil {
		t.Errorf("GetToken() = got %v, want %v", err, nil)
	}
}
