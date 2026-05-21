package common_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Suite")
}

var _ = Describe("RequestMethod", func() {
	Context("IsValid", func() {
		DescribeTable("should return true for all valid HTTP methods",
			func(method common.RequestMethod) {
				Expect(method.IsValid()).To(BeTrue())
			},
			Entry("GET", common.GET),
			Entry("POST", common.POST),
			Entry("PUT", common.PUT),
			Entry("PATCH", common.PATCH),
			Entry("DELETE", common.DELETE),
		)

		It("should return false for an unrecognised method string", func() {
			Expect(common.RequestMethod("OPTIONS").IsValid()).To(BeFalse())
		})

		It("should return false for an empty string", func() {
			Expect(common.RequestMethod("").IsValid()).To(BeFalse())
		})

		It("should be case-sensitive — lowercase is invalid", func() {
			Expect(common.RequestMethod("get").IsValid()).To(BeFalse())
		})
	})
})
