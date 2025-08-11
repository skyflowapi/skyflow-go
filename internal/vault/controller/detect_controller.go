package controller

import (
	"context"
	"fmt"
	"net/http"

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

	var header http.Header
	header = http.Header{}
	header.Set(constants.SDK_METRICS_HEADER_KEY, helpers.CreateJsonMetadata())

	client := text.NewClient(option.WithBaseURL(GetURLWithEnv(v.Config.Env, v.Config.ClusterId)),
		option.WithToken(token),
		option.WithHTTPHeader(header),
	)

	v.TextApiClient = *client

	clientFiles := files.NewClient(option.WithBaseURL(GetURLWithEnv(v.Config.Env, v.Config.ClusterId)),
		option.WithToken(token),
		option.WithHTTPHeader(header),
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
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid entity run: "+string(entity))
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
	if err := SetBearerTokenForDetectController(d); err != nil {
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
		logger.Error(fmt.Sprintf(logs.DEIDENTIFY_TEXT_REQUEST_FAILED, apiError.Error()))
		return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, apiError.Error())
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
