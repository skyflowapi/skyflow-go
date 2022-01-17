# Description
skyflow-go is the Skyflow SDK for the Go programming language.

## Table of Contents

  

*  [Service Account Token Generation](#service-account-token-generation)

* [Vault APIs](#vault-apis)

  *  [Insert](#insert)

  *  [Detokenize](#detokenize)

  *  [GetById](#get-by-id)

  *  [InvokeGateway](#invoke-gateway)

* [Logging](#logging)



## Usage

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

The [Vault](https://github.com/skyflowapi/skyflow-go) Go module is used to perform operations on the vault such as inserting records, detokenizing tokens, retrieving tokens for a skyflow_id and to invoke gateway.

  

To use this module, the skyflow client must first be initialized as follows.

  

```go
  
import (
     Skyflow "github.com/skyflowapi/skyflow-go/configuration/"
)
configuration := Skyflow.Configuration{
        VaultID: "<vauld_id>",  //Id of the vault that the client should connect to 
         VaultURL: "<vault_url>", //URL of the vault that the client should connect to
          TokenProvider: GetToken //helper function that retrieves a Skyflow bearer token from your backend
          }
```

  

All Vault APIs must be invoked using a client instance.

  

#### Insert

To insert data into the vault from the integrated application, use the insert(records, options) method of the Skyflow client. The records parameter takes an array of records to be inserted into the vault. The options parameter takes  tokens  as parameter.
'tokens' indicates whether or not tokens should be returned for the inserted data. Defaults to 'True'.

```go

//Initialize Client

 records := {
        "records": [
            {
                "table": "<TABLE_NAME>",
                "fields": {
                    "<FIELDNAME>": "<VALUE>"
                }
            }
        ]
    }
res, err := client.Insert(records, options)

```

  

An example of an insert call is given below:

  

```go
import (
    "fmt",
     Skyflow "github.com/skyflowapi/skyflow-go/configuration/"
    )   
    var records = make(map[string]interface{})
    var record = make(map[string]interface{})
	record["table"] = "cards"
    var fields = make(map[string]interface{})
	fields["cvv"] = "123"
	fields["fullname"] = "name"
	record["fields"] = fields
    var recordsArray []interface{}
	recordsArray = append(recordsArray, record)
	records["records"] = recordsArray
	res, err := client.Insert(records, options)
	if err == nil {
		result, jsonErr := json.Marshal(res)
		if jsonErr == nil {
			fmt.Println("result", string(result))
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
"card_number": "f37186-e7e2-466f-91e5-48e2bcbc1",
"cvv": "1989cb56-63a-4482-adf-1f74cd1a5",
},
}
]
}

```

  

#### Detokenize

  

For retrieving using tokens, use the detokenize(records) method. The records parameter takes an object that contains records to be fetched as shown below.

  

```go
records := {
	"records":[
		{
		"token": "string"  //token for the record to be fetched
		}
	]
}
res, err := client.Detokenize(records)

```

  

An example of a detokenize call:

```go
    import (
    "fmt",
     Skyflow "github.com/skyflowapi/skyflow-go/configuration/"
    )  
    var records = make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["token"] = "<token>"
	var record2 = make(map[string]interface{})
	record2["token"] = "<token>"
	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	recordsArray = append(recordsArray, record2)
	records["records"] = recordsArray
	res, err := client.Detokenize(records)
	if err == nil {
		jsonRes, err := json.Marshal(res)
		if err == nil {
			fmt.Println("result: ", string(jsonRes))
		}
	} 

```

  

Sample response:

```go
{
	"records": [
			{
			"token": "<token>",
			"value": "1990-01-01"
			}
		]	
}

```

  

#### Get By Id

  

For retrieving using SkyflowID's, use the getById(records) method. The records parameter takes an object that contains records to be fetched as shown below:

  

```go
    import (
    "fmt",
     Skyflow "github.com/skyflowapi/skyflow-go/configuration/"
    ) 
    records := {
        "records": [
            {
            "ids": ["id1","id2"], // List of SkyflowID's of the records to be fetched
            "table": "<table_name>", // name of table holding the above skyflow_id's
            "redaction": Skyflow.RedactionType, // redaction to be applied to retrieved data
            }
        ]
    }
	res, err := client.GetById(records)
```

There are 4 accepted values in Skyflow.RedactionTypes:

-  `PLAIN_TEXT`

-  `MASKED`

-  `REDACTED`

-  `DEFAULT`

  

An example of getById call:
```go
   import (
    "fmt",
     Skyflow "github.com/skyflowapi/skyflow-go/configuration/"
    ) 
    var records = make(map[string]interface{})
	var record1 = make(map[string]interface{})
	record1["ids"] = []interface{}{"<id1>", "<id2>"}
	record1["table"] = "cards"
	record1["redaction"] = "PLAIN_TEXT"

	var recordsArray []interface{}
	recordsArray = append(recordsArray, record1)
	records["records"] = recordsArray
	res, err := client.GetById(records)
	if err == nil {
		jsonRes, err := json.Marshal(res)
		if err == nil {
			fmt.Println("result: ", string(jsonRes))
		}
	}

```

Sample response:

```go

{

"records": [

{

"fields": {

"card_number": "4111111111111111",

"cvv": "127",

"expiry_date": "11/35",

"fullname": "myname",

"skyflow_id": "<id1>"

},

"table": "cards"

},

{

"fields": {

"card_number": "4111111111111111",

"cvv": "317",

"expiry_date": "10/23",

"fullname": "sam",

"skyflow_id": "<id2>"

},

"table": "cards"

}

],

"errors": [

{

"error": {

"code": "404",

"description": "No Records Found"

},

"skyflow_ids": ["invalid skyflow id"]

}

]

}

```