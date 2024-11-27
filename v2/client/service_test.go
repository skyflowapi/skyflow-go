package client

import (
	"os"
	vaultutils "skyflow-go/v2/utils/common"
	"skyflow-go/v2/utils/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleClientWithOneVaultConfig(t *testing.T) {
	vaultConfig1 := vaultutils.VaultConfig{
		VaultId: "vault1",
		Credentials: vaultutils.Credentials{
			Path:  "validPath1",
			Token: "token1",
		},
	}

	skyflowClient1 := Skyflow{}
	builder1, _ := skyflowClient1.Builder().
		WithVaultConfig(vaultConfig1).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	vaultService1, err1 := builder1.Vault("vault1")
	assert.NoError(t, err1)
	assert.NotNil(t, vaultService1)
	assert.Equal(t, "validPath1", vaultService1.config.Credentials.Path)
}
func TestSingleClientUseWithSkyflowCredentials(t *testing.T) {
	vaultConfig1 := vaultutils.VaultConfig{
		VaultId: "vault1",
	}

	skyflowClient1 := Skyflow{}
	builder1, _ := skyflowClient1.Builder().
		WithVaultConfig(vaultConfig1).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	vaultService1, err1 := builder1.Vault("vault1")
	assert.NoError(t, err1)
	assert.NotNil(t, vaultService1)
	assert.Equal(t, "token1", vaultService1.config.Credentials.Token)
	assert.Equal(t, "vault1", vaultService1.config.VaultId)
}

func TestSingleClientUseEnv(t *testing.T) {
	vaultConfig1 := vaultutils.VaultConfig{
		VaultId: "vault1",
	}
	_ = os.Setenv("SKYFLOW_CREDENTIALS", "{myValue:myValue}")

	skyflowClient1 := Skyflow{}
	builder1, _ := skyflowClient1.Builder().
		WithVaultConfig(vaultConfig1).
		WithLogLevel(logger.WARN).Build()

	vaultService1, err1 := builder1.Vault("vault1")
	assert.NoError(t, err1)
	assert.NotNil(t, vaultService1)

	assert.Equal(t, "{myValue:myValue}", vaultService1.config.Credentials.CredentialsString)

	// without vaultid
	vaultService2, err2 := builder1.Vault("vault1")
	assert.NoError(t, err2)
	assert.NotNil(t, vaultService2)

	assert.Equal(t, "{myValue:myValue}", vaultService2.config.Credentials.CredentialsString)
	_ = os.Unsetenv("SKYFLOW_CREDENTIALS")
}

func TestSingleClientWithMultipleVaultConfigs(t *testing.T) {
	vaultConfig1 := vaultutils.VaultConfig{
		VaultId: "vault1",
		Credentials: vaultutils.Credentials{
			Path:  "validPath1",
			Token: "token1",
		},
	}
	vaultConfig3 := vaultutils.VaultConfig{
		VaultId: "vault3",
		Env:     vaultutils.STAGE,
		Credentials: vaultutils.Credentials{
			Path:  "validPath2",
			Token: "token2",
		},
	}

	skyflowClient1 := Skyflow{}
	builder1, _ := skyflowClient1.Builder().
		WithVaultConfig(vaultConfig1).
		WithVaultConfig(vaultConfig3).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	// Check the first VaultConfig credentials
	vaultService1, err1 := builder1.Vault("vault1")
	assert.NoError(t, err1)
	assert.NotNil(t, vaultService1)
	assert.Equal(t, "validPath1", vaultService1.config.Credentials.Path)

	// Check the second VaultConfig credentials
	vaultService3, err3 := builder1.Vault("vault3")
	assert.NoError(t, err3)
	assert.NotNil(t, vaultService3)
	assert.Equal(t, "validPath2", vaultService3.config.Credentials.Path)

	vaultService4, err4 := builder1.Vault()
	assert.NoError(t, err4)
	assert.NotNil(t, vaultService4)
	assert.Equal(t, "validPath1", vaultService4.config.Credentials.Path)
	// Clean up environment variables if necessary
	_ = os.Unsetenv("SKYFLOW_CREDENTIALS")
}

func TestTwoClientsWithOneVaultConfigEach(t *testing.T) {
	vaultConfig1 := vaultutils.VaultConfig{
		VaultId: "vault1",
		Credentials: vaultutils.Credentials{
			Path:  "validPath1",
			Token: "token1",
		},
	}
	vaultConfig2 := vaultutils.VaultConfig{
		VaultId: "vault2",
		Credentials: vaultutils.Credentials{
			Path:  "validPath2",
			Token: "token2",
		},
	}
	vaultConfig3 := vaultutils.VaultConfig{
		VaultId: "vault3",
		Credentials: vaultutils.Credentials{
			Path:  "validPath3",
			Token: "token3",
		},
	}

	// Create two clients, each with one VaultConfig
	skyflowClient1 := Skyflow{}
	builder1, _ := skyflowClient1.Builder().
		WithVaultConfig(vaultConfig1).
		WithVaultConfig(vaultConfig2).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	skyflowClient2 := Skyflow{}
	builder2, _ := skyflowClient2.Builder().
		WithVaultConfig(vaultConfig3).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token2"}).
		WithLogLevel(logger.ERROR).Build()

	// Check the credentials for Client 1
	vaultService1, err1 := builder1.Vault("vault1")
	assert.NoError(t, err1)
	assert.NotNil(t, vaultService1)
	assert.Equal(t, "validPath1", vaultService1.config.Credentials.Path)

	// Check the credentials for Client 2
	vaultService2, err2 := builder2.Vault("vault3")
	assert.NoError(t, err2)
	assert.NotNil(t, vaultService2)
	assert.Equal(t, "validPath3", vaultService2.config.Credentials.Path)

	// check client 1 without vaultapi id
	vaultService3, err3 := builder1.Vault()
	assert.NoError(t, err3)
	assert.NotNil(t, vaultService3)
	assert.Equal(t, "vault1", vaultService3.config.VaultId)
	assert.Equal(t, "validPath1", vaultService3.config.Credentials.Path)

	// check client 2 without vaultapi id
	vaultService4, err4 := builder2.Vault()
	assert.NoError(t, err4)
	assert.NotNil(t, vaultService4)
	assert.Equal(t, "validPath3", vaultService4.config.Credentials.Path)

	_ = os.Unsetenv("SKYFLOW_CREDENTIALS")
}

func TestTwoClientsWithMultipleVaultConfigsEach(t *testing.T) {
	vaultConfig1 := vaultutils.VaultConfig{
		VaultId: "vault1",
		Credentials: vaultutils.Credentials{
			Path:  "validPath1",
			Token: "token1",
		},
	}
	vaultConfig2 := vaultutils.VaultConfig{
		VaultId: "vault2",
		Credentials: vaultutils.Credentials{
			Path:  "validPath2",
			Token: "token2",
		},
	}

	// Client 1 with multiple VaultConfigs
	skyflowClient1 := Skyflow{}
	builder1, _ := skyflowClient1.Builder().
		WithVaultConfig(vaultConfig1).
		WithVaultConfig(vaultConfig2).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
		WithLogLevel(logger.WARN).Build()

	// Client 2 with multiple VaultConfigs
	vaultConfig3 := vaultutils.VaultConfig{
		VaultId: "vault3",
		Env:     vaultutils.STAGE,
		Credentials: vaultutils.Credentials{
			Path:  "validPath3",
			Token: "token3",
		},
	}
	vaultConfig4 := vaultutils.VaultConfig{
		VaultId: "vault4",
		Credentials: vaultutils.Credentials{
			Path:  "validPath4",
			Token: "token4",
		},
	}

	skyflowClient2 := Skyflow{}
	builder2, _ := skyflowClient2.Builder().
		WithVaultConfig(vaultConfig3).
		WithVaultConfig(vaultConfig4).
		WithSkyflowCredentials(vaultutils.Credentials{Token: "token2"}).
		WithLogLevel(logger.ERROR).Build()

	// Validate VaultConfig for Client 1
	vaultService1, err1 := builder1.Vault("vault1")
	assert.NoError(t, err1)
	assert.NotNil(t, vaultService1)
	assert.Equal(t, "validPath1", vaultService1.config.Credentials.Path)

	vaultService2, err2 := builder1.Vault("vault2")
	assert.NoError(t, err2)
	assert.NotNil(t, vaultService2)
	assert.Equal(t, "validPath2", vaultService2.config.Credentials.Path)

	// Validate VaultConfig for Client 2
	vaultService3, err3 := builder2.Vault("vault3")
	assert.NoError(t, err3)
	assert.NotNil(t, vaultService3)
	assert.Equal(t, "validPath3", vaultService3.config.Credentials.Path)

	vaultService4, err4 := builder2.Vault("vault4")
	assert.NoError(t, err4)
	assert.NotNil(t, vaultService4)
	assert.Equal(t, "validPath4", vaultService4.config.Credentials.Path)

	// Clean up environment variables if necessary
	_ = os.Unsetenv("SKYFLOW_CREDENTIALS")
}
