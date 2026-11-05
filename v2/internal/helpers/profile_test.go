package helpers_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	. "github.com/skyflowapi/skyflow-go/v2/internal/helpers"
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
		// As in the Java and Python SkyVault Detect clients.
		Expect(profile.Defaults.Timeout).To(Equal(60 * time.Second))
		Expect(profile.Defaults.MaxRetries).To(Equal(2))
		Expect(profile.Defaults.ConnectTimeout).To(BeZero())
	})
})
