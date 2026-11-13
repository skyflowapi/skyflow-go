package helpers_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	. "github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	common "github.com/skyflowapi/skyflow-go/v2/utils/common"
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
	It("copies every v2 field and leaves the HTTP settings unset, so the defaults apply", func() {
		v2 := common.VaultConfig{
			VaultId: "v", ClusterId: "c", BaseVaultUrl: "https://a.example.com", BaseVaultURL: "https://b.example.com",
			Env: common.SANDBOX, Credentials: common.Credentials{ApiKey: "k"},
		}
		got := CommonVaultConfig(v2)
		Expect(got.VaultId).To(Equal("v"))
		Expect(got.ClusterId).To(Equal("c"))
		Expect(got.BaseVaultUrl).To(Equal("https://a.example.com"))
		Expect(got.BaseVaultURL).To(Equal("https://b.example.com"))
		Expect(got.Env).To(Equal(common.SANDBOX))
		Expect(got.Credentials.ApiKey).To(Equal("k"))
		Expect(got.HttpConfig()).To(BeZero())
	})
})
