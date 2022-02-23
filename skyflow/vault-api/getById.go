package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	"github.com/skyflowapi/skyflow-go/commonutils/messages"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type GetByIdApi struct {
	Configuration common.Configuration
	Records       map[string]interface{}
	Token         string
}

var getByIdTag = "GetById"

func (g *GetByIdApi) Get() (map[string]interface{}, *errors.SkyflowError) {

	err := g.doValidations()
	if err != nil {
		return nil, err
	}
	jsonRecord, _ := json.Marshal(g.Records)
	var getByIdRecord common.GetByIdInput
	if err := json.Unmarshal(jsonRecord, &getByIdRecord); err != nil {
		logger.Error(fmt.Sprintf(messages.INVALID_RECORDS, getByIdTag))
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_RECORDS, getByIdTag))
	}
	res, err := g.doRequest(getByIdRecord)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (g *GetByIdApi) doValidations() *errors.SkyflowError {
	var err = isValidVaultDetails(g.Configuration)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf(messages.VALIDATE_GET_BY_ID_INPUT, getByIdTag))

	var totalRecords = g.Records["records"]
	if totalRecords == nil {
		logger.Error(fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, getByIdTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.RECORDS_KEY_NOT_FOUND, getByIdTag))
	}
	var recordsArray = (totalRecords).([]interface{})
	if len(recordsArray) == 0 {
		logger.Error(fmt.Sprintf(messages.EMPTY_RECORDS, getByIdTag))
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORDS, getByIdTag))
	}
	for _, record := range recordsArray {
		var singleRecord = (record).(map[string]interface{})
		var table = singleRecord["table"]
		var ids = singleRecord["ids"]
		var redaction = singleRecord["redaction"]
		if table == nil {
			logger.Error(fmt.Sprintf(messages.MISSING_TABLE, getByIdTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_TABLE, getByIdTag))
		} else if table == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_TABLE_NAME, getByIdTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TABLE_NAME, getByIdTag))
		} else if ids == nil {
			logger.Error(fmt.Sprintf(messages.MISSING_KEY_IDS, getByIdTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_KEY_IDS, getByIdTag))
		} else if ids == "" {
			logger.Error(fmt.Sprintf(messages.EMPTY_RECORD_IDS, getByIdTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getByIdTag))
		} else if redaction == nil {
			logger.Error(fmt.Sprintf(messages.MISSING_REDACTION, getByIdTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.MISSING_REDACTION, getByIdTag))
		} else if redaction != common.PLAIN_TEXT && redaction != common.DEFAULT && redaction != common.REDACTED && redaction != common.MASKED {
			logger.Error(fmt.Sprintf(messages.INVALID_REDACTION_TYPE, getByIdTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.INVALID_REDACTION_TYPE, getByIdTag))
		}
		idArray := (ids).([]interface{})
		if len(idArray) == 0 {
			logger.Error(fmt.Sprintf(messages.EMPTY_RECORD_IDS, getByIdTag))
			return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_RECORD_IDS, getByIdTag))
		}
		for index := range idArray {
			if idArray[index] == "" {
				logger.Error(fmt.Sprintf(messages.EMPTY_TOKEN_ID, getByIdTag))
				return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(messages.EMPTY_TOKEN_ID, getByIdTag))
			}
		}
	}
	return nil
}

func (g *GetByIdApi) doRequest(records common.GetByIdInput) (map[string]interface{}, *errors.SkyflowError) {

	var finalSuccess []interface{}
	var finalError []map[string]interface{}

	responseChannel := make(chan map[string]interface{})

	for i := 0; i < len(records.Records); i++ {
		logger.Info(fmt.Sprintf(messages.GETTING_RECORDS_BY_ID, getByIdTag, records.Records[i].Table))
		go func(i int, responseChannel chan map[string]interface{}) {
			singleRecord := records.Records[i]
			requestUrl := fmt.Sprintf("%s/v1/vaults/%s/%s", g.Configuration.VaultURL, g.Configuration.VaultID, singleRecord.Table)
			url1, err := url.Parse(requestUrl)
			v := url.Values{}
			for j := 0; j < len(singleRecord.Ids); j++ {
				v.Add("skyflow_ids", singleRecord.Ids[j])
			}
			v.Add("redaction", string(singleRecord.Redaction))
			url1.RawQuery = v.Encode()
			if err == nil {
				request, _ := http.NewRequest(
					"GET",
					url1.String(),
					strings.NewReader(""),
				)
				bearerToken := fmt.Sprintf("Bearer %s", g.Token)
				request.Header.Add("Authorization", bearerToken)

				res, err := Client.Do(request)
				var requestId = ""
				if res != nil {
					requestId = res.Header.Get("x-request-id")
				}
				if err != nil {
					logger.Error(fmt.Sprintf(messages.GET_RECORDS_BY_ID_FAILED, getByIdTag, common.AppendRequestId(singleRecord.Table, requestId)))
					var error = make(map[string]interface{})
					var errorObj = make(map[string]interface{})
					errorObj["code"] = "500"
					errorObj["description"] = common.AppendRequestId(fmt.Sprintf(messages.SERVER_ERROR, getByIdTag, err), requestId)
					error["error"] = errorObj
					error["ids"] = singleRecord.Ids
					responseChannel <- error
					return
				}
				data, _ := ioutil.ReadAll(res.Body)
				defer res.Body.Close()
				var result map[string]interface{}
				err = json.Unmarshal(data, &result)
				if err != nil {
					logger.Error(fmt.Sprintf(messages.GET_RECORDS_BY_ID_FAILED, getByIdTag, common.AppendRequestId(singleRecord.Table, requestId)))
					var error = make(map[string]interface{})
					var errorObj = make(map[string]interface{})
					errorObj["code"] = "500"
					errorObj["description"] = fmt.Sprintf(messages.UNKNOWN_ERROR, getByIdTag, common.AppendRequestId(string(data), requestId))
					error["error"] = errorObj
					error["ids"] = singleRecord.Ids
					responseChannel <- error
				} else {
					errorResult := result["error"]
					if errorResult != nil {
						logger.Error(fmt.Sprintf(messages.GET_RECORDS_BY_ID_FAILED, getByIdTag, common.AppendRequestId(singleRecord.Table, requestId)))
						var generatedError = (errorResult).(map[string]interface{})
						var error = make(map[string]interface{})
						var errorObj = make(map[string]interface{})
						errorObj["code"] = fmt.Sprintf("%v", (errorResult.(map[string]interface{}))["http_code"])
						errorObj["description"] = common.AppendRequestId(generatedError["message"].(string), requestId)
						error["error"] = errorObj
						error["ids"] = singleRecord.Ids
						responseChannel <- error

					} else {
						logger.Info(fmt.Sprintf(messages.GET_RECORDS_BY_ID_SUCCESS, getByIdTag, singleRecord.Table))
						responseObj := make(map[string]interface{})
						var responseArr []interface{}

						records := (result["records"]).([]interface{})
						for k := 0; k < len(records); k++ {
							new := make(map[string]interface{})
							single := (records[k]).(map[string]interface{})
							fields := (single["fields"]).(map[string]interface{})
							fields["id"] = fields["skyflow_id"]
							delete(fields, "skyflow_id")
							new["fields"] = fields
							new["table"] = singleRecord.Table
							responseArr = append(responseArr, new)
						}
						responseObj["records"] = responseArr
						responseChannel <- responseObj
					}

				}
			}
		}(i, responseChannel)
	}

	for i := 0; i < len(records.Records); i++ {
		response := <-responseChannel
		if _, found := response["error"]; found {
			finalError = append(finalError, response)
		} else {
			finalSuccess = append(finalSuccess, response["records"].([]interface{})...)
		}
	}

	var finalRecord = make(map[string]interface{})
	if finalSuccess != nil {
		finalRecord["records"] = finalSuccess
	}
	if finalError != nil {
		finalRecord["errors"] = finalError
	}
	return finalRecord, nil
}
