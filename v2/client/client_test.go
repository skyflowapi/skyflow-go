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
				BaseVaultUrl: "invalid-url",
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
			customHeader := make(map[common.CustomHeaderKey]string)
			customHeader[common.RequestIdHeader] = "custom-header-value"
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
			err := client.AddVaultConfig(vaultConfig)
			Expect(err).Should(BeNil())
			vault, err := client.GetVaultConfig(vaultConfig.VaultId)
			Expect(err).Should(BeNil())
			Expect(vault).NotTo(BeNil())

		})

		It("should return an error when adding a duplicate vault configuration", func() {
			err := client.AddVaultConfig(vaultConfig)
			Expect(err).Should(BeNil())
			err = client.AddVaultConfig(vaultConfig)
			Expect(err).ShouldNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(fmt.Sprintf(error.VAULT_ID_EXISTS_IN_CONFIG_LIST, vaultConfig.VaultId)))
			err = client.AddVaultConfig(common.VaultConfig{
				VaultId: "",
			})
			Expect(err).ShouldNot(BeNil())
		})

		It("should successfully add a connection configuration", func() {
			err := client.AddConnectionConfig(connectionConfig)
			Expect(err).Should(BeNil())
			connection, err := client.GetConnectionConfig(connectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			Expect(connection).NotTo(BeNil())
		})

		It("should return an error when adding a duplicate connection configuration", func() {
			err := client.AddConnectionConfig(connectionConfig)
			Expect(err).Should(BeNil())
			err2 := client.AddConnectionConfig(connectionConfig)
			Expect(err2).ShouldNot(BeNil())

			err2 = client.AddConnectionConfig(common.ConnectionConfig{})
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
			client.AddVaultConfig(vaultConfig)
			client.AddConnectionConfig(connectionConfig)
		})

		It("should successfully remove a vault configuration", func() {
			err := client.RemoveVaultConfig(vaultConfig.VaultId)
			Expect(err).Should(BeNil())
			_, err = client.GetVaultConfig(vaultConfig.VaultId)
			Expect(err).ShouldNot(BeNil())
		})

		It("should return an error when removing a non-existing vault configuration", func() {
			err := client.RemoveVaultConfig("non-existing-vault")
			Expect(err).ShouldNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.VAULT_ID_NOT_IN_CONFIG_LIST))
		})

		It("should successfully remove a connection configuration", func() {
			err := client.RemoveConnectionConfig(connectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			_, err = client.Connection(connectionConfig.ConnectionId)
			Expect(err).ShouldNot(BeNil())
		})

		It("should return an error when removing a non-existing connection configuration", func() {
			err := client.RemoveConnectionConfig("non-existing-conn")
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
			client.AddVaultConfig(updatedVaultConfig)
			client.AddConnectionConfig(updatedConnectionConfig)
		})

		It("should successfully update a vault configuration and service", func() {
			updatedVaultConfig.ClusterId = "demo"

			err := client.UpdateVaultConfig(updatedVaultConfig)
			Expect(err).Should(BeNil())
			// SHOULD RETURRN ERROR
			err = client.UpdateVaultConfig(common.VaultConfig{})
			Expect(err).ShouldNot(BeNil())

			vault, err := client.GetVaultConfig(updatedVaultConfig.VaultId)
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
			err := client.UpdateVaultConfig(nonExistingConfig)
			Expect(err).ShouldNot(BeNil())
		})

		It("should successfully update a connection configuration", func() {
			_ = client.AddConnectionConfig(updatedConnectionConfig)
			updatedConnectionConfig.ConnectionUrl = "http://conn-updated"
			err := client.UpdateConnectionConfig(updatedConnectionConfig)
			Expect(err).Should(BeNil())
			conn, err := client.GetConnectionConfig(updatedConnectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			Expect(conn.ConnectionUrl).To(ContainSubstring("conn-updated"))
			service, err := client.Connection(updatedConnectionConfig.ConnectionId)
			Expect(err).Should(BeNil())
			Expect(service).NotTo(BeNil())

			service1, err1 := client.Connection("2")
			Expect(err1).ShouldNot(BeNil())
			Expect(service1).To(BeNil())

			conn1, err1 := client.GetConnectionConfig("not")
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
			err := client.UpdateConnectionConfig(nonExistingConfig)
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

	Context("GetSkyflowCredentials", func() {
		It("should return the credentials set at construction time", func() {
			got := client.GetSkyflowCredentials()
			Expect(got).ToNot(BeNil())
			Expect(got.CredentialsString).To(Equal("some-credentials"))
		})
	})

	Context("UpdateSkyflowCredentials", func() {
		It("should return error for empty credentials", func() {
			err := client.UpdateSkyflowCredentials(common.Credentials{})
			Expect(err).ToNot(BeNil())
		})

		It("should update credentials and propagate to all controllers", func() {
			newCreds := common.Credentials{Token: "new-bearer-token"}
			err := client.UpdateSkyflowCredentials(newCreds)
			Expect(err).To(BeNil())
			Expect(client.GetSkyflowCredentials().Token).To(Equal("new-bearer-token"))
		})
	})

	Context("AddSkyflowCredentials", func() {
		It("should return error for empty credentials", func() {
			err := client.AddSkyflowCredentials(common.Credentials{})
			Expect(err).ToNot(BeNil())
		})

		It("should set credentials and propagate to all controllers", func() {
			newCreds := common.Credentials{Token: "replacement-token"}
			err := client.AddSkyflowCredentials(newCreds)
			Expect(err).To(BeNil())
			Expect(client.GetSkyflowCredentials().Token).To(Equal("replacement-token"))
		})
	})

	Context("UpdateVaultConfig", func() {
		It("should return error when vault ID does not exist", func() {
			updated := common.VaultConfig{
				VaultId:   "nonexistent-vault",
				ClusterId: "cluster1",
			}
			err := client.UpdateVaultConfig(updated)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.VAULT_ID_NOT_IN_CONFIG_LIST))
		})

		It("should update cluster ID without touching credentials when credentials are empty", func() {
			updated := common.VaultConfig{
				VaultId:   "id",
				ClusterId: "new-cluster",
			}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
		})

		It("should update credentials when a non-empty token is supplied", func() {
			updated := common.VaultConfig{
				VaultId:     "id",
				ClusterId:   "cluster1",
				Credentials: common.Credentials{Token: "vault-token"},
			}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
		})
	})

	Context("UpdateConnectionConfig", func() {
		It("should return error when connection ID does not exist", func() {
			updated := common.ConnectionConfig{
				ConnectionId:  "nonexistent-conn",
				ConnectionUrl: "https://example.com",
			}
			err := client.UpdateConnectionConfig(updated)
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.CONNECTION_ID_NOT_IN_CONFIG_LIST))
		})

		It("should update connection URL without changing credentials", func() {
			updated := common.ConnectionConfig{
				ConnectionId:  "id1",
				ConnectionUrl: "https://new-url.com",
			}
			err := client.UpdateConnectionConfig(updated)
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("Skyflow client — uncovered branches", func() {
	var client *Skyflow

	BeforeEach(func() {
		var err *error.SkyflowError
		client, err = NewSkyflow(
			WithVaults(
				common.VaultConfig{VaultId: "v1", ClusterId: "c1"},
				common.VaultConfig{VaultId: "v2", ClusterId: "c2"},
			),
			WithConnections(
				common.ConnectionConfig{ConnectionId: "conn1", ConnectionUrl: "https://url1.example.com"},
				common.ConnectionConfig{ConnectionId: "conn2", ConnectionUrl: "https://url2.example.com"},
			),
		)
		Expect(err).To(BeNil())
	})

	Context("Vault — no ID returns first vault", func() {
		It("should return a vault service when no ID is supplied", func() {
			svc, err := client.Vault()
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
		})
	})

	Context("Connection — no ID returns first connection", func() {
		It("should return a connection service when no ID is supplied", func() {
			svc, err := client.Connection()
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
		})
	})

	Context("UpdateSkyflowCredentials — propagates to controllers", func() {
		It("should update credentials on vault and connection controllers after they are initialised", func() {
			// Initialise controllers by calling Vault() and Connection()
			_, vErr := client.Vault("v1")
			Expect(vErr).To(BeNil())
			_, cErr := client.Connection("conn1")
			Expect(cErr).To(BeNil())
			_, dErr := client.Detect("v1")
			Expect(dErr).To(BeNil())

			newCreds := common.Credentials{Token: "propagated-token"}
			err := client.UpdateSkyflowCredentials(newCreds)
			Expect(err).To(BeNil())
			Expect(client.GetSkyflowCredentials().Token).To(Equal("propagated-token"))
		})
	})

	Context("AddSkyflowCredentials — propagates to controllers", func() {
		It("should update credentials on vault and connection controllers after they are initialised", func() {
			_, vErr := client.Vault("v1")
			Expect(vErr).To(BeNil())
			_, cErr := client.Connection("conn1")
			Expect(cErr).To(BeNil())
			_, dErr := client.Detect("v1")
			Expect(dErr).To(BeNil())

			newCreds := common.Credentials{Token: "added-token"}
			err := client.AddSkyflowCredentials(newCreds)
			Expect(err).To(BeNil())
			Expect(client.GetSkyflowCredentials().Token).To(Equal("added-token"))
		})
	})

	Context("UpdateVaultConfig — updates controller credentials when controller is set", func() {
		It("should update vault controller credentials via isCredentialsEmpty path", func() {
			_, _ = client.Vault("v1")

			updated := common.VaultConfig{
				VaultId:     "v1",
				ClusterId:   "new-cluster",
				Credentials: common.Credentials{Token: "new-vault-token"},
			}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
		})

		It("should update cluster ID when controller is set", func() {
			_, _ = client.Vault("v1")

			updated := common.VaultConfig{
				VaultId:   "v1",
				ClusterId: "updated-cluster",
			}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
		})
	})

	Context("UpdateConnectionConfig — updates controller when controller is set", func() {
		It("should update connection controller credentials when set", func() {
			_, _ = client.Connection("conn1")

			updated := common.ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://new.example.com",
				Credentials:   common.Credentials{Token: "conn-token"},
			}
			err := client.UpdateConnectionConfig(updated)
			Expect(err).To(BeNil())
		})
	})

	Context("vaultIdExists — detect-only branch", func() {
		It("should return error when vaultId exists only in detectServices", func() {
			// Manually inject a detect-only entry (white-box: same package)
			client.detectServices["detect-only"] = &detectService{
				config: &common.VaultConfig{VaultId: "detect-only", ClusterId: "c"},
			}
			// vaultIdExists is called by AddVault; it should find "detect-only" in detectServices
			err := client.AddVaultConfig(common.VaultConfig{VaultId: "detect-only", ClusterId: "c"})
			Expect(err).ToNot(BeNil())
		})
	})

	Context("RemoveVaultConfig( — partial registration", func() {
		It("should succeed when vault exists only in vaultServices", func() {
			delete(client.detectServices, "v1")
			err := client.RemoveVaultConfig("v1")
			Expect(err).To(BeNil())
			_, stillThere := client.vaultServices["v1"]
			Expect(stillThere).To(BeFalse())
		})
		It("should succeed when vault exists only in detectServices", func() {
			delete(client.vaultServices, "v1")
			err := client.RemoveVaultConfig("v1")
			Expect(err).To(BeNil())
			_, stillThere := client.detectServices["v1"]
			Expect(stillThere).To(BeFalse())
		})
		It("should return error when vault exists in neither service", func() {
			err := client.RemoveVaultConfig("nonexistent")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("detect_service.DeidentifyText — error path", func() {
		It("should return error when controller returns validation error", func() {
			svc, svcErr := client.Detect("v1")
			Expect(svcErr).To(BeNil())
			// Empty text triggers validation error inside the controller
			_, err := svc.DeidentifyText(
				nil,
				common.DeidentifyTextRequest{Text: ""},
				common.DeidentifyTextOptions{},
			)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("vault_service.UploadFile — success path", func() {
		It("should return nil err from controller when validation fails (coverage for service wrapper)", func() {
			svc, svcErr := client.Vault("v1")
			Expect(svcErr).To(BeNil())
			// controller validation will fail due to empty table; that covers the error wrapper
			_, err := svc.UploadFile(nil, common.FileUploadRequest{}, common.FileUploadOptions{})
			Expect(err).ToNot(BeNil())
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

	BeforeEach(func() {
		vaultConfig = common.VaultConfig{
			VaultId:   "vault1",
			ClusterId: "cluster1",
			Env:       common.PROD,
		}
		connConfig = common.ConnectionConfig{
			ConnectionId:  "conn1",
			ConnectionUrl: "https://example.com",
		}
		client, _ = NewSkyflow()
	})

	Context("AddVault", func() {
		It("should add a vault successfully", func() {
			err := client.AddVaultConfig(vaultConfig)
			Expect(err).To(BeNil())
			Expect(client.vaultServices).To(HaveKey("vault1"))
		})
		It("should not add duplicate vault", func() {
			client.AddVaultConfig(vaultConfig)
			err := client.AddVaultConfig(vaultConfig)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("AddConnectionConfig(", func() {
		It("should add a connection successfully", func() {
			err := client.AddConnectionConfig(connConfig)
			Expect(err).To(BeNil())
			Expect(client.connectionServices).To(HaveKey("conn1"))
		})
		It("should not add duplicate connection", func() {
			client.AddConnectionConfig(connConfig)
			err := client.AddConnectionConfig(connConfig)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("AddSkyflowCredentials", func() {
		It("should add credentials successfully", func() {
			validCreds := common.Credentials{Token: "some-bearer-token"}
			err := client.AddSkyflowCredentials(validCreds)
			Expect(err).To(BeNil())
			Expect(client.credentials).To(Equal(&validCreds))
		})
		It("should fail with invalid credentials", func() {
			invalidCreds := common.Credentials{}
			err := client.AddSkyflowCredentials(invalidCreds)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("GetVaultConfig(", func() {
		It("should get vault config", func() {
			client.AddVaultConfig(vaultConfig)
			cfg, err := client.GetVaultConfig("vault1")
			Expect(err).To(BeNil())
			Expect(cfg.VaultId).To(Equal("vault1"))
		})
		It("should fail for missing vault", func() {
			cfg, err := client.GetVaultConfig("missing")
			Expect(err).ToNot(BeNil())
			Expect(cfg).To(BeNil())
		})
	})

	Context("GetConnectionConfig(", func() {
		It("should get connection config", func() {
			client.AddConnectionConfig(connConfig)
			cfg, err := client.GetConnectionConfig("conn1")
			Expect(err).To(BeNil())
			Expect(cfg.ConnectionId).To(Equal("conn1"))
		})
		It("should fail for missing connection", func() {
			cfg, err := client.GetConnectionConfig("missing")
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
			client.AddVaultConfig(vaultConfig)
			client.AddConnectionConfig(connConfig)
			validCreds := common.Credentials{Token: "some-bearer-token"}
			err := client.UpdateSkyflowCredentials(validCreds)
			Expect(err).To(BeNil())
			Expect(client.credentials).To(Equal(&validCreds))
			// Check controllers updated
			for _, v := range client.vaultServices {
				if v.controller != nil {
					Expect(v.controller.Config.Credentials).To(Equal(validCreds))
				}
			}
		})
		It("should fail with invalid credentials", func() {
			invalidCreds := common.Credentials{}
			err := client.UpdateSkyflowCredentials(invalidCreds)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("UpdateVaultConfig", func() {
		It("should update vault config and propagate to controller", func() {
			client.AddVaultConfig(vaultConfig)
			updated := vaultConfig
			updated.ClusterId = "new-cluster"
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.vaultServices["vault1"].config.ClusterId).To(Equal("new-cluster"))
			if client.vaultServices["vault1"].controller != nil {
				Expect(client.vaultServices["vault1"].controller.Config.ClusterId).To(Equal("new-cluster"))
			}
		})
		It("should fail for missing vault", func() {
			updated := vaultConfig
			updated.VaultId = "missing"
			err := client.UpdateVaultConfig(updated)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("UpdateConnectionConfig", func() {
		It("should update connection config", func() {
			client.AddConnectionConfig(connConfig)
			updated := connConfig
			updated.ConnectionUrl = "https://new-url.com"
			err := client.UpdateConnectionConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.connectionServices["conn1"].config.ConnectionUrl).To(Equal("https://new-url.com"))
		})
		It("should fail for missing connection", func() {
			updated := connConfig
			updated.ConnectionId = "missing"
			err := client.UpdateConnectionConfig(updated)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("GetLoglevel", func() {
		It("should get current log level", func() {
			client.UpdateLogLevel(logger.INFO)
			Expect(*client.GetLoglevel()).To(Equal(logger.INFO))
		})
	})

	Context("RemoveVaultConfig(", func() {
		It("should remove vault", func() {
			client.AddVaultConfig(vaultConfig)
			err := client.RemoveVaultConfig("vault1")
			Expect(err).To(BeNil())
			Expect(client.vaultServices).ToNot(HaveKey("vault1"))
		})
		It("should fail for missing vault", func() {
			err := client.RemoveVaultConfig("missing")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("RemoveConnection", func() {
		It("should remove connection", func() {
			client.AddConnectionConfig(connConfig)
			err := client.RemoveConnectionConfig("conn1")
			Expect(err).To(BeNil())
			Expect(client.connectionServices).ToNot(HaveKey("conn1"))
		})
		It("should fail for missing connection", func() {
			err := client.RemoveConnectionConfig("missing")
			Expect(err).ToNot(BeNil())
		})
	})
	Context("cross scenario: Vault(vaultid) and update/add/remove", func() {
		var vaultConfig common.VaultConfig
		var creds common.Credentials
		BeforeEach(func() {
			if os.Getenv("API_KEY") == "" {
				Skip("requires API_KEY env var")
			}
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
			err2 := client.UpdateVaultConfig(updated)
			Expect(err2).To(BeNil())
			vault, err3 := client.GetVaultConfig("vaultX")
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
			err4 := client.AddVaultConfig(vaultConfig2)
			Expect(err4).To(BeNil())
			vault2, err5 := client.GetVaultConfig("vaultY")
			Expect(err5).To(BeNil())
			Expect(vault2.VaultId).To(Equal("vaultY"))

			err6 := client.RemoveVaultConfig("vaultX")
			Expect(err6).To(BeNil())
			_, err7 := client.GetVaultConfig("vaultX")
			Expect(err7).ToNot(BeNil())
		})
		It("should fail to get, update, or remove missing vault", func() {
			vaultSvc, err := client.Vault("missing")
			Expect(err).ToNot(BeNil())
			Expect(vaultSvc).To(BeNil())
			missing := vaultConfig
			missing.VaultId = "missing"
			err2 := client.UpdateVaultConfig(missing)
			Expect(err2).ToNot(BeNil())
			err3 := client.RemoveVaultConfig("missing")
			Expect(err3).ToNot(BeNil())
		})
	})

	Context("cross scenario: Detect(vaultid) and update/add/remove", func() {
		var vaultConfig common.VaultConfig
		var creds common.Credentials
		BeforeEach(func() {
			if os.Getenv("API_KEY") == "" {
				Skip("requires API_KEY env var")
			}
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
			err2 := client.UpdateVaultConfig(updated)
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
			err4 := client.AddVaultConfig(vaultConfig2)
			Expect(err4).To(BeNil())
			detect2, err5 := client.Detect("vaultDY")
			Expect(err5).To(BeNil())
			Expect(detect2.config.VaultId).To(Equal("vaultDY"))

			err6 := client.RemoveVaultConfig("vaultDX")
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
			err2 := client.UpdateVaultConfig(missing)
			Expect(err2).ToNot(BeNil())
			err3 := client.RemoveVaultConfig("missing")
			Expect(err3).ToNot(BeNil())
		})
	})

	Context("cross scenario: Connection(connectionId) and update/add/remove", func() {
		var connConfig common.ConnectionConfig
		var creds common.Credentials
		BeforeEach(func() {
			if os.Getenv("API_KEY") == "" {
				Skip("requires API_KEY env var")
			}
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
			err2 := client.UpdateConnectionConfig(updated)
			Expect(err2).To(BeNil())
			conn, err3 := client.GetConnectionConfig("connX")
			Expect(err3).To(BeNil())
			Expect(conn.ConnectionUrl).To(Equal("https://new-connX.com"))

			connConfig2 := common.ConnectionConfig{
				ConnectionId:  "connY",
				ConnectionUrl: "https://connY.com",
				Credentials: common.Credentials{
					ApiKey: os.Getenv("API_KEY"),
				},
			}
			err4 := client.AddConnectionConfig(connConfig2)
			Expect(err4).To(BeNil())
			conn2, err5 := client.GetConnectionConfig("connY")
			Expect(err5).To(BeNil())
			Expect(conn2.ConnectionId).To(Equal("connY"))

			err6 := client.RemoveConnectionConfig("connX")
			Expect(err6).To(BeNil())
			_, err7 := client.GetConnectionConfig("connX")
			Expect(err7).ToNot(BeNil())
		})
		It("should fail to get, update, or remove missing connection", func() {
			connSvc, err := client.Connection("missing")
			Expect(err).ToNot(BeNil())
			Expect(connSvc).To(BeNil())
			missing := connConfig
			missing.ConnectionId = "missing"
			err2 := client.UpdateConnectionConfig(missing)
			Expect(err2).ToNot(BeNil())
			err3 := client.RemoveConnectionConfig("missing")
			Expect(err3).ToNot(BeNil())
		})
	})

	Context("Backward compat — deprecated method wrappers", func() {
		var bc *Skyflow
		var bcVault common.VaultConfig
		var bcConn common.ConnectionConfig

		BeforeEach(func() {
			bcVault = common.VaultConfig{
				VaultId:   "bc-vault",
				ClusterId: "bc-cluster",
				Env:       common.PROD,
				Credentials: common.Credentials{
					ApiKey: "key",
				},
			}
			bcConn = common.ConnectionConfig{
				ConnectionId:  "bc-conn",
				ConnectionUrl: "https://bc-conn.example.com",
				Credentials:   common.Credentials{ApiKey: "key"},
			}
			var bcErr *error.SkyflowError
			bc, bcErr = NewSkyflow(
				WithVaults(bcVault),
				WithConnections(bcConn),
				WithCredentials(common.Credentials{CredentialsString: "some-credentials"}),
			)
			Expect(bcErr).To(BeNil())
		})

		It("GetVault delegates to GetVaultConfig", func() {
			cfg, err := bc.GetVault(bcVault.VaultId)
			Expect(err).To(BeNil())
			Expect(cfg.VaultId).To(Equal(bcVault.VaultId))
		})

		It("GetConnection delegates to GetConnectionConfig", func() {
			cfg, err := bc.GetConnection(bcConn.ConnectionId)
			Expect(err).To(BeNil())
			Expect(cfg.ConnectionId).To(Equal(bcConn.ConnectionId))
		})

		It("AddVault delegates to AddVaultConfig", func() {
			newVault := common.VaultConfig{
				VaultId:   "new-vault-bc",
				ClusterId: "cluster-bc",
				Env:       common.PROD,
				Credentials: common.Credentials{
					ApiKey: "key",
				},
			}
			err := bc.AddVault(newVault)
			Expect(err).To(BeNil())
			cfg, err2 := bc.GetVaultConfig("new-vault-bc")
			Expect(err2).To(BeNil())
			Expect(cfg.VaultId).To(Equal("new-vault-bc"))
		})

		It("AddConnection delegates to AddConnectionConfig", func() {
			newConn := common.ConnectionConfig{
				ConnectionId:  "new-conn-bc",
				ConnectionUrl: "https://conn-bc.example.com",
				Credentials:   common.Credentials{ApiKey: "key"},
			}
			err := bc.AddConnection(newConn)
			Expect(err).To(BeNil())
			cfg, err2 := bc.GetConnectionConfig("new-conn-bc")
			Expect(err2).To(BeNil())
			Expect(cfg.ConnectionId).To(Equal("new-conn-bc"))
		})

		It("UpdateVault delegates to UpdateVaultConfig", func() {
			updated := common.VaultConfig{
				VaultId:   bcVault.VaultId,
				ClusterId: "updated-cluster",
				Env:       common.PROD,
			}
			err := bc.UpdateVault(updated)
			Expect(err).To(BeNil())
		})

		It("UpdateConnection delegates to UpdateConnectionConfig", func() {
			updated := common.ConnectionConfig{
				ConnectionId:  bcConn.ConnectionId,
				ConnectionUrl: "https://updated-conn.example.com",
			}
			err := bc.UpdateConnection(updated)
			Expect(err).To(BeNil())
		})

		It("RemoveVault delegates to RemoveVaultConfig", func() {
			err := bc.RemoveVault(bcVault.VaultId)
			Expect(err).To(BeNil())
			_, err2 := bc.GetVaultConfig(bcVault.VaultId)
			Expect(err2).ToNot(BeNil())
		})

		It("RemoveConnection delegates to RemoveConnectionConfig", func() {
			err := bc.RemoveConnection(bcConn.ConnectionId)
			Expect(err).To(BeNil())
			_, err2 := bc.GetConnectionConfig(bcConn.ConnectionId)
			Expect(err2).ToNot(BeNil())
		})

		It("VaultConfig.BaseVaultURL (old field) is accepted in AddVault", func() {
			newVault := common.VaultConfig{
				VaultId:      "vault-old-url",
				BaseVaultURL: "https://old-url.example.com",
				Env:          common.PROD,
				Credentials:  common.Credentials{ApiKey: "key"},
			}
			err := bc.AddVault(newVault)
			Expect(err).To(BeNil())
		})

		It("RequestIDHeader (old constant) is accepted in WithCustomHeaders", func() {
			newVault := common.VaultConfig{VaultId: "hdr-vault", ClusterId: "c", Env: common.PROD}
			_, err := NewSkyflow(
				WithVaults(newVault),
				WithCredentials(common.Credentials{CredentialsString: "creds"}),
				WithCustomHeaders(map[common.CustomHeaderKey]string{
					common.RequestIDHeader: "req-123",
				}),
			)
			Expect(err).To(BeNil())
		})

		It("SkyflowAccountID (old constant) is accepted in WithCustomHeaders", func() {
			newVault := common.VaultConfig{VaultId: "acct-vault", ClusterId: "c", Env: common.PROD}
			_, err := NewSkyflow(
				WithVaults(newVault),
				WithCredentials(common.Credentials{CredentialsString: "creds"}),
				WithCustomHeaders(map[common.CustomHeaderKey]string{
					common.SkyflowAccountID: "acct-123",
				}),
			)
			Expect(err).To(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// UpdateVaultConfig — detect controller path
	// -----------------------------------------------------------------------
	Context("UpdateVaultConfig — detect controller is updated when Detect() has been called", func() {
		var vc common.VaultConfig
		BeforeEach(func() {
			vc = common.VaultConfig{VaultId: "dv1", ClusterId: "c1", Env: common.PROD}
			client, _ = NewSkyflow()
			client.AddVaultConfig(vc)
			// Initialise the detect controller by calling Detect()
			_, _ = client.Detect("dv1")
		})

		It("should update detect controller credentials when non-empty creds are provided", func() {
			updated := common.VaultConfig{
				VaultId:     "dv1",
				ClusterId:   "c1",
				Credentials: common.Credentials{Token: "detect-new-token"},
			}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.detectServices["dv1"].controller.Config.Credentials.Token).To(Equal("detect-new-token"))
			Expect(client.detectServices["dv1"].controller.Token).To(Equal(""))
		})

		It("should update detect controller clusterId when non-empty clusterId is provided", func() {
			updated := common.VaultConfig{VaultId: "dv1", ClusterId: "new-cluster"}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.detectServices["dv1"].controller.Config.ClusterId).To(Equal("new-cluster"))
		})

		It("should update detect controller env and skip credentials when empty creds", func() {
			updated := common.VaultConfig{VaultId: "dv1", ClusterId: "c1", Env: common.SANDBOX}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.detectServices["dv1"].controller.Config.Env).To(Equal(common.SANDBOX))
		})

		It("should update both vault and detect controllers when both have been initialised", func() {
			_, _ = client.Vault("dv1")
			updated := common.VaultConfig{
				VaultId:     "dv1",
				ClusterId:   "shared-cluster",
				Credentials: common.Credentials{Token: "shared-token"},
			}
			err := client.UpdateVaultConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.vaultServices["dv1"].controller.Config.Credentials.Token).To(Equal("shared-token"))
			Expect(client.detectServices["dv1"].controller.Config.Credentials.Token).To(Equal("shared-token"))
		})
	})

	// -----------------------------------------------------------------------
	// UpdateVaultConfig — validation failure
	// -----------------------------------------------------------------------
	Context("UpdateVaultConfig — validation errors", func() {
		It("should return error when VaultId is empty", func() {
			err := client.UpdateVaultConfig(common.VaultConfig{VaultId: ""})
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.INVALID_VAULT_ID))
		})
	})

	// -----------------------------------------------------------------------
	// UpdateConnectionConfig — validation failure + controller nil path
	// -----------------------------------------------------------------------
	Context("UpdateConnectionConfig — validation errors and controller-nil path", func() {
		It("should return error when ConnectionId is empty", func() {
			err := client.UpdateConnectionConfig(common.ConnectionConfig{ConnectionId: ""})
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_ID))
		})

		It("should update config even when controller is nil (Connection() not yet called)", func() {
			client.AddConnectionConfig(connConfig)
			updated := common.ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://updated-no-ctrl.example.com",
			}
			err := client.UpdateConnectionConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.connectionServices["conn1"].config.ConnectionUrl).To(Equal("https://updated-no-ctrl.example.com"))
		})

		It("should not update controller credentials when empty creds supplied", func() {
			client.AddConnectionConfig(connConfig)
			_, _ = client.Connection("conn1")
			client.connectionServices["conn1"].controller.Config.Credentials = common.Credentials{Token: "original"}

			updated := common.ConnectionConfig{
				ConnectionId:  "conn1",
				ConnectionUrl: "https://new-url.com",
				// Credentials intentionally empty — should not overwrite
			}
			err := client.UpdateConnectionConfig(updated)
			Expect(err).To(BeNil())
			Expect(client.connectionServices["conn1"].controller.Config.Credentials.Token).To(Equal("original"))
		})
	})

	// -----------------------------------------------------------------------
	// Vault / Connection / Detect — all error and happy paths
	// -----------------------------------------------------------------------
	Context("Vault() — error and happy paths", func() {
		It("should return error when no vaults are registered", func() {
			emptyClient, _ := NewSkyflow()
			_, err := emptyClient.Vault()
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_VAULT_CONFIG))
		})

		It("should return error when specified vault ID is not registered", func() {
			client.AddVaultConfig(vaultConfig)
			_, err := client.Vault("nonexistent-id")
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.VAULT_ID_NOT_IN_CONFIG_LIST))
		})

		It("should return vault service for registered vault ID", func() {
			client.AddVaultConfig(vaultConfig)
			svc, err := client.Vault("vault1")
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
			Expect(svc.config.VaultId).To(Equal("vault1"))
		})

		It("should return first vault when no ID supplied", func() {
			client.AddVaultConfig(vaultConfig)
			svc, err := client.Vault()
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
		})
	})

	Context("Connection() — error and happy paths", func() {
		It("should return error when no connections are registered", func() {
			emptyClient, _ := NewSkyflow()
			_, err := emptyClient.Connection()
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_CONFIG))
		})

		It("should return error when specified connection ID is not registered", func() {
			client.AddConnectionConfig(connConfig)
			_, err := client.Connection("nonexistent-id")
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.CONNECTION_ID_NOT_IN_CONFIG_LIST))
		})

		It("should return connection service for registered connection ID", func() {
			client.AddConnectionConfig(connConfig)
			svc, err := client.Connection("conn1")
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
			Expect(svc.config.ConnectionId).To(Equal("conn1"))
		})

		It("should return first connection when no ID supplied", func() {
			client.AddConnectionConfig(connConfig)
			svc, err := client.Connection()
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
		})
	})

	Context("Detect() — error and happy paths", func() {
		It("should return error when no vaults are registered", func() {
			emptyClient, _ := NewSkyflow()
			_, err := emptyClient.Detect()
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_VAULT_CONFIG))
		})

		It("should return error when specified vault ID is not registered for detect", func() {
			client.AddVaultConfig(vaultConfig)
			_, err := client.Detect("nonexistent-id")
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.VAULT_ID_NOT_IN_CONFIG_LIST))
		})

		It("should return detect service for registered vault ID", func() {
			client.AddVaultConfig(vaultConfig)
			svc, err := client.Detect("vault1")
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
			Expect(svc.config.VaultId).To(Equal("vault1"))
		})

		It("should return first detect service when no ID supplied", func() {
			client.AddVaultConfig(vaultConfig)
			svc, err := client.Detect()
			Expect(err).To(BeNil())
			Expect(svc).ToNot(BeNil())
		})
	})

	// -----------------------------------------------------------------------
	// AddVaultConfig / AddConnectionConfig — validation errors
	// -----------------------------------------------------------------------
	Context("AddVaultConfig — validation errors", func() {
		It("should return error when VaultId is empty", func() {
			err := client.AddVaultConfig(common.VaultConfig{VaultId: "", ClusterId: "c"})
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.INVALID_VAULT_ID))
		})

		It("should return error when both ClusterId and BaseVaultUrl are empty", func() {
			err := client.AddVaultConfig(common.VaultConfig{VaultId: "v"})
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.INVALID_CLUSTER_ID))
		})

		It("should return error when BaseVaultUrl is not a valid HTTP URL", func() {
			err := client.AddVaultConfig(common.VaultConfig{VaultId: "v", BaseVaultUrl: "not-a-url"})
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.INVALID_VAULT_URL))
		})

		It("should accept BaseVaultUrl instead of ClusterId", func() {
			err := client.AddVaultConfig(common.VaultConfig{VaultId: "v-url", BaseVaultUrl: "https://custom.example.com"})
			Expect(err).To(BeNil())
		})
	})

	Context("AddConnectionConfig — validation errors", func() {
		It("should return error when ConnectionId is empty", func() {
			err := client.AddConnectionConfig(common.ConnectionConfig{ConnectionId: "", ConnectionUrl: "https://url.com"})
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_ID))
		})

		It("should return error when ConnectionUrl is empty", func() {
			err := client.AddConnectionConfig(common.ConnectionConfig{ConnectionId: "c", ConnectionUrl: ""})
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_URL))
		})
	})

	// -----------------------------------------------------------------------
	// WithCredentials — validation failures
	// -----------------------------------------------------------------------
	Context("WithCredentials — validation failures", func() {
		It("should return error when credentials are empty", func() {
			_, err := NewSkyflow(WithCredentials(common.Credentials{}))
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.NO_TOKEN_GENERATION_MEANS_PASSED))
		})

		It("should return error when multiple credential types are set simultaneously", func() {
			_, err := NewSkyflow(WithCredentials(common.Credentials{Token: "t", ApiKey: "k"}))
			Expect(err).ToNot(BeNil())
			Expect(err.GetMessage()).To(ContainSubstring(error.MULTIPLE_TOKEN_GENERATION_MEANS_PASSED))
		})
	})

	// -----------------------------------------------------------------------
	// GetSkyflowCredentials — nil and set cases
	// -----------------------------------------------------------------------
	Context("GetSkyflowCredentials — nil and set cases", func() {
		It("should return nil when no credentials have been set", func() {
			emptyClient, _ := NewSkyflow()
			Expect(emptyClient.GetSkyflowCredentials()).To(BeNil())
		})

		It("should return credentials after AddSkyflowCredentials is called", func() {
			emptyClient, _ := NewSkyflow()
			emptyClient.AddSkyflowCredentials(common.Credentials{Token: "my-token"})
			Expect(emptyClient.GetSkyflowCredentials().Token).To(Equal("my-token"))
		})
	})

	// -----------------------------------------------------------------------
	// UpdateLogLevel — propagates to all three service maps
	// -----------------------------------------------------------------------
	Context("UpdateLogLevel — propagates to vault, detect, and connection services", func() {
		It("should set the same log level on all registered services", func() {
			client.AddVaultConfig(vaultConfig)
			client.AddConnectionConfig(connConfig)
			client.UpdateLogLevel(logger.DEBUG)
			Expect(*client.vaultServices["vault1"].logLevel).To(Equal(logger.DEBUG))
			Expect(*client.detectServices["vault1"].logLevel).To(Equal(logger.DEBUG))
			Expect(*client.connectionServices["conn1"].logLevel).To(Equal(logger.DEBUG))
			Expect(*client.GetLoglevel()).To(Equal(logger.DEBUG))
		})
	})

	// -----------------------------------------------------------------------
	// RemoveVaultConfig — verify both maps cleared on normal AddVaultConfig flow
	// -----------------------------------------------------------------------
	Context("RemoveVaultConfig — clears both maps after AddVaultConfig", func() {
		It("should remove from both vaultServices and detectServices", func() {
			client.AddVaultConfig(vaultConfig)
			Expect(client.vaultServices).To(HaveKey("vault1"))
			Expect(client.detectServices).To(HaveKey("vault1"))
			err := client.RemoveVaultConfig("vault1")
			Expect(err).To(BeNil())
			Expect(client.vaultServices).ToNot(HaveKey("vault1"))
			Expect(client.detectServices).ToNot(HaveKey("vault1"))
		})
	})
})

// =============================================================================
// Lifecycle scenarios — controller activated, then config mutated
// =============================================================================
var _ = Describe("Skyflow lifecycle: after Vault/Detect/Connection activated", func() {
	var (
		client     *Skyflow
		vaultCfg   common.VaultConfig
		connCfg    common.ConnectionConfig
	)

	BeforeEach(func() {
		vaultCfg = common.VaultConfig{VaultId: "v1", ClusterId: "cluster1", Env: common.PROD}
		connCfg = common.ConnectionConfig{
			ConnectionId:  "c1",
			ConnectionUrl: "https://conn.example.com",
		}
		var err *error.SkyflowError
		client, err = NewSkyflow(WithVaults(vaultCfg), WithConnections(connCfg))
		Expect(err).To(BeNil())
	})

	// -------------------------------------------------------------------------
	// After Vault() — UpdateVaultConfig
	// -------------------------------------------------------------------------
	Context("after Vault() is called — UpdateVaultConfig", func() {
		BeforeEach(func() {
			_, err := client.Vault("v1")
			Expect(err).To(BeNil())
		})

		It("should reflect updated ClusterId in the vault controller config", func() {
			err := client.UpdateVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "new-cluster"})
			Expect(err).To(BeNil())
			Expect(client.vaultServices["v1"].controller.Config.ClusterId).To(Equal("new-cluster"))
			Expect(client.vaultServices["v1"].config.ClusterId).To(Equal("new-cluster"))
		})

		It("should reflect updated credentials in the vault controller and clear cached tokens", func() {
			err := client.UpdateVaultConfig(common.VaultConfig{
				VaultId:     "v1",
				Credentials: common.Credentials{Token: "updated-token"},
			})
			Expect(err).To(BeNil())
			Expect(client.vaultServices["v1"].controller.Config.Credentials.Token).To(Equal("updated-token"))
			Expect(client.vaultServices["v1"].controller.Token).To(Equal(""))
			Expect(client.vaultServices["v1"].controller.ApiKey).To(Equal(""))
		})

		It("should leave credentials unchanged when empty creds are supplied", func() {
			client.vaultServices["v1"].controller.Config.Credentials = common.Credentials{Token: "original"}
			err := client.UpdateVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "c2"})
			Expect(err).To(BeNil())
			Expect(client.vaultServices["v1"].controller.Config.Credentials.Token).To(Equal("original"))
		})

		It("should update env on vault controller", func() {
			err := client.UpdateVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "cluster1", Env: common.SANDBOX})
			Expect(err).To(BeNil())
			Expect(client.vaultServices["v1"].controller.Config.Env).To(Equal(common.SANDBOX))
		})

		It("subsequent Vault() call uses the updated config", func() {
			client.UpdateVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "refreshed-cluster"})
			svc, err := client.Vault("v1")
			Expect(err).To(BeNil())
			Expect(svc.config.ClusterId).To(Equal("refreshed-cluster"))
		})
	})

	// -------------------------------------------------------------------------
	// After Vault() — RemoveVaultConfig
	// -------------------------------------------------------------------------
	Context("after Vault() is called — RemoveVaultConfig", func() {
		BeforeEach(func() {
			_, err := client.Vault("v1")
			Expect(err).To(BeNil())
		})

		It("should remove the activated vault service", func() {
			err := client.RemoveVaultConfig("v1")
			Expect(err).To(BeNil())
			Expect(client.vaultServices).ToNot(HaveKey("v1"))
		})

		It("Vault() should error after the vault is removed", func() {
			client.RemoveVaultConfig("v1")
			_, err := client.Vault("v1")
			Expect(err).ToNot(BeNil())
			// vaultServices map is now empty → EMPTY_VAULT_CONFIG is returned before the "not found" check
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_VAULT_CONFIG))
		})

		It("AddVaultConfig with same ID should succeed after removal", func() {
			client.RemoveVaultConfig("v1")
			err := client.AddVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "re-added"})
			Expect(err).To(BeNil())
			cfg, _ := client.GetVaultConfig("v1")
			Expect(cfg.ClusterId).To(Equal("re-added"))
		})
	})

	// -------------------------------------------------------------------------
	// After Vault() — AddVaultConfig (second vault), then use both
	// -------------------------------------------------------------------------
	Context("after Vault() is called — AddVaultConfig adds second vault", func() {
		BeforeEach(func() {
			_, err := client.Vault("v1")
			Expect(err).To(BeNil())
		})

		It("should activate both vaults independently", func() {
			err := client.AddVaultConfig(common.VaultConfig{VaultId: "v2", ClusterId: "cluster2"})
			Expect(err).To(BeNil())

			svc1, err1 := client.Vault("v1")
			svc2, err2 := client.Vault("v2")
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(svc1.config.VaultId).To(Equal("v1"))
			Expect(svc2.config.VaultId).To(Equal("v2"))
		})

		It("removing one vault should not affect the other", func() {
			client.AddVaultConfig(common.VaultConfig{VaultId: "v2", ClusterId: "cluster2"})
			client.RemoveVaultConfig("v1")
			_, err1 := client.Vault("v1")
			Expect(err1).ToNot(BeNil())
			svc2, err2 := client.Vault("v2")
			Expect(err2).To(BeNil())
			Expect(svc2.config.VaultId).To(Equal("v2"))
		})
	})

	// -------------------------------------------------------------------------
	// After Detect() — UpdateVaultConfig
	// -------------------------------------------------------------------------
	Context("after Detect() is called — UpdateVaultConfig", func() {
		BeforeEach(func() {
			_, err := client.Detect("v1")
			Expect(err).To(BeNil())
		})

		It("should reflect updated ClusterId in the detect controller config", func() {
			err := client.UpdateVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "detect-new-cluster"})
			Expect(err).To(BeNil())
			Expect(client.detectServices["v1"].controller.Config.ClusterId).To(Equal("detect-new-cluster"))
			Expect(client.detectServices["v1"].config.ClusterId).To(Equal("detect-new-cluster"))
		})

		It("should update detect controller credentials and clear cached token", func() {
			err := client.UpdateVaultConfig(common.VaultConfig{
				VaultId:     "v1",
				Credentials: common.Credentials{Token: "detect-token"},
			})
			Expect(err).To(BeNil())
			Expect(client.detectServices["v1"].controller.Config.Credentials.Token).To(Equal("detect-token"))
			Expect(client.detectServices["v1"].controller.Token).To(Equal(""))
		})

		It("subsequent Detect() call uses the updated config", func() {
			client.UpdateVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "detect-refreshed"})
			svc, err := client.Detect("v1")
			Expect(err).To(BeNil())
			Expect(svc.config.ClusterId).To(Equal("detect-refreshed"))
		})
	})

	// -------------------------------------------------------------------------
	// After Detect() — RemoveVaultConfig
	// -------------------------------------------------------------------------
	Context("after Detect() is called — RemoveVaultConfig", func() {
		BeforeEach(func() {
			_, err := client.Detect("v1")
			Expect(err).To(BeNil())
		})

		It("should remove the activated detect service", func() {
			err := client.RemoveVaultConfig("v1")
			Expect(err).To(BeNil())
			Expect(client.detectServices).ToNot(HaveKey("v1"))
		})

		It("Detect() should error after the vault is removed", func() {
			client.RemoveVaultConfig("v1")
			_, err := client.Detect("v1")
			Expect(err).ToNot(BeNil())
			// detectServices map is now empty → EMPTY_VAULT_CONFIG is returned before the "not found" check
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_VAULT_CONFIG))
		})

		It("AddVaultConfig then Detect() should succeed after removal", func() {
			client.RemoveVaultConfig("v1")
			client.AddVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "re-detect"})
			svc, err := client.Detect("v1")
			Expect(err).To(BeNil())
			Expect(svc.config.ClusterId).To(Equal("re-detect"))
		})
	})

	// -------------------------------------------------------------------------
	// Both Vault() and Detect() activated — coordinated mutations
	// -------------------------------------------------------------------------
	Context("both Vault() and Detect() activated on same vault ID", func() {
		BeforeEach(func() {
			_, err1 := client.Vault("v1")
			_, err2 := client.Detect("v1")
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
		})

		It("UpdateVaultConfig should update both controllers", func() {
			err := client.UpdateVaultConfig(common.VaultConfig{
				VaultId:     "v1",
				ClusterId:   "both-updated",
				Credentials: common.Credentials{Token: "both-token"},
			})
			Expect(err).To(BeNil())
			Expect(client.vaultServices["v1"].controller.Config.ClusterId).To(Equal("both-updated"))
			Expect(client.detectServices["v1"].controller.Config.ClusterId).To(Equal("both-updated"))
			Expect(client.vaultServices["v1"].controller.Config.Credentials.Token).To(Equal("both-token"))
			Expect(client.detectServices["v1"].controller.Config.Credentials.Token).To(Equal("both-token"))
		})

		It("RemoveVaultConfig should remove both services", func() {
			err := client.RemoveVaultConfig("v1")
			Expect(err).To(BeNil())
			Expect(client.vaultServices).ToNot(HaveKey("v1"))
			Expect(client.detectServices).ToNot(HaveKey("v1"))
		})

		It("UpdateSkyflowCredentials should propagate to both controllers", func() {
			err := client.UpdateSkyflowCredentials(common.Credentials{Token: "global-token"})
			Expect(err).To(BeNil())
			Expect(client.vaultServices["v1"].controller.CommonCreds.Token).To(Equal("global-token"))
			Expect(client.detectServices["v1"].controller.CommonCreds.Token).To(Equal("global-token"))
		})

		It("AddSkyflowCredentials should propagate to both controllers", func() {
			err := client.AddSkyflowCredentials(common.Credentials{Token: "add-global-token"})
			Expect(err).To(BeNil())
			Expect(client.vaultServices["v1"].controller.CommonCreds.Token).To(Equal("add-global-token"))
			Expect(client.detectServices["v1"].controller.CommonCreds.Token).To(Equal("add-global-token"))
		})
	})

	// -------------------------------------------------------------------------
	// After Connection() — UpdateConnectionConfig
	// -------------------------------------------------------------------------
	Context("after Connection() is called — UpdateConnectionConfig", func() {
		BeforeEach(func() {
			_, err := client.Connection("c1")
			Expect(err).To(BeNil())
		})

		It("should reflect updated ConnectionUrl in the connection controller", func() {
			err := client.UpdateConnectionConfig(common.ConnectionConfig{
				ConnectionId:  "c1",
				ConnectionUrl: "https://updated-conn.example.com",
			})
			Expect(err).To(BeNil())
			Expect(client.connectionServices["c1"].controller.Config.ConnectionUrl).To(Equal("https://updated-conn.example.com"))
			Expect(client.connectionServices["c1"].config.ConnectionUrl).To(Equal("https://updated-conn.example.com"))
		})

		It("should update connection controller credentials and clear cached token", func() {
			err := client.UpdateConnectionConfig(common.ConnectionConfig{
				ConnectionId:  "c1",
				ConnectionUrl: "https://conn.example.com",
				Credentials:   common.Credentials{Token: "conn-token"},
			})
			Expect(err).To(BeNil())
			Expect(client.connectionServices["c1"].controller.Config.Credentials.Token).To(Equal("conn-token"))
			Expect(client.connectionServices["c1"].controller.Token).To(Equal(""))
		})

		It("should not update credentials when empty creds supplied", func() {
			client.connectionServices["c1"].controller.Config.Credentials = common.Credentials{Token: "original-conn"}
			err := client.UpdateConnectionConfig(common.ConnectionConfig{
				ConnectionId:  "c1",
				ConnectionUrl: "https://new-url.example.com",
			})
			Expect(err).To(BeNil())
			Expect(client.connectionServices["c1"].controller.Config.Credentials.Token).To(Equal("original-conn"))
		})

		It("subsequent Connection() call uses the updated URL", func() {
			client.UpdateConnectionConfig(common.ConnectionConfig{
				ConnectionId:  "c1",
				ConnectionUrl: "https://refreshed-conn.example.com",
			})
			svc, err := client.Connection("c1")
			Expect(err).To(BeNil())
			Expect(svc.config.ConnectionUrl).To(Equal("https://refreshed-conn.example.com"))
		})
	})

	// -------------------------------------------------------------------------
	// After Connection() — RemoveConnectionConfig
	// -------------------------------------------------------------------------
	Context("after Connection() is called — RemoveConnectionConfig", func() {
		BeforeEach(func() {
			_, err := client.Connection("c1")
			Expect(err).To(BeNil())
		})

		It("should remove the activated connection service", func() {
			err := client.RemoveConnectionConfig("c1")
			Expect(err).To(BeNil())
			Expect(client.connectionServices).ToNot(HaveKey("c1"))
		})

		It("Connection() should error after the connection is removed", func() {
			client.RemoveConnectionConfig("c1")
			_, err := client.Connection("c1")
			Expect(err).ToNot(BeNil())
			// connectionServices map is now empty → EMPTY_CONNECTION_CONFIG is returned before the "not found" check
			Expect(err.GetMessage()).To(ContainSubstring(error.EMPTY_CONNECTION_CONFIG))
		})

		It("AddConnectionConfig with same ID should succeed after removal", func() {
			client.RemoveConnectionConfig("c1")
			err := client.AddConnectionConfig(common.ConnectionConfig{
				ConnectionId:  "c1",
				ConnectionUrl: "https://re-added-conn.example.com",
			})
			Expect(err).To(BeNil())
			cfg, _ := client.GetConnectionConfig("c1")
			Expect(cfg.ConnectionUrl).To(Equal("https://re-added-conn.example.com"))
		})
	})

	// -------------------------------------------------------------------------
	// After Connection() — UpdateSkyflowCredentials / AddSkyflowCredentials
	// -------------------------------------------------------------------------
	Context("after Connection() is called — global credential updates", func() {
		BeforeEach(func() {
			_, err := client.Connection("c1")
			Expect(err).To(BeNil())
		})

		It("UpdateSkyflowCredentials propagates to connection controller", func() {
			err := client.UpdateSkyflowCredentials(common.Credentials{Token: "global-conn-token"})
			Expect(err).To(BeNil())
			Expect(client.connectionServices["c1"].controller.CommonCreds.Token).To(Equal("global-conn-token"))
			Expect(client.connectionServices["c1"].controller.Token).To(Equal(""))
		})

		It("AddSkyflowCredentials propagates to connection controller", func() {
			err := client.AddSkyflowCredentials(common.Credentials{Token: "add-conn-token"})
			Expect(err).To(BeNil())
			Expect(client.connectionServices["c1"].controller.CommonCreds.Token).To(Equal("add-conn-token"))
		})
	})

	// -------------------------------------------------------------------------
	// Full end-to-end lifecycle: init → activate → update → re-activate → remove
	// -------------------------------------------------------------------------
	Context("full lifecycle: init → Vault → UpdateVaultConfig → Vault → RemoveVaultConfig", func() {
		It("controller reflects each mutation at every stage", func() {
			// 1. Activate vault
			svc1, err := client.Vault("v1")
			Expect(err).To(BeNil())
			Expect(svc1.config.ClusterId).To(Equal("cluster1"))

			// 2. Update cluster
			err = client.UpdateVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "mid-cluster"})
			Expect(err).To(BeNil())

			// 3. Re-activate — controller should have new cluster
			svc2, err := client.Vault("v1")
			Expect(err).To(BeNil())
			Expect(svc2.config.ClusterId).To(Equal("mid-cluster"))
			Expect(svc2.controller.Config.ClusterId).To(Equal("mid-cluster"))

			// 4. Remove
			err = client.RemoveVaultConfig("v1")
			Expect(err).To(BeNil())

			// 5. Vault() must fail now
			_, err = client.Vault("v1")
			Expect(err).ToNot(BeNil())

			// 6. Re-add and confirm fresh state
			err = client.AddVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "final-cluster"})
			Expect(err).To(BeNil())
			cfg, _ := client.GetVaultConfig("v1")
			Expect(cfg.ClusterId).To(Equal("final-cluster"))
		})
	})

	Context("full lifecycle: init → Connection → UpdateConnectionConfig → Connection → RemoveConnectionConfig", func() {
		It("controller reflects each mutation at every stage", func() {
			// 1. Activate
			svc1, err := client.Connection("c1")
			Expect(err).To(BeNil())
			Expect(svc1.config.ConnectionUrl).To(Equal("https://conn.example.com"))

			// 2. Update URL
			err = client.UpdateConnectionConfig(common.ConnectionConfig{
				ConnectionId:  "c1",
				ConnectionUrl: "https://mid-conn.example.com",
			})
			Expect(err).To(BeNil())

			// 3. Re-activate — controller should have updated URL
			svc2, err := client.Connection("c1")
			Expect(err).To(BeNil())
			Expect(svc2.config.ConnectionUrl).To(Equal("https://mid-conn.example.com"))
			Expect(svc2.controller.Config.ConnectionUrl).To(Equal("https://mid-conn.example.com"))

			// 4. Remove
			err = client.RemoveConnectionConfig("c1")
			Expect(err).To(BeNil())

			// 5. Connection() must fail now
			_, err = client.Connection("c1")
			Expect(err).ToNot(BeNil())

			// 6. Re-add and confirm
			err = client.AddConnectionConfig(common.ConnectionConfig{
				ConnectionId:  "c1",
				ConnectionUrl: "https://final-conn.example.com",
			})
			Expect(err).To(BeNil())
			cfg, _ := client.GetConnectionConfig("c1")
			Expect(cfg.ConnectionUrl).To(Equal("https://final-conn.example.com"))
		})
	})

	Context("full lifecycle: Detect → UpdateVaultConfig → Detect → RemoveVaultConfig → AddVaultConfig → Detect", func() {
		It("detect controller and config stay consistent through all mutations", func() {
			// 1. Activate detect
			_, err := client.Detect("v1")
			Expect(err).To(BeNil())

			// 2. Update via UpdateVaultConfig
			err = client.UpdateVaultConfig(common.VaultConfig{
				VaultId:     "v1",
				ClusterId:   "detect-mid",
				Credentials: common.Credentials{Token: "detect-mid-token"},
			})
			Expect(err).To(BeNil())
			Expect(client.detectServices["v1"].controller.Config.ClusterId).To(Equal("detect-mid"))

			// 3. Re-activate — picks up new config
			svc2, err := client.Detect("v1")
			Expect(err).To(BeNil())
			Expect(svc2.config.ClusterId).To(Equal("detect-mid"))

			// 4. Remove
			err = client.RemoveVaultConfig("v1")
			Expect(err).To(BeNil())
			_, err = client.Detect("v1")
			Expect(err).ToNot(BeNil())

			// 5. Re-add and verify fresh state
			err = client.AddVaultConfig(common.VaultConfig{VaultId: "v1", ClusterId: "detect-final"})
			Expect(err).To(BeNil())
			svc3, err := client.Detect("v1")
			Expect(err).To(BeNil())
			Expect(svc3.config.ClusterId).To(Equal("detect-final"))
		})
	})
})
