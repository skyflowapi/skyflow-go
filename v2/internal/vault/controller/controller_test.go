package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	mockGetDetectRunInProgressJSON     = `{"status": "in_progress", "message": "Processing in progress"}`
	mockGetDetectRunFailedJSON         = `{"status": "FAILED", "message": "Processing failed", "outputType": "UNKNOWN"}`
	mockGetDetectRunExpiredJSON        = `{ "status": "UNKNOWN", "outputType": "UNKNOWN", "output": [], "message": "", "size": 0}`
	mockGetDetectRunApiErrorJSON       = `{"error": {"message": "Invalid run ID"}}`
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
			customHeader[RequestIDHeader] = "custom-header-value"
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
						SkyflowAccountID: "",
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
						SkyflowAccountID: "",
					},
				}
				res, err := vaultController.Detokenize(ctx, request, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
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
			}
			request := GetRequest{
				Table: "table",
				Ids:   []string{"id1"},
			}
			It("should return success response when valid ids passed in Get", func() {
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
						SkyflowAccountID: "",
					},
				}
				res, err := vaultController.Get(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
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
						SkyflowAccountID: "",
					},
				}
				res, err := vaultController.Delete(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
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
						SkyflowAccountID: "",
					},
				}
				res, err := vaultController.Query(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
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
						SkyflowAccountID: "",
					},
				}
				res, err := vaultController.Update(ctx, req, opts)
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
						SkyflowAccountID: "",
					},
				}
				res, err := vaultController.Tokenize(ctx, req, opts)
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
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
					SkyflowAccountID: "",
				},
			}
			res, err := vaultController.UploadFile(ctx, request, opts)
			Expect(err).ToNot(BeNil())
			Expect(res).To(BeNil())
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
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Length", "0")
					_, _ = w.Write([]byte(`67676`))
				}))
				ctrl.Config.ConnectionUrl = mockServer.URL
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, mockRequest)
				Expect(response).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
		Context("Invoke with different content types", func() {
			BeforeEach(func() {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			It("should correctly parse and set query parameters", func() {
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
			It("should correctly parse and set query parameters", func() {
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
		})
		Context("Handling Path parameters", func() {
			It("should correctly parse and set path parameters", func() {
				pathParams := map[string]string{"id": "123"}
				request := InvokeConnectionRequest{
					Method: "POST",
					Headers: map[string]string{
						"Content-Type": "multipart/form-data",
					},
					Body: map[string]interface{}{
						"key": "value",
					},
					PathParams: pathParams,
				}
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}
				response, err := ctrl.Invoke(ctx, request)
				Expect(err).ToNot(BeNil())
				Expect(response).To(BeNil())
			})
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
				RequestIDHeader:                   "req-123",
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
						SkyflowAccountID: "",
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
						SkyflowAccountID: "",
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

				// Status check endpoint always returns in_progress
				mux.HandleFunc("/v1/detect/runs/", func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"status":  "in_progress",
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
				Expect(result.Status).To(Equal("in_progress"))
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

			It("should return error when custom headers has empty value in DeidentifyFile", func() {
				req := DeidentifyFileRequest{
					File:     FileInput{FilePath: testFiles["txt"].Name()},
					Entities: []DetectEntities{Name},
				}
				opts := common.DeidentifyFileOptions{
					CustomHeaders: map[CustomHeaderKey]string{
						SkyflowAccountID: "",
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
				Expect(result.Status).To(Equal("in_progress"))
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
						SkyflowAccountID: "",
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
			RequestIDHeader: "my-request-id",
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
		Expect(capturedHeader.Get(string(RequestIDHeader))).To(Equal("my-request-id"))
	})

	It("should not panic when CustomHeaders map is nil", func() {
		vaultCtrl.CustomHeaders = nil
		Expect(func() {
			_ = CreateRequestClient(vaultCtrl, nil)
		}).ToNot(Panic())
	})

	It("should not panic when requestHeaders map is nil", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			RequestIDHeader: "req-id",
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
			RequestIDHeader: "controller-value",
		}
		requestHeaders := map[CustomHeaderKey]string{
			RequestIDHeader: "request-value",
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
		Expect(capturedHeader.Get(string(RequestIDHeader))).To(Equal("request-value"),
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
			RequestIDHeader: "",
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
		_, present := capturedHeader[http.CanonicalHeaderKey(string(RequestIDHeader))]
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
			RequestIDHeader: "from-controller",
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
		vals := capturedHeader[http.CanonicalHeaderKey(string(RequestIDHeader))]
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
				customHeaders[SkyflowAccountID] = "custom-account-id"
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
				Expect(capturedHeader.Get(string(SkyflowAccountID))).To(Equal("custom-account-id"))
				Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("custom-account-name"))
			})
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
				requestHeaders[RequestIDHeader] = "request-value"

				options := InsertOptions{
					ContinueOnError: false,
					CustomHeaders:   requestHeaders,
				}

				ctx := context.Background()
				res, insertError := contrl.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(capturedHeader.Get(string(RequestIDHeader))).To(Equal("request-value"))
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
				customHeaders[SkyflowAccountID] = "controller-value"
				customHeaders[RequestIDHeader] = "controller-common"

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
				requestHeaders[RequestIDHeader] = "request-common" // This should override controller header

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
				Expect(capturedHeader.Get(string(SkyflowAccountID))).To(Equal("controller-value"))
				// Request header should be present
				Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("request-value"))
				// Request header should override controller header with same key
				Expect(capturedHeader.Get(string(RequestIDHeader))).To(Equal("request-common"))
			})
		})

		Context("Custom headers with Detokenize operation", func() {
			It("should apply custom headers to detokenize request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[RequestIDHeader] = "trace-123"

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
				Expect(capturedHeader.Get(string(RequestIDHeader))).To(Equal("trace-123"))
			})
		})

		Context("Custom headers with Get operation", func() {
			It("should apply custom headers to get request", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetSuccessJSON), &response)

				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()

				customHeaders := make(map[CustomHeaderKey]string)
				customHeaders[RequestIDHeader] = "corr-456"

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
				customHeaders[SkyflowAccountID] = "account-id-123"
				customHeaders[SkyflowAccountName] = "account-name-456"
				customHeaders[RequestIDHeader] = "req-456"

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
				Expect(capturedHeader.Get(string(SkyflowAccountID))).To(Equal("account-id-123"))
				Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("account-name-456"))
				Expect(capturedHeader.Get(string(RequestIDHeader))).To(Equal("req-456"))
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
			RequestIDHeader: "req-999",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(RequestIDHeader))).To(Equal("req-999"))
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
					RequestIDHeader: "query-req-123",
				},
			}
			res, err := baseCtrl.Query(ctx, QueryRequest{
				Query: "SELECT * FROM persons WHERE skyflow_id='id'",
			}, opts)

			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(capturedReqHeaders[RequestIDHeader]).To(Equal("query-req-123"),
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
					SkyflowAccountID: "acct-abc",
				},
			}
			res, err := baseCtrl.Tokenize(ctx, arrReq, opts)

			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
			Expect(capturedReqHeaders[SkyflowAccountID]).To(Equal("acct-abc"),
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
					RequestIDHeader: "upload-req-456",
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
			Expect(capturedReqHeaders[RequestIDHeader]).To(Equal("upload-req-456"),
				"FileUploadOptions.CustomHeaders must be forwarded to CreateRequestClientFunc")
		})
	})
})

// 3. SkyflowAccountID and SkyflowAccountName enum constants end-to-end
var _ = Describe("SkyflowAccountID and SkyflowAccountName constants", func() {
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

	It("should send SkyflowAccountID header set at controller level", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			SkyflowAccountID: "acct-001",
		}
		err := CreateRequestClient(vaultCtrl, nil)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(SkyflowAccountID))).To(Equal("acct-001"))
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

	It("should send SkyflowAccountID and SkyflowAccountName together as request-level headers", func() {
		requestHeaders := map[CustomHeaderKey]string{
			SkyflowAccountID:   "acct-002",
			SkyflowAccountName: "partner-org",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(SkyflowAccountID))).To(Equal("acct-002"))
		Expect(capturedHeader.Get(string(SkyflowAccountName))).To(Equal("partner-org"))
	})

	It("request-level SkyflowAccountID should override controller-level value", func() {
		vaultCtrl.CustomHeaders = map[CustomHeaderKey]string{
			SkyflowAccountID: "controller-acct",
		}
		requestHeaders := map[CustomHeaderKey]string{
			SkyflowAccountID: "request-acct",
		}
		err := CreateRequestClient(vaultCtrl, requestHeaders)
		Expect(err).To(BeNil())
		makeDetokenizeCall()
		Expect(capturedHeader).ToNot(BeNil())
		Expect(capturedHeader.Get(string(SkyflowAccountID))).To(Equal("request-acct"),
			"request-level value must override controller-level value for the same key")
	})
})
