package validation

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

// ValidateDeidentifyTextRequest validates the required fields of DeidentifyTextRequest.
func ValidateDeidentifyTextRequest(req common.DeidentifyTextRequest) *skyflowError.SkyflowError {
	if strings.TrimSpace(req.Text) == "" {
		logger.Error(fmt.Sprintf(logs.INVALID_TEXT_IN_DEIDENTIFY, constants.REQUEST_DEIDENTIFY_TEXT))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TEXT_IN_DEIDENTIFY)
	}

	// Validate entities
	if len(req.Entities) > 0 {
		if err := validateEntities(req.Entities, constants.ENTITY_TYPE_TEXT); err != nil {
			return err
		}
	}

	// Validate token format
	if req.TokenFormat.DefaultType != "" {
		if _, err := vaultapis.NewTokenTypeMappingDefaultFromString(string(req.TokenFormat.DefaultType)); err != nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TOKEN_FORMAT)
		}
	}

	// Validate EntityOnly tokens
	if len(req.TokenFormat.EntityOnly) > 0 {
		if err := validateEntities(req.TokenFormat.EntityOnly, constants.ENTITY_TYPE_ENTITY_ONLY); err != nil {
			return err
		}
	}

	// Validate VaultToken entities
	if len(req.TokenFormat.VaultToken) > 0 {
		if err := validateEntities(req.TokenFormat.VaultToken, constants.ENTITY_TYPE_VAULT_TOKEN); err != nil {
			return err
		}
	}

	if err := validateTransformations(req.Transformations); err != nil {
		return err
	}

	return nil
}

// ValidateReidentifyTextRequest validates the required fields of ReidentifyTextRequest.
func ValidateReidentifyTextRequest(req common.ReidentifyTextRequest) *skyflowError.SkyflowError {
	if strings.TrimSpace(req.Text) == "" {
		logger.Error(fmt.Sprintf(logs.INVALID_TEXT_IN_REIDENTIFY, constants.REQUEST_REIDENTIFY_TEXT))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TEXT_IN_REIDENTIFY)
	}

	// Validate RedactedEntities
	if len(req.RedactedEntities) > 0 {
		if err := validateEntities(req.RedactedEntities, constants.ENTITY_TYPE_REDACTED); err != nil {
			return err
		}
	}

	// Validate MaskedEntities
	if len(req.MaskedEntities) > 0 {
		if err := validateEntities(req.MaskedEntities, constants.ENTITY_TYPE_MASKED); err != nil {
			return err
		}
	}

	// Validate PlainTextEntities
	if len(req.PlainTextEntities) > 0 {
		if err := validateEntities(req.PlainTextEntities, constants.ENTITY_TYPE_PLAIN_TEXT); err != nil {
			return err
		}
	}

	return nil
}

func validateEntities(entities []common.DetectEntities, entityType string) *skyflowError.SkyflowError {
	for _, entity := range entities {
		// add entity type validation
		if entityType == constants.ENTITY_TYPE_REDACTED {
			if _, err := vaultapis.NewFormatRedactedItemFromString(string(entity)); err != nil {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_ENTITY_TYPE, entity))
			}
		} else if entityType == constants.ENTITY_TYPE_MASKED {
			if _, err := vaultapis.NewFormatMaskedItemFromString(string(entity)); err != nil {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_ENTITY_TYPE, entity))
			}
		} else if entityType == constants.ENTITY_TYPE_PLAIN_TEXT {
			if _, err := vaultapis.NewFormatPlaintextItemFromString(string(entity)); err != nil {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_ENTITY_TYPE, entity))
			}
		} else if entityType == constants.ENTITY_TYPE_TEXT {
			if _, err := vaultapis.NewDeidentifyStringRequestEntityTypesItemFromString(string(entity)); err != nil {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_ENTITY_TYPE, entity))
			}
		} else if entityType == constants.ENTITY_TYPE_ENTITY_ONLY {
			if _, err := vaultapis.NewTokenTypeMappingEntityOnlyItemFromString(string(entity)); err != nil {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_ENTITY_TYPE, entity))
			}
		} else if entityType == constants.ENTITY_TYPE_VAULT_TOKEN {
			if _, err := vaultapis.NewTokenTypeMappingVaultTokenItemFromString(string(entity)); err != nil {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_ENTITY_TYPE, entity))
			}
		} else if entityType == constants.ENTITY_TYPE_ENTITY_UNIQUE_CTR {
			if _, err := vaultapis.NewTokenTypeMappingEntityUnqCounterItemFromString(string(entity)); err != nil {
				return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.INVALID_ENTITY_TYPE, entity))
			}
		}
	}
	return nil
}

func validateTransformations(transformations common.Transformations) *skyflowError.SkyflowError {
	shift := transformations.ShiftDates
	if shift.MinDays > 0 || shift.MaxDays > 0 || len(shift.Entities) > 0 {
		// Days must be strictly positive
		if shift.MinDays <= 0 || shift.MaxDays <= 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_SHIFT_DATES)
		}
		// Range validation
		if shift.MaxDays < shift.MinDays {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_DATE_TRANSFORMATION_RANGE)
		}
	}
	return nil
}

// ValidateGetDetectRunRequest validates the required fields of GetDetectRunRequest.
func ValidateGetDetectRunRequest(req common.GetDetectRunRequest) *skyflowError.SkyflowError {
	if strings.TrimSpace(req.RunId) == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_RUN_ID, constants.REQUEST_GET_DETECT_RUN))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_RUN_ID)
	}
	return nil
}

// ValidateDeidentifyFileRequest validates the required fields of DeidentifyFileRequest.
func ValidateDeidentifyFileRequest(req common.DeidentifyFileRequest) *skyflowError.SkyflowError {
		tag := constants.REQUEST_DEIDENTIFY_FILE
	// Validate if file or filepath is provided
	if req.File.File == nil && req.File.FilePath == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_FILE_AND_FILE_PATH_IN_DEIDENTIFY_FILE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_FILE_AND_FILE_PATH_IN_DEIDENTIFY_FILE)
	}

	// Check if both file and filepath are provided
	if req.File.File != nil && req.File.FilePath != "" {
		logger.Error(fmt.Sprintf(logs.BOTH_FILE_AND_FILE_PATH_PROVIDED, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.BOTH_FILE_AND_FILE_PATH_PROVIDED)
	}

	// Validate filepath if provided
	if req.File.FilePath != "" && strings.TrimSpace(req.File.FilePath) == "" {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_FILE_PATH)
	}

	if err := ValidateFilePermissions(req.File.FilePath, req.File.File); err != nil {
		return err
	}

	// Optional fields validation
	// Validate pixel density
	if req.PixelDensity != 0 && req.PixelDensity <= 0 {
		logger.Error(fmt.Sprintf(logs.INVALID_PIXEL_DENSITY_TO_DEIDENTIFY_FILE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_PIXEL_DENSITY)
	}

	if req.MaskingMethod != "" {
		switch req.MaskingMethod {
		case common.BLACKBOX, common.BLUR:
		default:
			logger.Error(fmt.Sprintf(logs.INVALID_MASKING_METHOD, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_MASKING_METHOD)
		}
	}

	if req.OutputTranscription != "" {
		switch req.OutputTranscription {
		case common.DIARIZED_TRANSCRIPTION, common.MEDICAL_DIARIZED_TRANSCRIPTION, common.MEDICAL_TRANSCRIPTION, common.PLAINTEXT_TRANSCRIPTION, common.TRANSCRIPTION:
		default:
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_OUTPUT_TRANSCRIPTION)
		}
	}

	// Validate max resolution
	if req.MaxResolution != 0 && req.MaxResolution <= 0 {
		logger.Error(fmt.Sprintf(logs.INVALID_MAX_RESOLUTION, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_MAX_RESOLUTION)
	}

	// Validate entities
	if len(req.Entities) > 0 {
		if err := validateEntities(req.Entities, constants.ENTITY_TYPE_ENTITIES); err != nil {
			return err
		}
	}

	if err := validateTransformations(req.Transformations); err != nil {
		return err
	}

	// Validate token format
	tokenFormat := req.TokenFormat
	if len(tokenFormat.VaultToken) > 0 {
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.VAULT_TOKEN_FORMAT_IS_NOT_ALLOWED_FOR_DEIDENTIFY_FILES)
	}

	if req.TokenFormat.DefaultType != "" {
		if _, err := vaultapis.NewTokenTypeMappingDefaultFromString(string(req.TokenFormat.DefaultType)); err != nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_TOKEN_FORMAT)
		}
	}

	if len(req.TokenFormat.EntityOnly) > 0 {
		if err := validateEntities(req.TokenFormat.EntityOnly, constants.REQUEST_ENTITY_ONLY); err != nil {
			return err
		}
	}

	if len(req.TokenFormat.EntityUniqueCounter) > 0 {
		if err := validateEntities(req.TokenFormat.EntityUniqueCounter, constants.ENTITY_TYPE_ENTITY_UNIQUE_CTR); err != nil {
			return err
		}
	}

	// Validate output directory if provided
	if req.OutputDirectory != "" {
		fileInfo, err := os.Stat(req.OutputDirectory)
		if os.IsNotExist(err) || !fileInfo.IsDir() {
			logger.Error(fmt.Sprintf(logs.OUTPUT_DIRECTORY_NOT_FOUND, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.OUTPUT_DIRECTORY_NOT_FOUND)
		}

		// Check directory permissions
		if err := checkDirWritePermission(req.OutputDirectory); err != nil {
			logger.Error(fmt.Sprintf(logs.INVALID_PERMISSIONS_FOR_OUTPUT_DIRECTORY, tag))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_PERMISSION)
		}
	}

	// Validate wait time
	if req.WaitTime != 0 {
		if req.WaitTime <= 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_WAIT_TIME)
		}
		if req.WaitTime > 64 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.WAIT_TIME_EXCEEDS_LIMIT)
		}
	}

	return nil
}

// Helper function to check directory write permission
func checkDirWritePermission(dir string) error {
	// Try to create a temporary file
	tempFile := filepath.Join(dir, constants.PERMISSION_CHECK_FILE)
	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(tempFile)
	return nil
}

func ValidateFilePermissions(filePath string, file *os.File) *skyflowError.SkyflowError {
	var info os.FileInfo
	var err error

	if filePath != "" {
		// Path-based check
		info, err = os.Stat(filePath)
		if os.IsNotExist(err) {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.FILE_NOT_FOUND_TO_DEIDENTIFY, filePath))
		}
		if err != nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.UNABLE_TO_STAT_FILE_TO_DEIDENTIFY, filePath))
		}

		// Ensure it's a regular file
		if !info.Mode().IsRegular() {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NOT_REGULAR_FILE_TO_DEIDENTIFY, filePath))
		}

		// Ensure file is not empty
		if info.Size() == 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.EMPTY_FILE_TO_DEIDENTIFY, filePath))
		}

		// Try to open file
		f, err := os.Open(filePath)
		if err != nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.FILE_NOT_READABLE_TO_DEIDENTIFY, filePath))
		}
		defer f.Close()

		return nil
	}

	if file != nil {
		// File-based check
		info, err = file.Stat()
		if err != nil {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.UNABLE_TO_STAT_FILE_TO_DEIDENTIFY, file.Name()))
		}
		if !info.Mode().IsRegular() {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.NOT_REGULAR_FILE_TO_DEIDENTIFY, file.Name()))
		}
		if info.Size() == 0 {
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, fmt.Sprintf(skyflowError.EMPTY_FILE_TO_DEIDENTIFY, filePath))
		}
		return nil
	}

	return nil

}

func ValidateInsertRequest(request common.InsertRequest, options common.InsertOptions) *skyflowError.SkyflowError {
	// Validate table
	tag := constants.REQUEST_INSERT
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
	if len(tokens) != len(values) - 1 && mode == common.ENABLE_STRICT {
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
	}
	if vaultConfig.BaseVaultURL == "" {
		if vaultConfig.ClusterId == "" {
			logger.Error(logs.CLUSTER_ID_IS_REQUIRED)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_CLUSTER_ID)
		}
	} else {
		// Parse the URL
		isValidHTTPURL := isValidHTTPURL(vaultConfig.BaseVaultURL)
		if !isValidHTTPURL {
			logger.Error(logs.VAULT_URL_IS_INVALID)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_VAULT_URL)
		}
	}
	return nil
}
func ValidateUpdateVaultConfig(vaultConfig common.VaultConfig) *skyflowError.SkyflowError {
	// Validate VaultId
	if vaultConfig.VaultId == "" {
		logger.Error(logs.VAULT_ID_IS_REQUIRED)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.INVALID_VAULT_ID)
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
		if len(credentials.ApiKey) != constants.MAX_REQUEST_SIZE || !strings.Contains(credentials.ApiKey, constants.API_KEY_PREFIX) {
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
		if len(request.Body) == 0 {
			logger.Error(fmt.Sprintf(logs.EMPTY_REQUEST_BODY, constants.REQUEST_INVOKE_CONNECTION))
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_REQUEST_BODY)
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
	skyflowId, _ := helpers.GetSkyflowID(request.Data)
	if request.Table == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_TABLE, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_TABLE)
	} else if skyflowId == "" {
		logger.Error(logs.INVALID_SKYFLOW_ID_IN_UPDATE)
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_ID_IN_UPDATE)
	}
	if request.Data == nil || len(request.Data) == 0 {
		logger.Error(fmt.Sprintf(logs.EMPTY_DATA, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.EMPTY_DATA)
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

	if strings.TrimSpace(req.SkyflowId) == "" {
		logger.Error(fmt.Sprintf(logs.EMPTY_SKYFLOW_ID, tag))
		return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.SKYFLOW_ID_KEY_ERROR)
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
