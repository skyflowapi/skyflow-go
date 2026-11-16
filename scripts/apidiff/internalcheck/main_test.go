package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every way an internal type can reach the public API is reported, and only those.
func TestLeakyModule(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"testdata/leaky"}, &out)
	if !errors.Is(err, errExposed) {
		t.Fatalf("run: got %v, want errExposed", err)
	}
	got := out.String()
	for _, want := range []string{
		"example.com/leaky/pub.Config -> example.com/leaky/internal/cfg.Config\n",
		"example.com/leaky/pub.New -> example.com/leaky/internal/cfg.Opt\n",
		"example.com/leaky/pub.Client.Base -> example.com/leaky/internal/cfg.Base\n",
		"example.com/leaky/pub.Client.Do -> example.com/leaky/internal/cfg.Result\n",
		"example.com/leaky/pub.Items -> example.com/leaky/internal/cfg.Item\n",
		"example.com/leaky/pub.Max -> example.com/leaky/internal/cfg.Number\n",
		"example.com/leaky/pub.ByName -> example.com/leaky/internal/cfg.Event\n",
		"example.com/leaky/pub.Stream -> example.com/leaky/internal/cfg.Event\n",
		"example.com/leaky/pub.Recent -> example.com/leaky/internal/cfg.Event\n",
		"example.com/leaky/pub.Handler.Handle -> example.com/leaky/internal/cfg.Event\n",
		"example.com/leaky/pub.Entity -> example.com/leaky/internal/cfg.Named\n",
		"example.com/leaky/pub.Severity -> example.com/leaky/internal/cfg.Level\n",
		"example.com/leaky/pub.Mixed -> example.com/leaky/internal/cfg.Level\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output is missing %q\n%s", want, got)
		}
	}
	for _, notWant := range []string{"Hidden", "helper", "cmd/tool", "Exported", "pub.Config.Deep", "pub.New.Base"} {
		if strings.Contains(got, notWant) {
			t.Errorf("output reports %q, which users can't reach\n%s", notWant, got)
		}
	}
}

// Internal code used in function bodies and constant values is not an exposure.
func TestCleanModule(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"testdata/clean"}, &out); err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Errorf("want no output, got\n%s", out.String())
	}
}

func TestErrors(t *testing.T) {
	if err := run(nil, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "usage") {
		t.Errorf("no args: got %v, want usage", err)
	}
	broken := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(broken, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/broken\n\ngo 1.22\n")
	write("b.go", "package b\n\nvar X int = \"not an int\"\n")
	err := run([]string{broken}, &bytes.Buffer{})
	if err == nil || errors.Is(err, errExposed) || !strings.Contains(err.Error(), "loading") {
		t.Errorf("package with type errors: got %v, want a load error", err)
	}
	if err := run([]string{filepath.Join(t.TempDir(), "missing")}, &bytes.Buffer{}); err == nil {
		t.Error("missing directory: want an error")
	}
}

func TestExitCode(t *testing.T) {
	for _, tc := range []struct {
		name    string
		err     error
		want    int
		wantErr string
	}{
		{"ok", nil, 0, ""},
		{"exposed", errExposed, 1, ""},
		{"tool error", errors.New("boom"), 2, "internalcheck: boom\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if got := exitCode(tc.err, &stderr); got != tc.want {
				t.Errorf("exitCode = %d, want %d", got, tc.want)
			}
			if stderr.String() != tc.wantErr {
				t.Errorf("stderr = %q, want %q", stderr.String(), tc.wantErr)
			}
		})
	}
}

func TestIsInternal(t *testing.T) {
	for path, want := range map[string]bool{
		"example.com/m/internal":       true,
		"example.com/m/internal/cfg":   true,
		"example.com/m/a/internal/b":   true,
		"example.com/m/internals":      false,
		"example.com/m/pub":            false,
		"example.com/m/notinternal/cf": false,
	} {
		if got := isInternal(path); got != want {
			t.Errorf("isInternal(%q) = %v, want %v", path, got, want)
		}
	}
}
