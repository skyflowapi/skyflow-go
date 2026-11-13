package helpers_test

import (
	"bytes"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	. "github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	common "github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

var _ = Describe("Profile", func() {
	It("describes v2 to common/vaultclient", func() {
		profile := Profile()
		Expect(profile.SDKPrefix).To(Equal(constants.SDK_PREFIX))
		Expect(profile.LogPrefix).To(Equal(constants.SDK_LOG_PREFIX))
		Expect(profile.Auth).NotTo(BeNil())
		Expect(profile.ClusterURL).NotTo(BeNil())
		Expect(profile.SDKMetadata).NotTo(BeNil())
		Expect(profile.Defaults.InitialRetryDelay).To(Equal(500 * time.Millisecond))
		Expect(profile.Defaults.MaxRetryDelay).To(Equal(2 * time.Second))
		Expect(profile.Defaults.Timeout).To(Equal(60 * time.Second))
		Expect(profile.Defaults.MaxRetries).To(Equal(2))
		Expect(profile.Defaults.ConnectTimeout).To(BeZero())
	})
})

var _ = Describe("CommonVaultConfig", func() {
	var out bytes.Buffer

	BeforeEach(func() {
		out.Reset()
		logger.SetOutput(&out)
		logger.SetLogLevel(logger.WARN)
		DeferCleanup(logger.SetOutput, os.Stderr)
		DeferCleanup(logger.SetLogLevel, logger.ERROR)
	})

	It("copies every v2 field and leaves the HTTP settings unset, so the defaults apply", func() {
		v2 := common.VaultConfig{
			VaultId: "v", ClusterId: "c", BaseVaultUrl: "https://a.example.com",
			Env: common.SANDBOX, Credentials: common.Credentials{ApiKey: "k"},
		}
		got := CommonVaultConfig(v2)
		Expect(got.VaultId).To(Equal("v"))
		Expect(got.ClusterId).To(Equal("c"))
		Expect(got.BaseVaultUrl).To(Equal("https://a.example.com"))
		Expect(got.Env).To(Equal(common.SANDBOX))
		Expect(got.Credentials.ApiKey).To(Equal("k"))
		Expect(got.HttpConfig()).To(BeZero())
		Expect(out.String()).NotTo(ContainSubstring("Deprecated"))
	})

	It("folds the deprecated BaseVaultURL into BaseVaultUrl with a warning, BaseVaultUrl winning", func() {
		got := CommonVaultConfig(common.VaultConfig{VaultId: "v", BaseVaultURL: "https://b.example.com"})
		Expect(got.BaseVaultUrl).To(Equal("https://b.example.com"))
		Expect(out.String()).To(ContainSubstring(logs.DEPRECATED_FIELD_BASE_VAULT_URL))

		got = CommonVaultConfig(common.VaultConfig{VaultId: "v", BaseVaultUrl: "https://a.example.com", BaseVaultURL: "https://b.example.com"})
		Expect(got.BaseVaultUrl).To(Equal("https://a.example.com"))
		Expect(BaseVaultUrl(common.VaultConfig{})).To(BeEmpty())
	})
})

var _ = Describe("Common service-account options", func() {
	var out bytes.Buffer

	BeforeEach(func() {
		out.Reset()
		logger.SetOutput(&out)
		logger.SetLogLevel(logger.WARN)
		DeferCleanup(logger.SetOutput, os.Stderr)
		DeferCleanup(logger.SetLogLevel, logger.ERROR)
	})

	It("passes RoleIds and Ctx through, dropping LogLevel", func() {
		got := CommonBearerTokenOptions(common.BearerTokenOptions{Ctx: "ctx", RoleIds: []string{"r1"}, LogLevel: logger.DEBUG})
		Expect(got.Ctx).To(Equal("ctx"))
		Expect(got.RoleIds).To(Equal([]string{"r1"}))
		Expect(out.String()).NotTo(ContainSubstring("Deprecated"))
	})

	It("falls back to the deprecated RoleIDs with a warning, RoleIds winning", func() {
		got := CommonBearerTokenOptions(common.BearerTokenOptions{RoleIDs: []string{"old"}})
		Expect(got.RoleIds).To(Equal([]string{"old"}))
		Expect(out.String()).To(ContainSubstring(logs.DEPRECATED_FIELD_ROLE_IDS))

		got = CommonBearerTokenOptions(common.BearerTokenOptions{RoleIds: []string{"new"}, RoleIDs: []string{"old"}})
		Expect(got.RoleIds).To(Equal([]string{"new"}))
	})

	It("converts signed data token options", func() {
		got := CommonSignedDataTokensOptions(common.SignedDataTokensOptions{DataTokens: []string{"t"}, TimeToLive: 30, Ctx: "c", LogLevel: logger.INFO})
		Expect(got.DataTokens).To(Equal([]string{"t"}))
		Expect(got.TimeToLive).To(Equal(30))
		Expect(got.Ctx).To(Equal("c"))
	})
})
