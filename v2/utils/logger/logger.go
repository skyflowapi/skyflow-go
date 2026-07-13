package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type LogLevel int

const (
	ERROR LogLevel = iota
	INFO
	DEBUG
	WARN
	OFF
)

var (
	// mu guards writer and serializes rebuilds. The hot logging path does not
	// take it; it reads the published logger via log.Load() instead.
	mu       sync.Mutex
	writer   io.Writer = os.Stderr
	levelVar           = new(slog.LevelVar)
	// writeMu serializes writes to the underlying writer so concurrent log
	// calls never produce interleaved lines (logrus locked writes the same way).
	writeMu sync.Mutex
	// log holds the current *slog.Logger. It is swapped atomically so that
	// concurrent log calls never race with SetOutput/SetLogLevel rebuilding it.
	log atomic.Pointer[slog.Logger]
)

func init() {
	levelVar.Set(slog.LevelError)
	rebuild()
}

// rebuild recreates the underlying logger and atomically publishes it. The
// handler is bound to a writer, so the logger must be rebuilt when the output
// changes; the level is held in a *slog.LevelVar, so level changes do not
// require it. Callers that mutate writer must hold mu; init runs before any
// goroutines, so it may call rebuild without the lock.
func rebuild() {
	log.Store(slog.New(&logrusTextHandler{w: writer, level: levelVar}))
}

// logrusTextHandler is a minimal slog.Handler that reproduces the line format
// of logrus's TextFormatter{FullTimestamp: true}, which this package used
// before migrating to log/slog. It emits exactly:
//
//	time="<RFC3339>" level=<lowercase> msg="<message>"
//
// The SDK only ever logs a preformatted message with no structured attributes,
// so WithAttrs/WithGroup are intentionally no-ops.
type logrusTextHandler struct {
	w     io.Writer
	level slog.Leveler
}

func (h *logrusTextHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *logrusTextHandler) Handle(_ context.Context, r slog.Record) error {
	buf := make([]byte, 0, 128)
	buf = append(buf, "time="...)
	buf = appendLogrusValue(buf, r.Time.Format(time.RFC3339))
	buf = append(buf, " level="...)
	buf = append(buf, logrusLevel(r.Level)...)
	buf = append(buf, " msg="...)
	buf = appendLogrusValue(buf, r.Message)
	buf = append(buf, '\n')

	writeMu.Lock()
	defer writeMu.Unlock()
	_, err := h.w.Write(buf)
	return err
}

func (h *logrusTextHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *logrusTextHandler) WithGroup(_ string) slog.Handler      { return h }

// logrusLevel maps slog levels to logrus's lowercase level names. Note logrus
// spells the warn level "warning".
func logrusLevel(l slog.Level) string {
	switch {
	case l < slog.LevelInfo:
		return "debug"
	case l < slog.LevelWarn:
		return "info"
	case l < slog.LevelError:
		return "warning"
	default:
		return "error"
	}
}

// appendLogrusValue appends s, quoting it with %q exactly when logrus's
// TextFormatter would (i.e. when it contains a character outside logrus's
// unquoted set). This is why the timestamp — containing ':' — is quoted.
func appendLogrusValue(b []byte, s string) []byte {
	if logrusNeedsQuoting(s) {
		return append(b, fmt.Sprintf("%q", s)...)
	}
	return append(b, s...)
}

func logrusNeedsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func Debug(args ...interface{}) {
	log.Load().Debug(fmt.Sprint(args...))
}

func Info(args ...interface{}) {
	log.Load().Info(fmt.Sprint(args...))
}

func Warn(args ...interface{}) {
	log.Load().Warn(fmt.Sprint(args...))
}

func Error(args ...interface{}) {
	log.Load().Error(fmt.Sprint(args...))
}

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	writer = w
	rebuild()
}

func SetLogLevel(level LogLevel) {
	switch level {
	case INFO:
		levelVar.Set(slog.LevelInfo)
	case DEBUG:
		levelVar.Set(slog.LevelDebug)
	case WARN:
		levelVar.Set(slog.LevelWarn)
	case ERROR:
		levelVar.Set(slog.LevelError)
	case OFF:
		SetOutput(io.Discard)
	default:
		levelVar.Set(slog.LevelError)
	}
}
