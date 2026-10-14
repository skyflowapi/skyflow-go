package common

// InsertRequest — flowvault specific, embeds common base.
type InsertRequest struct {
	BaseInsertRequest // Table, Values
	ContinueOnError   bool
}

// InsertOptions — flowvault specific, embeds common base.
// CustomHeaders uses plain string keys (v2 uses a typed CustomHeaderKey).
type InsertOptions struct {
	BaseInsertOptions // shared marker
	Upsert            string
	CustomHeaders     map[string]string
}

// InsertResponse — flowvault specific, embeds common base.
type InsertResponse struct {
	BaseInsertResponse // Errors
	InsertedFields     []map[string]interface{}
	Tokens             []map[string]interface{}
}
