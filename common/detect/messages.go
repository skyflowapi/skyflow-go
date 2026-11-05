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
	// firstClashSuffix numbers the second output that would get a used name: -2, then -3, ...
	firstClashSuffix   = 2
	extensionSeparator = "."
	fileTypeSeparator  = "_"
	// Temporary files are hidden and end in .tmp: .<output name>.<random>.tmp
	tempFilePrefix = "."
	tempFileSuffix = ".*.tmp"
)

// Message bodies without the SDK prefix; the controller prepends the owning SDK's prefix
// (same pattern as auth, config and vaultclient). Messages containing %s / %d are
// formatted with fmt.Sprintf.

// Error messages: prefixed with the SDK prefix. The wording is the Java SDK's, with Go's
// field names.
const (
	errInvalidTextDeidentifyString = " Validation error. Invalid text in deidentify string request. Specify a non-empty text."
	errInvalidTextReidentifyString = " Validation error. Invalid text in reidentify string request. Specify a non-empty text."
	errInvalidTextCheckGuardrails  = " Validation error. Invalid text in check guardrails request. Specify a non-empty text."
	errMissingFileSource           = " Validation error. Specify one of FilePath, FileObject or Value in deidentify file request."
	errMultipleFileSources         = " Validation error. Specify only one of FilePath, FileObject or Value in deidentify file request."
	errInvalidDeidentifyDataSource = " Validation error. Invalid DataSource in deidentify file request. Specify a DeidentifySourceType."
	errInvalidDeidentifyValue      = " Validation error. Invalid Value in deidentify file request. Specify the non-empty content or reference that DataSource describes."
	errFileNotFound                = " Validation error. The file to deidentify was not found at the specified path. Verify the file path and try again."
	errFileNotReadable             = " Validation error. The file to deidentify is not readable. Check the file permissions and try again."
	errInvalidWaitTime             = " Validation error. WaitTime in Poll must be between 1 and 300 seconds."
	errInvalidMaxAttempts          = " Validation error. MaxAttempts in Poll must be between 1 and 23."
	errOutputDirectoryNeedPoll     = " Validation error. OutputDirectory in deidentify file options requires Poll."
	errEmptyRunId                  = " Validation error. Invalid RunId in get run request. Specify a non-empty run id."
	errInvalidReidentifyDataSource = " Validation error. Invalid DataSource in reidentify file request. Specify a ReidentifySourceType."
	errInvalidReidentifyValue      = " Validation error. Invalid Value in reidentify file request. Specify the non-empty base64 content or presigned URL of the file."
	// errOutputNotWritten is the Java text with the cause appended.
	errOutputNotWritten = " Failed to write the processed output to %s: %s"
	errEmptyResponse    = " API error. %s returned an empty response."

	// Cause appended to errOutputNotWritten.
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
