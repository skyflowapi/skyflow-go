package controller

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	constants "github.com/skyflowapi/skyflow-go/v2/internal/constants"
	vaultapis "github.com/skyflowapi/skyflow-go/v2/internal/generated"
	files "github.com/skyflowapi/skyflow-go/v2/internal/generated/files"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	text "github.com/skyflowapi/skyflow-go/v2/internal/generated/strings"
	"github.com/skyflowapi/skyflow-go/v2/internal/helpers"
	"github.com/skyflowapi/skyflow-go/v2/internal/validation"
	"github.com/skyflowapi/skyflow-go/v2/serviceaccount"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
	logs "github.com/skyflowapi/skyflow-go/v2/utils/messages"
)

type DetectController struct {
	Config         common.VaultConfig
	Loglevel       *logger.LogLevel
	Token          string
	ApiKey         string
	TextApiClient  text.Client
	FilesApiClient files.Client
}

var CreateDetectRequestClientFunc = CreateDetectRequestClient

var SetBearerTokenForDetectControllerFunc = setBearerTokenForDetectController

// CreateRequestClient initializes the API client with the appropriate authorization header.
func CreateDetectRequestClient(v *DetectController) *skyflowError.SkyflowError {
	token := ""
	if v.Config.Credentials.ApiKey != "" {
		v.ApiKey = v.Config.Credentials.ApiKey
		logger.Info(logs.USING_API_KEY)
	} else if v.Config.Credentials.Token != "" {
		if serviceaccount.IsExpired(v.Config.Credentials.Token) {
			logger.Error(logs.BEARER_TOKEN_EXPIRED)
			return skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, skyflowError.TOKEN_EXPIRED)
		} else {
			logger.Info(logs.USING_BEARER_TOKEN)
			v.Token = v.Config.Credentials.Token
		}
	} else {
		err := setBearerTokenForDetectController(v)
		if err != nil {
			return err
		}
	}
	if v.ApiKey != "" {
		token = v.ApiKey
	} else if v.Token != "" {
		token = v.Token
	}

	var header http.Header
	header = http.Header{}
	header.Set(constants.SDK_METRICS_HEADER_KEY, helpers.CreateJsonMetadata())

	client := text.NewClient(option.WithBaseURL(GetURLWithEnv(v.Config.Env, v.Config.ClusterId)),
		option.WithToken(token),
		option.WithHTTPHeader(header),
		option.WithMaxAttempts(1),
	)

	v.TextApiClient = *client

	clientFiles := files.NewClient(option.WithBaseURL(GetURLWithEnv(v.Config.Env, v.Config.ClusterId)),
		option.WithToken(token),
		option.WithHTTPHeader(header),
		option.WithMaxAttempts(1),
	)

	v.FilesApiClient = *clientFiles

	return nil
}

// SetBearerTokenForDetectController checks and updates the token if necessary.
func setBearerTokenForDetectController(v *DetectController) *skyflowError.SkyflowError {
	// Validate token or generate a new one if expired or not set.
	if v.Token == "" || serviceaccount.IsExpired(v.Token) {
		logger.Info(logs.GENERATE_BEARER_TOKEN_TRIGGERED)
		token, err := GenerateToken(v.Config.Credentials)
		if err != nil {
			return err
		}
		v.Token = *token
	} else {
		logger.Info(logs.REUSE_BEARER_TOKEN)
	}
	return nil
}

func CreateDeidentifyTextRequest(request common.DeidentifyTextRequest, config common.VaultConfig) (*vaultapis.DeidentifyStringRequest, *skyflowError.SkyflowError) {
	payload := vaultapis.DeidentifyStringRequest{
		VaultId: config.VaultId,
	}

	// text
	if request.Text != "" {
		payload.Text = request.Text
	}

	// allowRegexList
	if len(request.AllowRegexList) > 0 {
		allowRegex := vaultapis.AllowRegex{}
		allowRegex = request.AllowRegexList
		payload.AllowRegex = &allowRegex
	}

	// Entities
	if len(request.Entities) > 0 {
		entities := vaultapis.EntityTypes{}
		for _, entity := range request.Entities {
			entityType, err := vaultapis.NewEntityTypeFromString(string(entity))
			if err != nil {
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid detect entity: "+string(entity))
			}
			entities = append(entities, entityType)
		}
		payload.EntityTypes = &entities
	}

	// TokenFormat
	if request.TokenFormat.DefaultType != "" {
		tokenFormat, err := vaultapis.NewTokenTypeDefaultFromString(string(request.TokenFormat.DefaultType))
		if err != nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid token format: "+string(request.TokenFormat.DefaultType))
		}
		payload.TokenType = &vaultapis.TokenType{
			Default: &tokenFormat,
		}
	}

	if len(request.TokenFormat.EntityOnly) > 0 || len(request.TokenFormat.VaultToken) > 0 {
		if payload.TokenType == nil {
			payload.TokenType = &vaultapis.TokenType{}
		}
	}

	if len(request.TokenFormat.EntityOnly) > 0 {
		entityOnly := []vaultapis.EntityType{}
		for _, entity := range request.TokenFormat.EntityOnly {
			entityType, err := vaultapis.NewEntityTypeFromString(string(entity))
			if err != nil {
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid entity only type: "+string(entity))
			}
			entityOnly = append(entityOnly, entityType)
		}
		payload.TokenType.EntityOnly = entityOnly
	}

	if len(request.TokenFormat.VaultToken) > 0 {
		vaultToken := []vaultapis.EntityType{}
		for _, entity := range request.TokenFormat.VaultToken {
			entityType, err := vaultapis.NewEntityTypeFromString(string(entity))
			if err != nil {
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid entity type: "+string(entity))
			}
			vaultToken = append(vaultToken, entityType)
		}
		payload.TokenType.VaultToken = vaultToken
	}

	// RestrictRegexList
	if len(request.RestrictRegexList) > 0 {
		restrictRegex := vaultapis.RestrictRegex{}
		restrictRegex = request.RestrictRegexList
		payload.RestrictRegex = &restrictRegex
	}

	// transformations
	if len(request.Transformations.ShiftDates.Entities) > 0 {
		entities := make([]vaultapis.TransformationsShiftDatesEntityTypesItem, len(request.Transformations.ShiftDates.Entities))
		for i, v := range request.Transformations.ShiftDates.Entities {
			entities[i] = vaultapis.TransformationsShiftDatesEntityTypesItem(v)
		}
		if payload.Transformations == nil {
			payload.Transformations = &vaultapis.Transformations{}
		}
		payload.Transformations.ShiftDates = &vaultapis.TransformationsShiftDates{
			EntityTypes: entities,
		}
	}

	if payload.Transformations != nil && payload.Transformations.ShiftDates != nil {
		if request.Transformations.ShiftDates.MaxDays != 0 {
			payload.Transformations.ShiftDates.MaxDays = &request.Transformations.ShiftDates.MaxDays
		}
		if request.Transformations.ShiftDates.MinDays != 0 {
			payload.Transformations.ShiftDates.MinDays = &request.Transformations.ShiftDates.MinDays
		}
	}
	return &payload, nil
}

func CreateReidentifyTextRequest(request common.ReidentifyTextRequest, config common.VaultConfig) (*vaultapis.ReidentifyStringRequest, *skyflowError.SkyflowError) {
	payload := vaultapis.ReidentifyStringRequest{
		VaultId: config.VaultId,
		Format:  &vaultapis.ReidentifyStringRequestFormat{},
	}

	// text
	if request.Text != "" {
		payload.Text = request.Text
	}

	// RedactedEntities
	if len(request.RedactedEntities) > 0 {
		redactedEntities, err := helpers.ValidateAndCreateEntityTypes(request.RedactedEntities)
		if err != nil {
			return nil, err
		}
		payload.Format.Redacted = redactedEntities
	}

	// MaskedEntities
	if len(request.MaskedEntities) > 0 {
		maskedEntities, err := helpers.ValidateAndCreateEntityTypes(request.MaskedEntities)
		if err != nil {
			return nil, err
		}
		payload.Format.Masked = maskedEntities
	}

	// PlainTextEntities
	if len(request.PlainTextEntities) > 0 {
		plainTextEntities, err := helpers.ValidateAndCreateEntityTypes(request.PlainTextEntities)
		if err != nil {
			return nil, err
		}
		payload.Format.Plaintext = plainTextEntities
	}

	return &payload, nil
}

func createTextFileRequest(request *common.DeidentifyFileRequest, base64Content, vaultID string) *vaultapis.DeidentifyTextRequest {
	return &vaultapis.DeidentifyTextRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifyTextRequestFile{
			Base64: base64Content,
		},
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		AllowRegex:      createAllowRegex(request.AllowRegexList),
		RestrictRegex:   createRestrictRegex(request.RestrictRegexList),
		Transformations: createTransformations(request.Transformations),
	}
}

func createImageRequest(request *common.DeidentifyFileRequest, base64Content, vaultId, fileExt string) *vaultapis.DeidentifyImageRequest {
	return &vaultapis.DeidentifyImageRequest{
		VaultId: vaultId,
		File: &vaultapis.DeidentifyImageRequestFile{
			Base64:     base64Content,
			DataFormat: vaultapis.DeidentifyImageRequestFileDataFormat(fileExt),
		},
		OutputProcessedImage: &request.OutputProcessedImage,
		OutputOcrText:        &request.OutputOcrText,
		MaskingMethod:        createMaskingMethod(request.MaskingMethod),
		EntityTypes:          createEntityTypes(request.Entities),
		TokenType:            createTokenType(request.TokenFormat),
		AllowRegex:           createAllowRegex(request.AllowRegexList),
		RestrictRegex:        createRestrictRegex(request.RestrictRegexList),
		Transformations:      createTransformations(request.Transformations),
	}
}

func createPdfRequest(request *common.DeidentifyFileRequest, base64Content, vaultID string) *vaultapis.DeidentifyPdfRequest {
	return &vaultapis.DeidentifyPdfRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifyPdfRequestFile{
			Base64: base64Content,
		},
		Density:         IntPtr(int(request.PixelDensity)),
		MaxResolution:   IntPtr(int(request.MaxResolution)),
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		AllowRegex:      createAllowRegex(request.AllowRegexList),
		RestrictRegex:   createRestrictRegex(request.RestrictRegexList),
		Transformations: createTransformations(request.Transformations),
	}
}

func createPresentationRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyPresentationRequest {
	return &vaultapis.DeidentifyPresentationRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifyPresentationRequestFile{
			Base64:     base64Content,
			DataFormat: vaultapis.DeidentifyPresentationRequestFileDataFormat(fileExt),
		},
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		AllowRegex:      createAllowRegex(request.AllowRegexList),
		RestrictRegex:   createRestrictRegex(request.RestrictRegexList),
		Transformations: createTransformations(request.Transformations),
	}
}

func createSpreadsheetRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifySpreadsheetRequest {
	return &vaultapis.DeidentifySpreadsheetRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifySpreadsheetRequestFile{
			Base64:     base64Content,
			DataFormat: vaultapis.DeidentifySpreadsheetRequestFileDataFormat(fileExt),
		},
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		AllowRegex:      createAllowRegex(request.AllowRegexList),
		RestrictRegex:   createRestrictRegex(request.RestrictRegexList),
		Transformations: createTransformations(request.Transformations),
	}
}

func createDocumentRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyDocumentRequest {
	return &vaultapis.DeidentifyDocumentRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifyDocumentRequestFile{
			Base64:     base64Content,
			DataFormat: vaultapis.DeidentifyDocumentRequestFileDataFormat(fileExt),
		},
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		AllowRegex:      createAllowRegex(request.AllowRegexList),
		RestrictRegex:   createRestrictRegex(request.RestrictRegexList),
		Transformations: createTransformations(request.Transformations),
	}
}

func createStructuredTextRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyStructuredTextRequest {
	return &vaultapis.DeidentifyStructuredTextRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifyStructuredTextRequestFile{
			Base64:     base64Content,
			DataFormat: vaultapis.DeidentifyStructuredTextRequestFileDataFormat(fileExt),
		},
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		AllowRegex:      createAllowRegex(request.AllowRegexList),
		RestrictRegex:   createRestrictRegex(request.RestrictRegexList),
		Transformations: createTransformations(request.Transformations),
	}
}

func createAudioRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyAudioRequest {
	req := &vaultapis.DeidentifyAudioRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifyAudioRequestFile{
			Base64:     base64Content,
			DataFormat: vaultapis.DeidentifyAudioRequestFileDataFormat(fileExt),
		},
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		Transformations: createTransformations(request.Transformations),
	}

	if request.OutputProcessedAudio {
		req.OutputProcessedAudio = &request.OutputProcessedAudio
	}

	if request.OutputTranscription != "" {
		trans := vaultapis.DeidentifyAudioRequestOutputTranscription(request.OutputTranscription)
		req.OutputTranscription = &trans
	}

	if request.Bleep.Frequency != 0 || request.Bleep.Gain != 0 ||
		request.Bleep.StartPadding != 0 || request.Bleep.StopPadding != 0 {
		if request.Bleep.Frequency != 0 {
			req.BleepFrequency = &request.Bleep.Frequency
		}
		if request.Bleep.Gain != 0 {
			req.BleepGain = &request.Bleep.Gain
		}
		if request.Bleep.StartPadding != 0 {
			req.BleepStartPadding = &request.Bleep.StartPadding
		}
		if request.Bleep.StopPadding != 0 {
			req.BleepStopPadding = &request.Bleep.StopPadding
		}
	}

	if len(request.AllowRegexList) > 0 {
		req.AllowRegex = createAllowRegex(request.AllowRegexList)
	}
	if len(request.RestrictRegexList) > 0 {
		req.RestrictRegex = createRestrictRegex(request.RestrictRegexList)
	}

	return req
}

func createGenericFileRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExtension string) *vaultapis.DeidentifyFileRequest {
	return &vaultapis.DeidentifyFileRequest{
		VaultId: vaultID,
		File: &vaultapis.DeidentifyFileRequestFile{
			Base64:     base64Content,
			DataFormat: vaultapis.DeidentifyFileRequestFileDataFormat(strings.ToUpper(fileExtension)),
		},
		EntityTypes:     createEntityTypes(request.Entities),
		TokenType:       createTokenType(request.TokenFormat),
		AllowRegex:      createAllowRegex(request.AllowRegexList),
		RestrictRegex:   createRestrictRegex(request.RestrictRegexList),
		Transformations: createTransformations(request.Transformations),
	}
}

func createEntityTypes(entities []common.DetectEntities) *vaultapis.EntityTypes {
	if len(entities) == 0 {
		return nil
	}

	entityTypes := make(vaultapis.EntityTypes, len(entities))
	for i, entity := range entities {
		entityType := vaultapis.EntityType(entity)
		entityTypes[i] = entityType
	}
	return &entityTypes
}

func createTokenType(format common.TokenFormat) *vaultapis.TokenTypeWithoutVault {
	if len(format.EntityOnly) == 0 && len(format.EntityUniqueCounter) == 0 {
		return nil
	}

	tokenType := &vaultapis.TokenTypeWithoutVault{}

	if len(format.EntityOnly) > 0 {
		tokenEntities, err := helpers.ValidateAndCreateEntityTypes(format.EntityOnly)
		if err == nil {
			tokenType.EntityOnly = tokenEntities
		}
	}

	if len(format.EntityUniqueCounter) > 0 {
		tokenEntities, err := helpers.ValidateAndCreateEntityTypes(format.EntityUniqueCounter)
		if err == nil {
			tokenType.EntityUnqCounter = tokenEntities
		}
	}

	return tokenType
}

func createAllowRegex(regex []string) *vaultapis.AllowRegex {
	if len(regex) == 0 {
		return nil
	}
	allowRegex := vaultapis.AllowRegex(regex)
	return &allowRegex
}

func createRestrictRegex(regex []string) *vaultapis.RestrictRegex {
	if len(regex) == 0 {
		return nil
	}
	restrictRegex := vaultapis.RestrictRegex(regex)
	return &restrictRegex
}

func createMaskingMethod(method common.MaskingMethod) *vaultapis.DeidentifyImageRequestMaskingMethod {
	if method == "" {
		return vaultapis.DeidentifyImageRequestMaskingMethodBlur.Ptr()
	}
	maskMethod := vaultapis.DeidentifyImageRequestMaskingMethod(method)
	return &maskMethod
}

func isZeroDateTransformation(dateTransformation common.DateTransformation) bool {
	return dateTransformation.MaxDays == 0 && dateTransformation.MinDays == 0 && len(dateTransformation.Entities) == 0
}

func createTransformations(transformations common.Transformations) *vaultapis.Transformations {
	if isZeroDateTransformation(transformations.ShiftDates) {
		return nil
	}

	shiftDates := &vaultapis.TransformationsShiftDates{
		MaxDays: &transformations.ShiftDates.MaxDays,
		MinDays: &transformations.ShiftDates.MinDays,
	}

	if n := len(transformations.ShiftDates.Entities); n > 0 {
		shiftDatesEntityTypesItem := make([]vaultapis.TransformationsShiftDatesEntityTypesItem, 0, n)
		for _, e := range transformations.ShiftDates.Entities {
			shiftDatesEntityTypesItem = append(shiftDatesEntityTypesItem, vaultapis.TransformationsShiftDatesEntityTypesItem(e))
		}
		shiftDates.EntityTypes = shiftDatesEntityTypesItem
	}

	return &vaultapis.Transformations{
		ShiftDates: shiftDates,
	}
}

func IntPtr(v int) *int {
	return &v
}

// DeidentifyText handles the de-identification of text using the DetectController.
func (d *DetectController) DeidentifyText(ctx context.Context, request common.DeidentifyTextRequest) (*common.DeidentifyTextResponse, *skyflowError.SkyflowError) {
	// Log the start of the operation
	logger.Info(logs.DEIDENTIFY_TEXT_TRIGGERED)
	logger.Info(logs.VALIDATE_DEIDENTIFY_TEXT_REQUEST)

	// Validate the deidentify text request
	if err := validation.ValidateDeidentifyTextRequest(request); err != nil {
		return nil, err
	}

	// Create the API client if needed
	if err := CreateDetectRequestClientFunc(d); err != nil {
		logger.Error(logs.BEARER_TOKEN_REJECTED, err)
		return nil, err
	}

	// Ensure the bearer token is valid
	if err := SetBearerTokenForDetectControllerFunc(d); err != nil {
		logger.Error(logs.BEARER_TOKEN_REJECTED, err)
		return nil, err
	}

	// Prepare the API request payload
	apiRequest, err := CreateDeidentifyTextRequest(request, d.Config)
	if err != nil {
		return nil, err
	}

	// Call the API
	response, apiError := d.TextApiClient.WithRawResponse.DeidentifyString(ctx, apiRequest)
	if apiError != nil {
		logger.Error(logs.DEIDENTIFY_TEXT_REQUEST_FAILED)
		return nil, skyflowError.SkyflowErrorApi(apiError)
	}

	// Check for empty response
	if response == nil || response.Body == nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Deidentify text response is empty")
	}

	// Map the API response to the common.DeidentifyTextResponse struct
	deidentifiedTextResponse := common.DeidentifyTextResponse{
		ProcessedText:  response.Body.ProcessedText,
		WordCount:      response.Body.WordCount,
		CharacterCount: response.Body.CharacterCount,
	}

	// Map entities if present
	if response.Body.Entities != nil {
		for _, entity := range response.Body.Entities {
			entityInfo := common.EntityInfo{
				Token:  *entity.Token,
				Value:  *entity.Value,
				Entity: *entity.EntityType,
				Scores: entity.EntityScores,
				TextIndex: common.TextIndex{
					StartIndex: *entity.Location.StartIndex,
					EndIndex:   *entity.Location.EndIndex,
				},
				ProcessedIndex: common.TextIndex{
					StartIndex: *entity.Location.StartIndexProcessed,
					EndIndex:   *entity.Location.EndIndexProcessed,
				},
			}
			deidentifiedTextResponse.Entities = append(deidentifiedTextResponse.Entities, entityInfo)
		}
	}

	logger.Info(logs.DEIDENTIFY_TEXT_SUCCESS)
	return &deidentifiedTextResponse, nil
}

// ReidentifyText handles the re-identification of text using the DetectController.
func (d *DetectController) ReidentifyText(ctx context.Context, request common.ReidentifyTextRequest) (*common.ReidentifyTextResponse, *skyflowError.SkyflowError) {
	// Log the start of the operation
	logger.Info(logs.REIDENTIFY_TEXT_TRIGGERED)
	logger.Info(logs.VALIDATE_REIDENTIFY_TEXT_REQUEST)

	// Validate the deidentify text request
	if err := validation.ValidateReidentifyTextRequest(request); err != nil {
		return nil, err
	}

	// Create the API client if needed
	if err := CreateDetectRequestClientFunc(d); err != nil {
		return nil, err
	}

	// Ensure the bearer token is valid
	if err := SetBearerTokenForDetectControllerFunc(d); err != nil {
		logger.Error(logs.BEARER_TOKEN_REJECTED, err)
		return nil, err
	}

	// Prepare the API request payload
	apiRequest, err := CreateReidentifyTextRequest(request, d.Config)
	if err != nil {
		return nil, err
	}

	// Call the API
	response, apiError := d.TextApiClient.WithRawResponse.ReidentifyString(ctx, apiRequest)
	if apiError != nil {
		logger.Error(logs.REIDENTIFY_TEXT_REQUEST_FAILED)
		return nil, skyflowError.SkyflowErrorApi(apiError)
	}

	// Map the API response to the common.ReidentifyTextResponse struct
	reidentifiedTextResponse := common.ReidentifyTextResponse{}

	if body := response.Body; body != nil && body.Text != nil {
		reidentifiedTextResponse.ProcessedText = *body.Text
	}

	logger.Info(logs.REIDENTIFY_TEXT_SUCCESS)
	return &reidentifiedTextResponse, nil
}

// DeidentifyFile handles the de-identification of files using the DetectController.
func (d *DetectController) DeidentifyFile(ctx context.Context, request common.DeidentifyFileRequest) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
	// Log the start of the operation
	logger.Info(logs.DEIDENTIFY_FILE_TRIGGERED)
	logger.Info(logs.VALIDATE_DEIDENTIFY_FILE_REQUEST)

	// Validate the deidentify file request
	if err := validation.ValidateDeidentifyFileRequest(request); err != nil {
		return nil, err
	}

	// Create the API client if needed
	if err := CreateDetectRequestClientFunc(d); err != nil {
		return nil, err
	}

	// Ensure the bearer token is valid
	if err := SetBearerTokenForDetectControllerFunc(d); err != nil {
		logger.Error(logs.BEARER_TOKEN_REJECTED, err)
		return nil, err
	}

	var fileContent []byte
	var fileName, fileExtension string

	if request.FileInput.FilePath != "" {
		fileContent, _ = os.ReadFile(request.FileInput.FilePath)
		fileName = filepath.Base(request.FileInput.FilePath)
		fileExtension = strings.ToLower(filepath.Ext(fileName))

	} else if request.FileInput.File != nil {
		// File provided
		fileContent, _ = io.ReadAll(request.FileInput.File)
		fileName = filepath.Base(request.FileInput.File.Name())
		fileExtension = strings.ToLower(filepath.Ext(fileName))
	}
	fileExtension = strings.TrimPrefix(fileExtension, ".")

	// Convert file to base64
	base64Content := base64.StdEncoding.EncodeToString(fileContent)

	// Process file based on type
	apiResponse, apiErr := d.processFileByType(ctx, fileExtension, base64Content, &request)
	if apiErr != nil {
		logger.Error(logs.DEIDENTIFY_FILE_REQUEST_FAILED)
		return nil, skyflowError.SkyflowErrorApi(apiErr)
	}

	// Poll for results
	response, pollErr := d.pollForResults(ctx, apiResponse.RunId, request.WaitTime)

	if pollErr != nil {
		logger.Error(logs.POLLING_FOR_RESULTS_FAILED)
		return nil, skyflowError.SkyflowErrorApi(pollErr)
	}

	// Handle successful response
	if strings.EqualFold(response.Status, string(common.SUCCESS)) && response.FileBase64 != "" {
		decodedBytes, err := base64.StdEncoding.DecodeString(response.FileBase64)
		if err != nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.SERVER, logs.FAILED_TO_DECODE_PROCESSED_FILE)
		}

		outputFileName := "processed-" + fileName
		if request.OutputDirectory != "" {
			outputFileName = filepath.Join(request.OutputDirectory, outputFileName)
		}

		if err := os.WriteFile(outputFileName, decodedBytes, 0644); err != nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.SERVER, skyflowError.FAILED_TO_SAVED_PROCESSED_FILE)
		}
	}

	logger.Info(logs.DEIDENTIFY_FILE_SUCCESS)
	return response, nil
}

func (d *DetectController) processFileByType(ctx context.Context, fileExtension, base64Content string, request *common.DeidentifyFileRequest) (*vaultapis.DeidentifyFileResponse, error) {
	var apiResponse *vaultapis.DeidentifyFileResponse
	var apiErr error

	switch fileExtension {
	case "txt":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyText(ctx, createTextFileRequest(request, base64Content, d.Config.VaultId))
	case "mp3", "wav":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyAudio(ctx, createAudioRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "pdf":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyPdf(ctx, createPdfRequest(request, base64Content, d.Config.VaultId))
	case "jpg", "jpeg", "png", "bmp", "tif", "tiff":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyImage(ctx, createImageRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "ppt", "pptx":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyPresentation(ctx, createPresentationRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "csv", "xls", "xlsx":
		apiResponse, apiErr = d.FilesApiClient.DeidentifySpreadsheet(ctx, createSpreadsheetRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "doc", "docx":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyDocument(ctx, createDocumentRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "json", "xml":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyStructuredText(ctx, createStructuredTextRequest(request, base64Content, d.Config.VaultId, fileExtension))
	default:
		apiResponse, apiErr = d.FilesApiClient.DeidentifyFile(ctx, createGenericFileRequest(request, base64Content, d.Config.VaultId, fileExtension))
	}
	if apiErr != nil {
		logger.Error(logs.DEIDENTIFY_FILE_REQUEST_FAILED)
		return nil, apiErr
	}

	return apiResponse, nil
}

func (d *DetectController) pollForResults(ctx context.Context, runID string, maxWaitTime int) (*common.DeidentifyFileResponse, error) {
	currentWaitTime := 1
	if maxWaitTime == 0 {
		maxWaitTime = 64
	}

	getRunRequest := vaultapis.GetRunRequest{
		VaultId: d.Config.VaultId,
	}

	for {
		response, err := d.FilesApiClient.WithRawResponse.GetRun(ctx, runID, &getRunRequest)

		if err != nil {
			logger.Error(logs.GET_DETECT_RUN_REQUEST_FAILED)
			return nil, skyflowError.SkyflowErrorApi(err)
		}

		if response == nil || response.Body == nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.SERVER, "Empty response received")
		}

		if strings.EqualFold(string(response.Body.Status), string(common.IN_PROGRESS)) {
			if currentWaitTime >= maxWaitTime {
				return &common.DeidentifyFileResponse{
					RunId:  runID,
					Status: string(common.IN_PROGRESS),
				}, nil
			}

			nextWaitTime := currentWaitTime * 2
			var waitTime int
			if nextWaitTime >= maxWaitTime {
				waitTime = maxWaitTime - currentWaitTime
				currentWaitTime = maxWaitTime
			} else {
				waitTime = nextWaitTime
				currentWaitTime = nextWaitTime
			}

			time.Sleep(time.Duration(waitTime) * time.Second)
			continue
		}

		if strings.EqualFold(string(response.Body.Status), string(common.SUCCESS)) || strings.EqualFold(string(response.Body.Status), string(common.FAILED)) {
			return parseDeidentifyFileResponse(response.Body, runID)
		}
	}
}

func parseDeidentifyFileResponse(response *vaultapis.DeidentifyStatusResponse, runID string) (*common.DeidentifyFileResponse, error) {
	if response == nil {
		return nil, errors.New(string(skyflowError.SERVER) + ": Empty response received")
	}

	fileResponse := &common.DeidentifyFileResponse{
		RunId:  runID,
		Status: string(response.Status),
	}

	if len(response.Output) > 0 {
		firstOutput := response.Output[0]
		if firstOutput != nil {
			// Handle processed file base64 and file extension
			if firstOutput.ProcessedFile != nil && firstOutput.ProcessedFileExtension != nil {
				decodedBytes, err := base64.StdEncoding.DecodeString(*firstOutput.ProcessedFile)
				if err != nil {
					return nil, errors.New(string(skyflowError.SERVER) + ": Failed to decode processed file")
				}
				fileResponse.File = common.FileInfo{
					Name:         "deidentified." + *firstOutput.ProcessedFileExtension,
					Size:         int64(len(decodedBytes)),
					Type:         "redacted_file",
					LastModified: time.Now().UnixMilli(),
				}
				fileResponse.FileBase64 = *firstOutput.ProcessedFile
				fileResponse.FileBase64 = *firstOutput.ProcessedFile
			}

			// Set file type and extension
			if firstOutput.ProcessedFileType != nil {
				fileResponse.Type = string(*firstOutput.ProcessedFileType)
			} else {
				fileResponse.Type = "UNKNOWN"
			}

			if firstOutput.ProcessedFileExtension != nil {
				fileResponse.Extension = *firstOutput.ProcessedFileExtension
			}
		}
	}

	wordCharCount := response.GetExtraProperties()["word_character_count"].(map[string]interface{})
	fileResponse.CharCount = int(wordCharCount["character_count"].(float64))
	fileResponse.WordCount = int(wordCharCount["word_count"].(float64))

	// Handle other metadata
	if response.Size != nil {
		fileResponse.SizeInKb = float64(*response.Size)
	}
	if response.Duration != nil {
		fileResponse.DurationInSeconds = float64(*response.Duration)
	}
	if response.Pages != nil {
		fileResponse.PageCount = *response.Pages
	}
	if response.Slides != nil {
		fileResponse.SlideCount = *response.Slides
	}

	// Get entities
	if len(response.Output) > 1 {
		secondOutput := response.Output[1]
		if secondOutput != nil {
			entityInfo := common.FileEntityInfo{}
			if secondOutput.ProcessedFile != nil {
				entityInfo.File = *secondOutput.ProcessedFile
			}
			if secondOutput.ProcessedFileType != nil {
				entityInfo.Type = string(*secondOutput.ProcessedFileType)
			}
			if secondOutput.ProcessedFileExtension != nil {
				entityInfo.Extension = *secondOutput.ProcessedFileExtension
			}
			fileResponse.Entities = append(fileResponse.Entities, entityInfo)
		}
	}

	return fileResponse, nil
}

func (d *DetectController) GetDetectRun(ctx context.Context, request common.GetDetectRunRequest) (*common.DeidentifyFileResponse, *skyflowError.SkyflowError) {
	// Log the start of the operation
	logger.Info(logs.GET_DETECT_RUN_TRIGGERED)
	logger.Info(logs.VALIDATE_GET_DETECT_RUN_REQUEST)

	// Validate the request
	if err := validation.ValidateGetDetectRunRequest(request); err != nil {
		return nil, err
	}

	// Create the API client if needed
	if err := CreateDetectRequestClientFunc(d); err != nil {
		return nil, err
	}

	// Ensure the bearer token is valid
	if err := SetBearerTokenForDetectControllerFunc(d); err != nil {
		logger.Error(logs.BEARER_TOKEN_REJECTED, err)
		return nil, err
	}

	response, err := d.FilesApiClient.WithRawResponse.GetRun(ctx, request.RunId, &vaultapis.GetRunRequest{
		VaultId: d.Config.VaultId,
	})

	if err != nil {
		logger.Error(logs.GET_DETECT_RUN_REQUEST_FAILED)
		return nil, skyflowError.SkyflowErrorApi(err)
	}

	if response == nil || response.Body == nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, " Empty response received")
	}

	if strings.EqualFold(string(response.Body.Status), string(common.IN_PROGRESS)) {
		return &common.DeidentifyFileResponse{
			RunId:  request.RunId,
			Status: string(common.IN_PROGRESS),
		}, nil
	}
	parsedResponse, err := parseDeidentifyFileResponse(response.Body, request.RunId)
	if err != nil {
		return nil, skyflowError.NewSkyflowError(skyflowError.SERVER, fmt.Sprintf("%v", err))
	}
	return parsedResponse, nil
}
