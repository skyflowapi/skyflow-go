package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	client "github.com/skyflowapi/skyflow-go/v2/internal/generated/client"
	. "github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("Vault controller Test cases", func() {
	Describe("Test helper functions", func() {
		Context("Test Get URL Function", func() {
			tests := []struct {
				name      string
				env       Env
				clusterId string
				expected  string
			}{
				{
					name:      "Development Environment",
					env:       DEV,
					clusterId: "test-cluster",
					expected:  constants.SECURE_PROTOCOL + "test-cluster" + constants.DEV_DOMAIN,
				},
				{
					name:      "Production Environment",
					env:       PROD,
					clusterId: "prod-cluster",
					expected:  constants.SECURE_PROTOCOL + "prod-cluster" + constants.PROD_DOMAIN,
				},
				{
					name:      "Staging Environment",
					env:       STAGE,
					clusterId: "stage-cluster",
					expected:  constants.SECURE_PROTOCOL + "stage-cluster" + constants.STAGE_DOMAIN,
				},
				{
					name:      "Sandbox Environment",
					env:       SANDBOX,
					clusterId: "sandbox-cluster",
					expected:  constants.SECURE_PROTOCOL + "sandbox-cluster" + constants.SANDBOX_DOMAIN,
				},
				{
					name:      "Default Environment",
					env:       Env(9),
					clusterId: "default-cluster",
					expected:  constants.SECURE_PROTOCOL + "default-cluster" + constants.PROD_DOMAIN,
				},
			}

			for _, tt := range tests {
				// Use a sub-describe block for each test case
				test := tt // capture range variable
				It("returns the expected URL", func() {
					result := GetURLWithEnv(test.env, test.clusterId)
					Expect(result).To(Equal(test.expected))
				})
			}
		})
		Context("CreateInsertBulkBodyRequest", func() {
			var records []*vaultapis.V1FieldRecords

			tests := []struct {
				name         string
				request      InsertRequest
				options      InsertOptions
				expectedBody *vaultapis.RecordServiceInsertRecordBody
			}{
				{
					name: "Default behavior",
					request: InsertRequest{
						Values: []map[string]interface{}{
							{"field1": "value1"},
							{"field2": "value2"},
						},
					},
					options: InsertOptions{
						ReturnTokens: true,
						Upsert:       "upsert",
						TokenMode:    DISABLE,
					},
					expectedBody: func() *vaultapis.RecordServiceInsertRecordBody {
						tokenMode := true
						upsert := "upsert"
						byot := vaultapis.V1ByotDisable
						body := vaultapis.RecordServiceInsertRecordBody{}
						body.Tokenization = &tokenMode
						body.Upsert = &upsert
						body.Byot = &byot
						body.Records = []*vaultapis.V1FieldRecords{
							{Fields: map[string]interface{}{"field1": "value1"}},
							{Fields: map[string]interface{}{"field2": "value2"}},
						}
						return &body
					}(),
				},
				{
					name: "With tokens",
					request: InsertRequest{
						Values: []map[string]interface{}{
							{"field1": "value1"},
						},
					},
					options: InsertOptions{
						ReturnTokens: true,
						Upsert:       "upsert",
						Tokens: []map[string]interface{}{
							{"token1": "value1_token"},
						},
						TokenMode: ENABLE_STRICT,
					},
					expectedBody: func() *vaultapis.RecordServiceInsertRecordBody {
						tokens := true
						upsert := "upsert"
						tokenMode := vaultapis.V1ByotEnableStrict
						body := vaultapis.RecordServiceInsertRecordBody{}
						body.Tokenization = &tokens
						body.Upsert = &upsert
						body.Byot = &tokenMode
						body.Records = []*vaultapis.V1FieldRecords{
							{
								Fields: map[string]interface{}{"field1": "value1"},
								Tokens: map[string]interface{}{"token1": "value1_token"},
							},
						}
						return &body
					}(),
				},
				{
					name: "Empty input",
					request: InsertRequest{
						Values: []map[string]interface{}{},
					},
					options: InsertOptions{
						ReturnTokens: false,
						Upsert:       "upsert",
						TokenMode:    ENABLE,
					},
					expectedBody: func() *vaultapis.RecordServiceInsertRecordBody {
						tokens := false
						upsert := "upsert"
						tokenMode := vaultapis.V1ByotEnable
						body := vaultapis.RecordServiceInsertRecordBody{}
						body.Tokenization = &tokens
						body.Upsert = &upsert
						body.Byot = &tokenMode
						body.Records = records
						return &body
					}(),
				},
			}

			for _, test := range tests {
				// Capture the current range variable to avoid closure issues
				test := test

				It("should create the correct request body", func() {
					actualBody, err := CreateInsertBulkBodyRequest(&test.request, &test.options)
					Expect(err).To(BeNil())
					Expect(*actualBody.Tokenization).To(Equal(*test.expectedBody.Tokenization))
					Expect(actualBody.Byot).To(Equal(test.expectedBody.Byot))
					Expect(actualBody.Upsert).To(Equal(test.expectedBody.Upsert))
					Expect(actualBody.Records).To(Equal(test.expectedBody.Records))
				})
			}

		})
		Context("SetTokenMode", func() {
			tests := []struct {
				name         string
				tokenMode    BYOT
				expectedByot vaultapis.V1Byot
			}{
				{
					name:         "Enable Strict Mode",
					tokenMode:    ENABLE_STRICT,
					expectedByot: vaultapis.V1ByotEnableStrict,
				},
				{
					name:         "Enable Mode",
					tokenMode:    ENABLE,
					expectedByot: vaultapis.V1ByotEnable,
				},
				{
					name:         "Default Disable Mode",
					tokenMode:    DISABLE,
					expectedByot: vaultapis.V1ByotDisable,
				},
				{
					name:         "Unknown Mode Defaults to Disable",
					tokenMode:    BYOT("UNKNOWN"),
					expectedByot: vaultapis.V1ByotDisable,
				},
			}

			for index, test := range tests {
				test := test // capture range variable

				It(fmt.Sprintf("should set the correct token mode %v", index), func() {
					body := vaultapis.RecordServiceBatchOperationBody{}

					mode, err := SetTokenMode(test.tokenMode)
					if err != nil {
						return
					}
					body.Byot = mode

					Expect(*mode).To(Equal(test.expectedByot),
						"Expected token mode to be %v but got %v", test.expectedByot, body.Byot)
				})
			}
		})

		type testCase struct {
			name          string
			record        interface{}
			requestIndex  int
			expected      map[string]interface{}
			expectedError error
		}
		Context("GetFormattedBatchInsertRecord", func() {
			var tests = []testCase{
				{
					name: "Valid record with skyflow_id and tokens",
					record: map[string]interface{}{
						"Body": map[string]interface{}{
							"records": []interface{}{
								map[string]interface{}{
									"skyflow_id": "12345",
									"tokens": map[string]interface{}{
										"token1": "value1",
										"token2": "value2",
									},
								},
							},
						},
					},
					requestIndex: 1,
					expected: map[string]interface{}{
						"skyflow_id":    "12345",
						"token1":        "value1",
						"token2":        "value2",
						"request_index": 1,
					},
					expectedError: nil,
				},
				{
					name: "Record missing Body field",
					record: map[string]interface{}{
						"SomeOtherField": "value",
					},
					requestIndex:  2,
					expected:      nil,
					expectedError: fmt.Errorf("Body field not found in JSON"),
				},
				{
					name: "Record with error field",
					record: map[string]interface{}{
						"Body": map[string]interface{}{
							"error": "Some error occurred",
						},
					},
					requestIndex: 3,
					expected: map[string]interface{}{
						"error":         "Some error occurred",
						"request_index": 3,
					},
					expectedError: nil,
				},
				{
					name: "Invalid record data type in records",
					record: map[string]interface{}{
						"Body": map[string]interface{}{
							"records": []interface{}{"invalid_record"},
						},
					},
					requestIndex: 4,
					expected: map[string]interface{}{
						"request_index": 4,
					},
					expectedError: nil,
				},
				{
					name:          "Failed to marshal record",
					record:        func() {}, // invalid type
					requestIndex:  5,
					expected:      nil,
					expectedError: fmt.Errorf("failed to marshal record"),
				},
			}

			for _, test := range tests {
				test := test // capture range variable
				It("should return the expected result or error", func() {
					result, err := GetFormattedBatchInsertRecord(test.record, test.requestIndex)

					if test.expectedError != nil {
						Expect(err).To(HaveOccurred())
						Expect(err.GetMessage()).To(ContainSubstring(skyflowError.INVALID_RESPONSE))
					} else {
						Expect(err).To(BeNil())
					}

					Expect(reflect.DeepEqual(result, test.expected)).To(BeTrue(),
						"Expected result: %v, got: %v", test.expected, result)
				})
			}
		})
		Context("GetFormattedBulkInsertRecord", func() {
			var (
				skyflowid = "12345"
				emptyid   = ""
			)

			tests := []struct {
				name     string
				record   vaultapis.V1RecordMetaProperties
				expected map[string]interface{}
			}{
				{
					name: "Record with skyflowId and tokens",
					record: vaultapis.V1RecordMetaProperties{
						SkyflowId: &skyflowid,
						Tokens: map[string]interface{}{
							"token1": "value1",
							"token2": "value2",
						},
					},
					expected: map[string]interface{}{
						"skyflow_id": skyflowid,
						"token1":     "value1",
						"token2":     "value2",
					},
				},
				{
					name: "Record with skyflowId and no tokens",
					record: vaultapis.V1RecordMetaProperties{
						SkyflowId: &skyflowid,
					},
					expected: map[string]interface{}{
						"skyflow_id": skyflowid,
					},
				},
				{
					name: "Record with no skyflowId and tokens",
					record: vaultapis.V1RecordMetaProperties{
						SkyflowId: &emptyid,
						Tokens: map[string]interface{}{
							"tokenA": "valueA",
						},
					},
					expected: map[string]interface{}{
						"skyflow_id": emptyid,
						"tokenA":     "valueA",
					},
				},
				{
					name: "Record with no skyflowId and no tokens",
					record: vaultapis.V1RecordMetaProperties{
						SkyflowId: &emptyid,
						Tokens:    map[string]interface{}{},
					},
					expected: map[string]interface{}{
						"skyflow_id": emptyid,
					},
				},
			}

			for _, test := range tests {
				test := test // capture the loop variable
				It("should return the expected formatted record", func() {
					// Call the function
					result := GetFormattedBulkInsertRecord(test.record)
					// Validate the result
					Expect(result).To(Equal(test.expected))
				})
			}
		})
	})
	Describe("Test Insert functions", func() {
		var (
			mockJSONResponse string
			response         map[string]interface{}
			contrl           VaultController
			ts               *httptest.Server
		)

		BeforeEach(func() {
			response = make(map[string]interface{})
			ts = nil
			contrl = VaultController{
				Config: VaultConfig{
					VaultId:   "id",
					ClusterId: "clusterid",
					Env:       PROD,
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
				},
			}
		})

		AfterEach(func() {
			if ts != nil {
				ts.Close()
			}
		})

		Context("Insert with ContinueOnError True - Success Case", func() {
			BeforeEach(func() {
				mockJSONResponse = `{"vaultID":"id", "responses":[{"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name_on_card":"token1"}}]}, "Status":200}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				// Setup mock server
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
				const mockJSONResponse = `{"vaultID":"id", "responses":[{"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}, {"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}]}`
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
					Config: VaultConfig{
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
					Config: VaultConfig{
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
					Config: VaultConfig{
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
				// Mock JSON response
				mockJSONResponse = `{"records":[{"skyflow_id":"skyflowid1", "tokens":{"name":"nameToken1"}}, {"skyflow_id":"skyflowid2", "tokens":{"expiry_month":"monthToken", "name":"nameToken2"}}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

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
				const mockJSONResponse = `{"error":{"grpc_code":3,"http_code":400,"message":"Insert failed. Table name card_detail is invalid. Specify a valid table name.","http_status":"Bad Request","details":[]}}`
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
					Config: VaultConfig{
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
				Expect(insertError).ToNot(BeNil(), "Expected error during insert operation")
				Expect(res).To(BeNil(), "Expected no response")
			})
		})
		Context("Insert Client Creation Failed", func() {
			It("should return an error when client creation fails", func() {
				const mockJSONResponse = `{"vaultID":"id", "responses":[{"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name_on_card":"token1"}}]}, "Status":200}]}`
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
				Config: VaultConfig{
					VaultId: "vaultID",
					Credentials: Credentials{
						ApiKey: "sky-token",
					},
					Env:       PROD,
					ClusterId: "clusterID",
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
		Context("Test Detokenize payload", func() {
			var (
				token1               = "token1"
				token2               = "token2"
				redaction            = vaultapis.RedactionEnumRedactionMasked
				ContinueOnError      = true
				ContinueOnErrorFalse = false
			)

			tests := []struct {
				name     string
				request  DetokenizeRequest
				options  DetokenizeOptions
				expected vaultapis.V1DetokenizePayload
			}{
				{
					name: "Test with valid tokens and redaction type",
					request: DetokenizeRequest{
						DetokenizeData: []DetokenizeData{{
							Token:         "token1",
							RedactionType: MASKED,
						}, {
							Token:         "token2",
							RedactionType: MASKED,
						}},
					},
					options: DetokenizeOptions{
						ContinueOnError: true,
					},
					expected: vaultapis.V1DetokenizePayload{
						DetokenizationParameters: []*vaultapis.V1DetokenizeRecordRequest{
							{
								Token:     &token1,
								Redaction: &redaction,
							},
							{
								Token:     &token2,
								Redaction: &redaction,
							},
						},
						ContinueOnError: &ContinueOnError,
					},
				},
				{
					name: "Test with no tokens",
					request: DetokenizeRequest{
						DetokenizeData: nil,
					},
					options: DetokenizeOptions{
						ContinueOnError: false,
					},
					expected: vaultapis.V1DetokenizePayload{
						DetokenizationParameters: nil,
						ContinueOnError:          &ContinueOnErrorFalse,
					},
				},
			}

			// Iterate over the test cases
			for _, test := range tests {
				Context(test.name, func() {
					It(test.name, func() {
						// Call the function being tested
						result := GetDetokenizePayload(test.request, test.options)

						// Compare the result with the expected value using Gomega's Expect
						Expect(result).To(Equal(test.expected))
					})
				})
			}
		})
		Context("When Detokenize is called", func() {
			It("should return detokenized data with no errors", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"records":[{"token":"token", "valueType":"STRING", "value":"*REDACTED*", "error":null}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				Expect(res.DetokenizedFields[0]["Token"]).To(Equal("token"))
				Expect(res.DetokenizedFields[0]["Value"]).To(Equal("*REDACTED*"))
				Expect(res.DetokenizedFields[0]["ValueType"]).To(Equal(vaultapis.DetokenizeRecordResponseValueType("STRING")))
			})
			It("should return detokenized data with errors", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Detokenize failed. All tokens are invalid. Specify valid tokens.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
					Config: VaultConfig{
						VaultId:     "vaultID",
						Credentials: Credentials{Token: "token"},
						Env:         PROD,
						ClusterId:   "clusterID",
					},
				}
				response := make(map[string]interface{})
				mockJSONResponse := `{"records":[{"token":"token1", "valueType":"STRING", "value":"*REDACTED*", "error":null}, {"token":"token1", "valueType":"NONE", "value":"", "error":"Token Not Found"}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				Config: VaultConfig{
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
				mockJSONResponse := `{"records":[{"fields":{"name":"name1", "skyflow_id":"id1"}, "tokens":null}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Get failed. [faild fail] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				Config: VaultConfig{
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
				mockJSONResponse := `{"RecordIDResponse":["id1"]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Delete failed. [id1] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				Config: VaultConfig{
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
				mockJSONResponse := `{"records":[{"fields":{"counter":null, "country":null, "date_of_birth":"XXXX-06-06", "email":"s******y@gmail.com", "name":"m***me", "phone_number":"XXXXXX8889", "skyflow_id":"id"}, "tokens":null}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Invalid request. Table name cards is invalid. Specify a valid table name.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				Config: VaultConfig{
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
				Id:     "skyflowid",
				Values: map[string]interface{}{"name": "john"},
				Tokens: nil,
			}
			It("should return success response when valid ids passed in Update", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"skyflow_id":"id","tokens":{"name":"token"}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
			})

			It("should return error response when invalid data passed in Update", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"grpc_code":3,"http_code":400,"message":"Invalid request. No fields were present. Specify valid fields and values.","http_status":"Bad Request","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				Config: VaultConfig{
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
				mockJSONResponse := `{"records":[{"token":"token1"}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
				mockJSONResponse := `{"error":{"grpc_code":3,"http_code":400,"message":"Tokenization failed. Column group group_name is invalid. Specify a valid column group.","http_status":"Bad Request","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
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
			Config: ConnectionConfig{
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
			Config: VaultConfig{
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
			Expect(err).ToNot(BeNil())
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

			vaultController.Config.Credentials.Path = ""
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
			vaultController.Config.Credentials.ApiKey = "sky-abcde-1234567890abcdef1234567890abcdef"

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
				Config: VaultConfig{
					Credentials: Credentials{
						Path: "test/path",
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
				Expect(err).ToNot(BeNil())
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
			It("should reuse token if valid token is provided", func() {
				detectController.Token = ""
				detectController.Config.Credentials.Path = "../../" + os.Getenv("CRED_FILE_PATH")

				err := SetBearerTokenForDetectControllerFunc(detectController)
				Expect(err).To(BeNil())
				Expect(detectController.Token).ToNot(BeNil())

				detectController.Config.Credentials.Path = ""
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
				detectController.Config.Credentials.ApiKey = "sky-abcde-1234567890abcdef1234567890abcdef"

				err := CreateDetectRequestClient(detectController)
				Expect(err).To(BeNil())
			})

		})
	})
	Describe("CreateDeidentifyTextRequest", func() {
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

				payload, err := CreateDeidentifyTextRequest(req)
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

		Context("when given an invalid entity", func() {
			It("should return an error", func() {
				req := DeidentifyTextRequest{
					Text:     "Sensitive text",
					Entities: []DetectEntities{"invalid_entity"},
				}
				payload, err := CreateDeidentifyTextRequest(req)
				Expect(payload).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})
		})

		Context("when given an invalid token format", func() {
			It("should return an error", func() {
				req := DeidentifyTextRequest{
					Text: "Sensitive text",
					TokenFormat: TokenFormat{
						DefaultType: "invalid_token_type",
					},
				}
				payload, err := CreateDeidentifyTextRequest(req)
				Expect(payload).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})
		})

		Context("when AllowRegexList and RestrictRegexList are empty", func() {
			It("should not set AllowRegex or RestrictRegex in payload", func() {
				req := DeidentifyTextRequest{
					Text: "Sensitive text",
				}
				payload, err := CreateDeidentifyTextRequest(req)
				Expect(err).To(BeNil())
				Expect(payload).ToNot(BeNil())
				Expect(payload.AllowRegex).To(BeNil())
				Expect(payload.RestrictRegex).To(BeNil())
			})
		})

		Context("when TokenFormat.EntityOnly contains invalid entity", func() {
			It("should return an error", func() {
				req := DeidentifyTextRequest{
					Text: "Sensitive text",
					TokenFormat: TokenFormat{
						EntityOnly: []DetectEntities{"invalid_entity"},
					},
				}
				payload, err := CreateDeidentifyTextRequest(req)
				Expect(payload).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})
		})

		Context("when TokenFormat.VaultToken contains invalid entity", func() {
			It("should return an error", func() {
				req := DeidentifyTextRequest{
					Text: "Sensitive text",
					TokenFormat: TokenFormat{
						VaultToken: []DetectEntities{"invalid_entity"},
					},
				}
				payload, err := CreateDeidentifyTextRequest(req)
				Expect(payload).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})
		})

		Context("when ShiftDates.Entities is empty", func() {
			It("should not set Transformations.ShiftDates in payload", func() {
				req := DeidentifyTextRequest{
					Text: "Sensitive text",
					Transformations: Transformations{
						ShiftDates: DateTransformation{
							MaxDays:  5,
							MinDays:  2,
							Entities: []TransformationsShiftDatesEntityTypesItem{},
						},
					},
				}
				payload, err := CreateDeidentifyTextRequest(req)
				Expect(err).To(BeNil())
				Expect(payload).ToNot(BeNil())
				if payload.Transformations != nil {
					Expect(payload.Transformations.ShiftDates).To(BeNil())
				}
			})
		})

		Context("when ShiftDates.MaxDays and MinDays are zero", func() {
			It("should not set MaxDays or MinDays in payload", func() {
				req := DeidentifyTextRequest{
					Text: "Sensitive text",
					Transformations: Transformations{
						ShiftDates: DateTransformation{
							MaxDays: 0,
							MinDays: 0,
							Entities: []TransformationsShiftDatesEntityTypesItem{
								TransformationsShiftDatesEntityTypesItemDate,
							},
						},
					},
				}
				payload, err := CreateDeidentifyTextRequest(req)
				Expect(err).To(BeNil())
				Expect(payload).ToNot(BeNil())
				Expect(payload.Transformations.ShiftDates.MaxDays).To(BeNil())
				Expect(payload.Transformations.ShiftDates.MinDays).To(BeNil())
			})
		})

	})

	Describe("CreateReidentifyTextRequest", func() {
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
				Expect(payload.VaultId).To(Equal(config.VaultId))
				Expect(payload.Text).To(Equal(request.Text))
				Expect(payload.Format.Redacted).To(HaveLen(2))
				Expect(payload.Format.Masked).To(HaveLen(1))
				Expect(payload.Format.Plaintext).To(HaveLen(1))
			})
		})

		Context("when handling invalid entities", func() {
			It("should return error for invalid redacted entities", func() {
				request := ReidentifyTextRequest{
					Text:             "Sample text",
					RedactedEntities: []DetectEntities{"INVALID_ENTITY"},
				}

				payload, err := CreateReidentifyTextRequest(request, config)

				Expect(payload).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error for invalid masked entities", func() {
				request := ReidentifyTextRequest{
					Text:           "Sample text",
					MaskedEntities: []DetectEntities{"INVALID_ENTITY"},
				}

				payload, err := CreateReidentifyTextRequest(request, config)

				Expect(payload).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})

			It("should return error for invalid plaintext entities", func() {
				request := ReidentifyTextRequest{
					Text:              "Sample text",
					PlainTextEntities: []DetectEntities{"INVALID_ENTITY"},
				}

				payload, err := CreateReidentifyTextRequest(request, config)

				Expect(payload).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
			})
		})
	})
	Describe("DeidentifyText tests", func() {
		var (
			detectController *DetectController
			ctx             context.Context
			mockRequest     DeidentifyTextRequest
		)
	
		BeforeEach(func() {
			ctx = context.Background()
			detectController = &DetectController{
				Config: VaultConfig{
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
				mockJSONResponse := `{
					"processed_text": "My name is [NAME] and email is [EMAIL]",
					"word_count": 8,
					"character_count": 45,
					"entities": [
						{
							"token": "token1",
							"value": "John Doe",
							"entity_type": "NAME",
							"entity_scores": {"score": 0.9},
							"location": {
								"start_index": 11,
								"end_index": 19,
								"start_index_processed": 11,
								"end_index_processed": 17
							}
						},
						{
							"token": "token2",
							"value": "john@example.com",
							"entity_type": "EMAIL_ADDRESS",
							"entity_scores": {"score": 0.95},
							"location": {
								"start_index": 30,
								"end_index": 45,
								"start_index_processed": 30,
								"end_index_processed": 37
							}
						}
					]
				}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
	
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
				Expect(result.CharacterCount).To(Equal(int(45)))
				Expect(result.Entities).To(HaveLen(2))
				Expect(result.Entities[0].Entity).To(Equal("NAME"))
				Expect(result.Entities[1].Entity).To(Equal("EMAIL_ADDRESS"))
			})
	
			It("should handle empty entities array in response", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{
					"processed_text": "No entities found in this text",
					"word_count": 6,
					"character_count": 30
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
				mockJSONResponse := `{"error":{"message":"Invalid request"}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
	
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
				fmt.Println("Error:", err)
	
				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal("Code: 400"))
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
								"start_index": 11,
								"end_index": 15,
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

				fmt.Println("## Result:", result)
	
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
				Config: VaultConfig{
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
				mockJSONResponse := `{"text": "Sample original text"}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/reidentify/string")
				defer ts.Close()

				fmt.Println("Mock server URL:", ts.URL)

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
				mockJSONResponse := `{"error":{"message":"Invalid request"}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

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
