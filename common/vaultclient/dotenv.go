package vaultclient

import (
	"os"
	"strings"
)

// dotenvFile is read for SKYFLOW_CREDENTIALS when the environment does not set it, as the
// Java and Python SDKs do.
const dotenvFile = ".env"

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

// parseDotenv reads KEY=VALUE lines. It skips blank lines and # comments and accepts an
// "export " prefix. A value can be single-quoted (taken as is), double-quoted (\" and \\
// are unescaped; other escapes such as the \n of a JSON private key are kept), or unquoted
// (a " #" starts a comment). A quoted value may span lines. A later line wins.
func parseDotenv(content string) map[string]string {
	values := map[string]string{}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", newline), newline)
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" || (value[0] != '"' && value[0] != '\'') {
			if comment := strings.Index(value, " #"); comment >= 0 {
				value = strings.TrimSpace(value[:comment])
			}
			values[key] = value
			continue
		}
		quote := value[0]
		// Join the following lines until one ends with the closing quote.
		for !closesQuote(value, quote) && i+1 < len(lines) {
			i++
			value += newline + strings.TrimRight(lines[i], " \t")
		}
		if !closesQuote(value, quote) {
			continue // unterminated: ignore the entry
		}
		value = value[1 : len(value)-1]
		if quote == '"' {
			value = strings.NewReplacer(`\"`, `"`, `\\`, `\`).Replace(value)
		}
		values[key] = value
	}
	return values
}

// closesQuote reports whether value, which starts with quote, also ends with it.
func closesQuote(value string, quote byte) bool {
	return len(value) >= 2 && value[len(value)-1] == quote && (quote == '\'' || !escaped(value, len(value)-1))
}

// escaped reports whether value[i] is preceded by an odd number of backslashes.
func escaped(value string, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && value[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}
