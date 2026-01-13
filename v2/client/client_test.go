package client

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	error "github.com/skyflowapi/skyflow-go/v2/utils/error"
	"github.com/skyflowapi/skyflow-go/v2/utils/logger"
)

func TestServiceAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client initialisation Suite")
}

var _ = Describe("Skyflow Client", func() {
	var client *Skyflow
	var logLevel logger.LogLevel
	var credentials common.Credentials

	BeforeEach(func() {
		// Initialize mock values before each test
		logLevel = logger.INFO
		credentials = common.Credentials{
			CredentialsString: "some-credentials",
		}
		var err *error.SkyflowError
		var vaultArr []common.VaultConfig
		vaultArr = append(vaultArr, common.VaultConfig{
			VaultId:   "id",
			ClusterId: "cluster1",
			Env:       0,
		})
		var connArr []common.ConnectionConfig
		connArr = append(connArr, common.ConnectionConfig{
			ConnectionId:  "id1",
			ConnectionUrl: "https://url",
		})

		client, err = NewSkyflow(
			WithLogLevel(logLevel),
			WithVaults(
				vaultArr...),
			WithConnections(connArr...),
			WithCredentials(credentials),
		)
		Expect(err).Should(BeNil())
	})

	Context("when initializing the Skyflow client", func() {
		It("should initialize with default configurations", func() {
			Expect(client).NotTo(BeNil())
			Expect(client.GetLoglevel()).To(Equal(&logLevel))
		})
		It("should return error when WithVaults called with nil vault array", func() {
			var nilConfig []common.VaultConfig = nil
			client, err := NewSkyflow(WithVaults(nilConfig...))
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_VAULT_CONFIG))
		})

		It("should return error when WithVaults called with no parameters", func() {
			client, err := NewSkyflow(WithVaults())
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_VAULT_CONFIG))
		})

		It("should return error when WithVaults called with empty vault array", func() {
			var config []common.VaultConfig = make([]common.VaultConfig, 0)
			client, err := NewSkyflow(WithVaults(config...))
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_VAULT_CONFIG))
		})
		It("should return error when initialize with configuration with incorrect vault config", func() {
			var config []common.VaultConfig
			config = append(config, common.VaultConfig{
				VaultId:   "",
				ClusterId: "cluster1",
				Env:       0,
			})
			client, err := NewSkyflow(
				WithVaults(config...))
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.INVALID_VAULT_ID))
		})
		It("should return error when initialize with custom invalid url", func() {
			var config []common.VaultConfig
			config = append(config, common.VaultConfig{
				VaultId:      "id",
				ClusterId:    "cluster1",
				Env:          0,
				BaseVaultURL: "invalid-url",
			})
			client, err := NewSkyflow(
				WithVaults(config...))
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.INVALID_VAULT_URL))
		})
		It("should initialize THE CLIENT with configuration with vault config", func() {
			var config []common.VaultConfig
			config = append(config, common.VaultConfig{
				VaultId:   "id",
				ClusterId: "cluster1",
				Env:       0,
			})
			client, err := NewSkyflow(
				WithVaults(config...))
			Expect(client).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should initialize THE CLIENT with configuration with vault config and custom headers", func() {
			var config []common.VaultConfig
			config = append(config, common.VaultConfig{
				VaultId:   "id",
				ClusterId: "cluster1",
				Env:       0,
			})
			customHeader := make(map[string]string)
			customHeader["x-custom-header"] = "custom-header-value"
			client, err := NewSkyflow(
				WithVaults(config...),
				WithCustomHeaders(customHeader),
			)
			Expect(client).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should return error when initialize with configuration with nil connection config array", func() {
			var nilConfig []common.ConnectionConfig = nil
			client, err := NewSkyflow(WithConnections(nilConfig...))
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_CONFIG))
		})

		It("should return error when initialize with configuration with empty connection config array", func() {
			emptyConfig := make([]common.ConnectionConfig, 0)
			client, err := NewSkyflow(WithConnections(emptyConfig...))
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_CONFIG))
		})

		It("should return error when WithConnections called with no parameters", func() {
			client, err := NewSkyflow(WithConnections())
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_CONFIG))
		})
		It("should return error when initialize with configuration with incorrect connection config config", func() {
			var config []common.ConnectionConfig
			config = append(config, common.ConnectionConfig{
				ConnectionId:  "",
				ConnectionUrl: "https://url",
			})
			client, err := NewSkyflow(
				WithConnections(config...))
			Expect(client).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_ID))
		})
		It("should initialize THE CLIENT with configuration with connection config config", func() {
			var config []common.ConnectionConfig
			config = append(config, common.ConnectionConfig{
				ConnectionId:  "ID",
				ConnectionUrl: "https://url",
			})
			client, err := NewSkyflow(
				WithConnections(config...))
			Expect(client).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())
		})

	})

	Context("when adding Vault and Connection Configs", func() {
		var vaultConfig common.VaultConfig
		var connectionConfig common.ConnectionConfig

		BeforeEach(func() {
			vaultConfig = common.VaultConfig{
				VaultId:   "vault2",
				ClusterId: "id",
			}
			connectionConfig = common.ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "http://url",
			}
		})

		It("should successfully add a vault configuration", func() {
			err := client.AddVault(vaultConfig)
			Expect(err).Should(BeNil())
			vault, err := client.GetVault(vaultConfig.VaultId)
			Expect(err).Should(BeNil())
			Expect(vault).NotTo(BeNil())

		})

		It("should return an error when adding a duplicate vault configuration", func() {
			err := client.AddVault(vaultConfig)
			Expect(err).Should(BeNil())
			err = client.AddVault(vaultConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(error.VAULT_ID_EXISTS_IN_CONFIG_LIST, vaultConfig.VaultId)))
			err = client.AddVault(common.VaultConfig{
				VaultId: "",
			})
			Expect(err).ShouldNot(BeNil())
		})

		It("should successfully add a connection configuration", func() {
			err := client.AddConnection(connectionConfig)
			Expect(err).Should(BeNil())
			connection, err := client.GetConnection(connectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			Expect(connection).NotTo(BeNil())
		})

		It("should return an error when adding a duplicate connection configuration", func() {
			err := client.AddConnection(connectionConfig)
			Expect(err).Should(BeNil())
			err2 := client.AddConnection(connectionConfig)
			Expect(err2).ShouldNot(BeNil())

			err2 = client.AddConnection(common.ConnectionConfig{})
			Expect(err2).ShouldNot(BeNil())
		})
	})

	Context("when removing Vault and Connection Configs", func() {
		var vaultConfig common.VaultConfig
		var connectionConfig common.ConnectionConfig

		BeforeEach(func() {
			vaultConfig = common.VaultConfig{
				VaultId:   "vault1",
				ClusterId: "id",
			}
			connectionConfig = common.ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "http://url",
			}
			client.AddVault(vaultConfig)
			client.AddConnection(connectionConfig)
		})

		It("should successfully remove a vault configuration", func() {
			err := client.RemoveVault(vaultConfig.VaultId)
			Expect(err).Should(BeNil())
			_, err = client.GetVault(vaultConfig.VaultId)
			Expect(err).ShouldNot(BeNil())
		})

		It("should return an error when removing a non-existing vault configuration", func() {
			err := client.RemoveVault("non-existing-vault")
			Expect(err).ShouldNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.VAULT_ID_NOT_IN_CONFIG_LIST))
		})

		It("should successfully remove a connection configuration", func() {
			err := client.RemoveConnection(connectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			_, err = client.Connection(connectionConfig.ConnectionId)
			Expect(err).ShouldNot(BeNil())
		})

		It("should return an error when removing a non-existing connection configuration", func() {
			err := client.RemoveConnection("non-existing-conn")
			Expect(err).ShouldNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.CONNECTION_ID_NOT_IN_CONFIG_LIST))
		})
	})

	Context("when updating configurations", func() {
		var updatedVaultConfig common.VaultConfig
		var updatedConnectionConfig common.ConnectionConfig

		BeforeEach(func() {
			updatedVaultConfig = common.VaultConfig{
				VaultId:   "vault1",
				ClusterId: "demo",
			}
			updatedConnectionConfig = common.ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "http://url",
			}
			client.AddVault(updatedVaultConfig)
			client.AddConnection(updatedConnectionConfig)
		})

		It("should successfully update a vault configuration and service", func() {
			updatedVaultConfig.ClusterId = "demo"

			err := client.UpdateVault(updatedVaultConfig)
			Expect(err).Should(BeNil())
			// SHOULD RETURRN ERROR
			err = client.UpdateVault(common.VaultConfig{})
			Expect(err).ShouldNot(BeNil())

			vault, err := client.GetVault(updatedVaultConfig.VaultId)
			Expect(err).Should(BeNil())
			Expect(vault.ClusterId).To(Equal("demo"))

			service, err := client.Vault(vault.VaultId)
			Expect(err).Should(BeNil())
			Expect(service).NotTo(BeNil())

			service1, err1 := client.Vault("1")
			Expect(err1).ShouldNot(BeNil())
			Expect(service1).To(BeNil())
		})

		It("should return an error when trying to update a non-existing vault configuration", func() {
			nonExistingConfig := common.VaultConfig{
				VaultId:   "non-existing-vault",
				ClusterId: "demo",
			}
			err := client.UpdateVault(nonExistingConfig)
			Expect(err).ShouldNot(BeNil())
		})

		It("should successfully update a connection configuration", func() {
			_ = client.AddConnection(updatedConnectionConfig)
			updatedConnectionConfig.ConnectionUrl = "http://conn-updated"
			err := client.UpdateConnection(updatedConnectionConfig)
			Expect(err).Should(BeNil())
			conn, err := client.GetConnection(updatedConnectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			Expect(conn.ConnectionUrl).To(ContainSubstring("conn-updated"))
			service, err := client.Connection(updatedConnectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			Expect(service).NotTo(BeNil())

			service1, err1 := client.Connection("2")
			Expect(err1).ShouldNot(BeNil())
			Expect(service1).To(BeNil())

			conn1, err1 := client.GetConnection("not")
			Expect(err1).ShouldNot(BeNil())
			Expect(conn1).To(BeNil())

			service2, err2 := client.Connection()
			Expect(err2).Should(BeNil())
			Expect(service2).NotTo(BeNil())
		})

		It("should return an error when trying to update a non-existing connection configuration", func() {
			nonExistingConfig := common.ConnectionConfig{
				ConnectionId: "non-existing-conn",
			}
			err := client.UpdateConnection(nonExistingConfig)
			Expect(err).ShouldNot(BeNil())
		})

		It("should return error a connection configuration", func() {
			client1, err := NewSkyflow(
				WithVaults(common.VaultConfig{
					VaultId:   "id",
					ClusterId: "id",
				},
				),
				WithCredentials(common.Credentials{}),
			)
			Expect(client1).To(BeNil())
			Expect(err).ShouldNot(BeNil())

			client1, err = NewSkyflow(
				WithConnections(common.ConnectionConfig{}))
			Expect(client1).To(BeNil())
			Expect(err).ShouldNot(BeNil())

		})
	})

	Context("when update loglevel", func() {
		It("should successfully update a loglevel", func() {
			client.UpdateLogLevel(logger.DEBUG)
			level := client.GetLoglevel()
			Expect(*level).Should(Equal(logger.DEBUG))
		})
		It("should successfully update a config and add config", func() {
			err := client.UpdateSkyflowCredentials(common.Credentials{
				Token: "token",
			})
			Expect(err).Should(BeNil())

			errr := client.AddSkyflowCredentials(common.Credentials{
				Token: "token1",
			})
			Expect(errr).Should(BeNil())

			// should return error when invalid cred passed
			errr1 := client.AddSkyflowCredentials(common.Credentials{Token: "token"})
			Expect(errr1).Should(BeNil())
		})
		It("should successfully update a config and remove config", func() {
			err := client.UpdateSkyflowCredentials(common.Credentials{})
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("client creation error", func() {
		It("should return an error when trying to create a new client", func() {
			client1, err := NewSkyflow(
				WithVaults(common.VaultConfig{
					VaultId: "vault1",
				}),
			)
			Expect(client1).Should(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error when trying to create a new client with vault config", func() {
			client1, err := NewSkyflow(
				WithVaults(common.VaultConfig{
					VaultId:   "vault1",
					ClusterId: "demo",
				}),
				WithVaults(common.VaultConfig{
					VaultId:     "vault1",
					ClusterId:   "demo",
					Env:         0,
					Credentials: common.Credentials{},
				}),
			)
			Expect(client1).Should(BeNil())
			Expect(err).To(HaveOccurred())
		})
		It("should return an error when trying to create a new client with connection config", func() {
			client1, err := NewSkyflow(
				WithConnections(common.ConnectionConfig{
					ConnectionId:  "conn1",
					ConnectionUrl: "http://url",
				}),
				WithConnections(common.ConnectionConfig{
					ConnectionId:  "conn1",
					ConnectionUrl: "http://url",
				}),
			)
			Expect(client1).Should(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("should return an error when trying to create a new client with connection config when validates fails", func() {
			client1, err := NewSkyflow(
				WithConnections(common.ConnectionConfig{
					ConnectionId:  "conn1",
					ConnectionUrl: "http://url",
				}),
				WithConnections(common.ConnectionConfig{
					ConnectionId: "conn1",
				}),
			)
			Expect(client1).Should(BeNil())
			Expect(err).To(HaveOccurred())
		})

	})
})

var _ = Describe("Detect and getDetectConfig scenarios", func() {
	It("should return error if no detect configs exist", func() {
		client, _ := NewSkyflow()
		service, err := client.Detect("any")
		Expect(service).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	It("should return error if detect config not found by ID", func() {
		vaultCfg := common.VaultConfig{VaultId: "v1", ClusterId: "c1"}
		client, _ := NewSkyflow(WithVaults(vaultCfg))
		service, err := client.Detect("notfound")
		Expect(service).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	It("should return detect service if found by ID", func() {
		vaultCfg := common.VaultConfig{VaultId: "v2", ClusterId: "c2", Credentials: common.Credentials{Token: "t"}}
		client, _ := NewSkyflow(WithVaults(vaultCfg))
		service, err := client.Detect("v2")
		Expect(err).To(BeNil())
		Expect(service).NotTo(BeNil())
	})

	It("should return first detect service if called with no ID", func() {
		vaultCfg1 := common.VaultConfig{VaultId: "v3", ClusterId: "c3", Credentials: common.Credentials{Token: "t1"}}
		vaultCfg2 := common.VaultConfig{VaultId: "v4", ClusterId: "c4", Credentials: common.Credentials{Token: "t2"}}
		client, _ := NewSkyflow(WithVaults(vaultCfg1, vaultCfg2))
		service, err := client.Detect()
		Expect(err).To(BeNil())
		Expect(service).NotTo(BeNil())
	})

	It("should use client credentials if detect config has empty credentials, client credentials not empty", func() {
		vaultCfg := common.VaultConfig{VaultId: "v6", ClusterId: "c6"}
		creds := common.Credentials{Token: "token"}
		client, _ := NewSkyflow(WithVaults(vaultCfg), WithCredentials(creds))
		service, err := client.Detect("v6")
		Expect(err).To(BeNil())
		Expect(service).NotTo(BeNil())
	})
})
var _ = Describe("Skyflow Management Methods", func() {
	var client *Skyflow
	var vaultConfig common.VaultConfig
	var connConfig common.ConnectionConfig
	var creds common.Credentials

	BeforeEach(func() {
		vaultConfig = common.VaultConfig{
			VaultId:   "vault1",
			ClusterId: "cluster1",
			Env:       common.PROD,
			Credentials: common.Credentials{
				ApiKey: os.Getenv("API_KEY"),
			},
		}
		connConfig = common.ConnectionConfig{
			ConnectionId:  "conn1",
			ConnectionUrl: "https://example.com",
			Credentials: common.Credentials{
				ApiKey: os.Getenv("API_KEY"),
			},
		}
		creds = common.Credentials{ApiKey: os.Getenv("API_KEY")}
		client, _ = NewSkyflow()
	})

	Context("AddVault", func() {
		It("should add a vault successfully", func() {
			err := client.AddVault(vaultConfig)
			Expect(err).To(BeNil())
			Expect(client.vaultServices).To(HaveKey("vault1"))
		})
		It("should not add duplicate vault", func() {
			client.AddVault(vaultConfig)
			err := client.AddVault(vaultConfig)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("AddConnection", func() {
		It("should add a connection successfully", func() {
			err := client.AddConnection(connConfig)
			Expect(err).To(BeNil())
			Expect(client.connectionServices).To(HaveKey("conn1"))
		})
		It("should not add duplicate connection", func() {
			client.AddConnection(connConfig)
			err := client.AddConnection(connConfig)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("AddSkyflowCredentials", func() {
		It("should add credentials successfully", func() {
			err := client.AddSkyflowCredentials(creds)
			Expect(err).To(BeNil())
			Expect(client.credentials).To(Equal(&creds))
		})
		It("should fail with invalid credentials", func() {
			invalidCreds := common.Credentials{}
			err := client.AddSkyflowCredentials(invalidCreds)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("GetVault", func() {
		It("should get vault config", func() {
			client.AddVault(vaultConfig)
			cfg, err := client.GetVault("vault1")
			Expect(err).To(BeNil())
			Expect(cfg.VaultId).To(Equal("vault1"))
		})
		It("should fail for missing vault", func() {
			cfg, err := client.GetVault("missing")
			Expect(err).ToNot(BeNil())
			Expect(cfg).To(BeNil())
		})
	})

	Context("GetConnection", func() {
		It("should get connection config", func() {
			client.AddConnection(connConfig)
			cfg, err := client.GetConnection("conn1")
			Expect(err).To(BeNil())
			Expect(cfg.ConnectionId).To(Equal("conn1"))
		})
		It("should fail for missing connection", func() {
			cfg, err := client.GetConnection("missing")
			Expect(err).ToNot(BeNil())
			Expect(cfg).To(BeNil())
		})
	})

	Context("UpdateLogLevel", func() {
		It("should update log level", func() {
			client.UpdateLogLevel(logger.DEBUG)
			Expect(*client.GetLoglevel()).To(Equal(logger.DEBUG))
		})
	})

	Context("UpdateSkyflowCredentials", func() {
		It("should update credentials and propagate to controllers", func() {
			client.AddVault(vaultConfig)
			client.AddConnection(connConfig)
			err := client.UpdateSkyflowCredentials(creds)
			Expect(err).To(BeNil())
			Expect(client.credentials).To(Equal(&creds))
			// Check controllers updated
			for _, v := range client.vaultServices {
				if v.controller != nil {
					Expect(v.controller.Config.Credentials).To(Equal(creds))
				}
			}
		})
		It("should fail with invalid credentials", func() {
			invalidCreds := common.Credentials{}
			err := client.UpdateSkyflowCredentials(invalidCreds)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("UpdateVault", func() {
		It("should update vault config and propagate to controller", func() {
			client.AddVault(vaultConfig)
			updated := vaultConfig
			updated.ClusterId = "new-cluster"
			err := client.UpdateVault(updated)
			Expect(err).To(BeNil())
			Expect(client.vaultServices["vault1"].config.ClusterId).To(Equal("new-cluster"))
			if client.vaultServices["vault1"].controller != nil {
				Expect(client.vaultServices["vault1"].controller.Config.ClusterId).To(Equal("new-cluster"))
			}
		})
		It("should fail for missing vault", func() {
			updated := vaultConfig
			updated.VaultId = "missing"
			err := client.UpdateVault(updated)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("UpdateConnection", func() {
		It("should update connection config", func() {
			client.AddConnection(connConfig)
			updated := connConfig
			updated.ConnectionUrl = "https://new-url.com"
			err := client.UpdateConnection(updated)
			Expect(err).To(BeNil())
			Expect(client.connectionServices["conn1"].config.ConnectionUrl).To(Equal("https://new-url.com"))
		})
		It("should fail for missing connection", func() {
			updated := connConfig
			updated.ConnectionId = "missing"
			err := client.UpdateConnection(updated)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("GetLoglevel", func() {
		It("should get current log level", func() {
			client.UpdateLogLevel(logger.INFO)
			Expect(*client.GetLoglevel()).To(Equal(logger.INFO))
		})
	})

	Context("RemoveVault", func() {
		It("should remove vault", func() {
			client.AddVault(vaultConfig)
			err := client.RemoveVault("vault1")
			Expect(err).To(BeNil())
			Expect(client.vaultServices).ToNot(HaveKey("vault1"))
		})
		It("should fail for missing vault", func() {
			err := client.RemoveVault("missing")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("RemoveConnection", func() {
		It("should remove connection", func() {
			client.AddConnection(connConfig)
			err := client.RemoveConnection("conn1")
			Expect(err).To(BeNil())
			Expect(client.connectionServices).ToNot(HaveKey("conn1"))
		})
		It("should fail for missing connection", func() {
			err := client.RemoveConnection("missing")
			Expect(err).ToNot(BeNil())
		})
	})
	Context("cross scenario: Vault(vaultid) and update/add/remove", func() {
		var vaultConfig common.VaultConfig
		var creds common.Credentials
		BeforeEach(func() {
			vaultConfig = common.VaultConfig{
				VaultId:   "vaultX",
				ClusterId: "clusterX",
				Env:       common.PROD,
				Credentials: common.Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}
			creds = common.Credentials{ApiKey: os.Getenv("API_KEY")}
			var err *error.SkyflowError
			client, err = NewSkyflow(WithVaults(vaultConfig), WithCredentials(creds))
			Expect(err).To(BeNil())
		})
		It("should get vault by id, update, add, and remove", func() {
			vaultSvc, err := client.Vault("vaultX")
			Expect(err).To(BeNil())
			Expect(vaultSvc).ToNot(BeNil())
			Expect(vaultSvc.config.VaultId).To(Equal("vaultX"))

			updated := vaultConfig
			updated.ClusterId = "new-clusterX"
			err2 := client.UpdateVault(updated)
			Expect(err2).To(BeNil())
			vault, err3 := client.GetVault("vaultX")
			Expect(err3).To(BeNil())
			Expect(vault.ClusterId).To(Equal("new-clusterX"))

			vaultConfig2 := common.VaultConfig{
				VaultId:   "vaultY",
				ClusterId: "clusterY",
				Env:       common.PROD,
				Credentials: common.Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}
			err4 := client.AddVault(vaultConfig2)
			Expect(err4).To(BeNil())
			vault2, err5 := client.GetVault("vaultY")
			Expect(err5).To(BeNil())
			Expect(vault2.VaultId).To(Equal("vaultY"))

			err6 := client.RemoveVault("vaultX")
			Expect(err6).To(BeNil())
			_, err7 := client.GetVault("vaultX")
			Expect(err7).ToNot(BeNil())
		})
		It("should fail to get, update, or remove missing vault", func() {
			vaultSvc, err := client.Vault("missing")
			Expect(err).ToNot(BeNil())
			Expect(vaultSvc).To(BeNil())
			missing := vaultConfig
			missing.VaultId = "missing"
			err2 := client.UpdateVault(missing)
			Expect(err2).ToNot(BeNil())
			err3 := client.RemoveVault("missing")
			Expect(err3).ToNot(BeNil())
		})
	})

	Context("cross scenario: Detect(vaultid) and update/add/remove", func() {
		var vaultConfig common.VaultConfig
		var creds common.Credentials
		BeforeEach(func() {
			vaultConfig = common.VaultConfig{
				VaultId:   "vaultDX",
				ClusterId: "clusterDX",
				Env:       common.PROD,
				Credentials: common.Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}
			creds = common.Credentials{ApiKey: os.Getenv("API_KEY")}
			var err *error.SkyflowError
			client, err = NewSkyflow(WithVaults(vaultConfig), WithCredentials(creds))
			Expect(err).To(BeNil())
		})
		It("should get detect by id, update vault, add, and remove", func() {
			detectSvc, err := client.Detect("vaultDX")
			Expect(err).To(BeNil())
			Expect(detectSvc).ToNot(BeNil())
			Expect(detectSvc.config.VaultId).To(Equal("vaultDX"))

			updated := vaultConfig
			updated.ClusterId = "new-clusterDX"
			err2 := client.UpdateVault(updated)
			Expect(err2).To(BeNil())
			detect, err3 := client.Detect("vaultDX")
			Expect(err3).To(BeNil())
			Expect(detect.config.ClusterId).To(Equal("new-clusterDX"))

			vaultConfig2 := common.VaultConfig{
				VaultId:   "vaultDY",
				ClusterId: "clusterDY",
				Env:       common.PROD,
				Credentials: common.Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}
			err4 := client.AddVault(vaultConfig2)
			Expect(err4).To(BeNil())
			detect2, err5 := client.Detect("vaultDY")
			Expect(err5).To(BeNil())
			Expect(detect2.config.VaultId).To(Equal("vaultDY"))

			err6 := client.RemoveVault("vaultDX")
			Expect(err6).To(BeNil())
			_, err7 := client.Detect("vaultDX")
			Expect(err7).ToNot(BeNil())
		})
		It("should fail to get, update, or remove missing detect", func() {
			detectSvc, err := client.Detect("missing")
			Expect(err).ToNot(BeNil())
			Expect(detectSvc).To(BeNil())
			missing := vaultConfig
			missing.VaultId = "missing"
			err2 := client.UpdateVault(missing)
			Expect(err2).ToNot(BeNil())
			err3 := client.RemoveVault("missing")
			Expect(err3).ToNot(BeNil())
		})
	})

	Context("cross scenario: Connection(connectionId) and update/add/remove", func() {
		var connConfig common.ConnectionConfig
		var creds common.Credentials
		BeforeEach(func() {
			connConfig = common.ConnectionConfig{
				ConnectionId:  "connX",
				ConnectionUrl: "https://connX.com",
				Credentials: common.Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}
			creds = common.Credentials{ApiKey: os.Getenv("API_KEY")}
			var err *error.SkyflowError
			client, err = NewSkyflow(WithConnections(connConfig), WithCredentials(creds))
			Expect(err).To(BeNil())
		})
		It("should get connection by id, update, add, and remove", func() {
			connSvc, err := client.Connection("connX")
			Expect(err).To(BeNil())
			Expect(connSvc).ToNot(BeNil())
			Expect(connSvc.config.ConnectionId).To(Equal("connX"))

			updated := connConfig
			updated.ConnectionUrl = "https://new-connX.com"
			err2 := client.UpdateConnection(updated)
			Expect(err2).To(BeNil())
			conn, err3 := client.GetConnection("connX")
			Expect(err3).To(BeNil())
			Expect(conn.ConnectionUrl).To(Equal("https://new-connX.com"))

			connConfig2 := common.ConnectionConfig{
				ConnectionId:  "connY",
				ConnectionUrl: "https://connY.com",
				Credentials: common.Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}
			err4 := client.AddConnection(connConfig2)
			Expect(err4).To(BeNil())
			conn2, err5 := client.GetConnection("connY")
			Expect(err5).To(BeNil())
			Expect(conn2.ConnectionId).To(Equal("connY"))

			err6 := client.RemoveConnection("connX")
			Expect(err6).To(BeNil())
			_, err7 := client.GetConnection("connX")
			Expect(err7).ToNot(BeNil())
		})
		It("should fail to get, update, or remove missing connection", func() {
			connSvc, err := client.Connection("missing")
			Expect(err).ToNot(BeNil())
			Expect(connSvc).To(BeNil())
			missing := connConfig
			missing.ConnectionId = "missing"
			err2 := client.UpdateConnection(missing)
			Expect(err2).ToNot(BeNil())
			err3 := client.RemoveConnection("missing")
			Expect(err3).ToNot(BeNil())
		})
	})
})
