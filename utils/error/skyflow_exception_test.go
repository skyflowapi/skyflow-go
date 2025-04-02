package errors_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/skyflowapi/skyflow-go/v2/utils/error"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServiceAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Skyflow Error Suite")
}

var _ = Describe("Skyflow Error", func() {

	Context("Getters", func() {
		var skyflowError *SkyflowError

		BeforeEach(func() {
			skyflowError = NewSkyflowError(INVALID_INPUT_CODE, "Invalid Input")
		})

		It("should return the correct message", func() {
			Expect(skyflowError.GetMessage()).To(Equal("Message: Invalid Input"))
		})

		It("should return the correct HTTP code", func() {
			Expect(skyflowError.GetCode()).To(Equal("Code: 400"))
		})

		It("should return the correct request ID", func() {
			Expect(skyflowError.GetRequestId()).To(Equal(""))
		})

		It("should return the correct gRPC code", func() {
			Expect(skyflowError.GetGrpcCode()).To(Equal(""))
		})

		It("should return the correct HTTP status code", func() {
			Expect(skyflowError.GetHttpStatusCode()).To(Equal("Bad Request"))
		})

		It("should return the correct details", func() {
			Expect(len(skyflowError.GetDetails())).To(Equal(0))
		})

		It("should return the correct response body", func() {
			Expect(len(skyflowError.GetResponseBody())).To(Equal(0))
		})
	})

	Context("SkyflowApiError", func() {
		It("should parse JSON response correctly", func() {
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			header.Set("x-request-id", "req-12345")
			header.Set("error-from-client", "true")
			response := http.Response{
				Header: header,
				Body: io.NopCloser(strings.NewReader(`{
					"error": {
						"http_code": 400,
						"message": "Invalid Input",
						"grpc_code": 3,
						"http_status": "Not found"
					}
				}`)),
			}

			skyflowError := SkyflowApiError(response)
			Expect(skyflowError.GetCode()).To(Equal("Code: 400"))
			Expect(skyflowError.GetMessage()).To(Equal("Message: Invalid Input"))
			Expect(skyflowError.GetGrpcCode()).To(Equal("3"))
			Expect(skyflowError.GetRequestId()).To(Equal("req-12345"))
			Expect(skyflowError.GetDetails()["errorFromClient"]).To(Equal("true"))
		})
		It("should parse error response correctly when error is string type", func() {
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			header.Set("x-request-id", "req-12345")
			response := http.Response{
				Header: header,
				Body: io.NopCloser(strings.NewReader(`{
					"error": "error occurred"
				}`)),
			}
			response.StatusCode = 400
			skyflowError := SkyflowApiError(response)
			Expect(skyflowError.GetMessage()).To(Equal("Message: error occurred"))
			Expect(skyflowError.GetRequestId()).To(Equal("req-12345"))
		})

		It("should parse JSON response correctly when error keys is missing from body", func() {
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			header.Set("x-request-id", "req-12345")
			response := http.Response{
				Header: header,
				Body: io.NopCloser(strings.NewReader(`{
					"error": {
					}
				}`)),
			}
			response.StatusCode = 400
			skyflowError := SkyflowApiError(response)
			Expect(skyflowError.GetCode()).To(Equal("Code: 400"))
			Expect(skyflowError.GetMessage()).To(Equal("Message: Unknown error"))
			Expect(skyflowError.GetGrpcCode()).To(Equal(""))
			Expect(skyflowError.GetRequestId()).To(Equal("req-12345"))
		})

		It("should parse JSON response correctly when message is missing", func() {
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			header.Set("x-request-id", "req-12345")
			response := http.Response{
				Header: header,
				Body: io.NopCloser(strings.NewReader(`{
					"error": {
					}
				}`)),
			}
			skyflowError := SkyflowApiError(response)
			Expect(skyflowError.GetCode()).To(Equal("Code: 0"))
			Expect(skyflowError.GetMessage()).To(Equal("Message: Unknown error"))
			Expect(skyflowError.GetGrpcCode()).To(Equal(""))
			Expect(skyflowError.GetRequestId()).To(Equal("req-12345"))
		})

		It("should parse plain text response correctly", func() {
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
				Body:   io.NopCloser(strings.NewReader("Plain text error")),
				Status: "400",
			}

			skyflowError := SkyflowApiError(response)
			Expect(skyflowError.GetMessage()).To(Equal("Message: Plain text error"))
			Expect(skyflowError.GetHttpStatusCode()).To(Equal("400"))
		})
	})
})
