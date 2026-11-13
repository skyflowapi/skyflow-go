package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

const (
	testSDKPrefix = "Skyflow Go SDK vTest"
	testLogPrefix = "[ Skyflow Go SDK vTest ] "
	validApiKey   = "sky-abcde-1234567890abcdef1234567890abcdef" // 42 chars
)

func expectInvalid(err *skyflowError.SkyflowError, errBase string) {
	Expect(err).NotTo(BeNil())
	Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
	Expect(err.GetMessage()).To(Equal("Message: " + testSDKPrefix + errBase))
}

var _ = Describe("Validator", func() {
	var (
		v      Validator
		logBuf *bytes.Buffer
	)

	BeforeEach(func() {
		v = NewValidator(testSDKPrefix, testLogPrefix)
		logBuf = &bytes.Buffer{}
		logger.SetOutput(logBuf)
		logger.SetLogLevel(logger.WARN)
	})

	// The logger is process-global: restore the defaults after every test.
	AfterEach(func() {
		logger.SetOutput(os.Stderr)
		logger.SetLogLevel(logger.ERROR)
	})

	It("should keep the prefixes it was built with", func() {
		Expect(v).To(Equal(Validator{SDKPrefix: testSDKPrefix, LogPrefix: testLogPrefix}))
	})

	It("should sanity check the test api key length", func() {
		Expect(apiKeyPattern.MatchString(validApiKey)).To(BeTrue())
	})

	Context("ValidateVaultConfig", func() {
		It("should reject a missing vault ID and log with the log prefix", func() {
			expectInvalid(v.ValidateVaultConfig(VaultConfig{ClusterId: "c"}), errInvalidVaultId)
			Expect(logBuf.String()).To(ContainSubstring("level=error"))
			Expect(logBuf.String()).To(ContainSubstring(testLogPrefix + logVaultIdIsRequired))
		})

		It("should require a cluster ID when no vault URL is set", func() {
			expectInvalid(v.ValidateVaultConfig(VaultConfig{VaultId: "v"}), errInvalidClusterId)
			Expect(logBuf.String()).To(ContainSubstring(logClusterIdIsRequired))
		})

		It("should accept a vault ID with a cluster ID", func() {
			Expect(v.ValidateVaultConfig(VaultConfig{VaultId: "v", ClusterId: "c", Env: DEV})).To(BeNil())
		})

		It("should accept the default Env", func() {
			Expect(v.ValidateVaultConfig(VaultConfig{VaultId: "v", ClusterId: "c"})).To(BeNil())
		})

		DescribeTable("should accept valid BaseVaultUrl without a cluster ID",
			func(u string) {
				Expect(v.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: u})).To(BeNil())
			},
			Entry("https", "https://example.vault.skyflowapis.com"),
			Entry("http localhost with port and path", "http://localhost:8080/v1"),
			Entry("http LOCALHOST", "http://LOCALHOST"),
			Entry("http IPv4 loopback", "http://127.0.0.1:9000"),
			Entry("http IPv6 loopback", "http://[::1]:9000"),
		)

		It("should accept plain http for any host with WithInsecureVaultUrl, but still not a malformed URL", func() {
			lenient := NewValidator(testSDKPrefix, testLogPrefix, WithInsecureVaultUrl())
			Expect(lenient.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "http://10.0.0.5"})).To(BeNil())
			Expect(lenient.ValidateUpdateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "http://vault.example.com"})).To(BeNil())
			expectInvalid(lenient.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "ftp://vault.example.com"}), errInvalidVaultUrl)
		})

		DescribeTable("should reject plain http for a non-loopback host",
			func(u string) {
				expectInvalid(v.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: u}), errInsecureVaultUrl)
				Expect(logBuf.String()).To(ContainSubstring(logVaultUrlIsInsecure))
			},
			Entry("public host", "http://example.vault.skyflowapis.com"),
			Entry("private IP", "http://10.0.0.5"),
			Entry("localhost lookalike", "http://localhost.example.com"),
		)

		DescribeTable("should reject an invalid BaseVaultUrl",
			func(u string) {
				expectInvalid(v.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: u}), errInvalidVaultUrl)
				Expect(logBuf.String()).To(ContainSubstring(logVaultUrlIsInvalid))
			},
			Entry("unparseable", "://bad"),
			Entry("unsupported scheme", "ftp://example.com"),
			Entry("no scheme", "example.com"),
			Entry("missing host", "https://"),
			Entry("bad escape", "http://%zz"),
		)

		It("should validate the vault's HTTP settings", func() {
			expectInvalid(v.ValidateVaultConfig(VaultConfig{VaultId: "v", ClusterId: "c", Timeout: -1}), errNegativeHttpConfig)
		})
	})

	Context("ValidateHttpConfig", func() {
		neg := -1
		zero := 0
		pos := 5

		It("should accept the zero value", func() {
			Expect(v.ValidateHttpConfig(HttpConfig{})).To(BeNil())
		})

		It("should accept positive and zero settings", func() {
			Expect(v.ValidateHttpConfig(HttpConfig{
				Timeout: time.Second, MaxRetries: &pos, InitialRetryDelay: time.Millisecond, MaxRetryDelay: time.Second,
				ConnectTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second,
			})).To(BeNil())
			Expect(v.ValidateHttpConfig(HttpConfig{MaxRetries: &zero})).To(BeNil())
		})

		DescribeTable("should reject negative settings",
			func(hc HttpConfig) {
				expectInvalid(v.ValidateHttpConfig(hc), errNegativeHttpConfig)
				Expect(logBuf.String()).To(ContainSubstring(logNegativeHttpConfig))
			},
			Entry("timeout", HttpConfig{Timeout: -time.Second}),
			Entry("connect timeout", HttpConfig{ConnectTimeout: -time.Second}),
			Entry("read timeout", HttpConfig{ReadTimeout: -time.Second}),
			Entry("write timeout", HttpConfig{WriteTimeout: -time.Second}),
			Entry("initial retry delay", HttpConfig{InitialRetryDelay: -time.Millisecond}),
			Entry("max retry delay", HttpConfig{MaxRetryDelay: -time.Second}),
			Entry("max retries", HttpConfig{MaxRetries: &neg}),
			Entry("connect timeout", HttpConfig{ConnectTimeout: -time.Second}),
			Entry("read timeout", HttpConfig{ReadTimeout: -time.Second}),
			Entry("write timeout", HttpConfig{WriteTimeout: -time.Second}),
		)
	})

	Context("ValidateUpdateVaultConfig", func() {
		It("should reject a missing vault ID", func() {
			expectInvalid(v.ValidateUpdateVaultConfig(VaultConfig{ClusterId: "c"}), errInvalidVaultId)
		})

		It("should accept a vault ID alone", func() {
			Expect(v.ValidateUpdateVaultConfig(VaultConfig{VaultId: "v"})).To(BeNil())
		})

		It("should check a base vault URL only when one is set", func() {
			Expect(v.ValidateUpdateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "https://vault.example.com"})).To(BeNil())
			expectInvalid(v.ValidateUpdateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "ftp://vault.example.com"}), errInvalidVaultUrl)
		})

		It("should reject an update to a plain http vault URL", func() {
			expectInvalid(v.ValidateUpdateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "http://vault.example.com"}), errInsecureVaultUrl)
			Expect(v.ValidateUpdateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "http://127.0.0.1:8080"})).To(BeNil())
		})
	})

	Context("ValidateCredentials", func() {
		It("should reject no credential source", func() {
			expectInvalid(v.ValidateCredentials(Credentials{}), errNoTokenGenerationMeansPassed)
			Expect(logBuf.String()).To(ContainSubstring(logNoTokenGenerationMeansPassed))
		})

		It("should reject roles-only credentials as having no source", func() {
			expectInvalid(v.ValidateCredentials(Credentials{Roles: []string{"r"}}), errNoTokenGenerationMeansPassed)
		})

		DescribeTable("should reject more than one credential source",
			func(c Credentials) {
				expectInvalid(v.ValidateCredentials(c), errMultipleTokenGenerationMeansPassed)
				Expect(logBuf.String()).To(ContainSubstring(logMultipleTokenGenerationMeansPassed))
			},
			Entry("path + token", Credentials{Path: "p", Token: "t"}),
			Entry("credentialsString + apiKey", Credentials{CredentialsString: "c", ApiKey: validApiKey}),
			Entry("all four", Credentials{Path: "p", CredentialsString: "c", Token: "t", ApiKey: validApiKey}),
		)

		DescribeTable("should accept exactly one credential source",
			func(c Credentials) {
				Expect(v.ValidateCredentials(c)).To(BeNil())
			},
			Entry("path", Credentials{Path: "p"}),
			Entry("credentialsString", Credentials{CredentialsString: "c"}),
			Entry("token", Credentials{Token: "t"}),
			Entry("apiKey", Credentials{ApiKey: validApiKey}),
			Entry("path with roles", Credentials{Path: "p", Roles: []string{"r1", "r2"}}),
		)

		DescribeTable("should reject a malformed api key",
			func(key string) {
				expectInvalid(v.ValidateCredentials(Credentials{ApiKey: key}), errInvalidApiKey)
				Expect(logBuf.String()).To(ContainSubstring(logInvalidApiKey))
			},
			Entry("too short", "sky-short"),
			Entry("too long", validApiKey+"x"),
			Entry("right length, no prefix", strings.Repeat("a", len(validApiKey))),
			// Checked against the full key format, not just length and prefix.
			Entry("prefix not at the start", "x"+validApiKey[:len(validApiKey)-1]),
			Entry("non-hex secret", validApiKey[:len(validApiKey)-1]+"z"),
			Entry("short account part", "sky-abcd-"+validApiKey[10:]+"0"),
		)

		It("should reject an empty roles slice", func() {
			expectInvalid(v.ValidateCredentials(Credentials{Token: "t", Roles: []string{}}), errEmptyRoles)
			Expect(logBuf.String()).To(ContainSubstring(logEmptyRoles))
		})

		It("should reject an empty role and log its index", func() {
			expectInvalid(v.ValidateCredentials(Credentials{Token: "t", Roles: []string{"r1", ""}}), errEmptyRoleInRoles)
			Expect(logBuf.String()).To(ContainSubstring(testLogPrefix + fmt.Sprintf(logEmptyOrNullRoleInRoles, 1)))
		})
	})

	Context("IsCredentialsEmpty", func() {
		It("should be true when no source is set", func() {
			Expect(IsCredentialsEmpty(Credentials{})).To(BeTrue())
			Expect(IsCredentialsEmpty(Credentials{Roles: []string{"r"}, Context: "ctx"})).To(BeTrue())
		})

		DescribeTable("should be false when any source is set",
			func(c Credentials) {
				Expect(IsCredentialsEmpty(c)).To(BeFalse())
			},
			Entry("path", Credentials{Path: "p"}),
			Entry("credentialsString", Credentials{CredentialsString: "c"}),
			Entry("token", Credentials{Token: "t"}),
			Entry("apiKey", Credentials{ApiKey: "k"}),
		)
	})
})

var _ = Describe("Validator options", func() {
	It("uses an SDK's own messages and error options, keeping the defaults for the rest", func() {
		v := NewValidator(testSDKPrefix, testLogPrefix,
			WithMessages(Messages{
				InvalidVaultId:    " vault ID must not be empty.",
				ClusterIdRequired: " give a cluster ID or a vault URL.",
				InvalidVaultUrl:   " bad URL.",
				InsecureVaultUrl:  " use https.",
			}),
			WithErrorOptions(skyflowError.RawMessage()),
		)
		Expect(v.ValidateVaultConfig(VaultConfig{}).GetMessage()).To(Equal(testSDKPrefix + " vault ID must not be empty."))
		Expect(v.ValidateVaultConfig(VaultConfig{VaultId: "v"}).GetMessage()).To(Equal(testSDKPrefix + " give a cluster ID or a vault URL."))
		Expect(v.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "ftp://x"}).GetMessage()).To(Equal(testSDKPrefix + " bad URL."))
		Expect(v.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultUrl: "http://vault.example.com"}).GetMessage()).To(Equal(testSDKPrefix + " use https."))
		Expect(v.ValidateCredentials(Credentials{}).GetMessage()).To(Equal(testSDKPrefix + errNoTokenGenerationMeansPassed))
		Expect(v.ValidateCredentials(Credentials{Token: "t", Roles: []string{""}}).GetMessage()).To(Equal(testSDKPrefix + errEmptyRoleInRoles))
	})

	It("keeps the default message where an SDK sets none", func() {
		v := NewValidator(testSDKPrefix, testLogPrefix, WithMessages(Messages{InvalidVaultId: " custom."}))
		expectInvalid(v.ValidateVaultConfig(VaultConfig{VaultId: "v"}), errInvalidClusterId)
	})
})
