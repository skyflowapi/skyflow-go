package logs

import (
	. "github.com/skyflowapi/skyflow-go/v2/internal/constants"
)

const (
	EMPTY_BEARER_TOKEN                 = SDK_LOG_PREFIX + "BearerToken is Empty"
	BEARER_TOKEN_EXPIRED               = SDK_LOG_PREFIX + "BearerToken is expired"
	GENERATE_BEARER_TOKEN_TRIGGERED    = SDK_LOG_PREFIX + "GenerateBearerToken is triggered"
	GENERATE_BEARER_TOKEN_SUCCESS      = SDK_LOG_PREFIX + "BearerToken is generated"
	GENERATE_SIGNED_DATA_TOKEN_SUCCESS = SDK_LOG_PREFIX + "Signed Data tokens are generated"

	VALIDATING_VAULT_CONFIG                  = SDK_LOG_PREFIX + "Validating vault config."
	VALIDATING_CONNECTION_CONFIG             = SDK_LOG_PREFIX + "Validating connection config."
	VAULT_CONTROLLER_INITIALIZED             = SDK_LOG_PREFIX + "Initialized vault controller with vault ID %s."
	CONNECTION_CONTROLLER_INITIALIZED        = SDK_LOG_PREFIX + "Initialized connection controller with connection ID %s."
	VALIDATING_CRED                   string = SDK_LOG_PREFIX + "Validating skyflow credentials."
	VAULT_ID_CONFIG_DOES_NOT_EXIST           = SDK_LOG_PREFIX + "Vault config with vault ID %s doesn't exist."

	CURRENT_LOG_LEVEL         = SDK_LOG_PREFIX + "Current log level is %v."
	CLIENT_INITIALIZED        = SDK_LOG_PREFIX + "Initialized skyflow client successfully."
	VALIDATE_INSERT_INPUT     = SDK_LOG_PREFIX + "Validating insert request."
	VALIDATE_DETOKENIZE_INPUT = SDK_LOG_PREFIX + "Validating detokenize request."
	VALIDATE_TOKENIZE_INPUT   = SDK_LOG_PREFIX + "Validating tokenize request."

	VALIDATE_GET_INPUT         = SDK_LOG_PREFIX + "Validating get method request."
	VALIDATE_QUERY_INPUT       = SDK_LOG_PREFIX + "Validating query method request."
	VALIDATE_DELETE_INPUT      = SDK_LOG_PREFIX + "Validating delete method request."
	VALIDATE_UPDATE_INPUT      = SDK_LOG_PREFIX + "Validating update method request."
	VALIDATE_CONNECTION_CONFIG = SDK_LOG_PREFIX + "Validating connection config."
	INSERT_DATA_SUCCESS        = SDK_LOG_PREFIX + "Data inserted."

	GET_SUCCESS      = SDK_LOG_PREFIX + "Data revealed."
	UPDATE_SUCCESS   = SDK_LOG_PREFIX + "Data updated."
	DELETE_SUCCESS   = SDK_LOG_PREFIX + "Data deleted."
	TOKENIZE_SUCCESS = SDK_LOG_PREFIX + "Data tokenized."
	QUERY_SUCCESS    = SDK_LOG_PREFIX + "Query executed."

	REUSE_BEARER_TOKEN = SDK_LOG_PREFIX + "Reusing bearer token."
	REUSE_API_KEY      = SDK_LOG_PREFIX + "Reusing api key."

	INSERT_TRIGGERED     = SDK_LOG_PREFIX + "Insert method triggered."
	DETOKENIZE_TRIGGERED = SDK_LOG_PREFIX + "Detokenize method triggered."
	TOKENIZE_TRIGGERED   = SDK_LOG_PREFIX + "Tokenize method triggered."

	GET_TRIGGERED   = SDK_LOG_PREFIX + "Get call triggered."
	QUERY_TRIGGERED = SDK_LOG_PREFIX + "Query call triggered."

	INVOKE_CONNECTION_TRIGGERED = SDK_LOG_PREFIX + "Invoke connection triggered."
	DELETE_TRIGGERED            = SDK_LOG_PREFIX + "Delete method Triggered"
	DELETE_REQUEST_RESOLVED     = SDK_LOG_PREFIX + "Delete method is resolved"

	QUERY_REQUEST_RESOLVED = SDK_LOG_PREFIX + "Query method is resolved"

	DETOKENIZE_REQUEST_RESOLVED = SDK_LOG_PREFIX + "Detokenize request is resolved."

	INSERT_BATCH_REQUEST_RESOLVED         = SDK_LOG_PREFIX + "Insert request is resolved."
	GET_REQUEST_RESOLVED                  = SDK_LOG_PREFIX + "Get request is resolved."
	INVOKE_CONNECTION_REQUEST_RESOLVED    = SDK_LOG_PREFIX + "Invoke connection request resolved."
	GENERATE_SIGNED_DATA_TOKENS_TRIGGERED = SDK_LOG_PREFIX + "generateSignedDataTokens is triggered"
	UPDATE_TRIGGERED                      = SDK_LOG_PREFIX + "Update method triggered."
	UPDATE_REQUEST_RESOLVED               = SDK_LOG_PREFIX + "Update request is resolved."

	USING_BEARER_TOKEN = SDK_LOG_PREFIX + "Using token from credentials"
	USING_API_KEY      = SDK_LOG_PREFIX + "Using api key from credentials"

	VALIDATING_INVOKE_CONNECTION_REQUEST = SDK_LOG_PREFIX + "Validating invoke connection request."
	DEIDENTIFY_TEXT_TRIGGERED = SDK_LOG_PREFIX + "Deidentify text triggered."
    VALIDATE_DEIDENTIFY_TEXT_REQUEST = SDK_LOG_PREFIX + "Validating deidentify text request."
    DEIDENTIFY_TEXT_SUCCESS = SDK_LOG_PREFIX + "Text data de-identified."
    DEIDENTIFY_TEXT_REQUEST_RESOLVED = SDK_LOG_PREFIX + "Deidentify text request resolved."
    VALIDATE_REIDENTIFY_TEXT_REQUEST = SDK_LOG_PREFIX + "Validating reidentify text request."
    REIDENTIFY_TEXT_TRIGGERED = SDK_LOG_PREFIX + "Reidentify text method triggered."
    REIDENTIFY_TEXT_REQUEST_RESOLVED = SDK_LOG_PREFIX + "Reidentify text request resolved."
    DEIDENTIFY_FILE_TRIGGERED = SDK_LOG_PREFIX + "Deidentify file method triggered."
    VALIDATE_DEIDENTIFY_FILE_REQUEST = SDK_LOG_PREFIX + "Validating deidentify file request."
    DEIDENTIFY_FILE_REQUEST_RESOLVED = SDK_LOG_PREFIX + "Deidentify file request resolved."
    DEIDENTIFY_FILE_SUCCESS = SDK_LOG_PREFIX + "File deidentified successfully."
    GET_DETECT_RUN_TRIGGERED = SDK_LOG_PREFIX + "Get detect run method triggered."
    VALIDATE_GET_DETECT_RUN_REQUEST = SDK_LOG_PREFIX + "Validating get detect run request."
    REIDENTIFY_TEXT_SUCCESS = SDK_LOG_PREFIX + "Text data re-identified."
	UPLOAD_FILE_TRIGGERED               = SDK_LOG_PREFIX + "Upload file method triggered."
	VALIDATE_FILE_UPLOAD_INPUT          = SDK_LOG_PREFIX + "Validating file upload request."
	VALIDATE_UPLOAD_INPUT               = SDK_LOG_PREFIX + "Validating upload file request."
	UPLOAD_FILE_REQUEST_RESOLVED       = SDK_LOG_PREFIX + "Upload file request is resolved."
)
