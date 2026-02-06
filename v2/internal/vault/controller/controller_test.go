package controller_test

import (
	// "bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	// "io"
	// "mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
			customHeader := make(map[string]string)
			customHeader["x-custom-header"] = "custom-header-value"
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					if (v.CustomHeaders != nil ) {
						for key, value := range v.CustomHeaders {
							header.Set(key, value)
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
				Expect(res.InsertedFields[0]["skyflow_id"]).To(Equal("skyflowid"))
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				Expect(res.InsertedFields[0]["skyflow_id"]).To(Equal("skyflowid"), "Expected the inserted field to have skyflow_id 'skyflowid'")
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				Expect(res.InsertedFields[0]["skyflow_id"]).To(Equal("skyflowid1"), "Expected first inserted field to have skyflow_id 'skyflowid1'")
				Expect(res.InsertedFields[1]["skyflow_id"]).To(Equal("skyflowid2"), "Expected second inserted field to have skyflow_id 'skyflowid2'")
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
					Env:       PROD,
					ClusterId: "clusterID",
					BaseVaultURL: "http://127.0.0.1",
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				// Call the Detokenize function
				res, err := vaultController.Detokenize(ctx, request, options)
				// Validate the response
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				mockJSONResponse := `{"records":[{"fields":{"name":"name1", "skyflow_id":"id1"}, "tokens":null}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Delete(ctx, request)
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Delete(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when invalid data passed in Delete", func() {
				request.Ids = []string{}
				res, err := vaultController.Delete(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Delete", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Delete(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				res, err := vaultController.Query(ctx, request)
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Query(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when invalid data passed in Query", func() {
				request.Query = ""
				res, err := vaultController.Query(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Query", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Query(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
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
				Data:   map[string]interface{}{"skyflow_id": "123", "name": "john"},
				Tokens: nil,
			}
			It("should return success response when valid ids passed in Update", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockUpdateSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Update(ctx, request, UpdateOptions{ReturnTokens: true, TokenMode: ENABLE_STRICT})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}

				res, err := vaultController.Tokenize(ctx, arrReq)
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				res, err := vaultController.Tokenize(ctx, arrReq)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when validations failed for invalid data passedin Tokenize", func() {
				arrReq = append(arrReq, TokenizeRequest{})
				res, err := vaultController.Tokenize(ctx, arrReq)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Tokenize", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				res, err := vaultController.Tokenize(ctx, arrReq)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
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
			CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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

			res, err := vaultController.UploadFile(ctx, request)
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
			CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
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

			res, err := vaultController.UploadFile(ctx, request)
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

			res, err := vaultController.UploadFile(ctx, request)
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(skyflowError.MISSING_FILE_SOURCE_IN_UPLOAD_FILE))
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
				// Should return as bytes when JSON parsing fails
				Expect(response.Data).To(Equal([]byte("invalid json content")))
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

var _ = Describe("Connection Utility Functions", func() {
	Describe("RUrlencode and renderKey", func() {
		It("should handle int values", func() {
			parents := make([]interface{}, 0)
			pairs := make(map[string]string)
			parents = append(parents, "count")
			result := RUrlencode(parents, pairs, 42)
			Expect(result["count"]).To(Equal("42"))
		})

		It("should handle float32 values", func() {
			parents := make([]interface{}, 0)
			pairs := make(map[string]string)
			parents = append(parents, "price")
			result := RUrlencode(parents, pairs, float32(19.99))
			Expect(result["price"]).To(ContainSubstring("19.99"))
		})

		It("should handle float64 values", func() {
			parents := make([]interface{}, 0)
			pairs := make(map[string]string)
			parents = append(parents, "amount")
			result := RUrlencode(parents, pairs, float64(99.95))
			Expect(result["amount"]).To(ContainSubstring("99.95"))
		})

		It("should handle bool values", func() {
			parents := make([]interface{}, 0)
			pairs := make(map[string]string)
			parents = append(parents, "active")
			result := RUrlencode(parents, pairs, true)
			Expect(result["active"]).To(Equal("true"))
		})

		It("should handle string values", func() {
			parents := make([]interface{}, 0)
			pairs := make(map[string]string)
			parents = append(parents, "name")
			result := RUrlencode(parents, pairs, "John Doe")
			Expect(result["name"]).To(Equal("John Doe"))
		})

		It("should handle nested map values", func() {
			parents := make([]interface{}, 0)
			pairs := make(map[string]string)
			parents = append(parents, "user")
			data := map[string]interface{}{
				"name": "Alice",
				"age":  30,
			}
			result := RUrlencode(parents, pairs, data)
			Expect(result["user[name]"]).To(Equal("Alice"))
			Expect(result["user[age]"]).To(Equal("30"))
		})

		It("should handle deeply nested map values", func() {
			parents := make([]interface{}, 0)
			pairs := make(map[string]string)
			parents = append(parents, "user")
			data := map[string]interface{}{
				"profile": map[string]interface{}{
					"firstName": "Bob",
					"age":       25,
				},
			}
			result := RUrlencode(parents, pairs, data)
			Expect(result["user[profile][firstName]"]).To(Equal("Bob"))
			Expect(result["user[profile][age]"]).To(Equal("25"))
		})
	})

	// Describe("writeFormData", func() {
	// 	It("should write form data with multiple types", func() {
	// 		buffer := new(bytes.Buffer)
	// 		writer := multipart.NewWriter(buffer)
			
	// 		requestBody := map[string]interface{}{
	// 			"string": "value",
	// 			"number": 123,
	// 			"bool":   true,
	// 			"nested": map[string]interface{}{
	// 				"key": "nested_value",
	// 			},
	// 		}
			
	// 		err := writeFormData(writer, requestBody)
	// 		Expect(err).To(BeNil())
	// 		writer.Close()
			
	// 		content := buffer.String()
	// 		Expect(content).To(ContainSubstring("string"))
	// 		Expect(content).To(ContainSubstring("value"))
	// 		Expect(content).To(ContainSubstring("number"))
	// 		Expect(content).To(ContainSubstring("123"))
	// 		Expect(content).To(ContainSubstring("bool"))
	// 		Expect(content).To(ContainSubstring("true"))
	// 		Expect(content).To(ContainSubstring("nested[key]"))
	// 		Expect(content).To(ContainSubstring("nested_value"))
	// 	})

	// 	It("should handle float values in form data", func() {
	// 		buffer := new(bytes.Buffer)
	// 		writer := multipart.NewWriter(buffer)
			
	// 		requestBody := map[string]interface{}{
	// 			"price": float64(19.99),
	// 		}
			
	// 		err := writeFormData(writer, requestBody)
	// 		Expect(err).To(BeNil())
	// 		writer.Close()
			
	// 		content := buffer.String()
	// 		Expect(content).To(ContainSubstring("price"))
	// 		Expect(content).To(ContainSubstring("19.99"))
	// 	})
	// })
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

			err := CreateRequestClient(vaultController)
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

			err := CreateRequestClient(vaultController)
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

			err := CreateRequestClient(vaultController)
			Expect(err).ToNot(BeNil())
		})
		It("should return an error if the token is expired", func() {
			vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			err := CreateRequestClient(vaultController)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
			vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			vaultController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

			err1 := SetBearerTokenForVaultController(vaultController)
			Expect(err1).To(BeNil())

			err2 := CreateRequestClient(vaultController)
			Expect(err2).ToNot(BeNil())
			Expect(err2.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))

		})
		It("should add apikey", func() {
			//vaultController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
			vaultController.Config.Credentials.Token = ""
			vaultController.Config.Credentials.Path = ""
			vaultController.Config.Credentials.ApiKey = "test-api-key"

			err := CreateRequestClient(vaultController)
			Expect(err).To(BeNil())
			//Expect(vaultController.Token).To(Equal(vaultController.Config.Credentials.ApiKey))
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

				err := CreateDetectRequestClient(detectController)
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

				err := CreateDetectRequestClient(detectController)
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

				err := CreateDetectRequestClient(detectController)
				Expect(err).ToNot(BeNil())
			})
			It("should return an error if the token is expired", func() {
				detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				err := CreateDetectRequestClient(detectController)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
				detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

				err1 := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err1).To(BeNil())

				err2 := CreateDetectRequestClient(detectController)
				Expect(err2).ToNot(BeNil())
				Expect(err2.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))

			})
			It("should add apikey", func() {
				//detectController.Config.Credentials.Token = os.Getenv("EXPIRED_TOKEN")
				detectController.Config.Credentials.Token = ""
				detectController.Config.Credentials.Path = ""
				detectController.Config.Credentials.ApiKey = "test-api-key"

				err := CreateDetectRequestClient(detectController)
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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyText(ctx, mockRequest)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyText(ctx, mockRequest)

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

				result, err := detectController.DeidentifyText(ctx, invalidRequest)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyText(ctx, mockRequest)
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid bearer token")
				}

				result, err := detectController.DeidentifyText(ctx, mockRequest)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyText(ctx, mockRequest)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyText(ctx, mockRequest)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.ReidentifyText(ctx, mockRequest)

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

				result, err := detectController.ReidentifyText(ctx, invalidRequest)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.ReidentifyText(ctx, mockRequest)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				result, err := detectController.ReidentifyText(ctx, mockRequest)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid bearer token")
				}

				result, err := detectController.ReidentifyText(ctx, mockRequest)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
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
						CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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
						result, err := detectController.DeidentifyFile(ctx, tc.mockRequest)

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

				result, err := detectController.DeidentifyFile(ctx, request)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyFile(ctx, request)
				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				request := DeidentifyFileRequest{
					File: FileInput{
						FilePath: testFiles["txt"].Name(),
					},
					Entities: []DetectEntities{Name},
				}

				result, err := detectController.DeidentifyFile(ctx, request)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
				Expect(result).To(BeNil())
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyFile(ctx, request)
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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyFile(ctx, request)
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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.DeidentifyFile(ctx, request)
				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal("FAILED"))
				Expect(result.Type).To(Equal("UNKNOWN"))
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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.GetDetectRun(ctx, request)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.GetDetectRun(ctx, request)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.GetDetectRun(ctx, request)

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

				result, err := detectController.GetDetectRun(ctx, request)

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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.GetDetectRun(ctx, request)
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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.GetDetectRun(ctx, request)
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
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
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

				result, err := detectController.GetDetectRun(ctx, request)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error when client creation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Failed to create client")
				}

				request := GetDetectRunRequest{
					RunId: "run123",
				}

				result, err := detectController.GetDetectRun(ctx, request)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error when bearer token validation fails", func() {
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid bearer token")
				}

				request := GetDetectRunRequest{
					RunId: "run123",
				}

				result, err := detectController.GetDetectRun(ctx, request)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
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
