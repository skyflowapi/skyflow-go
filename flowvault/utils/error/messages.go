package errors

import internal "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"

const (
	VAULT_ID_ALREADY_IN_CONFIG_LIST string = internal.SDK_PREFIX + " Validation error. VaultId is present in an existing config. Specify a new vaultId in config."
	VAULT_ID_NOT_IN_CONFIG_LIST     string = internal.SDK_PREFIX + " Validation error. VaultId is missing from the config. Specify the vaultIds from configs."
	EMPTY_VAULT_CONFIG              string = internal.SDK_PREFIX + " Validation error. No vault configurations available"
	VAULT_ID_EXISTS_IN_CONFIG_LIST  string = internal.SDK_PREFIX + " Validation error. %s already exists in the config list. Specify a new vaultId."
	EMPTY_VAULT_ID                  string = internal.SDK_PREFIX + " Validation error. Invalid vault ID. Specify a valid vault ID."
)
