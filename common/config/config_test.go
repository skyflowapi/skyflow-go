package config

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config types", func() {
	Context("Env", func() {
		It("should default to PROD", func() {
			var e Env
			Expect(e).To(Equal(PROD))
		})

		DescribeTable("should keep the released v2 values",
			func(e Env, expected int) {
				Expect(int(e)).To(Equal(expected))
			},
			Entry("PROD", PROD, 0),
			Entry("STAGE", STAGE, 1),
			Entry("SANDBOX", SANDBOX, 2),
			Entry("DEV", DEV, 3),
		)
	})

	Context("VaultConfig.HttpConfig", func() {
		It("should copy the vault's HTTP settings", func() {
			retries := 3
			vc := VaultConfig{
				VaultId:           "vault",
				Timeout:           5 * time.Second,
				MaxRetries:        &retries,
				InitialRetryDelay: 100 * time.Millisecond,
				MaxRetryDelay:     2 * time.Second,
			}
			Expect(vc.HttpConfig()).To(Equal(HttpConfig{
				Timeout:           5 * time.Second,
				MaxRetries:        &retries,
				InitialRetryDelay: 100 * time.Millisecond,
				MaxRetryDelay:     2 * time.Second,
			}))
		})

		It("should return the zero HttpConfig when nothing is set", func() {
			Expect(VaultConfig{}.HttpConfig()).To(Equal(HttpConfig{}))
		})
	})

	Context("CustomHeaderKey", func() {
		It("should expose the header names and deprecated aliases", func() {
			Expect(string(SkyflowAccountId)).To(Equal("x-skyflow-account-id"))
			Expect(string(SkyflowAccountName)).To(Equal("x-skyflow-account-name"))
			Expect(string(RequestIdHeader)).To(Equal("x-request-id"))
			Expect(SkyflowAccountID).To(Equal(SkyflowAccountId))
			Expect(RequestIDHeader).To(Equal(RequestIdHeader))
		})
	})
})
