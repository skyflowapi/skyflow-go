package skyflow_test

import (
	"github.com/stretchr/testify/assert"
	"skyflow-go/client"
	vaultutils "skyflow-go/utils/common"
	"skyflow-go/utils/logger"
	"testing"
)

// TO DO
func TestSkyflowClient(t *testing.T) {
	skyflow := client.Skyflow{}
	// initialise skyflow client
	skyflowClient, err := skyflow.Builder().WithVaultConfig(
		vaultutils.VaultConfig{
			VaultId:     "vaultID",
			ClusterId:   "clusterID",
			Env:         vaultutils.DEV,
			Credentials: vaultutils.Credentials{},
		}).WithLogLevel(logger.INFO).Build()
	assert.Nil(t, err)
	assert.NotNil(t, skyflowClient)
	//vaultService, err := skyflowClient.Vault()
}
