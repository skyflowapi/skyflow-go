package vaultclient

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDotenv(t *testing.T) {
	content := "# comment\r\n" +
		"\n" +
		"PLAIN=value # trailing comment\n" +
		"export EXPORTED=yes\n" +
		"SINGLE='{\"key\": \"a\\nb\"} # kept'\n" +
		"DOUBLE=\"{\\\"key\\\": \\\"a\\nb\\\"}\"\n" +
		"MULTI='{\n  \"clientID\": \"id\"\n}'\n" +
		"EMPTY=\n" +
		"NOEQUALS\n" +
		"=novalue\n" +
		"HASH=a#b\n" +
		"ESCAPED_END=\"ends with \\\\\"\n" +
		"LATER=first\n" +
		"LATER=second\n" +
		"COMMENTED=\"quoted\" # a \"comment\"\n" +
		"APOSTROPHE='it's' # note\n" +
		"export\tTABBED='yes'\n" +
		"exported=no\n" +
		"UNTERMINATED='never closed\n" +
		"AFTER=kept\n"
	got := parseDotenv(content)
	want := map[string]string{
		"PLAIN":       "value",
		"EXPORTED":    "yes",
		"SINGLE":      `{"key": "a\nb"} # kept`,
		"DOUBLE":      `{"key": "a\nb"}`,
		"MULTI":       "{\n  \"clientID\": \"id\"\n}",
		"EMPTY":       "",
		"HASH":        "a#b",
		"ESCAPED_END": `ends with \`,
		"LATER":       "second",
		"COMMENTED":   "quoted",
		"APOSTROPHE":  "it's",
		"TABBED":      "yes",
		"exported":    "no",
		"AFTER":       "kept",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries %v, want %d", len(got), got, len(want))
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %q, want %q", key, got[key], value)
		}
	}
}

func TestQuoteHandling(t *testing.T) {
	// An escaped closing quote does not end a double-quoted value; it does end a single-quoted one.
	if closingQuote(`"a\"`, '"') != -1 || closingQuote(`'a\'`, '\'') != 3 || closingQuote(`"`, '"') != -1 {
		t.Fatal("closingQuote")
	}
	// A byte order mark at the start of the file doesn't hide the first key.
	if got := parseDotenv("\ufeffFIRST=\"one\" # note\nSECOND=two\n"); got["FIRST"] != "one" || got["SECOND"] != "two" {
		t.Fatalf("with a BOM: %v", got)
	}
	got := parseDotenv("KEY=\"line one \\\"\nline two\"\n")
	if got["KEY"] != "line one \"\nline two" {
		t.Fatalf("KEY = %q", got["KEY"])
	}
}

// SKYFLOW_CREDENTIALS comes from the environment first, then ./.env.
func TestEnvCredentialsReadsTheDotenvFile(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(dir) }()
	t.Setenv(CredentialsEnv, "")
	if got := EnvCredentials(); got != "" {
		t.Fatalf("no env and no .env: got %q", got)
	}
	if err := os.WriteFile(filepath.Join(".", dotenvFile), []byte("SKYFLOW_CREDENTIALS='{\"from\":\"dotenv\"}'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := EnvCredentials(); got != `{"from":"dotenv"}` {
		t.Fatalf(".env: got %q", got)
	}
	t.Setenv(CredentialsEnv, `{"from":"env"}`)
	if got := EnvCredentials(); got != `{"from":"env"}` {
		t.Fatalf("env should win: got %q", got)
	}
}
