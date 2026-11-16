package errors_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

// The error logic lives in skyflow-go/common/errors and is tested there. These specs check
// that v2's wrappers forward to it and keep v2's error codes.

func TestSkyflowError(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SkyflowError Suite")
}

const apiErrorBody = `{"error":{"grpc_code":3,"http_code":400,"http_status":"Bad Request","message":"Invalid request","details":[]}}`

var _ = Describe("SkyflowError", func() {
	It("creates an error with a code and message", func() {
		err := skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "bad input")
		Expect(err.GetHttpCode()).To(Equal(string(skyflowError.INVALID_INPUT_CODE)))
		Expect(err.GetMessage()).To(ContainSubstring("bad input"))
	})

	It("keeps v2's error codes", func() {
		Expect(string(skyflowError.SERVER)).To(Equal("500"))
		Expect(string(skyflowError.INVALID_INPUT_CODE)).To(Equal("400"))
		Expect(string(skyflowError.PARTIAL)).To(Equal("404"))
	})

	It("builds an error from an API response", func() {
		response := http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"req-1"}},
			Body:       io.NopCloser(strings.NewReader(apiErrorBody)),
		}
		err := skyflowError.SkyflowApiError(response)
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("Invalid request"))
		Expect(err.GetGrpcCode()).To(Equal("3"))
		Expect(err.GetRequestId()).To(Equal("req-1"))
	})

	It("builds an error from a generated client error", func() {
		header := http.Header{"X-Request-Id": {"req-2"}}
		err := skyflowError.SkyflowErrorApi(errors.New("400: "+apiErrorBody), header)
		Expect(err.GetHttpCode()).To(Equal("400"))
		Expect(err.GetMessage()).To(ContainSubstring("Invalid request"))
		Expect(err.GetRequestId()).To(Equal("req-2"))
	})
})
