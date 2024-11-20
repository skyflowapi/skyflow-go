package client

import (
	"testing"

	vaultutils "skyflow-go/utils/common"
	"skyflow-go/utils/logger"

	"github.com/stretchr/testify/assert"
)

func TestSkyflowBuilder_ClientBuilder1(t *testing.T) {
	// Setup for first client
	vaultConfig1 := vaultutils.VaultConfig{VaultId: "vault1"}
	skyflowClient := Skyflow{}
	builder1, err1 := skyflowClient.Builder().
		WithVaultConfig(vaultConfig1).
		WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn1"}).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	assert.NoError(t, err1)
	assert.NotNil(t, builder1)

	assert.Equal(t, builder1.builder.logLevel, logger.WARN)

	expectedVaultConfigs1 := map[string]vaultutils.VaultConfig{
		"vault1": vaultConfig1,
	}
	assert.Equal(t, builder1.builder.vaultConfigs, expectedVaultConfigs1)

	// Check environment for builder1
	for vaultID, config := range builder1.builder.vaultConfigs {
		switch vaultID {
		case "vault1":
			assert.Equal(t, config.Env, vaultutils.PROD) // Default env if not set
		default:
			t.Errorf("Unexpected vault ID: %s", vaultID)
		}
	}
}

func TestSkyflowBuilder_ClientBuilder2(t *testing.T) {
	// Setup for second client
	vaultConfig2 := vaultutils.VaultConfig{VaultId: "vault2", Env: vaultutils.DEV}
	vaultConfig3 := vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.STAGE}
	skyflowClient2 := Skyflow{}
	builder2, err2 := skyflowClient2.Builder().
		WithVaultConfig(vaultConfig2).
		WithVaultConfig(vaultConfig3).
		WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn2"}).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token2"}).
		WithLogLevel(logger.ERROR).Build()

	assert.NoError(t, err2)
	assert.NotNil(t, builder2)

	assert.Equal(t, builder2.builder.logLevel, logger.ERROR)

	expectedVaultConfigs2 := map[string]vaultutils.VaultConfig{
		"vault2": vaultConfig2,
		"vault3": vaultConfig3,
	}
	assert.Equal(t, builder2.builder.vaultConfigs, expectedVaultConfigs2)

	for vaultID, config := range builder2.builder.vaultConfigs {
		switch vaultID {
		case "vault2":
			assert.Equal(t, config.Env, vaultutils.DEV)
		case "vault3":
			assert.Equal(t, config.Env, vaultutils.STAGE)
		default:
			t.Errorf("Unexpected vault ID: %s", vaultID)
		}
	}
}

func TestSkyflowBuilder_CompareClientBuilders(t *testing.T) {
	vaultConfig1 := vaultutils.VaultConfig{VaultId: "vault1"}
	vaultConfig2 := vaultutils.VaultConfig{VaultId: "vault2", Env: vaultutils.DEV}
	vaultConfig3 := vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.STAGE}

	skyflowClient := Skyflow{}
	builder1, _ := skyflowClient.Builder().
		WithVaultConfig(vaultConfig1).
		WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn1"}).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	skyflowClient2 := Skyflow{}
	builder2, _ := skyflowClient2.Builder().
		WithVaultConfig(vaultConfig2).
		WithVaultConfig(vaultConfig3).
		WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn2"}).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token2"}).
		WithLogLevel(logger.ERROR).Build()

	// Ensure builders are different
	assert.NotEqual(t, builder1, builder2)

	// Compare vault config lengths
	assert.NotEqual(t, len(builder1.builder.vaultConfigs), len(builder2.builder.vaultConfigs))

	// Ensure log levels are different between the two builders
	assert.NotEqual(t, builder1.builder.logLevel, builder2.builder.logLevel)
}

func TestSkyflowBuilder_DeleteFromVaultConfig(t *testing.T) {
	// Setup for deletion test
	vaultConfig1 := vaultutils.VaultConfig{VaultId: "vault1"}
	vaultConfig3 := vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.STAGE}

	skyflowClient := Skyflow{}
	builder1, _ := skyflowClient.Builder().
		WithVaultConfig(vaultConfig1).
		WithVaultConfig(vaultConfig3).
		WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn1"}).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	// Check initial vault configs
	initialVaultConfigs := map[string]vaultutils.VaultConfig{
		"vault1": vaultConfig1,
		"vault3": vaultConfig3,
	}
	assert.Equal(t, builder1.builder.vaultConfigs, initialVaultConfigs)

	// Delete a vault from the map
	delete(builder1.builder.vaultConfigs, "vault3")

	// Expected result after deletion
	expectedVaultConfigsAfterDeletion := map[string]vaultutils.VaultConfig{
		"vault1": vaultConfig1,
	}
	assert.Equal(t, builder1.builder.vaultConfigs, expectedVaultConfigsAfterDeletion)
}
