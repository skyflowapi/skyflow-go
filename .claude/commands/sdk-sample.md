---
name: sdk-sample
description: Generate a Skyflow Go SDK v2 sample file for a vault feature, detect operation, or service account operation. Compile-verified after creation.
context: fork
paths:
  - "samples/v2/**/*.go"
  - "v2/**/*.go"
exclude:
  - "**/internal/generated/**"
  - "**/vendor/**"
---

Create a Skyflow Go SDK v2 sample file demonstrating: $ARGUMENTS

## File placement

| Feature type | Directory |
|---|---|
| Vault ops (insert/get/update/delete/query/tokenize/detokenize) | `samples/v2/vaultapi/` |
| Service account auth (bearer token, signed data tokens) | `samples/v2/serviceaccount/` |
| Detect (deidentify text/file, reidentify, get run) | `samples/v2/detectapi/` |
| Connection (invoke) | `samples/v2/invoke_connection/` |

File name: `<feature_name>.go` (snake_case)

## Structure (follow this order)

```go
package main

import (
    "context"
    "fmt"

    "github.com/skyflowapi/skyflow-go/v2/client"
    "github.com/skyflowapi/skyflow-go/v2/utils/common"
    "github.com/skyflowapi/skyflow-go/v2/utils/logger"
    // add feature-specific imports as needed
)

func main() {
    // Step 1: Set up Skyflow vault credentials
    vaultConfig := common.VaultConfig{
        VaultId:   "<YOUR_VAULT_ID>",
        ClusterId: "<YOUR_CLUSTER_ID>",
        Env:       common.PROD,
        Credentials: common.Credentials{
            ApiKey: "<YOUR_API_KEY>",
            // OR: CredentialsString: "<YOUR_CREDENTIALS_STRING>"
            // OR: Path: "credentials.json"
        },
    }

    // Step 2: Configure the Skyflow client
    skyflowInstance, err := client.NewSkyflow(
        client.WithVaults(vaultConfig),
        client.WithLogLevel(logger.ERROR), // Use logger.ERROR in production
    )
    if err != nil {
        fmt.Println(*err)
        return
    }

    // Step 3: Get the vault/detect/connection service
    service, serviceErr := skyflowInstance.Vault("<YOUR_VAULT_ID>")
    if serviceErr != nil {
        fmt.Println(*serviceErr)
        return
    }

    ctx := context.TODO()
    // Step 4: Call the operation
    // ...

    // Step 5: Handle the response and errors
    // ...
}
```

## Rules

- Vault IDs / cluster IDs use placeholders: `"<YOUR_VAULT_ID>"`, `"<YOUR_CLUSTER_ID>"`
- Credential values use placeholders: `"<YOUR_API_KEY>"`, `"<YOUR_CREDENTIALS_STRING>"`
- Credentials file path: `"credentials.json"` (relative — no absolute paths)
- Never hardcode real tokens, IDs, or file paths — placeholder strings only
- Use `logger.ERROR` for production; comment that `logger.DEBUG` / `logger.INFO` are available for development
- Always check `err != nil` after `NewSkyflow` and after getting the service; print `*err` and return
- Keep under 100 lines

## After creating the file

Verify the sample compiles against the v2 module:
```bash
cd samples/v2 && go build ./<feature-directory>/... 2>&1 | tail -20
```

Report the file path and any compile errors.
