package logger

import (
	"io"
	"sync"
	"testing"
)

// TestConcurrentLogAndLevelChange guards against the data race that existed
// when `log` was a plain package var: request goroutines calling Info/Error
// read `log` while another goroutine reassigned it via SetLogLevel/SetOutput
// (OFF forces a rebuild). Run with -race to detect regressions:
//
//	go test ./utils/logger/ -race
func TestConcurrentLogAndLevelChange(t *testing.T) {
	SetOutput(io.Discard)
	defer func() {
		// Restore defaults so other specs in the package are unaffected.
		SetOutput(io.Discard)
		SetLogLevel(ERROR)
	}()

	var wg sync.WaitGroup
	// Simulate many in-flight SDK calls emitting logs.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Debug("request debug line")
			Info("request info line")
			Warn("request warn line")
			Error("request error line")
		}()
	}
	// Simulate a shared client's UpdateLogLevel being flipped at runtime.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			SetLogLevel(DEBUG)
			SetLogLevel(OFF) // OFF -> SetOutput -> rebuild() swaps the logger
			SetLogLevel(INFO)
			SetOutput(io.Discard)
		}
	}()
	wg.Wait()
}
