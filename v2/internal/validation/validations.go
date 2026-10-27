package validation

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"

	cfg "github.com/skyflowapi/skyflow-go/common/config"
	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

func ValidateInsertRequest(request common.InsertRequest, options common.InsertOptions) *skyflowError.SkyflowError {
	// Validate table
	tag := constants.REQUEST_INSERT
	if request.Table == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TABLE_KEY_ERROR)
	}

	// Validate upsert
	if options.Upsert != "" {
		if options.Homogeneous {
			logger.Error(fmt.Sprintf(logs.HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.HOMOGENOUS_NOT_SUPPORTED_WITH_UPSERT)
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
		if err := validateValues(request.Values, tag); err != nil {
			return err
		}
		if options.Tokens == nil {
			logger.Error(fmt.Sprintf(logs.TOKENS_REQUIRED_WITH_BYOT, tag, common.ENABLE))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKENS)
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
func validateValues(values []map[string]interface{}, tag string) *skyflowError.SkyflowError {
	if values == nil {
		logger.Error(fmt.Sprintf(logs.VALUES_IS_REQUIRED, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}
	if len(values) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_VALUES, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUES)
	}
	for _, valueMap := range values {
		for key := range valueMap {
			if key == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_KEY_IN_VALUES, tag))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_VALUES)
			}
		}
	}
	return nil
}

func ValidateTokensForInsertRequest(tokens []map[string]interface{}, values []map[string]interface{}, mode common.BYOT) *skyflowError.SkyflowError {
	tag := constants.REQUEST_INSERT_LOWER
	if tokens == nil || len(tokens) == 0 {
		if mode == common.ENABLE || mode == common.ENABLE_STRICT {
			logger.Error(fmt.Sprintf(logs.EMPTY_TOKENS, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, mode))
		}
	} else {
		for i, tokensMap := range tokens {
			if i >= len(values) {
				logger.Error(fmt.Sprintf(logs.MISMATCH_OF_FIELDS_AND_TOKENS, tag))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISMATCH_OF_FIELDS_AND_TOKENS)
			}
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
	// id will be ignored while comparing length
	if len(tokens) != len(values)-1 && mode == common.ENABLE_STRICT {
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
// configValidator is the shared implementation (skyflow-go/common/config) of the
// vault-config and credential checks, configured with v2's SDK prefix.
var configValidator = cfg.NewValidator(constants.SDK_PREFIX, constants.SDK_LOG_PREFIX)

func ValidateVaultConfig(vaultConfig common.VaultConfig) *skyflowError.SkyflowError {
	return configValidator.ValidateVaultConfig(vaultConfig)
}

func ValidateUpdateVaultConfig(vaultConfig common.VaultConfig) *skyflowError.SkyflowError {
	return configValidator.ValidateUpdateVaultConfig(vaultConfig)
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
func ValidateUpdateConnectionConfig(config common.ConnectionConfig) *skyflowError.SkyflowError {
	if config.ConnectionId == "" {
		logger.Error(logs.CONNECTION_ID_IS_REQUIRED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_CONNECTION_ID)
	}

	if config.ConnectionUrl != "" {
		_, err := url.Parse(config.ConnectionUrl)
		if err != nil {
			logger.Error(logs.INVALID_CONNECTION_URL)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CONNECTION_URL)
		}
	}
	return nil
}

// ValidateCredentials validates the credentials object
func ValidateCredentials(credentials common.Credentials) *skyflowError.SkyflowError {
	return configValidator.ValidateCredentials(credentials)
}

// ValidateInvokeConnectionRequest validates the fields of InvokeConnectionRequest
func ValidateInvokeConnectionRequest(request common.InvokeConnectionRequest) *skyflowError.SkyflowError {
	// Validate headers
	if request.Headers != nil {
		if len(request.Headers) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_REQUEST_HEADERS, constants.REQUEST_INVOKE_CONNECTION))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_HEADER)
		}
		for key, value := range request.Headers {
			if key == "" || value == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_REQUEST_HEADERS, constants.REQUEST_INVOKE_CONNECTION))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_REQUEST_HEADERS)
			}
		}
	}

	// Validate path parameters
	if request.PathParams != nil {
		if len(request.PathParams) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_PATH_PARAMS, constants.REQUEST_INVOKE_CONNECTION))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETERS)
		}
		for key, value := range request.PathParams {
			if key == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_PATH_PARAM, constants.REQUEST_INVOKE_CONNECTION))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETER_NAME)
			} else if value == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_PATH_PARAM, constants.REQUEST_INVOKE_CONNECTION))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_PARAMETER_VALUE)

			}
		}
	}

	// Validate query parameters
	if request.QueryParams != nil {
		if len(request.QueryParams) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_QUERY_PARAMS, constants.REQUEST_INVOKE_CONNECTION))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_QUERY_PARAM)
		}
		for key, value := range request.QueryParams {
			if key == "" || value == nil || value == "" {
				logger.Error(fmt.Sprintf(logs.INVALID_QUERY_PARAM, constants.REQUEST_INVOKE_CONNECTION))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_QUERY_PARAM)
			}
		}
	}
	// Validate body
	if request.Body != nil {
		// Check if body is a map and if it's empty
		if bodyMap, ok := request.Body.(map[string]interface{}); ok {
			if len(bodyMap) == 0 {
				logger.Error(fmt.Sprintf(logs.EMPTY_REQUEST_BODY, constants.REQUEST_INVOKE_CONNECTION))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_BODY)
			}
		}
	}
	if request.Method != "" {
		method := request.Method.IsValid()
		if !method {
			logger.Error(fmt.Sprintf(logs.INVALID_METHOD_NAME))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_METHOD_NAME)
		}
	}
	return nil
}

func ValidateDetokenizeRequest(request common.DetokenizeRequest) *skyflowError.SkyflowError {
	tag := constants.REQUEST_DETOKENIZE
	if request.DetokenizeData == nil {
		logger.Error(fmt.Sprintf(logs.DETOKENIZE_DATA_REQUIRED, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_DETOKENIZE_DATA)
	} else if len(request.DetokenizeData) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_DETOKENIZE_DATA, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_DETOKENIZE_DATA)
	} else {
		for index, token := range request.DetokenizeData {
			if token.Token == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_TOKEN_IN_DETOKENIZE_DATA, tag, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TOKEN_IN_DETOKENIZE_DATA)
			}
		}
	}
	return nil
}

func ValidateGetRequest(getRequest common.GetRequest, options common.GetOptions) *skyflowError.SkyflowError {
	// Check if the table is valid
	tag := constants.REQUEST_GET
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
	if options.ReturnTokens && options.RedactionType != "" {
		logger.Error(fmt.Sprintf(logs.TOKENIZATION_NOT_SUPPORTED_WITH_REDACTION, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.REDACTION_WITH_TOKENS_NOT_SUPPORTED)
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
	tag := constants.REQUEST_DELETE
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
				logger.Error(fmt.Sprintf(logs.EMPTY_COLUMN_GROUP_IN_COLUMN_VALUES, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_VALUE_IN_COLUMN_VALUES)
			} else if tokenize.Value == "" {
				logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_COLUMN_VALUE_IN_COLUMN_VALUES, constants.REQUEST_TOKENIZE, index))
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_COLUMN_VALUES)
			}
		}
	}
	return nil
}

func ValidateUpdateRequest(request common.UpdateRequest, options common.UpdateOptions) *skyflowError.SkyflowError {
	tag := constants.REQUEST_UPDATE
	if request.Table == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	}
	if request.Data == nil || len(request.Data) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_DATA, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_DATA)
	}
	skyflowId, _ := helpers.GetSkyflowID(request.Data)
	if skyflowId == "" {
		logger.Error(logs.INVALID_SKYFLOW_ID_IN_UPDATE)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_UPDATE)
	}

	for key, data := range request.Data {
		if data == "" {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_VALUE_IN_DATA, tag, key))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_DATA_IN_DATA_KEY)
		} else if key == "" {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_KEY_IN_DATA, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_KEY_IN_DATA)
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
		err := validateTokenForStrict(request.Tokens, request.Data, options.TokenMode, tag)
		if err != nil {
			return err
		}
	case common.ENABLE_STRICT:
		if request.Tokens == nil {
			logger.Error(fmt.Sprintf(logs.TOKENS_REQUIRED_WITH_BYOT, tag, common.ENABLE_STRICT))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NO_TOKENS_WITH_BYOT, options.TokenMode))
		}
		err := validateTokenForStrict(request.Tokens, request.Data, options.TokenMode, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateFileUploadRequest validates the required fields of FileUploadRequest.
func ValidateFileUploadRequest(req common.FileUploadRequest) *skyflowError.SkyflowError {
	tag := constants.REQUEST_UPLOAD_FILE

	if strings.TrimSpace(req.Table) == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TABLE_KEY_ERROR)
	}

	if strings.TrimSpace(req.ColumnName) == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_COLUMN_NAME, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.COLUMN_NAME_KEY_ERROR)
	}

	// At least one file source must be provided
	if strings.TrimSpace(req.FilePath) == "" && strings.TrimSpace(req.Base64) == "" && req.FileObject == (os.File{}) {
		logger.Error(fmt.Sprintf(logs.MISSING_FILE_SOURCE_IN_UPLOAD_FILE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.MISSING_FILE_SOURCE_IN_UPLOAD_FILE)
	}

	// Validate FilePath
	if strings.TrimSpace(req.FilePath) != "" {
		fileInfo, err := os.Stat(req.FilePath)
		if err != nil || fileInfo.IsDir() {
			logger.Error(fmt.Sprintf(logs.INVALID_FILE_PATH, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_FILE_PATH)
		}
	}

	// Validate Base64
	if strings.TrimSpace(req.Base64) != "" {
		if strings.TrimSpace(req.FileName) == "" {
			logger.Error(logs.FILE_NAME_MUST_BE_PROVIDED_WITH_FILE_OBJECT)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.FILE_NAME_MUST_BE_PROVIDED_WITH_FILE_OBJECT)
		}
		_, err := base64.StdEncoding.DecodeString(req.Base64)
		if err != nil {
			logger.Error(fmt.Sprintf(logs.INVALID_BASE64, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_BASE64)
		}
	}

	// Validate FileObject
	if req.FileObject != (os.File{}) {
		fileInfo, err := req.FileObject.Stat()
		if err != nil || fileInfo.IsDir() {
			logger.Error(logs.INVALID_FILE_OBJECT)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_FILE_OBJECT)
		}
	}

	return nil
}

// ValidateCustomHeaders checks that every key in headers is one of the
func ValidateCustomHeaders(headers map[common.CustomHeaderKey]string, tag string) *skyflowError.SkyflowError {
	if headers != nil && len(headers) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_REQUEST_HEADERS, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_HEADER)
	}
	allowedKeys := map[common.CustomHeaderKey]struct{}{
		common.SkyflowAccountId:   {},
		common.SkyflowAccountName: {},
		common.RequestIdHeader:    {},
	}
	for key := range headers {
		if _, ok := allowedKeys[key]; !ok {
			logger.Error(fmt.Sprintf(logs.INVALID_HEADER_KEY, tag, key))
			return skyflowError.NewSkyflowError(
				skyflowError.INVALID_INPUT_CODE,
				fmt.Sprintf(skyflowError.INVALID_HEADER_KEY, tag, key),
			)
		}
		if strings.TrimSpace(headers[key]) == "" {
			logger.Error(fmt.Sprintf(logs.EMPTY_OR_NULL_VALUE_IN_HEADERS, tag, key))
			return skyflowError.NewSkyflowError(
				skyflowError.INVALID_INPUT_CODE,
				fmt.Sprintf(skyflowError.EMPTY_OR_NULL_VALUE_IN_HEADERS, tag, key),
			)
		}
	}
	return nil
}

func isValidHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}

	if u.Scheme != constants.HTTP_PROTOCOL && u.Scheme != constants.HTTPS_PROTOCOL {
		return false
	}

	return u.Host != ""
}
