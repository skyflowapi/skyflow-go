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

var SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController

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
		err := SetBearerTokenForDetectController(v)
		if err != nil {
			return err
		}
	}
	if v.ApiKey != "" {
		token = v.ApiKey
	} else if v.Token != "" {
		token = v.Token
	}

	header := http.Header{}
	header.Set(constants.SDK_METRICS_HEADER_KEY, helpers.CreateJsonMetadata())

	client := text.NewClient(option.WithBaseURL(helpers.GetURLWithEnv(v.Config.Env, v.Config.ClusterId)),
		option.WithToken(token),
		option.WithHTTPHeader(header),
		option.WithMaxAttempts(1),
	)

	v.TextApiClient = *client

	clientFiles := files.NewClient(option.WithBaseURL(helpers.GetURLWithEnv(v.Config.Env, v.Config.ClusterId)),
		option.WithToken(token),
		option.WithHTTPHeader(header),
		option.WithMaxAttempts(1),
	)

	v.FilesApiClient = *clientFiles

	return nil
}

// SetBearerTokenForDetectController checks and updates the token if necessary.
func SetBearerTokenForDetectController(v *DetectController) *skyflowError.SkyflowError {
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
		Text:    request.Text,
	}

	if len(request.AllowRegexList) > 0 {
		payload.AllowRegex = request.AllowRegexList
	}

	// RestrictRegex
	if len(request.RestrictRegexList) > 0 {
		payload.RestrictRegex = request.RestrictRegexList
	}

	// Entities
	if len(request.Entities) > 0 {
		entityTypes := make([]vaultapis.DeidentifyStringRequestEntityTypesItem, len(request.Entities))
		for i, entity := range request.Entities {
			entityTypes[i] = vaultapis.DeidentifyStringRequestEntityTypesItem(entity)
		}
		payload.EntityTypes = entityTypes
	}

	// TokenFormat
	if request.TokenFormat.DefaultType != "" || len(request.TokenFormat.EntityOnly) > 0 || len(request.TokenFormat.VaultToken) > 0 {
		payload.TokenType = &vaultapis.TokenTypeMapping{}

		if request.TokenFormat.DefaultType != "" {
			tokenFormat := vaultapis.TokenTypeMappingDefault(request.TokenFormat.DefaultType)
			payload.TokenType.Default = &tokenFormat
		}

		if len(request.TokenFormat.EntityOnly) > 0 {
			entityOnly := make([]vaultapis.TokenTypeMappingEntityOnlyItem, len(request.TokenFormat.EntityOnly))
			for i, entity := range request.TokenFormat.EntityOnly {
				entityOnly[i] = vaultapis.TokenTypeMappingEntityOnlyItem(entity)
			}
			payload.TokenType.EntityOnly = entityOnly
		}

		if len(request.TokenFormat.VaultToken) > 0 {
			vaultToken := make([]vaultapis.TokenTypeMappingVaultTokenItem, len(request.TokenFormat.VaultToken))
			for i, entity := range request.TokenFormat.VaultToken {
				vaultToken[i] = vaultapis.TokenTypeMappingVaultTokenItem(entity)
			}
			payload.TokenType.VaultToken = vaultToken
		}
	}

	// Transformations
	if len(request.Transformations.ShiftDates.Entities) > 0 ||
		request.Transformations.ShiftDates.MinDays != 0 ||
		request.Transformations.ShiftDates.MaxDays != 0 {

		shiftDates := &vaultapis.ShiftDates{
			MinDays: &request.Transformations.ShiftDates.MinDays,
			MaxDays: &request.Transformations.ShiftDates.MaxDays,
		}

		if len(request.Transformations.ShiftDates.Entities) > 0 {
			entityTypes := make([]vaultapis.ShiftDatesEntityTypesItem, len(request.Transformations.ShiftDates.Entities))
			for i, entity := range request.Transformations.ShiftDates.Entities {
				entityTypes[i] = vaultapis.ShiftDatesEntityTypesItem(entity)
			}
			shiftDates.EntityTypes = entityTypes
		}

		payload.Transformations = &vaultapis.Transformations{
			ShiftDates: shiftDates,
		}
	}

	return &payload, nil
}

func CreateReidentifyTextRequest(request common.ReidentifyTextRequest, config common.VaultConfig) (*vaultapis.ReidentifyStringRequest, *skyflowError.SkyflowError) {
	payload := vaultapis.ReidentifyStringRequest{
		VaultId: &config.VaultId,
		Format:  &vaultapis.Format{},
	}

	// text
	if request.Text != "" {
		payload.Text = &request.Text
	}

	// RedactedEntities
	if len(request.RedactedEntities) > 0 {
		redactedEntities := CreateEntityTypes(request.RedactedEntities, "redacted").([]vaultapis.FormatRedactedItem)
		if len(redactedEntities) > 0 {
			payload.Format.Redacted = redactedEntities
		}
	}

	// MaskedEntities
	if len(request.MaskedEntities) > 0 {
		maskedEntities := CreateEntityTypes(request.MaskedEntities, "masked").([]vaultapis.FormatMaskedItem)
		if len(maskedEntities) > 0 {
			payload.Format.Masked = maskedEntities
		}
	}

	// PlainTextEntities
	if len(request.PlainTextEntities) > 0 {
		plainTextEntities := CreateEntityTypes(request.PlainTextEntities, "plaintext").([]vaultapis.FormatPlaintextItem)
		if len(plainTextEntities) > 0 {
			payload.Format.Plaintext = plainTextEntities
		}
	}

	return &payload, nil
}

func CreateTextFileRequest(request *common.DeidentifyFileRequest, base64Content, vaultID string) *vaultapis.DeidentifyFileRequestDeidentifyText {
	var entityTypes []vaultapis.DeidentifyFileRequestDeidentifyTextEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "text"); result != nil {
		entityTypes = result.([]vaultapis.DeidentifyFileRequestDeidentifyTextEntityTypesItem)
	}
	return &vaultapis.DeidentifyFileRequestDeidentifyText{
		VaultId: vaultID,
		File: &vaultapis.FileDataDeidentifyText{
			Base64: base64Content,
		},
		EntityTypes:     entityTypes,
		TokenType:       CreateTokenType(request.TokenFormat),
		AllowRegex:      CreateAllowRegex(request.AllowRegexList),
		RestrictRegex:   CreateRestrictRegex(request.RestrictRegexList),
		Transformations: CreateTransformations(request.Transformations),
	}
}

func CreateImageRequest(request *common.DeidentifyFileRequest, base64Content, vaultId, fileExt string) *vaultapis.DeidentifyFileImageRequestDeidentifyImage {
	var entityTypes []vaultapis.DeidentifyFileImageRequestDeidentifyImageEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "image"); result != nil {
		entityTypes = result.([]vaultapis.DeidentifyFileImageRequestDeidentifyImageEntityTypesItem)
	}
	return &vaultapis.DeidentifyFileImageRequestDeidentifyImage{
		VaultId: vaultId,
		File: &vaultapis.FileDataDeidentifyImage{
			Base64:     base64Content,
			DataFormat: vaultapis.FileDataDeidentifyImageDataFormat(fileExt),
		},
		OutputProcessedImage: &request.OutputProcessedImage,
		OutputOcrText:        &request.OutputOcrText,
		MaskingMethod:        CreateMaskingMethod(request.MaskingMethod),
		EntityTypes:          entityTypes,
		TokenType:            CreateTokenType(request.TokenFormat),
		AllowRegex:           CreateAllowRegex(request.AllowRegexList),
		RestrictRegex:        CreateRestrictRegex(request.RestrictRegexList),
	}
}

func CreatePdfRequest(request *common.DeidentifyFileRequest, base64Content, vaultID string) *vaultapis.DeidentifyFileDocumentPdfRequestDeidentifyPdf {
	var entityTypes []vaultapis.DeidentifyFileDocumentPdfRequestDeidentifyPdfEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "pdf"); result != nil {
		entityTypes = result.([]vaultapis.DeidentifyFileDocumentPdfRequestDeidentifyPdfEntityTypesItem)
	}
	return &vaultapis.DeidentifyFileDocumentPdfRequestDeidentifyPdf{
		VaultId: vaultID,
		File: &vaultapis.FileDataDeidentifyPdf{
			Base64: base64Content,
		},
		Density:       &request.PixelDensity,
		MaxResolution: &request.MaxResolution,
		EntityTypes:   entityTypes,
		TokenType:     CreateTokenType(request.TokenFormat),
		AllowRegex:    CreateAllowRegex(request.AllowRegexList),
		RestrictRegex: CreateRestrictRegex(request.RestrictRegexList),
	}
}

func CreatePresentationRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyFileRequestDeidentifyPresentation {
	var entityTypes []vaultapis.DeidentifyFileRequestDeidentifyPresentationEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "ppt"); result != nil {
		entityTypes = result.([]vaultapis.DeidentifyFileRequestDeidentifyPresentationEntityTypesItem)
	}
	return &vaultapis.DeidentifyFileRequestDeidentifyPresentation{
		VaultId: vaultID,
		File: &vaultapis.FileDataDeidentifyPresentation{
			Base64:     base64Content,
			DataFormat: vaultapis.FileDataDeidentifyPresentationDataFormat(fileExt),
		},
		EntityTypes:   entityTypes,
		TokenType:     CreateTokenType(request.TokenFormat),
		AllowRegex:    CreateAllowRegex(request.AllowRegexList),
		RestrictRegex: CreateRestrictRegex(request.RestrictRegexList),
	}
}

func CreateSpreadsheetRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyFileRequestDeidentifySpreadsheet {
	var entityTypes []vaultapis.DeidentifyFileRequestDeidentifySpreadsheetEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "spread").([]vaultapis.DeidentifyFileRequestDeidentifySpreadsheetEntityTypesItem); result != nil {
		entityTypes = result
	}
	return &vaultapis.DeidentifyFileRequestDeidentifySpreadsheet{
		VaultId: vaultID,
		File: &vaultapis.FileDataDeidentifySpreadsheet{
			Base64:     base64Content,
			DataFormat: vaultapis.FileDataDeidentifySpreadsheetDataFormat(fileExt),
		},
		EntityTypes:   entityTypes,
		TokenType:     CreateTokenType(request.TokenFormat),
		AllowRegex:    CreateAllowRegex(request.AllowRegexList),
		RestrictRegex: CreateRestrictRegex(request.RestrictRegexList),
	}
}

func CreateDocumentRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyFileRequestDeidentifyDocument {
	var entityTypes []vaultapis.DeidentifyFileRequestDeidentifyDocumentEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "document").([]vaultapis.DeidentifyFileRequestDeidentifyDocumentEntityTypesItem); result != nil {
		entityTypes = result
	}
	return &vaultapis.DeidentifyFileRequestDeidentifyDocument{
		VaultId: vaultID,
		File: &vaultapis.FileDataDeidentifyDocument{
			Base64:     base64Content,
			DataFormat: vaultapis.FileDataDeidentifyDocumentDataFormat(fileExt),
		},
		EntityTypes:   entityTypes,
		TokenType:     CreateTokenType(request.TokenFormat),
		AllowRegex:    CreateAllowRegex(request.AllowRegexList),
		RestrictRegex: CreateRestrictRegex(request.RestrictRegexList),
	}
}

func CreateStructuredTextRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyFileRequestDeidentifyStructuredText {
	var entityTypes []vaultapis.DeidentifyFileRequestDeidentifyStructuredTextEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "structured").([]vaultapis.DeidentifyFileRequestDeidentifyStructuredTextEntityTypesItem); result != nil {
		entityTypes = result
	}
	return &vaultapis.DeidentifyFileRequestDeidentifyStructuredText{
		VaultId: vaultID,
		File: &vaultapis.FileDataDeidentifyStructuredText{
			Base64:     base64Content,
			DataFormat: vaultapis.FileDataDeidentifyStructuredTextDataFormat(fileExt),
		},
		EntityTypes:     entityTypes,
		TokenType:       CreateTokenType(request.TokenFormat),
		AllowRegex:      CreateAllowRegex(request.AllowRegexList),
		RestrictRegex:   CreateRestrictRegex(request.RestrictRegexList),
		Transformations: CreateTransformations(request.Transformations),
	}
}

func CreateAudioRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExt string) *vaultapis.DeidentifyFileAudioRequestDeidentifyAudio {
	var entityTypes []vaultapis.DeidentifyFileAudioRequestDeidentifyAudioEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "audio").([]vaultapis.DeidentifyFileAudioRequestDeidentifyAudioEntityTypesItem); result != nil {
		entityTypes = result
	}
	req := &vaultapis.DeidentifyFileAudioRequestDeidentifyAudio{
		VaultId: vaultID,
		File: &vaultapis.FileDataDeidentifyAudio{
			Base64:     base64Content,
			DataFormat: vaultapis.FileDataDeidentifyAudioDataFormat(fileExt),
		},
		EntityTypes:          entityTypes,
		TokenType:            CreateTokenType(request.TokenFormat),
		AllowRegex:           CreateAllowRegex(request.AllowRegexList),
		RestrictRegex:        CreateRestrictRegex(request.RestrictRegexList),
		Transformations:      CreateTransformations(request.Transformations),
		BleepGain:            &request.Bleep.Gain,
		BleepFrequency:       &request.Bleep.Frequency,
		BleepStartPadding:    &request.Bleep.StartPadding,
		BleepStopPadding:     &request.Bleep.StopPadding,
		OutputProcessedAudio: &request.OutputProcessedAudio,
	}

	if request.OutputTranscription != "" {
		trans := vaultapis.DeidentifyFileAudioRequestDeidentifyAudioOutputTranscription(request.OutputTranscription)
		req.OutputTranscription = &trans
	}

	return req
}

func CreateGenericFileRequest(request *common.DeidentifyFileRequest, base64Content, vaultID, fileExtension string) *vaultapis.DeidentifyFileRequest {
	var entityTypes []vaultapis.DeidentifyFileRequestEntityTypesItem
	if result := CreateEntityTypesRef(request.Entities, "generic").([]vaultapis.DeidentifyFileRequestEntityTypesItem); result != nil {
		entityTypes = result
	}
	return &vaultapis.DeidentifyFileRequest{
		VaultId: vaultID,
		File: &vaultapis.FileData{
			Base64:     base64Content,
			DataFormat: vaultapis.FileDataDataFormat(strings.ToUpper(fileExtension)),
		},
		EntityTypes:     entityTypes,
		TokenType:       CreateTokenType(request.TokenFormat),
		AllowRegex:      CreateAllowRegex(request.AllowRegexList),
		RestrictRegex:   CreateRestrictRegex(request.RestrictRegexList),
		Transformations: CreateTransformations(request.Transformations),
	}
}

func CreateEntityTypesRef(entities []common.DetectEntities, dataType string) any {
	if len(entities) == 0 {
		return nil
	}

	switch strings.ToLower(dataType) {

	case "text":
		entityTypes := make([]vaultapis.DeidentifyFileRequestDeidentifyTextEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileRequestDeidentifyTextEntityTypesItem(e)
		}
		return entityTypes

	case "image":
		entityTypes := make([]vaultapis.DeidentifyFileImageRequestDeidentifyImageEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileImageRequestDeidentifyImageEntityTypesItem(e)
		}
		return entityTypes

	case "pdf":
		entityTypes := make([]vaultapis.DeidentifyFileDocumentPdfRequestDeidentifyPdfEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileDocumentPdfRequestDeidentifyPdfEntityTypesItem(e)
		}
		return entityTypes

	case "ppt":
		entityTypes := make([]vaultapis.DeidentifyFileRequestDeidentifyPresentationEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileRequestDeidentifyPresentationEntityTypesItem(e)
		}
		return entityTypes

	case "spread":
		entityTypes := make([]vaultapis.DeidentifyFileRequestDeidentifySpreadsheetEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileRequestDeidentifySpreadsheetEntityTypesItem(e)
		}
		return entityTypes

	case "document":
		entityTypes := make([]vaultapis.DeidentifyFileRequestDeidentifyDocumentEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileRequestDeidentifyDocumentEntityTypesItem(e)
		}
		return entityTypes

	case "structured":
		entityTypes := make([]vaultapis.DeidentifyFileRequestDeidentifyStructuredTextEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileRequestDeidentifyStructuredTextEntityTypesItem(e)
		}
		return entityTypes

	case "audio":
		entityTypes := make([]vaultapis.DeidentifyFileAudioRequestDeidentifyAudioEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileAudioRequestDeidentifyAudioEntityTypesItem(e)
		}
		return entityTypes

	case "generic":
		entityTypes := make([]vaultapis.DeidentifyFileRequestEntityTypesItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.DeidentifyFileRequestEntityTypesItem(e)
		}
		return entityTypes

	default:
		return nil
	}
}

func CreateEntityTypes(entities []common.DetectEntities, entityType string) any {
	if len(entities) == 0 {
		return nil
	}

	switch strings.ToLower(entityType) {
	case "redacted":
		entityTypes := make([]vaultapis.FormatRedactedItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.FormatRedactedItem(e)
		}
		return entityTypes

	case "masked":
		entityTypes := make([]vaultapis.FormatMaskedItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.FormatMaskedItem(e)
		}
		return entityTypes

	case "plaintext":
		entityTypes := make([]vaultapis.FormatPlaintextItem, len(entities))
		for i, e := range entities {
			entityTypes[i] = vaultapis.FormatPlaintextItem(e)
		}
		return entityTypes

	default:
		return nil
	}
}

func CreateTokenType(format common.TokenFormat) *vaultapis.TokenTypeMapping {
	// If nothing is set, return nil
	if len(format.VaultToken) == 0 &&
		len(format.EntityOnly) == 0 &&
		len(format.EntityUniqueCounter) == 0 &&
		format.DefaultType == "" {
		return nil
	}

	tokenType := &vaultapis.TokenTypeMapping{}

	if len(format.VaultToken) > 0 {
		vaultTokens := make([]vaultapis.TokenTypeMappingVaultTokenItem, len(format.VaultToken))
		for i, e := range format.VaultToken {
			vaultTokens[i] = vaultapis.TokenTypeMappingVaultTokenItem(e)
		}
		tokenType.VaultToken = vaultTokens
	}

	if len(format.EntityOnly) > 0 {
		entityOnly := make([]vaultapis.TokenTypeMappingEntityOnlyItem, len(format.EntityOnly))
		for i, e := range format.EntityOnly {
			entityOnly[i] = vaultapis.TokenTypeMappingEntityOnlyItem(e)
		}
		tokenType.EntityOnly = entityOnly
	}

	if len(format.EntityUniqueCounter) > 0 {
		entityUnqCounter := make([]vaultapis.TokenTypeMappingEntityUnqCounterItem, len(format.EntityUniqueCounter))
		for i, e := range format.EntityUniqueCounter {
			entityUnqCounter[i] = vaultapis.TokenTypeMappingEntityUnqCounterItem(e)
		}
		tokenType.EntityUnqCounter = entityUnqCounter
	}

	if format.DefaultType != "" {
		defaultType := vaultapis.TokenTypeMappingDefault(format.DefaultType)
		tokenType.Default = &defaultType
	}

	return tokenType
}

func CreateAllowRegex(regex []string) []string {
	if len(regex) == 0 {
		return nil
	}
	return regex
}

func CreateRestrictRegex(regex []string) []string {
	if len(regex) == 0 {
		return nil
	}
	return regex
}

func CreateMaskingMethod(method common.MaskingMethod) *vaultapis.DeidentifyFileImageRequestDeidentifyImageMaskingMethod {
	if method == "" {
		return nil
	}
	maskMethod := vaultapis.DeidentifyFileImageRequestDeidentifyImageMaskingMethod(method)
	return &maskMethod
}

func isZeroDateTransformation(dateTransformation common.DateTransformation) bool {
	return dateTransformation.MaxDays == 0 && dateTransformation.MinDays == 0 && len(dateTransformation.Entities) == 0
}

func CreateTransformations(transformations common.Transformations) *vaultapis.Transformations {
	if isZeroDateTransformation(transformations.ShiftDates) {
		return nil
	}

	shiftDates := &vaultapis.ShiftDates{
		MinDays: &transformations.ShiftDates.MinDays,
		MaxDays: &transformations.ShiftDates.MaxDays,
	}

	if len(transformations.ShiftDates.Entities) > 0 {
		entityTypes := make([]vaultapis.ShiftDatesEntityTypesItem, len(transformations.ShiftDates.Entities))
		for i, e := range transformations.ShiftDates.Entities {
			entityTypes[i] = vaultapis.ShiftDatesEntityTypesItem(e)
		}
		shiftDates.EntityTypes = entityTypes
	}

	return &vaultapis.Transformations{
		ShiftDates: shiftDates,
	}
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
		header, _ := helpers.GetHeader(apiError)
		return nil, skyflowError.SkyflowErrorApi(apiError, header)
	}

	deidentifiedTextResponse := common.DeidentifyTextResponse{}

	// Check for empty response
	if response == nil || response.Body == nil {
		return &deidentifiedTextResponse, nil
	}
	// Map the API response to the common.DeidentifyTextResponse struct
	if response.Body.ProcessedText != nil {
		deidentifiedTextResponse.ProcessedText = *response.Body.ProcessedText
	}
	if response.Body.WordCount != nil {
		deidentifiedTextResponse.WordCount = *response.Body.WordCount
	}
	if response.Body.CharacterCount != nil {
		deidentifiedTextResponse.CharCount = *response.Body.CharacterCount
	}
	// Map entities if present
	if response.Body.Entities != nil {
		for _, entity := range response.Body.Entities {
			entityInfo := common.EntityInfo{}
			if entity != nil {
				if entity.EntityType != nil {
					entityInfo.Entity = *entity.EntityType
				}
				if entity.Value != nil {
					entityInfo.Value = *entity.Value
				}
				if entity.Token != nil {
					entityInfo.Token = *entity.Token
				}
				if entity.Location != nil {
					if entity.Location.StartIndex != nil && entity.Location.EndIndex != nil {
						entityInfo.TextIndex = common.TextIndex{
							Start: *entity.Location.StartIndex,
							End:   *entity.Location.EndIndex,
						}
					}
					if entity.Location.StartIndexProcessed != nil && entity.Location.EndIndexProcessed != nil {
						entityInfo.ProcessedIndex = common.TextIndex{
							Start: *entity.Location.StartIndexProcessed,
							End:   *entity.Location.EndIndexProcessed,
						}
					}
				}
				if entity.EntityScores != nil {
					entityInfo.Scores = entity.EntityScores
				}
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
		header, _ := helpers.GetHeader(apiError)
		return nil, skyflowError.SkyflowErrorApi(apiError, header)
	}

	// Map the API response to the common.ReidentifyTextResponse struct
	reidentifiedTextResponse := common.ReidentifyTextResponse{}

	if body := response.Body; body != nil && body.Text != "" {
		reidentifiedTextResponse.ProcessedText = body.Text
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

	if request.File.FilePath != "" {
		fileContent, _ = os.ReadFile(request.File.FilePath)
		fileName = filepath.Base(request.File.FilePath)
		fileExtension = strings.ToLower(filepath.Ext(fileName))

	} else if request.File.File != nil {
		// File provided
		fileContent, _ = io.ReadAll(request.File.File)
		fileName = filepath.Base(request.File.File.Name())
		fileExtension = strings.ToLower(filepath.Ext(fileName))
	}
	fileExtension = strings.TrimPrefix(fileExtension, ".")

	// Convert file to base64
	base64Content := base64.StdEncoding.EncodeToString(fileContent)

	// Process file based on type
	apiResponse, apiErr := d.processFileByType(ctx, fileExtension, base64Content, &request)
	if apiErr != nil {
		logger.Error(logs.DEIDENTIFY_FILE_REQUEST_FAILED)
		header, _ := helpers.GetHeader(apiErr)
		return nil, skyflowError.SkyflowErrorApi(apiErr, header)
	}

	// Poll for results
	pollResponse, pollErr := d.pollForResults(ctx, *apiResponse.RunId, request.WaitTime)

	if pollErr != nil {
		logger.Error(logs.POLLING_FOR_RESULTS_FAILED)
		header, _ := helpers.GetHeader(pollErr)
		return nil, skyflowError.SkyflowErrorApi(pollErr, header)
	}

	response := &common.DeidentifyFileResponse{}

	// Handle successful response
	if pollResponse.Status != nil && strings.EqualFold(string(*pollResponse.Status), string(common.SUCCESS)) {
		processDeidentifyFileResponse(pollResponse, request.OutputDirectory, fileName, strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	}

	response, _ = parseDeidentifyFileResponse(pollResponse, *apiResponse.RunId)

	logger.Info(logs.DEIDENTIFY_FILE_SUCCESS)
	return response, nil
}

func (d *DetectController) processFileByType(ctx context.Context, fileExtension, base64Content string, request *common.DeidentifyFileRequest) (*vaultapis.DeidentifyFileResponse, error) {
	var apiResponse *vaultapis.DeidentifyFileResponse
	var apiErr error

	switch fileExtension {
	case "txt":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyText(ctx, CreateTextFileRequest(request, base64Content, d.Config.VaultId))
	case "mp3", "wav":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyAudio(ctx, CreateAudioRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "pdf":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyPdf(ctx, CreatePdfRequest(request, base64Content, d.Config.VaultId))
	case "jpg", "jpeg", "png", "bmp", "tif", "tiff":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyImage(ctx, CreateImageRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "ppt", "pptx":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyPresentation(ctx, CreatePresentationRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "csv", "xls", "xlsx":
		apiResponse, apiErr = d.FilesApiClient.DeidentifySpreadsheet(ctx, CreateSpreadsheetRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "doc", "docx":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyDocument(ctx, CreateDocumentRequest(request, base64Content, d.Config.VaultId, fileExtension))
	case "json", "xml":
		apiResponse, apiErr = d.FilesApiClient.DeidentifyStructuredText(ctx, CreateStructuredTextRequest(request, base64Content, d.Config.VaultId, fileExtension))
	default:
		apiResponse, apiErr = d.FilesApiClient.DeidentifyFile(ctx, CreateGenericFileRequest(request, base64Content, d.Config.VaultId, fileExtension))
	}
	if apiErr != nil {
		logger.Error(logs.DEIDENTIFY_FILE_REQUEST_FAILED)
		return nil, apiErr
	}

	return apiResponse, nil
}

func (d *DetectController) pollForResults(ctx context.Context, runID string, maxWaitTime int) (*vaultapis.DetectRunsResponse, error) {
	currentWaitTime := 1
	if maxWaitTime == 0 {
		maxWaitTime = 64
	}

	getRunRequest := vaultapis.GetRunRequest{
		VaultId: &d.Config.VaultId,
	}

	for {
		response, err := d.FilesApiClient.WithRawResponse.GetRun(ctx, runID, &getRunRequest)

		if err != nil {
			logger.Error(logs.GET_DETECT_RUN_REQUEST_FAILED)
			header, _ := helpers.GetHeader(err)
			return nil, skyflowError.SkyflowErrorApi(err, header)
		}

		if response == nil || response.Body == nil {
			return nil, skyflowError.NewSkyflowError(skyflowError.SERVER, logs.EMPTY_DEIDENTIFY_FILE_RESPONSE)
		}

		if response.Body.Status != nil && strings.EqualFold(string(*response.Body.Status), string(common.IN_PROGRESS)) {
			if currentWaitTime >= maxWaitTime {
				status := vaultapis.DetectRunsResponseStatus(common.IN_PROGRESS)
				deidentifyStatusRes := &vaultapis.DetectRunsResponse{
					Status: &status,
				}
				return deidentifyStatusRes, nil
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

		return response.Body, nil
	}
}

func processDeidentifyFileResponse(data *vaultapis.DetectRunsResponse, outputDir, fileName, fileBaseName string) *skyflowError.SkyflowError {
	if data == nil || len(data.Output) == 0 {
		return skyflowError.NewSkyflowError(skyflowError.SERVER, logs.EMPTY_DEIDENTIFY_FILE_RESPONSE)
	}
	if outputDir == "" {
		return nil
	}

	deidentifyFilePrefix := "processed-"
	processedFile := data.Output[0].ProcessedFile
	if processedFile != nil {
		decodedBytes, err := base64.StdEncoding.DecodeString(string(*processedFile))
		if err != nil {
			return skyflowError.NewSkyflowError(skyflowError.SERVER, logs.FAILED_TO_DECODE_PROCESSED_FILE)
		}

		outputFileName := filepath.Join(outputDir, deidentifyFilePrefix+fileName)

		if err := os.WriteFile(outputFileName, decodedBytes, 0644); err != nil {
			return skyflowError.NewSkyflowError(skyflowError.SERVER, skyflowError.FAILED_TO_SAVED_PROCESSED_FILE)
		}
	}
	if len(data.Output) > 1 {
		for _, output := range data.Output[1:] {
			if output != nil && output.ProcessedFile != nil {
				outputFileName := fmt.Sprintf("%s%s.%s", deidentifyFilePrefix, fileBaseName, *output.ProcessedFileExtension)
				outputPath := filepath.Join(outputDir, outputFileName)

				decodedData, err := base64.StdEncoding.DecodeString(string(*output.ProcessedFile))
				if err != nil {
					return skyflowError.NewSkyflowError(skyflowError.SERVER, logs.FAILED_TO_DECODE_PROCESSED_FILE)
				}

				if err := os.WriteFile(outputPath, decodedData, 0644); err != nil {
					return skyflowError.NewSkyflowError(skyflowError.SERVER, skyflowError.FAILED_TO_SAVED_PROCESSED_FILE)
				}
			}
		}
	}
	return nil
}

func parseDeidentifyFileResponse(response *vaultapis.DetectRunsResponse, runID string) (*common.DeidentifyFileResponse, error) {
	if response == nil {
		return nil, errors.New(string(skyflowError.SERVER) + logs.EMPTY_DEIDENTIFY_FILE_RESPONSE)
	}

	fileResponse := &common.DeidentifyFileResponse{
		RunId:  runID,
		Status: string(*response.Status),
	}

	// In case of expired/invalid run id
	if len(response.Output) == 0 {
		if response.OutputType != nil {
			fileResponse.Type = string(*response.OutputType)
		}
		return fileResponse, nil
	}

	if len(response.Output) > 0 {
		firstOutput := response.Output[0]
		if firstOutput != nil {
			// Handle processed file base64 and file extension
			if firstOutput.ProcessedFile != nil && firstOutput.ProcessedFileExtension != nil {
				decodedBytes, err := base64.StdEncoding.DecodeString(*firstOutput.ProcessedFile)
				if err != nil {
					return nil, errors.New(string(skyflowError.SERVER) + logs.FAILED_TO_DECODE_PROCESSED_FILE)
				}
				fileResponse.File = common.FileInfo{
					Name:         "deidentified." + string(*firstOutput.ProcessedFileExtension),
					Size:         int64(len(decodedBytes)),
					Type:         "redacted_file",
					LastModified: time.Now().UnixMilli(),
				}
				fileResponse.FileBase64 = *firstOutput.ProcessedFile
			}

			// Set file type and extension
			if firstOutput.ProcessedFileType != nil {
				fileResponse.Type = string(*firstOutput.ProcessedFileType)
			} else {
				fileResponse.Type = "UNKNOWN"
			}

			if firstOutput.ProcessedFileExtension != nil {
				fileResponse.Extension = string(*firstOutput.ProcessedFileExtension)
			}
		}
	}
	if response.WordCharacterCount != nil {
		if response.WordCharacterCount.CharacterCount != nil {
			fileResponse.CharCount = *response.WordCharacterCount.CharacterCount
		}
		if response.WordCharacterCount.WordCount != nil {
			fileResponse.WordCount = *response.WordCharacterCount.WordCount
		}
	}

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
	for _, output := range response.Output[1:] {
		if output == nil || output.ProcessedFileType == nil {
			continue
		}

		entityInfo := common.FileEntityInfo{}
		if output.ProcessedFile != nil {
			entityInfo.File = *output.ProcessedFile
		}
		if output.ProcessedFileType != nil {
			entityInfo.Type = string(*output.ProcessedFileType)
		}
		if output.ProcessedFileExtension != nil {
			entityInfo.Extension = string(*output.ProcessedFileExtension)
		}
		fileResponse.Entities = append(fileResponse.Entities, entityInfo)
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
		VaultId: &d.Config.VaultId,
	})

	if err != nil {
		logger.Error(logs.GET_DETECT_RUN_REQUEST_FAILED)
		header, _ := helpers.GetHeader(err)
		return nil, skyflowError.SkyflowErrorApi(err, header)
	}

	deidentifyFileRes := common.DeidentifyFileResponse{}

	if response == nil || response.Body == nil {
		return &deidentifyFileRes, nil
	}

	if response.Body.Status != nil && strings.EqualFold(string(*response.Body.Status), string(common.IN_PROGRESS)) {
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
