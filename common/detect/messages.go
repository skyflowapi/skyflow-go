package detect

// Method names used in logs and error messages, and operation names passed to interceptors.
const (
	methodDeidentifyString = "DeidentifyString"
	methodDeidentifyFile   = "DeidentifyFile"
	methodGetRun           = "GetRun"
	methodReidentifyString = "ReidentifyString"
	methodReidentifyFile   = "ReidentifyFile"
	methodCheckGuardrails  = "CheckGuardrails"

	OperationDeidentifyString = "DEIDENTIFY_STRING"
	OperationDeidentifyFile   = "DEIDENTIFY_FILE"
	OperationGetRun           = "GET_RUN"
	OperationReidentifyString = "REIDENTIFY_STRING"
	OperationReidentifyFile   = "REIDENTIFY_FILE"
	OperationCheckGuardrails  = "CHECK_GUARDRAILS"
)

// API paths, relative to the vault URL.
const (
	pathDeidentifyString = "/v2/detect/deidentify/string"
	pathDeidentifyFile   = "/v2/detect/deidentify/file"
	pathRuns             = "/v2/detect/runs/"
	pathReidentifyString = "/v2/detect/reidentify/string"
	pathReidentifyFile   = "/v2/detect/reidentify/file"
	pathGuardrails       = "/v2/detect/guardrails"
)

// Poll limits.
const (
	defaultWaitTimeSeconds = 60
	maxWaitTimeSeconds     = 300
	defaultMaxAttempts     = 23
	maxMaxAttempts         = 23
	maxPollDelaySeconds    = 32
)

// Output file naming.
const (
	processedFilePrefix = "processed-"
	defaultOutputName   = "output"
	binaryExtension     = "bin"
	nameSeparator       = "-"
	extensionSeparator  = "."
	fileTypeSeparator   = "_"
)

// Message bodies without the SDK prefix; the controller prepends the owning SDK's prefix
// (same pattern as auth, config and vaultclient). Messages containing %s / %d are
// formatted with fmt.Sprintf.

// Error messages: prefixed with the SDK prefix.
const (
	errEmptyText               = " Validation error. %s failed. Text must not be empty."
	errFileSource              = " Validation error. %s failed. Set exactly one of Value, FilePath or FileObject."
	errMissingDeidentifySource = " Validation error. %s failed. DataSource is required with Value. Specify BASE64, SKYFLOW_ID or PRESIGNED_URL."
	errMissingReidentifySource = " Validation error. %s failed. DataSource is required with Value. Specify BASE64 or PRESIGNED_URL."
	errLocalFileDataSource     = " Validation error. %s failed. A local file is sent as BASE64; DataSource must be empty or BASE64."
	errFileNotReadable         = " Validation error. %s failed. Could not read file: %s"
	errInvalidWaitTime         = " Validation error. DeidentifyFile failed. Poll WaitTime must be between 1 and 300 seconds."
	errInvalidMaxAttempts      = " Validation error. DeidentifyFile failed. Poll MaxAttempts must be between 1 and 23."
	errOutputDirectoryNeedPoll = " Validation error. DeidentifyFile failed. OutputDirectory requires Poll."
	errEmptyRunId              = " Validation error. GetRun failed. RunId must not be empty."
	errRedactionSource         = " Validation error. %s failed. Redaction level at index %d needs exactly one of TokenGroupName or EntityName."
	errRedactionReplacement    = " Validation error. %s failed. Redaction level at index %d needs exactly one of RedactionPattern or RedactionType."
	errEmptyDenyTopic          = " Validation error. CheckGuardrails failed. DenyTopics must not contain empty values."
	errOutputNotWritten        = " Error occurred. The run succeeded but its output could not be written to %s: %s"
	errEmptyResponse           = " API error. %s returned an empty response."

	// Cause appended to errFileNotReadable and errOutputNotWritten.
	notRegularFile = "not a regular file"
)

// Log messages: prefixed with the SDK log prefix.
const (
	logMethodTriggered     = "%s method triggered."
	logValidatingRequest   = "Validating %s request."
	logRequestResolved     = "%s request resolved."
	logRequestRejected     = "%s request resulted in failure."
	logPollStarted         = "Polling run %s for up to %v and %d attempts."
	logPollTerminal        = "Run %s finished with status %s after %d attempts."
	logPollBudgetExhausted = "Run %s is still %s after %d attempts and %v; returning its run ID."
	logPollCancelled       = "Polling run %s stopped by the context while %s."
	logOutputNotWritten    = "Run outputs are %s, not BASE64, so nothing is written."
	logOutputFileWritten   = "Wrote run output to %s."
)
