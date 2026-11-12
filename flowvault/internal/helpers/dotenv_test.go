package helpers

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SKYFLOW_CREDENTIALS from the environment or .env", func() {
	const key = "SKYFLOW_CREDENTIALS"

	BeforeEach(func() {
		GinkgoT().Setenv(key, "")
		previous := dotenvPath
		dotenvPath = filepath.Join(GinkgoT().TempDir(), ".env")
		DeferCleanup(func() { dotenvPath = previous })
	})

	writeDotenv := func(content string) {
		Expect(os.WriteFile(dotenvPath, []byte(content), 0o600)).To(Succeed())
	}

	It("prefers the environment variable, trimmed", func() {
		writeDotenv(`SKYFLOW_CREDENTIALS={"from":"file"}`)
		GinkgoT().Setenv(key, "  {\"from\":\"env\"}\n")
		Expect(EnvCredentials()).To(Equal(`{"from":"env"}`))
	})

	It("falls back to the .env file, and to nothing without one", func() {
		Expect(EnvCredentials()).To(BeEmpty())
		writeDotenv("# Skyflow\nOTHER=1\nexport SKYFLOW_CREDENTIALS='{\"clientID\":\"c\",\"privateKey\":\"-----BEGIN\\nKEY\"}'\n")
		Expect(EnvCredentials()).To(Equal(`{"clientID":"c","privateKey":"-----BEGIN\nKEY"}`))
	})

	DescribeTable("reads the usual .env forms",
		func(content, want string) { Expect(dotenvValue(content, key)).To(Equal(want)) },
		Entry("unquoted", "SKYFLOW_CREDENTIALS=abc", "abc"),
		Entry("unquoted with a comment", "SKYFLOW_CREDENTIALS=abc # note", "abc"),
		Entry("spaces around =", "SKYFLOW_CREDENTIALS = abc ", "abc"),
		Entry("export prefix", "export SKYFLOW_CREDENTIALS=abc", "abc"),
		Entry("single quotes, as they are", `SKYFLOW_CREDENTIALS='{"a":"b\n"}' # c`, `{"a":"b\n"}`),
		Entry("double quotes with escaped quotes", `SKYFLOW_CREDENTIALS="{\"a\":\"b\\n\"}"`, `{"a":"b\n"}`),
		Entry("a quoted value over several lines", "SKYFLOW_CREDENTIALS='{\n  \"a\": 1\n}'\nNEXT=2", "{\n  \"a\": 1\n}"),
		Entry("an unclosed quote takes the rest", "SKYFLOW_CREDENTIALS=\"{\nabc", "{\nabc"),
		Entry("CRLF line endings", "A=1\r\nSKYFLOW_CREDENTIALS=abc\r\n", "abc"),
		Entry("the last assignment wins", "SKYFLOW_CREDENTIALS=one\nSKYFLOW_CREDENTIALS=two", "two"),
		Entry("a commented-out assignment", "#SKYFLOW_CREDENTIALS=abc", ""),
		Entry("another key only", "SKYFLOW_CREDENTIALS_X=abc\nNOT_IT=1", ""),
		Entry("a line without =", "SKYFLOW_CREDENTIALS", ""),
		Entry("an empty value", "SKYFLOW_CREDENTIALS=", ""),
	)
})
