package errors_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
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
	It("builds an error from a code and message", func() {
		err := NewSkyflowError(INVALID_INPUT_CODE, "bad input")
		Expect(err.GetMessage()).To(ContainSubstring("bad input"))
		Expect(err.GetCode()).To(ContainSubstring("400"))
	})

	It("builds an error from an API response", func() {
		header := http.Header{}
		header.Set("Content-Type", "application/json")
		header.Set("x-request-id", "req-1")
		err := SkyflowApiError(http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader(`{"error":{"http_code":400,"message":"Invalid query"}}`)),
		})
		Expect(err.GetMessage()).To(ContainSubstring("Invalid query"))
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetRequestId()).To(Equal("req-1"))
	})

	It("builds an error from an API error and its headers", func() {
		header := http.Header{}
		header.Set("x-request-id", "req-2")
		err := SkyflowErrorApi(errors.New(`404: {"error":{"http_code":404,"message":"Not found"}}`), header)
		Expect(err.GetMessage()).To(ContainSubstring("Not found"))
		Expect(err.GetRequestId()).To(Equal("req-2"))
	})
})
