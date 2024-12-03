package validation_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "skyflow-go/v2/internal/validation"
	"skyflow-go/v2/utils/common"
	errors "skyflow-go/v2/utils/error"
	"testing"
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
		It("should return error when tokens are not passed for all values objectin BYOT ENABLE mode", func() {
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
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT))
		})
		It("should return error when tokens are not passed for all values objectin BYOT ENABLE mode", func() {
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
		It("should return error when tokens are passed in BYOT ENABLE_STRICT mode", func() {
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
			validCredentials   common.Credentials
			invalidCredentials common.Credentials
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

			invalidCredentials = common.Credentials{}
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

		Context("Invalid Credentials", func() {
			It("should return error for invalid Credentials", func() {
				config := common.VaultConfig{
					VaultId:     "valid-vault-id",
					ClusterId:   "valid-cluster-id",
					Env:         common.PROD,
					Credentials: invalidCredentials,
				}
				err := ValidateVaultConfig(config)
				Expect(err).ToNot(BeNil())
				Expect(err.GetMessage()).ToNot(BeEmpty())
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
					Expect(err.GetCode()).To(ContainSubstring("400"))
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
					Expect(err.GetCode()).To(ContainSubstring("400"))
					Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_CONNECTION_URL))
				})

				It("should return an error if ConnectionUrl is invalid", func() {
					config := common.ConnectionConfig{
						ConnectionId:  "valid-id",
						ConnectionUrl: "invalid-url",
						Credentials:   common.Credentials{},
					}

					err := ValidateConnectionConfig(config)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring("400"))
					Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_CONNECTION_URL))
				})

				It("should return an error if ConnectionUrl is not HTTPS", func() {
					config := common.ConnectionConfig{
						ConnectionId:  "valid-id",
						ConnectionUrl: "http://valid.url",
						Credentials:   common.Credentials{},
					}

					err := ValidateConnectionConfig(config)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring("400"))
					Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_CONNECTION_URL))
				})
				It("should return an error if ConnectionUrl is invalid", func() {
					config := common.ConnectionConfig{
						ConnectionId:  "valid-id",
						ConnectionUrl: "demo",
						Credentials:   common.Credentials{},
					}

					err := ValidateConnectionConfig(config)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring("400"))
					Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_CONNECTION_URL))
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
					Expect(err.GetCode()).To(ContainSubstring("400"))
					Expect(err.GetMessage()).To(ContainSubstring(errors.NO_TOKEN_GENERATION_MEANS_PASSED))
				})

				It("should return an error if multiple token generation means are passed", func() {
					credentials := common.Credentials{
						Path:              "some/path",
						CredentialsString: "some-string",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring("400"))
					Expect(err.GetMessage()).To(ContainSubstring(errors.MULTIPLE_TOKEN_GENERATION_MEANS_PASSED))
				})

				It("should return an error for invalid API key format", func() {
					credentials := common.Credentials{
						ApiKey: "invalid-api-key",
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring("400"))
					Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_API_KEY))
				})

				It("should return an error if roles list is empty", func() {
					credentials := common.Credentials{
						Token: "token",
						Roles: []string{},
					}
					err := ValidateCredentials(credentials)
					Expect(err).To(HaveOccurred())
					Expect(err.GetCode()).To(ContainSubstring("400"))
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
					Expect(err.GetCode()).To(ContainSubstring("400"))
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
				Tokens: nil,
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_DATA_TOKENS))
		})

		It("should return an error if tokens are empty", func() {
			request := common.DetokenizeRequest{
				Tokens: []string{},
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKENS_DETOKENIZE))
		})

		It("should return an error if any token is empty", func() {
			request := common.DetokenizeRequest{
				Tokens: []string{"validToken", ""},
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TOKEN_IN_DATA_TOKEN))
		})

		It("should not return an error if all tokens are valid", func() {
			request := common.DetokenizeRequest{
				Tokens: []string{"token1", "token2"},
			}
			err := ValidateDetokenizeRequest(request)
			Expect(err).To(BeNil())
		})
	})
	Context("when validating update requests", func() {
		It("should return an error if the table is empty", func() {
			request := common.UpdateRequest{
				Table:  "",
				Id:     "123",
				Values: map[string]interface{}{"key": "value"},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TABLE))
		})

		It("should return an error if the ID is empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Id:     "",
				Values: map[string]interface{}{"key": "value"},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_ID_IN_UPDATE))
		})

		It("should return an error if the values are nil or empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Id:     "123",
				Values: nil,
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUES))
		})

		It("should return an error if tokens are nil or empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Id:     "123",
				Values: map[string]interface{}{"key": "value"},
				Tokens: nil,
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).To(BeNil())
		})

		It("should return an error if a value is empty", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Id:     "123",
				Values: map[string]interface{}{"key": ""},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUE_IN_VALUES))
		})

		It("should return an error if a key is empty in values", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Id:     "123",
				Values: map[string]interface{}{"": "value"},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_KEY_IN_VALUES))
		})

		It("should return an error if tokens are passed with TokenMode DISABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Id:     "123",
				Values: map[string]interface{}{"key": "value"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value"},
				Tokens: map[string]interface{}{"key": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).To(BeNil())
		})

		It("should return an error for invalid input with TokenMode ENABLE", func() {
			request := common.UpdateRequest{
				Table:  "test_table",
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": "value2"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": "value2"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": "value2"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": "value2"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": "value2"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": "value2"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": "value2"},
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
				Id:     "123",
				Values: map[string]interface{}{"key": "value", "key2": nil},
				Tokens: map[string]interface{}{"key": "value", "key2": "token"},
			}
			options := common.UpdateOptions{TokenMode: common.ENABLE}
			err := ValidateUpdateRequest(request, options)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(errors.MISMATCH_OF_FIELDS_AND_TOKENS))
		})

	})
	Context("ValidateTokenizeRequest", func() {
		var (
			request *[]common.TokenizeRequest
		)
		It("should return INVALID_TOKENIZE_REQUEST error", func() {
			request = nil
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring("400"))
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_TOKENIZE_REQUEST))
		})
		It("should return INVALID_TOKENIZE_REQUEST error", func() {
			request = &[]common.TokenizeRequest{}
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring("400"))
			Expect(err.GetMessage()).To(ContainSubstring(errors.INVALID_TOKENIZE_REQUEST))
		})
		It("should return EMPTY_VALUE_IN_COLUMN_VALUES error", func() {
			request = &[]common.TokenizeRequest{
				{ColumnGroup: "", Value: "valid_value"},
			}
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring("400"))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_VALUE_IN_COLUMN_VALUES))
		})
		It("should return EMPTY_COLUMN_VALUES error", func() {
			request = &[]common.TokenizeRequest{
				{ColumnGroup: "valid_group", Value: ""},
			}
			err := ValidateTokenizeRequest(request)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring("400"))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_COLUMN_VALUES))
		})
		It("should return nil", func() {
			request = &[]common.TokenizeRequest{
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
			Expect(err.GetCode()).To(ContainSubstring("400"))
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
			Expect(err.GetCode()).To(ContainSubstring("400"))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_TABLE))
		})
		It("should return an EMPTY_IDS error", func() {
			request := common.DeleteRequest{
				Table: "test_table",
				Ids:   nil,
			}

			err := ValidateDeleteRequest(request)

			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring("400"))
			Expect(err.GetMessage()).To(ContainSubstring(errors.EMPTY_IDS))
		})
		It("should return an EMPTY_ID_IN_IDS error", func() {
			request := common.DeleteRequest{
				Table: "test_table",
				Ids:   []string{"id1", ""},
			}

			err := ValidateDeleteRequest(request)

			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(ContainSubstring("400"))
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

})
