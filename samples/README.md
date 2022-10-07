# GO-SDK sample templates
Use this folder to test the functionalities of GO-SDK just by adding `VAULT-ID` `VAULT-URL` and `SERVICE-ACCOUNT` details at the required place.

## Prerequisites
- A Skylow account. If you don't have one, you can register for one on the [Try Skyflow](https://skyflow.com/try-skyflow) page.
- go 1.15 and above

## Configure
- Before you can run the sample app, create a vault
- Navigate to `samples/vaultapi` and run the following command :

        go get
        

### Create the vault
1. In a browser, navigate to Skyflow Studio and log in.
2. Create a vault by clicking **Create Vault** > **Start With a Template** > **Quickstart vault**.
3. Once the vault is created, click the gear icon and select **Edit Vault** Details.
4. Note your Vault URL and Vault ID values, then click Cancel. You'll need these later.


### Create a service account
1. In the side navigation click, **IAM** > **Service Accounts** > **New Service Account**.
2. For Name, enter **Test-Go-Sdk-Sample**. For Roles, choose Roles corresponding to the action.
3. Click **Create**. Your browser downloads a **credentials.json** file. Keep this file secure, as you'll need it in the next steps.

### Different types of functionalities of Go-Sdk
- [**detokenize**](vaultapi/detokenize.go)
    - Detokenize the data token from the vault. 
    - Make sure the token is of the data which exists in the Vault. If not so please make use of [insert.go](insert.go) to insert the data in the data and use this token for detokenization.
    - Configure
        - Replace **<vault_id>** with **VAULT ID**
        - Replace **<vault_url>** with **VAULT URL**.
        - Replace **<token>** with data token of the data present in the vault.
        - Replace **<file_path>** with relative  path of **SERVICE ACCOUNT CREDENTIAL FILE**.
    - Execution
            
            go run detokenize.go
- [**get_by_id**](vaultapi/get_by_id.go)
    - Get data using skyflow id. 
    - Configure
        - Replace **<vault_id>** with **VAULT ID**
        - Replace **<vault_url>** with **VAULT URL**.
        - Replace **<id1>** with **Skyflow Id 1**.
        - Replace **<id2>** with **Skyflow Id 2**.
        - Replace **<file_path>** with relative  path of **SERVICE ACCOUNT CREDENTIAL FILE**.
    - Execution
        
            go run get_by_id.go
- [**insert**](vaultapi/insert.go)
    - Insert data in the vault.
    - Configure
        - Replace **<vault_id>** with **VAULT ID**.
        - Replace **<vault_url>** with **VAULT URL**.
        - Replace **<file_path>** with relative  path of **SERVICE ACCOUNT CREDENTIAL FILE**.
        - Execution
                
                go run insert.go
- [**invoke_connection**](vaultapi/invoke_connection.go)
    - Invoke connection
    - Configure
        - Replace **<vault_id>** with **VAULT ID**.
        - Replace **<vault_url>** with **VAULT URL**.
        - Replace **<file_path>** with relative  path of **SERVICE ACCOUNT CREDENTIAL FILE**.
        - Replace **pathParams** data with required params by the connection url.
        - Replace **<connection_url>** with **Connection url**.
        - Give **<Your-Authorization-Value>** value as the tokens.
        - Replace key and value pair of **requestBody** with your's request body content.

        - Execution
            
                go run invoke_connection.go
- [**service_account_token**](serviceaccount/token/main/service_account_token.go)
    - generates SA Token using path of credentials file.
    - Configure
        - Replace **<file_path>** with relative  path of **SERVICE ACCOUNT CREDENTIAL FILE**.

        - Execution
                
                go run service_account_token.go
- [**service_account_token_using_cred_string**](serviceaccount/token/main/service_account_token_using_cred_String.go)
    - generates SA Token using path of credentials file.
    - Configure
        - Replace **<credentials_in_string_format>** with relative  path of **SERVICE ACCOUNT CREDENTIAL IN STRING FORMAT**.

        - Execution
                
                go run service_account_token_using_cred_string.go