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
		"pkg example.com/m/a, type Config struct\n",
		"pkg example.com/m/a, type Config struct, field Name string\n",
		"pkg example.com/m/a, method (Config) Get() string\n",
		"pkg example.com/m/a, method (*Config) Set(string)\n",
		"pkg example.com/m/a, func New(string, ...int) (*example.com/m/a.Config, error)\n",
		"pkg example.com/m/a, type Doer interface { Do(example.com/m/a.Config) error }\n",
		`pkg example.com/m/a, const Msg untyped string = "hello"` + "\n",
		"pkg example.com/m/a, type Level string\n",
		"pkg example.com/m/a, type Alias = string\n",
		"pkg example.com/m/a, var Default example.com/m/a.Config\n",
		"pkg example.com/m/a, type Box[T any] struct\n",
		"pkg example.com/m/a, type Sealed interface { Public() int; // has unexported methods }\n",
		"pkg example.com/m/a, type Number interface { ~int | ~float64 }\n",
	} {
		if !strings.Contains(old, want) {
			t.Errorf("snapshot is missing %q\n%s", want, old)
		}
	}
	if strings.Contains(old, "inner") || strings.Contains(old, "internal") {
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
	base := write(t, dir, "base.api", "a\nb\n")
	tests := []struct {
		name, oldPath, newBody, want string
		breaking                     bool
	}{
		{"same", base, "b\na\n", "No API changes.", false},
		{"addition", base, "a\nb\nc\n", "Added: 1 line(s)\n+ c", false},
		{"removal", base, "a\nc\n", "Breaking: 1 line(s) removed or changed\n- b", true},
		{"new module", filepath.Join(dir, "missing.api"), "a\n", "New module: 1 API lines", false},
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
	if err := run([]string{"compare", t.TempDir(), "x"}, &bytes.Buffer{}); err == nil {
		t.Error("compare with a directory as the old file: want an error")
	}
	if got := (errBreaking{2}).Error(); got != "2 breaking change(s)" {
		t.Errorf("errBreaking.Error() = %q", got)
	}
}
