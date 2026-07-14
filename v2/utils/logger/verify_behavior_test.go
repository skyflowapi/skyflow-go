package logger

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

// TestBehavior_Verify checks level filtering is unchanged AND that the output
// matches logrus's TextFormatter{FullTimestamp:true} line format exactly:
//
//	time="<RFC3339>" level=<lowercase> msg="<message>"
func TestBehavior_Verify(t *testing.T) {
	calls := []struct {
		name string
		fn   func(...interface{})
	}{{"DEBUG", Debug}, {"INFO", Info}, {"WARN", Warn}, {"ERROR", Error}}
	cases := []struct {
		level   LogLevel
		name    string
		visible map[string]bool
	}{
		{DEBUG, "DEBUG", map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true}},
		{INFO, "INFO", map[string]bool{"DEBUG": false, "INFO": true, "WARN": true, "ERROR": true}},
		{WARN, "WARN", map[string]bool{"DEBUG": false, "INFO": false, "WARN": true, "ERROR": true}},
		{ERROR, "ERROR", map[string]bool{"DEBUG": false, "INFO": false, "WARN": false, "ERROR": true}},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		SetOutput(&buf)
		SetLogLevel(c.level)
		for _, cl := range calls {
			cl.fn(cl.name + " message")
		}
		out := buf.String()
		for _, cl := range calls {
			if strings.Contains(out, cl.name+" message") != c.visible[cl.name] {
				t.Errorf("level %s: %s present=%v want %v", c.name, cl.name, !c.visible[cl.name], c.visible[cl.name])
			}
		}
	}

	// OFF discards everything.
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLogLevel(OFF)
	Error("discarded")
	if buf.Len() != 0 {
		t.Errorf("OFF: expected no output, got %q", buf.String())
	}

	// Exact logrus-style format, including lowercase levels and "warning".
	line := regexp.MustCompile(`^time="[^"]+" level=(\w+) msg="([^"]*)"\n$`)
	checks := []struct {
		fn       func(...interface{})
		level    string
		inputMsg string
	}{
		{Debug, "debug", "hello world"},
		{Info, "info", "hello world"},
		{Warn, "warning", "hello world"},
		{Error, "error", "hello world"},
	}
	for _, ch := range checks {
		var b bytes.Buffer
		SetOutput(&b)
		SetLogLevel(DEBUG)
		ch.fn(ch.inputMsg)
		got := b.String()
		m := line.FindStringSubmatch(got)
		if m == nil {
			t.Errorf("format mismatch: %q", got)
			continue
		}
		if m[1] != ch.level {
			t.Errorf("level: got %q want %q (line %q)", m[1], ch.level, got)
		}
		if m[2] != ch.inputMsg {
			t.Errorf("msg: got %q want %q", m[2], ch.inputMsg)
		}
	}
}
