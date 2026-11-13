package errors_test

import (
	"errors"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

func TestErrors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Errors Suite")
}

var _ = Describe("SkyflowError", func() {
	It("keeps the cause of a cancelled request", func() {
		cause := errors.New("context canceled")
		err := NewSkyflowErrorWithCause(REQUEST_CANCELLED, "cancelled", cause)
		Expect(errors.Is(err, cause)).To(BeTrue())
		Expect(err.GetHttpCode()).To(Equal("499"))
	})

	It("builds an error from a code and message", func() {
		err := NewSkyflowError(INVALID_INPUT_CODE, "bad input")
		Expect(err.GetMessage()).To(ContainSubstring("bad input"))
		Expect(err.GetCode()).To(ContainSubstring("400"))
	})

	It("builds an error from an API error and its headers", func() {
		header := http.Header{}
		header.Set("x-request-id", "req-2")
		err := SkyflowErrorApi(errors.New(`404: {"error":{"http_code":404,"message":"Not found"}}`), header)
		Expect(err.GetMessage()).To(ContainSubstring("Not found"))
		Expect(err.GetRequestId()).To(Equal("req-2"))
	})
})
