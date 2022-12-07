# Description
This go SDK is designed to help developers easily implement Skyflow into their go backend. 

[![CI](https://img.shields.io/static/v1?label=CI&message=passing&color=green?style=plastic&logo=github)](https://github.com/skyflowapi/skyflow-go/actions)
[![GitHub release](https://img.shields.io/github/v/release/skyflowapi/skyflow-go.svg)](https://github.com/skyflowapi/skyflow-go/releases)
[![License](https://img.shields.io/github/license/skyflowapi/skyflow-go)](https://github.com/skyflowapi/skyflow-go/blob/main/LICENSE)


# Table of Contents

- [Description](#description)
- [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
    - [Requirements](#requirements)
    - [Configuration](#configuration)
    - [Service Account Token Generation](#service-account-token-generation)
    - [Vault APIs](#vault-apis)
      - [Insert](#insert)
      - [Detokenize](#detokenize)
      - [Get By Id](#get-by-id)
    - [Invoke-connection](#invoke-connection)
    - [Logging](#logging)
  - [Reporting a Vulnerability](#reporting-a-vulnerability)


## Features

- Authentication with a Skyflow Service Account and generation of a bearer token
- Vault API operations to insert, retrieve and tokenize sensitive data
- Invoking connections to call downstream third party APIs without directly handling sensitive data

## Installation

### Requirements
- go 1.15 and above

### Configuration

Make sure your project is using Go Modules (it will have a go.mod file in its root if it already is):

```go
go mod init
```

Then, reference skyflow-go in a Go program with import:

```go
import (
  saUtil "github.com/skyflowapi/skyflow-go/serviceaccount/util"
  Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
  "github.com/skyflowapi/skyflow-go/skyflow/common"
  "github.com/skyflowapi/skyflow-go/commonutils/logger"
)
```
Alternatively, `go get <package_name>` can also be used to download the required dependencies 

### Service Account Token Generation
[This](https://github.com/skyflowapi/skyflow-go/tree/main/serviceaccount) go module is used to generate service account tokens from service account credentials file which is downloaded upon creation of service account. The token generated from this module is valid for 60 minutes and can be used to make API calls to vault services as well as management API(s) based on the permissions of the service account.

The **GenerateBearerToken(filepath)** function takes the credentials file path for token generation, alternatively, you can also send the entire credentials as string, by using **GenerateBearerTokenFromCreds(credentials)**.

[Example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/serviceaccount/token/main/service_account_token.go):

```go
package main
    
import (
    "fmt"
    saUtil "github.com/skyflowapi/skyflow-go/serviceaccount/util"
)
    
var bearerToken = ""

func GetSkyflowBearerToken() (string, error) {

	filePath := "<file_path>"
	if saUtil.IsExpired(bearerToken) {
		newToken, err := saUtil.GenerateBearerToken(filePath)
		if err != nil {
			return "", err
		} else {
			bearerToken = newToken.AccessToken
			return bearerToken, nil
		}
	}
	return bearerToken, nil
}
```


### Vault APIs

The [Vault](https://github.com/skyflowapi/skyflow-go/skyflow/vault-api) Go module is used to perform operations on the vault such as inserting records, detokenizing tokens, retrieving tokens for a skyflow_id and to invoke a connection.

To use this module, the skyflow client must first be initialized as follows.

```go
import (
     Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
     "github.com/skyflowapi/skyflow-go/skyflow/common"
)

configuration := common.Configuration {
        VaultID: "<vauld_id>",      //Id of the vault that the client should connect to 
        VaultURL: "<vault_url>",    //URL of the vault that the client should connect to
        TokenProvider: GetToken     //helper function that retrieves a Skyflow bearer token from your backend
}

skyflowClient := Skyflow.Init(configuration)
```

All Vault APIs must be invoked using a skyflowClient instance.

### Insert data into the vault

To insert data into your vault, use the **Insert(records map[string]interface{}, options common.InsertOptions)** method of the Skyflow client. The **insertInput** parameter requires a `records` key and takes an array of records to insert as a value into the vault. The `options` parameter is a InsertOptions object that provides further options, including Upsert operations, for your insert call, as shown below.

Insert call schema:

```go
import (
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

//Initialize the  SkyflowClient.

var records = make(map[string] interface {})

var record = make(map[string] interface {})
record["table"] = "<your_table_name>"
var fields = make(map[string] interface {})
fields["<field_name>"] = "<field_value>"
record["fields"] = fields

var recordsArray[] interface {}
recordsArray = append(recordsArray, record)

records["records"] = recordsArray

var upsertArray []common.UpsertOptions
var upsertOption = common.UpsertOptions{Table:"<table_name>",Column:"<column_name>"}
upsertArray = append(upsertArray,upsertOption)

options = common.InsertOptions {
        Tokens: true //Optional, indicates whether tokens should be returned for the inserted data. This value defaults to "true".
        Upsert: upsertArray //Optional, upsert support.
}

res, err: = skyflowClient.Insert(records, options)
```

[Insert call example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vault-api/insert.go):

```go
package main

import (
    "fmt"
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

    //Initialize the SkyflowClient.

    var records = make(map[string] interface {})
    var record = make(map[string] interface {})
    record["table"] = "cards"
    var fields = make(map[string] interface {})
    fields["cardNumber"] = "411111111111"
    fields["fullname"] = "name"
    record["fields"] = fields
    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record)
    records["records"] = recordsArray

    var options = common.InsertOptions {
        Tokens: true
    }

    res, err: = skyflowClient.Insert(records, options)

    if err == nil {
        fmt.Println(res.Records)
    }
}
```

Sample response :

```json
{
  "records": [
    {
      "table": "cards",
      "fields": {
        "cardNumber": "f37186-e7e2-466f-91e5-48e2bcbc1",
        "fullname": "1989cb56-63a-4482-adf-1f74cd1a5"
      }
    }
  ]
}

```

[Upsert call example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vault-api/upsert.go):

```go
package main

import (
    "fmt"
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

    //Initialize the SkyflowClient.

    var records = make(map[string] interface {})
    var record = make(map[string] interface {})
    record["table"] = "cards"
    var fields = make(map[string] interface {})
    fields["cardNumber"] = "411111111111"
    fields["fullname"] = "name"
    record["fields"] = fields
    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record)
    records["records"] = recordsArray

    //Create an upsert array.

    var upsertArray []common.UpsertOptions
    var upsertOption = common.UpsertOptions{Table:"cards",Column:"cardNumber"}
    upsertArray = append(upsertArray,upsertOption)

    var options = common.InsertOptions {
        Tokens: true
        Upsert: upsertArray
    }

    res, err: = skyflowClient.Insert(records, options)

    if err == nil {
        fmt.Println(res.Records)
    }
}
```

Sample response :

```json
{
  "records": [
    {
      "table": "cards",
      "fields": {
        "cardNumber": "f37186-e7e2-466f-91e5-48e2bcbc1",
        "fullname": "1989cb56-63a-4482-adf-1f74cd1a5"
      }
    }
  ]
}
```

#### Detokenize
To retrieve tokens from your vault, you can use the **Detokenize(records map[string]interface{})** method.The `records` parameter takes an array of SkyflowIDs to return, as shown below:

```go
import (
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

//initialize skyflowClient

var records = make(map[string] interface {})

var record1 = make(map[string] interface {})
record1["token"] = "<token>"    // token for the record to be fetched
var record2 = make(map[string] interface {})
record2["token"] = "<token>"

var recordsArray[] interface {}
recordsArray = append(recordsArray, record1)
recordsArray = append(recordsArray, record2)

records["records"] = recordsArray

res, err := skyflowClient.Detokenize(records)
```

An [example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vault-api/detokenize.go) of a Detokenize call:

```go
package main

import (
    "fmt"
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

    //initialize skyflowClient

    var records = make(map[string] interface {})
    var record1 = make(map[string] interface {})
    record1["token"] = "45012507-f72b-4f5c-9bf9-86b133bae719"
    var record2 = make(map[string] interface {})
    record2["token"] = "invalid-token"
    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record1)
    recordsArray = append(recordsArray, record2)
    records["records"] = recordsArray

    res, err: = skyflowClient.Detokenize(records)

    if err == nil {
        fmt.Println("Records:",res.Records)
        fmt.Println("Errors:",res.Errors)
    }
}  
```

Sample response:

```json
{
  "records": [
    {
      "token": "131e70dc-6f76-4319-bdd3-96281e051051",
      "value": "1990-01-01"
    }
  ],
  "errors": [
    {
      "token": "invalid-token",
      "error": {
        "code": 404,
        "description": "Tokens not found for invalid-token"
      }
    }
  ]
}
```

#### GetById
 
In order to retrieve data from your vault using SkyflowIDs, use the **GetById(records map[string]interface{})** method. The `records` parameter takes a map that has an array of SkyflowIDs to return, as shown below:

```go
import (
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

//initialize skyflowClient

var records = make(map[string] interface {})

var record1 = make(map[string] interface {})
record1["ids"] = [] string {} {     // List of SkyflowID's of the records to be fetched
    "<skyflow_id1>", "<skyflow_id2>"
}
record1["table"] = "<table_name>"   // name of table holding the above skyflow_id's
record1["redaction"] =  common.PLAIN_TEXT   // redaction to be applied to retrieved data

var recordsArray[] interface {}
recordsArray = append(recordsArray, record1)
records["records"] = recordsArray

res, err := skyflowClient.GetById(records)
```

There are 4 accepted values in Skyflow.RedactionTypes:

-  `PLAIN_TEXT`
-  `MASKED`
-  `REDACTED`
-  `DEFAULT`

  

An [example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vault-api/getById.go) of GetById call:

```go
package main

import (
    "fmt"
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

    //initialize skyflowClient

    var records = make(map[string] interface {})
    var record1 = make(map[string] interface {})
    record1["ids"] = [] string {} {
        "f8d8a622-b557-4c6b-a12c-c5ebe0b0bfd9", "da26de53-95d5-4bdb-99db-8d8c66a35ff9"
    }
    record1["table"] = "cards"
    record1["redaction"] = common.PLAIN_TEXT

    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record1)
    records["records"] = recordsArray

    res, err: = skyflowClient.GetById(records)

    if err == nil {
      fmt.Println("Records:",res.Records)
      fmt.Println("Errors:",res.Errors)
    }
}
```

Sample response:

```json
{
  "records": [
    {
      "fields": {
        "card_number": "4111111111111111",
        "expiry_date": "11/35",
        "fullname": "myname",
        "skyflow_id": "f8d8a622-b557-4c6b-a12c-c5ebe0b0bfd9"
      },
      "table": "cards"
    },
    {
      "fields": {
        "card_number": "4111111111111111",
        "expiry_date": "10/23",
        "fullname": "sam",
        "skyflow_id": "da26de53-95d5-4bdb-99db-8d8c66a35ff9"
      },
      "table": "cards"
    }
  ]
}
```

### InvokeConnection

End-user apps can use InvokeConnection to integrate checkout and card issuance flows with their apps and systems. To invoke a connection, use the invokeConnection(config ConnectionConfig) method of the Skyflow client. The config object must have `connectionURL`,`methodName` and the remaining are optional. 

The InvokeConnection method lets you bypass handling sensitive data by integrating third-party server-side application using APIs. Before invoking the `InvokeConnection` method, you must create a connection and generate a connectionURL. Once you have the connectionURL, you can invoke a connection by using the **InvokeConnection(config ConnectionConfig)** method. The config parameter must include a `connectionURL` and `methodName`. The other fields are optional.

```go

  pathParams := make(map[string]string)
  pathParams["<path_param_key>"] = "<path_param_value>"

  queryParams := make(map[string]interface{})
  queryParams["<query_param_key>"] = "<query_param_value>"

  requestHeader := make(map[string]string)
  requestHeader["<request_header_key>"] = "<request_header_value>"

  requestBody := make(map[string]interface{})
  requestBody["<request_body_key>"] = "<request_body_value>"

  connectionConfig := common.ConnectionConfig{ConnectionURL : "<your_connection_url>",MethodName : "<Method_Name>",PathParams : pathParams,QueryParams : queryParams, RequestBody : requestBody, RequestHeaders : requestHeader}

  skyflowClient.InvokeConnection(connectionConfig)  


```

`methodName` supports the following methods:
- GET
- POST
- PUT
- PATCH
- DELETE

**pathParams, queryParams, requestHeader, requestBody**  objects will be sent through the connection integration url as shown below.

An [example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vault-api/invokeConnection.go) of InvokeConnection call:

```go

package main

import (
    "fmt"
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

    //initialize skyflowClient

    pathParams := make(map[string]string)
    pathParams["card_number"] = "1852-344-234-34251"

    requestHeader := make(map[string]string)
    requestHeaders["Authorization"] = "<YOUR_CONNECTION_AUTH>"

    requestBody := make(map[string]interface{})
    requestBody["expirationDate"] = "12/2026"

    connectionConfig := common.ConnectionConfig{ConnectionURL : "<Connection_URL>",MethodName : "<Method_Name>",PathParams : pathParams, RequestBody : requestBody, RequestHeaders : requestHeader}

    res, err: = skyflowClient.InvokeConnection(connectionConfig)

    if err == nil {
          jsonRes, err: = json.Marshal(res)
          if err == nil {
                fmt.Println("result: ", string(jsonRes))
          }
    }
}
```

Sample invokeConnection Response
```go
{
    "receivedTimestamp": "2021-11-05 13:43:12.534",
    "processingTimeinMs": 12,
    "resource": {
        "cvv2": "558"
    }
}
```

### Logging

The skyflow-go SDK provides useful logging using go libray `github.com/sirupsen/logrus`. By default the logging level of the SDK is set to `LogLevel.ERROR`. This can be changed by using `SetLogLevel(LogLevel)` as shown below:

```go
import "github.com/skyflowapi/skyflow-go/commonutils/logger"

// sets the skyflow-go SDK log level to INFO
logger.SetLogLevel(logger.LogLevel.INFO);
```

Currently the following log levels are supported:

- `DEBUG`:

   When `LogLevel.DEBUG` is passed, all level of logs will be printed(DEBUG, INFO, WARN, ERROR)
   
- `INFO`: 

   When `LogLevel.INFO` is passed, INFO logs for every event that has occurred during the SDK flow execution will be printed along with WARN and ERROR logs
   
- `WARN`: 

   When `LogLevel.WARN` is passed, WARN and ERROR logs will be printed
   
- `ERROR`:

   When `LogLevel.ERROR` is passed, only ERROR logs will be printed.

`Note`:
  - The ranking of logging levels is as follows :  `DEBUG` < `INFO` < `WARN` < `ERROR`.
## Reporting a Vulnerability

If you discover a potential security issue in this project, please reach out to us at security@skyflow.com. Please do not create public GitHub issues or Pull Requests, as malicious actors could potentially view them.
