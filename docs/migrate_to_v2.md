
## Migration from v1 to v2
Below are the steps to migrate the go sdk from v1 to v2.

## Breaking Changes

- **Client initialization:** `TokenProvider`/`VaultURL` pattern replaced by `Credentials` + `VaultConfig` passed to `client.NewSkyflow()`. `VaultURL` is now split into `VaultId` + `ClusterId`.
- **Request/response types:** Operations like insert, get, and detokenize now use typed request/response structs (e.g., `common.InsertRequest` / `InsertResponse`) instead of raw `map[string]interface{}`.
- **Error handling:** Error objects restructured to include `httpStatusCode`, `details`, and `requestId` for richer debugging.
- **Logging:** Global log level replaced by per-instance log level set on the Skyflow client via `client.WithLogLevel(logger.INFO)`.
- **Import paths:** All imports updated from `github.com/skyflowapi/skyflow-go/skyflow/...` to `github.com/skyflowapi/skyflow-go/v2/...`.

---


### **Authentication options**
In V2, we have introduced multiple authentication options.
You can now provide credentials in the following ways:

- **Passing credentials in ENV.** (`SKYFLOW_CREDENTIALS`) (**Recommended**)
- **API Key**
- **Path to your credentials JSON file**
- **Stringified JSON of your credentials**
- **Bearer token**

These options allow you to choose the authentication method that best suits your use case.

#### V1 (Old): Passing the token provider function below as a parameter to the Configuration.

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

#### V2(New): Passing one of the following:
```go
// Option 1: API Key (Recommended) 
skyflowCredentials := common.Credentials{ApiKey: "<YOUR_API_KEY>"} // Replace <API_KEY> with your actual API key

 // Option 2: Environment Variables 
// Set SKYFLOW_CREDENTIALS in your environment

// Option 3: Credentials File
skyflowCredentials := common.Credentials{Path: "<YOUR_CREDENTIALS_FILE_PATH>"} // Replace with the path to credentials file

// Option 4: Stringified JSON
skyflowCredentials := common.Credentials{CredentialsString: "<YOUR_CREDENTIALS_STRING>"} // Replace with the credentials string

// Option 5: Bearer Token
skyflowCredentials := common.Credentials{Token: "<BEARER_TOKEN>"} // Replace <BEARER_TOKEN> with your actual authentication token.
```


### Initializing the client
In V2, we have introduced a functional options design pattern for client initialization and added support for multi-vault. This allows you to configure multiple vaults during client initialization. 

In V2, the log level is tied to each individual client instance.

During client initialization, you can pass the following parameters:

- `VaultID` and `VaultURL`: These values are derived from the vault ID & vault URL.
- `Env`: Specify the environment (e.g., SANDBOX or PROD).
- `Credentials`: The necessary authentication credentials.

#### V1 (Old):
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

#### V2 (New):
```go
import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func main() {
	creds := common.Credentials{Path: "<YOUR_CREDENTIALS_FILE_PATH_1>"}    // Replace with the path to the credentials file
        vaultConfig1 := common.VaultConfig{VaultId: "<VAULT_ID1>", ClusterId: "<CLUSTER_ID1>", Env: common.DEV, Credentials: creds} // Replace with the Cluster and Vault ID of the first vault, Set the environment (e.g., DEV, STAGE, PROD)
        var arr []common.VaultConfig
	arr = append(arr, vaultConfig1)
       // Create a Skyflow client and add vault configurations
        skyflowClient, err := client.NewSkyflow(
		client.WithVaults(arr...), // Add the first vault configuration
		client.WithCredentials(common.Credentials{}), // Add the first vault configuration
		client.WithLogLevel(logger.DEBUG), // Enable debugging for detailed logs
	)
}	
```

#### Key Changes:
- `vaultUrl` replaced with `ClusterId`.
- Added environment specification (`Env`).
- Instance-specific log levels.

###  Request & response structure
In V2, we have removed the use of JSON objects from a third-party package. Instead, we have transitioned to accepting native list and map data structures. This request needs:
- **Table**: The name of the table.
- **Values**: An array list of objects containing the data to be inserted.
The response will be of type `InsertResponse` struct, which contains `InsertedFields` and `Errors`.

#### V1 (Old) :  Request Building
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

options := common.InsertOptions {
        Tokens: true //Optional, indicates whether tokens should be returned for the inserted data. This value defaults to "true".
        Upsert: upsertArray //Optional, upsert support.
        ContinueOnError: true // Optional, decides whether to continue if error encountered or not
}

res, err: = skyflowClient.Insert(records, options)
```
#### V2 (New) : Request building
```go
service, serviceError := skyflowClient.Vault("<VAULT_ID>")
if serviceError != nil {
	fmt.Println(serviceError)
} else {
	ctx := context.TODO()
	values := make([]map[string]interface{}, 0)
	values = append(values, map[string]interface{}{
      "<COLUMN_NAME_1>": "<COLUMN_VALUE_1>", // Replace with column name and value
    })
	values = append(values, map[string]interface{}{
      "<COLUMN_NAME_2>": "<COLUMN_VALUE_2>",  // Replace with another column name and value
    })
    tokens := make([]map[string]interface{}, 0)
    tokens = append(values, map[string]interface{}{
                "<COLUMN_NAME_2>": "<TOKEN_VALUE_2>",
    })
	insert, err := service.Insert(ctx, common.InsertRequest{
      Table:  "<TABLE_NAME>",
      Values: values,
    }, common.InsertOptions{ContinueOnError: false, ReturnTokens: true, TokenMode: common.ENABLE, Tokens: tokens})
	
	if err != nil {
		fmt.Println("Error occurred ", *err)
	} else {
		fmt.Println("RESPONSE:", insert)
	}
}
```
#### V1 (Old) :  Response structure
```json
{
    "Records": [
        {
            "table": "cards",
            "fields": {
                "skyflow_id": "16419435-aa63-4823-aae7-19c6a2d6a19f",
                "cardNumber": "f3907186-e7e2-466f-91e5-48e12c2bcbc1",
                "cvv": "1989cb56-63da-4482-a2df-1f74cd0dd1a5"
            }
        }
    ]
}
```
#### V2 (New) :  Response  structure
```json
{
    "InsertedFields": [
          {
               "card_number": "5484-7829-1702-9110",
               "request_index": "0",
               "skyflow_id": "9fac9201-7b8a-4446-93f8-5244e1213bd1",
               "cardholder_name": "b2308e2a-c1f5-469b-97b7-1f193159399b"
          }
     ],
     "Errors": []
}
```
### Request options
In V2, with the introduction of the Functional options design pattern has made handling optional fields in Go more efficient and straightforward.
#### V1 (Old):
```go
options := common.InsertOptions {
        Tokens: true //Optional, indicates whether tokens should be returned for the inserted data. This value defaults to "true".
        Upsert: upsertArray //Optional, upsert support.
        ContinueOnError: true // Optional, decides whether to continue if error encountered or not
}
```
#### V2 (New):
```go
options := common.InsertOptions{ContinueOnError: false, ReturnTokens: true, TokenMode: common.DISABLE, Upsert: "<UPSERT_COLUMN>"}
```

#### Error structure
In V2, we have enriched the error details to provide better debugging capabilities.
The error response now includes:
- **httpStatusCode**: The HTTP status code.
- **grpcCode**: The gRPC code associated with the error.
- **details & message**: A detailed description of the error.
- **requestId**: A unique request identifier for easier debugging.


#### V1 (Old): Error structure
```json
{
  "code": "<http_code>",
  "message": "<message>",
}
```
#### V2 (New): Error structure
```js
{
  "httpStatusCode": "<http_status>",
  "grpcCode": "<grpc_code>",
  "httpCode": "<http_code>",
  "message": "<message>",
  "requestId": "<request_id>",
  "details": ["<details>"]
}
```

## Credential field names (v2.1+)

The credentials JSON file field names are updated to follow camelCase conventions. Both old and new forms are permanently accepted.

| Old form (still accepted) | New form (preferred) |
|---|---|
| `clientID` | `clientId` |
| `keyID` | `keyId` |
| `tokenURI` | `tokenUri` |

---

## Response field names (v2.1+)

Response maps now use `SkyflowId` (PascalCase). The legacy keys are still present for backward compatibility but are deprecated.

| Deprecated (still returned) | Preferred |
|---|---|
| `skyflow_id` | `SkyflowId` |
| `request_index` | `RequestIndex` |

---

For the full list of changes see [CHANGELOG.md](../CHANGELOG.md).
