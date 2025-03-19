package client_test

import (
	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/skyflowapi/skyflow-go/v2/client"
	vaultapi2 "github.com/skyflowapi/skyflow-go/v2/internal/generated/vaultapi"
	. "github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Vault controller Test cases", func() {
	Describe("Test Insert functions", func() {
		var (
			//mockJSONResponse string
			response map[string]interface{}
			client   *Skyflow
			ts       *httptest.Server
			err      *skyflowError.SkyflowError
		)
		BeforeEach(func() {
			client, err = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
				},
			}))
			response = make(map[string]interface{})
			ts = nil
		})

		AfterEach(func() {
			if ts != nil {
				ts.Close()
			}
		})

		Context("Insert with ContinueOnError True - Success Case", func() {
			BeforeEach(func() {
				mockJSONResponse := `{"vaultID":"id", "responses":[{"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name_on_card":"token1"}}]}, "Status":200}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				// Setup mock server
				ts = setupMockServer(response, "ok", "/vaults/v1/vaults/")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
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
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, insertError := service.Insert(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(insertError).To(BeNil())
				Expect(len(res.InsertedFields)).To(Equal(1))
				Expect(res.InsertedFields[0]["skyflow_id"]).To(Equal("skyflowid"))
			})
		})

		Context("Insert with ContinueOnError True - Error Case", func() {

			It("should return an error when insert fails and ContinueOnError is true", func() {
				mockJSONResponse := `{"vaultID":"id", "responses":[{"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}, {"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}]}`

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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}

				// Call the Insert method
				ctx := context.Background()
				var service, _ = client.Vault()
				res, insertError := service.Insert(ctx, request, options)
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

				// Call the Insert method
				ctx := context.Background()
				var service, _ = client.Vault()
				res, insertError := service.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil(), "Expected an error during insert operation")
				Expect(res).To(BeNil(), "Expected no response due to error in insert operation")
			})

		})

		Context("Insert with ContinueOnError True - Partial Error Case", func() {
			It("should return partial success and error fields", func() {
				mockJSONResponse := `{"vaultID":"id", "responses":[{"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}, {"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name":"token1"}}]}, "Status":200}]}`

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
					TokenMode:       DISABLE,
				}

				// Set up the mock server using the reusable function
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")
				defer ts.Close()
				// Set the mock server URL in the controller's client
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}

				// Call the Insert method
				ctx := context.Background()
				var service, _ = client.Vault()
				res, insertError := service.Insert(ctx, request, options)

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
				mockJSONResponse := `{"records":[{"skyflow_id":"skyflowid1", "tokens":{"name":"nameToken1"}}, {"skyflow_id":"skyflowid2", "tokens":{"expiry_month":"monthToken", "name":"nameToken2"}}]}`
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}

				// Call the Insert method
				ctx := context.Background()
				var service, _ = client.Vault()
				res, insertError := service.Insert(ctx, request, options)

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
				mockJSONResponse := `{"error":{"grpc_code":3,"http_code":400,"message":"Insert failed. Table name card_detail is invalid. Specify a valid table name.","http_status":"Bad Request","details":[]}}`
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
					ContinueOnError: false,
					Upsert:          "upsert",
					TokenMode:       DISABLE,
				}

				// Set up the mock server using the reusable function
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")
				defer ts.Close()

				// Set the mock server URL in the controller's client
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}

				// Create the VaultController instance

				// Call the Insert method
				ctx := context.Background()
				var service, _ = client.Vault()
				res, insertError := service.Insert(ctx, request, options)
				// Assertions
				Expect(insertError).ToNot(BeNil(), "Expected error during insert operation")
				Expect(res).To(BeNil(), "Expected no response")
			})
		})

		Context("Insert Client Creation Failed", func() {
			It("should return an error when client creation fails", func() {
				mockJSONResponse := `{"vaultID":"id", "responses":[{"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name_on_card":"token1"}}]}, "Status":200}]}`
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
				var service, _ = client.Vault()
				_, insertError := service.Insert(ctx, request, options)

				// Assertions
				Expect(insertError).ToNot(BeNil(), "Expected an error when client creation fails")
			})
		})
	})

	Describe("Test Detokenize functions", func() {
		var (
			request DetokenizeRequest
			options DetokenizeOptions
			client  *Skyflow
		)
		BeforeEach(func() {
			// Initialize the VaultController instance
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
				},
			}))

			// Initialize context, request, and options
			request = DetokenizeRequest{
				DetokenizeData: []DetokenizeData{
					{Token: "token1",
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
				mockJSONResponse := `{"records":[{"token":"token", "valueType":"STRING", "value":"*REDACTED*", "error":null}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				// Call the Detokenize function
				ctx := context.Background()
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Detokenize(ctx, request, options)
				// Validate the response
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.DetokenizedFields).To(HaveLen(1))
				Expect(res.DetokenizedFields[0]["Token"]).To(Equal("token"))
				Expect(res.DetokenizedFields[0]["Value"]).To(Equal("*REDACTED*"))
				Expect(res.DetokenizedFields[0]["ValueType"]).To(Equal("STRING"))
			})
			It("should return detokenized data with errors", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Detokenize failed. All tokens are invalid. Specify valid tokens.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				// Call the Detokenize function
				ctx := context.Background()
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Detokenize(ctx, request, options) // Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return detokenized data with errors", func() {
				request.DetokenizeData = nil
				// Call the Detokenize function
				ctx := context.Background()
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Detokenize(ctx, request, options) // Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return detokenized data with partial success response", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"records":[{"token":"token1", "valueType":"STRING", "value":"*REDACTED*", "error":null}, {"token":"token1", "valueType":"NONE", "value":"", "error":"Token Not Found"}]}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				// Call the Detokenize function
				ctx := context.Background()
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Detokenize(ctx, request, options) // Validate the response
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
			It("should return error while creating client in detokenize", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				// Call the Detokenize function
				ctx := context.Background()
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Detokenize(ctx, request, options) // Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
			It("should return error in get token while calling in detokenize", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				// Call the Detokenize function
				ctx := context.Background()
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Detokenize(ctx, request, options) // Validate the response
				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})
	})

	Describe("Test Get functions", func() {
		var client *Skyflow
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
				},
			}))
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

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Get(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
			It("should return error response when invalid ids passed in Get", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Get failed. [faild fail] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Get(ctx, request, options)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error client creation step Get", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				service, _ := client.Vault()
				res, err := service.Get(ctx, request, options)
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
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Get(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
		})
	})

	Describe("Test Delete functions", func() {
		var client *Skyflow
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
				},
			}))
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

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Delete(ctx, request)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid ids passed in Delete", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Delete failed. [id1] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Delete(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when invalid data passed in Delete", func() {
				request.Ids = []string{}
				service, _ := client.Vault()
				res, err := service.Delete(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Delete", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				service, _ := client.Vault()
				res, err := service.Delete(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Test Query functions", func() {
		var client *Skyflow
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
				},
			}))
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

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Query(ctx, request)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid ids passed in Query", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Invalid request. Table name cards is invalid. Specify a valid table name.","http_status":"Not Found","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Query(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when invalid data passed in Query", func() {
				request.Query = ""
				service, _ := client.Vault()
				res, err := service.Query(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Query", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				service, _ := client.Vault()
				res, err := service.Query(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Test Update functions", func() {
		var client *Skyflow
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
				},
			}))
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

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Update(ctx, request, UpdateOptions{
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
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")
				request.Tokens = map[string]interface{}{"name": "token"}
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Update(ctx, request, UpdateOptions{ReturnTokens: false, TokenMode: ENABLE})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when validation fail for invalid data passed in Update", func() {
				request.Tokens = nil
				service, _ := client.Vault()
				res, err := service.Update(ctx, request, UpdateOptions{ReturnTokens: false, TokenMode: ENABLE})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Update", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				service, _ := client.Vault()
				res, err := service.Update(ctx, request, UpdateOptions{ReturnTokens: true, TokenMode: ENABLE_STRICT})
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("Test Tokenize functions", func() {
		var client *Skyflow
		var ctx context.Context
		BeforeEach(func() {
			// Initialize the VaultController instance
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
				},
			}))
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

				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Tokenize(ctx, arrReq)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid data passed in Tokenize", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"grpc_code":3,"http_code":400,"message":"Tokenization failed. Column group group_name is invalid. Specify a valid column group.","http_status":"Bad Request","details":[]}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					configuration := vaultapi2.NewConfiguration()
					configuration.AddDefaultHeader("Authorization", "Bearer token")
					configuration.AddDefaultHeader("Content-Type", "application/json")
					configuration.Servers[0].URL = ts.URL + "/vaults"
					apiClient := vaultapi2.NewAPIClient(configuration)
					v.ApiClient = *apiClient
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Tokenize(ctx, arrReq)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
			It("should return error response when validations failed for invalid data passedin Tokenize", func() {
				arrReq = append(arrReq, TokenizeRequest{})
				service, _ := client.Vault()
				res, err := service.Tokenize(ctx, arrReq)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})

			It("should return error client creation step Tokenize", func() {
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "error occurred in client fucntion")
				}
				service, _ := client.Vault()
				res, err := service.Tokenize(ctx, arrReq)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})
})

var _ = Describe("ConnectionController", func() {
	var (
		client     *Skyflow
		mockServer *httptest.Server
		//mockToken    string
		mockRequest  InvokeConnectionRequest
		mockResponse map[string]interface{}
	)

	BeforeEach(func() {
		//mockToken = "mock-valid-token"
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
				client, _ = NewSkyflow(WithConnections(ConnectionConfig{
					ConnectionId:  "failed",
					ConnectionUrl: mockServer.URL,
					Credentials: Credentials{
						Token: "TOKEN",
					},
				}))
			})

			AfterEach(func() {
				mockServer.Close()
			})

			It("should return a valid response", func() {
				SetBearerTokenForConnectionControllerFunc = func(v *ConnectionController) *skyflowError.SkyflowError {
					return nil
				}

				service, err := client.Connection("failed")
				response, err := service.Invoke(ctx, mockRequest)
				Expect(err).To(BeNil())
				Expect(response.Data).To(Equal(mockResponse))
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
				service, _ := client.Connection("failed")

				response, err := service.Invoke(ctx, request)
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
				service, err := client.Connection("failed")
				response, err := service.Invoke(ctx, request)
				Expect(err).ToNot(BeNil())
				Expect(response).To(BeNil())
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
			w.WriteHeader(http.StatusMultiStatus)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
		//_, _ = w.Write(jsonData)
		w.Write(jsonData)

	})

	// Start the server and return it
	return httptest.NewServer(mockServer)
}
