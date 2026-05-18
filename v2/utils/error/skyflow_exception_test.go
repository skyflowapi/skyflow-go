package errors

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServiceAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Skyflow Error Suite")
}

// errReader always returns an error on Read, used to simulate body-read failures.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, fmt.Errorf("forced read error") }

var _ = Describe("Skyflow Error", func() {

	Context("Error() Method Safe Checks", func() {
		It("should return message without panic when originalError is nil", func() {
			skyflowError := &SkyflowError{
				message:       "Simple error",
				originalError: nil,
			}
			Expect(skyflowError.Error()).To(Equal("Message: Simple error"))
		})

		It("should return combined message when originalError is present", func() {
			originalErr := errors.New("database connection failed")
			skyflowError := &SkyflowError{
				message:       "Operation failed",
				originalError: originalErr,
			}
			expectedMsg := "Message: Operation failed, Original Error (if any): database connection failed"
			Expect(skyflowError.Error()).To(Equal(expectedMsg))
		})
	})

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
			Expect(skyflowError.GetDetails()[0]).To(Equal(map[string]interface{}{"errorFromClient": true}))
		})
		It("should parse JSON response correctly when error-from-client is in wrong format", func() {
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			header.Set("x-request-id", "req-12345")
			header.Set("error-from-client", "maybe")
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
			Expect(skyflowError.GetDetails()[0]).To(Equal(map[string]interface{}{"errorFromClient": false}))
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

	Context("SkyflowApiError — text/plain; charset=utf-8 content type", func() {
		It("should parse plain body when body is not JSON", func() {
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
					"X-Request-Id": []string{"req-utf8"},
				},
				Body:   io.NopCloser(strings.NewReader("plain error message")),
				Status: "503 Service Unavailable",
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("plain error message"))
			Expect(skyflowErr.GetHttpStatusCode()).To(Equal("503 Service Unavailable"))
		})

		It("should parse JSON error body when body is valid JSON", func() {
			body := `{"error":{"http_code":403,"message":"Forbidden","grpc_code":7,"http_status":"Forbidden"}}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body:   io.NopCloser(strings.NewReader(body)),
				Status: "403 Forbidden",
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("Forbidden"))
			Expect(skyflowErr.GetCode()).To(Equal("Code: 403"))
			Expect(skyflowErr.GetGrpcCode()).To(Equal("7"))
		})

		It("should parse string error when JSON error field is a string", func() {
			body := `{"error":"access denied"}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body:   io.NopCloser(strings.NewReader(body)),
				Status: "401 Unauthorized",
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("access denied"))
		})
	})

	Context("SkyflowApiError — application/json with invalid JSON body", func() {
		It("should return a parse-failure error when JSON is malformed", func() {
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(strings.NewReader(`{invalid json`)),
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("unmarshal"))
		})
	})

	Context("SkyflowApiError — application/json with details in error body", func() {
		It("should parse details array when present in error body", func() {
			body := `{"error":{"http_code":400,"message":"Bad Request","grpc_code":3,"http_status":"Bad Request","details":[{"type":"detail1"}]}}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(strings.NewReader(body)),
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetCode()).To(Equal("Code: 400"))
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("Bad Request"))
			Expect(skyflowErr.GetDetails()).To(HaveLen(1))
		})
	})

	Context("SkyflowApiError — application/json with error as neither string nor map", func() {
		It("should use raw body as message when error field is numeric", func() {
			body := `{"error":42}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(strings.NewReader(body)),
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetMessage()).To(ContainSubstring(body))
		})
	})

	Context("SkyflowApiError — text/plain read error", func() {
		It("should return parse-failure error when body read fails", func() {
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain"},
				},
				Body: io.NopCloser(errReader{}),
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("Failed to read error"))
		})
	})

	Context("SkyflowApiError — text/plain; charset=utf-8 read error", func() {
		It("should return parse-failure error when body read fails", func() {
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body: io.NopCloser(errReader{}),
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("Failed to read error"))
		})
	})

	Context("SkyflowApiError — text/plain; charset=utf-8 missing http_code in error body", func() {
		It("should use response StatusCode when http_code is absent from error body", func() {
			body := `{"error":{"message":"Service Unavailable","grpc_code":14,"http_status":"Service Unavailable"}}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body:       io.NopCloser(strings.NewReader(body)),
				Status:     "503 Service Unavailable",
				StatusCode: 503,
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetCode()).To(Equal("Code: 503"))
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("Service Unavailable"))
		})
	})

	Context("SkyflowApiError — text/plain; charset=utf-8 missing message in error body", func() {
		It("should return Unknown error when message is absent", func() {
			body := `{"error":{"http_code":500}}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body:   io.NopCloser(strings.NewReader(body)),
				Status: "500 Internal Server Error",
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("Unknown error"))
		})
	})

	Context("SkyflowApiError — text/plain; charset=utf-8 with details in error body", func() {
		It("should parse details array when present", func() {
			body := `{"error":{"http_code":403,"message":"Forbidden","grpc_code":7,"http_status":"Forbidden","details":[{"reason":"insufficient_permissions"}]}}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body:   io.NopCloser(strings.NewReader(body)),
				Status: "403 Forbidden",
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetMessage()).To(ContainSubstring("Forbidden"))
			Expect(skyflowErr.GetDetails()).To(HaveLen(1))
		})
	})

	Context("SkyflowApiError — text/plain; charset=utf-8 with error as neither string nor map", func() {
		It("should use raw body as message when error field is numeric", func() {
			body := `{"error":99}`
			response := http.Response{
				Header: http.Header{
					"Content-Type": []string{"text/plain; charset=utf-8"},
				},
				Body:   io.NopCloser(strings.NewReader(body)),
				Status: "500",
			}
			skyflowErr := SkyflowApiError(response)
			Expect(skyflowErr.GetMessage()).To(ContainSubstring(body))
		})
	})

	Context("SkyflowErrorApi", func() {
		var header http.Header
		BeforeEach(func() {
			header = http.Header{}
			header.Set("x-request-id", "req-67890")
		})

		It("should correctly parse a valid API error message", func() {
			errorJSON := `{"error": {"http_code": 400, "message": "Invalid request", "grpc_code": 3, "http_status": "Bad Request", "details": {"demo": "demo"}}}`
			err := errors.New("API Error: " + errorJSON)
			skyflowErr := SkyflowErrorApi(err, header)

			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetCode()).To(Equal("Code: 400"))
			Expect(skyflowErr.GetMessage()).To(Equal("Message: Invalid request"))
			Expect(skyflowErr.GetGrpcCode()).To(Equal("3"))
			Expect(skyflowErr.GetRequestId()).To(Equal("req-67890"))
			Expect(skyflowErr.GetHttpStatusCode()).To(Equal("Bad Request"))
		})

		It("should return an error when JSON parsing fails", func() {
			err := errors.New("API Error: {invalid_json}")

			skyflowErr := SkyflowErrorApi(err, header)

			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetMessage()).To(Equal(fmt.Sprintf(`Message: %s`, err.Error())))
		})

		It("should return an error when the message doesn't contain a colon", func() {
			err := errors.New("Invalid API response")

			skyflowErr := SkyflowErrorApi(err, header)

			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetMessage()).To(Equal("Message: Invalid API response"))
		})

		It("should handle an error message without required fields", func() {
			errorJSON := `{"error": {}}`
			err := errors.New("API Error: " + errorJSON)

			skyflowErr := SkyflowErrorApi(err, header)

			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetMessage()).To(Equal("Message: Unknown error"))
			Expect(skyflowErr.GetCode()).To(Equal("Code: "))
			Expect(skyflowErr.GetGrpcCode()).To(BeEmpty())
			Expect(skyflowErr.GetRequestId()).To(Equal("req-67890"))
			Expect(skyflowErr.GetHttpStatusCode()).To(BeEmpty())
		})

		It("should handle an error message where error is a string", func() {
			errorJSON := `{"error": "Something went wrong"}`
			err := errors.New("API Error: " + errorJSON)

			skyflowErr := SkyflowErrorApi(err, header)
			Expect(skyflowErr.GetRequestId()).To(Equal("req-67890"))
			Expect(skyflowErr.GetMessage()).To(Equal("Message: Something went wrong"))
		})

		It("should parse details array when present in SkyflowErrorApi", func() {
			errorJSON := `{"error":{"http_code":400,"message":"Bad Request","grpc_code":3,"http_status":"Bad Request","details":[{"type":"detail1"},{"type":"detail2"}]}}`
			err := errors.New("API Error: " + errorJSON)

			skyflowErr := SkyflowErrorApi(err, header)
			Expect(skyflowErr.GetCode()).To(Equal("Code: 400"))
			Expect(skyflowErr.GetMessage()).To(Equal("Message: Bad Request"))
			Expect(skyflowErr.GetDetails()).To(HaveLen(2))
			Expect(skyflowErr.GetRequestId()).To(Equal("req-67890"))
		})

		It("should use original error message when error field is neither string nor map", func() {
			errorJSON := `{"error": 42}`
			originalErr := errors.New("API Error: " + errorJSON)

			skyflowErr := SkyflowErrorApi(originalErr, header)
			Expect(skyflowErr).ToNot(BeNil())
			Expect(skyflowErr.GetMessage()).To(ContainSubstring(originalErr.Error()))
		})
	})

})
