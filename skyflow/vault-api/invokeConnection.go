package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/skyflowapi/skyflow-go/errors"
	"github.com/skyflowapi/skyflow-go/skyflow/common"
)

type InvokeConnectionApi struct {
	ConnectionConfig common.ConnectionConfig
	Token            string
}

func (InvokeConnectionApi *InvokeConnectionApi) doValidations() *errors.SkyflowError {
	if InvokeConnectionApi.ConnectionConfig.ConnectionURL == "" {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), errors.EMPTY_CONNECTION_URL)
	} else if !isValidUrl(InvokeConnectionApi.ConnectionConfig.ConnectionURL) {
		return errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(errors.INVALID_CONNECTION_URL, InvokeConnectionApi.ConnectionConfig.ConnectionURL))
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
		requestUrl = strings.Replace(requestUrl, index, value, -1)
	}
	requestBody, err := json.Marshal(InvokeConnectionApi.ConnectionConfig.RequestBody)
	if err != nil {
		return nil, errors.NewSkyflowError(errors.ErrorCodesEnum(errors.SdkErrorCode), fmt.Sprintf(errors.UNKNOWN_ERROR, err))
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
		default:
			fmt.Printf("Invalid type, we dont allow these types")
		}
	}
	request.URL.RawQuery = query.Encode()
	request.Header.Add("X-Skyflow-Authorization", InvokeConnectionApi.Token)
	for index, value := range InvokeConnectionApi.ConnectionConfig.RequestHeader {
		request.Header.Add(index, value)
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("error from server: ", err)
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, nil
	}
	return nil, nil
}
