---
name: sdk-sample
description: Generate a Go sample/example file for a feature or operation. Compile-verified after creation.
context: fork
paths:
  - "**/*.go"
  - "**/go.mod"
  # - "**/<EXT>"       # add other extensions as needed
exclude:
  - "**/vendor/**"
  - "**/internal/generated/**"   # replace with your generated-code path
  # - "**/<GENERATED_DIR>/**"    # add other auto-generated dirs as needed
---

Create a Go sample file demonstrating: $ARGUMENTS

## File placement

Infer the correct directory from the feature described in `$ARGUMENTS` and the existing project layout. Look for a `samples/`, `examples/`, or `cmd/` directory and follow the existing naming convention. If none exists, create the file at the project root or in a `examples/<feature>/` directory.

File name: `<feature_name>.go` (snake_case)

## Structure (follow this order)

```go
package main

import (
    "context"
    "fmt"
    "log"
    // add feature-specific imports as needed
)

func main() {
    ctx := context.Background()

    // Step 1: Configure the client
    // ...

    // Step 2: Call the operation
    // ...

    // Step 3: Handle the response
    // ...
    _ = ctx
}
```

## Rules

- Use placeholder values for any IDs, tokens, or credentials: `"<YOUR_API_KEY>"`, `"<YOUR_CLIENT_ID>"`
- Credentials or config read from environment: use `os.Getenv("ENV_VAR_NAME")` with a comment
- Never hardcode real tokens, keys, or file paths — placeholder strings only
- Always check errors and handle them (`log.Fatal` or `fmt.Fprintf(os.Stderr, ...)` is acceptable in `main`)
- Keep under 100 lines where possible
- Add a short comment at the top explaining what the sample demonstrates

## After creating the file

Verify the sample compiles:
```bash
go build ./<sample-directory>/... 2>&1 | tail -20
```

Report the file path and any compile errors. If compile errors exist, fix them before reporting completion.
