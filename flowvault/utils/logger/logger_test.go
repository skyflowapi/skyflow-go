package logger_test

import (
	"bytes"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Logger Suite")
}

var _ = Describe("Logger", func() {
	var out *bytes.Buffer
	BeforeEach(func() {
		out = &bytes.Buffer{}
		SetOutput(out)
		DeferCleanup(func() {
			SetOutput(os.Stderr)
			SetLogLevel(ERROR)
		})
	})

	It("logs every level at DEBUG", func() {
		SetLogLevel(DEBUG)
		Debug("debug message")
		Info("info message")
		Warn("warn message")
		Error("error message")
		Expect(out.String()).To(ContainSubstring(`level=debug msg="debug message"`))
		Expect(out.String()).To(ContainSubstring(`level=info msg="info message"`))
		Expect(out.String()).To(ContainSubstring(`level=warning msg="warn message"`))
		Expect(out.String()).To(ContainSubstring(`level=error msg="error message"`))
	})

	It("logs only errors at ERROR and nothing when OFF", func() {
		SetLogLevel(ERROR)
		Info("hidden")
		Error("shown")
		Expect(out.String()).ToNot(ContainSubstring("hidden"))
		Expect(out.String()).To(ContainSubstring("shown"))

		out.Reset()
		SetLogLevel(OFF)
		Error("silent")
		Expect(out.String()).To(BeEmpty())
	})
})
