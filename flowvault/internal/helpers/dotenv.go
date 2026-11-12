package helpers

import (
	"os"
	"strings"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
)

// dotenvPath is the .env file EnvCredentials reads: the working directory's, as in the Java
// SDK. Tests point it elsewhere.
var dotenvPath = ".env"

// EnvCredentials returns the SKYFLOW_CREDENTIALS value: the environment variable, else the
// same key in the working directory's .env file, as the Java and Python SDKs read it. The file
// is only read, never loaded into the environment. The value is trimmed, as in the Python SDK;
// "" means it is not set.
func EnvCredentials() string {
	if value := strings.TrimSpace(os.Getenv(fvconsts.SKYFLOW_CREDENTIALS_ENV)); value != "" {
		return value
	}
	content, err := os.ReadFile(dotenvPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(dotenvValue(string(content), fvconsts.SKYFLOW_CREDENTIALS_ENV))
}

// dotenvValue returns key's value in .env content, or "" when it isn't there. It reads the
// usual .env forms: KEY=value, an optional "export " prefix, # comments, and values in single
// quotes (taken as they are) or double quotes (with \" and \\ unescaped), which may span
// lines. An unquoted value ends at a " #" comment. The last assignment wins.
func dotenvValue(content, key string) string {
	value, found := "", false
	rest := strings.ReplaceAll(content, "\r\n", "\n")
	for rest != "" {
		var line string
		line, rest = cutLine(rest)
		line = strings.TrimSpace(line)
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		name, raw, isAssignment := strings.Cut(line, "=")
		if !isAssignment || strings.HasPrefix(line, "#") || strings.TrimSpace(name) != key {
			continue
		}
		raw = strings.TrimLeft(raw, " \t")
		if raw != "" && (raw[0] == '"' || raw[0] == '\'') {
			var parsed string
			parsed, rest = quotedValue(raw, rest)
			value, found = parsed, true
			continue
		}
		if comment := strings.Index(raw, " #"); comment >= 0 {
			raw = raw[:comment]
		}
		value, found = strings.TrimSpace(raw), true
	}
	if !found {
		return ""
	}
	return value
}

// cutLine splits off the first line of s.
func cutLine(s string) (string, string) {
	line, rest, _ := strings.Cut(s, "\n")
	return line, rest
}

// quotedValue reads a quoted value that starts at raw (its opening quote) and may continue
// on the following lines in rest. It returns the value and the content after its line.
func quotedValue(raw, rest string) (string, string) {
	quote := raw[0]
	text := raw[1:]
	for {
		if end := closingQuote(text, quote); end >= 0 {
			text = text[:end]
			break
		}
		if rest == "" {
			break // no closing quote: take the rest of the file
		}
		var line string
		line, rest = cutLine(rest)
		text += "\n" + line
	}
	if quote == '"' {
		text = strings.NewReplacer(`\\`, `\`, `\"`, `"`).Replace(text)
	}
	return text, rest
}

// closingQuote is the index of the quote that closes text, or -1. In double quotes a quote
// after a backslash is escaped; single quotes have no escapes.
func closingQuote(text string, quote byte) int {
	for i := 0; i < len(text); i++ {
		switch {
		case quote == '"' && text[i] == '\\':
			i++ // skip the escaped character
		case text[i] == quote:
			return i
		}
	}
	return -1
}
