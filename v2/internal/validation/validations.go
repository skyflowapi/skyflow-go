package validation

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	. "skyflow-go/v2/utils/common"
	skyflowError "skyflow-go/v2/utils/error"
)

func ValidateInsertRequest(request InsertRequest, options InsertOptions) *skyflowError.SkyflowError {
	// Validate table
	if request.Table == "" {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TABLE_KEY_ERROR)
	}

	// Validate values
	if request.Values == nil {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}
	if len(request.Values) == 0 {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}

	// Validate upsert
	if options.Upsert != "" {
		if options.Homogeneous {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT)
		}
	}

	// Validate each key-value pair in values
	for _, valueMap := range request.Values {
		for key, value := range valueMap {
			if value == nil || value == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_VALUES)
			} else if key == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_VALUES)
			}

		}
	}

	// Validate BYOT token strictness
	switch options.TokenMode {
	case DISABLE:
		if options.Tokens != nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKENS_PASSED_FOR_BYOT_DISABLE)
		}
	case ENABLE:
		if options.Tokens == nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
		}
		if len(options.Tokens) != len(request.Values) {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
		}
		if err := ValidateTokensForInsertRequest(options.Tokens, request.Values, ENABLE); err != nil {
			return err
		}
	case ENABLE_STRICT:
		if options.Tokens == nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
		}
		if len(options.Tokens) != len(request.Values) {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
		}
		if err := ValidateTokensForInsertRequest(options.Tokens, request.Values, ENABLE_STRICT); err != nil {
			return err
		}
	}

	return nil
}

func ValidateTokensForInsertRequest(tokens []map[string]interface{}, values []map[string]interface{}, mode BYOT) *skyflowError.SkyflowError {
	if tokens == nil || len(tokens) == 0 {
		if mode == ENABLE || mode == ENABLE_STRICT {
			//logger.Error(fmt.Sprintf(messages.NO_TOKENS_IN_INSERT, insertTag, insertApi.Options.Byot))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, mode))
		}
	} else {
		for i, tokensMap := range tokens {
			fieldsMap := values[i]
			if len(tokensMap) != len(fieldsMap) && mode == ENABLE_STRICT {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
			}
			if len(tokensMap) == 0 {
				//logger.Error(fmt.Sprintf(messages.EMPTY_TOKENS_IN_INSERT, insertTag))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
			}
			for key, tokenKey := range tokensMap {
				if _, exists := fieldsMap[key]; !exists {
					//logger.Error(fmt.Sprintf(messages.MISMATCH_OF_FIELDS_AND_TOKENS, insertTag))
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
				}
				if tokenKey == nil {
					//logError(fmt.Sprintf("Value for key '%s' in tokens or values map is nil.", key))
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_TOKENS)
				} else if fieldsMap[key] == nil {
					return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
				}
			}
		}
	}
	return nil
}
func validateTokenForStrict(tokens map[string]interface{}, values map[string]interface{}, mode BYOT) *skyflowError.SkyflowError {
	if len(tokens) == 0 {
		//logger.Error(fmt.Sprintf(messages.EMPTY_TOKENS_IN_INSERT, insertTag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
	}
	if len(tokens) != len(values) && mode == ENABLE_STRICT {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INSUFFICIENT_TOKENS_PASSED_FOR_BYOT_ENABLE_STRICT)
	}
	for key, tokenKey := range tokens {
		if key == "" {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_TOKENS)
		} else if tokenKey == nil {
			//logError(fmt.Sprintf("Value for key '%s' in tokens or values map is nil.", key))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_TOKENS)
		}
		if _, exists := values[key]; !exists {
			//logger.Error(fmt.Sprintf(messages.MISMATCH_OF_FIELDS_AND_TOKENS, insertTag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
		}
		if values[key] == nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
		}
	}
	return nil
}

// validateVaultConfig function
func ValidateVaultConfig(vaultConfig VaultConfig) *skyflowError.SkyflowError {
	// Validate VaultId
	if vaultConfig.VaultId == "" {
		log.Println("Error: Vault ID is required.")
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_VAULT_ID)
	} else if vaultConfig.ClusterId == "" {
		log.Println("Error: Cluster ID is required.")
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CLUSTER_ID)
	}
	return nil
}

// ValidateConnectionConfig validates the ConnectionConfig struct
func ValidateConnectionConfig(config ConnectionConfig) *skyflowError.SkyflowError {
	if config.ConnectionId == "" {
		log.Println("Error: Connection ID is required.")
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CONNECTION_ID)
	} else if config.ConnectionUrl == "" {
		log.Println("Error: Connection URL is required.", config)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CONNECTION_URL)
	}

	// Parse the URL
	_, err := url.Parse(config.ConnectionUrl)
	if err != nil {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CONNECTION_URL)
	}
	//if parsedURL.Scheme != "https" {
	//	return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CONNECTION_URL)
	//}
	return nil
}

// ValidateCredentials validates the credentials object
func ValidateCredentials(credentials Credentials) *skyflowError.SkyflowError {
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
		//log.Println("Error: Multiple token generation means passed")
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MULTIPLE_TOKEN_GENERATION_MEANS_PASSED)
	} else if credPresent < 1 {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.NO_TOKEN_GENERATION_MEANS_PASSED)
	}

	// API key validation
	if credentials.ApiKey != "" {
		// Validate API key format
		apiKeyRegex := `^sky-[a-zA-Z0-9]{5}-[a-fA-F0-9]{32}$` // Replace this with the actual regex
		matched, err := regexp.MatchString(apiKeyRegex, credentials.ApiKey)
		if err != nil {
			log.Println("Error: Failed to validate API key regex")
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_API_KEY)
		}
		if !matched {
			log.Println("Error: Invalid API key format")
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_API_KEY)
		}
	}

	// Roles validation
	if credentials.Roles != nil {
		if len(credentials.Roles) == 0 {
			log.Println("Error: Roles list is empty")
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ROLES)
		}
		for index, role := range credentials.Roles {
			if role == "" {
				log.Printf("Error: Role at index %d is empty or null\n", index)
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ROLE_IN_ROLES)
			}
		}
	}

	return nil
}

// ValidateInvokeConnectionRequest validates the fields of InvokeConnectionRequest
func ValidateInvokeConnectionRequest(request InvokeConnectionRequest) *skyflowError.SkyflowError {
	// Validate headers
	if request.Headers != nil {
		if len(request.Headers) == 0 {
			log.Println("Error: Empty request headers")
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_HEADER)
		}
		for key, value := range request.Headers {
			if key == "" || value == "" {
				log.Println("Error: Invalid request header key or value")
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_REQUEST_HEADERS)
			}
		}
	}

	// Validate path parameters
	if request.PathParams != nil {
		if len(request.PathParams) == 0 {
			log.Println("Error: Empty path parameters")
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETERS)
		}
		for key, value := range request.PathParams {
			if key == "" {
				log.Println("Error: Invalid path parameter key or value")
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETER_NAME)
			} else if value == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETER_VALUE)

			}
		}
	}

	// Validate query parameters
	if request.QueryParams != nil {
		if len(request.QueryParams) == 0 {
			log.Println("Error: Empty query parameters")
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_QUERY_PARAM)
		}
		for key, value := range request.QueryParams {
			if key == "" || value == nil || value == "" {
				log.Println("Error: Invalid query parameter key or value")
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_QUERY_PARAM)
			}
		}
	}
	// Validate body
	if request.Body != nil {
		if len(request.Body) == 0 {
			log.Println("Error: Empty request body")
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_BODY)
		}
	}
	return nil
}

func ValidateDetokenizeRequest(request DetokenizeRequest) *skyflowError.SkyflowError {
	if request.Tokens == nil {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_DATA_TOKENS)
	} else if len(request.Tokens) == 0 {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS_DETOKENIZE)
	} else {
		for _, token := range request.Tokens {
			if token == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKEN_IN_DATA_TOKEN)
			}
		}
	}
	return nil
}

func ValidateGetRequest(getRequest GetRequest, options GetOptions) *skyflowError.SkyflowError {
	// Check if the table is valid
	if getRequest.Table == "" {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	}

	// Validate Ids
	if getRequest.Ids != nil {
		if len(getRequest.Ids) == 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_IDS)
		}
		for _, id := range getRequest.Ids {
			if id == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_IDS)
			}
		}
	}

	// Validate Fields
	if options.Fields != nil {
		if len(options.Fields) == 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_FIELDS)
		}
		for _, field := range options.Fields {
			if field == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_FIELD_IN_FIELDS)
			}
		}
	}

	if options.ReturnTokens {
		if options.ColumnName != "" || options.ColumnValues != nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKENS_GET_COLUMN_NOT_SUPPORTED)
		}
	}

	// ColumnName and ColumnValues logic
	if getRequest.Ids == nil && options.ColumnName == "" && options.ColumnValues == nil {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.UNIQUE_COLUMN_OR_IDS_KEY_ERROR)
	}
	if getRequest.Ids != nil && (options.ColumnName != "" || options.ColumnValues != nil) {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.BOTH_IDS_AND_COLUMN_DETAILS_SPECIFIED)
	}
	if options.ColumnName == "" && options.ColumnValues != nil {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.COLUMN_NAME_KEY_ERROR)
	}
	if options.ColumnName != "" && options.ColumnValues == nil {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_COLUMN_VALUES)
	}
	if options.ColumnName != "" {
		if len(options.ColumnValues) == 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_COLUMN_VALUES)
		}
		for _, columnValue := range options.ColumnValues {
			if columnValue == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_COLUMN_VALUES)
			}
		}
	}
	return nil
}

func ValidateDeleteRequest(request DeleteRequest) *skyflowError.SkyflowError {
	if request.Table == "" {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	} else if request.Ids == nil || len(request.Ids) == 0 {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_IDS)
	}
	for _, id := range request.Ids {
		if id == "" {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_IDS)
		}
	}
	return nil
}

func ValidateQueryRequest(request QueryRequest) *skyflowError.SkyflowError {
	if request.Query == "" {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_QUERY)
	}
	return nil
}

func ValidateTokenizeRequest(request []TokenizeRequest) *skyflowError.SkyflowError {
	if request == nil {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TOKENIZE_REQUEST)
	} else if len(request) == 0 {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TOKENIZE_REQUEST)
	} else {
		for _, tokenize := range request {
			if tokenize.ColumnGroup == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_COLUMN_VALUES)
			} else if tokenize.Value == "" {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_COLUMN_VALUES)
			}
		}
	}
	return nil
}

func ValidateUpdateRequest(request UpdateRequest, options UpdateOptions) *skyflowError.SkyflowError {
	if request.Table == "" {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	} else if request.Id == "" {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_UPDATE)
	} else if request.Values == nil || len(request.Values) == 0 {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}

	for key, value := range request.Values {
		if value == "" {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_VALUES)
		} else if key == "" {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_VALUES)
		}
	}
	switch options.TokenMode {
	case DISABLE:
		if request.Tokens != nil || len(request.Tokens) != 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKENS_PASSED_FOR_BYOT_DISABLE)
		}
	case ENABLE:
		if request.Tokens == nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, options.TokenMode))
		}
		err := validateTokenForStrict(request.Tokens, request.Values, options.TokenMode)
		if err != nil {
			return err
		}
	case ENABLE_STRICT:
		if request.Tokens == nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, options.TokenMode))
		}
		err := validateTokenForStrict(request.Tokens, request.Values, options.TokenMode)
		if err != nil {
			return err
		}
	}
	return nil
}
