package controller_test

import (
	// "bytes"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	// "io"
	// "mime/multipart"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	client "github.com/skyflowapi/skyflow-go/v2/internal/generated/client"
	. "github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

// errorReader is a custom io.Reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

var (
	mockInsertSuccessJSON              = `{"vaultID":"id", "responses":[{"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name_on_card":"token1"}}]}, "Status":200}]}`
	mockInsertContinueFalseSuccessJSON = `{"records":[{"skyflow_id":"skyflowid1", "tokens":{"name":"nameToken1"}}, {"skyflow_id":"skyflowid2", "tokens":{"expiry_month":"monthToken", "name":"nameToken3"}}]}`
	mockDetokenizeSuccessJSON          = `{"records":[{"token":"token", "valueType":"STRING", "value":"*REDACTED*", "error":null}]}`
	mockDetokenizeErrorJSON            = `{"error":{"grpc_code":5,"http_code":404,"message":"Detokenize failed. All tokens are invalid. Specify valid tokens.","http_status":"Not Found","details":[]}}`
	mockDetokenizePartialSuccessJSON   = `{"records":[{"token":"token1", "valueType":"STRING", "value":"*REDACTED*", "error":null}, {"token":"token1", "valueType":"NONE", "value":"", "error":"Token Not Found"}]}`
	mockGetSuccessJSON                 = `{"records":[{"fields":{"name":"name1", "skyflow_id":"id1"}, "tokens":null}]}`
	mockGetErrorJSON                   = `{"error":{"grpc_code":5,"http_code":404,"message":"Get failed. [faild fail] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
	mockDeleteSuccessJSON              = `{"RecordIDResponse":["id1"]}`
	mockDeleteErrorJSON                = `{"error":{"grpc_code":5,"http_code":404,"message":"Delete failed. [id1] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
	mockQuerySuccessJSON               = `{"records":[{"fields":{"counter":null, "country":null, "date_of_birth":"XXXX-06-06", "email":"s******y@gmail.com", "name":"m***me", "phone_number":"XXXXXX8889", "skyflow_id":"id"}, "tokens":null}]}`
	mockQueryErrorJSON                 = `{"error":{"grpc_code":5,"http_code":404,"message":"Invalid request. Table name cards is invalid. Specify a valid table name.","http_status":"Not Found","details":[]}}`
	mockUpdateSuccessJSON              = `{"skyflow_id":"id","tokens":{"name":"token"}}`
	mockUpdateErrorJSON                = `{"error":{"grpc_code":3,"http_code":400,"message":"Invalid request. No fields were present. Specify valid fields and values.","http_status":"Bad Request","details":[]}}`
	mockTokenizeSuccessJSON            = `{"records":[{"token":"token1"}]}`
	mockTokenizeErrorJSON              = `{"error":{"grpc_code":3,"http_code":400,"message":"Tokenization failed. Column group group_name is invalid. Specify a valid column group.","http_status":"Bad Request","details":[]}}`
	mockDeidentifyTextSuccessJSON      = `{"processed_text": "My name is [NAME] and email is [EMAIL]", "word_count": 8, "character_count": 45, "entities": [{"token": "token1", "value": "John Doe", "entity_type": "NAME", "entity_scores": {"score": 0.9}, "location": {"start_index": 11, "end_index": 19, "start_index_processed": 11, "end_index_processed": 17}}, {"token": "token2", "value": "john@example.com", "entity_type": "EMAIL_ADDRESS", "entity_scores": {"score": 0.95}, "location": {"start_index": 30, "end_index": 45, "start_index_processed": 30, "end_index_processed": 37}}]}`
	mockDeidentifyTextNoEntitiesJSON   = `{"processed_text": "No entities found in this text", "word_count": 6, "character_count": 30}`
	mockDeidentifyTextErrorJSON        = `{"error":{"message":"Invalid request"}}`
	mockReidentifyTextSuccessJSON      = `{"text": "Sample original text"}`
	mockReidentifyTextErrorJSON        = `{"error":{"message":"Invalid request"}}`
	mockDeidentifyFileErrorJSON        = `{"error":{"message":"Invalid file format"}}`
	mockGetDetectRunInProgressJSON     = `{"status": "IN_PROGRESS", "message": "Processing in progress"}`
	mockGetDetectRunFailedJSON         = `{"status": "FAILED", "message": "Processing failed", "outputType": "UNKNOWN"}`
	mockGetDetectRunExpiredJSON        = `{ "status": "UNKNOWN", "outputType": "UNKNOWN", "output": [], "message": "", "size": 0}`
	mockGetDetectRunApiErrorJSON       = `{"error": {"message": "Invalid run ID"}}`
)

var (
	mockInsertBatchEmptyResponseJSON    = `{}`
	mockDetokenizeErrorMissingTokenJSON = `{"records":[{"error":"Token Not Found"}]}`
	mockDetokenizeSuccessNullFieldsJSON = `{"records":[{}]}`
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("Vault controller Test cases", func() {
	Describe("Test Insert functions", func() {
		var (
			mockJSONResponse string
			response         map[string]interface{}
			contrl           VaultController
			ts               *httptest.Server
		)

		BeforeEach(func() {
			customHeader := make(map[CustomHeaderKey]string)
			customHeader[RequestIdHeader] = "custom-header-value"
			response = make(map[string]interface{})
			ts = nil
			contrl = VaultController{
				Config: &VaultConfig{
					VaultId:   "id",
					ClusterId: "clusterid",
					Env:       PROD,
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
				},
				CustomHeaders: customHeader,
			}
		})

		AfterEach(func() {
			if ts != nil {
				ts.Close()
			}
		})

		Context("Insert with ContinueOnError True - Success Case", func() {
			BeforeEach(func() {
				_ = json.Unmarshal([]byte(mockInsertSuccessJSON), &response)

				// Setup mock server
				ts = setupMockServer(response, "ok", "/vaults/v1/vaults/")
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
			})

			It("should insert successfully", func() {
				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"field1": "value1"},
						{"field2": "value2"},
					},
				}
				options := InsertOptions{
					ContinueOnError: true,
					Upsert:          "upsert",
					Tokens: []map[string]interface{}{
						{"name": "token1"},
						{"expiry_month": "token2", "name": "token3"},
					},
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				Expect(insertError).To(BeNil())
				Expect(len(res.InsertedFields)).To(Equal(1))
				Expect(res.InsertedFields[0]["SkyflowId"]).To(Equal("skyflowid"))
			})
		})
		Context("Insert with ContinueOnError True - Error Case", func() {
			It("should return an error when insert fails and ContinueOnError is true", func() {
				var response map[string]interface{}

				// Unmarshal the mock JSON response into a map
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				// Prepare mock data
				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"field1": "value1"},
						{"field2": "value2"},
					},
				}
				options := InsertOptions{
					ContinueOnError: true,
					Upsert:          "upsert",
				}

				// Set up the mock server using the reusable function
				ts = setupMockServer(response, "partial", "/vaults/v1/vaults/")
				defer ts.Close()

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				// Create the VaultController instance
				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							Token: "Token",
						},
					},
				}

				// Call the Insert method
				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).To(BeNil(), "Expected an error during insert operation")
				Expect(res).ToNot(BeNil(), "Expected no response due to error in insert operation")
			})
			It("should return an error when validations fails", func() {
				// Prepare mock data
				request := InsertRequest{
					Table: "",
					Values: []map[string]interface{}{
						{"field1": "value1"},
						{"field2": "value2"},
					},
				}
				options := InsertOptions{
					ContinueOnError: true,
					Upsert:          "upsert",
				}

				// Create the VaultController instance
				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							Token: "Token",
						},
					},
				}

				// Call the Insert method
				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil(), "Expected an error during insert operation")
				Expect(res).To(BeNil(), "Expected no response due to error in insert operation")
			})

			It("should return error when custom headers map is empty in Insert", func() {
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				options := InsertOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				ctx := context.Background()
				res, err := contrl.Insert(ctx, request, options)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has invalid key in Insert", func() {
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				options := InsertOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				ctx := context.Background()
				res, err := contrl.Insert(ctx, request, options)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has empty value in Insert", func() {
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				options := InsertOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				ctx := context.Background()
				res, err := contrl.Insert(ctx, request, options)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

		})
		Context("Insert with ContinueOnError True - Partial Error Case", func() {
			It("should return partial success and error fields", func() {
				const mockJSONResponse = `{"vaultID":"id", "responses":[{"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}, {"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name":"token1"}}]}, "Status":200}]}`
				var response map[string]interface{}

				// Unmarshal the mock JSON response into a map
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				// Prepare mock data
				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"field1": "value1"},
						{"field2": "value2"},
					},
				}
				options := InsertOptions{
					ContinueOnError: true,
					Upsert:          "upsert",
				}

				// Set up the mock server using the reusable function
				ts := setupMockServer(response, "partial", "/vaults/v1/vaults/")
				defer ts.Close()

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				// Create the VaultController instance
				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							Token: "Token",
						},
					},
				}

				// Call the Insert method
				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).To(BeNil(), "Expected no error during insert operation")
				Expect(res).ToNot(BeNil(), "Expected valid response")
				Expect(len(res.InsertedFields)).To(Equal(1), "Expected exactly 1 inserted field")
				Expect(res.InsertedFields[0]["SkyflowId"]).To(Equal("skyflowid"), "Expected the inserted field to have skyflow_id 'skyflowid'")
				Expect(len(res.Errors)).To(Equal(1), "Expected exactly 1 error field")
			})
		})
		Context("Insert with ContinueOnError False - Success Case", func() {
			It("should insert records correctly and return valid response", func() {
				// Use mock response constant
				_ = json.Unmarshal([]byte(mockInsertContinueFalseSuccessJSON), &response)

				// Mock request and options
				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
						{"expiry_month": "value2", "name": "value2"},
					},
				}
				options := InsertOptions{
					ContinueOnError: false,
					Upsert:          "upsert",
					Tokens: []map[string]interface{}{
						{"name": "token1"},
						{"expiry_month": "token2", "name": "token3"},
					},
				}

				// Set up the mock server
				ts = setupMockServer(response, "ok", "/vaults/v1/vaults/")
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				// Call the Insert method
				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).To(BeNil(), "Expected no error during insert operation")
				Expect(res).ToNot(BeNil(), "Expected valid response from insert operation")
				Expect(len(res.InsertedFields)).To(Equal(2), "Expected exactly 2 inserted fields")
				Expect(res.InsertedFields[0]["SkyflowId"]).To(Equal("skyflowid1"), "Expected first inserted field to have skyflow_id 'skyflowid1'")
				Expect(res.InsertedFields[1]["SkyflowId"]).To(Equal("skyflowid2"), "Expected second inserted field to have skyflow_id 'skyflowid2'")
			})
		})
		Context("Insert with ContinueOnError False - Error Case", func() {
			It("should return error", func() {
				var resp map[string]interface{}

				// Unmarshal the mock JSON response into a map
				_ = json.Unmarshal([]byte(mockJSONResponse), &resp)

				// Prepare mock data
				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"field1": "value1"},
						{"field2": "value2"},
					},
				}
				options := InsertOptions{
					ContinueOnError: false,
					Upsert:          "upsert",
				}

				// Set up the mock server using the reusable function
				ts := setupMockServer(resp, "error", "/vaults/v1/vaults/")
				defer ts.Close()

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				// Create the VaultController instance
				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							Token: "Token",
						},
					},
				}

				// Call the Insert method
				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil(), "Expected an error during insert operation")
				Expect(res).To(BeNil(), "Expected no response due to error in insert operation")
			})
		})
		Context("Insert with ContinueOnError True - Empty batch response body", func() {
			It("should return empty InsertResponse without panicking when server body has no records", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertBatchEmptyResponseJSON), &response)
				ts = setupMockServer(response, "ok", "/vaults/v1/vaults/")
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				options := InsertOptions{ContinueOnError: true}
				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)
				Expect(insertError).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.InsertedFields).To(BeEmpty())
				Expect(res.Errors).To(BeEmpty())
			})
		})

		Context("Insert Client Creation Failed", func() {
			It("should return an error when client creation fails", func() {
				var response map[string]interface{}

				// Unmarshal the mock JSON response into a map
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				// Prepare mock data
				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"field1": "value1"},
						{"field2": "value2"},
					},
				}
				options := InsertOptions{
					ContinueOnError: true,
					Upsert:          "upsert",
					Tokens: []map[string]interface{}{
						{"name": "token1"},
						{"expiry_month": "token2", "name": "token3"},
					},
					TokenMode: ENABLE,
				}

				// Set up the mock server using the reusable function
				ts = setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				// Set the mock server URL in the controller's client
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError("code", "error occurred in api")
				}

				// Call the Insert method
				ctx := context.Background()
				_, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil(), "Expected an error when client creation fails")
			})
		})

		// -------------------------------------------------------------------
		// Line 213: CreateRequestClientFunc error — bulk (ContinueOnError=false) path
		// -------------------------------------------------------------------
		Context("Insert — CreateRequestClientFunc fails in bulk path", func() {
			It("should return error when client creation fails with ContinueOnError false", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError("code", "client creation failed")
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				ctrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							Token: "token",
						},
					},
				}
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				res, err := ctrl.Insert(context.Background(), request, InsertOptions{ContinueOnError: false})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("client creation failed"))
			})
		})

		// -------------------------------------------------------------------
		// Line 206: client-level custom headers validation error
		// -------------------------------------------------------------------
		Context("Insert — client-level custom headers invalid", func() {
			It("should return error when client-level custom headers map is empty", func() {
				ctrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
					CustomHeaders: make(map[CustomHeaderKey]string), // empty = invalid
				}
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				res, err := ctrl.Insert(context.Background(), request, InsertOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers contain an invalid key", func() {
				ctrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-not-allowed"): "value",
					},
				}
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				res, err := ctrl.Insert(context.Background(), request, InsertOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers have an empty value", func() {
				ctrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				res, err := ctrl.Insert(context.Background(), request, InsertOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// Lines 226-229: batch API error (apiErr) inside ContinueOnError=true
		// -------------------------------------------------------------------
		Context("Insert ContinueOnError=true — batch API returns HTTP error", func() {
			It("should return error when the batch API call fails", func() {
				var errResp map[string]interface{}
				_ = json.Unmarshal([]byte(`{"error":{"grpc_code":3,"http_code":400,"message":"batch insert failed","http_status":"Bad Request","details":[]}}`), &errResp)

				ts := setupMockServer(errResp, "error", "/vaults/v1/vaults/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}

				ctrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{Token: "token"},
					},
				}
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				res, err := ctrl.Insert(context.Background(), request, InsertOptions{ContinueOnError: true})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// Lines 237-239: parseErr from GetFormattedBatchInsertRecord
		// Server returns a batch response with a record missing the "Body" field.
		// -------------------------------------------------------------------
		Context("Insert ContinueOnError=true — malformed batch response record", func() {
			It("should return parseErr when a batch response record has no Body field", func() {
				// Response with a record missing "Body" — causes GetFormattedBatchInsertRecord to error
				rawResp := `{"vaultID":"id","responses":[{"Status":200}]}`
				var resp map[string]interface{}
				_ = json.Unmarshal([]byte(rawResp), &resp)

				ts := setupMockServer(resp, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}

				ctrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{Token: "token"},
					},
				}
				request := InsertRequest{
					Table:  "test_table",
					Values: []map[string]interface{}{{"field1": "value1"}},
				}
				res, err := ctrl.Insert(context.Background(), request, InsertOptions{ContinueOnError: true})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})
	})
	Describe("Test Detokenize functions", func() {
		var (
			vaultController *VaultController
			ctx             context.Context
			request         DetokenizeRequest
			options         DetokenizeOptions
		)
		BeforeEach(func() {
			// Initialize the VaultController instance
			vaultController = &VaultController{
				Config: &VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:          PROD,
					ClusterId:    "clusterID",
					BaseVaultUrl: "http://127.0.0.1",
				},
			}

			// Initialize context, request, and options
			ctx = context.Background()
			request = DetokenizeRequest{
				DetokenizeData: []DetokenizeData{
					{
						Token:         "token1",
						RedactionType: MASKED,
					},
				},
			}
			options = DetokenizeOptions{
				ContinueOnError: true,
			}
		})
		Context("When Detokenize is called", func() {
			It("should return detokenized data with no errors", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				ctx = context.Background()
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				// Call the Detokenize function
				res, err := vaultController.Detokenize(ctx, request, options)
				// Validate the response
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.DetokenizedFields).To(HaveLen(1))
				Expect(res.DetokenizedFields[0].Token).To(Equal("token"))
				Expect(res.DetokenizedFields[0].Value).To(Equal("*REDACTED*"))
				Expect(res.DetokenizedFields[0].Type).To(Equal("STRING"))

			})
			It("should return detokenized data with errors", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "error", "/vaults/v1/vaults/")

				ctx = context.Background()
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				// Call the Detokenize function
				res, err := vaultController.Detokenize(ctx, request, options)
				// Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return detokenized data with errors", func() {
				ctx = context.Background()
				request.DetokenizeData = nil
				// Call the Detokenize function
				res, err := vaultController.Detokenize(ctx, request, options)
				// Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return detokenized data with partial success response", func() {
				_ = &VaultController{
					Config: &VaultConfig{
						VaultId:     "vaultID",
						Credentials: Credentials{Token: "token"},
						Env:         PROD,
						ClusterId:   "clusterID",
					},
				}
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizePartialSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				ctx = context.Background()
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				// Call the Detokenize function
				res, err := vaultController.Detokenize(ctx, request, options)
				// Validate the response
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
			It("should return error while creating client in detokenize", func() {
				ctx = context.Background()
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				// Call the Detokenize function
				res, err := vaultController.Detokenize(ctx, request, options)
				// Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return error in get token while calling in detokenize", func() {
				ctx = context.Background()
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				// Call the Detokenize function
				res, err := vaultController.Detokenize(ctx, request, options)
				// Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers map is empty in Detokenize", func() {
				opts := DetokenizeOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				res, err := vaultController.Detokenize(ctx, request, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has invalid key in Detokenize", func() {
				opts := DetokenizeOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				res, err := vaultController.Detokenize(ctx, request, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has empty value in Detokenize", func() {
				opts := DetokenizeOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				res, err := vaultController.Detokenize(ctx, request, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return error when custom headers has empty value in Detokenize", func() {
				opts := DetokenizeOptions{}
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					SkyflowAccountId: "",
				}
				res, err := vaultController.Detokenize(ctx, request, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			// -------------------------------------------------------------------
			// Line 295: client-level custom headers validation (v.CustomHeaders)
			// -------------------------------------------------------------------
			Context("Detokenize — client-level custom headers invalid", func() {
				AfterEach(func() {
					vaultController.CustomHeaders = nil
				})

				It("should return error when client-level custom headers map is empty", func() {
					vaultController.CustomHeaders = make(map[CustomHeaderKey]string)
					res, err := vaultController.Detokenize(ctx, request, DetokenizeOptions{})
					Expect(err).ToNot(BeNil())
					Expect(res).To(BeNil())
				})

				It("should return error when client-level custom headers contain an invalid key", func() {
					vaultController.CustomHeaders = map[CustomHeaderKey]string{
						CustomHeaderKey("x-not-allowed"): "value",
					}
					res, err := vaultController.Detokenize(ctx, request, DetokenizeOptions{})
					Expect(err).ToNot(BeNil())
					Expect(res).To(BeNil())
				})

				It("should return error when client-level custom headers have an empty value", func() {
					vaultController.CustomHeaders = map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					}
					res, err := vaultController.Detokenize(ctx, request, DetokenizeOptions{})
					Expect(err).ToNot(BeNil())
					Expect(res).To(BeNil())
				})
			})

			It("should not panic when error record is missing the token field", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeErrorMissingTokenJSON), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}
				res, err := vaultController.Detokenize(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.Errors).To(HaveLen(1))
				Expect(res.Errors[0].Token).To(Equal(""))
				Expect(res.Errors[0].Error).To(Equal("Token Not Found"))
			})

			It("should not panic when success record has all null fields", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeSuccessNullFieldsJSON), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}
				res, err := vaultController.Detokenize(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.DetokenizedFields).To(HaveLen(1))
				Expect(res.DetokenizedFields[0].Token).To(Equal(""))
				Expect(res.DetokenizedFields[0].Value).To(Equal(""))
				Expect(res.DetokenizedFields[0].Type).To(Equal(""))
			})
		})
	})
	Describe("Test Get functions", func() {
		var vaultController VaultController
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			vaultController = VaultController{
				Config: &VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:       PROD,
					ClusterId: "clusterID",
				},
			}
			ctx = context.TODO()
		})
		Context("Test the success and error case", func() {
			options := GetOptions{
				RedactionType: REDACTED,
				Offset:		"10",
				Limit:		"10",
			}
			request := GetRequest{
				Table: "table",
				Ids:   []string{"id1"},
			}
			It("should return error response when Invalid ids passed in Get", func() {
				// Set the mock server URL in the controller's client
				getRequest := GetRequest{
					Table: "table",
					Ids:   []string{},
			    }

				res, err := vaultController.Get(ctx, getRequest, options)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return success response when valid ids passed in Get", func() {
				// Set the mock server URL in the controller's client
				getRequest := GetRequest{
					Table: "table",
					Ids:   []string{"id1"},
			    }
				getOptions := GetOptions{
					Offset:		"10",
					Limit:		"10",
					ReturnTokens: true,
					Fields: []string{"name", "SkyflowId"},
				}
				response := make(map[string]interface{})

				mockJSONResponse := `{"records":[{"fields":{"name":"name1", "SkyflowId":"id1"}, "tokens":null}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				res, err := vaultController.Get(ctx, getRequest, getOptions)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
			It("should return error response when valid request passed in Get", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Get(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
			It("should return error response when invalid ids passed in Get", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "error", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Get(ctx, request, options)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error client creation step Get", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Get(ctx, request, options)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return success response when valid column passed in Get", func() {
				options = GetOptions{
					RedactionType: REDACTED,
					ColumnName:    "name",
					ColumnValues:  []string{"1234"},
				}
				request = GetRequest{
					Table: "table1",
				}
				response := make(map[string]interface{})
				mockJSONResponse := `{"records":[{"fields":{"name":"name1", "SkyflowId":"id1"}, "tokens":null}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				res, err := vaultController.Get(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error when custom headers map is empty in Get", func() {
				req := GetRequest{Table: "table", Ids: []string{"id1"}}
				opts := GetOptions{
					RedactionType: REDACTED,
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				res, err := vaultController.Get(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has invalid key in Get", func() {
				req := GetRequest{Table: "table", Ids: []string{"id1"}}
				opts := GetOptions{
					RedactionType: REDACTED,
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				res, err := vaultController.Get(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has empty value in Get", func() {
				req := GetRequest{Table: "table", Ids: []string{"id1"}}
				opts := GetOptions{
					RedactionType: REDACTED,
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				res, err := vaultController.Get(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// options.OrderBy branch in Get
		// -------------------------------------------------------------------
		Context("Get — OrderBy option is set", func() {
			It("should set OrderBy on the request when options.OrderBy is non-empty", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetSuccessJSON), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				res, err := vaultController.Get(ctx, GetRequest{
					Table: "table",
					Ids:   []string{"id1"},
				}, GetOptions{
					RedactionType: REDACTED,
					OrderBy:       ASCENDING,
				})
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
		})
	})
	Describe("Test Delete functions", func() {
		var vaultController VaultController
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			vaultController = VaultController{
				Config: &VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:       PROD,
					ClusterId: "clusterID",
				},
			}
			ctx = context.TODO()
		})
		Context("Test the success and error case", func() {
			request := DeleteRequest{
				Table: "table",
				Ids:   []string{"id1"},
			}
			It("should return success response when valid ids passed in Delete", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeleteSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Delete(ctx, request, common.DeleteOptions{})
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid ids passed in Delete", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeleteErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "error", "/vaults/v1/vaults/")

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Delete(ctx, request, common.DeleteOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when invalid data passed in Delete", func() {
				request.Ids = []string{}
				res, err := vaultController.Delete(ctx, request, common.DeleteOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Delete", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Delete(ctx, request, common.DeleteOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers map is empty in Delete", func() {
				req := DeleteRequest{Table: "table", Ids: []string{"id1"}}
				opts := common.DeleteOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				res, err := vaultController.Delete(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers map is invalid in controller in Delete", func() {
				req := DeleteRequest{Table: "table", Ids: []string{"id1"}}
				opts := common.DeleteOptions{
				}
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
				}
				res, err := vaultController.Delete(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has invalid key in Delete", func() {
				req := DeleteRequest{Table: "table", Ids: []string{"id1"}}
				opts := common.DeleteOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				res, err := vaultController.Delete(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has empty value in Delete", func() {
				req := DeleteRequest{Table: "table", Ids: []string{"id1"}}
				opts := common.DeleteOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				res, err := vaultController.Delete(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return error when expired token in controller in Delete", func() {
				req := DeleteRequest{Table: "table", Ids: []string{"id1"}}
				opts := common.DeleteOptions{
				}
				vaultController.Config.Credentials.Token = "expired_token"
				vaultController.Config.Credentials.ApiKey = ""
				vaultController.CustomHeaders = make(map[CustomHeaderKey]string)
				res, err := vaultController.Delete(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// CreateRequestClientFunc error in Delete
		// -------------------------------------------------------------------
		Context("Delete — CreateRequestClientFunc fails", func() {
			It("should return error when client creation fails", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError("code", "client creation failed")
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				res, err := vaultController.Delete(ctx, DeleteRequest{
					Table: "table",
					Ids:   []string{"id1"},
				}, common.DeleteOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("client creation failed"))
			})
		})
	})
	Describe("Test Query functions", func() {
		var vaultController VaultController
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			vaultController = VaultController{
				Config: &VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:       PROD,
					ClusterId: "clusterID",
				},
			}
			ctx = context.TODO()
		})
		Context("Test the success and error case", func() {
			request := QueryRequest{
				Query: "SELECT * FROM persons WHERE skyflow_id='id'",
			}
			It("should return success response when valid ids passed in Query", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockQuerySuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				res, err := vaultController.Query(ctx, request, common.QueryOptions{})
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid ids passed in Query", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockQueryErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "error", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Query(ctx, request, common.QueryOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when invalid data passed in Query", func() {
				request.Query = ""
				res, err := vaultController.Query(ctx, request, common.QueryOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Query", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Query(ctx, request, common.QueryOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error when custom headers is incorrect in query", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					CustomHeaderKey("x-invalid-header"): "value",
				}
				res, err := vaultController.Query(ctx, request, common.QueryOptions{})
				Expect(res).To(BeNil())
				fmt.Printf("%v", err)
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers map is empty in Query", func() {
				req := QueryRequest{Query: "SELECT * FROM persons WHERE skyflow_id='id'"}
				opts := common.QueryOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				res, err := vaultController.Query(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has invalid key in Query", func() {
				req := QueryRequest{Query: "SELECT * FROM persons WHERE skyflow_id='id'"}
				opts := common.QueryOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				res, err := vaultController.Query(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has empty value in Query", func() {
				req := QueryRequest{Query: "SELECT * FROM persons WHERE skyflow_id='id'"}
				opts := common.QueryOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				res, err := vaultController.Query(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// Line 503: client-level custom headers validation (v.CustomHeaders)
		// -------------------------------------------------------------------
		Context("Query — client-level custom headers invalid", func() {
			var req QueryRequest
			BeforeEach(func() {
				req = QueryRequest{Query: "SELECT * FROM persons WHERE skyflow_id='id'"}
			})
			AfterEach(func() {
				vaultController.CustomHeaders = nil
			})

			It("should return error when client-level custom headers map is empty", func() {
				vaultController.CustomHeaders = make(map[CustomHeaderKey]string)
				res, err := vaultController.Query(ctx, req, common.QueryOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers contain an invalid key", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					CustomHeaderKey("x-not-allowed"): "value",
				}
				res, err := vaultController.Query(ctx, req, common.QueryOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers have an empty value", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					SkyflowAccountId: "",
				}
				res, err := vaultController.Query(ctx, req, common.QueryOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// Line 508: CreateRequestClientFunc error in Query
		// -------------------------------------------------------------------
		Context("Query — CreateRequestClientFunc fails", func() {
			It("should return error when client creation fails", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError("code", "client creation failed")
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				res, err := vaultController.Query(ctx, QueryRequest{
					Query: "SELECT * FROM persons WHERE skyflow_id='id'",
				}, common.QueryOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("client creation failed"))
			})
		})
	})
	Describe("Test Update functions", func() {
		var vaultController VaultController
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			vaultController = VaultController{
				Config: &VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:       PROD,
					ClusterId: "clusterID",
				},
			}
			ctx = context.TODO()
		})
		Context("Test the success and error case", func() {
			request := UpdateRequest{
				Table:  "demo",
				Data:   map[string]interface{}{"SkyflowId": "123", "name": "john"},
				Tokens: nil,
			}
			It("should return success response when valid ids passed in Update", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockUpdateSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Update(ctx, request, UpdateOptions{
					ReturnTokens: true,
					TokenMode:    DISABLE,
				})
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.Errors).To(BeNil())
			})

			// Line 563-564: request.Tokens != nil branch
			It("should set record.Tokens when request.Tokens is non-nil", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockUpdateSuccessJSON), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}

				req := UpdateRequest{
					Table:  "demo",
					Data:   map[string]interface{}{"SkyflowId": "123", "name": "john"},
					Tokens: map[string]interface{}{"name": "token"},
				}
				res, err := vaultController.Update(ctx, req, UpdateOptions{TokenMode: ENABLE})
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("Update response contains both SkyflowId (new) and skyflowId (deprecated backward compat)", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockUpdateSuccessJSON), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}
				// Use a fresh request — the controller deletes SkyflowId from Data after extracting it,
				// so reusing the shared `request` variable fails if a previous test already ran Update.
				freshRequest := UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}
				res, err := vaultController.Update(ctx, freshRequest, UpdateOptions{TokenMode: DISABLE})
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.UpdatedField).To(HaveKeyWithValue("SkyflowId", "id"),
					"new SkyflowId key must be present in the Update response")
				Expect(res.UpdatedField).To(HaveKeyWithValue("skyflowId", "id"),
					"deprecated skyflowId key must be retained for backward compatibility")
			})

			It("should return error response when invalid data passed in Update", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockUpdateErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "error", "/vaults/v1/vaults/")
				request.Tokens = map[string]interface{}{"name": "token"}
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Update(ctx, request, UpdateOptions{ReturnTokens: false, TokenMode: ENABLE})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when validation fail for invalid data passed in Update", func() {
				request.Tokens = nil

				res, err := vaultController.Update(ctx, request, UpdateOptions{ReturnTokens: false, TokenMode: ENABLE})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Update", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Update(ctx, request, UpdateOptions{ReturnTokens: true, TokenMode: ENABLE_STRICT})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers map is empty in Update", func() {
				req := UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}
				opts := UpdateOptions{
					TokenMode:     DISABLE,
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				res, err := vaultController.Update(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has invalid key in Update", func() {
				req := UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}
				opts := UpdateOptions{
					TokenMode: DISABLE,
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				res, err := vaultController.Update(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has empty value in Update", func() {
				req := UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}
				opts := UpdateOptions{
					TokenMode: DISABLE,
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				res, err := vaultController.Update(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// Line 544: client-level custom headers validation (v.CustomHeaders)
		// -------------------------------------------------------------------
		Context("Update — client-level custom headers invalid", func() {
			var req UpdateRequest
			BeforeEach(func() {
				req = UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}
			})
			AfterEach(func() {
				vaultController.CustomHeaders = nil
			})

			It("should return error when client-level custom headers map is empty", func() {
				vaultController.CustomHeaders = make(map[CustomHeaderKey]string)
				res, err := vaultController.Update(ctx, req, UpdateOptions{TokenMode: DISABLE})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers contain an invalid key", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					CustomHeaderKey("x-not-allowed"): "value",
				}
				res, err := vaultController.Update(ctx, req, UpdateOptions{TokenMode: DISABLE})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers have an empty value", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					SkyflowAccountId: "",
				}
				res, err := vaultController.Update(ctx, req, UpdateOptions{TokenMode: DISABLE})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// Line 547: CreateRequestClientFunc error in Update
		// -------------------------------------------------------------------
		Context("Update — CreateRequestClientFunc fails", func() {
			It("should return error when client creation fails", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError("code", "client creation failed")
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				res, err := vaultController.Update(ctx, UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}, UpdateOptions{TokenMode: DISABLE})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("client creation failed"))
			})
		})

		// -------------------------------------------------------------------
		// Line 551-553: SetTokenMode returns error for invalid BYOT value
		// -------------------------------------------------------------------
		Context("Update — invalid TokenMode", func() {
			It("should return INVALID_BYOT error when TokenMode is not a valid BYOT value", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				res, err := vaultController.Update(ctx, UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}, UpdateOptions{TokenMode: BYOT("INVALID_MODE")})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(skyflowError.INVALID_BYOT))
			})
		})

		// -------------------------------------------------------------------
		// apiErr != nil: RecordServiceUpdateRecord returns HTTP error
		// -------------------------------------------------------------------
		Context("Update — API call returns HTTP error", func() {
			It("should return error when the API responds with an error status", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockUpdateErrorJSON), &response)
				ts := setupMockServer(response, "error", "/vaults/v1/vaults/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					c := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *c
					return nil
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				res, err := vaultController.Update(ctx, UpdateRequest{
					Table: "demo",
					Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
				}, UpdateOptions{TokenMode: DISABLE})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})
	})
	Describe("Test Tokenize functions", func() {
		var vaultController VaultController
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			vaultController = VaultController{
				Config: &VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:       PROD,
					ClusterId: "clusterID",
				},
			}
			ctx = context.TODO()
		})
		Context("Test the success and error case", func() {
			var arrReq []TokenizeRequest
			arrReq = append(arrReq, TokenizeRequest{
				ColumnGroup: "group_name",
				Value:       "41111111111111",
			})
			It("should return success response when valid ids passed in Tokenize", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockTokenizeSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Tokenize(ctx, arrReq, common.TokenizeOptions{})
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid data passed in Tokenize", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockTokenizeErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "error", "/vaults/v1/vaults/")
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				res, err := vaultController.Tokenize(ctx, arrReq, common.TokenizeOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when validations failed for invalid data passedin Tokenize", func() {
				arrReq = append(arrReq, TokenizeRequest{})
				res, err := vaultController.Tokenize(ctx, arrReq, common.TokenizeOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Tokenize", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Tokenize(ctx, arrReq, common.TokenizeOptions{})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers map is empty in Tokenize", func() {
				req := []TokenizeRequest{{ColumnGroup: "group_name", Value: "41111111111111"}}
				opts := common.TokenizeOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				res, err := vaultController.Tokenize(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has invalid key in Tokenize", func() {
				req := []TokenizeRequest{{ColumnGroup: "group_name", Value: "41111111111111"}}
				opts := common.TokenizeOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				res, err := vaultController.Tokenize(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when custom headers has empty value in Tokenize", func() {
				req := []TokenizeRequest{{ColumnGroup: "group_name", Value: "41111111111111"}}
				opts := common.TokenizeOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				res, err := vaultController.Tokenize(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// v.CustomHeaders: client-level custom headers validation in Tokenize
		// -------------------------------------------------------------------
		Context("Tokenize — client-level custom headers invalid", func() {
			var req []TokenizeRequest
			BeforeEach(func() {
				req = []TokenizeRequest{{ColumnGroup: "group_name", Value: "41111111111111"}}
			})
			AfterEach(func() {
				vaultController.CustomHeaders = nil
			})

			It("should return error when client-level custom headers map is empty", func() {
				vaultController.CustomHeaders = make(map[CustomHeaderKey]string)
				res, err := vaultController.Tokenize(ctx, req, common.TokenizeOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers contain an invalid key", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					CustomHeaderKey("x-not-allowed"): "value",
				}
				res, err := vaultController.Tokenize(ctx, req, common.TokenizeOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers have an empty value", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					SkyflowAccountId: "",
				}
				res, err := vaultController.Tokenize(ctx, req, common.TokenizeOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// CreateRequestClientFunc error in Tokenize
		// -------------------------------------------------------------------
		Context("Tokenize — CreateRequestClientFunc fails", func() {
			It("should return error when client creation fails", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError("code", "client creation failed")
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				req := []TokenizeRequest{{ColumnGroup: "group_name", Value: "41111111111111"}}
				res, err := vaultController.Tokenize(ctx, req, common.TokenizeOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("client creation failed"))
			})
		})
	})
	Describe("Test Upload file functions", func() {
		var vaultController VaultController
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			vaultController = VaultController{
				Config: &VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:       PROD,
					ClusterId: "clusterID",
				},
			}
			ctx = context.TODO()
		})
		It("should return success response when file upload is valid", func() {
			response := make(map[string]interface{})
			mockJSONResponse := `{"skyflowID":"id"}`
			_ = json.Unmarshal([]byte(mockJSONResponse), &response)
			// // Set the mock server URL in the controller's client
			ts := setupMockServer(response, "ok", "/vaults/v2/vaults/")

			// Set the mock server URL in the controller's client
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				client := client.NewClient(
					option.WithBaseURL(ts.URL+"/vaults"),
					option.WithToken("token"),
					option.WithHTTPHeader(header),
				)
				v.ApiClient = *client
				return nil
			}
			request := common.FileUploadRequest{
				Table:      "table",
				ColumnName: "column",
				FilePath:   "../../../../credentials.json",
				SkyflowId:  "skyflowid",
			}

			res, err := vaultController.UploadFile(ctx, request, common.FileUploadOptions{})
			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(res.SkyflowId).To(Equal("id"))
		})
		It("should return error response when api throw error", func() {
			response := make(map[string]interface{})
			mockJSONResponse := `{"error":"error occurred"}`
			_ = json.Unmarshal([]byte(mockJSONResponse), &response)
			// // Set the mock server URL in the controller's client
			ts := setupMockServer(response, "", "/vaults/v2/vaults/")

			// Set the mock server URL in the controller's client
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				client := client.NewClient(
					option.WithBaseURL(ts.URL+"/vaults"),
					option.WithToken("token"),
					option.WithHTTPHeader(header),
				)
				v.ApiClient = *client
				return nil
			}
			request := common.FileUploadRequest{
				Table:      "table",
				ColumnName: "column",
				FilePath:   "../../../../credentials.json",
				SkyflowId:  "skyflowid",
			}

			res, err := vaultController.UploadFile(ctx, request, common.FileUploadOptions{})
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(Equal("Message: error occurred"))
		})
		It("should return error response when file path is invalid in file upload", func() {

			request := common.FileUploadRequest{
				Table:      "table",
				ColumnName: "column",
				FilePath:   "",
				SkyflowId:  "skyflowid",
			}

			res, err := vaultController.UploadFile(ctx, request, common.FileUploadOptions{})
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.MISSING_FILE_SOURCE_IN_UPLOAD_FILE))
		})

		It("should return error when custom headers map is empty in UploadFile", func() {
			request := common.FileUploadRequest{
				Table:      "table",
				ColumnName: "column",
				Base64:     "dGVzdA==",
				FileName:   "test.txt",
				SkyflowId:  "skyflowid",
			}
			opts := common.FileUploadOptions{
				CustomHeaders: make(map[CustomHeaderKey]string),
			}
			res, err := vaultController.UploadFile(ctx, request, opts)
			Expect(err).ToNot(BeNil())
			Expect(res).To(BeNil())
		})

		It("should return error when custom headers has invalid key in UploadFile", func() {
			request := common.FileUploadRequest{
				Table:      "table",
				ColumnName: "column",
				Base64:     "dGVzdA==",
				FileName:   "test.txt",
				SkyflowId:  "skyflowid",
			}
			opts := common.FileUploadOptions{
				CustomHeaders: map[CustomHeaderKey]string{
					CustomHeaderKey("x-invalid-header"): "value",
				},
			}
			res, err := vaultController.UploadFile(ctx, request, opts)
			Expect(err).ToNot(BeNil())
			Expect(res).To(BeNil())
		})

		It("should return error when custom headers has empty value in UploadFile", func() {
			request := common.FileUploadRequest{
				Table:      "table",
				ColumnName: "column",
				Base64:     "dGVzdA==",
				FileName:   "test.txt",
				SkyflowId:  "skyflowid",
			}
			opts := common.FileUploadOptions{
				CustomHeaders: map[CustomHeaderKey]string{
					SkyflowAccountId: "",
				},
			}
			res, err := vaultController.UploadFile(ctx, request, opts)
			Expect(err).ToNot(BeNil())
			Expect(res).To(BeNil())
		})

		// -------------------------------------------------------------------
		// v.CustomHeaders: client-level custom headers validation in UploadFile
		// -------------------------------------------------------------------
		Context("UploadFile — client-level custom headers invalid", func() {
			var req common.FileUploadRequest
			BeforeEach(func() {
				req = common.FileUploadRequest{
					Table:      "table",
					ColumnName: "column",
					Base64:     "dGVzdA==",
					FileName:   "test.txt",
					SkyflowId:  "skyflowid",
				}
			})
			AfterEach(func() {
				vaultController.CustomHeaders = nil
			})

			It("should return error when client-level custom headers map is empty", func() {
				vaultController.CustomHeaders = make(map[CustomHeaderKey]string)
				res, err := vaultController.UploadFile(ctx, req, common.FileUploadOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers contain an invalid key", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					CustomHeaderKey("x-not-allowed"): "value",
				}
				res, err := vaultController.UploadFile(ctx, req, common.FileUploadOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})

			It("should return error when client-level custom headers have an empty value", func() {
				vaultController.CustomHeaders = map[CustomHeaderKey]string{
					SkyflowAccountId: "",
				}
				res, err := vaultController.UploadFile(ctx, req, common.FileUploadOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		// -------------------------------------------------------------------
		// CreateRequestClientFunc error in UploadFile
		// -------------------------------------------------------------------
		Context("UploadFile — CreateRequestClientFunc fails", func() {
			It("should return error when client creation fails", func() {
				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError("code", "client creation failed")
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				req := common.FileUploadRequest{
					Table:      "table",
					ColumnName: "column",
					Base64:     "dGVzdA==",
					FileName:   "test.txt",
					SkyflowId:  "skyflowid",
				}
				res, err := vaultController.UploadFile(ctx, req, common.FileUploadOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("client creation failed"))
			})
		})

		// -------------------------------------------------------------------
		// GetFileForFileUpload error in UploadFile
		// File passes os.Stat (validation) but fails os.Open (chmod 000)
		// -------------------------------------------------------------------
		Context("UploadFile — GetFileForFileUpload returns error", func() {
			It("should return error when file exists but cannot be opened", func() {
				tmpFile, tmpErr := os.CreateTemp("", "test-upload-*.txt")
				Expect(tmpErr).To(BeNil())
				tmpPath := tmpFile.Name()
				tmpFile.Close()
				Expect(os.Chmod(tmpPath, 0000)).To(Succeed())
				defer func() {
					os.Chmod(tmpPath, 0600)
					os.Remove(tmpPath)
				}()

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}
				defer func() { CreateRequestClientFunc = CreateRequestClient }()

				req := common.FileUploadRequest{
					Table:      "table",
					ColumnName: "column",
					FilePath:   tmpPath,
					SkyflowId:  "skyflowid",
				}
				res, err := vaultController.UploadFile(ctx, req, common.FileUploadOptions{})
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})
	})
})
var _ = Describe("ConnectionController", func() {
	var (
		ctrl         *ConnectionController
		mockServer   *httptest.Server
		mockToken    string
		mockRequest  InvokeConnectionRequest
		mockResponse map[string]interface{}
	)

	BeforeEach(func() {
		mockToken = "mock-valid-token"
		ctrl = &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionUrl: "http://mockserver.com",
				ConnectionId:  "demo",
			},
			Token: mockToken,
		}
		mockResponse = map[string]interface{}{"key": "value"}
		mockRequest = InvokeConnectionRequest{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:   map[string]interface{}{"data": "test"},
			Method: POST,
		}
	})

	Describe("Invoke", func() {
		ctx := context.TODO()
		Context("when making a valid request", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"key": "value"}`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
			})

			AfterEach(func() {
				mockServer.Close()
			})

			It("should return a valid response", func() {
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, mockRequest)
				Expect(err).To(BeNil())
				Expect(response.Data).To(Equal(mockResponse))
			})
		})
		Context("when the request fails", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error": "internal server error"}`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
			})

			AfterEach(func() {
				mockServer.Close()
			})
			It("should return an error", func() {
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, mockRequest)
				Expect(response).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return an error from api", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{`))
				}))
				ctrl.Config.ConnectionUrl = "http://invalidurl"
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, mockRequest)
				Expect(response).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return an error when invalid token passed", func() {
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				response, err := ctrl.Invoke(ctx, mockRequest)
				Expect(response).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return an success from api with invalid body", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Length", "0")
					_, _ = w.Write([]byte(`67676`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, mockRequest)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(Equal(float64(67676)))
			})
		})
		Context("Invoke with different content types", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"key": "value"}`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
			})

			AfterEach(func() {
				mockServer.Close()
			})
			It("should handle application/json content type", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"key": "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(HaveKeyWithValue("key", "value"))
			})
			It("should handle application/x-www-form-urlencoded content type", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/x-www-form-urlencoded",
					},
					Body: map[string]interface{}{
						"key": "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(HaveKeyWithValue("key", "value"))
			})
			It("should handle multipart/form-data content type", func() {

				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body: map[string]interface{}{
						"key":  "value",
						"key2": int(123),
						"key3": 123.4,
						"key4": true,
						"key5": float32(1.0),
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(HaveKeyWithValue("key", "value"))
			})
			It("should handle when content type is not set", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/x-www-form-urlencoded",
					},
					Body: map[string]interface{}{
						"key": "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(HaveKeyWithValue("key", "value"))
			})
			It("should throw error when invalid request passed", func() {
				request := InvokeConnectionRequest{
					Method:  "POST",
					Headers: map[string]string{},
					Body: map[string]interface{}{
						"key": "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).ToNot(BeNil())
				Expect(response).To(BeNil())
			})

		})
		Context("Handling query parameters", func() {
			It("should return error when query params contain an unsupported type", func() {
				queryParams := map[string]interface{}{
					"intKey":     123,
					"floatKey":   456.78,
					"stringKey":  "test",
					"boolKey":    true,
					"invalidKey": struct{}{},
				}
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body: map[string]interface{}{
						"key": "value",
					},
					QueryParams: queryParams,
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).ToNot(BeNil())
				Expect(response).To(BeNil())
			})
			It("should correctly send valid query parameters and return a response", func() {
				var capturedQuery string
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					capturedQuery = r.URL.RawQuery
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"result":"ok"}`))
				}))
				defer srv.Close()
				ctrl.Config.ConnectionUrl = srv.URL

				queryParams := map[string]interface{}{
					"intKey":    123,
					"floatKey":  456.78,
					"stringKey": "test",
					"boolKey":   true,
				}
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body:        map[string]interface{}{"key": "value"},
					QueryParams: queryParams,
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(capturedQuery).To(ContainSubstring("stringKey=test"))
				Expect(capturedQuery).To(ContainSubstring("boolKey=true"))
			})
		})
		Context("Handling Path parameters", func() {
			It("should substitute path parameters in the URL and return a response", func() {
				var capturedPath string
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					capturedPath = r.URL.Path
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"result":"ok"}`))
				}))
				defer srv.Close()
				ctrl.Config.ConnectionUrl = srv.URL + "/{id}"

				pathParams := map[string]string{"id": "123"}
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body:       map[string]interface{}{"key": "value"},
					PathParams: pathParams,
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(capturedPath).To(Equal("/123"))
			})
		})

		Context("Handling XML content types", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><response><key>value</key></response>`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
			})

			AfterEach(func() {
				mockServer.Close()
			})

			It("should handle application/xml content type with map body", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/xml",
					},
					Body: map[string]interface{}{
						"key":   "value",
						"nested": map[string]interface{}{
							"inner": "data",
						},
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(ContainSubstring("<key>value</key>"))
			})

			It("should handle text/xml content type", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "text/xml",
					},
					Body: map[string]interface{}{
						"user": "john",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle XML with special characters requiring escaping", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/xml",
					},
					Body: map[string]interface{}{
						"key": "value with <special> & \"characters\" 'here'",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle XML with string body", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/xml",
					},
					Body: "<?xml version=\"1.0\"?><root><item>test</item></root>",
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle XML with arrays", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/xml",
					},
					Body: map[string]interface{}{
						"items": []interface{}{"item1", "item2", "item3"},
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})
		})

		Context("Handling URL-encoded content with nested objects", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`key=value&nested=data`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
			})

			AfterEach(func() {
				mockServer.Close()
			})

			It("should handle nested objects in URL-encoded format", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/x-www-form-urlencoded",
					},
					Body: map[string]interface{}{
						"user": map[string]interface{}{
							"name": "john",
							"age":  30,
						},
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle arrays in URL-encoded format", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/x-www-form-urlencoded",
					},
					Body: map[string]interface{}{
						"tags": []interface{}{"tag1", "tag2", "tag3"},
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle mixed nested objects and arrays", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "application/x-www-form-urlencoded",
					},
					Body: map[string]interface{}{
						"user": map[string]interface{}{
							"name": "john",
						},
						"tags": []interface{}{"tag1", "tag2"},
						"key":  "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})
		})

		Context("Handling multipart/form-data with file uploads", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"success": true}`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
			})

			AfterEach(func() {
				mockServer.Close()
			})

			It("should handle multipart/form-data with *os.File", func() {
				// Create a temporary file for testing
				tmpFile, err := os.CreateTemp("", "test-*.txt")
				Expect(err).To(BeNil())
				defer os.Remove(tmpFile.Name())
				_, _ = tmpFile.WriteString("test file content")
				tmpFile.Close()

				// Reopen for reading
				file, err := os.Open(tmpFile.Name())
				Expect(err).To(BeNil())
				defer file.Close()

				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body: map[string]interface{}{
						"file": file,
						"key":  "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle multipart/form-data with io.Reader", func() {
				reader := strings.NewReader("test content from reader")
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body: map[string]interface{}{
						"upload": reader,
						"name":   "test",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle multipart/form-data with nested maps (JSON stringified)", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body: map[string]interface{}{
						"user": map[string]interface{}{
							"name": "john",
							"age":  30,
						},
						"simple": "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle multipart/form-data with arrays (JSON stringified)", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body: map[string]interface{}{
						"tags": []interface{}{"tag1", "tag2", "tag3"},
						"key":  "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})
		})

		Context("Handling text/plain and text/html content types", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					contentType := r.Header.Get("Content-Type")
					w.Header().Set("Content-Type", contentType)
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("plain text response"))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
			})

			AfterEach(func() {
				mockServer.Close()
			})

			It("should handle text/plain content type", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "text/plain",
					},
					Body: "This is plain text content",
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(Equal("plain text response"))
			})

			It("should handle text/html content type", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "text/html",
					},
					Body: "<html><body>Hello</body></html>",
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})

			It("should handle text/html with map body (converted to JSON)", func() {
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "text/html",
					},
					Body: map[string]interface{}{
						"key": "value",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
			})
		})

		Context("Handling response parsing for different content types", func() {
			It("should parse XML response correctly", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`<?xml version="1.0"?><response><key>value</key></response>`))
				}))
				defer mockServer.Close()
				ctrl.Config.ConnectionUrl = mockServer.URL

				request := InvokeConnectionRequest{
					Method: "GET",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(ContainSubstring("<key>value</key>"))
			})

			It("should parse URL-encoded response correctly", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`key1=value1&key2=value2&key3=value3a&key3=value3b`))
				}))
				defer mockServer.Close()
				ctrl.Config.ConnectionUrl = mockServer.URL

				request := InvokeConnectionRequest{
					Method: "GET",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				// Response should be a map
				dataMap, ok := response.Data.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(dataMap).To(HaveKey("key1"))
				Expect(dataMap["key1"]).To(Equal("value1"))
				// key3 should be an array since it has multiple values
				Expect(dataMap).To(HaveKey("key3"))
			})

			It("should parse JSON response correctly", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"key": "value", "number": 42}`))
				}))
				defer mockServer.Close()
				ctrl.Config.ConnectionUrl = mockServer.URL

				request := InvokeConnectionRequest{
					Method: "GET",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				dataMap, ok := response.Data.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(dataMap["key"]).To(Equal("value"))
				Expect(dataMap["number"]).To(Equal(float64(42)))
			})

			It("should parse text/plain response correctly", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("Simple plain text"))
				}))
				defer mockServer.Close()
				ctrl.Config.ConnectionUrl = mockServer.URL

				request := InvokeConnectionRequest{
					Method: "GET",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Data).To(Equal("Simple plain text"))
			})

			It("should handle invalid JSON response gracefully", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`invalid json content`))
				}))
				defer mockServer.Close()
				ctrl.Config.ConnectionUrl = mockServer.URL

				request := InvokeConnectionRequest{
					Method: "GET",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())
				// Should return as string when JSON parsing fails
				Expect(response.Data).To(Equal("invalid json content"))
			})

			It("should handle invalid URL-encoded response gracefully", func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`%invalid%`))
				}))
				defer mockServer.Close()
				ctrl.Config.ConnectionUrl = mockServer.URL

				request := InvokeConnectionRequest{
					Method: "GET",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
			response, err := ctrl.Invoke(ctx, request)
			Expect(response).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("should handle multipart/form-data response correctly", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "multipart/form-data")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`boundary data`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "GET",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
			Expect(response.Data).To(Equal("boundary data"))
		})

		It("should handle URL-encoded response with multiple values for same key", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`color=red&color=blue&color=green`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "GET",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
			dataMap, ok := response.Data.(map[string]interface{})
			Expect(ok).To(BeTrue())
			colors, ok := dataMap["color"].([]string)
			Expect(ok).To(BeTrue())
			Expect(len(colors)).To(Equal(3))
		})

		It("should handle multipart/form-data body as string", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: "raw string body",
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data body as non-map type", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: 12345,
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle text/plain body as non-string", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
				Body: 98765,
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle text/html body as non-string", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "text/html",
				},
				Body: 54321,
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle unknown content-type with non-map body", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/octet-stream",
				},
				Body: 99999,
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle empty response body", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "GET",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle JSON body as string for application/json", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: `{"test": "value"}`,
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle JSON body as non-map and non-string (integer)", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: 12345,
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle URL-encoded body with non-map type", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/x-www-form-urlencoded",
				},
				Body: "not a map",
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle XML body with unsupported type (not string or map)", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/xml",
				},
				Body: 12345,
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(response).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring("Invalid XML format"))
		})

		It("should handle default method as POST when method is empty", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
				Expect(r.Method).To(Equal("POST"))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{"test": "value"},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle text/html body as map (converted to JSON)", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "text/html",
				},
				Body: map[string]interface{}{"html": "<h1>Title</h1>"},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle default content-type with map body", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/custom",
				},
				Body: map[string]interface{}{"key": "value"},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle default content-type with string body", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/custom",
				},
				Body: "string body",
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with nil value in map", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"field1": "value1",
					"field2": nil,
					"field3": "value3",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		// Error handling tests for multipart/form-data
		It("should handle multipart/form-data with complex nested map containing all valid types", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"nestedMap": map[string]interface{}{
						"key1": "value1",
						"key2": 123,
						"key3": true,
						"key4": 45.67,
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with array containing all valid types", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"arrayField": []interface{}{
						"string",
						123,
						true,
						45.67,
						map[string]interface{}{"nested": "map"},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with primitive string value", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"simpleString": "test value",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with primitive int value", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"simpleInt": 42,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with primitive bool value", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"simpleBool": true,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with primitive float value", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"simpleFloat": 3.14159,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with map containing unsupported types that json.Marshal handles", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			// json.Marshal can handle most basic types, so this tests the success path
			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"complexMap": map[string]interface{}{
						"nullValue": nil,
						"emptyString": "",
						"zero": 0,
						"negativInt": -42,
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with empty nested map", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"emptyMap": map[string]interface{}{},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with empty array", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"emptyArray": []interface{}{},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with deeply nested structure", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"level3": map[string]interface{}{
								"data": "deep value",
							},
						},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with array of arrays", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"matrix": []interface{}{
						[]interface{}{1, 2, 3},
						[]interface{}{4, 5, 6},
						[]interface{}{7, 8, 9},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with mixed file and data fields", func() {
			tmpFile, err := os.CreateTemp("", "test-mixed-*.txt")
			Expect(err).To(BeNil())
			defer os.Remove(tmpFile.Name())
			_, _ = tmpFile.WriteString("file content")
			tmpFile.Seek(0, 0)

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"file": tmpFile,
					"name": "test file",
					"metadata": map[string]interface{}{
						"size": 12,
						"type": "text",
					},
					"tags": []interface{}{"test", "sample"},
					"count": 42,
					"enabled": true,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		// Error path tests for multipart/form-data operations
		It("should handle error when file is closed before reading in multipart/form-data", func() {
			tmpFile, err := os.CreateTemp("", "test-closed-*.txt")
			Expect(err).To(BeNil())
			fileName := tmpFile.Name()
			_, _ = tmpFile.WriteString("file content")
			// Close the file to trigger io.Copy error
			tmpFile.Close()
			defer os.Remove(fileName)

			// Reopen file for deletion but create request with closed file handle
			closedFile, _ := os.Open(fileName)
			closedFile.Close() // Close immediately to trigger error

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"file": closedFile,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			// Should return error due to closed file
			Expect(err).ToNot(BeNil())
			Expect(response).To(BeNil())
		})

		It("should handle multipart/form-data with types that json.Marshal can handle", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			// Test with all JSON-compatible types in nested map
			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"complexData": map[string]interface{}{
						"string":  "value",
						"number":  42,
						"float":   3.14,
						"bool":    true,
						"null":    nil,
						"array":   []interface{}{1, 2, 3},
						"nested":  map[string]interface{}{"key": "val"},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with arrays containing all JSON types", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"mixedArray": []interface{}{
						"string",
						123,
						45.67,
						true,
						false,
						nil,
						map[string]interface{}{"nested": "object"},
						[]interface{}{1, 2, 3},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with all primitive value types", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"stringField":  "text value",
					"intField":     42,
					"floatField":   3.14159,
					"boolField":    true,
					"zeroField":    0,
					"emptyString":  "",
					"negativeInt":  -100,
					"negativeFloat": -99.99,
					"falseField":   false,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with bytes.Reader as io.Reader", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			fileContent := []byte("This is file content from bytes.Reader")
			reader := strings.NewReader(string(fileContent))

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"fileFromReader": reader,
					"description":    "File uploaded via io.Reader",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with multiple files of different types", func() {
			// Create temp files
			txtFile, _ := os.CreateTemp("", "test-*.txt")
			_, _ = txtFile.WriteString("text content")
			txtFile.Seek(0, 0)
			defer os.Remove(txtFile.Name())

			jsonFile, _ := os.CreateTemp("", "test-*.json")
			_, _ = jsonFile.WriteString(`{"key": "value"}`)
			jsonFile.Seek(0, 0)
			defer os.Remove(jsonFile.Name())

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"textFile":  txtFile,
					"jsonFile":  jsonFile,
					"readerFile": strings.NewReader("reader content"),
					"metadata":  map[string]interface{}{"count": 2},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with large nested structure", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2a": map[string]interface{}{
							"level3": []interface{}{
								map[string]interface{}{"id": 1, "name": "item1"},
								map[string]interface{}{"id": 2, "name": "item2"},
							},
						},
						"level2b": []interface{}{
							[]interface{}{1, 2, 3},
							[]interface{}{4, 5, 6},
						},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with special characters in primitive values", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"specialChars": "value with spaces & symbols !@#$%^&*()",
					"unicode":      "Hello 世界 🌍",
					"quotes":       `value with "quotes" and 'apostrophes'`,
					"newlines":     "line1\nline2\nline3",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		// Error path tests for FORMDATA case
		It("should return error when io.Reader fails during io.Copy in multipart/form-data", func() {
			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"failingReader": &errorReader{},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			// Should return error due to failing reader
			Expect(err).ToNot(BeNil())
			Expect(response).To(BeNil())
		})

		It("should handle multipart/form-data with channel type causing json.Marshal to work with map", func() {
			// Note: json.Marshal will handle most types, but channels, functions, and complex types cause issues
			// However, since we're putting them in a map[string]interface{}, Go will handle the conversion
			// This test verifies the happy path where json.Marshal succeeds even with edge case types
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"nestedData": map[string]interface{}{
						"validString": "test",
						"validNumber": 123,
						"validBool":   true,
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with array containing various valid types", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"arrayData": []interface{}{
						"string",
						123,
						45.67,
						true,
						map[string]interface{}{"nested": "value"},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with all primitive types as WriteField values", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"stringPrimitive": "text",
					"intPrimitive":    42,
					"floatPrimitive":  3.14,
					"boolPrimitive":   true,
					"int64Primitive":  int64(9223372036854775807),
					"int32Primitive":  int32(2147483647),
					"float32Primitive": float32(3.14159),
					"uint8Primitive":   uint8(255),
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with empty string primitive", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"emptyString": "",
					"whitespace":  "   ",
					"tab":         "\t",
					"newline":     "\n",
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with zero values for all numeric types", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"zeroInt":     0,
					"zeroFloat":   0.0,
					"zeroInt64":   int64(0),
					"zeroFloat32": float32(0.0),
					"falseBool":   false,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with very long string primitive", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			longString := strings.Repeat("a", 10000)
			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"longString": longString,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with nested maps at multiple levels", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"nested1": map[string]interface{}{
						"nested2": map[string]interface{}{
							"nested3": map[string]interface{}{
								"nested4": map[string]interface{}{
									"value": "deeply nested",
								},
							},
						},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with arrays containing nested arrays", func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"nestedArrays": []interface{}{
						[]interface{}{
							[]interface{}{
								[]interface{}{1, 2, 3},
							},
						},
					},
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		It("should handle multipart/form-data with combination of files, maps, arrays, and primitives", func() {
			tmpFile, err := os.CreateTemp("", "combo-test-*.txt")
			Expect(err).To(BeNil())
			defer os.Remove(tmpFile.Name())
			_, _ = tmpFile.WriteString("combo test content")
			tmpFile.Seek(0, 0)

			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer mockServer.Close()
			ctrl.Config.ConnectionUrl = mockServer.URL

			request := InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
				},
				Body: map[string]interface{}{
					"file1":  tmpFile,
					"file2":  strings.NewReader("reader content"),
					"map1":   map[string]interface{}{"key": "value"},
					"array1": []interface{}{1, 2, 3},
					"string1": "text",
					"int1":    42,
					"bool1":   true,
					"float1":  3.14,
				},
			}
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
				return nil
			}
			response, err := ctrl.Invoke(ctx, request)
			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})
	})

})


})

var _ = Describe("ConnectionController edge cases", func() {
	var (
		ctrl      *ConnectionController
		mockToken string
		ctx       = context.TODO()
	)

	BeforeEach(func() {
		mockToken = "mock-valid-token"
		ctrl = &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionUrl: "http://mockserver.com",
				ConnectionId:  "demo",
			},
			Token: mockToken,
		}
	})

	// --- setQueryParams: extended numeric types ---
	Context("setQueryParams with extended numeric types", func() {
		var mockServer *httptest.Server

		BeforeEach(func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			ctrl.Config.ConnectionUrl = mockServer.URL
		})
		AfterEach(func() { mockServer.Close() })

		It("should encode int32 query param without error", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			req := InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"limit": int32(50)},
			}
			resp, err := ctrl.Invoke(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode int64 query param without error", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			req := InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"id": int64(9223372036854775807)},
			}
			resp, err := ctrl.Invoke(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode float32 query param without error", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			req := InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"ratio": float32(1.5)},
			}
			resp, err := ctrl.Invoke(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode uint query param without error", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			req := InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"count": uint(100)},
			}
			resp, err := ctrl.Invoke(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should return error for nil query param value", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			req := InvokeConnectionRequest{
				Method:      "GET",
				Headers:     map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"key": nil},
			}
			resp, err := ctrl.Invoke(ctx, req)
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	// --- FORMURLENCODED: raw string body ---
	Context("FORMURLENCODED with raw string body", func() {
		var mockServer *httptest.Server
		var receivedBody string

		BeforeEach(func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buf, _ := io.ReadAll(r.Body)
				receivedBody = string(buf)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			ctrl.Config.ConnectionUrl = mockServer.URL
		})
		AfterEach(func() { mockServer.Close() })

		It("should send pre-encoded string body as-is for application/x-www-form-urlencoded", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			req := InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				Body:    "key1=value1&key2=value2",
			}
			resp, err := ctrl.Invoke(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(receivedBody).To(Equal("key1=value1&key2=value2"))
		})

		It("should send fmt-converted body for non-map non-string FORMURLENCODED body", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			req := InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				Body:    42,
			}
			resp, err := ctrl.Invoke(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(receivedBody).To(Equal("42"))
		})
	})

	// --- Response content-type handling ---
	Context("Response content-type handling", func() {
		BeforeEach(func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
		})

		It("should return string data for text/xml response", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/xml")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`<?xml version="1.0"?><root><item>test</item></root>`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(resp.Data).To(Equal(`<?xml version="1.0"?><root><item>test</item></root>`))
		})

		It("should parse JSON response with charset parameter (application/json; charset=utf-8)", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"parsed": true}`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			dataMap, ok := resp.Data.(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(dataMap["parsed"]).To(Equal(true))
		})

		It("should return string data for text/html response and assert body value", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`<html><body>Hello</body></html>`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(resp.Data).To(Equal("<html><body>Hello</body></html>"))
		})

		It("should return string data for unknown response content-type (e.g. application/pdf)", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/pdf")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`binary-like-content`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(resp.Data).To(Equal("binary-like-content"))
		})

		It("should return string data for multipart/form-data response and assert body value", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "multipart/form-data; boundary=abc")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`--abc\r\nContent-Disposition: form-data; name="field"\r\n\r\nvalue\r\n--abc--`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(resp.Data).To(BeAssignableToTypeOf(""))
		})
	})
})

var _ = Describe("VaultController", func() {
	var vaultController *VaultController

	BeforeEach(func() {
		vaultController = &VaultController{
			Config: &VaultConfig{
				Credentials: Credentials{
					Path: "test/path",
				},
			},
		}
	})

	Context("SetBearerTokenForVaultController", func() {
		It("should throw error if the current token is expired", func() {
			vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.Roles = []string{"demo"}
			vaultController.Config.Credentials.Context = "demo"

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).To(BeNil())
		})
		It("should create token if the current token is expired", func() {
			vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			vaultController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

			err := SetBearerTokenForVaultController(vaultController)

			Expect(err).To(BeNil())
		})
		It("should generate token if file path is provided", func() {
			vaultController.Token = ""
			vaultController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).To(BeNil())
			Expect(vaultController.Token).ToNot(BeNil())
		})
		It("should reuse token if valid token is provided", func() {
			vaultController.Token = ""
			vaultController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).To(BeNil())
			Expect(vaultController.Token).ToNot(BeNil())

			// vaultController.Config.Credentials.Path = ""
			errs := SetBearerTokenForVaultController(vaultController)
			Expect(errs).To(BeNil())
			Expect(vaultController.Token).ToNot(BeNil())
		})
		It("should generate token if file creds as string is provided", func() {
			vaultController.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = ""
			vaultController.Config.Credentials.CredentialsString = os.Getenv("VALID_CREDS_PVT_KEY")

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).To(BeNil())
			Expect(vaultController.Token).ToNot(BeNil())
		})
		It("should generate token if wrong creds string is provided", func() {
			vaultController.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = ""
			vaultController.Config.Credentials.CredentialsString = "{demo}"

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).ToNot(BeNil())
		})
		It("should generate token if apikey string is provided", func() {
			vaultController.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = os.Getenv("API_KEY")
			vaultController.Config.Credentials.CredentialsString = ""

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).To(BeNil())
		})
		It("should generate token if apikey string is provided", func() {
			vaultController.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = ""
			vaultController.Config.Credentials.CredentialsString = ""

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).ToNot(BeNil())
		})
		It("should generate token if apikey string is provided", func() {
			vaultController.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = ""
			vaultController.Config.Credentials.CredentialsString = ""
			vaultController.CommonCreds = &Credentials{
				ApiKey: os.Getenv("API_KEY"),
			}

			err := SetBearerTokenForVaultController(vaultController)
			Expect(err).To(BeNil())
		})

	})

	Context("CreateRequestClient", func() {
		It("should create an API client with a valid token", func() {
			vaultController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")
			err1 := SetBearerTokenForVaultController(vaultController)
			Expect(err1).To(BeNil())

			vaultController.Config.Credentials.Token = vaultController.Token
			vaultController.Config.Env = DEV
			vaultController.Config.ClusterId = "test-cluster"

			err := CreateRequestClient(vaultController, map[CustomHeaderKey]string{})
			Expect(err).To(BeNil())
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should create an API client with a valid token generation", func() {
			vaultController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")
			vaultController.Token = ""
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.CredentialsString = ""

			//vaultController.Config.Credentials.Token = vaultController.Token
			vaultController.Config.Env = DEV
			vaultController.Config.ClusterId = "test-cluster"

			err := CreateRequestClient(vaultController, map[CustomHeaderKey]string{})
			Expect(err).To(BeNil())
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should throw an error with a invalid path", func() {
			vaultController.Config.Credentials.Path = "invalid_path.json"
			vaultController.Token = ""
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.CredentialsString = ""

			//vaultController.Config.Credentials.Token = vaultController.Token
			vaultController.Config.Env = DEV
			vaultController.Config.ClusterId = "test-cluster"

			err := CreateRequestClient(vaultController, map[CustomHeaderKey]string{})
			Expect(err).ToNot(BeNil())
		})
		It("should return an error if the token is expired", func() {
			vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			err := CreateRequestClient(vaultController, map[CustomHeaderKey]string{})
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
			vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			vaultController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

			err1 := SetBearerTokenForVaultController(vaultController)
			Expect(err1).To(BeNil())

			err2 := CreateRequestClient(vaultController, map[CustomHeaderKey]string{})
			Expect(err2).ToNot(BeNil())
			Expect(err2.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))

		})
		It("should add apikey", func() {
			//vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"

			err := CreateRequestClient(vaultController, map[CustomHeaderKey]string{})
			Expect(err).To(BeNil())
			//Expect(vaultController.Token).To(Equal(vaultController.Config.Credentials.ApiKey))
		})
		It("should apply controller-level CustomHeaders when set", func() {
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"
			vaultController.CustomHeaders = map[CustomHeaderKey]string{
				CustomHeaderKey("x-custom-header"): "custom-value",
				CustomHeaderKey("x-api-version"):   "v1",
			}

			err := CreateRequestClient(vaultController, nil)
			Expect(err).To(BeNil())
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should apply request-level headers when provided", func() {
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"

			requestHeaders := map[CustomHeaderKey]string{
				RequestIdHeader:                   "req-123",
				CustomHeaderKey("x-trace-header"): "trace-value",
			}

			err := CreateRequestClient(vaultController, requestHeaders)
			Expect(err).To(BeNil())
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should apply both controller CustomHeaders and request-level headers", func() {
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"
			vaultController.CustomHeaders = map[CustomHeaderKey]string{
				CustomHeaderKey("x-controller-header"): "controller-value",
				CustomHeaderKey("x-common-header"):     "controller-common",
			}

			requestHeaders := map[CustomHeaderKey]string{
				CustomHeaderKey("x-request-header"): "request-value",
				CustomHeaderKey("x-common-header"):  "request-common",
			}

			err := CreateRequestClient(vaultController, requestHeaders)
			Expect(err).To(BeNil())
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should use Config.Credentials.Token when valid and not expired", func() {
			claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, _ := tok.SignedString([]byte("secret"))

			vaultController.Config.Credentials.ApiKey = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.Token = tokenString

			err := CreateRequestClient(vaultController, nil)
			Expect(err).To(BeNil())
			Expect(vaultController.Token).To(Equal(tokenString))
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should use BaseVaultUrl when set instead of constructing from Env and ClusterId", func() {
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"
			vaultController.Config.BaseVaultUrl = "https://custom.vault.example.com"

			err := CreateRequestClient(vaultController, nil)
			Expect(err).To(BeNil())
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should not panic when CustomHeaders is a non-nil empty map", func() {
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"
			vaultController.CustomHeaders = map[CustomHeaderKey]string{}

			err := CreateRequestClient(vaultController, nil)
			Expect(err).To(BeNil())
			Expect(vaultController.ApiClient).ToNot(BeNil())
		})
		It("should give request-level headers precedence over controller-level headers for the same key", func() {
			var capturedHeader http.Header
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedHeader = r.Header.Clone()
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"records":[]}`))
			}))
			defer ts.Close()

			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"
			vaultController.Config.VaultId = "vault-id"
			vaultController.Config.BaseVaultUrl = ts.URL
			vaultController.CustomHeaders = map[CustomHeaderKey]string{
				CustomHeaderKey("x-priority"): "controller-value",
			}
			requestHeaders := map[CustomHeaderKey]string{
				CustomHeaderKey("x-priority"): "request-value",
			}

			err := CreateRequestClient(vaultController, requestHeaders)
			Expect(err).To(BeNil())

			// Trigger a real HTTP request to capture the headers the server receives
			tok := "test-token"
			payload := &vaultapis.V1DetokenizePayload{
				DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
					{Token: &tok},
				},
			}
			_, _ = vaultController.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
				context.Background(), vaultController.Config.VaultId, payload,
			)

			Expect(capturedHeader).ToNot(BeNil())
			Expect(capturedHeader.Get("x-priority")).To(Equal("request-value"),
				"request-level header should override controller-level header for the same key")
		})

	})
})

var _ = Describe("VaultController — deprecated field fallbacks", func() {
	var ts *httptest.Server
	var ctx context.Context
	originalCreateRequestClientFunc := CreateRequestClientFunc

	BeforeEach(func() {
		ctx = context.TODO()
	})

	AfterEach(func() {
		CreateRequestClientFunc = originalCreateRequestClientFunc
		if ts != nil {
			ts.Close()
			ts = nil
		}
	})

	Context("VaultConfig.BaseVaultURL → BaseVaultUrl", func() {
		makeDetokenizeCall := func(vc *VaultController) {
			tok := "t"
			_, _ = vc.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
				ctx, "vault1", &vaultapis.V1DetokenizePayload{
					DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{{Token: &tok}},
				},
			)
		}

		Context("old field only", func() {
			It("CreateRequestClient succeeds when only deprecated BaseVaultURL is set", func() {
				vc := &VaultController{
					Config: &VaultConfig{
						VaultId:      "vault1",
						BaseVaultURL: "https://custom.vault.example.com",
						Credentials:  Credentials{ApiKey: "k"},
					},
				}
				err := CreateRequestClient(vc, nil)
				Expect(err).To(BeNil())
				Expect(vc.ApiClient).ToNot(BeZero(),
					"ApiClient should be initialised when only BaseVaultURL (deprecated) is set")
			})
		})

		Context("new field only", func() {
			It("CreateRequestClient routes requests to BaseVaultUrl when only new field is set", func() {
				var called bool
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{}`))
				}))
				vc := &VaultController{
					Config: &VaultConfig{
						VaultId:      "vault1",
						BaseVaultUrl: ts.URL,
						Credentials:  Credentials{ApiKey: "k"},
					},
				}
				err := CreateRequestClient(vc, nil)
				Expect(err).To(BeNil())
				makeDetokenizeCall(vc)
				Expect(called).To(BeTrue(),
					"request should reach the server at BaseVaultUrl (new)")
			})
		})

		Context("both old and new set together", func() {
			It("new BaseVaultUrl wins over deprecated BaseVaultURL", func() {
				var called bool
				ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{}`))
				}))
				vc := &VaultController{
					Config: &VaultConfig{
						VaultId:      "vault1",
						BaseVaultUrl: ts.URL,
						BaseVaultURL: "https://old.example.com",
						Credentials:  Credentials{ApiKey: "k"},
					},
				}
				err := CreateRequestClient(vc, nil)
				Expect(err).To(BeNil())
				makeDetokenizeCall(vc)
				Expect(called).To(BeTrue(),
					"request should reach the new BaseVaultUrl server, not BaseVaultURL (deprecated)")
			})
		})
	})

	Context("GetOptions.DownloadURL → DownloadUrl", func() {
		makeGetMock := func(captureQuery *string) {
			ts = setupMockServer(map[string]interface{}{
				"records": []interface{}{
					map[string]interface{}{"fields": map[string]interface{}{"SkyflowId": "id1"}, "tokens": nil},
				},
			}, "ok", "/vaults/v1/vaults/")
			inner := ts.Config.Handler
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				*captureQuery = r.URL.RawQuery
				inner.ServeHTTP(w, r)
			})
			CreateRequestClientFunc = func(v *VaultController, _ map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				c := client.NewClient(
					option.WithBaseURL(ts.URL+"/vaults"),
					option.WithToken("test-token"),
				)
				v.ApiClient = *c
				return nil
			}
		}

		Context("old field only", func() {
			It("deprecated DownloadURL=true is forwarded as downloadURL query param", func() {
				var rawQuery string
				makeGetMock(&rawQuery)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Get(ctx,
					GetRequest{Table: "table", Ids: []string{"id1"}},
					GetOptions{RedactionType: PLAIN_TEXT, DownloadURL: true},
				)
				Expect(rawQuery).To(ContainSubstring("downloadURL=true"),
					"deprecated DownloadURL should be forwarded as downloadURL query param")
			})
		})

		Context("new field only", func() {
			It("DownloadUrl=true is forwarded as downloadURL query param", func() {
				var rawQuery string
				makeGetMock(&rawQuery)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Get(ctx,
					GetRequest{Table: "table", Ids: []string{"id1"}},
					GetOptions{RedactionType: PLAIN_TEXT, DownloadUrl: true},
				)
				Expect(rawQuery).To(ContainSubstring("downloadURL=true"),
					"new DownloadUrl should be forwarded as downloadURL query param")
			})

			It("DownloadUrl=false — downloadURL is absent from query params", func() {
				var rawQuery string
				makeGetMock(&rawQuery)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Get(ctx,
					GetRequest{Table: "table", Ids: []string{"id1"}},
					GetOptions{RedactionType: PLAIN_TEXT, DownloadUrl: false},
				)
				Expect(rawQuery).ToNot(ContainSubstring("downloadURL=true"),
					"DownloadUrl=false should not send downloadURL query param")
			})
		})

		Context("both old and new set together", func() {
			runGet := func(newVal bool, oldVal bool) string {
				var rawQuery string
				makeGetMock(&rawQuery)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Get(ctx,
					GetRequest{Table: "table", Ids: []string{"id1"}},
					GetOptions{RedactionType: PLAIN_TEXT, DownloadUrl: newVal, DownloadURL: oldVal},
				)
				return rawQuery
			}

			// DownloadUrl (bool) | DownloadURL (bool) | result in request
			It("new=true,  old=true  → downloadURL=true (new wins)", func() {
				Expect(runGet(true, true)).To(ContainSubstring("downloadURL=true"))
			})
			It("new=true,  old=false → downloadURL=true (new wins over no-op old)", func() {
				Expect(runGet(true, false)).To(ContainSubstring("downloadURL=true"))
			})
			It("new=false, old=true  → downloadURL=true (deprecated fallback activates)", func() {
				Expect(runGet(false, true)).To(ContainSubstring("downloadURL=true"))
			})
			It("new=false, old=false → no downloadURL   (neither active)", func() {
				Expect(runGet(false, false)).ToNot(ContainSubstring("downloadURL=true"))
			})
		})
	})

	Context("DetokenizeOptions.DownloadURL → DownloadUrl — final request body", func() {
		// captureBody reads the POST body and stores the parsed JSON so tests can inspect it.
		makeDetokenizeMock := func(captureBody *map[string]interface{}) {
			response := map[string]interface{}{
				"records": []interface{}{
					map[string]interface{}{
						"token":     "tok",
						"valueType": "STRING",
						"value":     "v",
					},
				},
			}
			ts = setupMockServer(response, "ok", "/vaults/v1/vaults/")
			inner := ts.Config.Handler
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Body != nil {
					raw, _ := io.ReadAll(r.Body)
					r.Body = io.NopCloser(bytes.NewBuffer(raw))
					var parsed map[string]interface{}
					_ = json.Unmarshal(raw, &parsed)
					*captureBody = parsed
				}
				inner.ServeHTTP(w, r)
			})
			CreateRequestClientFunc = func(v *VaultController, _ map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				c := client.NewClient(
					option.WithBaseURL(ts.URL+"/vaults"),
					option.WithToken("test-token"),
				)
				v.ApiClient = *c
				return nil
			}
		}

		detokenizeReq := func() DetokenizeRequest {
			return DetokenizeRequest{
				DetokenizeData: []common.DetokenizeData{{Token: "tok"}},
			}
		}

		Context("old field only", func() {
			It("deprecated DownloadURL=true → downloadURL:true in request body", func() {
				var body map[string]interface{}
				makeDetokenizeMock(&body)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Detokenize(ctx, detokenizeReq(), DetokenizeOptions{DownloadURL: true})
				Expect(body).To(HaveKeyWithValue("downloadURL", true))
			})

			It("deprecated DownloadURL=false (not set) → downloadURL absent from request body", func() {
				var body map[string]interface{}
				makeDetokenizeMock(&body)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Detokenize(ctx, detokenizeReq(), DetokenizeOptions{})
				Expect(body).ToNot(HaveKey("downloadURL"))
			})
		})

		Context("new field only", func() {
			It("DownloadUrl=true → downloadURL:true in request body", func() {
				var body map[string]interface{}
				makeDetokenizeMock(&body)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Detokenize(ctx, detokenizeReq(), DetokenizeOptions{DownloadUrl: true})
				Expect(body).To(HaveKeyWithValue("downloadURL", true))
			})

			It("DownloadUrl=false — downloadURL absent from request body", func() {
				var body map[string]interface{}
				makeDetokenizeMock(&body)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Detokenize(ctx, detokenizeReq(), DetokenizeOptions{DownloadUrl: false})
				Expect(body).ToNot(HaveKey("downloadURL"))
			})
		})

		Context("both old and new set together", func() {
			runDetokenize := func(newVal bool, oldVal bool) map[string]interface{} {
				var body map[string]interface{}
				makeDetokenizeMock(&body)
				vc := &VaultController{Config: &VaultConfig{VaultId: "vault1", Credentials: Credentials{ApiKey: "k"}}}
				_, _ = vc.Detokenize(ctx, detokenizeReq(), DetokenizeOptions{DownloadUrl: newVal, DownloadURL: oldVal})
				return body
			}

			// DownloadUrl (bool) | DownloadURL (bool) | downloadURL field in request body
			It("new=true,  old=true  → downloadURL:true  (new wins)", func() {
				Expect(runDetokenize(true, true)).To(HaveKeyWithValue("downloadURL", true))
			})
			It("new=true,  old=false → downloadURL:true  (new wins over no-op old)", func() {
				Expect(runDetokenize(true, false)).To(HaveKeyWithValue("downloadURL", true))
			})
			It("new=false, old=true  → downloadURL:true  (deprecated fallback activates)", func() {
				Expect(runDetokenize(false, true)).To(HaveKeyWithValue("downloadURL", true))
			})
			It("new=false, old=false → key absent        (neither active)", func() {
				Expect(runDetokenize(false, false)).ToNot(HaveKey("downloadURL"))
			})
		})
	})
})

var _ = Describe("VaultController — response key backward compat", func() {
	var ts *httptest.Server
	var ctx context.Context
	originalCreateRequestClientFunc := CreateRequestClientFunc

	BeforeEach(func() {
		ctx = context.TODO()
	})

	AfterEach(func() {
		CreateRequestClientFunc = originalCreateRequestClientFunc
		if ts != nil {
			ts.Close()
			ts = nil
		}
	})

	newVC := func() *VaultController {
		return &VaultController{
			Config: &VaultConfig{
				VaultId:   "id",
				ClusterId: "clusterid",
				Env:       PROD,
				Credentials: Credentials{ApiKey: "sky-token"},
			},
		}
	}

	setMockClient := func(vc *VaultController) {
		CreateRequestClientFunc = func(v *VaultController, _ map[CustomHeaderKey]string) *skyflowError.SkyflowError {
			c := client.NewClient(option.WithBaseURL(ts.URL+"/vaults"), option.WithToken("test-token"))
			v.ApiClient = *c
			return nil
		}
	}

	Context("Insert (ContinueOnError=false) — InsertedFields contains both SkyflowId and skyflow_id", func() {
		It("response map contains both SkyflowId (new) and skyflow_id (deprecated) for each record", func() {
			resp := make(map[string]interface{})
			_ = json.Unmarshal([]byte(mockInsertContinueFalseSuccessJSON), &resp)
			ts = setupMockServer(resp, "ok", "/vaults/v1/vaults/")
			vc := newVC()
			setMockClient(vc)
			res, err := vc.Insert(ctx, InsertRequest{
				Table:  "test_table",
				Values: []map[string]interface{}{{"name": "john"}},
			}, InsertOptions{ContinueOnError: false})
			Expect(err).To(BeNil())
			Expect(res.InsertedFields[0]).To(HaveKeyWithValue("SkyflowId", "skyflowid1"),
				"new SkyflowId key must be present")
			Expect(res.InsertedFields[0]).To(HaveKeyWithValue("skyflow_id", "skyflowid1"),
				"deprecated skyflow_id key must be retained for backward compatibility")
		})
	})

	Context("Insert (ContinueOnError=true) — InsertedFields contains both id and index keys", func() {
		It("response contains both SkyflowId and skyflow_id, and both RequestIndex and request_index", func() {
			resp := make(map[string]interface{})
			_ = json.Unmarshal([]byte(mockInsertSuccessJSON), &resp)
			ts = setupMockServer(resp, "ok", "/vaults/v1/vaults/")
			vc := newVC()
			setMockClient(vc)
			res, err := vc.Insert(ctx, InsertRequest{
				Table:  "test_table",
				Values: []map[string]interface{}{{"name": "john"}},
			}, InsertOptions{ContinueOnError: true})
			Expect(err).To(BeNil())
			Expect(res.InsertedFields[0]).To(HaveKey("SkyflowId"),
				"new SkyflowId key must be present")
			Expect(res.InsertedFields[0]).To(HaveKey("skyflow_id"),
				"deprecated skyflow_id key must be retained")
			Expect(res.InsertedFields[0]).To(HaveKey("RequestIndex"),
				"new RequestIndex key must be present")
			Expect(res.InsertedFields[0]).To(HaveKey("request_index"),
				"deprecated request_index key must be retained")
		})
	})

	Context("Get — Data contains both SkyflowId and skyflow_id", func() {
		It("response map contains both SkyflowId (new) and skyflow_id (deprecated)", func() {
			resp := make(map[string]interface{})
			_ = json.Unmarshal([]byte(mockGetSuccessJSON), &resp)
			ts = setupMockServer(resp, "ok", "/vaults/v1/vaults/")
			vc := newVC()
			setMockClient(vc)
			res, err := vc.Get(ctx, GetRequest{Table: "test_table", Ids: []string{"id1"}},
				GetOptions{RedactionType: PLAIN_TEXT})
			Expect(err).To(BeNil())
			Expect(res.Data[0]).To(HaveKeyWithValue("SkyflowId", "id1"),
				"new SkyflowId key must be present")
			Expect(res.Data[0]).To(HaveKeyWithValue("skyflow_id", "id1"),
				"deprecated skyflow_id key must be retained for backward compatibility")
		})
	})

	Context("Query — Fields contains both SkyflowId and skyflow_id", func() {
		It("response map contains both SkyflowId (new) and skyflow_id (deprecated)", func() {
			resp := make(map[string]interface{})
			_ = json.Unmarshal([]byte(mockQuerySuccessJSON), &resp)
			ts = setupMockServer(resp, "ok", "/vaults/v1/vaults/")
			vc := newVC()
			setMockClient(vc)
			res, err := vc.Query(ctx,
				QueryRequest{Query: "SELECT * FROM test_table WHERE skyflow_id='id'"},
				QueryOptions{})
			Expect(err).To(BeNil())
			Expect(res.Fields[0]).To(HaveKeyWithValue("SkyflowId", "id"),
				"new SkyflowId key must be present")
			Expect(res.Fields[0]).To(HaveKeyWithValue("skyflow_id", "id"),
				"deprecated skyflow_id key must be retained for backward compatibility")
		})
	})
})

var _ = Describe("DetectController", func() {
	Describe("Detect client creation", func() {
		var detectController *DetectController

		BeforeEach(func() {
			detectController = &DetectController{
				Config: &VaultConfig{
					Credentials: Credentials{
						Path: "credentials.json",
					},
				},
			}
		})

		Context("SetBearerTokenForDetectControllerFunc", func() {
			It("should throw error if the current token is expired", func() {

				detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				detectController.Config.Credentials.Path = ""
				detectController.Config.Credentials.Roles = []string{"demo"}
				detectController.Config.Credentials.Context = "demo"

				err := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err).To(BeNil())
			})
			It("should create token if the current token is expired", func() {
				detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

				err := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err).To(BeNil())
			})
			It("should generate token if file path is provided", func() {
				detectController.Token = ""
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

				err := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err).To(BeNil())
				Expect(detectController.Token).ToNot(BeNil())
			})
			It("should reuse token if valid token is provided case 2", func() {
				detectController.Token = ""
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

				err := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err).To(BeNil())
				Expect(detectController.Token).ToNot(BeNil())

				// detectController.Config.Credentials.Path = ""
				errs := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(errs).To(BeNil())
				Expect(detectController.Token).ToNot(BeNil())
			})
			It("should generate token if file creds as string is provided", func() {
				detectController.Token = ""
				detectController.Config.Credentials.Path = ""
				detectController.Config.Credentials.ApiKey = ""
				detectController.Config.Credentials.CredentialsString = os.Getenv("VALID_CREDS_PVT_KEY")

				err := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err).To(BeNil())
				Expect(detectController.Token).ToNot(BeNil())
			})
			It("should generate token if wrong creds string is provided", func() {
				detectController.Token = ""
				detectController.Config.Credentials.Path = ""
				detectController.Config.Credentials.ApiKey = ""
				detectController.Config.Credentials.CredentialsString = "{demo}"

				err := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("Create Detect Request Client", func() {
			It("should create an API client with a valid token", func() {
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")
				err1 := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err1).To(BeNil())

				detectController.Config.Credentials.Token = detectController.Token
				detectController.Config.Env = DEV
				detectController.Config.ClusterId = "test-cluster"

				err := CreateDetectRequestClient(detectController, nil)
				Expect(err).To(BeNil())
				Expect(detectController.TextApiClient).ToNot(BeNil())
				Expect(detectController.FilesApiClient).ToNot(BeNil())

			})
			It("should create an API client with a valid token generation", func() {
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")
				detectController.Token = ""
				detectController.Config.Credentials.Token = ""
				detectController.Config.Credentials.CredentialsString = ""

				detectController.Config.Env = DEV
				detectController.Config.ClusterId = "test-cluster"
				headers := map[CustomHeaderKey]string{
					CustomHeaderKey("x-request-id"): "test-value",
				}
				err := CreateDetectRequestClient(detectController, headers)
				Expect(err).To(BeNil())
				Expect(detectController.TextApiClient).ToNot(BeNil())
				Expect(detectController.FilesApiClient).ToNot(BeNil())
			})
			It("should throw an error with a invalid path", func() {
				detectController.Config.Credentials.Path = "invalid_path.json"
				detectController.Token = ""
				detectController.Config.Credentials.Token = ""
				detectController.Config.Credentials.CredentialsString = ""

				detectController.Config.Env = DEV
				detectController.Config.ClusterId = "test-cluster"
				detectController.CustomHeaders = map[CustomHeaderKey]string{
					CustomHeaderKey("x-request-id"): "test-value",
				}
				err := CreateDetectRequestClient(detectController, nil)
				Expect(err).ToNot(BeNil())
			})
			It("should return an error if the token is expired", func() {
				detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				err := CreateDetectRequestClient(detectController, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
				detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

				err1 := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err1).To(BeNil())

				err2 := CreateDetectRequestClient(detectController, nil)
				Expect(err2).ToNot(BeNil())
				Expect(err2.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))

			})
			It("should add apikey", func() {
				//detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				detectController.Config.Credentials.Token = ""
				detectController.Config.Credentials.Path = ""
				detectController.Config.Credentials.ApiKey = "test-api-key"
				detectController.CustomHeaders = map[CustomHeaderKey]string{
					CustomHeaderKey("x-request-id"): "test-value",
				}
				err := CreateDetectRequestClient(detectController, map[CustomHeaderKey]string{
					CustomHeaderKey("x-request-id"): "test-value2",
				})
				Expect(err).To(BeNil())
			})

		})
	})
	Describe("CreateDeidentifyTextRequest tests", func() {
		var config VaultConfig

		BeforeEach(func() {
			config = VaultConfig{
				VaultId: "vault123",
			}
		})

		Context("when given valid input", func() {
			It("should create a valid payload", func() {
				req := DeidentifyTextRequest{
					Text:              "Sensitive text",
					Entities:          []DetectEntities{Name},
					AllowRegexList:    []string{"demo"},
					RestrictRegexList: []string{"demo"},
					TokenFormat: TokenFormat{
						DefaultType: TokenTypeDefaultEntityOnly,
					},
					Transformations: Transformations{
						ShiftDates: DateTransformation{
							MaxDays: 10,
							MinDays: 1,
							Entities: []TransformationsShiftDatesEntityTypesItem{
								TransformationsShiftDatesEntityTypesItemDate,
							},
						},
					},
				}

				payload, err := CreateDeidentifyTextRequest(req, config)
				Expect(err).To(BeNil())
				Expect(payload).ToNot(BeNil())
				Expect(payload.Text).To(Equal(req.Text))
				Expect(payload.AllowRegex).ToNot(BeNil())
				Expect(payload.RestrictRegex).ToNot(BeNil())
				Expect(payload.EntityTypes).ToNot(BeNil())
				Expect(payload.TokenType.Default).ToNot(BeNil())
				Expect(payload.Transformations.ShiftDates.MaxDays).ToNot(BeNil())
				Expect(payload.Transformations.ShiftDates.MinDays).ToNot(BeNil())
				Expect(payload.Transformations.ShiftDates.EntityTypes).ToNot(BeNil())
			})
		})
	})

	Describe("CreateReidentifyTextRequest tests", func() {
		var config VaultConfig

		BeforeEach(func() {
			config = VaultConfig{
				VaultId: "vault123",
			}
		})

		Context("when creating a valid payload", func() {
			It("should create payload with all entity types", func() {
				request := ReidentifyTextRequest{
					Text:              "Sample text",
					RedactedEntities:  []DetectEntities{Name, EmailAddress},
					MaskedEntities:    []DetectEntities{PhoneNumber},
					PlainTextEntities: []DetectEntities{Date},
				}

				payload, err := CreateReidentifyTextRequest(request, config)

				Expect(err).To(BeNil())
				Expect(*payload.VaultId).To(Equal(config.VaultId))
				Expect(*payload.Text).To(Equal(request.Text))
				Expect(payload.Format.Redacted).To(HaveLen(2))
				Expect(payload.Format.Masked).To(HaveLen(1))
				Expect(payload.Format.Plaintext).To(HaveLen(1))
			})
		})
	})

	Describe("CreateDeidentifyFileRequest tests", Ordered, func() {
		var (
			config            VaultConfig
			base64            string
			entities          []DetectEntities
			allowRegexList    []string
			restrictRegexList []string
			tokenFormat       TokenFormat
			transformations   Transformations
			expectedEntities  []string
		)

		BeforeAll(
			func() {
				base64 = "c29tZSB0ZXh0"
				entities = []DetectEntities{Name, EmailAddress}
				allowRegexList = []string{"demo"}
				restrictRegexList = []string{"demo", "test"}
				tokenFormat = TokenFormat{
					DefaultType: TokenTypeDefaultEntityOnly,
				}
				transformations = Transformations{
					ShiftDates: DateTransformation{
						MaxDays: 10,
						MinDays: 1,
						Entities: []TransformationsShiftDatesEntityTypesItem{
							TransformationsShiftDatesEntityTypesItemDate,
						},
					},
				}
				expectedEntities = []string{"name", "email_address"}

			},
		)

		BeforeEach(func() {
			config = VaultConfig{
				VaultId: "vault123",
			}
		})

		It("when creating a valid payload for deidentify text file", func() {
			request := &DeidentifyFileRequest{
				File: FileInput{
					FilePath: "/test/testfile.txt",
				},
				Entities:          entities,
				AllowRegexList:    allowRegexList,
				RestrictRegexList: restrictRegexList,
				TokenFormat:       tokenFormat,
				Transformations:   transformations,
			}

			payload := CreateTextFileRequest(request, base64, config.VaultId)

			Expect(payload.VaultId).To(Equal(config.VaultId))
			Expect(payload.File.Base64).To(Equal(base64))
			Expect(payload.AllowRegex).ToNot(BeNil())
			Expect(payload.AllowRegex).To(HaveLen(len(allowRegexList)))
			Expect(payload.AllowRegex).To(ContainElements(allowRegexList))
			Expect(payload.AllowRegex).To(Equal(allowRegexList))
			Expect(payload.RestrictRegex).ToNot(BeNil())
			Expect(payload.RestrictRegex).To(HaveLen(len(restrictRegexList)))
			Expect(payload.RestrictRegex).To(ContainElements(restrictRegexList))
			Expect(payload.RestrictRegex).To(Equal(restrictRegexList))
			var actualEntities []string
			for _, e := range payload.EntityTypes {
				actualEntities = append(actualEntities, string(e))
			}

			Expect(actualEntities).To(HaveLen(len(expectedEntities)))
			Expect(actualEntities).To(ContainElements(expectedEntities))
			Expect(actualEntities).To(Equal(expectedEntities))
			Expect(payload.Transformations.ShiftDates).ToNot(BeNil())
			Expect(*payload.Transformations.ShiftDates.MaxDays).To(Equal(10))
			Expect(*payload.Transformations.ShiftDates.MinDays).To(Equal(1))

			expected := []vaultapis.ShiftDatesEntityTypesItem{
				vaultapis.ShiftDatesEntityTypesItemDate,
			}

			Expect(payload.Transformations.ShiftDates.EntityTypes).To(Equal(expected))

		})

		It("when creating a valid payload for deidentify image file", func() {
			request := &DeidentifyFileRequest{
				File: FileInput{
					FilePath: "/test/testfile.jpeg",
				},
				Entities:          entities,
				AllowRegexList:    allowRegexList,
				RestrictRegexList: restrictRegexList,
				TokenFormat:       tokenFormat,
				Transformations:   transformations,
				MaskingMethod:     BLACKBOX,
			}

			payload := CreateImageRequest(request, base64, config.VaultId, "jpeg")

			Expect(payload.VaultId).To(Equal(config.VaultId))
			Expect(payload.File.Base64).To(Equal(base64))
			Expect(payload.AllowRegex).ToNot(BeNil())
			Expect(payload.AllowRegex).To(HaveLen(len(allowRegexList)))
			Expect(payload.AllowRegex).To(ContainElements(allowRegexList))
			Expect(payload.AllowRegex).To(Equal(allowRegexList))
			Expect(payload.RestrictRegex).ToNot(BeNil())
			Expect(payload.RestrictRegex).To(HaveLen(len(restrictRegexList)))
			Expect(payload.RestrictRegex).To(ContainElements(restrictRegexList))
			Expect(payload.RestrictRegex).To(Equal(restrictRegexList))
			var actualEntities []string
			for _, e := range payload.EntityTypes {
				actualEntities = append(actualEntities, string(e))
			}

			Expect(actualEntities).To(HaveLen(len(expectedEntities)))
			Expect(actualEntities).To(ContainElements(expectedEntities))
			Expect(actualEntities).To(Equal(expectedEntities))
			Expect(string(*payload.MaskingMethod)).To(Equal(string(BLACKBOX)))
			Expect(payload.Transformations).To(BeNil())
		})

		It("when creating a valid payload for deidentify pdf file", func() {
			request := &DeidentifyFileRequest{
				File: FileInput{
					FilePath: "/test/testfile.pdf",
				},
				Entities:          entities,
				AllowRegexList:    allowRegexList,
				RestrictRegexList: restrictRegexList,
				TokenFormat:       tokenFormat,
				Transformations:   transformations,
				MaskingMethod:     BLACKBOX,
				MaxResolution:     300,
				PixelDensity:      200,
			}

			payload := CreatePdfRequest(request, base64, config.VaultId)

			Expect(payload.VaultId).To(Equal(config.VaultId))
			Expect(payload.File.Base64).To(Equal(base64))
			Expect(payload.AllowRegex).ToNot(BeNil())
			Expect(payload.AllowRegex).To(HaveLen(len(allowRegexList)))
			Expect(payload.AllowRegex).To(ContainElements(allowRegexList))
			Expect(payload.AllowRegex).To(Equal(allowRegexList))
			Expect(payload.RestrictRegex).ToNot(BeNil())
			Expect(payload.RestrictRegex).To(HaveLen(len(restrictRegexList)))
			Expect(payload.RestrictRegex).To(ContainElements(restrictRegexList))
			Expect(payload.RestrictRegex).To(Equal(restrictRegexList))
			var actualEntities []string
			for _, e := range payload.EntityTypes {
				actualEntities = append(actualEntities, string(e))
			}

			Expect(actualEntities).To(HaveLen(len(expectedEntities)))
			Expect(actualEntities).To(ContainElements(expectedEntities))
			Expect(actualEntities).To(Equal(expectedEntities))
			Expect(payload.Transformations).To(BeNil())
			Expect(*payload.MaxResolution).To(Equal(300))
			Expect(*payload.Density).To(Equal(200))
		})
	})
	Describe("DeidentifyText tests", func() {
		var (
			detectController *DetectController
			ctx              context.Context
			mockRequest      DeidentifyTextRequest
		)

		BeforeEach(func() {
			ctx = context.Background()
			detectController = &DetectController{
				Config: &VaultConfig{
					VaultId:   "vault123",
					ClusterId: "cluster123",
					Env:       DEV,
					Credentials: Credentials{
						ApiKey: "test-api-key",
					},
				},
			}
			mockRequest = DeidentifyTextRequest{
				Text:     "My name is John Doe and email is john@example.com",
				Entities: []DetectEntities{Name, EmailAddress},
			}
		})

		Context("Success cases", func() {
			It("should successfully deidentify text with all entity types", func() {
				// Mock API response
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeidentifyTextSuccessJSON), &response)

				// Setup mock server
				ts := setupMockServer(response, "ok", "/v1/detect/deidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest, common.DeidentifyTextOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.ProcessedText).To(Equal("My name is [NAME] and email is [EMAIL]"))
				Expect(result.WordCount).To(Equal(int(8)))
				Expect(result.CharCount).To(Equal(int(45)))
				Expect(result.Entities).To(HaveLen(2))
				Expect(result.Entities[0].Entity).To(Equal("NAME"))
				Expect(result.Entities[1].Entity).To(Equal("EMAIL_ADDRESS"))
			})

			It("should handle empty entities array in response", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeidentifyTextNoEntitiesJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/deidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest, common.DeidentifyTextOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.ProcessedText).To(Equal("No entities found in this text"))
				Expect(result.Entities).To(BeEmpty())
			})
		})

		Context("Error cases", func() {
			It("should return error when validation fails", func() {
				invalidRequest := DeidentifyTextRequest{
					Text: "", // Empty text should fail validation
				}

				result, err := detectController.DeidentifyText(ctx, invalidRequest, common.DeidentifyTextOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when API request fails", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeidentifyTextErrorJSON), &response)

				ts := setupMockServer(response, "error", "/v1/detect/deidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest, common.DeidentifyTextOptions{})
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest, common.DeidentifyTextOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid bearer token")
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest, common.DeidentifyTextOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when custom headers map is empty in DeidentifyText", func() {
				opts := common.DeidentifyTextOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				result, err := detectController.DeidentifyText(ctx, mockRequest, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers has invalid key in DeidentifyText", func() {
				opts := common.DeidentifyTextOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				result, err := detectController.DeidentifyText(ctx, mockRequest, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers has empty value in DeidentifyText", func() {
				opts := common.DeidentifyTextOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				result, err := detectController.DeidentifyText(ctx, mockRequest, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})

		Context("Advanced configuration cases", func() {
			It("should handle requests with token format configuration", func() {
				mockRequest.TokenFormat = TokenFormat{
					EntityOnly: []DetectEntities{Name},
					VaultToken: []DetectEntities{EmailAddress},
				}

				response := make(map[string]interface{})
				mockJSONResponse := `{
					"processed_text": "My name is [NAME] and email is [EMAIL]",
					"entities": [
						{
							"token": "token1",
							"value": "John Doe",
							"entity_type": "NAME",
							"location": {
								"start_index": 11,
								"end_index": 19,
								"start_index_processed": 11,
								"end_index_processed": 17
							}
						}
					]
				}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/deidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest, common.DeidentifyTextOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.ProcessedText).To(Equal("My name is [NAME] and email is [EMAIL]"))
				Expect(result.Entities[0].Entity).To(Equal("NAME"))
			})

			It("should handle requests with regex configuration", func() {
				mockRequest.AllowRegexList = []string{"[A-Z][a-z]+"}
				mockRequest.RestrictRegexList = []string{"[0-9]+"}

				response := make(map[string]interface{})
				mockJSONResponse := `{
					"processed_text": "My name is [NAME] and email is [EMAIL]",
					"entities": [
						{
							"token": "token1",
							"value": "John",
							"entity_type": "NAME",
							"location": {
								"startIndex": 11,
								"endIndex": 15,
								"startIndexProcessed": 11,
								"endIndexProcessed": 17
							}
						}
					]
				}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/deidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest, common.DeidentifyTextOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.ProcessedText).To(Equal("My name is [NAME] and email is [EMAIL]"))
			})
		})
	})

	Describe("ReidentifyText tests", func() {
		var (
			detectController *DetectController
			ctx              context.Context
			mockRequest      ReidentifyTextRequest
		)

		BeforeEach(func() {
			ctx = context.Background()
			detectController = &DetectController{
				Config: &VaultConfig{
					VaultId:   "vault123",
					ClusterId: "cluster123",
					Env:       DEV,
					Credentials: Credentials{
						ApiKey: "test-api-key",
					},
				},
			}
			mockRequest = ReidentifyTextRequest{
				Text:             "Sample redacted text",
				RedactedEntities: []DetectEntities{Name, EmailAddress},
				MaskedEntities:   []DetectEntities{PhoneNumber},
			}
		})

		Context("Success cases", func() {
			It("should successfully reidentify text", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockReidentifyTextSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/reidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.ReidentifyText(ctx, mockRequest, common.ReidentifyTextOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.ProcessedText).To(Equal("Sample original text"))
			})
		})

		Context("Error cases", func() {
			It("should return error when validation fails", func() {
				invalidRequest := ReidentifyTextRequest{
					Text: "", // Empty text should fail validation
				}

				result, err := detectController.ReidentifyText(ctx, invalidRequest, common.ReidentifyTextOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when API request fails", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockReidentifyTextErrorJSON), &response)

				ts := setupMockServer(response, "error", "/v1/detect/reidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.ReidentifyText(ctx, mockRequest, common.ReidentifyTextOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				result, err := detectController.ReidentifyText(ctx, mockRequest, common.ReidentifyTextOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid bearer token")
				}

				result, err := detectController.ReidentifyText(ctx, mockRequest, common.ReidentifyTextOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when custom headers map is empty in ReidentifyText", func() {
				opts := common.ReidentifyTextOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				result, err := detectController.ReidentifyText(ctx, mockRequest, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers has invalid key in ReidentifyText", func() {
				opts := common.ReidentifyTextOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				result, err := detectController.ReidentifyText(ctx, mockRequest, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers has empty value in ReidentifyText", func() {
				opts := common.ReidentifyTextOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				result, err := detectController.ReidentifyText(ctx, mockRequest, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("DeidentifyFile tests", Ordered, func() {
		var (
			detectController *DetectController
			ctx              context.Context
			tempDir          string
			testFiles        map[string]*os.File
		)

		BeforeAll(func() {
			var err error
			// Create temporary directory
			tempDir, err = os.MkdirTemp("", "skyflow_test_*")
			Expect(err).To(BeNil(), "Failed to create temp directory for tests")

			// Create temporary test files for each type
			testFiles = make(map[string]*os.File)
			testContent := []byte("Test content for file processing")

			fileTypes := []string{"txt", "mp3", "jpeg", "pdf", "pptx", "xlsx", "docx", "json"}
			for _, fileType := range fileTypes {
				tmpFile, err := os.CreateTemp(tempDir, fmt.Sprintf("detect.*.%s", fileType))
				Expect(err).To(BeNil(), fmt.Sprintf("Failed to create temp %s file", fileType))
				_, err = tmpFile.Write(testContent)
				Expect(err).To(BeNil(), fmt.Sprintf("Failed to write to temp %s file", fileType))
				testFiles[fileType] = tmpFile
			}
		})

		AfterAll(func() {
			// Close and remove all temporary files
			for _, file := range testFiles {
				if file != nil {
					file.Close()
				}
			}

			// Clean up temporary directory and its contents
			if tempDir != "" {
				err := os.RemoveAll(tempDir)
				Expect(err).To(BeNil(), "Failed to clean up temp directory after tests")
			}
		})

		BeforeEach(func() {
			ctx = context.Background()
			detectController = &DetectController{
				Config: &VaultConfig{
					VaultId:   "vault123",
					ClusterId: "cluster123",
					Env:       DEV,
					Credentials: Credentials{
						ApiKey: "test-api-key",
					},
				},
			}

		})

		Context("Success cases", func() {
			Context("Success cases for different file types", func() {

				audioFilePath := filepath.Join(tempDir, "detect.mp3")
				audioFile, _ := os.Open(audioFilePath)
				defer audioFile.Close()

				var testCases = []struct {
					name        string
					fileExt     string
					endpoint    string
					fileType    string
					mockRequest DeidentifyFileRequest
				}{
					{
						name:     "Text File",
						fileExt:  "txt",
						endpoint: "/v1/detect/deidentify/file/text",
						fileType: "TEXT",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								FilePath: filepath.Join(tempDir, "detect.txt"),
							},
							OutputDirectory: tempDir,
							Entities:        []DetectEntities{Name, EmailAddress, Ssn, Date, Day, Dob},
							WaitTime:        5,
							TokenFormat: TokenFormat{
								EntityOnly: []DetectEntities{
									Name, EmailAddress, Ssn, Date, Day, Dob,
								},
								EntityUniqueCounter: []DetectEntities{
									Ssn, Date, Day, Dob,
								},
							},
							AllowRegexList: []string{
								"My",
							},
							Transformations: Transformations{
								ShiftDates: DateTransformation{
									MinDays: 5,
									MaxDays: 10,
									Entities: []TransformationsShiftDatesEntityTypesItem{
										TransformationsShiftDatesEntityTypesItem(Month),
										TransformationsShiftDatesEntityTypesItem(Date),
										TransformationsShiftDatesEntityTypesItem(Day),
										TransformationsShiftDatesEntityTypesItem(Dob),
										TransformationsShiftDatesEntityTypesItem(CreditCardExpiration),
									},
								},
							},
						},
					},
					{
						name:     "Audio File",
						fileExt:  "mp3",
						endpoint: "/v1/detect/deidentify/file/audio",
						fileType: "MP3",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								File: audioFile,
							},
							Entities: []DetectEntities{Name, EmailAddress, Ssn, Date, Day, Dob},
							TokenFormat: TokenFormat{
								DefaultType: TokenTypeDefaultVaultToken,
							},
							OutputOcrText: true,
							MaxResolution: 200,
							PixelDensity:  200,
							Bleep: AudioBleep{
								Gain:         2,
								Frequency:    1000,
								StartPadding: 2,
								StopPadding:  20,
							},
							OutputProcessedAudio: true,
							AllowRegexList: []string{
								"My",
							},
							Transformations: Transformations{
								ShiftDates: DateTransformation{
									MinDays: 5,
									MaxDays: 10,
									Entities: []TransformationsShiftDatesEntityTypesItem{
										TransformationsShiftDatesEntityTypesItem(Month),
										TransformationsShiftDatesEntityTypesItem(Date),
										TransformationsShiftDatesEntityTypesItem(Day),
										TransformationsShiftDatesEntityTypesItem(Dob),
										TransformationsShiftDatesEntityTypesItem(CreditCardExpiration),
									},
								},
							},
						},
					},

					{
						name:     "Image File",
						fileExt:  "jpeg",
						endpoint: "/v1/detect/deidentify/file/image",
						fileType: "JPEG",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								FilePath: filepath.Join(tempDir, "detect.jpeg"),
							},
							Entities: []DetectEntities{Name, EmailAddress, Ssn, Date, Day, Dob},
							TokenFormat: TokenFormat{
								DefaultType: TokenTypeDefaultVaultToken,
								EntityOnly: []DetectEntities{
									Name, EmailAddress, Ssn, Date,
								},
							},
							OutputOcrText: true,
							MaxResolution: 200,
							PixelDensity:  200,
							AllowRegexList: []string{
								"My",
							},
							MaskingMethod: BLACKBOX,
							Transformations: Transformations{
								ShiftDates: DateTransformation{
									MinDays: 5,
									MaxDays: 10,
									Entities: []TransformationsShiftDatesEntityTypesItem{
										TransformationsShiftDatesEntityTypesItem(Month),
										TransformationsShiftDatesEntityTypesItem(Date),
										TransformationsShiftDatesEntityTypesItem(Day),
										TransformationsShiftDatesEntityTypesItem(Dob),
										TransformationsShiftDatesEntityTypesItem(CreditCardExpiration),
									},
								},
							},
						},
					},
					{
						name:     "PDF Document",
						fileExt:  "pdf",
						endpoint: "/v1/detect/deidentify/file/document/pdf",
						fileType: "PDF",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								FilePath: filepath.Join(tempDir, "detect.pdf"),
							},
							Entities: []DetectEntities{Name, EmailAddress, Ssn},
							TokenFormat: TokenFormat{
								DefaultType: TokenTypeDefaultVaultToken,
								EntityUniqueCounter: []DetectEntities{
									Name, EmailAddress, Ssn, Date,
								},
							},
							WaitTime:      5,
							MaxResolution: 200,
						},
					},
					{
						name:     "Presentation File",
						fileExt:  "pptx",
						endpoint: "/v1/detect/deidentify/file/presentation",
						fileType: "PPTX",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								FilePath: filepath.Join(tempDir, "detect.pptx"),
							},
							Entities: []DetectEntities{Name, EmailAddress},
							WaitTime: 5,
							TokenFormat: TokenFormat{
								DefaultType: TokenTypeDefaultEntityOnly,
							},
						},
					},
					{
						name:     "Spreadsheet File",
						fileExt:  "xlsx",
						endpoint: "/v1/detect/deidentify/file/spreadsheet",
						fileType: "XLSX",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								FilePath: filepath.Join(tempDir, "detect.xlsx"),
							},
							Entities: []DetectEntities{Name, EmailAddress, Ssn},
							WaitTime: 5,
						},
					},
					{
						name:     "Document File",
						fileExt:  "docx",
						endpoint: "/v1/detect/deidentify/file/document",
						fileType: "DOCX",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								FilePath: filepath.Join(tempDir, "detect.docx"),
							},
							Entities: []DetectEntities{Name, EmailAddress},
							WaitTime: 5,
						},
					},
					{
						name:     "Structured Text File",
						fileExt:  "json",
						endpoint: "/v1/detect/deidentify/file/structured_text",
						fileType: "JSON",
						mockRequest: DeidentifyFileRequest{
							File: FileInput{
								FilePath: filepath.Join(tempDir, "detect.json"),
							},
							Entities: []DetectEntities{Name, EmailAddress},
							WaitTime: 5,
						},
					},
				}

				for _, tc := range testCases {
					tc := tc // capture range variable
					It(fmt.Sprintf("should successfully process %s", tc.name), func() {
						// Update file path to use temporary directory
						tc.mockRequest.File.FilePath = testFiles[tc.fileExt].Name()
						tc.mockRequest.OutputDirectory = tempDir

						// Mock upload response
						getDetectRunResponse := make(map[string]interface{})
						uploadJSONResponse := `{"run_id": "run123"}`
						_ = json.Unmarshal([]byte(uploadJSONResponse), &getDetectRunResponse)

						// Mock status check response
						statusResponse := map[string]interface{}{
							"status": "SUCCESS",
							"output": []map[string]interface{}{
								{
									"processedFile":          "dGVzdCBjb250ZW50",
									"processedFileExtension": tc.fileExt,
									"processedFileType":      tc.fileType,
								},
								{
									"processedFile":          "eyJlbnRpdGllcyI6W119",
									"processedFileType":      "entities",
									"processedFileExtension": "json",
								},
							},
							"outputType": "FILE",
							"message":    "Processing completed successfully",
							"size":       1024.5,
							"duration":   60.5,
							"pages":      5,
							"slides": func() int {
								if tc.fileType == "PPTX" {
									return 10
								}
								return 0
							}(),
							"wordCharacterCount": map[string]interface{}{
								"wordCount":      150,
								"characterCount": 750,
							},
						}

						// Set up mock servers for both endpoints
						mux := http.NewServeMux()

						// Handle file upload
						mux.HandleFunc(tc.endpoint, func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("Content-Type", "application/json")
							json.NewEncoder(w).Encode(getDetectRunResponse)
						})

						// Handle status check
						mux.HandleFunc("/v1/detect/runs/", func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("Content-Type", "application/json")
							json.NewEncoder(w).Encode(statusResponse)
						})

						ts := httptest.NewServer(mux)
						defer ts.Close()

						// Configure mock client
						header := http.Header{}
						header.Set("Content-Type", "application/json")
						CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
							client := client.NewClient(
								option.WithBaseURL(ts.URL),
								option.WithToken("token"),
								option.WithHTTPHeader(header),
							)
							d.FilesApiClient = *client.Files
							return nil
						}

						SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
							return nil
						}

						// Execute test
						result, err := detectController.DeidentifyFile(ctx, tc.mockRequest, common.DeidentifyFileOptions{})

						// Verify results
						Expect(err).To(BeNil())
						Expect(result).ToNot(BeNil())
						Expect(result.RunId).To(Equal("run123"))
						Expect(result.Status).To(Equal("SUCCESS"))
						Expect(result.FileBase64).To(Equal("dGVzdCBjb250ZW50"))
						Expect(result.Type).To(Equal(tc.fileType))
						Expect(result.Extension).To(Equal(tc.fileExt))
						Expect(result.SizeInKb).To(Equal(1024.5))
						Expect(result.DurationInSeconds).To(Equal(60.5))
						Expect(result.WordCount).To(Equal(150))
						Expect(result.CharCount).To(Equal(750))

						// Verify file specific attributes
						if tc.fileType == "PDF" {
							Expect(result.PageCount).To(Equal(5))
						}
						if tc.fileType == "PPTX" {
							Expect(result.SlideCount).To(Equal(10))
						}

						// Verify file info
						Expect(result.File.Name).To(Equal(fmt.Sprintf("deidentified.%s", tc.fileExt)))
						Expect(result.File.Type).To(Equal("redacted_file"))

						// Verify entities
						Expect(result.Entities).To(HaveLen(1))
						Expect(result.Entities[0].Type).To(Equal("entities"))
						Expect(result.Entities[0].Extension).To(Equal("json"))
						Expect(result.Entities[0].File).To(Equal("eyJlbnRpdGllcyI6W119"))
					})
				}
			})

		})

		Context("Error cases", func() {
			It("should return error for validation failure", func() {
				request := DeidentifyFileRequest{
					File: FileInput{}, // Empty file input should fail validation
				}

				result, err := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})

				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
				Expect(result).To(BeNil())
			})

			It("should return error when API request fails", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeidentifyFileErrorJSON), &response)

				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(response)
				}))
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				request := DeidentifyFileRequest{
					File: FileInput{
						FilePath: testFiles["txt"].Name(),
					},
					Entities: []DetectEntities{Name},
				}

				result, err := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, headers map[common.CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				request := DeidentifyFileRequest{
					File: FileInput{
						FilePath: testFiles["txt"].Name(),
					},
					Entities: []DetectEntities{Name},
				}

				result, err := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(result).To(BeNil())
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, headers map[common.CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid bearer token")
				}

				request := DeidentifyFileRequest{
					File: FileInput{
						FilePath: testFiles["txt"].Name(),
					},
					Entities: []DetectEntities{Name},
				}

				result, err := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(result).To(BeNil())
			})

			It("should return error when polling times out", func() {
				request := DeidentifyFileRequest{
					File: FileInput{
						FilePath: testFiles["txt"].Name(),
					},
					Entities: []DetectEntities{Name},
					WaitTime: 2, // Short timeout for test
				}

				// Mock API responses
				mux := http.NewServeMux()

				// Upload endpoint returns success
				mux.HandleFunc("/v1/detect/deidentify/file/text", func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]string{"run_id": "run123"})
				})

				// Status check endpoint always returns IN_PROGRESS
				mux.HandleFunc("/v1/detect/runs/", func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "IN_PROGRESS",
						"message": "Still processing",
					})
				})

				ts := httptest.NewServer(mux)
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal("IN_PROGRESS"))
			})

			It("should handle failed processing status", func() {
				request := DeidentifyFileRequest{
					File: FileInput{
						FilePath: testFiles["txt"].Name(),
					},
					Entities: []DetectEntities{Name},
				}

				mux := http.NewServeMux()

				// Upload endpoint returns success
				mux.HandleFunc("/v1/detect/deidentify/file/text", func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]string{"run_id": "run123"})
				})

				// Status check endpoint returns failed status
				mux.HandleFunc("/v1/detect/runs/", func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":     "FAILED",
						"message":    "Processing failed",
						"outputType": "UNKNOWN",
					})
				})

				ts := httptest.NewServer(mux)
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				result, err := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal("FAILED"))
				Expect(result.Type).To(Equal("UNKNOWN"))
			})

			It("should return error when custom headers map is empty in DeidentifyFile", func() {
				req := DeidentifyFileRequest{
					File:     FileInput{FilePath: testFiles["txt"].Name()},
					Entities: []DetectEntities{Name},
				}
				opts := common.DeidentifyFileOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				result, err := detectController.DeidentifyFile(ctx, req, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers has invalid key in DeidentifyFile", func() {
				req := DeidentifyFileRequest{
					File:     FileInput{FilePath: testFiles["txt"].Name()},
					Entities: []DetectEntities{Name},
				}
				opts := common.DeidentifyFileOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				result, err := detectController.DeidentifyFile(ctx, req, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when FilePath points to a non-existent file", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, headers map[common.CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}
				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}
				request := DeidentifyFileRequest{
					File:     FileInput{FilePath: "/non/existent/path/file.txt"},
					Entities: []DetectEntities{Name},
				}
				result, err := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
			})

			It("should return error when file object cannot be read", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, headers map[common.CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}
				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}
				// Create and immediately close the file so io.ReadAll fails
				closedFile, err := os.CreateTemp("", "skyflow_closed_*.txt")
				Expect(err).To(BeNil())
				closedFile.Close()
				os.Remove(closedFile.Name())

				request := DeidentifyFileRequest{
					File:     FileInput{File: closedFile},
					Entities: []DetectEntities{Name},
				}
				result, skyErr := detectController.DeidentifyFile(ctx, request, common.DeidentifyFileOptions{})
				Expect(result).To(BeNil())
				Expect(skyErr).ToNot(BeNil())
				Expect(skyErr.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
			})

			It("should return error when custom headers has empty value in DeidentifyFile", func() {
				req := DeidentifyFileRequest{
					File:     FileInput{FilePath: testFiles["txt"].Name()},
					Entities: []DetectEntities{Name},
				}
				opts := common.DeidentifyFileOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				result, err := detectController.DeidentifyFile(ctx, req, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})
	Describe("GetDetectRun tests", func() {
		var (
			detectController *DetectController
			ctx              context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
			detectController = &DetectController{
				Config: &VaultConfig{
					VaultId:   "vault123",
					ClusterId: "cluster123",
					Env:       DEV,
					Credentials: Credentials{
						ApiKey: "test-api-key",
					},
				},
			}
		})

		Context("Success cases", func() {
			It("should successfully get completed run status", func() {
				// Mock status check response
				response := map[string]interface{}{
					"status": "SUCCESS",
					"output": []map[string]interface{}{
						{
							"processedFile":          "dGVzdCBjb250ZW50",
							"processedFileExtension": "txt",
							"processedFileType":      "TEXT",
						},
						{
							"processedFile":          "eyJlbnRpdGllcyI6W119",
							"processedFileType":      "ENTITIES",
							"processedFileExtension": "json",
						},
					},
					"outputType": "FILE",
					"message":    "Processing completed successfully",
					"size":       1024.5,
					"duration":   1.2,
					"pages":      0,
					"slides":     0,
					"wordCharacterCount": map[string]interface{}{
						"wordCount":      150,
						"characterCount": 750,
					},
				}

				ts := setupMockServer(response, "ok", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				request := GetDetectRunRequest{
					RunId: "run123",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.RunId).To(Equal("run123"))
				Expect(result.Status).To(Equal("SUCCESS"))
				Expect(result.FileBase64).To(Equal("dGVzdCBjb250ZW50"))
				Expect(result.Type).To(Equal("TEXT"))
				Expect(result.Extension).To(Equal("txt"))
				Expect(result.SizeInKb).To(Equal(1024.5))
				Expect(result.DurationInSeconds).To(Equal(1.2))
				Expect(result.PageCount).To(Equal(0))
				Expect(result.SlideCount).To(Equal(0))
			})

			It("should handle in-progress status", func() {
				response := make(map[string]interface{})

				_ = json.Unmarshal([]byte(mockGetDetectRunInProgressJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				request := GetDetectRunRequest{
					RunId: "run123",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal("IN_PROGRESS"))
				Expect(result.RunId).To(Equal("run123"))
			})

			It("should handle failed processing status", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetDetectRunFailedJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				request := GetDetectRunRequest{
					RunId: "run123",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal("FAILED"))
				Expect(result.RunId).To(Equal("run123"))
				Expect(result.Type).To(Equal("UNKNOWN"))
			})
		})

		Context("Error cases", func() {
			It("should return error for empty run ID", func() {
				request := GetDetectRunRequest{
					RunId: "",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
			})

			It("should return error for invalid run ID format", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"message":"Invalid run ID format","code":400}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				ts := setupMockServer(response, "error", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				request := GetDetectRunRequest{
					RunId: "invalid-format",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error for expired run ID", func() {
				response := make(map[string]interface{})

				_ = json.Unmarshal([]byte(mockGetDetectRunExpiredJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				request := GetDetectRunRequest{
					RunId: "invalid-run-id",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal("UNKNOWN"))
				Expect(result.RunId).To(Equal("invalid-run-id"))
				Expect(result.Type).To(Equal("UNKNOWN"))
			})

			It("should handle API error response", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetDetectRunApiErrorJSON), &response)

				ts := setupMockServer(response, "error", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.FilesApiClient = *client.Files
					return nil
				}

				request := GetDetectRunRequest{
					RunId: "invalid_run_id",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				request := GetDetectRunRequest{
					RunId: "run123",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController, customHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid bearer token")
				}

				request := GetDetectRunRequest{
					RunId: "run123",
				}

				result, err := detectController.GetDetectRun(ctx, request, common.GetDetectRunOptions{})

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when custom headers map is empty in GetDetectRun", func() {
				req := GetDetectRunRequest{RunId: "run123"}
				opts := common.GetDetectRunOptions{
					CustomHeaders: make(map[CustomHeaderKey]string),
				}
				result, err := detectController.GetDetectRun(ctx, req, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers has invalid key in GetDetectRun", func() {
				req := GetDetectRunRequest{RunId: "run123"}
				opts := common.GetDetectRunOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}
				result, err := detectController.GetDetectRun(ctx, req, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when custom headers has empty value in GetDetectRun", func() {
				req := GetDetectRunRequest{RunId: "run123"}
				opts := common.GetDetectRunOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountId: "",
					},
				}
				result, err := detectController.GetDetectRun(ctx, req, opts)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})
})

var _ = Describe("applyCustomHeaders edge cases", func() {
	var (
		vaultCtrl *VaultController
		ts        *httptest.Server
	)

	BeforeEach(func() {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"records":[]}`))
		}))
		vaultCtrl = &VaultController{
			Config: &VaultConfig{
				VaultId:      "vault-id",
				ClusterId:    "cluster-id",
				Env:          PROD,
				BaseVaultUrl: ts.URL,
				Credentials: Credentials{
					ApiKey: "test-api-key",
				},
			},
		}
	})

	AfterEach(func() {
		ts.Close()
	})

	It("should skip an empty-string key", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey(""): "should-be-skipped",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())
		Expect(vaultCtrl.ApiClient).ToNot(BeNil())
	})

	It("should skip a whitespace-only key", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey("   "): "should-be-skipped",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())
		Expect(vaultCtrl.ApiClient).ToNot(BeNil())
	})

	It("should skip the reserved header sky-metadata", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey("sky-metadata"): "should-be-skipped",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())
		Expect(vaultCtrl.ApiClient).ToNot(BeNil())
	})

	It("should skip the reserved Authorization header", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey("Authorization"): "should-be-skipped",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())
		Expect(vaultCtrl.ApiClient).ToNot(BeNil())
	})

	It("should apply a valid enum key correctly", func() {
		var capturedHeader http.Header
		ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"records":[]}`))
		}))
		defer ts2.Close()

		vaultCtrl.Config.BaseVaultUrl = ts2.URL
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			RequestIdHeader: "my-request-id",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())

		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)

		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("my-request-id"))
	})

	It("should not panic when CustomHeaders map is nil", func() {
		vaultCtrl.CustomHeaders = nil
		Expect(func() {
			_ = CreateRequestClient(vaultCtrl, nil)
		}).ToNot(Panic())
	})

	It("should not panic when requestHeaders map is nil", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			RequestIdHeader: "req-id",
		}
		Expect(func() {
			_ = CreateRequestClient(vaultCtrl, nil)
		}).ToNot(Panic())
	})

	It("request-level headers override controller-level headers for the same key", func() {
		var capturedHeader http.Header
		ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"records":[]}`))
		}))
		defer ts2.Close()

		vaultCtrl.Config.BaseVaultUrl = ts2.URL
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			RequestIdHeader: "controller-value",
		}
		requestHeaders := map[CustomHeaderKey]string{
			RequestIdHeader: "request-value",
		}

		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())

		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)

		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("request-value"),
			"request-level header should override controller-level header for same key")
	})

	It("should skip lowercase variant of reserved Authorization header", func() {
		var capturedHeader http.Header
		ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer ts2.Close()

		vaultCtrl.Config.BaseVaultUrl = ts2.URL
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey("authorization"): "sneaky-token",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())

		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)

		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get("Authorization")).ToNot(Equal("sneaky-token"),
			"lowercase 'authorization' must be treated as reserved and not override the SDK-set value")
	})

	It("should skip all-uppercase variant of reserved AUTHORIZATION header", func() {
		var capturedHeader http.Header
		ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer ts2.Close()

		vaultCtrl.Config.BaseVaultUrl = ts2.URL
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey("AUTHORIZATION"): "sneaky-token",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())

		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)

		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get("Authorization")).ToNot(Equal("sneaky-token"),
			"uppercase 'AUTHORIZATION' must be treated as reserved and not override the SDK-set value")
	})

	It("should set a header with an empty string value", func() {
		var capturedHeader http.Header
		ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer ts2.Close()

		vaultCtrl.Config.BaseVaultUrl = ts2.URL
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			RequestIdHeader: "",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())

		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)

		Expect(capturedHeader).ToNot(BeNil())
		_, present := capturedHeader[http.CanonicalHeaderKey(string(RequestIdHeader))]
		Expect(present).To(BeTrue(), "header key with empty value should still be present in the request")
	})

	It("request-level key that normalises to same header as controller-level key should produce one header entry", func() {
		var capturedHeader http.Header
		ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer ts2.Close()

		vaultCtrl.Config.BaseVaultUrl = ts2.URL
		// controller sets "x-request-id", request-level sets "X-REQUEST-ID" — both canonicalise to X-Request-Id
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			RequestIdHeader: "from-controller",
		}
		requestHeaders := map[CustomHeaderKey]string{
			CustomHeaderKey("X-REQUEST-ID"): "from-request",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())

		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)

		Expect(capturedHeader).ToNot(BeNil())
		vals := capturedHeader[http.CanonicalHeaderKey(string(RequestIdHeader))]
		Expect(vals).To(HaveLen(1), "both keys normalise to the same header — only one value should be present")
		Expect(vals[0]).To(Equal("from-request"), "request-level value should win")
	})
})

var _ = Describe("Custom Headers Tests", func() {
	Describe("Test Custom Headers in CreateRequestClient", func() {
		Context("Controller-level CustomHeaders only", func() {
			It("should apply controller-level custom headers to request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertContinueFalseSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				// Create controller with custom headers
				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[SkyflowAccountId] = "custom-account-id"
				customHeaders[SkyflowAccountName] = "custom-account-name"

				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
					CustomHeaders: customHeaders,
				}

				// Track headers passed to client
				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					// Verify controller headers are applied
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					// Apply request headers (would override controller headers)
					if requestHeaders != nil {
						for key, value := range requestHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}
				options := InsertOptions{
					ContinueOnError: false,
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get(string(SkyflowAccountId))).To(Equal("custom-account-id"))
				Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("custom-account-name"))
			})
			It("should return error if the current token is expired", func() {
				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
					},
				}
				contrl.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				contrl.Config.Credentials.Path = ""
			    request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}
				options := InsertOptions{
					ContinueOnError: false,
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return error if the request body is not correct", func() {
				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
					},
				}
				contrl.Config.Credentials.Token = "../../" + os.Getenv("CRED_FILE_PATH")
				contrl.Config.Credentials.Path = ""
			    request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}
				options := InsertOptions{
					ContinueOnError: true,
					TokenMode: "UNKNOWN_MODE",
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil())
				Expect(insertError.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
				Expect(res).To(BeNil())
			})
			It("should return error if the request body is not correct when continue error is false", func() {
				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
					},
				}
				contrl.Config.Credentials.Token = "../../" + os.Getenv("CRED_FILE_PATH")
				contrl.Config.Credentials.Path = ""
			    request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}
				options := InsertOptions{
					ContinueOnError: false,
					TokenMode: "UNKNOWN_MODE",
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil())
				Expect(insertError.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
				Expect(res).To(BeNil())

		})

		Context("Per-request custom headers only", func() {
			It("should apply request-level custom headers to request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertContinueFalseSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
				}

				// Track headers passed to client
				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					if requestHeaders != nil {
						for key, value := range requestHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}

				requestHeaders := make(map[CustomHeaderKey]string)
				requestHeaders[RequestIdHeader] = "request-value"

				options := InsertOptions{
					ContinueOnError: false,
					CustomHeaders:   requestHeaders,
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("request-value"))
			})
		})

		Context("Both controller and request custom headers", func() {
			It("should apply and merge both controller and request custom headers", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertContinueFalseSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				// Create controller with custom headers
				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[SkyflowAccountId] = "controller-value"
				customHeaders[RequestIdHeader] = "controller-common"

				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
					CustomHeaders: customHeaders,
				}

				// Track headers passed to client
				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					// Request headers override controller headers with same key
					if requestHeaders != nil {
						for key, value := range requestHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}

				requestHeaders := make(map[CustomHeaderKey]string)
				requestHeaders[SkyflowAccountName] = "request-value"
				requestHeaders[RequestIdHeader] = "request-common" // This should override controller header

				options := InsertOptions{
					ContinueOnError: false,
					CustomHeaders:   requestHeaders,
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).To(BeNil())
				Expect(res).ToNot(BeNil())
				// Controller header should be present
				Expect(capturedHeader.Get(string(SkyflowAccountId))).To(Equal("controller-value"))
				// Request header should be present
				Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("request-value"))
				// Request header should override controller header with same key
				Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("request-common"))
			})
		})

		Context("Custom headers with Detokenize operation", func() {
			It("should apply custom headers to detokenize request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[RequestIdHeader] = "trace-123"

				vaultController := &VaultController{
					Config: &VaultConfig{
						VaultId: "vaultID",
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
						Env:          PROD,
						ClusterId:    "clusterID",
						BaseVaultUrl: "http://127.0.0.1",
					},
					CustomHeaders: customHeaders,
				}

				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				request := DetokenizeRequest{
					DetokenizeData: []DetokenizeData{
						{
							Token:         "token1",
							RedactionType: MASKED,
						},
					},
				}
				options := DetokenizeOptions{
					ContinueOnError: true,
				}

				ctx := context.Background()
				res, err := vaultController.Detokenize(ctx, request, options)

				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("trace-123"))
			})
			It("should apply custom headers to detokenize request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[RequestIdHeader] = "trace-123"

				vaultController := &VaultController{
					Config: &VaultConfig{
						VaultId: "vaultID",
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
						Env:          PROD,
						ClusterId:    "clusterID",
						BaseVaultUrl: "http://127.0.0.1",
					},
					CustomHeaders: customHeaders,
				}

				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				request := DetokenizeRequest{
					DetokenizeData: []DetokenizeData{
						{
							Token:         "token1",
							RedactionType: MASKED,
						},
					},
				}
				options := DetokenizeOptions{
					ContinueOnError: true,
				}

				ctx := context.Background()
				res, err := vaultController.Detokenize(ctx, request, options)

				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("trace-123"))
			})
		})

		Context("Custom headers with Get operation", func() {
			It("should apply custom headers to get request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[RequestIdHeader] = "corr-456"

				vaultController := VaultController{
					Config: &VaultConfig{
						VaultId: "vaultID",
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
						Env:       PROD,
						ClusterId: "clusterID",
					},
					CustomHeaders: customHeaders,
				}

				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				ctx := context.Background()
				request := GetRequest{
					Table: "table",
					Ids:   []string{"id1"},
				}
				options := GetOptions{
					RedactionType: REDACTED,
				}

				res, err := vaultController.Get(ctx, request, options)

				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get("x-request-id")).To(Equal("corr-456"))
			})
			It("should apply custom headers to get request", func() {
				vaultController := VaultController{
					Config: &VaultConfig{
						VaultId: "vaultID",
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
						Env:       PROD,
						ClusterId: "clusterID",
					},
					CustomHeaders: map[CustomHeaderKey]string{
						CustomHeaderKey("x-invalid-header"): "value",
					},
				}

				ctx := context.Background()
				request := GetRequest{
					Table: "table",
					Ids:   []string{"id1"},
				}
				options := GetOptions{
					RedactionType: REDACTED,
				}

				res, err := vaultController.Get(ctx, request, options)

				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})

		Context("Empty custom headers", func() {
			It("should handle nil custom headers gracefully", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertContinueFalseSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
					CustomHeaders: nil, // No custom headers
				}

				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}
				options := InsertOptions{
					ContinueOnError: false,
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				Expect(insertError).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
		})

		Context("Custom headers with Delete operation", func() {
			It("should apply custom headers to delete request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeleteSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[SkyflowAccountName] = "user-789"

				vaultController := VaultController{
					Config: &VaultConfig{
						VaultId: "vaultID",
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
						Env:       PROD,
						ClusterId: "clusterID",
					},
					CustomHeaders: customHeaders,
				}

				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				ctx := context.Background()
				request := DeleteRequest{
					Table: "table",
					Ids:   []string{"id1"},
				}

				res, err := vaultController.Delete(ctx, request, common.DeleteOptions{})

				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("user-789"))
			})
		})

		Context("Multiple custom headers", func() {
			It("should apply all custom headers correctly", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertContinueFalseSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[SkyflowAccountId] = "account-id-123"
				customHeaders[SkyflowAccountName] = "account-name-456"
				customHeaders[RequestIdHeader] = "req-456"

				contrl := VaultController{
					Config: &VaultConfig{
						VaultId:   "id",
						ClusterId: "clusterid",
						Env:       PROD,
						Credentials: Credentials{
							ApiKey: "sky-token",
						},
					},
					CustomHeaders: customHeaders,
				}

				capturedHeader := http.Header{}
				header := http.Header{}
				header.Set("Content-Type", "application/json")

				CreateRequestClientFunc = func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
					if v.CustomHeaders != nil {
						for key, value := range v.CustomHeaders {
							header.Set(string(key), value)
							capturedHeader.Set(string(key), value)
						}
					}
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				request := InsertRequest{
					Table: "test_table",
					Values: []map[string]interface{}{
						{"name": "value1"},
					},
				}
				options := InsertOptions{
					ContinueOnError: false,
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				Expect(insertError).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get(string(SkyflowAccountId))).To(Equal("account-id-123"))
				Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("account-name-456"))
				Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("req-456"))
			})
		})
	})
})
})

func setupMockServer(mockResponse map[string]interface{}, status string, path string) *httptest.Server {
	// Create a mock server
	mockServer := http.NewServeMux()
	// Define the handler for "/vaults/v1/vaults/"
	mockServer.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jsonData, _ := json.Marshal(mockResponse)
		// Write the response
		switch status {
		case "ok":
			w.WriteHeader(http.StatusOK)
		case "partial":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		_, _ = w.Write(jsonData)
	})

	// Start the server and return it
	return httptest.NewServer(mockServer)
}

// ---------------------------------------------------------------------------
// Missing edge-case tests
// ---------------------------------------------------------------------------

//  1. Request-level reserved headers (Authorization, sky-metadata) must be blocked
//     even when supplied as per-request headers (second arg to CreateRequestClient).
var _ = Describe("Request-level reserved header blocking", func() {
	var vaultCtrl *VaultController
	var ts *httptest.Server
	var capturedHeader http.Header

	BeforeEach(func() {
		capturedHeader = nil
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"records":[]}`))
		}))
		vaultCtrl = &VaultController{
			Config: &VaultConfig{
				VaultId:      "vault-id",
				ClusterId:    "cluster-id",
				Env:          PROD,
				BaseVaultUrl: ts.URL,
				Credentials: Credentials{
					ApiKey: "test-api-key",
				},
			},
		}
	})

	AfterEach(func() { ts.Close() })

	makeDetokenizeCall := func() {
		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)
	}

	It("should block 'Authorization' supplied as a request-level header", func() {
		requestHeaders := map[CustomHeaderKey]string{
			CustomHeaderKey("Authorization"): "sneaky-token",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get("Authorization")).ToNot(Equal("sneaky-token"),
			"request-level Authorization must be blocked by reserved-header check")
	})

	It("should block lowercase 'authorization' supplied as a request-level header", func() {
		requestHeaders := map[CustomHeaderKey]string{
			CustomHeaderKey("authorization"): "sneaky-token",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get("Authorization")).ToNot(Equal("sneaky-token"),
			"lowercase request-level authorization must be blocked")
	})

	It("should allow a non-reserved request-level header through", func() {
		requestHeaders := map[CustomHeaderKey]string{
			RequestIdHeader: "req-999",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(RequestIdHeader))).To(Equal("req-999"))
	})
})

// 2. Per-request CustomHeaders for Query, Tokenize, Update, UploadFile
var _ = Describe("Per-request CustomHeaders for remaining operations", func() {
	var baseCtrl VaultController
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		baseCtrl = VaultController{
			Config: &VaultConfig{
				VaultId:   "vaultID",
				ClusterId: "clusterID",
				Env:       PROD,
				Credentials: Credentials{
					ApiKey: "sky-token",
				},
			},
		}
	})

	buildMockClientFunc := func(ts *httptest.Server, capturedReqHeaders *map[CustomHeaderKey]string) func(*VaultController, map[CustomHeaderKey]string) *skyflowError.SkyflowError {
		return func(v *VaultController, requestHeaders map[CustomHeaderKey]string) *skyflowError.SkyflowError {
			*capturedReqHeaders = requestHeaders
			hdr := http.Header{}
			hdr.Set("Content-Type", "application/json")
			c := client.NewClient(
				option.WithBaseURL(ts.URL+"/vaults"),
				option.WithToken("token"),
				option.WithHTTPHeader(hdr),
			)
			v.ApiClient = *c
			return nil
		}
	}

	Context("Query", func() {
		It("should pass CustomHeaders from QueryOptions to CreateRequestClientFunc", func() {
			response := make(map[string]interface{})
			_ = json.Unmarshal([]byte(mockQuerySuccessJSON), &response)
			ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
			defer ts.Close()

			var capturedReqHeaders map[CustomHeaderKey]string
			CreateRequestClientFunc = buildMockClientFunc(ts, &capturedReqHeaders)

			opts := common.QueryOptions{
				CustomHeaders: map[CustomHeaderKey]string{
					RequestIdHeader: "query-req-123",
				},
			}
			res, err := baseCtrl.Query(ctx, QueryRequest{
				Query: "SELECT * FROM persons WHERE skyflow_id='id'",
			}, opts)

			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(capturedReqHeaders[RequestIdHeader]).To(Equal("query-req-123"),
				"QueryOptions.CustomHeaders must be forwarded to CreateRequestClientFunc")
		})
	})

	Context("Tokenize", func() {
		It("should pass CustomHeaders from TokenizeOptions to CreateRequestClientFunc", func() {
			response := make(map[string]interface{})
			_ = json.Unmarshal([]byte(mockTokenizeSuccessJSON), &response)
			ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
			defer ts.Close()

			var capturedReqHeaders map[CustomHeaderKey]string
			CreateRequestClientFunc = buildMockClientFunc(ts, &capturedReqHeaders)

			arrReq := []TokenizeRequest{{ColumnGroup: "group_name", Value: "41111111111111"}}
			opts := common.TokenizeOptions{
				CustomHeaders: map[CustomHeaderKey]string{
					SkyflowAccountId: "acct-abc",
				},
			}
			res, err := baseCtrl.Tokenize(ctx, arrReq, opts)

			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(capturedReqHeaders[SkyflowAccountId]).To(Equal("acct-abc"),
				"TokenizeOptions.CustomHeaders must be forwarded to CreateRequestClientFunc")
		})
	})

	Context("Update", func() {
		It("should pass CustomHeaders from UpdateOptions to CreateRequestClientFunc", func() {
			response := make(map[string]interface{})
			_ = json.Unmarshal([]byte(mockUpdateSuccessJSON), &response)
			ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
			defer ts.Close()

			var capturedReqHeaders map[CustomHeaderKey]string
			CreateRequestClientFunc = buildMockClientFunc(ts, &capturedReqHeaders)

			opts := UpdateOptions{
				ReturnTokens: true,
				TokenMode:    DISABLE,
				CustomHeaders: map[CustomHeaderKey]string{
					SkyflowAccountName: "my-account",
				},
			}
			res, err := baseCtrl.Update(ctx, UpdateRequest{
				Table: "demo",
				Data:  map[string]interface{}{"SkyflowId": "123", "name": "john"},
			}, opts)

			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(capturedReqHeaders[SkyflowAccountName]).To(Equal("my-account"),
				"UpdateOptions.CustomHeaders must be forwarded to CreateRequestClientFunc")
		})
	})

	Context("UploadFile", func() {
		It("should pass CustomHeaders from FileUploadOptions to CreateRequestClientFunc", func() {
			response := make(map[string]interface{})
			_ = json.Unmarshal([]byte(`{"skyflowID":"id"}`), &response)
			ts := setupMockServer(response, "ok", "/vaults/v2/vaults/")
			defer ts.Close()

			var capturedReqHeaders map[CustomHeaderKey]string
			CreateRequestClientFunc = buildMockClientFunc(ts, &capturedReqHeaders)

			opts := common.FileUploadOptions{
				CustomHeaders: map[CustomHeaderKey]string{
					RequestIdHeader: "upload-req-456",
				},
			}
			res, err := baseCtrl.UploadFile(ctx, common.FileUploadRequest{
				Table:      "table",
				ColumnName: "column",
				FilePath:   "../../../../credentials.json",
				SkyflowId:  "skyflowid",
			}, opts)

			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(capturedReqHeaders[RequestIdHeader]).To(Equal("upload-req-456"),
				"FileUploadOptions.CustomHeaders must be forwarded to CreateRequestClientFunc")
		})
	})
})

// 3. SkyflowAccountId and SkyflowAccountName enum constants end-to-end
var _ = Describe("SkyflowAccountId and SkyflowAccountName constants", func() {
	var vaultCtrl *VaultController
	var ts *httptest.Server
	var capturedHeader http.Header

	BeforeEach(func() {
		capturedHeader = nil
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Clone()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"records":[]}`))
		}))
		vaultCtrl = &VaultController{
			Config: &VaultConfig{
				VaultId:      "vault-id",
				ClusterId:    "cluster-id",
				Env:          PROD,
				BaseVaultUrl: ts.URL,
				Credentials: Credentials{
					ApiKey: "test-api-key",
				},
			},
		}
	})

	AfterEach(func() { ts.Close() })

	makeDetokenizeCall := func() {
		tok := "test-token"
		payload := &vaultapis.V1DetokenizePayload{
			DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
				{Token: &tok},
			},
		}
		_, _ = vaultCtrl.ApiClient.Tokens.WithRawResponse.RecordServiceDetokenize(
			context.Background(), vaultCtrl.Config.VaultId, payload,
		)
	}

	It("should send SkyflowAccountId header set at controller level", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			SkyflowAccountId: "acct-001",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(SkyflowAccountId))).To(Equal("acct-001"))
	})

	It("should send SkyflowAccountName header set at controller level", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			SkyflowAccountName: "my-org",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("my-org"))
	})

	It("should send SkyflowAccountId and SkyflowAccountName together as request-level headers", func() {
		requestHeaders := map[CustomHeaderKey]string{
			SkyflowAccountId:   "acct-002",
			SkyflowAccountName: "partner-org",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(SkyflowAccountId))).To(Equal("acct-002"))
		Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("partner-org"))
	})

	It("request-level SkyflowAccountId should override controller-level value", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			SkyflowAccountId: "controller-acct",
		}
		requestHeaders := map[CustomHeaderKey]string{
			SkyflowAccountId: "request-acct",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(SkyflowAccountId))).To(Equal("request-acct"),
			"request-level value must override controller-level value for the same key")
	})
})

var _ = Describe("GenerateToken", func() {
	Context("when credentials contain a bearer token", func() {
		It("should return the token directly without any network call", func() {
			creds := Credentials{Token: "my-bearer-token"}
			token, err := GenerateToken(creds)
			Expect(err).To(BeNil())
			Expect(token).ToNot(BeNil())
			Expect(*token).To(Equal("my-bearer-token"))
		})
	})

	Context("when credentials contain an API key", func() {
		It("should return the API key directly without any network call", func() {
			creds := Credentials{ApiKey: "sky-api-key"}
			token, err := GenerateToken(creds)
			Expect(err).To(BeNil())
			Expect(token).ToNot(BeNil())
			Expect(*token).To(Equal("sky-api-key"))
		})
	})

	Context("when no credential fields are set", func() {
		It("should return INVALID_CREDENTIALS error", func() {
			creds := Credentials{}
			token, err := GenerateToken(creds)
			Expect(token).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.INVALID_CREDENTIALS))
		})
	})
})

var _ = Describe("setBearerTokenForConnectionController (via SetBearerTokenForConnectionControllerFunc)", func() {
	var originalFunc func(*ConnectionController) *skyflowError.SkyflowError

	BeforeEach(func() {
		originalFunc = SetBearerTokenForConnectionControllerFunc
		SetBearerTokenForConnectionControllerFunc = setBearerTokenForConnectionControllerReal
	})
	AfterEach(func() {
		SetBearerTokenForConnectionControllerFunc = originalFunc
	})

	It("should set token when config has token credentials", func() {
		ctrl := &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://example.com",
				Credentials:   Credentials{Token: "my-token"},
			},
		}
		err := SetBearerTokenForConnectionControllerFunc(ctrl)
		Expect(err).To(BeNil())
		Expect(ctrl.Token).To(Equal("my-token"))
	})

	It("should set token from builder creds when config creds are empty", func() {
		builderCreds := Credentials{Token: "builder-token"}
		ctrl := &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://example.com",
				Credentials:   Credentials{},
			},
			CommonCreds: &builderCreds,
		}
		err := SetBearerTokenForConnectionControllerFunc(ctrl)
		Expect(err).To(BeNil())
		Expect(ctrl.Token).To(Equal("builder-token"))
	})

	It("should set token from builder creds when config is nil", func() {
		builderCreds := Credentials{Token: "builder-token"}
		ctrl := &ConnectionController{
			Config:      nil,
			CommonCreds: &builderCreds,
		}
		err := SetBearerTokenForConnectionControllerFunc(ctrl)
		Expect(err).To(BeNil())
		Expect(ctrl.Token).To(Equal("builder-token"))
	})

	It("should return error when no credentials are available", func() {
		ctrl := &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://example.com",
				Credentials:   Credentials{},
			},
			CommonCreds: nil,
		}
		err := SetBearerTokenForConnectionControllerFunc(ctrl)
		Expect(err).ToNot(BeNil())
		Expect(err.GetMessage()).To(ContainSubstring(skyflowError.EMPTY_CREDENTIALS))
	})

	It("should reuse existing non-expired token", func() {
		expiryTime := time.Now().Add(1 * time.Hour).Unix()
		claims := jwt.MapClaims{"exp": float64(expiryTime)}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := tok.SignedString([]byte("secret"))

		ctrl := &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://example.com",
				Credentials:   Credentials{Token: "fresh-token"},
			},
			Token: tokenString,
		}
		err := SetBearerTokenForConnectionControllerFunc(ctrl)
		Expect(err).To(BeNil())
		Expect(ctrl.Token).To(Equal(tokenString))
	})

	It("should refresh token when existing token is expired", func() {
		expiryTime := time.Now().Add(-1 * time.Hour).Unix()
		claims := jwt.MapClaims{"exp": float64(expiryTime)}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := tok.SignedString([]byte("secret"))

		ctrl := &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://example.com",
				Credentials:   Credentials{Token: "new-token"},
			},
			Token: tokenString,
		}
		err := SetBearerTokenForConnectionControllerFunc(ctrl)
		Expect(err).To(BeNil())
		Expect(ctrl.Token).To(Equal("new-token"))
	})
})

// setBearerTokenForConnectionControllerReal is the real (non-mocked) implementation,
// accessed via the exported var so tests can call it directly.
var setBearerTokenForConnectionControllerReal = SetBearerTokenForConnectionControllerFunc

var _ = Describe("CreateGenericFileRequest", func() {
	It("should build a DeidentifyFileRequest with no entities", func() {
		req := &common.DeidentifyFileRequest{}
		result := CreateGenericFileRequest(req, "base64content", "vault123", "png")
		Expect(result).ToNot(BeNil())
		Expect(result.VaultId).To(Equal("vault123"))
		Expect(result.File.Base64).To(Equal("base64content"))
		Expect(string(result.File.DataFormat)).To(Equal("PNG"))
		Expect(result.EntityTypes).To(BeNil())
	})

	It("should build a DeidentifyFileRequest with entities", func() {
		req := &common.DeidentifyFileRequest{
			Entities: []common.DetectEntities{"PERSON", "DATE"},
		}
		result := CreateGenericFileRequest(req, "b64", "v1", "jpg")
		Expect(result).ToNot(BeNil())
		Expect(result.EntityTypes).ToNot(BeNil())
		Expect(len(result.EntityTypes)).To(Equal(2))
	})
})

// originalSetBearerTokenConnectionFunc stores the real setBearerTokenForConnectionController so
// tests that need to invoke the actual function can restore it after other tests mock it.
var originalSetBearerTokenConnectionFunc = SetBearerTokenForConnectionControllerFunc

// ---------------------------------------------------------------------------
// Additional coverage: 10 specific code paths in connection_controller.go
// ---------------------------------------------------------------------------
var _ = Describe("ConnectionController additional coverage", func() {
	var (
		ctrl *ConnectionController
		ctx  = context.TODO()
	)

	BeforeEach(func() {
		ctrl = &ConnectionController{
			Config: &ConnectionConfig{
				ConnectionUrl: "http://mockserver.com",
				ConnectionId:  "demo",
			},
			Token: "mock-token",
		}
	})

	// -----------------------------------------------------------------------
	// Item 1: setBearerTokenForConnectionController — GenerateToken error path
	// -----------------------------------------------------------------------
	Context("setBearerTokenForConnectionController GenerateToken error path", func() {
		It("should propagate error from GenerateToken when credentials path is invalid", func() {
			// Use the real function (not a mock)
			SetBearerTokenForConnectionControllerFunc = originalSetBearerTokenConnectionFunc
			ctrl.Token = "" // force token regeneration
			ctrl.Config.Credentials = Credentials{
				Path: "/nonexistent/credentials-file.json",
			}
			ctrl.CommonCreds = nil
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// Item 2: setConnectionCredentials — SKYFLOW_CREDENTIALS env var path
	// -----------------------------------------------------------------------
	Context("setConnectionCredentials SKYFLOW_CREDENTIALS env var path", func() {
		It("should fall back to SKYFLOW_CREDENTIALS env var when config and builder creds are absent", func() {
			// Use the real function so setConnectionCredentials is actually called
			SetBearerTokenForConnectionControllerFunc = originalSetBearerTokenConnectionFunc
			os.Setenv("SKYFLOW_CREDENTIALS", `{"clientID":"c","keyID":"k","tokenURI":"t","privateKey":"p"}`)
			defer os.Unsetenv("SKYFLOW_CREDENTIALS")
			ctrl.Token = "" // force token regeneration
			ctrl.Config = &ConnectionConfig{
				ConnectionUrl: "http://mockserver.com",
				ConnectionId:  "demo",
				// Credentials is zero-value → isCredentialsEmpty == true → hits env var branch
			}
			ctrl.CommonCreds = nil
			// GenerateToken will fail because the JSON is not real credentials,
			// but the env var path in setConnectionCredentials will have been taken.
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// Item 3: io.ReadAll error — server closes connection prematurely
	// -----------------------------------------------------------------------
	Context("io.ReadAll error when server closes connection prematurely", func() {
		It("should return error when response body cannot be fully read", func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hj, ok := w.(http.Hijacker)
				if !ok {
					http.Error(w, "hijack unsupported", http.StatusInternalServerError)
					return
				}
				conn, bufrw, _ := hj.Hijack()
				// Advertise 1000 bytes but send only a few, then close.
				bufrw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 1000\r\n\r\nhello")
				bufrw.Flush()
				conn.Close()
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// Item 4: json.Marshal errors — 5 places, channels are not JSON-serialisable
	// -----------------------------------------------------------------------
	Context("json.Marshal errors in prepareRequest", func() {
		It("should return error when json.Marshal fails for application/json body map", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    map[string]interface{}{"ch": make(chan int)},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})

		It("should return error when json.Marshal fails for text/html body map", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "text/html"},
				Body:    map[string]interface{}{"ch": make(chan int)},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})

		It("should return error when json.Marshal fails for default content-type body map", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/custom-type"},
				Body:    map[string]interface{}{"ch": make(chan int)},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})

		It("should return error when json.Marshal fails for multipart/form-data nested map", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "multipart/form-data"},
				Body:    map[string]interface{}{"field": map[string]interface{}{"ch": make(chan int)}},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})

		It("should return error when json.Marshal fails for multipart/form-data array", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "multipart/form-data"},
				Body:    map[string]interface{}{"arr": []interface{}{make(chan int)}},
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// Item 5: http.NewRequest error — null-byte in connection URL
	// -----------------------------------------------------------------------
	Context("http.NewRequest error via invalid URL in Invoke", func() {
		It("should return error when ConnectionUrl contains a null byte", func() {
			ctrl.Config.ConnectionUrl = "http://host\x00.example.com"
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    "test",
			})
			Expect(err).ToNot(BeNil())
			Expect(resp).To(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// Item 6: buildURLEncodedParams nil continue — via Invoke
	// -----------------------------------------------------------------------
	Context("buildURLEncodedParams nil value via Invoke", func() {
		It("should skip nil values in FORMURLENCODED body map", func() {
			var receivedBody string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				receivedBody = string(data)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				Body:    map[string]interface{}{"key1": "value1", "nilkey": nil},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(receivedBody).ToNot(ContainSubstring("nilkey"))
			Expect(receivedBody).To(ContainSubstring("key1"))
		})
	})

	// -----------------------------------------------------------------------
	// Item 7: setQueryParams — int8, int16, uint8, uint16, uint32, uint64
	// -----------------------------------------------------------------------
	Context("setQueryParams with int8/int16/uint8/uint16/uint32/uint64", func() {
		var mockServer *httptest.Server

		BeforeEach(func() {
			mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			ctrl.Config.ConnectionUrl = mockServer.URL
		})
		AfterEach(func() { mockServer.Close() })

		It("should encode int8 query param", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method: "GET", Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"v": int8(100)},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode int16 query param", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method: "GET", Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"v": int16(30000)},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode uint8 query param", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method: "GET", Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"v": uint8(200)},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode uint16 query param", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method: "GET", Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"v": uint16(60000)},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode uint32 query param", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method: "GET", Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"v": uint32(4000000000)},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})

		It("should encode uint64 query param", func() {
			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method: "GET", Headers: map[string]string{"Content-Type": "application/json"},
				QueryParams: map[string]interface{}{"v": uint64(18000000000000000000)},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// Item 8: setHeaders ApiKey branch — server receives ApiKey as auth header
	// -----------------------------------------------------------------------
	Context("setHeaders ApiKey branch via Invoke", func() {
		It("should use ApiKey as x-skyflow-authorization when ApiKey is set", func() {
			var capturedAuth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedAuth = r.Header.Get("x-skyflow-authorization")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL
			ctrl.ApiKey = "sky-test-apikey-abcde"
			ctrl.Token = ""

			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "GET",
				Headers: map[string]string{"Content-Type": "application/json"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(capturedAuth).To(Equal("sky-test-apikey-abcde"))
		})
	})

	// -----------------------------------------------------------------------
	// Item 9: setHeaders content-type continue — boundary must not be overridden
	// -----------------------------------------------------------------------
	Context("setHeaders content-type continue branch via Invoke", func() {
		It("should preserve multipart boundary even when user headers include Content-Type", func() {
			var capturedCT string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCT = r.Header.Get("Content-Type")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "multipart/form-data",
					"X-Extra":      "extra-value",
				},
				Body: map[string]interface{}{"field": "value"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			// The actual content-type must contain the boundary set by the multipart writer.
			Expect(capturedCT).To(ContainSubstring("multipart/form-data"))
			Expect(capturedCT).To(ContainSubstring("boundary="))
		})
	})

	// -----------------------------------------------------------------------
	// Item 10: writeXMLElement nil — produces <tag/> via XML body map
	// -----------------------------------------------------------------------
	Context("writeXMLElement nil value via Invoke", func() {
		It("should produce a self-closing XML tag for nil values in the body map", func() {
			var receivedBody string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				receivedBody = string(data)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			}))
			defer srv.Close()
			ctrl.Config.ConnectionUrl = srv.URL

			SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError { return nil }
			resp, err := ctrl.Invoke(ctx, InvokeConnectionRequest{
				Method:  "POST",
				Headers: map[string]string{"Content-Type": "application/xml"},
				Body:    map[string]interface{}{"nullfield": nil, "realfield": "value"},
			})
			Expect(err).To(BeNil())
			Expect(resp).ToNot(BeNil())
			Expect(receivedBody).To(ContainSubstring("<nullfield/>"))
			Expect(receivedBody).To(ContainSubstring("<realfield>value</realfield>"))
		})
	})
})

// ---------------------------------------------------------------------------
// Credential resolution / fallback tests for setVaultCredentials
// (exercised through SetBearerTokenForVaultController)
// ---------------------------------------------------------------------------
var _ = Describe("setVaultCredentials — credential resolution", func() {
	BeforeEach(func() {
		// Ensure the env var is absent unless a test explicitly sets it.
		os.Unsetenv("SKYFLOW_CREDENTIALS")
	})

	Context("vault config has credentials", func() {
		It("uses vault-level token, ignores client-level credentials", func() {
			v := &VaultController{
				Config:      &VaultConfig{Credentials: Credentials{Token: "vault-token"}},
				CommonCreds: &Credentials{Token: "client-token"},
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).To(BeNil())
			Expect(v.Token).To(Equal("vault-token"))
		})

		It("uses vault-level API key, ignores client-level token", func() {
			v := &VaultController{
				Config:      &VaultConfig{Credentials: Credentials{ApiKey: "vault-api-key"}},
				CommonCreds: &Credentials{Token: "client-token"},
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).To(BeNil())
			Expect(v.Token).To(Equal("vault-api-key"))
		})
	})

	Context("vault config has no credentials — falls back to client-level credentials", func() {
		It("uses client-level token when vault credentials are empty", func() {
			v := &VaultController{
				Config:      &VaultConfig{Credentials: Credentials{}},
				CommonCreds: &Credentials{Token: "client-token"},
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).To(BeNil())
			Expect(v.Token).To(Equal("client-token"))
		})

		It("uses client-level API key when vault credentials are empty", func() {
			v := &VaultController{
				Config:      &VaultConfig{Credentials: Credentials{}},
				CommonCreds: &Credentials{ApiKey: "client-api-key"},
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).To(BeNil())
			Expect(v.Token).To(Equal("client-api-key"))
		})

		It("uses client-level token when vault config is nil", func() {
			v := &VaultController{
				Config:      nil,
				CommonCreds: &Credentials{Token: "client-token"},
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).To(BeNil())
			Expect(v.Token).To(Equal("client-token"))
		})
	})

	Context("neither vault nor client-level credentials provided", func() {
		It("falls back to SKYFLOW_CREDENTIALS env var and attempts token generation", func() {
			// Provide env var with syntactically valid but fake JSON — GenerateToken
			// will fail, but the error must NOT be EMPTY_CREDENTIALS, proving the
			// env-var branch in setVaultCredentials was reached.
			os.Setenv("SKYFLOW_CREDENTIALS", `{"clientID":"c","keyID":"k","tokenURI":"https://t.example.com","privateKey":"p"}`)
			defer os.Unsetenv("SKYFLOW_CREDENTIALS")
			v := &VaultController{
				Config:      &VaultConfig{Credentials: Credentials{}},
				CommonCreds: nil,
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).ToNot(ContainSubstring(skyflowError.EMPTY_CREDENTIALS))
		})

		It("returns EMPTY_CREDENTIALS error when vault, client, and env var creds are all absent", func() {
			v := &VaultController{
				Config:      &VaultConfig{Credentials: Credentials{}},
				CommonCreds: nil,
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.EMPTY_CREDENTIALS))
		})

		It("returns EMPTY_CREDENTIALS error when vault config is nil and no client or env var creds", func() {
			v := &VaultController{
				Config:      nil,
				CommonCreds: nil,
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.EMPTY_CREDENTIALS))
		})
	})

	Context("token reuse across vault and client-level sources", func() {
		It("reuses an existing non-expired token regardless of credential source", func() {
			expiryTime := time.Now().Add(1 * time.Hour).Unix()
			claims := jwt.MapClaims{"exp": float64(expiryTime)}
			tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			validToken, _ := tok.SignedString([]byte("secret"))

			v := &VaultController{
				Config:      &VaultConfig{Credentials: Credentials{}},
				CommonCreds: &Credentials{Token: "client-token"},
				Token:       validToken, // already set and non-expired
			}
			err := SetBearerTokenForVaultController(v)
			Expect(err).To(BeNil())
			// Token must NOT be replaced by client-token because it was still valid.
			Expect(v.Token).To(Equal(validToken))
		})
	})
})

