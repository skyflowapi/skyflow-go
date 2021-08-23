package token
import (
    "testing"
)

func TestGetToken(t *testing.T) {
	_, err := GetToken("")
    if err != nil {
        t.Errorf("GetToken() = got %v, want %v", err, nil)
    }
}