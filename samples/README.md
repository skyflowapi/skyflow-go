# Go SDK samples
Test the SDK by adding your `VAULT_ID`, `VAULT_URL`, and `SERVICE-ACCOUNT `details as the corresponding values in each sample.

## Prerequisites
- Sign in to your Skyflow account:
    * For trial environments, use https://try.skyflow.com/ .
    * For sandbox and production environments, use your dedicated sign-in URL.
  If you don't have an account,  [sign up for a free trial account](https://skyflow.com/try-skyflow).
- go 1.15 or higher

## Get started
- Navigate to `samples/vaultapi` and run the following command :

        go get
        

### Create a vault
1. Sign in to Skyflow Studio.
2. Click Create Vault > Start With a Template. 
3. Under Quickstart, click Create.

To run the following commands, you'll need to retrieve your vault-specific values, <VAULT_URL> and <VAULT_ID>. From your vault page, click the gear icon and select Edit Vault Details. Create a service account

### Create a service account
1. In Studio, click **Settings** in the upper navigation.
2. In the side navigation, click **Vault**, then choose the **Quickstart** vault from the dropdown menu.
3. Under **IAM**, click  **Service Accounts > New Service Account**.
4. For **Name**, enter "SDK Sample". For **Roles**, choose **Vault Editor**.
5. Click **Create**. 
6. Your browser downloads a credentials.json file. Keep this file secure. You'll need it to generate bearer tokens.

## SDK samples
### [Detokenize data](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/detokenize.go)
This sample demonstrates how to detokenize a data token from the vault. Make sure the token you specify exists in the vault. If you need a valid token for detokenization, use insert.go to insert the records and return a data token.

1. Replace **<vault_id>** and **<vault_url>** with your vault-specific values. 
2. Replace **<token>** with the data token you want to detokenize..
3. Replace **<file_path>** with the relative path for your service account credentials file downloaded while #Create a service account . 

#### Run the following command:
            
        go run detokenize.go

### [Get a record by ID](https://github.com/skyflowapi/skyflow-go/tree/main/samples/vaultapi)

Get data using skyflow id. 
#### Configure
1. Replace **<vault_id>** and **<vault_url>** with your vault-specific values. 
2. Replace **<skyflow_id1>** and **<skyflow_id2>** with the Skyflow IDs you want to retrieve.
3. Replace **<file_path>** with the relative path for your service account credentials file downloaded while #Create a service account . 
#### Run the following command:
        
        go run get_by_id.go
### [Insert data into a vault](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/insert.go)
Insert data in the vault.
1. Replace **<vault_id>** and **<vault_url>** with your vault-specific values. 
3. Replace **<file_path>** with the relative path for your service account credentials file downloaded while #Create a service account . 

#### Run the following command:
                
        go run insert.go
### [Invoke a connection](https://github.com/skyflowapi/skyflow-go/blob/main/samples/vaultapi/invoke_connection.go)
Skyflow Connections is a gateway service that uses Skyflow's underlying tokenization capabilities to securely connect to first-party and third-party services. This way, you never expose your infrastructure to sensitive records, and you offload security and compliance requirements to Skyflow.
1. Replace **<vault_id>** and **<vault_url>** with your vault-specific values.
2. Replace **<file_path>** with the relative path for your service account credentials.
3. Replace `pathParams` data with the connection URL params.
4. Replace **<your_connection_url>** with the Connection URL value.
5. Enter the token values.
6. Replace the requestBody key and value pair with your request body content.

#### Run the following command:
    
        go run invoke_connection.go

### [Generate a service account bearer token from a file](https://github.com/skyflowapi/skyflow-go/blob/main/samples/serviceaccount/token/main/service_account_token.go)
Generates a service account bearer token using the path of a credentials file.
1. Replace **<file_path>** with the relative path for your service account credentials file downloaded while #Create a service account.


#### Run the following command:

        go run service_account_token.go

### [Generate a service account bearer token from a credentials string](https://github.com/skyflowapi/skyflow-go/blob/main/samples/serviceaccount/token/main/service_account_token.go) 
Generates service account bearer token using the JSON content of a credentials file.
#### Configure
1. Replace **<file_path>** with the relative path for your service account credentials file downloaded while #Create a service account.

#### Run the following command:
        
       go run service_account_token_using_cred_string.go