# Authentication credentials options

> **Note:** A single `Credentials` struct accepts only one credential type. Passing more than one in the same struct is a validation error. When credentials are set at more than one level, vault-level credentials take precedence over client-level (common) credentials. Client-level credentials take precedence over the `SKYFLOW_CREDENTIALS` environment variable.

1. **API keys**

   A unique identifier used to authenticate and authorize requests to an API.

   ```go
   credentials := common.Credentials{
    ApiKey: "<API_KEY>",
  }
   ```

2. **Bearer tokens**

   A temporary access token used to authenticate API requests, typically included in the
   Authorization header.

   ```go
   credentials := common.Credentials{
    Token: "<YOUR_BEARER_TOKEN>",
    }
   ```

3. **Service account credentials file path**

   The file path pointing to a JSON file containing credentials for a service account, used
   for secure API access.

   ```go
   credentials := common.Credentials{
    Path: "<YOUR_CREDENTIALS_FILE_PATH>",
    }
   ```

4. **Service account credentials string**

   JSON-formatted string containing service account credentials, often used as an alternative to a file for programmatic authentication.

   ```go
   credentials := common.Credentials{
       CredentialsString: os.Getenv("SKYFLOW_CREDENTIALS_JSON_STRING"),
   }
   ```

5. **Environment variables**

   If no credentials are explicitly provided, the SDK automatically looks for the `SKYFLOW_CREDENTIALS` environment variable. This variable must contain a JSON string like one of the examples above.