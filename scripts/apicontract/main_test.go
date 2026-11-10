package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func snap(t *testing.T, dir string) string {
	t.Helper()
	var out bytes.Buffer
	if err := run([]string{"snapshot", dir}, &out); err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	return out.String()
}

// A type moved behind an alias, with the same API, gives the same snapshot.
func TestSnapshotIsAliasAware(t *testing.T) {
	old, moved := snap(t, "testdata/old"), snap(t, "testdata/new")
	if old != moved {
		t.Fatalf("alias move changed the snapshot\nold:\n%s\nnew:\n%s", old, moved)
	}
	for _, want := range []string{
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type VaultConfig struct\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type VaultConfig struct, field VaultId string\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, method (VaultConfig) GetVaultId() string\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, method (*VaultConfig) SetVaultId(string)\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, func NewVaultConfig(string, ...string) (*github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault.VaultConfig, error)\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type Validator interface { Validate(github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault.VaultConfig) error }\n",
		`pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, const SdkName untyped string = "skyflow-go"` + "\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type LogLevel string\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type TokenGroupName = string\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, var DefaultVaultConfig github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault.VaultConfig\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type Response[T any] struct\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type Credentials interface { GetToken() string; // has unexported methods }\n",
		"pkg github.com/skyflowapi/skyflow-go/scripts/apicontract/testdata/vault, type Numeric interface { ~int | ~float64 }\n",
	} {
		if !strings.Contains(old, want) {
			t.Errorf("snapshot is missing %q\n%s", want, old)
		}
	}
	if strings.Contains(old, "clusterId") || strings.Contains(old, "internal") {
		t.Errorf("snapshot leaks unexported or internal names:\n%s", old)
	}
}

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCompare(t *testing.T) {
	dir := t.TempDir()
	const (
		getVaultId = "pkg github.com/skyflowapi/skyflow-go/v2/vault, method (VaultConfig) GetVaultId() string\n"
		setVaultId = "pkg github.com/skyflowapi/skyflow-go/v2/vault, method (*VaultConfig) SetVaultId(string)\n"
		getEnv     = "pkg github.com/skyflowapi/skyflow-go/v2/vault, method (VaultConfig) GetEnv() string\n"
	)
	base := write(t, dir, "base.api", getVaultId+setVaultId)
	tests := []struct {
		name, oldPath, newBody, want string
		breaking                     bool
	}{
		{"same", base, setVaultId + getVaultId, "No API changes.", false},
		{"addition", base, getVaultId + setVaultId + getEnv, "Added: 1 line(s)\n+ " + getEnv, false},
		{"removal", base, getVaultId + getEnv, "Breaking: 1 line(s) removed or changed\n- " + setVaultId, true},
		{"new module", filepath.Join(dir, "missing.api"), getVaultId, "New module: 1 API lines", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newPath := write(t, dir, tc.name+".api", tc.newBody)
			var out bytes.Buffer
			err := run([]string{"compare", tc.oldPath, newPath}, &out)
			var b errBreaking
			if got := errors.As(err, &b); got != tc.breaking {
				t.Fatalf("breaking = %v (err %v), want %v", got, err, tc.breaking)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Errorf("output %q does not contain %q", out.String(), tc.want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	if err := run(nil, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "usage") {
		t.Errorf("no args: got %v, want usage", err)
	}
	if err := run([]string{"snapshot", t.TempDir()}, &bytes.Buffer{}); err == nil {
		t.Error("snapshot of a directory without a module: want an error")
	}
	if err := run([]string{"compare", t.TempDir(), "vault.api"}, &bytes.Buffer{}); err == nil {
		t.Error("compare with a directory as the old file: want an error")
	}
	if got := (errBreaking{2}).Error(); got != "2 breaking change(s)" {
		t.Errorf("errBreaking.Error() = %q", got)
	}
}
