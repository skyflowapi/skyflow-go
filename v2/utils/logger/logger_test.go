package logger

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

var _ = Describe("Logger", func() {
	Context("log level functions", func() {
		It("should call Debug without panicking", func() {
			Expect(func() { Debug("debug message") }).NotTo(Panic())
		})
		It("should call Info without panicking", func() {
			Expect(func() { Info("info message") }).NotTo(Panic())
		})
		It("should call Warn without panicking", func() {
			Expect(func() { Warn("warn message") }).NotTo(Panic())
		})
		It("should call Error without panicking", func() {
			Expect(func() { Error("error message") }).NotTo(Panic())
		})
	})

	Context("SetLogLevel", func() {
		It("should set level to INFO", func() {
			Expect(func() { SetLogLevel(INFO) }).NotTo(Panic())
		})
		It("should set level to DEBUG", func() {
			Expect(func() { SetLogLevel(DEBUG) }).NotTo(Panic())
		})
		It("should set level to WARN", func() {
			Expect(func() { SetLogLevel(WARN) }).NotTo(Panic())
		})
		It("should set level to ERROR", func() {
			Expect(func() { SetLogLevel(ERROR) }).NotTo(Panic())
		})
		It("should set level to OFF (discard output)", func() {
			Expect(func() { SetLogLevel(OFF) }).NotTo(Panic())
		})
		It("should use default for unknown level", func() {
			Expect(func() { SetLogLevel(LogLevel(99)) }).NotTo(Panic())
		})
	})
})
