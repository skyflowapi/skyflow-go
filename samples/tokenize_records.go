package main

import (
    "context"
    "fmt"
    "github.com/skyflowapi/skyflow-go/flowservice"
    "github.com/skyflowapi/skyflow-go/api"
)

// tokenizeRecords tokenizes sensitive data.
func tokenizeRecords(client *flowservice.Client) {
    vaultID := "<VAULT_ID>"

    // Data to tokenize
    data := []*api.V1FlowTokenizeRequestObject{
        {
            Value:          "<VALUE_1>",
            DataType:       api.FlowEnumDataTypeString.Ptr(),
            TokenGroupNames: []string{"<TOKEN_GROUP_1>"},
        },
    }

    // Create the tokenize request
    request := &api.V1FlowTokenizeRequest{
        VaultId: &vaultID,
        Data:    data,
    }

    // Call the Tokenize function
    ctx := context.Background()
    response, err := client.Tokenize(ctx, request)
    if err != nil {
        fmt.Println("Error during tokenize:", err)
        return
    }

    fmt.Println("Tokenize response:", response)
}