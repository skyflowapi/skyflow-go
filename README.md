# Skyflow Go SDK
This go SDK is designed to help developers easily implement Skyflow into their go backend. 

[![CI](https://img.shields.io/static/v1?label=CI&message=passing&color=green?style=plastic&logo=github)](https://github.com/skyflowapi/skyflow-go/actions)
[![GitHub release](https://img.shields.io/github/v/release/skyflowapi/skyflow-go.svg)](https://github.com/skyflowapi/skyflow-go/releases)
[![License](https://img.shields.io/github/license/skyflowapi/skyflow-go)](https://github.com/skyflowapi/skyflow-go/blob/main/LICENSE)


# Table of Contents

- [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
    - [Requirements](#requirements)
    - [Configuration](#configuration)
    - [Service Account Token Generation](#service-account-token-generation)
    - [Vault APIs](#vault-apis)
      - [Insert data into the vault](#insert-data-into-the-vault)
      - [Detokenize](#detokenize)
      - [GetById](#getbyid)
      - [Get](#get)
        - [Use Skyflow IDs](#use-skyflow-ids)
        - [Use column name and values](#use-column-name-and-values)
    - [InvokeConnection](#invokeconnection)
    - [Logging](#logging)
  - [Reporting a Vulnerability](#reporting-a-vulnerability)


## Features

- Authentication with a Skyflow Service Account and generation of a bearer token
- Vault API operations to insert, retrieve and tokenize sensitive data
- Invoking connections to call downstream third party APIs without directly handling sensitive data

## Installation

### Requirements
- go 1.18 and above

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
  "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
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
[Example using cred json string](https://github.com/skyflowapi/skyflow-go/blob/main/samples/serviceaccount/token/main/service_account_token_using_cred_string.go):
```go
package main

import (
	"fmt"

	logger "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"
	saUtil "github.com/skyflowapi/skyflow-go/serviceaccount/util"
)

var token = ""

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error : ", err)
		}
	}()
	logger.SetLogLevel(logger.INFO) //set loglevel to INFO
	credentials:= "<credentials_in_string_format>"
	if saUtil.IsExpired(token) {
		newToken, err := saUtil.GenerateBearerTokenFromCreds(credentials)
		if err != nil {
			panic(err)
		} else {
			token = newToken.AccessToken
		}
		fmt.Println("%v", token)
	}
}
```
### Vault APIs

The [Vault](https://github.com/skyflowapi/skyflow-go/tree/main/skyflow/vaultapi) Go module is used to perform operations on the vault such as inserting records, detokenizing tokens, retrieving tokens for a skyflow_id and to invoke a connection.

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

#### Insert data into the vault

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
        ContinueOnError: true // Optional, decides whether to continue if error encountered or not
}

res, err: = skyflowClient.Insert(records, options)
```

[Insert call example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/insert.go):

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
      "request_index": 0,
      "table": "cards",
      "fields": {
        "cardNumber": "f37186-e7e2-466f-91e5-48e2bcbc1",
        "fullname": "1989cb56-63a-4482-adf-1f74cd1a5",
        "skyflow_id": "da26de53-95d5-4bdb-99db-8d8c66a35ff9"
      }
    }
  ]
}

```


[Insert call example with ContinueOnError](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/insert_with_continueOnError.go):

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

    var record2 = make(map[string] interface {})
    record2["table"] = "pii_field"
    var fields2 = make(map[string] interface {})
    fields2["name"] = "name"
    record2["fields"] = fields2

    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record)

    records["records"] = recordsArray

    var options = common.InsertOptions {
        Tokens: true,
        ContinueOnError: true
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
      "request_index": 0,
      "table": "cards",
      "fields": {
        "cardNumber": "f37186-e7e2-466f-91e5-48e2bcbc1",
        "fullname": "1989cb56-63a-4482-adf-1f74cd1a5",
        "skyflow_id": "da26de53-95d5-4bdb-99db-8d8c66a35ff9"
      }
    }
  ],
  "errors": [
    {
      "error": {
        "request_index": 1,
        "code": 404,
        "description": "Object Name pii_field was not found for Vault - requestId : id1234"
      }
    }
  ]
}

```

[Upsert call example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/upsert.go):

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
      "request_index": 0,
      "table": "cards",
      "fields": {
        "cardNumber": "f37186-e7e2-466f-91e5-48e2bcbc1",
        "fullname": "1989cb56-63a-4482-adf-1f74cd1a5",
        "skyflow_id": "da26de53-95d5-4bdb-99db-8d8c66a35ff9"
      }
    }
  ]
}
```

#### Detokenize
To retrieve tokens from your vault, you can use the **Detokenize(records map[string]interface{},options common.DetokenizeOptions)** method.The `records` parameter takes an array of SkyflowIDs to return.The options parameter is a DetokenizeOptions object that provides further options, including `ContinueOnError` operation, for your detokenize call, as shown below:

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
record2["redaction"] = "<RedactionType>" // Optional. Redaction to be applied for retrieved data.

var recordsArray[] interface {}
recordsArray = append(recordsArray, record1)
recordsArray = append(recordsArray, record2)

records["records"] = recordsArray
options := common.DetokenizeOptions {
        ContinueOnError: true //Optional, true indicates making individual API calls. false indicates to make a bulk API call.. This value defaults to "true".
}
res, err := skyflowClient.Detokenize(records, options)

Note: `redaction` defaults to `common.PLAIN_TEXT`
```

An [example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/detokenize.go) of a Detokenize call:

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
[Detokenize call with the ContinueOnError example.](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/detokenize.go):

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
    record2["token"] = "131e70dc-6f76-4319-bdd3-96281e051051"
    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record1)
    recordsArray = append(recordsArray, record2)
    records["records"] = recordsArray
    options := common.DetokenizeOptions {
        ContinueOnError: false
    }
    res, err: = skyflowClient.Detokenize(records, options)

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
      "token": "45012507-f72b-4f5c-9bf9-86b133bae719",
      "value": "Jhon"
    },
    {
      "token": "131e70dc-6f76-4319-bdd3-96281e051051",
      "value": "1990-01-01"
    }
  ],
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

  

An [example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/get_by_id.go) of GetById call:

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

#### Get
 
In order to retrieve data from your vault using Skyflow IDs or by Unique Column Values, use the **Get(records map[string]interface{}, options common.GetOptions)** method. The `records` parameter takes a map that should contain

1. Either an array of Skyflow IDs to fetch
2. Or a column name and array of column values

The second parameter, options, is a GetOptions object that retrieves tokens of Skyflow IDs. 

Note: 
1. GetOptions parameter applicable only for retrieving tokens using Skyflow ID.
2. You can't pass GetOptions along with the redaction type.
3. `tokens` defaults to false.


##### Use Skyflow IDs
1. Retrieve data using Redaction type:

```go
import (
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

//initialize skyflowClient

var records = make(map[string] interface {})

var record1 = make(map[string] interface {})
record1["ids"] = [] string {} {              // List of SkyflowID's of the records to be fetched
    "<skyflow_id1>", "<skyflow_id2>"
}
record1["table"] = "<table_name>"            // Name of table holding the records in the vault.
record1["redaction"] =  common.PLAIN_TEXT    // Redaction type to apply to retrieved data.

var recordsArray[] interface {}
recordsArray = append(recordsArray, record1)
records["records"] = recordsArray

res, err := skyflowClient.Get(records)
```

2. Retrieve tokens using GetOptions:

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
record1["table"] = "<table_name>"   // // Name of table holding the records in the vault.

var recordsArray[] interface {}
recordsArray = append(recordsArray, record1)
records["records"] = recordsArray

res, err := skyflowClient.Get(records, common.GetOptions{Tokens: true})
```

##### Use column name and values
```go
import (
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

//initialize skyflowClient

var records = make(map[string] interface {})

var record1 = make(map[string] interface {})
record1["columnValues"] = [] string {} {     // List of given unique column values.
    "<column_value1>", "<column_value2>"
}
record1["columnName"] = "<column_name>"      // Unique column name in the vault.
record1["redaction"] =  common.PLAIN_TEXT 
record1["table"] = "<table_name>"            // Name of table holding the above skyflow_id's

var recordsArray[] interface {}
recordsArray = append(recordsArray, record1)
records["records"] = recordsArray

res, err := skyflowClient.Get(records)
```

There are 4 accepted values in Skyflow.RedactionTypes:

-  `PLAIN_TEXT`
-  `MASKED`
-  `REDACTED`
-  `DEFAULT`

Examples

An example call using Skyflow IDs with RedactionType:

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

    var record2 = make(map[string] interface {})
    record2["ids"] = [] string {} { "invalid-id" }
    record2["table"] = "cards"
    record2["redaction"] = common.PLAIN_TEXT

    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record1)
    recordsArray = append(recordsArray, record2)

    records["records"] = recordsArray

    res, err: = skyflowClient.Get(records)

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
  ],
  "errors": [
    {
      "error": {
        "code": "404",
        "description": "No Records Found - requestId: fc531b8d-412e-9775-b945-4feacc9b8616"
      },
      "ids": ["Invalid Skyflow ID"]
    }
  ]
}
```

An example call using Skyflow IDs with GetOptions:

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

    var record2 = make(map[string] interface {})
    record2["ids"] = [] string {} { "Invalid Skyflow ID" }
    record2["table"] = "cards"

    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record1)
    recordsArray = append(recordsArray, record2)

    records["records"] = recordsArray

    res, err: = skyflowClient.Get(records, common.GetOptions{Tokens: true})

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
        "card_number": "4555-5176-5936-1930",
        "expiry_date": "23396425-93c9-419b-834b-7750b76a34b0",
        "fullname": "d6bb7fe5-6b77-4842-b898-221c51c3cc20",
        "id": "f8d8a622-b557-4c6b-a12c-c5ebe0b0bfd9"
      },
      "table": "cards"
    },
    {
      "fields": {
        "card_number": "8882-7418-2776-6660",
        "expiry_date": "284fb1f6-3c29-449f-8899-83a7839821bc",
        "fullname": "45a69af3-e22a-4668-9016-08bb2ef2259d",
        "id": "da26de53-95d5-4bdb-99db-8d8c66a35ff9"
      },
      "table": "cards"
    }
  ],
  "errors": [
    {
      "error": {
        "code": "404",
        "description": "No Records Found - requestId: fc531b8d-412e-9775-b945-4feacc9b8616"
      },
      "ids": ["Invalid Skyflow ID"]
    }
  ]
}
```
An example call using column names and values. 

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
    record1["columnValues"] = [] string {} {
        "123455432112345", "123455432112346"
    }
    record1["columnName"] = "bank_account_number"
    record1["table"] = "account_details"
    record1["redaction"] = common.PLAIN_TEXT

    var record2 = make(map[string] interface {})
    record2["columnValues"] = [] string {} { "Invalid Skyflow column value" }
    record1["columnName"] = "bank_account_number"
    record2["table"] = "account_details"
    record2["redaction"] = common.PLAIN_TEXT


    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record1)
    recordsArray = append(recordsArray, record2)

    records["records"] = recordsArray

    res, err: = skyflowClient.Get(records)

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
        "bank_account_number": "123455432112345",
        "pin_code": "123123",
        "name": "vivek jain",
        "id": "492c21a1-107f-4d10-ba2c-3482a411827d"
      },
      "table": "account_details"
    },
    {
      "fields": {
        "bank_account_number": "123455432112346",
        "pin_code": "123123",
        "name": "vivek",
        "id": "ac6c6221-bcd1-4265-8fc7-ae7a8fb6dfd5"
      },
      "table": "account_details"
    }
  ],
  "errors": [
    {
      "columnName": ["bank_account_number"],
      "error": {
        "code": 404,
        "description": "No Records Found - requestId: fc531b8d-412e-9775-b945-4feacc9b8616"
      }
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

An [example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/invoke_connection.go) of InvokeConnection call:

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
import "github.com/skyflowapi/skyflow-go/commonutils/logwrapper"

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

- `OFF`:

  `LogLevel.OFF` can be used to turn off all logging from the Skyflow SDK.

`Note`:
  - The ranking of logging levels is as follows :  `DEBUG` < `INFO` < `WARN` < `ERROR`< `OFF`

.
## Reporting a Vulnerability

If you discover a potential security issue in this project, please reach out to us at security@skyflow.com. Please do not create public GitHub issues or Pull Requests, as malicious actors could potentially view them.
