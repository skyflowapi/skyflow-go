# Description
skyflow-go is the Skyflow SDK for the Go programming language.

## Usage


## Table of Contents

*  [Service Account Token Generation](#service-account-token-generation)
* [Vault APIs](#vault-apis)
  *  [Insert](#insert)
  *  [Detokenize](#detokenize)
  *  [GetById](#get-by-id)
  *  [InvokeGateway](#invoke-gateway)
* [Logging](#logging)

### Service Account Token Generation
[This](https://github.com/skyflowapi/skyflow-go/tree/main/service-account) go module is used to generate service account tokens from service account credentials file which is downloaded upon creation of service account. The token generated from this module is valid for 60 minutes and can be used to make API calls to vault services as well as management API(s) based on the permissions of the service account.

[Example](https://github.com/skyflowapi/skyflow-go/blob/main/examples/service-account/token/main/service_account_token.go):

```go
package main
    
import (
    "fmt"
    saUtil "github.com/skyflowapi/skyflow-go/service-account/util"
)
    
func main() {
    token, err := saUtil.GenerateToken("<path_to_sa_credentials_file>")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("token %v", *token)
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

#### Insert

To insert data into the vault from the integrated application, use the insert(records map[string]interface{}, options common.InsertOptions) method of the Skyflow client. The first parameter records is a map that has a `records` key and takes an array of records to be inserted into the vault as value. The second parameter options is optional, which takes struct as shown below:

```go
import (
    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

//initialize skyflowClient

var records = make(map[string] interface {})

var record = make(map[string] interface {})
record["table"] = "<your_table_name>"
var fields = make(map[string] interface {})
fields["<field_name>"] = "<field_value>"
record["fields"] = fields

var recordsArray[] interface {}
recordsArray = append(recordsArray, record)

records["records"] = recordsArray

// Indicates whether or not tokens should be returned for the inserted data. Defaults to 'true'
options = common.InsertOptions {
        Tokens: true
}

res, err: = skyflowClient.Insert(records, options)
```

An example of an insert call is given below:


```go
package main

import (
    "fmt"

    Skyflow "github.com/skyflowapi/skyflow-go/skyflow/client"
    "github.com/skyflowapi/skyflow-go/skyflow/common"
)

func main() {

    //initialize skyflowClient

    var records = make(map[string]interface{})
    var record = make(map[string]interface{})
	record["table"] = "cards"
    var fields = make(map[string]interface{})
	fields["cardNumber"] = "411111111111"
	fields["fullname"] = "name"
	record["fields"] = fields
    var recordsArray []interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray

    var options = common.InsertOptions{Tokens: true}

	res, err := skyflowClient.Insert(records, options)

	if err == nil {
		result, jsonErr := json.Marshal(res)
		if jsonErr == nil {
			fmt.Println("result", string(result))
		}
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

For retrieving data using tokens, use the detokenize(records map[string]interface{}) method. The records parameter takes a map that contains records to be fetched as shown below.


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

An example of a detokenize call:

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
        jsonRes, err: = json.Marshal(res)
        if err == nil {
            fmt.Println("result: ", string(jsonRes))
        }
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

#### Get By Id

For retrieving data using SkyflowID's, use the getById(records map[string]interface{}) method. The records parameter takes a map that contains records to be fetched as shown below:
  

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
record1["redaction"] =  Skyflow.RedactionType   // redaction to be applied to retrieved data

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

  

An example of getById call:

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
    record1["redaction"] = "PLAIN_TEXT"

    var recordsArray[] interface {}
    recordsArray = append(recordsArray, record1)
    records["records"] = recordsArray

    res, err: = skyflowClient.GetById(records)

    if err == nil {
        jsonRes, err: = json.Marshal(res)
        if err == nil {
            fmt.Println("result: ", string(jsonRes))
        }
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