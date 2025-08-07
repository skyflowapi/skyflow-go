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
	Config    common.DetectConfig
	Loglevel  *logger.LogLevel
	Token     string
	ApiKey    string
	TextApiClient text.Client
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
func CreateDeidentifyTextRequest(request common.DeidentifyTextRequest) (*vaultapis.DeidentifyStringRequest, *skyflowError.SkyflowError) {
	// Create the API request object
	payload := vaultapis.DeidentifyStringRequest{}
	// text
	if request.Text != "" {
		payload.Text = request.Text
	} 
	// allowRegexList
	if (request.AllowRegexList != nil && len(request.AllowRegexList) > 0) {
		allowRegex := vaultapis.AllowRegex{}
		allowRegex = request.AllowRegexList
		payload.AllowRegex = &allowRegex
	}
	// Entities
	if len(request.Entities) > 0 {
		entities := vaultapis.EntityTypes{}
		for _, entity := range request.Entities {
			entityStr := fmt.Sprintf("%v", entity)
			entityType, err := vaultapis.NewEntityTypeFromString(entityStr)
			if err != nil {
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid entity type: "+entityStr)
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
		payload.TokenType.Default = &tokenFormat
	}
	if len(request.TokenFormat.EntityOnly) > 0 {
		tokenFormat := []vaultapis.EntityType{}
		for _, entity := range request.TokenFormat.EntityOnly {
			entity, err := vaultapis.NewEntityTypeFromString(string(entity))
			if err != nil {
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid entity type: "+string(entity))
			}
			tokenFormat = append(tokenFormat, entity)
		}
		payload.TokenType.EntityOnly = tokenFormat
	}
	if request.TokenFormat.VaultToken != nil {
		tokenFormat := []vaultapis.EntityType{}
		for _, entity := range request.TokenFormat.VaultToken {
			entity, err := vaultapis.NewEntityTypeFromString(string(entity))
			if err != nil {
				return nil, skyflowError.NewSkyflowError(skyflowError.INVALID_INPUT_CODE, "Invalid entity type: "+string(entity))
			}
			tokenFormat = append(tokenFormat, entity)
		}
		payload.TokenType.VaultToken = tokenFormat
	}
	// RestrictRegexList
	if request.RestrictRegexList != nil && len(request.RestrictRegexList) > 0 {
		restrictRegex := vaultapis.RestrictRegex{}
		restrictRegex = request.RestrictRegexList
		payload.RestrictRegex = &restrictRegex
	}
	// transformations
	if request.Transformations.ShiftDates.Entities != nil && len(request.Transformations.ShiftDates.Entities) > 0 {
		entities := make([]vaultapis.TransformationsShiftDatesEntityTypesItem, len(request.Transformations.ShiftDates.Entities))
		for i, v := range request.Transformations.ShiftDates.Entities {
			entities[i] = vaultapis.TransformationsShiftDatesEntityTypesItem(v)
		}
		payload.Transformations.ShiftDates.EntityTypes = entities
	}
	if request.Transformations.ShiftDates.MaxDays != 0 {
		payload.Transformations.ShiftDates.MaxDays = &request.Transformations.ShiftDates.MaxDays
	}
	if request.Transformations.ShiftDates.MinDays != 0 {
		payload.Transformations.ShiftDates.MinDays = &request.Transformations.ShiftDates.MinDays
	}
	return &payload, nil
}

// // create client for DetectController
// // create deidentifyText request
// // process the response


