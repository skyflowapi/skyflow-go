# Description
The Skyflow Go SDK is designed to help with integrating Skyflow into a go backend.

[![CI](https://img.shields.io/static/v1?label=CI&message=passing&color=green?style=plastic&logo=github)](https://github.com/skyflowapi/skyflow-go/actions)
[![GitHub release](https://img.shields.io/github/v/release/skyflowapi/skyflow-go.svg)](https://github.com/skyflowapi/skyflow-go/releases)
[![License](https://img.shields.io/github/license/skyflowapi/skyflow-go)](https://github.com/skyflowapi/skyflow-go/blob/main/LICENSE)


# Table of Contents

- [Table of Contents](#table-of-contents)
- [Overview](#overview)
- [Install](#install)
  - [Requirements](#requirements)
  - [Configuration](#configuration)
- [Migration from v1 to v2](#migration-from-v1-and-v2)
  - [Authentication options](#authentication-options)
  - [Initializing the client](#initializing-the-client)
  - [Request & response structure](#request--response-structure)
  - [Request options](#request-options)
  - [Error structure](#error-structure)
- [Quickstart](#quickstart)
  - [Authenticate](#authenticate)
  - [Initialize the client](#initialize-the-client)
  - [Insert data into the vault](#insert-data-into-the-vault)
- [Vault](#vault)
  - [Insert data into the vault](#insert-data-into-the-vault-1)
  - [Detokenize](#detokenize)
  - [Tokenize](#tokenize)
  - [Get](#get)
    - [Get by skyflow IDs](#get-by-skyflow-ids)
    - [Get tokens](#get-tokens)
    - [Get By column name and column values](#get-by-column-name-and-column-values)
    - [Redaction types](#redaction-types)
  - [Update](#update)
  - [Delete](#delete)
  - [Query](#query)
- [Connections](#connections)
  - [Invoke a Connection](#invoke-connection)
- [Authentication with bearer tokens](#authenticate-with-bearer-tokens)
  - [Generate a bearer token](#generate-a-bearer-token)
  - [Generate bearer tokens with context](#generate-bearer-tokens-with-context)
  - [Generate scoped bearer tokens](#generate-scoped-bearer-tokens)
  - [Generate signed data tokens](#generate-signed-data-tokens)
- [Logging](#logging)
- [Reporting a Vulnerability](#reporting-a-vulnerability)


## Overview
- Authenticate using a Skyflow service account and generate bearer tokens for secure access.
- Perform Vault API operations such as inserting, retrieving, and tokenizing sensitive data with ease.
- Invoke connections to third-party APIs without directly handling sensitive data, ensuring compliance and data protection.

## Install

### Requirements
- go 1.22.0 and above

### Configuration

Make sure your project is using Go Modules (it will have a go.mod file in its root if it already is):

```go
go mod init
```

Then, reference skyflow-go in a Go program with import:

```go
import (
"github.com/skyflowapi/skyflow-go/v2/client"
"github.com/skyflowapi/skyflow-go/v2/utils/common"
"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
```
Alternatively, `go get <package_name>` can also be used to download the required dependencies

## Migration from v1 and v2
Below are the steps to migrate the go sdk from v1 to v2.

### **Authentication options**
In V2, we have introduced multiple authentication options. You can now provide credentials in the following ways:
- API Key (Recommended)
- Passing credentials in ENV. (SKYFLOW_CREDENTIALS) (Recommended)
- Path to your credentials JSON file
- Stringified JSON of your credentials
- Bearer token
  These options allow you to choose the authentication method that best suits your use case.

#### v1(old):
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

#### Notes
- Use only ONE authentication method.
- API Key or environment variables are recommended for production use.
- Secure storage of credentials is essential.
- For overriding behavior and priority order of credentials, please refer to Initialize the client section in Quickstart.

### Initializing the client
In V2, we have introduced a functional options design pattern for client initialization and added support for multi-vault. This allows you to configure multiple vaults during client initialization. In V2, the log level is tied to each individual client instance. During client initialization, you can pass the following parameters:
- `vaultId` and `clusterId`: These values are derived from the vault ID & vault URL.
- `env`: Specify the environment (e.g., SANDBOX or PROD).
- `credentials`: The necessary authentication credentials.

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
- `vaultUrl` replaced with `clusterId`.
- Added environment specification (`env`).
- Instance-specific log levels.

###  Request & response structure
In V2, we have removed the use of JSON objects from a third-party package. Instead, we have transitioned to accepting native list and map data structures. This request needs:
- Table: The name of the table.
- Values: An array list of objects containing the data to be inserted.
  The response will be of type InsertResponse class, which contains insertedFields and errors.

#### V1 (Old) :  Request building
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
In V2, with the introduction of the builder design pattern has made handling optional fields in Java more efficient and straightforward.
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
- `httpStatus`: The HTTP status code.
- `grpcCode`: The gRPC code associated with the error.
- `details & message`: A detailed description of the error.
- `requestId`: A unique request identifier for easier debugging.


#### V1 (Old): Error structure
```go
{
  code: "<http_code>",
  description: "<description>",
}
```
#### V2 (New): Error structure
```json
{
    httpStatus: "<http_status>",
    grpcCode: <grpc_code>,
    httpCode: <http_code>,
    message: "<message>",
    requestId: "<request_ID>",
    details: [ "<details>" ],
}
```
## Quickstart
Get started quickly with the essential steps: authenticate, initialize the client, and perform a basic vault operation. This section provides a minimal setup to help you integrate the SDK efficiently.

### Authenticate
You can use an API key to authenticate and authorize requests to an API. For authenticating via bearer tokens and different supported bearer token types, refer to the Authenticate with bearer tokens section.
```go
skyflowCredentials := common.Credentials{ApiKey: "<YOUR_API_KEY>"} // Replace <API_KEY> with your actual API key
```

### Initialize the client
To get started, you must first initialize the skyflow client. While initializing the skyflow client, you can specify different types of credentials.

1. API keys
   A unique identifier used to authenticate and authorize requests to an API.
2. Bearer tokens
   A temporary access token used to authenticate API requests, typically included in the Authorization header.
3. Service account credentials file path
   The file path pointing to a JSON file containing credentials for a service account, used
   for secure API access.
4. Service account credentials string (JSON formatted)
   A JSON-formatted string containing service account credentials, often used as an alternative to a file for programmatic authentication.

`Note`: Only one type of credential can be used at a time. If multiple credentials are provided, the last one added will take precedence.

```go
package main

import (
	"context"
	"fmt"
	"github.com/skyflowapi/skyflow-go/v2/client"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
/**
 * Example program to initialize the Skyflow client with various configurations.
 * The Skyflow client facilitates secure interactions with the Skyflow vault,
 * such as securely managing sensitive data.
 */

func main() {
    // Step 1: Define the primary credentials for authentication.
    // Note: Only one type of credential can be used at a time. You can choose between: 
	// - API key
    //  - Bearer token
   //   - A credentials string (JSON-formatted)
   //   - A file path to a credentials file.
  // Initialize primary credentials using a Bearer token for authentication.
  primaryCredentials := common.Credentials{Token: "<BEARER_TOKEN1>"} // Replace <BEARER_TOKEN> with your actual authentication token.
  
	// Step 2: Configure the primary vault details.
  // VaultConfig stores all necessary details to connect to a specific Skyflow vault.
  primaryConfig := common.VaultConfig{
	  VaultId: "<PRIMARY_VAULT_ID>", // Replace with your primary vault's ID.
      ClusterId: "<CLUSTER_ID>", // Replace with the cluster ID (part of the vault URL, e.g., https://{clusterId}.vault.skyflowapis.com).
      Env: common.DEV,           // Set the environment (PROD, SANDBOX, STAGE, DEV).
      Credentials: primaryCredentials,       // Attach the primary credentials to this vault configuration.
  }
  
  // Step 3: Create credentials as a JSON object (if a Bearer Token is not provided).
  // Demonstrates an alternate approach to authenticate with Skyflow using a credentials object.
    credentialsObject = `<CREDS_JSON_OBJECT>`
  // Step 4: Use credentials string.
    skyflowCredentials = common.Credentials{CredentialsString: credentialsObject}

  // Step 5: Define secondary credentials (API key-based authentication as an example).
  // Demonstrates a different type of authentication mechanism for Skyflow vaults.
  secondaryCredentials := common.Credentials{ApiKey: "<API_KEY>"} // Replace with your API Key for authentication.
  
  // Step 6: Configure the secondary vault details.
  // A secondary vault configuration can be used for operations involving multiple vaults.
  secondaryConfig := common.VaultConfig{
		VaultId: "<SECONDARY_VAULT_ID>", // Replace with your secondary vault's ID.
        ClusterId: "<CLUSTER_ID>", // Replace with the corresponding cluster ID.
      Env: common.SANDBOX,  // Set the environment for this vault.
      Credentials: secondaryCredentials, // Attach the secondary credentials to this configuration.
    }
  
	// Step 7: Define tertiary credentials using a path to a credentials JSON file.
  // This method demonstrates an alternative authentication method.
  tertiaryCredentials := common.Credentials{Path: "<PATH_TO_YOUR_CREDENTIALS_JSON_FILE>"}

  // Step 8: Configure the tertiary vault details.
  tertiaryConfig := common.VaultConfig{
    VaultId: "<TERTIARY_VAULT_ID>", // Replace with your tertiary vault's ID.
    ClusterId: "<CLUSTER_ID>", // Replace with the corresponding cluster ID.
    Env: common.SANDBOX,  // Set the environment for this vault.
    Credentials: tertiaryCredentials, // Attach the secondary credentials to this configuration.
  }
  // Step 9: Build and initialize the Skyflow client.
  // Skyflow client is configured with multiple vaults and credentials.
  
  var arr []common.VaultConfig
	arr = append(arr, primaryConfig, secondaryConfig, tertiaryConfig)
  skyflowClient, err := client.NewSkyflow(
		client.WithVaults(arr...),
		client.WithCredentials(skyflowCredentials), // Add JSON-formatted credentials if applicable.
        client.WithLogLevel(logger.DEBUG), // Set log level for debugging or monitoring purposes.
  )
  // The Skyflow client is now fully initialized.
  // Use the `skyflowClient` object to perform secure operations such as:
  // - Inserting data
  // - Retrieving data
  // - Deleting data
  // within the configured Skyflow vaults.
}
```
#### Notes:
- If both Skyflow common credentials and individual credentials at the configuration level are specified, the individual credentials at the configuration level will take precedence.
- If neither Skyflow common credentials nor individual configuration-level credentials are provided, the SDK attempts to retrieve credentials from the `SKYFLOW_CREDENTIALS` environment variable.
- All Vault operations require a client instance.

### Insert data into the vault
To insert data into your vault, use the `Insert` method.  The `InsertRequest` class creates an insert request, which includes the values to be inserted as a list of records. Below is a simple example to get started. For advanced options, check out [Insert data into the vault]() section.

```go
/**
 * This example demonstrates how to insert sensitive data (e.g., card information) into a Skyflow vault using the Skyflow client.
 *
 * 1. Initializes the Skyflow client.
 * 2. Prepares a record with sensitive data (e.g., card number and cardholder name).
 * 3. Creates an insert request for inserting the data into the Skyflow vault.
 * 4. Prints the response of the insert operation.
 */
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func main() {
  // Initialize Skyflow client

  // Step 1: Initialize data to be inserted into the Skyflow vault
  ctx := context.TODO() // Create a context for the operation.

  // Create a slice to hold the data records for insertion.
  values := make([]map[string]interface{}, 0)

  // Add a record with sensitive fields (e.g., card number and cardholder name).
  values = append(values, map[string]interface{}{
    "card_number":     "4111111111111111",// Replace with actual card number (sensitive data)
    "cardholder_name": "john doe",         // Replace with the actual cardholder name  (sensitive data)
  })
  insertRequest := common.InsertRequest{
    Table:  "table1", // Specify the table in the vault where the data will be inserted
    Values: values,   // Attach the data (records) to be inserted
  }
  insertOptions := common.InsertOptions{
    ReturnTokens:    true,  // Request tokenized values to be returned in the response.
  }
  // Step 2: Obtain a Vault service instance for performing operations.
  service, serviceError := skyflowClient.Vault("9f27764a10f7946fe56b3258e117") // Replace the vault ID "9f27764a10f7946fe56b3258e117" with your actual Skyflow vault ID
  if serviceError != nil {
    // Handle errors while getting the vault service instance.
    fmt.Println("Error obtaining Vault service:", serviceError)
  }

  // Step 3: Perform the insert operation using the Skyflow client
  insert, err4 := service.Insert(ctx, insertRequest, insertOptions)
  if err4 != nil {
    // Step 4: Handle any errors that occur during the insert operation.
    fmt.Println("Error occurred: ", *err4)
  } else {
    // Step 5: Print the response from the insert operation.
    fmt.Println("Insert Response: ", insert)
  }
}
```

Skyflow returns tokens for the record that was just inserted.
```javascript
Insert Response: {
	"InsertedFields": [{
		"card_number": "5484-7829-1702-9110",
		"request_index": "0",
		"skyflow_id": "9fac9201-7b8a-4446-93f8-5244e1213bd1",
		"cardholder_name": "b2308e2a-c1f5-469b-97b7-1f193159399b",
	}],
	"Errors": []
}

```

## Vault

The [Vault](https://github.com/skyflowapi/skyflow-go/tree/main/skyflow/vaultapi) module performs operations on the vault, including inserting records, detokenizing tokens, and retrieving tokens associated with a `skyflow_id`.

### Insert data into the vault
Apart from using the `insert` method to insert data into your vault covered in [Quickstart](#quickstart), you can also specify options in `InsertRequest`, such as returning tokenized data, upserting records, or continuing the operation in case of errors.

#### Construct an insert request

```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
/**
 * Example program to demonstrate inserting data into a Skyflow vault, along with corresponding InsertRequest schema.
 *
 */

func main() {
  //  Initialise Skyflow client

  // Step 1: Prepare the data to be inserted into the Skyflow vault.
  ctx := context.TODO() // Create a context for the operation.

  // Create the first record with field names and their respective values
  values := make([]map[string]interface{}, 0)

  // Add the first record with field names and their respective values.
  values = append(values, map[string]interface{}{
    "<FIELD_NAME1_1>": "<VALUE_1>", // Replace with actual field name and value.
  })

  // Create the second record with field names and their respective values
  values = append(values, map[string]interface{}{
    "<FIELD_NAME_2>": "<VALUE_1>", // Replace with actual field name and value.
  })
  // Step 2: Build an InsertRequest object with the table name and the data to insert
  insertRequest := common.InsertRequest{
    Table:  "<TABLE_NAME>", // Replace with the actual table name in your Skyflow vault.
    Values: values,         // Attach the data to be inserted.
  }
  
  // Step 3: Use the Skyflow client to perform the insert operation
  //Obtain a Vault service instance for performing operations. 
  service, err := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with your actual vault ID
  if err != nil {
    // Handle errors while getting the vault service instance.
    fmt.Println("Error obtaining Vault service:", err)
  }
  // Step 4: Perform the insert operation using the Vault service.
  insert, errs := service.Insert(ctx, insertRequest)
  if errs != nil {
    // Handle any exceptions that occur during the insert operation
    fmt.Println("Error occurred while inserting data: ", *err4)
  } else {
    // Print the response from the insert operation
    fmt.Println("Insert Response: ", insert)
  }
}
```

[Insert call example with ContinueOnError option](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/insert_records.go):
  The `ContinueOnError` flag is a boolean that determines whether insert operation should proceed despite encountering partial errors. Set to `true` to allow the process to continue even if some errors occur.

```go
/**
 * This example demonstrates how to insert multiple records into a Skyflow vault using the Skyflow client.
 *
 * 1. Initializes the Skyflow client.
 * 2. Prepares multiple records with sensitive data (e.g., card number and cardholder name).
 * 3. Creates an insert request with the records to insert into the Skyflow vault.
 * 4. Specifies options to continue on error and return tokens.
 * 5. Prints the response of the insert operation.
 */
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func main() {
  // Initialize Skyflow client
  // Step 1: Initialize a list to hold the data records to be inserted into the vault
    ctx := context.TODO() // Create a context for the operation.

  // Create a slice to hold the data records for insertion.
  insertData := make([]map[string]interface{}, 0)

  // Step 2: Create the first record with card number and cardholder name
  insertRecord1 := map[string]interface{}{
    "card_number":     "4111111111111111", // Replace with the actual card number.
    "cardholder_name": "john doe",         // Replace with the actual cardholder name.
  }
  // Step 3: Create the second record with card number and cardholder name
    insertRecord2 := map[string]interface{}{
      "card_number":     "42222222222222222", // Replace with the actual card number.
      "cardholder_name": "john doe",         // Replace with the actual cardholder name.
    }
  // Step 4: Add the records to the insertData map
  insertData = append(insertData, insertRecord1)
  insertData = append(insertData, insertRecord2)
  
  // Step 5: Build the InsertRequest object with the data records to insert
   insertRequest := common.InsertRequest{
     Table:  "table1", // Replace with the actual table name in your Skyflow vault.
     Values: insertData,   // Attach the prepared data for insertion.
   }
  // Step 6: Create insert options to support the continue on error 
   insertOptions := common.InsertOptions{
     ContinueOnError: true,  // Specify to continue inserting records even if an error occurs for some records
     ReturnTokens:    true,  // Specify if tokens should be returned upon successful insertion
   }
   
  // Step 7: Obtain a Vault service instance for performing operations.
  service, serviceError := skyflowClient.Vault("9f27764a10f7946fe56b3258e117") // Replace with your actual vault ID.
  if serviceError != nil {
    // Handle errors while getting the vault service instance.
    fmt.Println("Error obtaining Vault service:", serviceError)

  }
  
  // Step 8: Perform the insert operation using the Vault service.
  insert, err4 := service.Insert(ctx, insertRequest , insertOptions)
  if err4 != nil {
    // Handle any errors that occur during the insert operation.
    fmt.Println("Error occurred: ", *err4)
  } else {
    // Print the response from the insert operation.
    fmt.Println("Insert Response: ", insert)
  }
}

```

Sample response :

```json
{
  "insertedFields": [{
    "card_number": "5484-7829-1702-9110",
    "request_index": "0",
    "skyflow_id": "9fac9201-7b8a-4446-93f8-5244e1213bd1",
    "cardholder_name": "b2308e2a-c1f5-469b-97b7-1f193159399b",
  }],
  "errors": [{
    "request_index": "1",
    "error": "Insert failed. Column card_numbe is invalid. Specify a valid column."
  }]
}
```

[Insert call example with upsert option]():
An upsert operation checks for a record based on a unique column's value. If a match exists, the record is updated; otherwise, a new record is inserted.

```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * This example demonstrates how to insert or upsert a record into a Skyflow vault using the Skyflow client, with the option to return tokens.
 *
 * 1. Initializes the Skyflow client.
 * 2. Prepares a record to insert or upsert (e.g., cardholder name).
 * 3. Creates an insert request with the data to be inserted or upserted into the Skyflow vault.
 * 4. Specifies the field (cardholder_name) for upsert operations.
 * 5. Prints the response of the insert or upsert operation.
 */

func main() {
  // Initialize Skyflow client
  // Step 1: Initialize a list to hold the data records for the insert/upsert operation
  upsertData := make([]map[string]interface{}, 0)
  ctx := context.TODO()

  // Step 2: Create a record with the field 'cardholder_name' to insert or upsert
  upsertRecord := map[string]interface{}{
    "cardholder_name": "jane doe", // Replace with the actual cardholder name
  }
  // Step 3: Add the record to the upsertData list
  upsertData = append(upsertData, upsertRecord)

  // Step 4: Build the InsertRequest object with the upsertData
    insertRequest := common.InsertRequest{
      Table:  "table1", // Specify the table in the vault where data will be inserted/upserted
      Values: upsertData, // Attach the data records to be inserted/upserted
    }
   // Step 5: Create the insert options object 
  insertOptions := common.InsertOptions{
    ReturnTokens: true, // Specify if tokens should be returned upon successful operation
    Upsert: "cardholder_name", // Specify the field to be used for upsert operations (e.g., cardholder_name)
  }

  // Step 5: Obtain a Vault service instance for performing operations.
  service, serviceError := skyflowClient.Vault("<VAULT_ID>")
  if serviceError != nil {
    fmt.Println("Error obtaining Vault service:", serviceError)
  }

  // Step 6: Perform the insert/upsert operation using the Skyflow client
  insert, err4 := service.Insert(ctx, insertRequest, insertOptions)
  if err4 != nil {
    fmt.Println("Error occurred", *err4)
  } else {
    fmt.Println("RESPONSE:", insert)
  }
}
```

Sample response :

```json
{
  "InsertedFields": [{
    "skyflowId": "9fac9201-7b8a-4446-93f8-5244e1213bd1",
    "cardholder_name": "73ce45ce-20fd-490e-9310-c1d4f603ee83"
  }],
  "Errors": []
}
```

### Detokenize
To retrieve tokens from your vault, use the `Detokenize` method. The `DetokenizeRequest` class requires a list of detokenization data as input. Additionally, you can provide optional parameters, such as the redaction type and the option to continue on error.

**Construct a detokenize request**

```go
package vaultapi

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
/**
 * This example demonstrates how to detokenize sensitive data from tokens stored in a Skyflow vault, along with corresponding DetokenizeRequest schema.
 *
 */

func main() {
  // Configure the vaults and Skyflow client
  ctx := context.TODO() // Create a context for the detokenization operation.

  // Step 1: Initialize a list of tokens to be detokenized (replace with actual tokens)
  tokens := []string{"<TOKEN1>", "<TOKEN2>"} // Replace with actual token values.

  // Step 2: Create the DetokenizeRequest object with the tokens and redaction type
  detokenizeRequest := common.DetokenizeRequest{
    ReturnTokens:        tokens,           // Provide the list of tokens to be detokenized
    RedactionType: common.PLAIN_TEXT,    // Specify how the detokenized data should be returned (plain text)
    ContinueOnError: true,               // Continue even if one token cannot be detokenized
  }
  // Step 2: Create the DetokenizeOptions object with the ContinueOnError
  options := common.DetokenizeOptions{
    ContinueOnError: true, // Continue even if one token cannot be detokenized.
  }
  
  // Step 3: Obtain a Vault service instance for performing operations.
  service, serviceError := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with the specific vault ID.
  if serviceError != nil {
    // Handle errors while getting the vault service instance.
    fmt.Println("Error obtaining Vault service:", serviceError)

  }
  // Step 4: Call the Skyflow vault to detokenize the provided tokens
  res, err := service.Detokenize(ctx, detokenizeRequest, options)
  if err != nil {
    // Step 5: Handle any errors that occur during the detokenization process.
    fmt.Println("Error occurred ", err)
  } else {
    // Step 6: Print the detokenization response.
    fmt.Println("RESPONSE: ", res)
  }
}
```

Notes:
- `RedactionType` defaults to `RedactionType.PLAIN_TEXT`.
- `ContinueOnError` defaults to `true`.

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/detokenize.go) of a Detokenize call:

```go
package vaultapi

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
/**
 * This example demonstrates how to detokenize sensitive data from tokens stored in a Skyflow vault.
 *
 * 1. Initializes the Skyflow client.
 * 2. Creates a list of tokens (e.g., credit card tokens) that represent the sensitive data.
 * 3. Builds a detokenization request using the provided tokens and specifies how the redacted data should be returned.
 * 4. Calls the Skyflow vault to detokenize the tokens and retrieves the detokenized data.
 * 5. Prints the detokenization response, which contains the detokenized values or errors.
 */
func main() {
  // Initialize Skyflow client
  // Step 1: Initialize a list of tokens to be detokenized (replace with actual token values)
  tokens := []string{"9738-1683-0486-1480", "6184-6357-8409-6668", "4914-9088-2814-3840"} // Replace with actual token values.

  ctx := context.TODO() // Create a context for the detokenization operation.

  // Step 2: Create the DetokenizeRequest object with the tokens and redaction type.
  detokenizeRequest := common.DetokenizeRequest{
    ReturnTokens:        tokens,               // List of tokens to detokenize.
    RedactionType: common.PLAIN_TEXT,    // Specify the redaction type (e.g., PLAIN_TEXT).
  }
  // Step 3: Obtain a Vault service instance for performing operations.
  service, serviceError := skyflowClient.Vault("9f27764a10f7946fe56b3258e117")             // Replace "9f27764a10f7946fe56b3258e117" with your actual Skyflow vault ID
  if serviceError != nil {
    // Handle errors while getting the vault service instance.
    fmt.Println("Error obtaining Vault service:", serviceError)

  }
  options := common.DetokenizeOptions{
    ContinueOnError: false, // Continue even if one token cannot be detokenized.
  }
  // Step 4: Perform the detokenization operation using the Vault service.
  res, errDetokenize := service.Detokenize(ctx, detokenizeRequest, options)
  
  if errDetokenize != nil {
    // Step 5: Handle any errors that occur during the detokenization process.
    fmt.Println("Error occurred: ", errDetokenize)
  } else {
    // Step 6: Print the detokenization response.
    fmt.Println("RESPONSE: ", res)
  }
}
```

Sample response:
```json
{
  "DetokenizedFields": [{
    "token": "9738-1683-0486-1480",
    "value": "4111111111111115",
    "type": "STRING"
  }, {
    "token": "6184-6357-8409-6668",
    "value": "4111111111111119",
    "type": "STRING"
  }, {
    "token": "4914-9088-2814-3840",
    "value": "4111111111111118",
    "type": "STRING"
  }],
  "Errors": []
}

```
[An example of a detokenize call with `ContinueOnError` option](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/detokenize.go):

```go
package vaultapi

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
/**
 * This example demonstrates how to detokenize sensitive data (e.g., credit card numbers) from tokens in a Skyflow vault.
 *
 * 1. Initializes the Skyflow client.
 * 2. Creates a list of tokens (e.g., credit card tokens) to be detokenized.
 * 3. Builds a detokenization request with the tokens and specifies the redaction type for the detokenized data.
 * 4. Calls the Skyflow vault to detokenize the tokens and retrieves the detokenized data.
 * 5. Prints the detokenization response, which includes the detokenized values or errors.
 */
func main() {
  // Initialize Skyflow client
  // Step 1: Initialize a list of tokens to be detokenized (replace with actual token values)
  tokens := []string{"9738-1683-0486-1480", "6184-6357-8409-6668", "4914-9088-2814-3840"} // Replace with actual token values.
  
  // Step 2: Create the DetokenizeRequest and  DetokenizeOptions object with the tokens and redaction type
  request := common.DetokenizeRequest{
    Tokens:        tokens,              // Provide the list of tokens to detokenize
    RedactionType: common.PLAIN_TEXT,    // Specify the format for the detokenized data (plain text)
  }
  options := common.DetokenizeOptions{
    ContinueOnError: false, // Continue even if one token cannot be detokenized.
  }
  ctx := context.TODO() // Create a context for the detokenization operation.

  // Step 3: Obtain a Vault service instance for performing operations.
  service, serviceError := skyflowClient.Vault("9f27764a10f7946fe56b3258e117")             // Replace "9f27764a10f7946fe56b3258e117" with your actual Skyflow vault ID
  
  if serviceError != nil {
    // Handle errors while getting the vault service instance.
    fmt.Println("Error obtaining Vault service:", serviceError)

  }
  
  // Step 4: Call the Skyflow vault to detokenize the provided tokens
  res, errDetokenize := service.Detokenize(ctx, request, options)
  if errDetokenize != nil {
    // Step 5: Handle any errors that occur during the detokenization process.
    fmt.Println("Error occurred ", errDetokenize)
  } else {
    // Step 6: Print the detokenization response, which contains the detokenized data or errors 
	  fmt.Println("RESPONSE: ", res)
  }
}
```
Sample response:
```json
{
  "DetokenizedFields": [{
    "token": "9738-1683-0486-1480",
    "value": "4111111111111115",
    "type": "STRING"
  }, {
    "token": "6184-6357-8409-6668",
    "value": "4111111111111119",
    "type": "STRING"
  }],
  "Errors": [{
    "token": "4914-9088-2814-384",
    "error": "Token Not Found"
  }]
}
```

### Tokenize
Tokenization replaces sensitive data with unique identifier tokens. This approach protects sensitive information by securely storing the original data while allowing the use of tokens within your application.
To tokenize data, use the `Tokenize` method. The `TokenizeRequest` creates a tokenize request. In this request, you specify the `values` parameter, which is a list of `ColumnValue` objects. Each `ColumnValue` contains two properties: `Value` and `ColumnGroup`.
**Constructing your Tokenize request**
```go
package vaultapi

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * This example demonstrates how to tokenize sensitive data (e.g., credit card information)
 * using the Skyflow client, along with corresponding TokenizeRequest schema.
 */
func main() {
  // Initialize Skyflow client
	
  // Step 1: Create a TokenizeRequest array to hold sensitive data
  ctx := context.TODO()
  var reqArray []common.TokenizeRequest

  // Step 2: Create column values for each sensitive data field (e.g., card number and cardholder name)
  columnValue1 := common.TokenizeRequest{
    ColumnGroup: "<COLUMN_GROUP>", 
    Value:       "<VALUE>",           // Replace <VALUE> and <COLUMN_GROUP> with actual data
  }
  columnValue2 := common.TokenizeRequest{
    ColumnGroup: "<COLUMN_GROUP>", 
    Value:       "<VALUE>",           
  } // Replace <VALUE> and <COLUMN_GROUP> with actual data
  
  // Add the created column values to the TokenizeRequest array
  reqArray = append(reqArray, columnValue1)
  reqArray = append(reqArray, columnValue2)


  // Step 3: Access the vault
  service, serviceErr := skyflowClient.Vault("<VAULT_ID>")
  if serviceErr != nil {
    //  Handle error accessing the vault
    fmt.Println(serviceErr)
  } else {

    // Step 4: Call the Skyflow vault to tokenize the sensitive data
    res, tokenizeErr := service.Tokenize(ctx, reqArray)
    if tokenizeErr != nil {
      // Step 5: Handle error during the tokenization process
      fmt.Println("Error occurred: ", tokenizeErr)
    } else {
      // Step 6: Print the tokenization response, which contains the generated tokens or errors 
      fmt.Println("RESPONSE: ", res)
    }
  }
}
```

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/tokenize_records.go) of Tokenize call
```go
/**
 * This example demonstrates how to tokenize sensitive data (e.g., credit card information) using the Skyflow client.
 *
 * 1. Initializes the Skyflow client.
 * 2. Creates a column value for sensitive data (e.g., credit card number).
 * 3. Builds a tokenize request with the column value to be tokenized.
 * 4. Sends the request to the Skyflow vault for tokenization.
 * 5. Prints the tokenization response, which includes the token or errors.
**/

func main() {
// Initialize Skyflow client
// Step 1: Initialize a array of column values to be tokenized (replace with actual sensitive data)
  ctx := context.TODO()
  var reqArray []common.TokenizeRequest

// Step 2: Create a column value for the sensitive data (e.g., card number with its column group)
columnValue := common.TokenizeRequest{
ColumnGroup: "card_number_cg",              // Replace with actual column group name
Value:       "4111111111111111",            // Replace with the actual sensitive data (e.g., card number)
}
//Step 3:  Add the created column value to the list
reqArray = append(reqArray, columnValue)
ctx := context.TODO()

// Access the vault
service, serviceErr := skyflowClient.Vault("9f27764a10f7946fe56b3258e117") // Replace "9f27764a10f7946fe56b3258e117" with your actual Skyflow vault ID

if serviceErr != nil {
//  Handle error accessing the vault
fmt.Println(serviceErr)
} else {
// Step 4 : Call the Skyflow vault to tokenize the sensitive data
res, tokenizeErr := service.Tokenize(ctx, reqArray)
if tokenizeErr != nil {
// Handle error during the tokenization process
fmt.Println("Error occurred ", tokenizeErr)
} else {
// Step 5: Print the tokenization response, which contains the generated tokens or errors 
fmt.Println("RESPONSE: ", res)
}
}
}
```

Sample response:
```json
{
  "tokens": [5479-4229-4622-1393]
}
```

### Get
To retrieve data using Skyflow IDs or unique column values, use the `Get` method. The `GetRequest` class creates a get request, where you specify parameters such as the table name, redaction type, Skyflow IDs, column names, column values, and whether to return tokens. If you specify Skyflow IDs, you can't use column names and column values, and the inverse is true—if you specify column names and column values, you can't use Skyflow IDs.

**Constructing your get request:**
```go
package vaultapi

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to retrieve data from the Skyflow vault using different methods,
 * along with corresponding GetRequest schema.
 */
func main() {
  // Initialize Skyflow client
  // Step 1: Initialize a array of Skyflow IDs to retrieve records (replace with actual Skyflow IDs)
  ctx := context.TODO() // Prepare the context for the request
  ids := []string{"<SKYFLOW_ID_1>", "<SKYFLOW_ID_2>"} // Replace with actual Skyflow ID

  // Step 2: Create a GetRequest and GetOptions to retrieve records by Skyflow ID without returning tokens
  getRequest := common.GetRequest{
    Table: "<TABLE_NAME>",       // Replace with the actual table name
    Ids:   ids,
  }
  options := common.GetOptions{
    Tokens: false, // Set to false to avoid returning tokens
    RedactionType: common.PLAIN_TEXT, // Redact data as plain text
  }
  
  //  Initialize the Skyflow service and replace <VAULT_ID> with your actual Skyflow vault ID
  service, serviceError := skyflowClient.Vault("<VAULT_ID>")
  if serviceError != nil {
    // Handle any errors during initialization
    fmt.Println("Error occurred while initializing Skyflow service:", serviceError)

  }

  // Send the request to the Skyflow vault and retrieve the records
  res, getErr := service.Get(ctx, getRequest, options)
  if getErr != nil {
    // Handle any errors during the retrieval process
    fmt.Println("Error occurred while retrieving records by ID:", getErr)
  } else {
    // Print the retrieved records
    fmt.Println("Response for records by ID:", res.Data)
  }

  // Step 3: Create another GetRequest and GetOptions to retrieve records by Skyflow ID with tokenized values
  getTokensRequest := common.GetRequest{
    Table: "<TABLE_NAME>",       // Replace with the actual table name
    Ids:   ids, // Replace with actual Skyflow IDs
  }
  options := common.GetOptions{
    Tokens: true, // Set to true to return tokenized values
  }
  
  // Send the request to the Skyflow vault and retrieve the tokenized records
  resWithTokens, getErrWithTokens := service.Get(ctx, getTokensRequest, options)
  if getErrWithTokens != nil {
    // Handle any errors during the retrieval process
    fmt.Println("Error occurred while retrieving tokenized records:", getErrWithTokens)
  } else {
    // Print the retrieved tokenized records
    fmt.Println("Response for tokenized records:", resWithTokens.Data)
  }

  // Step 4: Create a GetRequest to retrieve records based on specific column values
  columnValues := []string{"<COLUMN_VALUE_1>", "<COLUMN_VALUE_2>"} // Replace with the actual column value
  getByColumnRequest := common.GetRequest{
    Table:       "<TABLE_NAME>", // Replace with the actual table name
    ColumnName:  "<COLUMN_NAME>", // Replace with the actual column name
    ColumnValues: columnValues,   // Add the list of column values
  }
  // Send the request to the Skyflow vault and retrieve the records filtered by column values
  getByColumnResponse, getErrByColumn := service.Get(ctx, getByColumnRequest, common.GetOptions{
    RedactionType: common.PLAIN_TEXT, // Redact data as plain text
  })
  if getErrByColumn != nil {
    // Handle any errors during the retrieval process
    fmt.Println("Error occurred while retrieving records by column values:", getErrByColumn)
  } else {
    // Print the retrieved records filtered by column values
    fmt.Println("Response for records by column values:", getByColumnResponse.Data)
  }
}
```

#### Get by skyflow IDs
Retrieve specific records using `skyflow_ids`. Ideal for fetching exact records when IDs are known.

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/get_records.go) of a get call to retrieve data using Redaction type:

```go 
package vaultapi

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to retrieve data from the Skyflow vault using a list of Skyflow IDs.
 *
 * 1. Initializes the Skyflow client with a given vault ID.
 * 2. Creates a request to retrieve records based on Skyflow IDs.
 * 3. Specifies that the response should not return tokens.
 * 4. Uses plain text redaction type for the retrieved records.
 * 5. Prints the response to display the retrieved records.
 */
func main() {
  // Initialize Skyflow client
  // Step 1: Initialize a list of Skyflow IDs (replace with actual Skyflow IDs)
  ids := []string{
    "a581d205-1969-4350-acbe-a2a13eb871a6", // Replace with actual Skyflow ID
    "5ff887c3-b334-4294-9acc-70e78ae5164a", // Replace with actual Skyflow ID
  }

  // Step 2: Create a GetRequest and GetOptions to retrieve records based on Skyflow IDs
  // The request specifies:
  // - `ids`: The list of Skyflow IDs to retrieve
  // - `table`: The table from which the records will be retrieved
  getRequest := common.GetRequest{
    Table: "table1", // Replace with the actual table name
    Ids: ids,
  }
  // The options specifies:
  // - `ReturnTokens`: Set to false, meaning tokens will not be returned in the response
  // - `RedactionType`: Set to PLAIN_TEXT, meaning the retrieved records will have data redacted as plain text
  getOptions := common.GetOptions{
    ReturnTokens: false, // Tokens will not be returned
    RedactionType: common.PLAIN_TEXT, // Data will be redacted as plain text
  } 
  
  // Initialize the Skyflow service
  // Replace <VAULT_ID> with your actual Skyflow vault ID
  service, serviceError := skyflowClient.Vault("<VAULT_ID>")
  if serviceError != nil {
    // Step 4: Handle any errors that occur during the initialization process
    fmt.Println("Error occurred while initializing Skyflow service:", serviceError)

  }

  ctx := context.TODO()
  // Step 3: Send the request to the Skyflow vault and retrieve the records
  res, getErr := service.Get(ctx, getRequest, getOptions)
  if getErr != nil {
    // Step 4: Handle any errors that occur during the data retrieval process
    fmt.Println("Error occurred while retrieving records:", getErr)
  } else {
    // Step 5: Print the retrieved records from the response
    fmt.Println("Response:", res.Data)
  }
}
```

Sample response:
```json
{
  "Data": [{
    "card_number": "4555555555555553",
    "email": "john.doe@gmail.com",
    "name": "john doe",
    "skyflow_id": "a581d205-1969-4350-acbe-a2a13eb871a6",
  }, {
    "card_number": "4555555555555559",
    "email": "jane.doe@gmail.com",
    "name": "jane doe",
    "skyflow_id": "5ff887c3-b334-4294-9acc-70e78ae5164a",
  }],
  "Errors": []
}
```

#### Get tokens
Return tokens for records. Ideal for securely processing sensitive data while maintaining data privacy.

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/get_records.go) of get call to retrieve tokens using Skyflow IDs:
```go
/**
 * This example demonstrates how to retrieve data from the Skyflow vault and return tokens along with the records.
 *
 * 1. Initializes the Skyflow client with a given vault ID.
 * 2. Creates a request to retrieve records based on Skyflow IDs and ensures tokens are returned.
 * 3. Prints the response to display the retrieved records along with the tokens.
 */
func main() {
// Initialize Skyflow client
// Step 1: Initialize a list of Skyflow IDs (replace with actual Skyflow IDs)
ids := []string{
    "a581d205-1969-4350-acbe-a2a13eb871a6", // Replace with actual Skyflow ID
    "5ff887c3-b334-4294-9acc-70e78ae5164a", // Replace with actual Skyflow ID
    }
// Step 2: Create a GetRequest to retrieve records based on Skyflow IDs
// The request specifies:
// - `ids`: The list of Skyflow IDs to retrieve
// - `table`: The table from which the records will be retrieved
getRequest := common.GetRequest{
Table: "table1", // Replace with the actual table name
Ids: ids,
}
// Specify options for the request
// - `returnTokens`: Set to true, meaning tokens will be included in the response	
getOptions := common.GetOptions{
Tokens: true, // Tokens will be returned
}

// Prepare the context for the request
ctx := context.TODO()

// Initialize the Skyflow service
// Replace <VAULT_ID> with your actual Skyflow vault ID
service, serviceError := skyflowClient.Vault("<VAULT_ID>")
if serviceError != nil {
// Handle any errors that occur during the initialization process
fmt.Println("Error occurred while initializing Skyflow service:", serviceError)
}
// Step 3: Send the request to the Skyflow vault and retrieve the records with tokens
res, getErr := service.Get(ctx, getRequest, getOptions)
if getErr != nil {
// Step 4: Handle any errors that occur during the data retrieval process
fmt.Println("Error occurred while retrieving records:", getErr)
} else {
// Step 5: Print the retrieved records from the response
fmt.Println("Response:", res.Data)
}
}
```

Sample response:
```json
{
  "Data": [{
    "card_number": "3998-2139-0328-0697",
    "email": "c9a6c9555060@82c092e7.bd52",
    "name": "82c092e7-74c0-4e60-bd52-c9a6c9555060",
    "skyflow_id": "a581d205-1969-4350-acbe-a2a13eb871a6",
  }, {
    "card_number": "3562-0140-8820-7499",
    "email": "6174366e2bc6@59f82e89.93fc",
    "name": "59f82e89-138e-4f9b-93fc-6174366e2bc6",
    "skyflow_id": "5ff887c3-b334-4294-9acc-70e78ae5164a",
  }],
  "Errors": []
}
```

#### Get By column name and column values
Retrieve records by unique column values. Ideal for querying data without knowing Skyflow IDs, using alternate unique identifiers.

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/get_column_values.go) of get call to retrieve data using column name and column values
```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)
/**
 * This example demonstrates how to retrieve data from the Skyflow vault based on column values.
 *
 * 1. Initializes the Skyflow client with a given vault ID.
 * 2. Creates a request to retrieve records based on specific column values (e.g., email addresses).
 * 3. Prints the response to display the retrieved records after redacting sensitive data based on the specified redaction type.
 */

func main() {
  // Initialize Skyflow client
  // Step 1: Initialize a list of column values (email addresses in this case)
  columnValues := []string{"john.doe@gmail.com", "jane.doe@gmail.com"} // Replace with actual values

  // Step 2: Create a GetRequest and GetOptions to retrieve records based on column values
  // The request specifies:
  // - `table`: The table from which the records will be retrieved
  // - `columnName`: The column to filter the records by (e.g., "email")
  // - `columnValues`: The list of values to match in the specified column
  // - `redactionType`: Defines how sensitive data should be redacted (set to PLAIN_TEXT here)
  request := common.GetRequest{
    Table:        "table1",       // Replace with the actual table name
    ColumnName:   "email",        // The column to filter by (e.g., "email")
    ColumnValues: columnValues,   // The list of column values to match
  }
  options := common.GetOptions{
    RedactionType: common.PLAIN_TEXT, // Set the redaction type (e.g., PLAIN_TEXT)
  }
  
  // Set up the Skyflow vault service
  service, serviceError := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with the actual vault ID
  if serviceError != nil {
    fmt.Println(serviceError) // Print any errors that occur during service initialization
  }

  // Define the context for the API call
  ctx := context.TODO() // Using context to manage the API request lifecycle
  
  // Step 3: Send the Get request to the Skyflow vault and retrieve the records
  response, getErr := service.Get(ctx, request, options)
  if getErr != nil {
    // Step 4: Handle any errors that occur during the data retrieval process
    fmt.Println("Error occurred", getErr)
  } else {
    // Print the response to display the retrieved records
    fmt.Println("RESPONSE:", response.Data)
  }
}
```

Sample response:
```json
{
  "Data": [{
    "card_number": "4555555555555553",
    "email": "john.doe@gmail.com",
    "name": "john doe",
    "skyflow_id": "a581d205-1969-4350-acbe-a2a13eb871a6",
  }, {
    "card_number": "4555555555555559",
    "email": "jane.doe@gmail.com",
    "name": "jane doe",
    "skyflow_id": "5ff887c3-b334-4294-9acc-70e78ae5164a",
  }],
  "Errors": []
}
```
#### Redaction types
Redaction types determine how sensitive data is displayed when retrieved from the vault.

**Available Redaction Types**
- `DEFAULT`: Applies the vault-configured default redaction setting.
- `REDACTED`: Completely removes sensitive data from view.
- `MASKED`: Partially obscures sensitive information.
- `PLAIN_TEXT`: Displays the full, unmasked data.

**Choosing the Right Redaction Type**
- Use `REDACTED` for scenarios requiring maximum data protection to prevent exposure of sensitive information.
- Use `MASKED` to provide partial visibility of sensitive data for less critical use cases.
- Use `PLAIN_TEXT` for internal, authorized access where full data visibility is necessary.

### Update
To update data in your vault, use the `Update` method. The `UpdateRequest` class creates an update request, where you specify parameters such as the table name, data (as a map of key-value pairs), tokens, `ReturnTokens`, and `TokenMode`. If `ReturnTokens` is set to true, Skyflow returns tokens for the updated records. If `ReturnTokens` is set to false, Skyflow returns IDs for the updated records.

**Construct an update request**
```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to update records in the Skyflow vault by providing new data and/or tokenized values, along with corresponding UpdateRequest schema.
 *
 */
func main() {
  // Initialize Skyflow client
  // Step 1: Prepare the data to update in the vault
  // Use a map to store the data that will be updated in the specified table
  data := map[string]interface{}{
    "skyflow_id": "<SKYFLOW_ID>", // Skyflow ID for identifying the record to update
    "<COLUMN_NAME_1>": "<COLUMN_VALUE_1>", // Example of a column name and its value to update
    "<COLUMN_NAME_2>": "<COLUMN_VALUE_2>", // Another example of a column name and its value to update
  }
  // Step 2: Prepare the tokens (if necessary) for certain columns that require tokenization
  // Use a map to specify columns that need tokens in the update request
    tokens := map[string]interface{}{
		"COLUMN_NAME_2": "<TOKEN_VALUE_2>",
	}
  // Define the context for the API call
  ctx := context.TODO() // Using context to manage the API request lifecycle
  
  // Step 3: Create an UpdateRequest to specify the update operation
  // The request includes the table name, token mode, data, tokens, and the returnTokens flag
  updateRequest := common.UpdateRequest{
    Table:  "<TABLE_NAME>",         // Replace with the actual table name
    Id:     "<SKYFLOW_ID>",         // The Skyflow ID to identify the record to update
    Values: data,                   // The data to update in the record
  }
  updateOptions := common.UpdateOptions{
    Tokens: true,             // Specify whether to return tokens in the response
    TokenMode:    common.DISABLE,   // Specify the tokenization mode (e.g., ENABLE or DISABLE)
  }
  // Set up the Skyflow vault service
  service, serviceErr := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with the actual vault ID
  if serviceErr != nil {
    // Handle errors that occur during service initialization
    fmt.Println(serviceErr) // Print the error
  }
  
  // Step 4: Send the request to the Skyflow vault and update the record
  response, errUpdate := service.Update(ctx, updateRequest, updateOptions)
  if errUpdate != nil {
    // Handle errors that occur during the update operation
    fmt.Println("Error occurred", *errUpdate) // Print the error for debugging purposes
  } else {
    // Print the response to confirm the update result
    fmt.Println("response:", response)
  }
}
```

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/update_record.go) of update call
```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to update a record in the Skyflow vault with specified data and tokens.
 *
 * 1. Initializes the Skyflow client with a given vault ID.
 * 2. Constructs an update request with data to modify and tokens to include.
 * 3. Sends the request to update the record in the vault.
 * 4. Prints the response to confirm the success or failure of the update operation.
 */

func main() {
  // Initialize Skyflow client
  // Step 1: Prepare the data to update in the vault
  // Use a map to store the data that will be updated in the specified table
  data := map[string]interface{}{
    "skyflow_id":  "5b699e2c-4301-4f9f-bcff-0a8fd3057413",   // Skyflow ID identifies the record to update
    "name":        "john doe",       // Updating the "name" column with a new value
    "card_number": "4111111111111115", // Updating the "card_number" column with a new value
  }
  // Step 2: Prepare the tokens to include in the update request
  // Tokens can be included to update sensitive data with tokenized values
  tokens := map[string]interface{}{
    "name": "72b8ffe3-c8d3-4b4f-8052-38b2a7405b5a",
  }
  
  // Step 3: Create an UpdateRequest to define the update operation
  // The request specifies the table name, token mode, data, and tokens for the update
  updateRequest := common.UpdateRequest{
    Table:  "table1",  // Replace with the actual table name
    Id:     "5b699e2c-4301-4f9f-bcff-0a8fd3057413",  // Skyflow ID to identify the record to update
    Values: data,            // The data to update in the record
  }
  // Define update options, including tokenization mode
  updateOptions := common.UpdateOptions{
    ReturnTokens: true,      // Specify whether to return tokens in the response
    TokenMode:    common.DISABLE, // Specify tokenization mode (e.g., DISABLE means no tokenization)
  }

  
  //Set up the Skyflow vault service
  // Initialize the Skyflow client with the provided Vault ID
  service, serviceErr := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with your actual Vault ID
  if serviceErr != nil {
    // Handle errors that occur during the service initialization
    fmt.Println(serviceErr) // Print the error for debugging purposes
  }
  // Step 4: Send the update request to the Skyflow vault
  response, errUpdate := service.Update(context.TODO(), updateRequest, updateOptions)
  if errUpdate != nil {
    // Handle errors that occur during the update operation
    fmt.Println("Error occurred", *errUpdate) // Print the error for debugging purposes
  } else {
    // Print the response to confirm the update result
    fmt.Println("response:", response)
  }
}
```

Sample response:
When `ReturnTokens` is set to `true`
```json
{
  "skyflowId": "5b699e2c-4301-4f9f-bcff-0a8fd3057413",
  "name": "72b8ffe3-c8d3-4b4f-8052-38b2a7405b5a",
  "card_number": "4315-7650-1359-9681"
}
```

When `ReturnTokens` is set to `false`
```json
{
  "skyflowId": "5b699e2c-4301-4f9f-bcff-0a8fd3057413"
}
```

### Delete
To delete records using Skyflow IDs, use the `Delete` method. The `DeleteRequest` struct accepts a list of Skyflow IDs that you want to delete, as shown below:
**Construct a delete request**
```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to delete records from a Skyflow vault using specified Skyflow IDs, along with corresponding DeleteRequest schema.
 *
 */
func main() {
  // Initialize Skyflow client 
  // Step 1: Prepare a list of Skyflow IDs for the records to delete
  // The list stores the Skyflow IDs of the records that need to be deleted from the vault
  ids := []string{
    "<SKYFLOW_ID_1>", // Replace with actual Skyflow ID 1
    "<SKYFLOW_ID_2>", // Replace with actual Skyflow ID 2
    "<SKYFLOW_ID_3>", // Replace with actual Skyflow ID 3
  }
  // Define the context for the API call
  ctx := context.TODO() // Use context to manage the lifecycle of the API request

  // Step 2: Create a DeleteRequest to define the delete operation
  // The request specifies the table from which to delete the records and the IDs of the records to delete
  deleteRequest := common.DeleteRequest{
    Table: "<TABLE_NAME>", // Replace with the actual table name from which to delete records
    Ids:   ids,            // List of Skyflow IDs to delete
  }

  // Set up the Skyflow vault service
  service, serviceErr := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with the actual vault ID
  if serviceErr != nil {
    // Handle errors during service initialization
    fmt.Println(serviceErr) // Print the error message

  }
  
  // Step 3: Send the delete request to the Skyflow vault
  deleteResponse, errDelete := service.Delete(ctx, deleteRequest) // Call to delete records from the vault
  if errDelete != nil {
    // Handle errors during the delete operation
    fmt.Println("Error occurred", *errDelete) // Print the error message for debugging
  } else {
    // Step 4: Print the response to confirm the delete result
    // The response confirms whether the delete operation was successful
    fmt.Println("response:", deleteResponse) // Print the delete response
  }
}
```

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/delete.go) of delete call
```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to delete records from a Skyflow vault using specified Skyflow IDs.
 *
 * 1. Initializes the Skyflow client with a given Vault ID.
 * 2. Constructs a delete request by specifying the IDs of the records to delete.
 * 3. Sends the delete request to the Skyflow vault to delete the specified records.
 * 4. Prints the response to confirm the success or failure of the delete operation.
 **/

func main() {
  // Step 1: Set up the Skyflow vault service
  service, serviceErr := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with your actual vault ID
  if serviceErr != nil {
    // Handle any errors that occur during service initialization
    fmt.Println(serviceErr) // Print the error
  }

  // Prepare a list of Skyflow IDs for the records to delete
  // The list stores the Skyflow IDs of the records that need to be deleted from the vault
  ids := []string{
    "9cbf66df-6357-48f3-b77b-0f1acbb69280", // Replace with actual Skyflow ID 1
    "ea74bef4-f27e-46fe-b6a0-a28e91b4477b", // Replace with actual Skyflow ID 2
    "47700796-6d3b-4b54-9153-3973e281cafb", // Replace with actual Skyflow ID 3
  }

  // Step 2: Create a DeleteRequest to define the delete operation
  // The request specifies the table from which to delete the records and the IDs of the records to delete
  
  deleteRequest := common.DeleteRequest{
    Table: "<TABLE_NAME>", // Replace with the actual table name from which to delete
    Ids:   ids,            // List of Skyflow IDs to delete
  }

  // Step 3: Send the delete request to the Skyflow vault
  deleteResponse, errDelete := service.Delete(context.TODO(), deleteRequest)
  if errDelete != nil {
    // Handle errors that occur during the delete operation
    fmt.Println("Error occurred", *errDelete) // Print the error for debugging purposes
  } else {
    // Step 4: Print the response to confirm the delete result
    fmt.Println("response:", deleteResponse)
  }
}
```

Sample response:
```json
{
  "DeletedIds": [
    "9cbf66df-6357-48f3-b77b-0f1acbb69280",
    "ea74bef4-f27e-46fe-b6a0-a28e91b4477b",
    "47700796-6d3b-4b54-9153-3973e281cafb"
  ]
}
```

### Query
To retrieve data with SQL queries, use the `Query` method. The `QueryRequest` struct accepts a `query` parameter, as shown below.

**Construct a query request**
Refer to [Query your data](https://docs.skyflow.com/query-data/) and [Execute Query](https://docs.skyflow.com/record/#QueryService_ExecuteQuery) for guidelines and restrictions on supported SQL statements, operators, and keywords.
```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to execute a custom SQL query on a Skyflow vault, along with QueryRequest schema.
 *
 */
func main() {
  // Initialize Skyflow client
  // Step 1: Define the SQL query to execute on the Skyflow vault
  // Replace "<YOUR_SQL_QUERY>" with the actual SQL query you want to run
  query := "<YOUR_SQL_QUERY>" // Example: "SELECT * FROM demo WHERE skyflow_id='<ID>'"

  // Step 2: Create a QueryRequest with the specified SQL query
  queryRequest := common.QueryRequest{
    Query: query, // Pass the query string to the request
  }
  // Step 3: Initialize the Skyflow vault service
  service, serviceError := skyflowClient.Vault("<VAULT_ID>") // Replace <VAULT_ID> with the actual Vault ID
  if serviceError != nil {
    // Handle errors that occur during the service initialization
    fmt.Println(serviceError) // Print the error message

  }
  ctx := context.TODO() // Using context to manage the API request lifecycle

  // Step 4: Execute the query request on the specified Skyflow vault
  res, queryErr := service.Query(ctx, queryRequest) // Execute the query

  // Step 5: Handle the response or any errors from the query execution
  if queryErr != nil {
    // Handle any errors that occur during query execution
    fmt.Println("Error occurred: ", *queryErr) // Print the error message
  } else {
    // Print the response containing the query results
    fmt.Println("RESPONSE: ", res)
  }
}
```

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/query_record.go) of query call
```go
package main

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)

/**
 * This example demonstrates how to execute a SQL query on a Skyflow vault to retrieve data.
 *
 * 1. Initializes the Skyflow client with the Vault ID.
 * 2. Constructs a query request with a specified SQL query.
 * 3. Executes the query against the Skyflow vault.
 * 4. Prints the response from the query execution.
 **/


func main() {
  // Initialize Skyflow client
  // Step 1: Define the SQL query to execute
  // Example query: Retrieve all records from the "demo" table with a specific skyflow_id
  query := "SELECT * FROM cards WHERE skyflow_id='3ea3861-x107-40w8-la98-106sp08ea83f'" // Replace with the actual Skyflow ID to filter the query 

  // Step 2: Create a QueryRequest with the SQL query
  queryRequest := common.QueryRequest{
    Query: query, // SQL query to execute
  }

  // Set up the Skyflow vault service
  service, serviceError := skyflowClient.Vault("9f27764a10f7946fe56b3258e117") // Replace with the actual vault ID
  if serviceError != nil {
    // Handle errors that occur during service initialization
    fmt.Println(serviceError) // Print the error

  }

  // Step 3: Execute the query request on the specified Skyflow vault and handle the response
  ctx := context.TODO() // Context for managing the lifecycle of the query request
  res, queryErr := service.Query(ctx, queryRequest) // Execute the query request

  if queryErr != nil {
    // Handle any errors that occur during query execution
    fmt.Println("Error occurred ", *queryErr) // Print the error for debugging purposes
  } else {
    // Step 5: Print the response from the query execution
    // The response contains the query results retrieved from the Skyflow vault
    fmt.Println("RESPONSE: ", res) // Print the query response to show the results
  }
}
```

Sample response:
```json
{
  "fields": [{
    "card_number": "XXXXXXXXXXXX1112",
    "name": "S***ar",
    "skyflow_id": "3ea3861-x107-40w8-la98-106sp08ea83f",
    "tokenizedData": null
  }]
}
```

## Connections
Skyflow Connections is a gateway service leveraging tokenization to securely send and receive data between your systems and first- or third-party services. The [connections](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/invoke_connection.go) module is used to invoke both INBOUND and/or OUTBOUND connections.
- **Inbound Connections**: Act as intermediaries between your client and server, tokenizing sensitive data before it reaches your backend, ensuring downstream services handle only tokenized data.
- **Outbound Connections**: Enable secure extraction of data from the vault and transfer it to third-party services via your backend server.

### Invoke Connection
To invoke a connection, use the `Invoke` method of the Skyflow client.

**Construct an invoke connection request**
```go
package vaultapi

import (
  "context"
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
  . "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
)
/**
 * This example demonstrates how to invoke an external connection using the Skyflow SDK, along with corresponding InvokeConnectionRequest schema.
 *
 */


func main() {
  // Initialize Skyflow client 
  // Step 1: Define the request body parameters
  // These are the values you want to send in the request body
  ctx := context.TODO() // Define the context of the request
  body := map[string]interface{}{ // Set your data in the body of the request
    "<KEY>": "<VALUE>", // Example BODY
  }
  // Step 2: Define the request headers
  // Add any required headers that need to be sent with the request
  headers := map[string]string{ 
    "<HEADER_NAME_1>": "<HEADER_VALUE_1>", 
    "<HEADER_NAME_2>": "<HEADER_VALUE_2>", 
  }
  // Step 3: Define the path parameters
  // Path parameters are part of the URL and typically used in RESTful APIs
  pathParams := map[string]string{ 
    "<YOUR_PATH_PARAM_KEY_1>": "<YOUR_PATH_PARAM_VALUE_1>", 
  }
  // Step 4: Define the query parameters
  // Query parameters are included in the URL after a '?' and are used to filter or modify the response
  queryParams := map[string]interface{}{ 
    "<YOUR_QUERY_PARAM_KEY_1>": "<YOUR_QUERY_PARAM_VALUE_1>",
    "<YOUR_QUERY_PARAM_KEY_2>": "<YOUR_QUERY_PARAM_VALUE_2>",
  }
  // Step 5: Build the InvokeConnectionRequest
  // Construct the request by specifying method, headers, body, query parameters, and path parameters
  req := InvokeConnectionRequest{
    Method:      POST,        // The HTTP method to use for the request (POST in this case)
    Headers:     headers,     // The headers to include in the request
    Body:        body,        // Attach the request body
    QueryParams: queryParams,  // The query parameters to append to the URL
    PathParams:  pathParams,  // The path parameters for the URL
  }
  // Replace "<CONNECTION_ID>" with the actual connection ID you are using
  service, conError := client1.Connection("<CONNECTION_ID1>") // Replace with actual connection ID
  if conError != nil {
    // Handle errors when establishing the connection
    fmt.Println("Error:", conError) // Print the connection error if it occurs
  } else {
    // Step 6: Invoke the connection using the request
    // Send the request to the external connection and receive the response
    res, invokeError := service.Invoke(ctx, req) // Invoke the connection with the provided request
    if invokeError != nil {
      // Handle any errors that occur during the connection invocation
      fmt.Println("Error occurred ", *invokeError) // Print the error if the invocation fails
    } else {
      // Step 7: Print the response from the invoked connection
      // The response contains the result of the request sent to the external system
      fmt.Println("RESPONSE", res) // Print the successful response
    }
  }
}
```

`method` supports the following methods:
- GET
- POST
- PUT
- PATCH
- DELETE

`PathParams`, `QueryParams`, `RequestHeader`, `RequestBody` are the objects represented as map, that will be sent through the connection integration url.

An [example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/vaultapi/invoke_connection.go) of invokeConnection
```go
import (
"context"
"fmt"
"github.com/skyflowapi/skyflow-go/v2/utils/logger"

. "github.com/skyflowapi/skyflow-go/v2/client"
. "github.com/skyflowapi/skyflow-go/v2/utils/common"
)
/**
 * This example demonstrates how to invoke an external connection using the Skyflow SDK.
 * It configures a connection, sets up the request, and sends a POST request to the external service.
 *
 * 1. Initialize Skyflow client with connection details.
 * 2. Define the request body, headers, and method.
 * 3. Execute the connection request.
 * 4. Print the response from the invoked connection.
 */

func main() {
// Initialize Skyflow client 
// Step 1: Set up credentials and connection configuration
// Load credentials from a JSON file (you need to provide the correct path)
credentials := Credentials{Path: "../cred.json"}

// Define the connection configuration (URL and credentials)
connConfig1 := ConnectionConfig{
	ConnectionId: "<CONNECTION_ID1>", // Replace with actual connection ID
    ConnectionUrl: "https://connection.url.com", // Replace with actual connection URL
    Credentials: credentials, // Set credentials for the connection
}
// Add connection configurations to an array
var arr []ConnectionConfig
arr = append(arr, connConfig1)

// Initialize the Skyflow client with the connection configuration
skyflowClient, clientError := NewSkyflow(
WithConnections(arr...),       // Add the connection configurations to the client
WithLogLevel(logger.DEBUG),    // Set log level to DEBUG for detailed logs
)
if clientError != nil {
// Handle any errors that occur during Skyflow client initialization
fmt.Println("Error:", clientError)
} else {
// Replace "<CONNECTION_ID1>" with the actual connection ID
service, conError := skyflowClient.Connection("<CONNECTION_ID1>")
if conError != nil {
// Handle errors that occur during the connection setup
fmt.Println("Error:", conError)
} else {
// Step 2: Define the request body and headers
// Map for request body parameters
ctx := context.TODO() // Define the context for the API call
body := map[string]interface{}{ // Set your request data
"card_number": "4337-1696-5866-0865", // Example card number
"ssn": "524-41-4248",                // Example SSN
}
// Map for request headers
headers := map[string]string{ // Set the request headers
"Content-Type": "application/json", // Specify the content type for the request
}
// Step 3: Build the InvokeConnectionRequest with required parameters
// Set HTTP method to POST, include the request body and headers
req := InvokeConnectionRequest{
Method:      POST,    // Set the HTTP method to POST
Headers:     headers, // Add request headers
Body:        body,    // Add the body with request data
}

// Step 4: Invoke the connection and capture the response
res, invokeError := service.Invoke(ctx, req)
if invokeError != nil {
// Handle any errors that occur during the connection invocation
fmt.Println("Error occurred ", *invokeError)
} else {
// Step 8: Print the response from the connection invocation
fmt.Println("RESPONSE", res)
}
}
}
}
```
Sample response:
```json
{
  "data": {
    "card_number": "4337-1696-5866-0865",
    "ssn": "524-41-4248"
  },
  "metadata": {
    "requestId": "4a3453b5-7aa4-4373-98d7-cf102b1f6f97"
  }
}

```


## Authenticate with bearer tokens
This section covers methods for generating and managing tokens to authenticate API calls:
1. **Generate a bearer token**:
Enable the creation of bearer tokens using service account credentials. These tokens, valid for 60 minutes, provide secure access to Vault services and management APIs based on the service account's permissions. Use this for general API calls when you only need basic authentication without additional context or role-based restrictions.
2. **Generate a bearer token with context**:
Support embedding context values into bearer tokens, enabling dynamic access control and the ability to track end-user identity. These tokens include context claims and allow flexible authorization for Vault services. Use this when policies depend on specific contextual attributes or when tracking end-user identity is required.
3. **Generate a scoped bearer token**:
Facilitate the creation of bearer tokens with role-specific access, ensuring permissions are limited to the operations allowed by the designated role. This is particularly useful for service accounts with multiple roles. Use this to enforce fine-grained role-based access control, ensuring tokens only grant permissions for a specific role.
4. **Generate signed data tokens**:
Add an extra layer of security by digitally signing data tokens with the service account's private key. These signed tokens can be securely detokenized, provided the necessary bearer token and permissions are available. Use this to add cryptographic protection to sensitive data, enabling secure detokenization with verified integrity and authenticity.

### Generate a bearer token
The [Service Account]() go module is designed to generate service account tokens using a service account credentials file, which is provided when a service account is created. The tokens generated by this module are valid for 60 minutes and can be used to make API calls to Vault services and management APIs, depending on the permissions assigned to the service account.

The **GenerateBearerToken(filepath)** utility provides functionality for generating bearer tokens using a credentials JSON file. Alternatively, you can pass the credentials as a string to achieve the same result.

[Example](https://github.com/skyflowapi/skyflow-go/blob/main/samples/serviceaccount/token/main/service_account_token.go):

```go
import (
"fmt"
saUtil "github.com/skyflowapi/skyflow-go/v2/serviceaccount"
"github.com/skyflowapi/skyflow-go/v2/utils/common"
"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
/**
 * Example program to generate a Bearer Token using Skyflow's BearerToken utility.
 * The token can be generated in two ways:
 * 1. Using the file path to a credentials.json file.
 * 2. Using the JSON content of the credentials file as a string.
 */
func BearerTokenGenerationExample() {
// Variable to store the generated token
var token string

// Example 1: Generate Bearer Token using a credentials.json file
// Specify the full file path to the credentials.json file
var filePath = "<YOUR_CREDENTIALS_FILE_PATH>"

// Check if the token is either not initialized or has expired
if saUtil.IsExpired(token) {
// Create a BearerToken using the credentials file
res, err := saUtil.GenerateBearerToken(filePath, common.BearerTokenOptions{
LogLevel: logger.DEBUG,
})
if err != nil {
fmt.Println("errors", *err)
} else {
token = res.AccessToken
}
}

// Print the generated Bearer Token to the console
fmt.Println("Generated Bearer Token (from file): " + token)

// Example 2: Generate Bearer Token using the credentials JSON as a string
// Provide the credentials JSON content as a string
var fileContents = "<YOUR_CREDENTIALS_FILE_CONTENTS_AS_STRING>"

// Check if the token is either not initialized or has expired
if saUtil.IsExpired(token) {
// Create a BearerToken using the credentials string
res, err := saUtil.GenerateBearerTokenFromCreds(fileContents, common.BearerTokenOptions{
LogLevel: logger.DEBUG,
})
if err != nil {
fmt.Println("Errors", *err)
} else {
fmt.Println("Token", res.AccessToken)
}
token = res.AccessToken
}

// Print the generated Bearer Token to the console
fmt.Println("Generated Bearer Token: " + token)
}
```
### Generate bearer tokens with context
`Context-Aware Authorization`  embeds context values into a bearer token during its generation and so you can reference those values in your policies. This enables more flexible access controls, such as helping you track end-user identity when making API calls using service accounts, and facilitates using signed data tokens during detokenization.
A service account with the `context_id` identifier generates bearer tokens containing context information, represented as a JWT claim in a Skyflow-generated bearer token. Tokens generated from such service accounts include a `context_identifier` claim, are valid for 60 minutes, and can be used to make API calls to the Data and Management APIs, depending on the service account's permissions.

[Example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/serviceaccount/token_generation_with_context.go)
```go
import (
"fmt"
saUtil "github.com/skyflowapi/skyflow-go/v2/serviceaccount"
"github.com/skyflowapi/skyflow-go/v2/utils/common"
"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)
/**
 * Example program to generate a Bearer Token using Skyflow's BearerToken utility.
 * The token is generated using two approaches:
 * 1. By providing the credentials.json file path.
 * 2. By providing the contents of credentials.json as a string.
 */

func BearerTokenGenerationWithContextExample() {
// Variable to store the generated Bearer Token
var bearerToken = "";

// Approach 1: Generate Bearer Token by specifying the path to the credentials.json file
// Replace <YOUR_CREDENTIALS_FILE_PATH> with the full path to your credentials.json file
var filePath = "<YOUR_CREDENTIALS_FILE_PATH>";

// Create a BearerToken using the file path
res, err := saUtil.GenerateBearerToken(filePath, // Set credentials using a File object
common.BearerTokenOptions{
	LogLevel: logger.DEBUG, 
	Ctx: "<CONTEXT>" // Set context string (example: "abc")
})

if err != nil {
// Handle exceptions specific to Skyflow operations
fmt.Println("errors:", *err)
} else {
// Print the generated Bearer Token to the console
bearerToken = res.AccessToken
fmt.Println("Token", res.AccessToken)
}
// Print the generated Bearer Token to the console
fmt.Println(bearerToken);

// Approach 2: Generate Bearer Token by specifying the contents of credentials.json as a string
// Replace <YOUR_CREDENTIALS_FILE_CONTENTS_AS_STRING> with the actual contents of your credentials.json file
var fileContents = "<YOUR_CREDENTIALS_FILE_CONTENTS_AS_STRING>";

// Create a BearerToken object using the file contents as a string
res, err = saUtil.GenerateBearerTokenFromCreds(fileContents, common.BearerTokenOptions{LogLevel: logger.DEBUG, Ctx: "<CONTEXT>"})

if err != nil {
fmt.Println("errors:", *err)
} else {
// Retrieve the Bearer Token as a string
bearerToken = res.AccessToken
// Print the generated Bearer Token to the console
fmt.Println("Token", res.AccessToken)
}
// Handle exceptions specific to Skyflow operations
fmt.Println(bearerToken);
}
```

### Generate scoped bearer tokens
A service account with multiple roles can generate bearer tokens with access limited to a specific role by specifying the appropriate `roleID`. It can be used to limit access to specific roles for services with multiple responsibilities, such as segregating access for billing vs. analytics. The generated bearer tokens are valid for 60 minutes and can only execute operations permitted by the permissions associated with the designated role.

[Example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/serviceaccount/scoped_token_generation.go):
```go
import (
"fmt"
saUtil "github.com/skyflowapi/skyflow-go/v2/serviceaccount"
"github.com/skyflowapi/skyflow-go/v2/utils/common"
"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * Example program to generate a Scoped Token using Skyflow's BearerToken utility.
 * The token is generated by providing the file path to the credentials.json file
 * and specifying roles associated with the token.
 */
func ScopedTokenGenerationExample() {
// Variable to store the generated scoped token
var scopedToken interface{}

// Example: Generate Scoped Token by specifying the credentials.json file path
// Create a list of roles that the generated token will be scoped to
var roles = []string{"<ROLE_ID_1>", "<ROLE_ID_2>", "<ROLE_ID_3>"}

// Specify the full file path to the service account's credentials.json file
var filePath = "<YOUR_CREDENTIALS_FILE_PATH>"

// Create a BearerToken using the credentials file and associated roles
res, err := saUtil.GenerateBearerToken(filePath, common.BearerTokenOptions{LogLevel: logger.DEBUG, RoleIDs: roles}) // Set the roles that the token should be scoped to

if err != nil {
fmt.Println("Errors", *err)
} else {
// retrieve token
fmt.Println("Token", res.AccessToken)
}

// Retrieve the generated scoped token
scopedToken = res.AccessToken

// Print the generated scoped token to the console
fmt.Println(scopedToken);
}
```
Notes:
- You can pass either the file path of a service account key credentials file or the service account key credentials as a string to the methods `GenerateBearerToken` and `GenerateBearerTokenFromCreds`.

### Generate signed data tokens
Skyflow generates data tokens when sensitive data is inserted into the vault. These data tokens can be digitally signed with a service account's private key, adding an extra layer of protection. Signed tokens can only be detokenized by providing the signed data token along with a bearer token generated from the service account's credentials. The service account must have the necessary permissions and context to successfully detokenize the signed data tokens.

[Example](https://github.com/skyflowapi/skyflow-go/blob/v2/samples/serviceaccount/signed_token_generation.go):
```go
import (
"fmt"
saUtil "github.com/skyflowapi/skyflow-go/v2/serviceaccount"
"github.com/skyflowapi/skyflow-go/v2/utils/common"
"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

// Example program to generate Signed Data Tokens using Skyflow's SignedDataTokens utility.
// Signed Data Tokens can be generated in two ways:
// 1. By specifying the file path to the credentials.json file.
// 2. By providing the credentials as a JSON string.
func SignedTokenGenerationExample() {
// Example 1: Generate Signed Data Tokens by specifying the credentials.json file path
// File path to the service account's credentials.json file
var filePath = "<YOUR_CREDENTIALS_FILE_PATH>";

// Context value to associate with the token
var context = "abc";

var tokens []string
tokens = append(tokens, "<TOKEN>") // List of data tokens to sign; replace with your actual data tokens

// Build the SignedDataTokensOptions object using the file path and required configurations
options := common.SignedDataTokensOptions{
Ctx: context, // Set the context value
DataTokens: tokens,  // Set the data tokens to be signed
TimeToLive: 60, // Set the token's time-to-live (TTL) in seconds
LogLevel: logger.ERROR,
}
// Generate and retrieve the signed data tokens
res, err := saUtil.GenerateSignedDataTokens(filePath, options)
if err != nil {
fmt.Println("Error occurred ", err)
} else {
// retrieve the signed data tokens 
fmt.Println("RESPONSE:", res)
}

// Example 2: Generate Signed Data Tokens by specifying credentials as a JSON string
// Provide the credentials JSON content as a string
var fileContents = "<YOUR_CREDENTIALS_FILE_CONTENTS_AS_STRING>";

// Context value to associate with the token
context = "abc";

tokens = nil
tokens = append(tokens, "<TOKEN>")

// Create the SignedDataTokensOptions object using the required configurations
options = common.SignedDataTokensOptions{
DataTokens: tokens,  // Set the data tokens to be signed
TimeToLive: 60, // in seconds
LogLevel: logger.ERROR,
}
// Generate and retrieve the signed data tokens
res, err = saUtil.GenerateSignedDataTokensFromCreds(fileContents, options)

if err != nil {
fmt.Println("Error occurred ", err)
} else {
// retrieve the signed data tokens 
fmt.Println("RESPONSE: ", res)
}
}
```

Response:
```json
[
  {
    "Token":"5530-4316-0674-5748",
    "signedToken":"signed_token_eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJzLCpZjA"
  }
]
```

Notes:
- The **time to live (TTL)** value should be specified in seconds.
- By default, the TTL value is set to 60 seconds.

## Logging
The Skyflow Go SDK provides useful logging using go's built-in logging library. By default, the SDK's logging level is set to `LogLevel.ERROR`. This can be changed using the UpdateLogLevel(logLevel) method, as shown below:

Currently, the following five log levels are supported:
- `DEBUG`:
  When `LogLevel.DEBUG` is passed, logs at all levels will be printed (DEBUG, INFO, WARN, ERROR).
- `INFO`:
  When `LogLevel.INFO` is passed, INFO logs for every event that occurs during SDK flow execution will be printed, along with WARN and ERROR logs.
- `WARN`:
  When `LogLevel.WARN` is passed, only WARN and ERROR logs will be printed.
- `ERROR`:
  When `LogLevel.ERROR` is passed, only ERROR logs will be printed.
- `OFF`:
  `LogLevel.OFF` can be used to turn off all logging from the Skyflow Go SDK.

`Note`: The ranking of logging levels is as follows: `DEBUG` < `INFO` < `WARN` < `ERROR` < `OFF`.

```go
package main

import (
  "fmt"
  "github.com/skyflowapi/skyflow-go/v2/client"
  "github.com/skyflowapi/skyflow-go/v2/utils/common"
  "github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

/**
 * This example demonstrates how to configure the Skyflow client with custom log levels
 * and authentication credentials (either token, credentials string, or other methods).
 * It also shows how to configure a vault connection using specific parameters.
 *
 * 1. Set up credentials with a Bearer token or credentials string.
 * 2. Define the Vault configuration.
 * 3. Build the Skyflow client with the chosen configuration and set log level.
 * 4. Example of changing the log level from ERROR (default) to INFO.
 */
func main() {
  // Step 1: Set up credentials - either pass token or use credentials string
  // In this case, we are using a Bearer token for authentication.
  credentials := common.Credentials{
    Token: "<BEARER_TOKEN>",        // Replace with the actual Bearer token
  }
  // Step 2: Define the Vault configuration
  // Configure the vault with necessary details like vault ID, cluster ID, and environment
  config := common.VaultConfig{
    VaultId:   "<VAULT_ID>",            // Replace with the actual Vault ID (first vault)
    ClusterId: "<CLUSTER_ID>",          // Replace with the actual Cluster ID (from vault URL)
    Env:       common.DEV,               // Set the environment (default is DEV, can also use PROD)
    Credentials: credentials,
  }
  credentialString := "<CREDENTIAL_AS_JSON_STRING>"
  skyflowCredentials := common.Credentials{
    CredentialsString: credentialString,
  }
  // Step 2: Define the Vault configuration
  // Create an array of Vault configurations to be used for multiple Vaults.
  var arr []common.VaultConfig
  arr = append(arr, config) // Add Vault configurations to the array

  // Step 3: Build the Skyflow client with the chosen configuration and log level
  // Using the Vault configurations and setting the log level to DEBUG.
  skyflowClient, err := client.NewSkyflow(
    client.WithVaults(arr...),              // Add the Vault configurations from the array
    client.WithCredentials(skyflowCredentials), // Use Skyflow credentials if no token is passed
    client.WithLogLevel(logger.INFO),     // Set log level to INFO (default is ERROR)
  )

  // Step 4: Handle any errors that occur during client creation
  if err != nil {
    // Print the error if something went wrong during client initialization
    fmt.Println("Error occurred while creating Skyflow client:", err)
  } else {
    skyflowClient.UpdateLogLevel(logger.DEBUG)
  }

  // Step 5: Client is now ready to use with the specified log level and credentials
  fmt.Println("Skyflow client has been successfully configured with log level: DEBUG.")
}
```


## Reporting a Vulnerability
If you discover a potential security issue in this project, please reach out to us at security@skyflow.com. Please refrain from creating public GitHub issues or pull requests, as malicious actors could potentially view them.