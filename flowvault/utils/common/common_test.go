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

var _ = Describe("RequestContext", func() {
	It("has no headers until one is added", func() {
		var rc RequestContext
		Expect(rc.Headers()).To(BeEmpty())
	})

	It("adds headers, keeping the last value per key, and returns a copy", func() {
		rc := &RequestContext{Operation: "QUERY"}
		rc.AddHeader(RequestIdHeader, "first")
		rc.AddHeader(RequestIdHeader, "second")
		rc.AddHeader(SkyflowAccountId, "acc")

		headers := rc.Headers()
		Expect(headers).To(Equal(map[CustomHeaderKey]string{RequestIdHeader: "second", SkyflowAccountId: "acc"}))
		headers[SkyflowAccountName] = "changed"
		Expect(rc.Headers()).ToNot(HaveKey(SkyflowAccountName))
	})
})
