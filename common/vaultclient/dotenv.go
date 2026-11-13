package vaultclient

import (
	"os"
	"strings"
)

// dotenvFile is read for SKYFLOW_CREDENTIALS when the environment does not set it.
const dotenvFile = ".env"

// commentMarker starts a comment line, or a comment after a value.
const commentMarker = "#"

// newline is the line separator parseDotenv splits on and joins multi-line values with.
const newline = "\n"

// EnvCredentials returns SKYFLOW_CREDENTIALS from the environment or, when the environment
// does not set it, from the .env file in the working directory. It returns "" when neither
// has it. The .env file is only read, never loaded into the environment.
func EnvCredentials() string {
	if value := os.Getenv(CredentialsEnv); value != "" {
		return value
	}
	return dotenvLookup(dotenvFile, CredentialsEnv)
}

// dotenvLookup returns key's value in the .env file at path, or "" when the file can't be
// read or doesn't set it.
func dotenvLookup(path, key string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return parseDotenv(string(data))[key]
}

// byteOrderMark is the UTF-8 BOM some editors put at the start of a file.
const byteOrderMark = "\ufeff"

// parseDotenv reads KEY=VALUE lines. It skips blank lines and # comments and accepts an
// "export" prefix. A value can be single-quoted (taken as is), double-quoted (\" and \\
// are unescaped; other escapes such as the \n of a JSON private key are kept), or unquoted
// (a " #" starts a comment). A quoted value may span lines and ends at its closing quote:
// what follows it on the line, such as a comment, is ignored. A later line wins.
func parseDotenv(content string) map[string]string {
	values := map[string]string{}
	content = strings.TrimPrefix(content, byteOrderMark)
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", newline), newline)
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, commentMarker) {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "export"); ok && rest != "" && (rest[0] == ' ' || rest[0] == '\t') {
			line = strings.TrimSpace(rest)
		}
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || (value[0] != '"' && value[0] != '\'') {
			if comment := strings.Index(value, " "+commentMarker); comment >= 0 {
				value = strings.TrimSpace(value[:comment])
			}
			values[key] = value
			continue
		}
		quote := value[0]
		// Join the following lines until one has the closing quote.
		first := i
		end := closingQuote(value, quote)
		for end < 0 && i+1 < len(lines) {
			i++
			value += newline + strings.TrimRight(lines[i], " \t")
			end = closingQuote(value, quote)
		}
		if end < 0 {
			i = first // unterminated: ignore the entry, but still read the lines after it
			continue
		}
		value = value[1:end]
		if quote == '"' {
			value = strings.NewReplacer(`\"`, `"`, `\\`, `\`).Replace(value)
		}
		values[key] = value
	}
	return values
}

// closingQuote returns the index of the quote that closes value, which starts with quote,
// or -1 when there is none. It is the first quote followed, on its line, by nothing or by a
// comment, so a quote inside the value (the apostrophe of 'it's') doesn't end it. A
// backslash escapes a double quote, not a single one.
func closingQuote(value string, quote byte) int {
	for i := 1; i < len(value); i++ {
		if value[i] != quote || (quote == '"' && escaped(value, i)) {
			continue
		}
		rest, _, _ := strings.Cut(value[i+1:], newline)
		if rest = strings.TrimSpace(rest); rest == "" || strings.HasPrefix(rest, commentMarker) {
			return i
		}
	}
	return -1
}

// escaped reports whether value[i] is preceded by an odd number of backslashes.
func escaped(value string, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && value[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}
