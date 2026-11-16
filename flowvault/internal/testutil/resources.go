// Package testutil loads test fixtures from flowvault/test/resources. Fake credentials live
// there instead of in test source, so secret scanners don't flag the tests.
package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Resource returns the trimmed contents of a file in flowvault/test/resources. It panics if
// the file can't be read, which fails the test that asked for it.
func Resource(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "test", "resources", name)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(content))
}
