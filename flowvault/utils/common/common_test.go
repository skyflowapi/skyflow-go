package common_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/skyflowapi/skyflow-go/flowvault/internal/testutil"
	. "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Common Suite")
}

var _ = Describe("Credentials", func() {
	It("is empty only when no credential source is set", func() {
		Expect(IsCredentialsEmpty(Credentials{})).To(BeTrue())
		Expect(IsCredentialsEmpty(Credentials{Roles: []string{"r"}})).To(BeTrue())
		Expect(IsCredentialsEmpty(Credentials{ApiKey: testutil.Resource("client_api_key.txt")})).To(BeFalse())
		Expect(IsCredentialsEmpty(Credentials{Token: "t"})).To(BeFalse())
		Expect(IsCredentialsEmpty(Credentials{Path: "p"})).To(BeFalse())
		Expect(IsCredentialsEmpty(Credentials{CredentialsString: "{}"})).To(BeFalse())
	})
})

var _ = Describe("Ptr", func() {
	It("returns a pointer to its own copy of the value", func() {
		value := 2
		pointer := Ptr(value)
		value = 3
		Expect(*pointer).To(Equal(2))
		Expect(*Ptr(true)).To(BeTrue())
		Expect(*Ptr(float32(0.5))).To(Equal(float32(0.5)))
	})
})

var _ = Describe("DeleteFilesStatusFromWire", func() {
	It("matches DELETED case-insensitively and maps anything else to UNKNOWN", func() {
		Expect(DeleteFilesStatusFromWire("DELETED")).To(Equal(DeleteFilesStatusDeleted))
		Expect(DeleteFilesStatusFromWire(" deleted ")).To(Equal(DeleteFilesStatusDeleted))
		Expect(DeleteFilesStatusFromWire("ARCHIVED")).To(Equal(DeleteFilesStatusUnknown))
		Expect(DeleteFilesStatusFromWire("")).To(Equal(DeleteFilesStatusUnknown))
	})
})
