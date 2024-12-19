package validation

import (
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	"github.com/skyflowapi/skyflow-go/v2/utils/messages"
	"net/url"
	"strings"
)

func ValidateInsertRequest(request common.InsertRequest, options common.InsertOptions) *skyflowError.SkyflowError {
	// Validate table
	tag := "Insert"
	if request.Table == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TABLE_KEY_ERROR)
	}

	// Validate values
	if request.Values == nil {
		logger.Error(fmt.Sprintf(logs.VALUES_IS_REQUIRED, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}
	if len(request.Values) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_VALUES, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}

	// Validate upsert
	if options.Upsert != "" {
		if options.Homogeneous {
			logger.Error(fmt.Sprintf(logs.HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT)
		}
	}

	// Validate each key-value pair in values
	for _, valueMap := range request.Values {
		for key, value := range valueMap {
			if value == nil || value == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_VALUE_IN_VALUES, tag, key))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_VALUES)
			} else if key == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_KEY_IN_VALUES, tag))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_VALUES)
			}

		}
	}

	// Validate BYOT token strictness
	switch options.TokenMode {
	case common.DISABLE:
		if options.Tokens != nil {
			logger.Error(fmt.Sprintf(logs.TOKENS_NOT_ALLOWED_WITH_BYOT_DISABLE, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKENS_PASSED_FOR_BYOT_DISABLE)
		}
	case common.ENABLE:
		if options.Tokens == nil {
			logger.Error(fmt.Sprintf(logs.TOKENS_REQUIRED_WITH_BYOT, tag, common.ENABLE))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
		}
		if len(options.Tokens) != len(request.Values) {
			logger.Error(fmt.Sprintf(logs.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
		}
		if err := ValidateTokensForInsertRequest(options.Tokens, request.Values, common.ENABLE); err != nil {
			return err
		}
	case common.ENABLE_STRICT:
		if options.Tokens == nil {
			logger.Error(fmt.Sprintf(logs.TOKENS_REQUIRED_WITH_BYOT, tag, common.ENABLE))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
		}
		if len(options.Tokens) != len(request.Values) {
			logger.Error(fmt.Sprintf(logs.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
		}
		if err := ValidateTokensForInsertRequest(options.Tokens, request.Values, common.ENABLE_STRICT); err != nil {
			return err
		}
	}

	return nil
}

func ValidateTokensForInsertRequest(tokens []map[string]interface{}, values []map[string]interface{}, mode common.BYOT) *skyflowError.SkyflowError {
	tag := "insert"
	if tokens == nil || len(tokens) == 0 {
		if mode == common.ENABLE || mode == common.ENABLE_STRICT {
			logger.Error(fmt.Sprintf(logs.EMPTY_TOKENS, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, mode))
		}
	} else {
		for i, tokensMap := range tokens {
			fieldsMap := values[i]
			if len(tokensMap) != len(fieldsMap) && mode == common.ENABLE_STRICT {
				logger.Error(fmt.Sprintf(logs.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT, tag))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
			}
			if len(tokensMap) == 0 {
				logger.Error(fmt.Sprintf(logs.EMPTY_TOKENS, tag))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
			}
			for key, tokenKey := range tokensMap {
				if _, exists := fieldsMap[key]; !exists {
					logger.Error(fmt.Sprintf(logs.MISMATCH_OF_FIELDS_AND_TOKENS, tag))
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
				}
				if tokenKey == nil {
					logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_VALUE_IN_TOKENS, tag, key))
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_TOKENS)
				} else if fieldsMap[key] == nil {
					logger.Error(fmt.Sprintf(logs.MISMATCH_OF_FIELDS_AND_TOKENS, tag))
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
				}
			}
		}
	}
	return nil
}
func validateTokenForStrict(tokens map[string]interface{}, values map[string]interface{}, mode common.BYOT, tag string) *skyflowError.SkyflowError {
	if len(tokens) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_TOKENS, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
	}
	if len(tokens) != len(values) && mode == common.ENABLE_STRICT {
		logger.Error(fmt.Sprintf(logs.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
	}
	for key, tokenKey := range tokens {
		if key == "" {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_KEY_IN_TOKENS, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_TOKENS)
		} else if tokenKey == nil {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_VALUE_IN_TOKENS, tag, key))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_TOKENS)
		}
		if _, exists := values[key]; !exists {
			logger.Error(fmt.Sprintf(logs.MISMATCH_OF_FIELDS_AND_TOKENS, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
		}
		if values[key] == nil {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_VALUE_IN_VALUES, tag, key))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
		}
	}
	return nil
}

// validateVaultConfig function
func ValidateVaultConfig(vaultConfig common.VaultConfig) *skyflowError.SkyflowError {
	// Validate VaultId
	if vaultConfig.VaultId == "" {
		logger.Error(logs.VAULT_ID_IS_REQUIRED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_VAULT_ID)
	} else if vaultConfig.ClusterId == "" {
		logger.Error(logs.CLUSTER_ID_IS_REQUIRED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CLUSTER_ID)
	}
	return nil
}

// ValidateConnectionConfig validates the ConnectionConfig struct
func ValidateConnectionConfig(config common.ConnectionConfig) *skyflowError.SkyflowError {
	if config.ConnectionId == "" {
		logger.Error(logs.CONNECTION_ID_IS_REQUIRED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CONNECTION_ID)
	} else if config.ConnectionUrl == "" {
		logger.Error(logs.CONNECTION_URL_IS_REQUIRED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CONNECTION_URL)
	}

	// Parse the URL
	_, err := url.Parse(config.ConnectionUrl)
	if err != nil {
		logger.Error(logs.INVALID_CONNECTION_URL)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CONNECTION_URL)
	}
	//if parsedURL.Scheme != "https" {
	//	return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CONNECTION_URL)
	//}
	return nil
}

// ValidateCredentials validates the credentials object
func ValidateCredentials(credentials common.Credentials) *skyflowError.SkyflowError {
	credPresent := 0

	// Count non-null members
	if credentials.Path != "" {
		credPresent++
	}
	if credentials.CredentialsString != "" {
		credPresent++
	}
	if credentials.Token != "" {
		credPresent++
	}
	if credentials.ApiKey != "" {
		credPresent++
	}

	// Validation logic
	if credPresent > 1 {
		logger.Error(logs.MULTIPLE_TOKEN_GENERATION_MEANS_PASSED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MULTIPLE_TOKEN_GENERATION_MEANS_PASSED)
	} else if credPresent < 1 {
		logger.Error(logs.NO_TOKEN_GENERATION_MEANS_PASSED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.NO_TOKEN_GENERATION_MEANS_PASSED)
	}

	// API key validation
	if credentials.ApiKey != "" {
		// Validate API key format
		if len(credentials.ApiKey) != 42 || !strings.Contains(credentials.ApiKey, "sky-") {
			logger.Error(logs.INVALID_API_KEY)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_API_KEY)

		}
	}

	// Roles validation
	if credentials.Roles != nil {
		if len(credentials.Roles) == 0 {
			logger.Error(logs.EMPTY_ROLES)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ROLES)
		}
		for index, role := range credentials.Roles {
			if role == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_ROLE_IN_ROLES, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ROLE_IN_ROLES)
			}
		}
	}

	return nil
}

// ValidateInvokeConnectionRequest validates the fields of InvokeConnectionRequest
func ValidateInvokeConnectionRequest(request common.InvokeConnectionRequest) *skyflowError.SkyflowError {
	// Validate headers
	if request.Headers != nil {
		if len(request.Headers) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_REQUEST_HEADERS, "InvokeConnectionRequest"))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_HEADER)
		}
		for key, value := range request.Headers {
			if key == "" || value == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_REQUEST_HEADERS, "InvokeConnectionRequest"))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_REQUEST_HEADERS)
			}
		}
	}

	// Validate path parameters
	if request.PathParams != nil {
		if len(request.PathParams) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_PATH_PARAMS, "InvokeConnectionRequest"))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETERS)
		}
		for key, value := range request.PathParams {
			if key == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_PATH_PARAM, "InvokeConnectionRequest"))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETER_NAME)
			} else if value == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_PATH_PARAM, "InvokeConnectionRequest"))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETER_VALUE)

			}
		}
	}

	// Validate query parameters
	if request.QueryParams != nil {
		if len(request.QueryParams) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_QUERY_PARAMS, "InvokeConnectionRequest"))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_QUERY_PARAM)
		}
		for key, value := range request.QueryParams {
			if key == "" || value == nil || value == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_QUERY_PARAM, "InvokeConnectionRequest"))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_QUERY_PARAM)
			}
		}
	}
	// Validate body
	if request.Body != nil {
		if len(request.Body) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_REQUEST_BODY, "InvokeConnectionRequest"))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_BODY)
		}
	}
	return nil
}

func ValidateDetokenizeRequest(request common.DetokenizeRequest) *skyflowError.SkyflowError {
	tag := "DetokenizeRequest"
	if request.Tokens == nil {
		logger.Error(fmt.Sprintf(logs.TOKENS_REQUIRED, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_DATA_TOKENS)
	} else if len(request.Tokens) == 0 {
		logger.Error(logs.EMPTY_TOKENS_IN_DETOKENIZE)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS_DETOKENIZE)
	} else {
		for index, token := range request.Tokens {
			if token == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_TOKEN_IN_TOKENS, tag, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKEN_IN_DATA_TOKEN)
			}
		}
	}
	return nil
}

func ValidateGetRequest(getRequest common.GetRequest, options common.GetOptions) *skyflowError.SkyflowError {
	// Check if the table is valid
	tag := "Get"
	if getRequest.Table == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	}

	// Validate Ids
	if getRequest.Ids != nil {
		if len(getRequest.Ids) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_IDS, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_IDS)
		}
		for index, id := range getRequest.Ids {
			if id == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_ID_IN_IDS, tag, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_IDS)
			}
		}
	}

	// Validate Fields
	if options.Fields != nil {
		if len(options.Fields) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_FIELDS, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_FIELDS)
		}
		for index, field := range options.Fields {
			if field == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_FIELD_IN_FIELDS, tag, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_FIELD_IN_FIELDS)
			}
		}
	}

	if options.ReturnTokens {
		if options.ColumnName != "" || options.ColumnValues != nil {
			logger.Error(fmt.Sprintf(logs.TOKENIZATION_NOT_SUPPORTED_WITH_REDACTION, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKENS_GET_COLUMN_NOT_SUPPORTED)
		}
	}

	// ColumnName and ColumnValues logic
	if getRequest.Ids == nil && options.ColumnName == "" && options.ColumnValues == nil {
		logger.Error(fmt.Sprintf(logs.NEITHER_IDS_NOR_COLUMN_NAME_PASSED, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.UNIQUE_COLUMN_OR_IDS_KEY_ERROR)
	}
	if getRequest.Ids != nil && (options.ColumnName != "" || options.ColumnValues != nil) {
		logger.Error(fmt.Sprintf(logs.BOTH_IDS_AND_COLUMN_NAME_PASSED, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.BOTH_IDS_AND_COLUMN_DETAILS_SPECIFIED)
	}
	if options.ColumnName == "" && options.ColumnValues != nil {
		logger.Error(logs.EMPTY_COLUMN_NAME_IN_GET_COLUMN)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.COLUMN_NAME_KEY_ERROR)
	}
	if options.ColumnName != "" && options.ColumnValues == nil {
		logger.Error(logs.EMPTY_COLUMN_VALUES_IN_GET_COLUMN)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_COLUMN_VALUES)
	}
	if options.ColumnName != "" {
		if len(options.ColumnValues) == 0 {
			logger.Error(logs.EMPTY_COLUMN_VALUES_IN_GET_COLUMN)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_COLUMN_VALUES)
		}
		for index, columnValue := range options.ColumnValues {
			if columnValue == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_COLUMN_VALUE_IN_COLUMN_VALUES, tag, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_COLUMN_VALUES)
			}
		}
	}
	return nil
}

func ValidateDeleteRequest(request common.DeleteRequest) *skyflowError.SkyflowError {
	tag := "delete"
	if request.Table == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	} else if request.Ids == nil || len(request.Ids) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_IDS, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_IDS)
	}
	for index, id := range request.Ids {
		if id == "" {
			logger.Error(fmt.Sprintf(logs.INVALID_ID, tag, index))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_IDS)
		}
	}
	return nil
}

func ValidateQueryRequest(request common.QueryRequest) *skyflowError.SkyflowError {
	if request.Query == "" {
		logger.Error(logs.EMPTY_QUERY)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_QUERY)
	}
	return nil
}

func ValidateTokenizeRequest(request []common.TokenizeRequest) *skyflowError.SkyflowError {
	if request == nil {
		logger.Error(logs.INVALID_TOKENIZE_REQUEST)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TOKENIZE_REQUEST)
	} else if len(request) == 0 {
		logger.Error(logs.INVALID_TOKENIZE_REQUEST)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TOKENIZE_REQUEST)
	} else {
		for index, tokenize := range request {
			if tokenize.ColumnGroup == "" {
				logger.Error(logs.EMPTY_COLUMN_GROUP_IN_COLUMN_VALUES, index)
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_COLUMN_VALUES)
			} else if tokenize.Value == "" {
				logger.Error(logs.EMPTY_OR_NULL_COLUMN_VALUE_IN_COLUMN_VALUES)
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_COLUMN_VALUES)
			}
		}
	}
	return nil
}

func ValidateUpdateRequest(request common.UpdateRequest, options common.UpdateOptions) *skyflowError.SkyflowError {
	tag := "update"
	if request.Table == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	} else if request.Id == "" {
		logger.Error(logs.INVALID_SKYFLOW_ID_IN_UPDATE)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_UPDATE)
	} else if request.Values == nil || len(request.Values) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_VALUES, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}

	for key, value := range request.Values {
		if value == "" {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_VALUE_IN_VALUES, tag, key))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_VALUES)
		} else if key == "" {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_KEY_IN_VALUES, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_VALUES)
		}
	}
	switch options.TokenMode {
	case common.DISABLE:
		if request.Tokens != nil || len(request.Tokens) != 0 {
			logger.Error(fmt.Sprintf(logs.TOKENS_NOT_ALLOWED_WITH_BYOT_DISABLE, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKENS_PASSED_FOR_BYOT_DISABLE)
		}
	case common.ENABLE:
		if request.Tokens == nil {
			logger.Error(fmt.Sprintf(logs.TOKENS_REQUIRED_WITH_BYOT, tag, common.ENABLE))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, options.TokenMode))
		}
		err := validateTokenForStrict(request.Tokens, request.Values, options.TokenMode, tag)
		if err != nil {
			return err
		}
	case common.ENABLE_STRICT:
		if request.Tokens == nil {
			logger.Error(fmt.Sprintf(logs.TOKENS_REQUIRED_WITH_BYOT, tag, common.ENABLE_STRICT))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, options.TokenMode))
		}
		err := validateTokenForStrict(request.Tokens, request.Values, options.TokenMode, tag)
		if err != nil {
			return err
		}
	}
	return nil
}
