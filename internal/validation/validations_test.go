package validation_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	errors "github.com/skyflowapi/skyflow-go/v2/utils/error"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServiceAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validations Suite")
}

var _ = Describe("ValidateTokensForInsertRequest", func() {
	Context("with empty tokens array", func() {
		It("should return an error", func() {
			err := ValidateTokensForInsertRequest([]map[string]interface{}{}, []map[string]interface{}{}, common.ENABLE_STRICT)
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.NO_TOKENS_WITH_BYOT, common.ENABLE_STRICT)))
		})
	})

	Context("with mismatched token and values sizes in ENABLE_STRICT mode", func() {
		It("should return an error", func() {
			tokens := []map[string]interface{}{
				{"key1": "token1", "key2": "token2"},
			}
			values := []map[string]interface{}{
				{"key1": "value1"},
			}
			err := ValidateTokensForInsertRequest(tokens, values, common.ENABLE_STRICT)
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)))
		})
	})

	Context("with missing keys in values map", func() {
		It("should return an error", func() {
			tokens := []map[string]interface{}{
				{"key1": "token1", "key2": "token2"},
			}
			values := []map[string]interface{}{
				{"key1": "value1"},
			}
			err := ValidateTokensForInsertRequest(tokens, values, common.ENABLE)
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(errors.MISMATCH_OF_FIELDS_AND_TOKENS))
		})
	})

	Context("with nil tokens or values map", func() {
		It("should return an error", func() {
			tokens := []map[string]interface{}{
				make(map[string]interface{}),
			}
			values := []map[string]interface{}{
				{"key1": "value1", "key2": nil},
			}
			err := ValidateTokensForInsertRequest(tokens, values, common.ENABLE)
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKENS))
		})
		It("should return an error", func() {
			tokens := []map[string]interface{}{
				{"key1": "token1", "key2": "token2"},
			}
			values := []map[string]interface{}{
				{"key1": "value1", "key2": nil},
			}
			err := ValidateTokensForInsertRequest(tokens, values, common.ENABLE)
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(errors.MISMATCH_OF_FIELDS_AND_TOKENS))
		})
	})

	Context("with valid tokens and values", func() {
		It("should not return an error", func() {
			tokens := []map[string]interface{}{
				{"key1": "token1", "key2": "token2"},
			}
			values := []map[string]interface{}{
				{"key1": "value1", "key2": "value2"},
			}
			err := ValidateTokensForInsertRequest(tokens, values, common.ENABLE_STRICT)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should not return an error when empty token passed", func() {
			tokens := []map[string]interface{}{
				{"key1": "token1", "key2": nil},
			}
			values := []map[string]interface{}{
				{"key1": "value1", "key2": "value2"},
			}
			err := ValidateTokensForInsertRequest(tokens, values, common.ENABLE)
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUE_IN_TOKENS))
		})
	})
	Context("ValidateInsertRequest", func() {
		It("should return TABLE_KEY_ERROR when table is empty", func() {
			request := common.InsertRequest{
				Table:  "",
				Values: []map[string]interface{}{{"key": "value"}},
			}
			options := common.InsertOptions{}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.TABLE_KEY_ERROR))
		})

		It("should return EMPTY_VALUES when values are nil", func() {
			request := common.InsertRequest{
				Table:  "testTable",
				Values: nil,
			}
			options := common.InsertOptions{}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUES))
		})

		It("should return EMPTY_VALUES when values are empty", func() {
			request := common.InsertRequest{
				Table:  "testTable",
				Values: []map[string]interface{}{},
			}
			options := common.InsertOptions{}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUES))
		})

		It("should return HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT when homogeneous is true with upsert", func() {
			request := common.InsertRequest{
				Table:  "testTable",
				Values: []map[string]interface{}{{"key": "value"}},
			}
			options := common.InsertOptions{
				Upsert:      "upsertValue",
				Homogeneous: true,
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT))
		})

		It("should return EMPTY_VALUE_IN_VALUES when a value is nil or empty", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": nil},
				},
			}
			options := common.InsertOptions{}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUE_IN_VALUES))
		})

		It("should return EMPTY_KEY_IN_VALUES when a key is empty", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"": "value"},
				},
			}
			options := common.InsertOptions{}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_KEY_IN_VALUES))
		})

		It("should return TOKENS_PASSED_FOR_BYOT_DISABLE when tokens are passed in BYOT DISABLE mode", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.DISABLE,
				Tokens:    []map[string]interface{}{{"key": "token"}},
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.TOKENS_PASSED_FOR_BYOT_DISABLE))
		})
		It("should return error when tokens are passed in BYOT ENABLE mode", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.ENABLE,
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKENS))
		})
		It("should not return error when tokens are not passed for all values object in BYOT ENABLE mode", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
					{"key2": "value2"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.ENABLE,
				Tokens:    []map[string]interface{}{{"key": "token"}},
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).To(BeNil())
		})
		It("should return error when tokens are not passed for all values object in BYOT ENABLE mode", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
					{"key2": "value2", "key3": "value3"},
					{"key4": "value4", "key5": "value5"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.ENABLE,
				Tokens:    []map[string]interface{}{{"key": "token"}, {"key2": "value2"}, make(map[string]interface{})},
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKENS))
		})
		It("should return error when tokens and values are not passed with equal length in BYOT ENABLE_strict mode", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.ENABLE_STRICT,
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKENS))
		})
		It("should return error when tokens and values are not passed with equal length in BYOT ENABLE_STRICT mode", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
					{"key2": "value2", "key3": "value3"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.ENABLE_STRICT,
				Tokens:    []map[string]interface{}{{"key": "token"}},
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT))
		})
		It("should return error when tokens are invalid in BYOT ENABLE STRICT mode", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
					{"key2": "value2", "key3": "value3"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.ENABLE_STRICT,
				Tokens:    []map[string]interface{}{{"key": nil}, {"key2": nil, "key3": nil}},
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUE_IN_TOKENS))
		})
		It("Valid request", func() {
			request := common.InsertRequest{
				Table: "testTable",
				Values: []map[string]interface{}{
					{"key": "value"},
				},
			}
			options := common.InsertOptions{
				TokenMode: common.ENABLE_STRICT,
				Tokens:    []map[string]interface{}{{"key": "nil"}},
			}

			err := ValidateInsertRequest(request, options)
			Expect(err).To(BeNil())
		})
	})

	Describe("Validate Config", func() {
		var (
			validCredentials common.Credentials
		)

		BeforeEach(func() {
			validCredentials = common.Credentials{
				Path:              "valid/path",
				Roles:             []string{"role1", "role2"},
				Context:           "valid-context",
				CredentialsString: "valid-credentials",
				Token:             "valid-token",
				ApiKey:            "valid-apikey",
			}

		})

		Context("Invalid VaultConfig", func() {
			It("should return error for empty VaultId", func() {
				config := common.VaultConfig{
					VaultId:     "",
					ClusterId:   "valid-cluster-id",
					Env:         common.PROD,
					Credentials: validCredentials,
				}
				err := ValidateVaultConfig(config)
				Expect(err).ToNot(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_VAULT_ID))
			})

			It("should return error for empty ClusterId", func() {
				config := common.VaultConfig{
					VaultId:     "valid-vault-id",
					ClusterId:   "",
					Env:         common.PROD,
					Credentials: validCredentials,
				}
				err := ValidateVaultConfig(config)
				Expect(err).ToNot(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_CLUSTER_ID))
			})
		})

		Context("Valid VaultConfig", func() {
			It("should return nil for valid VaultConfig", func() {
				config := common.VaultConfig{
					VaultId:   "valid-vault-id",
					ClusterId: "valid-cluster-id",
					Env:       common.PROD,
					Credentials: common.Credentials{
						Path: "valid-path",
					},
				}
				err := ValidateVaultConfig(config)
				Expect(err).To(BeNil())
			})
		})
		Describe("ValidateConnectionConfig", func() {
			Context("Invalid common.ConnectionConfig", func() {
				It("should return an error if ConnectionId is empty", func() {
					config := common.ConnectionConfig{
						ConnectionId:  "",
						ConnectionUrl: "https://valid.url",
						Credentials:   common.Credentials{},
					}

					err := ValidateConnectionConfig(config)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
					Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_CONNECTION_ID))
				})

				It("should return an error if ConnectionUrl is empty", func() {
					config := common.ConnectionConfig{
						ConnectionId:  "valid-id",
						ConnectionUrl: "",
						Credentials:   common.Credentials{},
					}
					err := ValidateConnectionConfig(config)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
					Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_CONNECTION_URL))
				})

			})

			Context("Valid common.ConnectionConfig", func() {
				It("should return nil for a valid common.ConnectionConfig", func() {
					config := common.ConnectionConfig{
						ConnectionId:  "valid-id",
						ConnectionUrl: "https://valid.url",
						Credentials:   common.Credentials{},
					}

					err := ValidateConnectionConfig(config)
					Expect(err).To(BeNil())
				})
			})
		})
		Describe("ValidateCredentials", func() {
			Context("Invalid Credentials", func() {
				It("should return an error if no token generation means are passed", func() {
					credentials := common.Credentials{}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
					Expect(err.GetMessage()).To(ContainSubstring(errors.NO_TOKEN_GENERATION_MEANS_PASSED))
				})

				It("should return an error if multiple token generation means are passed", func() {
					credentials := common.Credentials{
						Path:              "some/path",
						CredentialsString: "some-string",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
					Expect(err.GetMessage()).To(ContainSubstring(errors.MULTIPLE_TOKEN_GENERATION_MEANS_PASSED))
				})

				It("should return an error for invalid API key format", func() {
					credentials := common.Credentials{
						ApiKey: "invalid-api-key",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
					Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_API_KEY))
				})

				It("should return an error if roles list is empty", func() {
					credentials := common.Credentials{
						Token: "token",
						Roles: []string{},
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
					Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_ROLES))
				})

				It("should return an error if a role in roles is empty", func() {
					var role string
					credentials := common.Credentials{
						Token: "token",
						Roles: []string{"admin", role},
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
					Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_ROLE_IN_ROLES))
				})
			})

			Context("Valid Credentials", func() {
				It("should return nil for valid Path", func() {
					credentials := common.Credentials{
						Path: "some/path",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(BeNil())
				})

				It("should return nil for valid CredentialsString", func() {
					credentials := common.Credentials{
						CredentialsString: "valid-credentials-string",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(BeNil())
				})

				It("should return nil for valid Token", func() {
					credentials := common.Credentials{
						Token: "valid-token",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(BeNil())
				})

				It("should return nil for valid API key", func() {
					credentials := common.Credentials{
						ApiKey: "sky-abcde-1234567890abcdef1234567890abcdef",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(BeNil())
				})

				It("should return nil for valid roles list", func() {
					credentials := common.Credentials{
						Token: "token",
						Roles: []string{"admin", "user"},
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(BeNil())
				})
			})
		})
	})
	Context("when validating headers", func() {
		It("should return an error for empty headers", func() {
			request := common.InvokeConnectionRequest{
				Headers: make(map[string]string),
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_REQUEST_HEADER))
		})

		It("should return an error for invalid header key or value", func() {
			request := common.InvokeConnectionRequest{
				Headers: map[string]string{"": "value", "key": ""},
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_REQUEST_HEADERS))
		})
	})

	Context("when validating invoke connection path parameters", func() {
		It("should return an error for empty path parameters", func() {
			request := common.InvokeConnectionRequest{
				PathParams: make(map[string]string),
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_PARAMETERS))
		})

		It("should return an error for invalid path parameter key", func() {
			request := common.InvokeConnectionRequest{
				PathParams: map[string]string{"": "value"},
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_PARAMETER_NAME))
		})

		It("should return an error for invalid path parameter value", func() {
			request := common.InvokeConnectionRequest{
				PathParams: map[string]string{"key": ""},
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_PARAMETER_VALUE))
		})
	})

	Context("when validating connection query parameters", func() {
		It("should return an error for empty query parameters", func() {
			request := common.InvokeConnectionRequest{
				QueryParams: make(map[string]interface{}),
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_QUERY_PARAM))
		})

		It("should return an error for invalid query parameter key or value", func() {
			request := common.InvokeConnectionRequest{
				QueryParams: map[string]interface{}{"": "value", "key": nil},
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_QUERY_PARAM))
		})
	})

	Context("when validating connection request body", func() {
		It("should return an error for empty request body", func() {
			request := common.InvokeConnectionRequest{
				Body: make(map[string]interface{}),
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_REQUEST_BODY))
		})
	})
	Context("when validating connection request should return no error when valid request", func() {
		It("should return no error for valid request", func() {
			request := common.InvokeConnectionRequest{
				Body: map[string]interface{}{
					"key": "value",
				},
				Method: common.POST,
			}
			err := ValidateInvokeConnectionRequest(request)
			Expect(err).To(BeNil())
		})
	})

	Context("when validating tokens", func() {
		It("should return an error if tokens are nil", func() {
			request := common.DetokenizeRequest{
				DetokenizeData: nil,
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_DETOKENIZE_DATA))
		})

		It("should return an error if tokens are empty", func() {
			request := common.DetokenizeRequest{
				DetokenizeData: []common.DetokenizeData{},
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKENS_DETOKENIZE))
		})

		It("should return an error if any token is empty", func() {
			request := common.DetokenizeRequest{
				DetokenizeData: []common.DetokenizeData{
					{Token: ""},
				},
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKEN_IN_DETOKENIZE_DATA))
		})

		It("should not return an error if all tokens are valid", func() {
			request := common.DetokenizeRequest{
				DetokenizeData: []common.DetokenizeData{
					{Token: "token1"},
					{Token: "token2"},
				},
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).To(BeNil())
		})
	})
	Context("when validating update requests", func() {
		var validData = map[string]interface{}{"skyflow_id": "123", "key": "value", "key2": "value2"}
		It("should return an error if the table is empty", func() {
			request := common.UpdateRequest{
				Table:  "",
				Data:   validData,
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TABLE))
		})

		It("should return an error if the id is empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   map[string]interface{}{"skyflow_id": "", "key": "value"},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_ID_IN_UPDATE))
		})

		It("should return an error if the data are nil or empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   nil,
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_ID_IN_UPDATE))
		})

		It("should return an error if tokens are nil or empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: nil,
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).To(BeNil())
		})

		It("should return an error if a data is empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   map[string]interface{}{"skyflow_id": "123", "key": ""},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_DATA_IN_DATA_KEY))
		})

		It("should return an error if a key is empty in data", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   map[string]interface{}{"skyflow_id": "123", "": "value"},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_KEY_IN_DATA))
		})

		It("should return an error if tokens are passed with TokenMode DISABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.DISABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.TOKENS_PASSED_FOR_BYOT_DISABLE))
		})

		It("should return an error if no tokens are passed with TokenMode ENABLE_STRICT", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: nil,
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE_STRICT}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.NO_TOKENS_WITH_BYOT, common.ENABLE_STRICT)))
		})

		It("should not return an error for valid input with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).To(BeNil())
		})

		It("should return an error for invalid input with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: make(map[string]interface{}),
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKENS))
		})

		It("should return an error for tokens empty with TokenMode ENABLE_STRICT", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: nil,
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE_STRICT}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.NO_TOKENS_WITH_BYOT, common.ENABLE_STRICT)))
		})
		It("should return an error for tokens empty with TokenMode ENABLE_STRICT", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE_STRICT}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT))
		})

		It("should return an error for tokens empty with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: nil,
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.NO_TOKENS_WITH_BYOT, common.ENABLE)))
		})
		It("should return an error for tokens key value is empty with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: map[string]interface{}{"key": nil},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUE_IN_TOKENS))
		})
		It("should return an error for tokens key is empty with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: map[string]interface{}{"": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_KEY_IN_TOKENS))
		})
		It("should return an error for tokens key not exist in values with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   validData,
				Tokens: map[string]interface{}{"demo": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.MISMATCH_OF_FIELDS_AND_TOKENS))
		})
		It("should return an error for tokens key not exist in values with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   map[string]interface{}{"skyflow_id": "123", "key": "value", "key2": nil},
				Tokens: map[string]interface{}{"key": "value", "key2": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.MISMATCH_OF_FIELDS_AND_TOKENS))
		})

		It("should return error if sufficient tokens is not passed for all values object in BYOT ENABLE STRICT mode", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   map[string]interface{}{"skyflow_id": "123", "key": "value", "key2": "value2"},
				Tokens: map[string]interface{}{"key2": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE_STRICT}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT))
		})

		It("should not return error if tokens and values count is not equal in BYOT ENABLE mode", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   map[string]interface{}{"skyflow_id": "123", "key": "value", "key2": "value2"},
				Tokens: map[string]interface{}{"key2": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).To(BeNil())
		})

		It("should return error if tokens and values count is not equal in BYOT ENABLE STRICT mode", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Data:   map[string]interface{}{"skyflow_id": "123", "key": "value", "key2": "value2"},
				Tokens: map[string]interface{}{"key2": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE_STRICT}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT))
		})
	})
	Context("ValidateTokenizeRequest", func() {
		var (
			request []common.TokenizeRequest
		)
		It("should return INVALID_TOKENIZE_REQUEST error", func() {
			request = nil
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_TOKENIZE_REQUEST))
		})
		It("should return INVALID_TOKENIZE_REQUEST error", func() {
			request = []common.TokenizeRequest{}
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_TOKENIZE_REQUEST))
		})
		It("should return EMPTY_VALUE_IN_COLUMN_VALUES error", func() {
			request = []common.TokenizeRequest{
				{ColumnGroup: "", Value: "valid_value"},
			}
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUE_IN_COLUMN_VALUES))
		})
		It("should return EMPTY_COLUMN_VALUES error", func() {
			request = []common.TokenizeRequest{
				{ColumnGroup: "valid_group", Value: ""},
			}
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_COLUMN_VALUES))
		})
		It("should return nil", func() {
			request = []common.TokenizeRequest{
				{ColumnGroup: "valid_group", Value: "valid_value"},
			}
			err := ValidateTokenizeRequest(request)
			Expect(err).To(BeNil())
		})
	})
	Context("When QueryRequest is validated", func() {

		It("should return an error if the query is empty", func() {
			// Arrange
			request := common.QueryRequest{
				Query: "",
			}

			// Act
			err := ValidateQueryRequest(request)

			// Assert
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_QUERY))
		})

		It("should return nil if the query is valid", func() {
			// Arrange
			request := common.QueryRequest{
				Query: "SELECT * FROM users",
			}

			// Act
			err := ValidateQueryRequest(request)

			// Assert
			Expect(err).To(BeNil())
		})
	})
	Context("ValidateDeleteRequest", func() {
		It("should return an EMPTY_TABLE error", func() {
			request := common.DeleteRequest{
				Table: "",
				Ids:   []string{"id1", "id2"},
			}

			err := ValidateDeleteRequest(request)

			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TABLE))
		})
		It("should return an EMPTY_IDS error", func() {
			request := common.DeleteRequest{
				Table: "test_table",
				Ids:   nil,
			}

			err := ValidateDeleteRequest(request)

			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_IDS))
		})
		It("should return an EMPTY_ID_IN_IDS error", func() {
			request := common.DeleteRequest{
				Table: "test_table",
				Ids:   []string{"id1", ""},
			}

			err := ValidateDeleteRequest(request)

			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_ID_IN_IDS))
		})
		It("should return nil", func() {
			request := common.DeleteRequest{
				Table: "test_table",
				Ids:   []string{"id1", "id2"},
			}

			err := ValidateDeleteRequest(request)

			Expect(err).To(BeNil())
		})
	})
	Context("Validate the deidentify text request", func() {
		It("should return error when Text is empty", func() {
			req := common.DeidentifyTextRequest{
				Text: "",
			}
			err := ValidateDeidentifyTextRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.INVALID_TEXT_IN_DEIDENTIFY)))
		})

		It("should return error when Text is only whitespace", func() {
			req := common.DeidentifyTextRequest{
				Text: "   ",
			}
			err := ValidateDeidentifyTextRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.INVALID_TEXT_IN_DEIDENTIFY)))
		})

		It("should return nil when Text is non-empty", func() {
			req := common.DeidentifyTextRequest{
				Text: "valid text",
			}
			err := ValidateDeidentifyTextRequest(req)
			Expect(err).To(BeNil())
		})

		Context("when given an invalid entity", func() {
			It("should return an error", func() {
				req := common.DeidentifyTextRequest{
					Text:     "Sensitive text",
					Entities: []common.DetectEntities{"invalid_entity"},
				}
				err := ValidateDeidentifyTextRequest(req)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			})
		})

		Context("when given an invalid token format", func() {
			It("should return an error", func() {
				req := common.DeidentifyTextRequest{
					Text: "Sensitive text",
					TokenFormat: common.TokenFormat{
						DefaultType: "invalid_token_type",
					},
				}
				err := ValidateDeidentifyTextRequest(req)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			})
		})

		Context("when TokenFormat.EntityOnly contains invalid entity", func() {
			It("should return an error", func() {
				req := common.DeidentifyTextRequest{
					Text: "Sensitive text",
					TokenFormat: common.TokenFormat{
						EntityOnly: []common.DetectEntities{"invalid_entity"},
					},
				}
				err := ValidateDeidentifyTextRequest(req)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			})
		})

		Context("when TokenFormat.VaultToken contains invalid entity", func() {
			It("should return an error", func() {
				req := common.DeidentifyTextRequest{
					Text: "Sensitive text",
					TokenFormat: common.TokenFormat{
						VaultToken: []common.DetectEntities{"invalid_entity"},
					},
				}
				err := ValidateDeidentifyTextRequest(req)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			})
		})

		Context("when ShiftDates.Entities is empty", func() {
			It("should return an error", func() {
				req := common.DeidentifyTextRequest{
					Text: "Sensitive text",
					Transformations: common.Transformations{
						ShiftDates: common.DateTransformation{
							MaxDays:  5,
							MinDays:  2,
							Entities: []common.TransformationsShiftDatesEntityTypesItem{},
						},
					},
				}
				err := ValidateDeidentifyTextRequest(req)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
				Expect(err.GetMessage()).To(ContainSubstring(errors.DETECT_ENTITIES_REQUIRED_ON_SHIFT_DATES))
			})
		})

		Context("when ShiftDates.MaxDays and MinDays are zero", func() {
			It("should throw an error", func() {
				req := common.DeidentifyTextRequest{
					Text: "Sensitive text",
					Transformations: common.Transformations{
						ShiftDates: common.DateTransformation{
							MaxDays: 0,
							MinDays: 0,
							Entities: []common.TransformationsShiftDatesEntityTypesItem{
								common.TransformationsShiftDatesEntityTypesItemDate,
							},
						},
					},
				}
				err := ValidateDeidentifyTextRequest(req)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
				Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_SHIFT_DATES))
			})
		})

		Context("when ShiftDates.MinDays is greater than MaxDays", func() {
			It("should throw an error", func() {
				req := common.DeidentifyTextRequest{
					Text: "Sensitive text",
					Transformations: common.Transformations{
						ShiftDates: common.DateTransformation{
							MaxDays: 5,
							MinDays: 7,
							Entities: []common.TransformationsShiftDatesEntityTypesItem{
								common.TransformationsShiftDatesEntityTypesItemDate,
							},
						},
					},
				}
				err := ValidateDeidentifyTextRequest(req)
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
				Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_DATE_TRANSFORMATION_RANGE))
			})
		})

	})

	Context("Validate the reidentify text request", func() {
		It("should return error when Text is empty", func() {
			req := common.ReidentifyTextRequest{
				Text: "",
			}
			err := ValidateReidentifyTextRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.INVALID_TEXT_IN_REIDENTIFY)))
		})

		It("should return error when Text is only whitespace", func() {
			req := common.ReidentifyTextRequest{
				Text: "   ",
			}
			err := ValidateReidentifyTextRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.INVALID_TEXT_IN_REIDENTIFY)))
		})

		It("should return nil when Text is non-empty", func() {
			req := common.ReidentifyTextRequest{
				Text: "valid text",
			}
			err := ValidateReidentifyTextRequest(req)
			Expect(err).To(BeNil())
		})
	})

	Context("ValidateGetDetectRunRequest", func() {
		It("should return error when RunId is empty", func() {
			req := common.GetDetectRunRequest{
				RunId: "",
			}
			err := ValidateGetDetectRunRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_RUN_ID))
		})

		It("should return nil when RunId is valid", func() {
			req := common.GetDetectRunRequest{
				RunId: "valid-run-id",
			}
			err := ValidateGetDetectRunRequest(req)
			Expect(err).To(BeNil())
		})
	})

	Context("ValidateDeidentifyFileRequest", Ordered, func() {

		var (
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

			fileTypes := []string{"txt"}
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

		It("should return error when FileInput is empty", func() {
			req := common.DeidentifyFileRequest{}
			err := ValidateDeidentifyFileRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
		})

		It("should return error when both FilePath and File are empty", func() {
			req := common.DeidentifyFileRequest{
				File: common.FileInput{},
			}
			err := ValidateDeidentifyFileRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring(string(errors.INVALID_INPUT_CODE)))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_FILE_AND_FILE_PATH_IN_DEIDENTIFY_FILE))
		})

		It("should return nil when FilePath is valid", func() {
			testFilePath := filepath.Join(tempDir, "detect.txt")
			err := os.WriteFile(testFilePath, []byte("test content"), 0644)
			Expect(err).To(BeNil(), "Failed to create test file")

			req := common.DeidentifyFileRequest{
				File: common.FileInput{
					FilePath: testFilePath,
				},
			}
			validationErr := ValidateDeidentifyFileRequest(req)
			Expect(validationErr).To(BeNil())
		})

		It("should return nil when File is valid", func() {
			// First create and write to the test file
			testFilePath := filepath.Join(tempDir, "detect.txt")
			err := os.WriteFile(testFilePath, []byte("test content"), 0644)
			Expect(err).To(BeNil(), "Failed to create test file")

			// Now open the file
			file, err := os.Open(testFilePath)
			Expect(err).To(BeNil(), "Failed to open test file")
			defer file.Close()

			req := common.DeidentifyFileRequest{
				File: common.FileInput{
					File: file,
				},
			}
			validationErr := ValidateDeidentifyFileRequest(req)
			Expect(validationErr).To(BeNil())
		})

		It("should return error when both FilePath and File are provided", func() {
			file, err := os.Open(filepath.Join(tempDir, "detect.txt"))
			Expect(err).To(BeNil())
			defer file.Close()

			req := common.DeidentifyFileRequest{
				File: common.FileInput{
					FilePath: filepath.Join(tempDir, "detect.txt"),
					File:     file,
				},
			}
			validationErr := ValidateDeidentifyFileRequest(req)
			Expect(validationErr).ToNot(BeNil())
			Expect(validationErr.GetMessage()).To(ContainSubstring(errors.BOTH_FILE_AND_FILE_PATH_PROVIDED))
		})

		It("should return error when FilePath is whitespace", func() {
			req := common.DeidentifyFileRequest{
				File: common.FileInput{
					FilePath: "   ",
				},
			}
			err := ValidateDeidentifyFileRequest(req)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_FILE_PATH))
		})

		It("should return error when pixel density is negative", func() {
			req := common.DeidentifyFileRequest{
				File: common.FileInput{
					FilePath: filepath.Join(tempDir, "detect.txt"),
				},
				PixelDensity: -1,
			}
			validationErr := ValidateDeidentifyFileRequest(req)
			Expect(validationErr).ToNot(BeNil())
			Expect(validationErr.GetMessage()).To(ContainSubstring(errors.INVALID_PIXEL_DENSITY))
		})

		It("should return error for invalid masking method", func() {
			req := common.DeidentifyFileRequest{
				File: common.FileInput{
					FilePath: filepath.Join(tempDir, "detect.txt"),
				},
				MaskingMethod: "INVALID_METHOD",
			}
			validationErr := ValidateDeidentifyFileRequest(req)
			Expect(validationErr).ToNot(BeNil())
			Expect(validationErr.GetMessage()).To(ContainSubstring(errors.INVALID_MASKING_METHOD))
		})

		It("should accept valid masking methods", func() {
			validMethods := []common.MaskingMethod{common.BLACKBOX, common.BLUR}
			for _, method := range validMethods {
				req := common.DeidentifyFileRequest{
					File: common.FileInput{
						FilePath: filepath.Join(tempDir, "detect.txt"),
					},
					MaskingMethod: method,
				}
				err := ValidateDeidentifyFileRequest(req)
				Expect(err).To(BeNil())
			}
		})

		It("should return error when max resolution is negative", func() {
			req := common.DeidentifyFileRequest{
				File: common.FileInput{
					FilePath: filepath.Join(tempDir, "detect.txt"),
				},
				MaxResolution: -1,
			}
			validationErr := ValidateDeidentifyFileRequest(req)
			Expect(validationErr).ToNot(BeNil())
			Expect(validationErr.GetMessage()).To(ContainSubstring(errors.INVALID_MAX_RESOLUTION))
		})

		Context("Output Directory Validation", func() {
			It("should return error for non-existent directory", func() {
				req := common.DeidentifyFileRequest{
					File: common.FileInput{
						FilePath: filepath.Join(tempDir, "detect.txt"),
					},
					OutputDirectory: "/non/existent/directory",
				}
				validationErr := ValidateDeidentifyFileRequest(req)
				Expect(validationErr).ToNot(BeNil())
				Expect(validationErr.GetMessage()).To(ContainSubstring(errors.OUTPUT_DIRECTORY_NOT_FOUND))
			})

			It("should return nil for valid directory", func() {
				req := common.DeidentifyFileRequest{
					File: common.FileInput{
						FilePath: filepath.Join(tempDir, "detect.txt"),
					},
					OutputDirectory: tempDir,
				}
				validationErr := ValidateDeidentifyFileRequest(req)
				Expect(validationErr).To(BeNil())
			})

			It("should return error for invalid directory permissions", func() {
				tempDir, err := os.MkdirTemp("", "testdir")
				Expect(err).To(BeNil())
				defer os.RemoveAll(tempDir)

				testFile := filepath.Join(tempDir, "detect.txt")
				err = os.WriteFile(testFile, []byte("dummy content"), 0644)
				Expect(err).To(BeNil())

				restrictedDir := filepath.Join(tempDir, "restricted")
				err = os.Mkdir(restrictedDir, 0700)
				Expect(err).To(BeNil())

				err = os.Chmod(restrictedDir, 0000)
				Expect(err).To(BeNil())

				defer os.Chmod(restrictedDir, 0755)

				// Build request with restricted output directory
				req := common.DeidentifyFileRequest{
					File: common.FileInput{
						FilePath: testFile,
					},
					OutputDirectory: restrictedDir,
				}
				validationErr := ValidateDeidentifyFileRequest(req)
				Expect(validationErr).ToNot(BeNil())
				Expect(validationErr.GetMessage()).To(ContainSubstring(errors.INVALID_PERMISSION))
			})
		})

		Context("Wait Time Validation", func() {
			It("should return error for negative wait time", func() {
				req := common.DeidentifyFileRequest{
					File: common.FileInput{
						FilePath: filepath.Join(tempDir, "detect.txt"),
					},
					WaitTime: -1,
				}
				validationErr := ValidateDeidentifyFileRequest(req)
				Expect(validationErr).ToNot(BeNil())
				Expect(validationErr.GetMessage()).To(ContainSubstring(errors.INVALID_WAIT_TIME))
			})

			It("should return error for wait time exceeding limit", func() {
				req := common.DeidentifyFileRequest{
					File: common.FileInput{
						FilePath: filepath.Join(tempDir, "detect.txt"),
					},
					WaitTime: 65,
				}
				validationErr := ValidateDeidentifyFileRequest(req)
				Expect(validationErr).ToNot(BeNil())
				Expect(validationErr.GetMessage()).To(ContainSubstring(errors.WAIT_TIME_EXCEEDS_LIMIT))
			})

			It("should accept valid wait time", func() {
				req := common.DeidentifyFileRequest{
					File: common.FileInput{
						FilePath: filepath.Join(tempDir, "detect.txt"),
					},
					WaitTime: 30,
				}
				err := ValidateDeidentifyFileRequest(req)
				Expect(err).To(BeNil())
			})
		})

		Context("ValidateFilePermissions", func() {
			var resourcePath string

			BeforeEach(func() {
				resourcePath = filepath.Join(tempDir, "detect.txt")
			})

			It("should return error when file does not exist", func() {
				err := ValidateFilePermissions("/non/existent/file.txt", nil)
				Expect(err).ToNot(BeNil())
				Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.FILE_NOT_FOUND_TO_DEIDENTIFY, "/non/existent/file.txt")))
			})

			It("should return error when file is not regular", func() {
				dirPath := tempDir
				validationErr := ValidateFilePermissions(dirPath, nil)
				Expect(validationErr).ToNot(BeNil())
				Expect(validationErr.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.NOT_REGULAR_FILE_TO_DEIDENTIFY, dirPath)))
			})

			Context("File based validation", func() {
				It("should return error when file stat fails for file pointer", func() {
					file, err := os.Open(resourcePath)
					Expect(err).To(BeNil())

					// Close file to cause stat to fail
					file.Close()

					validationErr := ValidateFilePermissions("", file)
					Expect(validationErr).ToNot(BeNil())
					Expect(validationErr.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.UNABLE_TO_STAT_FILE_TO_DEIDENTIFY, file.Name())))
				})

				It("should return error when file pointer is not a regular file", func() {
					// Use directory instead of file
					testFilePath := tempDir

					// Now open the file
					dirFile, err := os.Open(testFilePath)
					Expect(err).To(BeNil(), "Failed to open test file")
					defer dirFile.Close()

					validationErr := ValidateFilePermissions("", dirFile)
					Expect(validationErr).ToNot(BeNil())
					Expect(validationErr.GetMessage()).To(ContainSubstring(fmt.Sprintf(errors.NOT_REGULAR_FILE_TO_DEIDENTIFY, dirFile.Name())))
				})

				It("should return nil for valid file", func() {
					file, err := os.Open(resourcePath)
					Expect(err).To(BeNil())
					defer file.Close()

					err = ValidateFilePermissions("", file)
					Expect(err).To(BeNil())
				})
			})
		})
	})
})
