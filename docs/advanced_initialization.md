# Advanced Skyflow Client Initialization

This guide demonstrates advanced initialization patterns for the Skyflow Python SDK, including multiple vault configurations and different credential types.

Use multiple vault configurations when your application needs to access data across different Skyflow vaults, such as managing data across different geographic regions or distinct environments.

To get started, you must first initialize the skyflow client. While initializing the skyflow client, you can specify different types of credentials.  

**1. API keys**
- A unique identifier used to authenticate and authorize requests to an API.

**2. Bearer tokens**
- A temporary access token used to authenticate API requests, typically included in the Authorization header.

**3. Service account credentials file path**
- The file path pointing to a JSON file containing credentials for a service account, used
   for secure API access.

**4. Service account credentials string (JSON formatted)**
- A JSON-formatted string containing service account credentials, often used as an alternative to a file for programmatic authentication.

Note: Only one type of credential can be used at a time.

```go
package main
import (
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
  primaryCredentials := common.Credentials {
    Token: "<BEARER_TOKEN1>",
  } // Replace <BEARER_TOKEN> with your actual authentication token.

  // Step 2: Configure the primary vault details.
  // VaultConfig stores all necessary details to connect to a specific Skyflow vault.
  primaryConfig: = common.VaultConfig {
    VaultId: "<PRIMARY_VAULT_ID>", // Replace with your primary vault's ID.
    ClusterId: "<CLUSTER_ID>", // Replace with the cluster ID (part of the vault URL, e.g., https://{clusterId}.vault.skyflowapis.com).
    Env: common.DEV, // Set the environment (PROD, SANDBOX, STAGE, DEV).
    Credentials: primaryCredentials, // Attach the primary credentials to this vault configuration.
  }

  // Step 3: Create credentials as a JSON object (if a Bearer Token is not provided).
  // Demonstrates an alternate approach to authenticate with Skyflow using a credentials object.
  credentialsObject := `<CREDS_JSON_OBJECT>`
  // Step 4: Use credentials string.
  skyflowCredentials = common.Credentials {
    CredentialsString: credentialsObject,
  }

  // Step 5: Define secondary credentials (API key-based authentication as an example).
  // Demonstrates a different type of authentication mechanism for Skyflow vaults.
  secondaryCredentials := common.Credentials {
    ApiKey: "<API_KEY>",
  } // Replace with your API Key for authentication.

  // Step 6: Configure the secondary vault details.
  // A secondary vault configuration can be used for operations involving multiple vaults.
  secondaryConfig := common.VaultConfig {
    VaultId: "<SECONDARY_VAULT_ID>", // Replace with your secondary vault's ID.
    ClusterId: "<CLUSTER_ID>", // Replace with the corresponding cluster ID.
    Env: common.SANDBOX, // Set the environment for this vault.
    Credentials: secondaryCredentials, // Attach the secondary credentials to this configuration.
  }

  // Step 7: Define tertiary credentials using a path to a credentials JSON file.
  // This method demonstrates an alternative authentication method.
  tertiaryCredentials := common.Credentials {
    Path: "<PATH_TO_YOUR_CREDENTIALS_JSON_FILE>",
  }

  // Step 8: Configure the tertiary vault details.
  tertiaryConfig := common.VaultConfig {
    VaultId: "<TERTIARY_VAULT_ID>", // Replace with your tertiary vault's ID.
    ClusterId: "<CLUSTER_ID>", // Replace with the corresponding cluster ID.
    Env: common.SANDBOX, // Set the environment for this vault.
    Credentials: tertiaryCredentials, // Attach the secondary credentials to this configuration.
  }
  // Step 9: Build and initialize the Skyflow client.
  // Skyflow client is configured with multiple vaults and credentials.

  var arr[] common.VaultConfig
  arr = append(arr, primaryConfig, secondaryConfig, tertiaryConfig)
  skyflowClient, err: = client.NewSkyflow(
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