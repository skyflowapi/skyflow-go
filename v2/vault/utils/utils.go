package utils

import "context"

type VaultConfig struct {
	VaultId     string
	ClusterId   string
	Env         string
	Credentials string
}

type VaultConfigDetails struct {
	// vault-specific details
}

type ConnectionConfig struct {
	// connection-specific details here
	ConnectionId  string
	ConnectionUrl string
}

type Credentials struct {
	// credentials-related fields here
	Path              string
	Roles             []string
	Context           context.Context
	CredentialsString string
	Token             string
}

type InsertRequest struct {
	Table  string
	Values []map[string]string
}

type InsertResponse struct {
	// Response fields
}

type DetokenizeRequest struct {
	Tokens        []string
	RedactionType string
}

type DetokenizeResponse struct {
	// Response fields
	Tokens string
}

type DeleteRequest struct {
	Table string
	Ids   []string
}

type DeleteResponse struct {
	// Response fields
}

type UpdateRequest struct {
	Table string
	Data  []map[string]string
}

type UpdateResponse struct {
	// Response fields
}

type GetRequest struct {
	Table         string
	Ids           []string
	RedactionType string
}

type GetResponse struct {
	// Response fields
}

type UploadFileRequest struct {
	TableName  string
	DummyId    string
	ColumnName string
	FilePath   string
}

type UploadFileResponse struct {
	// Response fields
}

type InsertOptions struct {
	ReturnTokens bool
	Upsert       string
	Homogenes    bool
	TokenMode    bool
	TokenStrict  string
}

type DetokenizeOptions struct {
	ContinueOnError bool
}

type UpdateOptions struct {
	ReturnTokens bool
}

type DeleteOptions struct {
	// Optional fields
}
