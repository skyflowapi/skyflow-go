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
		Expect(validApiKey).To(HaveLen(apiKeyLength))
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
			Entry("http with port and path", "http://localhost:8080/v1"),
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

		It("should warn and fall back to the deprecated BaseVaultURL", func() {
			err := v.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultURL: "https://example.com"})
			Expect(err).To(BeNil())
			Expect(logBuf.String()).To(ContainSubstring("level=warning"))
			Expect(logBuf.String()).To(ContainSubstring(testLogPrefix + logDeprecatedFieldBaseVaultUrl))
		})

		It("should validate the deprecated BaseVaultURL when used as fallback", func() {
			expectInvalid(v.ValidateVaultConfig(VaultConfig{VaultId: "v", BaseVaultURL: "not a url"}), errInvalidVaultUrl)
		})

		It("should prefer BaseVaultUrl over the deprecated BaseVaultURL", func() {
			err := v.ValidateVaultConfig(VaultConfig{
				VaultId:      "v",
				BaseVaultUrl: "https://example.com",
				BaseVaultURL: "not a url",
			})
			Expect(err).To(BeNil())
			Expect(logBuf.String()).To(ContainSubstring(logDeprecatedFieldBaseVaultUrl))
		})

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
			})).To(BeNil())
			Expect(v.ValidateHttpConfig(HttpConfig{MaxRetries: &zero})).To(BeNil())
		})

		DescribeTable("should reject negative settings",
			func(hc HttpConfig) {
				expectInvalid(v.ValidateHttpConfig(hc), errNegativeHttpConfig)
				Expect(logBuf.String()).To(ContainSubstring(logNegativeHttpConfig))
			},
			Entry("timeout", HttpConfig{Timeout: -time.Second}),
			Entry("initial retry delay", HttpConfig{InitialRetryDelay: -time.Millisecond}),
			Entry("max retry delay", HttpConfig{MaxRetryDelay: -time.Second}),
			Entry("max retries", HttpConfig{MaxRetries: &neg}),
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

		It("should check the deprecated base vault URL when it is the only one set", func() {
			expectInvalid(v.ValidateUpdateVaultConfig(VaultConfig{VaultId: "v", BaseVaultURL: "not a url"}), errInvalidVaultUrl)
			Expect(logBuf.String()).To(ContainSubstring(logDeprecatedFieldBaseVaultUrl))
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
			Entry("right length, no prefix", strings.Repeat("a", apiKeyLength)),
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
