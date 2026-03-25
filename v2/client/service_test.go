package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/skyflowapi/skyflow-go/v2/client"
	client2 "github.com/skyflowapi/skyflow-go/v2/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	. "github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

var (
	mockInsertSuccessJSON     = `{"vaultID":"id", "responses":[{"Body":{"records":[{"skyflow_id":"skyflowid", "tokens":{"name_on_card":"token1"}}]}, "Status":200}]}`
	mockInsertErrorJSON       = `{"vaultID":"id", "responses":[{"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}, {"Body":{"error":"Insert failed. Table name card_detail is invalid. Specify a valid table name."}, "Status":400}]}`
	mockDetokenizeSuccessJSON = `{"records":[{"token":"token", "valueType":"STRING", "value":"*REDACTED*", "error":null}]}`
	mockDetokenizeErrorJSON   = `{"error":{"grpc_code":5,"http_code":404,"X-Request-Id": "123455","message":"Detokenize failed. All tokens are invalid. Specify valid tokens.","http_status":"Not Found","details":[]}}`
	mockGetSuccessJSON        = `{"records":[{"fields":{"name":"name1", "skyflow_id":"id1"}, "tokens":null}]}`
	mockGetErrorJSON          = `{"error":{"grpc_code":5,"http_code":404,"message":"Get failed. [faild fail] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
	mockDeleteSuccessJSON     = `{"RecordIDResponse":["id1"]}`
	mockDeleteErrorJSON       = `{"error":{"grpc_code":5,"http_code":404,"message":"Delete failed. [id1] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
	mockQuerySuccessJSON      = `{"records":[{"fields":{"counter":null, "country":null, "date_of_birth":"XXXX-06-06", "email":"s******y@gmail.com", "name":"m***me", "phone_number":"XXXXXX8889", "skyflow_id":"id"}, "tokens":null}]}`
	mockQueryErrorJSON        = `{"error":{"grpc_code":5,"http_code":404,"message":"Invalid request. Table name cards is invalid. Specify a valid table name.","http_status":"Not Found","details":[]}}`
	mockUpdateSuccessJSON     = `{"skyflow_id":"id","tokens":{"name":"token"}}`
	mockUpdateErrorJSON       = `{"error":{"grpc_code":3,"http_code":400,"message":"Invalid request. No fields were present. Specify valid fields and values.","http_status":"Bad Request","details":[]}}`
	mockTokenizeSuccessJSON   = `{"records":[{"token":"token1"}]}`
	mockTokenizeErrorJSON     = `{"error":{"grpc_code":3,"http_code":400,"message":"Tokenization failed. Column group group_name is invalid. Specify a valid column group.","http_status":"Bad Request","details":[]}}`
	mockDeidentifyTextJSON    = `{"processed_text": "My name is [NAME] and my email is [EMAIL]", "word_count": 8, "character_count": 45, "entities": [{"token": "token1", "value": "John Doe", "entity_type": "NAME", "entity_scores": {"score": 0.9}, "location": {"start_index": 11, "end_index": 19, "start_index_processed": 11, "end_index_processed": 17}}, {"token": "token2", "value": "john@example.com", "entity_type": "EMAIL_ADDRESS", "entity_scores": {"score": 0.95}, "location": {"start_index": 30, "end_index": 45, "start_index_processed": 30, "end_index_processed": 37}}]}`
	mockNoEntitiesFoundJSON   = `{"processed_text": "No entities found in this text", "word_count": 6, "character_count": 30, "entities": []}`
	mockReidentifyTextJSON    = `{"text": "My SSN is 123-45-6789 and my card is *REDACTED*."}`
	mockDetectErrorJSON       = `{"error":{"grpc_code":3,"http_code":400,"message":"Invalid request","http_status":"Bad Request","details":[]}}`
	mockInvalidRunIdJSON      = `{ "status": "UNKNOWN", "outputType": "UNKNOWN", "output": [], "message": "", "size": 0}`
	mockDetectRunResponse     = map[string]interface{}{
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
)

var _ = Describe("Vault controller Test cases", func() {
	Describe("Test Insert functions", func() {
		var (
			response map[string]interface{}
			client   *Skyflow
			ts       *httptest.Server
			err      *skyflowError.SkyflowError
		)
		BeforeEach(func() {
			customHeader := make(map[string]string)
			customHeader["x-custom-header"] = "custom-header-value"
			client, err = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "test-api-key",
				},
			}),
				WithCustomHeaders(customHeader))
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
				response = make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertSuccessJSON), &response)

				// Setup mock server
				ts = setupMockServer(response, "ok", "/vaults/v1/vaults/")
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
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
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInsertErrorJSON), &response)

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
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
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
					ApiKey: "test-api-key",
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
				_ = json.Unmarshal([]byte(mockDetokenizeSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
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
				Expect(res.DetokenizedFields[0].Token).To(Equal("token"))
				Expect(res.DetokenizedFields[0].Value).To(Equal("*REDACTED*"))
				Expect(res.DetokenizedFields[0].Type).To(Equal("STRING"))

			})
			It("should return detokenized data with errors", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetokenizeErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				// Call the Detokenize function
				ctx := context.Background()
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				res, err := service.Detokenize(ctx, request, options) // Validate the response
				if err != nil {
					Expect(err).ToNot(BeNil())
					Expect(err.GetRequestId()).To(Equal("123456"))
				}
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
					ApiKey: "test-api-key",
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
				_ = json.Unmarshal([]byte(mockGetSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				var service, err1 = client.Vault()
				Expect(err1).To(BeNil())
				ctx := context.TODO()
				res, err := service.Get(ctx, request, options)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})
			It("should return error response when invalid ids passed in Get", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockGetErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Get(ctx, request, options)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
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
					ApiKey: "test-api-key",
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
				_ = json.Unmarshal([]byte(mockDeleteSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Delete(ctx, request)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid ids passed in Delete", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeleteErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Delete(ctx, request)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})
	Context("Test Upload file function", func() {
		var client *Skyflow
		var ctx context.Context
		var request FileUploadRequest
		BeforeEach(func() {
			// Initialize the VaultController instance
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "id",
				ClusterId: "cid",
				Env:       0,
				Credentials: Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}))
			request = FileUploadRequest{
				Table:      "table",
				SkyflowId:  "id1",
				FilePath:   os.Getenv("CRED_FILE_PATH"),
				ColumnName: "fileColumn",
			}
		})

		It("should return success response when valid id passed in Uploadfile", func() {
			response := make(map[string]interface{})
			mockJSONResponse := `{"SkyflowId":"id1"}`
			_ = json.Unmarshal([]byte(mockJSONResponse), &response)
			// Set the mock server URL in the controller's client
			ts := setupMockServer(response, "ok", "/vaults/v2/vaults/")

			// Set the mock server URL in the controller's client
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
				client := client2.NewClient(
					option.WithBaseURL(ts.URL+"/vaults"),
					option.WithToken("token"),
					option.WithHTTPHeader(header),
				)
				v.ApiClient = *client
				return nil
			}
			var service, err1 = client.Vault()
			Expect(err1).To(BeNil())
			ctx = context.TODO()
			res, err := service.UploadFile(ctx, request)
			Expect(err).To(BeNil())
			Expect(res).ToNot(BeNil())
		})
		It("should return error response when invalid ids passed in Get", func() {
			response := make(map[string]interface{})
			mockJSONResponse := `{"error":{"grpc_code":5,"http_code":404,"message":"Get failed. [faild fail] isn't a valid Skyflow ID. Specify a valid Skyflow ID.","http_status":"Not Found","details":[]}}`
			_ = json.Unmarshal([]byte(mockJSONResponse), &response)
			// Set the mock server URL in the controller's client
			ts := setupMockServer(response, "err", "/vaults/v2/vaults/")

			// Set the mock server URL in the controller's client
			header := http.Header{}
			header.Set("Content-Type", "application/json")
			CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
				client := client2.NewClient(
					option.WithBaseURL(ts.URL+"/vaults"),
					option.WithToken("token"),
					option.WithHTTPHeader(header),
				)
				v.ApiClient = *client
				return nil
			}
			service, _ := client.Vault()
			res, err := service.UploadFile(ctx, request)
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
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
					ApiKey: "test-api-key",
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
				_ = json.Unmarshal([]byte(mockQuerySuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Query(ctx, request)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid ids passed in Query", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockQueryErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
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
					ApiKey: "test-api-key",
				},
			}))
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
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
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
				_ = json.Unmarshal([]byte(mockUpdateErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")
				request.Tokens = map[string]interface{}{"name": "token"}
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Update(ctx, request, UpdateOptions{ReturnTokens: false, TokenMode: ENABLE})
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
					ApiKey: "test-api-key",
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
				_ = json.Unmarshal([]byte(mockTokenizeSuccessJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "ok", "/vaults/v1/vaults/")

				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Tokenize(ctx, arrReq)
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
			})

			It("should return error response when invalid data passed in Tokenize", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockTokenizeErrorJSON), &response)
				// Set the mock server URL in the controller's client
				ts := setupMockServer(response, "err", "/vaults/v1/vaults/")
				// Set the mock server URL in the controller's client
				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateRequestClientFunc = func(v *VaultController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL+"/vaults"),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					v.ApiClient = *client
					return nil
				}
				service, _ := client.Vault()
				res, err := service.Tokenize(ctx, arrReq)
				Expect(res).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})
	})
})

var _ = Describe("Detect controller Test cases", func() {
	var (
		client *Skyflow
		ctx    context.Context
	)

	BeforeEach(func() {
		client, _ = NewSkyflow(WithVaults(VaultConfig{
			VaultId:   "vault123",
			ClusterId: "cluster123",
			Env:       0,
			Credentials: Credentials{
				ApiKey: "test-api-key",
			},
		}))
		ctx = context.Background()
	})

	Describe("DeidentifyText tests", func() {
		Context("when request is valid", func() {
			It("should successfully deidentify text with all entity types", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDeidentifyTextJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/deidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				request := DeidentifyTextRequest{
					Text:     "My name is John Doe and my email is john@example.com",
					Entities: []DetectEntities{Name, EmailAddress},
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				service, _ := client.Detect()
				res, err := service.DeidentifyText(ctx, request)

				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.ProcessedText).To(Equal("My name is [NAME] and my email is [EMAIL]"))
				Expect(res.WordCount).To(Equal(int(8)))
				Expect(res.CharCount).To(Equal(int(45)))
				Expect(res.Entities).To(HaveLen(2))
				Expect(res.Entities[0].Entity).To(Equal("NAME"))
				Expect(res.Entities[1].Entity).To(Equal("EMAIL_ADDRESS"))
			})

			It("should handle empty entities array in response", func() {
				response := make(map[string]interface{})

				_ = json.Unmarshal([]byte(mockNoEntitiesFoundJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/deidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				request := DeidentifyTextRequest{
					Text:     "Simple text without sensitive information",
					Entities: []DetectEntities{Name, EmailAddress},
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				service, _ := client.Detect()
				res, err := service.DeidentifyText(ctx, request)

				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.ProcessedText).To(Equal("No entities found in this text"))
				Expect(res.Entities).To(BeEmpty())
			})
		})
	})

	Describe("ReidentifyText tests", func() {
		Context("when request is valid", func() {
			It("should successfully reidentify text", func() {
				response := make(map[string]interface{})

				_ = json.Unmarshal([]byte(mockReidentifyTextJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/reidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				request := ReidentifyTextRequest{
					Text: "My SSN is [SSN_123] and my card is [CREDIT_CARD_4321].",
					PlainTextEntities: []DetectEntities{
						Ssn,
					},
					RedactedEntities: []DetectEntities{
						CreditCard,
					},
				}
				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				service, _ := client.Detect()
				res, err := service.ReidentifyText(ctx, request)

				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				Expect(res.ProcessedText).To(Equal("My SSN is 123-45-6789 and my card is *REDACTED*."))
			})
		})
		Context("when API request fails", func() {
			It("should return error for API failure", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockDetectErrorJSON), &response)

				ts := setupMockServer(response, "err", "/v1/detect/reidentify/string")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
						option.WithBaseURL(ts.URL),
						option.WithToken("token"),
						option.WithHTTPHeader(header),
					)
					d.TextApiClient = *client.Strings
					return nil
				}

				request := ReidentifyTextRequest{
					Text: "Some text with <TOKEN>invalid-token</TOKEN>",
				}

				service, _ := client.Detect()
				res, err := service.ReidentifyText(ctx, request)

				Expect(err).ToNot(BeNil())
				Expect(res).To(BeNil())
			})
		})
	})

	Describe("DeidentifyFile tests", Ordered, func() {
		var (
			ctx       context.Context
			tempDir   string
			testFiles map[string]*os.File
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
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "vault123",
				ClusterId: "cluster123",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "test-api-key",
				},
			}))
			ctx = context.Background()
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
							Entities: []DetectEntities{Name, EmailAddress, Ssn, Date, Day, Dob},
							WaitTime: 5,
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
							client := client2.NewClient(
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

						service, _ := client.Detect()
						result, err := service.DeidentifyFile(ctx, tc.mockRequest)

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

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				service, _ := client.Detect()
				result, err := service.DeidentifyFile(ctx, request)

				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
				Expect(result).To(BeNil())
			})

			It("should return error when API request fails", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error":{"message":"Invalid file format"}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(response)
				}))
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
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

				service, _ := client.Detect()
				result, err := service.DeidentifyFile(ctx, request)

				Expect(err).ToNot(BeNil())
				Expect(result).To(BeNil())
			})

		})
	})

	Describe("GetDetectRun tests", func() {
		var (
			ctx context.Context
		)

		BeforeEach(func() {
			client, _ = NewSkyflow(WithVaults(VaultConfig{
				VaultId:   "vault123",
				ClusterId: "cluster123",
				Env:       0,
				Credentials: Credentials{
					ApiKey: "test-api-key",
				},
			}))
			ctx = context.Background()
		})

		Context("Success cases", func() {
			It("should successfully get completed run status", func() {
				ts := setupMockServer(mockDetectRunResponse, "ok", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
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

				service, _ := client.Detect()
				result, err := service.GetDetectRun(ctx, request)

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

		})

		Context("Error cases", func() {
			It("should return error for empty run ID", func() {
				request := GetDetectRunRequest{
					RunId: "",
				}

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				service, _ := client.Detect()
				result, err := service.GetDetectRun(ctx, request)

				Expect(result).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(fmt.Sprintf("Code: %v", skyflowError.INVALID_INPUT_CODE)))
			})

			It("should return error for expired run ID", func() {
				response := make(map[string]interface{})
				_ = json.Unmarshal([]byte(mockInvalidRunIdJSON), &response)

				ts := setupMockServer(response, "ok", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
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

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				service, _ := client.Detect()
				result, err := service.GetDetectRun(ctx, request)

				Expect(err).To(BeNil())
				Expect(result).ToNot(BeNil())
				Expect(result.Status).To(Equal("UNKNOWN"))
				Expect(result.RunId).To(Equal("invalid-run-id"))
				Expect(result.Type).To(Equal("UNKNOWN"))
			})

			It("should handle API error response", func() {
				response := make(map[string]interface{})
				mockJSONResponse := `{"error": {"message": "Invalid run ID"}}`
				_ = json.Unmarshal([]byte(mockJSONResponse), &response)

				ts := setupMockServer(response, "error", "/v1/detect/runs/")
				defer ts.Close()

				header := http.Header{}
				header.Set("Content-Type", "application/json")
				CreateDetectRequestClientFunc = func(d *DetectController) *skyflowError.SkyflowError {
					client := client2.NewClient(
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

				SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
					return nil
				}

				service, _ := client.Detect()
				result, err := service.GetDetectRun(ctx, request)

				Expect(result).To(BeNil())
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
		//mockResponse map[string]interface{}
	)

	BeforeEach(func() {
		//mockToken = "mock-valid-token"
		//mockResponse = map[string]interface{}{"key": "value"}
		mockRequest = InvokeConnectionRequest{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:   map[string]interface{}{"data": "test"},
			Method: POST,
		}
	})

	Describe("Invoke tests", func() {
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
				Expect(response.Data).To(Equal(fmt.Sprintf("%v", `{"key": "value"}`)))
			})
		})
		Context("Handling query parameters", func() {
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
		w.Header().Set("X-Request-Id", "123456")
		jsonData, _ := json.Marshal(mockResponse)
		// Write the response
		switch status {
		case "ok":
			w.WriteHeader(http.StatusOK)
		case "partial":
			w.WriteHeader(http.StatusMultiStatus)
		default:
			fmt.Println("status is", status)
			w.WriteHeader(http.StatusBadRequest)
		}
		//_, _ = w.Write(jsonData)
		w.Write(jsonData)

	})

	// Start the server and return it
	return httptest.NewServer(mockServer)
}
