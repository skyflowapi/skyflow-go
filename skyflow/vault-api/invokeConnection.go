package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/skyflowapi/skyflow-go/commonutils"
	"github.com/skyflowapi/skyflow-go/commonutils/errors"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type InvokeConnectionApi struct {
	ConnectionConfig common.ConnectionConfig
	Token            string
}

func (InvokeConnectionApi *InvokeConnectionApi) doValidations() *errors.SkyflowError {
	if InvokeConnectionApi.ConnectionConfig.ConnectionURL == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), commonutils.EMPTY_CONNECTION_URL)
	} else if !isValidUrl(InvokeConnectionApi.ConnectionConfig.ConnectionURL) {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(commonutils.INVALID_CONNECTION_URL, InvokeConnectionApi.ConnectionConfig.ConnectionURL))
	}
	return nil
}

func (InvokeConnectionApi *InvokeConnectionApi) Post() (map[string]interface{}, *errors.SkyflowError) {

	validationError := InvokeConnectionApi.doValidations()
	if validationError != nil {
		return nil, validationError
	}
	requestUrl := InvokeConnectionApi.ConnectionConfig.ConnectionURL
	for index, value := range InvokeConnectionApi.ConnectionConfig.PathParams {
		requestUrl = strings.Replace(requestUrl, fmt.Sprintf("{%s}", index), value, -1)
	}
	requestBody, err := json.Marshal(InvokeConnectionApi.ConnectionConfig.RequestBody)
	if err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(commonutils.UNKNOWN_ERROR, err))
	}
	request, _ := http.NewRequest(
		InvokeConnectionApi.ConnectionConfig.MethodName.String(),
		requestUrl,
		strings.NewReader(string(requestBody)),
	)
	query := request.URL.Query()
	for index, value := range InvokeConnectionApi.ConnectionConfig.QueryParams {
		switch v := value.(type) {
		case int:
			query.Set(index, strconv.Itoa(v))
		case float64:
			query.Set(index, fmt.Sprintf("%f", v))
		case string:
			query.Set(index, v)
		case bool:
			query.Set(index, strconv.FormatBool(v))
			// default:
			// 	fmt.Printf("Invalid type, we dont allow these types")
		}
	}
	request.URL.RawQuery = query.Encode()
	request.Header.Set("X-Skyflow-Authorization", InvokeConnectionApi.Token)
	request.Header.Set("Content-Type", "application/json")
	for index, value := range InvokeConnectionApi.ConnectionConfig.RequestHeader {
		request.Header.Set(index, value)
	}
	fmt.Println((request.URL))
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(commonutils.SERVER_ERROR, err))
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(commonutils.UNKNOWN_ERROR, string(data)))
	}
	return result, nil
}
