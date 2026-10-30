package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Suite")
}

// errWriter always fails, used to check that Handle surfaces write errors.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("forced write error") }

var _ = Describe("Logger", func() {
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		SetOutput(buf)
	})

	// The logger is process-global: restore the defaults after every test.
	AfterEach(func() {
		SetOutput(os.Stderr)
		SetLogLevel(ERROR)
	})

	Context("default state", func() {
		It("should log at ERROR level by default", func() {
			Expect(levelVar.Level()).To(Equal(slog.LevelError))
		})
	})

	Context("SetLogLevel", func() {
		DescribeTable("should set the matching slog level",
			func(level LogLevel, expected slog.Level) {
				SetLogLevel(level)
				Expect(levelVar.Level()).To(Equal(expected))
			},
			Entry("INFO", INFO, slog.LevelInfo),
			Entry("DEBUG", DEBUG, slog.LevelDebug),
			Entry("WARN", WARN, slog.LevelWarn),
			Entry("ERROR", ERROR, slog.LevelError),
			Entry("unknown level falls back to ERROR", LogLevel(99), slog.LevelError),
		)

		It("should emit only messages at or above the configured level", func() {
			SetLogLevel(WARN)
			Debug("debug msg")
			Info("info msg")
			Warn("warn msg")
			Error("error msg")
			out := buf.String()
			Expect(out).NotTo(ContainSubstring("debug msg"))
			Expect(out).NotTo(ContainSubstring("info msg"))
			Expect(out).To(ContainSubstring(`level=warning msg="warn msg"`))
			Expect(out).To(ContainSubstring(`level=error msg="error msg"`))
		})

		It("should emit every level when set to DEBUG", func() {
			SetLogLevel(DEBUG)
			Debug("d")
			Info("i")
			Warn("w")
			Error("e")
			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			Expect(lines).To(HaveLen(4))
			Expect(lines[0]).To(HaveSuffix("level=debug msg=d"))
			Expect(lines[1]).To(HaveSuffix("level=info msg=i"))
			Expect(lines[2]).To(HaveSuffix("level=warning msg=w"))
			Expect(lines[3]).To(HaveSuffix("level=error msg=e"))
		})

		It("should discard all output when set to OFF", func() {
			SetLogLevel(OFF)
			Error("should not appear")
			Expect(buf.Len()).To(Equal(0))
		})

		It("should log again to the same writer when a level is set after OFF", func() {
			SetLogLevel(OFF)
			SetLogLevel(INFO)
			Info("visible again")
			Expect(buf.String()).To(ContainSubstring("msg=\"visible again\""))
			Expect(writer).To(Equal(buf))
		})
	})

	Context("line format", func() {
		It("should write a logrus-style line with a quoted RFC3339 timestamp", func() {
			Error("plain")
			line := buf.String()
			Expect(line).To(MatchRegexp(`^time="[^"]+" level=error msg=plain\n$`))
			ts := strings.TrimSuffix(strings.TrimPrefix(strings.SplitN(line, " ", 2)[0], `time="`), `"`)
			_, err := time.Parse(time.RFC3339, ts)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should join multiple args with fmt.Sprint semantics", func() {
			Error("a", 1, 2)
			Expect(buf.String()).To(ContainSubstring(`msg="a1 2"`))
		})
	})

	Context("appendLogrusValue / logrusNeedsQuoting", func() {
		DescribeTable("should quote only when needed",
			func(in, expected string) {
				Expect(string(appendLogrusValue(nil, in))).To(Equal(expected))
			},
			Entry("allowed charset is left bare", "aZ09-._/@^+", "aZ09-._/@^+"),
			Entry("empty string is left bare", "", ""),
			Entry("space forces quoting", "a b", `"a b"`),
			Entry("colon forces quoting", "a:b", `"a:b"`),
			Entry("embedded quote is escaped", `a"b`, `"a\"b"`),
			Entry("newline is escaped", "a\nb", `"a\nb"`),
		)
	})

	Context("logrusLevel", func() {
		DescribeTable("should map slog levels to logrus names",
			func(l slog.Level, expected string) {
				Expect(logrusLevel(l)).To(Equal(expected))
			},
			Entry("below debug", slog.LevelDebug-4, levelDebug),
			Entry("debug", slog.LevelDebug, levelDebug),
			Entry("info", slog.LevelInfo, levelInfo),
			Entry("warn", slog.LevelWarn, levelWarn),
			Entry("error", slog.LevelError, levelError),
			Entry("above error", slog.LevelError+4, levelError),
		)
	})

	Context("logrusTextHandler", func() {
		It("should report Enabled against its level", func() {
			lv := new(slog.LevelVar)
			lv.Set(slog.LevelWarn)
			h := &logrusTextHandler{w: buf, level: lv}
			Expect(h.Enabled(context.Background(), slog.LevelInfo)).To(BeFalse())
			Expect(h.Enabled(context.Background(), slog.LevelWarn)).To(BeTrue())
			Expect(h.Enabled(context.Background(), slog.LevelError)).To(BeTrue())
		})

		It("should return itself from WithAttrs and WithGroup", func() {
			h := &logrusTextHandler{w: buf, level: slog.LevelInfo}
			Expect(h.WithAttrs([]slog.Attr{slog.String("k", "v")})).To(BeIdenticalTo(h))
			Expect(h.WithGroup("g")).To(BeIdenticalTo(h))
		})

		It("should return the writer error from Handle", func() {
			h := &logrusTextHandler{w: errWriter{}, level: slog.LevelInfo}
			rec := slog.NewRecord(time.Now(), slog.LevelError, "msg", 0)
			Expect(h.Handle(context.Background(), rec)).To(MatchError("forced write error"))
		})
	})
})
