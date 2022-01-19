package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()

	url := "www.google.com/card/cvv"

	path := make(map[string]interface{})

	path["card"] = "1234"
	path["cvv"] = "234"

	query := make(map[string]interface{})

	query["cvv"] = 456.1
	query["cc"] = true

	for index, value := range path {
		url = strings.Replace(url, index, value.(string), -1)
	}

	req := make(map[string]interface{})
	req["sam"] = 123
	req["xx"] = 456
	requestBody, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	request, _ := http.NewRequest(
		"POST",
		url,
		strings.NewReader(string(requestBody)),
	)
	query1 := request.URL.Query()
	for index, value := range query {
		switch v := value.(type) {
		case int:
			query1.Set(index, strconv.Itoa(v))
		case float64:
			query1.Set(index, fmt.Sprintf("%f", v))
		case string:
			query1.Set(index, v)
		case bool:
			query1.Set(index, strconv.FormatBool(v))
		default:
			fmt.Printf("Invalid type, we dont allow these types")
		}
	}
	request.URL.RawQuery = query1.Encode()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	fmt.Println(request.URL)

}
