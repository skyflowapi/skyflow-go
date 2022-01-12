package vaultapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/skyflowapi/skyflow-go/errors"
)

type invokeConnectionApi struct {
	Configuration ConnectionConfig
	token         string
}

func (invokeConnectionApi *invokeConnectionApi) post() (map[string]interface{}, *errors.SkyflowError) {

	requestUrl := invokeConnectionApi.Configuration.connectionURL
	for index, value := range invokeConnectionApi.Configuration.pathParams {
		requestUrl = strings.Replace(requestUrl, index, value, -1)
	}
	requestBody, err := json.Marshal(invokeConnectionApi.Configuration.requestBody)
	request, _ := http.NewRequest(
		invokeConnectionApi.Configuration.methodName.String(),
		requestUrl,
		strings.NewReader(string(requestBody)),
	)
	query := request.URL.Query()
	for index, value := range invokeConnectionApi.Configuration.queryParams {
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
	request.Header.Add("X-Skyflow-Authorization", invokeConnectionApi.token)
	for index, value := range invokeConnectionApi.Configuration.requestHeader {
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
	return result, nil
}
