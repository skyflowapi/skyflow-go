package client_test

import (
	"fmt"
	vaultutils "skyflow-go/v2/utils/common"
	"skyflow-go/v2/utils/logger"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "skyflow-go/v2/client"
)

func TestServiceAccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client initialisation Suite")
}

var _ = Describe("Skyflow Builder", func() {
	var skyflowClient Skyflow
	BeforeEach(func() {
		skyflowClient = Skyflow{}
	})
	Context("ClientBuilder1", func() {
		It("should build a client with the correct configurations", func() {
			vaultConfig1 := vaultutils.VaultConfig{VaultId: "vault1"}
			builder1, err := skyflowClient.Builder().
				WithVaultConfig(vaultConfig1).
				WithVaultConfig(vaultConfig1).
				WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn1"}).
				WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
				WithLogLevel(logger.WARN).
				Build()

			expectNoError(err)
			Expect(builder1).NotTo(BeNil())
			log, _ := builder1.GetLoglevel()
			Expect(*log).To(Equal(logger.WARN))

			expectedVaultConfigs := map[string]vaultutils.VaultConfig{
				"vault1": vaultConfig1,
			}
			config, err := builder1.GetVaultConfig("vault1")
			Expect(config).To(Equal(expectedVaultConfigs["vault1"]))
			Expect(err).To(BeNil())
			Expect(config.Env).To(Equal(expectedVaultConfigs["vault1"].Env))
		})
	})
	Context("ClientBuilder2", func() {
		It("should handle multiple vault configurations", func() {
			vaultConfig2 := vaultutils.VaultConfig{VaultId: "vault2", Env: vaultutils.DEV}
			vaultConfig3 := vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.STAGE}
			builder2, err := skyflowClient.Builder().
				WithVaultConfig(vaultConfig2).
				WithVaultConfig(vaultConfig3).
				WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn2"}).
				WithSkyflowCredentials(vaultutils.Credentials{Token: "token2"}).
				WithLogLevel(logger.ERROR).Build()

			expectNoError(err)
			Expect(builder2).NotTo(BeNil())
			log, err := builder2.GetLoglevel()
			Expect(*log).To(Equal(logger.ERROR))

			expectedVaultConfigs := map[string]vaultutils.VaultConfig{
				"vault2": vaultConfig2,
				"vault3": vaultConfig3,
			}
			v2config, err := builder2.GetVaultConfig("vault2")
			expectNoError(err)
			Expect(v2config).To(Equal(expectedVaultConfigs["vault2"]))
			Expect(v2config.Env).To(Equal(vaultutils.DEV))

		})
	})
	Context("CompareClientBuilders", func() {
		It("should verify that two builders are different", func() {
			vaultConfig1 := vaultutils.VaultConfig{VaultId: "vault1"}
			vaultConfig2 := vaultutils.VaultConfig{VaultId: "vault2", Env: vaultutils.DEV}
			vaultConfig3 := vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.STAGE}

			builder1, _ := skyflowClient.Builder().
				WithVaultConfig(vaultConfig1).
				WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn1"}).
				WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
				WithLogLevel(logger.WARN).Build()

			builder2, _ := skyflowClient.Builder().
				WithVaultConfig(vaultConfig2).
				WithVaultConfig(vaultConfig3).
				WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn2"}).
				WithSkyflowCredentials(vaultutils.Credentials{Token: "token2"}).
				WithLogLevel(logger.ERROR).Build()

			Expect(builder1).NotTo(Equal(builder2))
			b1config, err := builder1.GetVaultConfig("vault1")
			Expect(err).To(BeNil())
			b2config, err := builder2.GetVaultConfig("vault2")
			Expect(err).To(BeNil())

			// check vault1 is present or not in b2
			b3config, err := builder2.GetVaultConfig("vault1")
			Expect(err).ToNot(BeNil())
			Expect(b3config).To(Equal(vaultutils.VaultConfig{}))
			Expect(b1config).NotTo(Equal(b2config))
		})
	})
	Context("DeleteFromVaultConfig", func() {
		It("should delete a vault configuration and verify the update", func() {
			vaultConfig1 := vaultutils.VaultConfig{VaultId: "vault1"}
			vaultConfig3 := vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.STAGE}

			builder1, _ := skyflowClient.Builder().
				WithVaultConfig(vaultConfig1).
				WithVaultConfig(vaultConfig3).
				WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn1"}).
				WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
				WithLogLevel(logger.WARN).Build()

			initialVaultConfigs := map[string]vaultutils.VaultConfig{
				"vault1": vaultConfig1,
				"vault3": vaultConfig3,
			}
			config, err := builder1.GetVaultConfig("vault1")
			Expect(err).To(BeNil())
			Expect(config).To(Equal(initialVaultConfigs["vault1"]))

			builder1.RemoveVaultConfig("vault1")
			config, err = builder1.GetVaultConfig("vault1")
			Expect(err).ToNot(BeNil())
			Expect(config).To(Equal(vaultutils.VaultConfig{}))
			// remove when vault config not present
			err1 := builder1.RemoveVaultConfig("vault1")
			Expect(err1).ToNot(BeNil())
			config, err = builder1.GetVaultConfig("vault1")
			Expect(err).ToNot(BeNil())
			Expect(config).To(Equal(vaultutils.VaultConfig{}))
		})
	})
	Context("ClientBuilder1", func() {
		It("should build a client with the correct configurations", func() {
			vaultConfig1 := vaultutils.VaultConfig{VaultId: "vault1"}
			builder1, err := skyflowClient.Builder().
				WithVaultConfig(vaultConfig1).
				WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "conn1"}).
				WithSkyflowCredentials(vaultutils.Credentials{Token: "token1"}).
				WithLogLevel(logger.WARN).
				Build()

			expectNoError(err)
			Expect(builder1).NotTo(BeNil())
			log, _ := builder1.GetLoglevel()
			Expect(*log).To(Equal(logger.WARN))

			expectedVaultConfigs := map[string]vaultutils.VaultConfig{
				"vault1": vaultConfig1,
			}
			config, err := builder1.GetVaultConfig("vault1")
			Expect(config).To(Equal(expectedVaultConfigs["vault1"]))
			Expect(err).To(BeNil())
			Expect(config.Env).To(Equal(expectedVaultConfigs["vault1"].Env))
		})
	})
	Context("Test LogLevel and credentials", func() {
		It("should return the correct default log level", func() {
			builder, _ := skyflowClient.Builder().Build()
			logLevel, err := builder.GetLoglevel()
			expectNoError(err)
			Expect(*logLevel).To(Equal(logger.DEBUG))
		})

		It("should get the log level correctly", func() {
			builder, err := skyflowClient.Builder().WithLogLevel(logger.DEBUG).Build()
			expectNoError(err)
			logLevel, err := builder.GetLoglevel()
			expectNoError(err)
			Expect(*logLevel).To(Equal(logger.DEBUG))
		})
		It("should update the log level correctly", func() {
			builder, err := skyflowClient.Builder().WithLogLevel(logger.DEBUG).Build()
			expectNoError(err)
			logLevel, err := builder.GetLoglevel()
			expectNoError(err)
			Expect(*logLevel).To(Equal(logger.DEBUG))

			builder.UpdateLogLevel(logger.WARN)
			logLevel, err = builder.GetLoglevel()
			expectNoError(err)
			Expect(*logLevel).To(Equal(logger.WARN))
		})
		It("should update the config at skyflow client level correctly", func() {
			builder, err := skyflowClient.Builder().WithLogLevel(logger.DEBUG).WithSkyflowCredentials(vaultutils.Credentials{
				Token: "token1",
			}).Build()
			expectNoError(err)

			builder.UpdateSkyflowCredentials(vaultutils.Credentials{Token: "token2"})
		})

	})
	Context("RemoveConnectionConfig", func() {
		It("should remove an existing connection configuration", func() {
			skyflowClient = Skyflow{}
			connectionConfig := vaultutils.ConnectionConfig{ConnectionId: "id", ConnectionUrl: "https://demo.com", Credentials: vaultutils.Credentials{
				Token: "token1",
			}}
			builder, err := skyflowClient.Builder().WithConnectionConfig(connectionConfig).WithConnectionConfig(connectionConfig).Build()
			expectNoError(err)
			config1, err1 := builder.GetConnectionConfig("id")
			expectNoError(err1)
			Expect(config1.ConnectionId).To(Equal("id"))
			errr := builder.RemoveConnectionConfig("id")
			Expect(errr).To(BeNil())
			config2, err2 := builder.GetConnectionConfig("id")
			Expect(err2).ToNot(BeNil())
			Expect(config2).To(Equal(vaultutils.ConnectionConfig{}))

			// remove deleted config
			errr = builder.RemoveConnectionConfig("id")
			Expect(errr).ToNot(BeNil())
			config2, err2 = builder.GetConnectionConfig("id")
			Expect(err2).ToNot(BeNil())
			Expect(config2).To(Equal(vaultutils.ConnectionConfig{}))

			// add config
			errr1 := builder.AddConnectionConfig(vaultutils.ConnectionConfig{
				ConnectionId:  "id2",
				ConnectionUrl: "https://demo.com/",
				Credentials: vaultutils.Credentials{
					Token: "token",
				},
			})
			Expect(errr1).To(BeNil())
			err3 := builder.UpdateConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "id2", ConnectionUrl: "https://demo2.com"})
			Expect(err3).To(BeNil())

			c, err4 := builder.GetConnectionConfig("id2")
			Expect(err4).To(BeNil())
			Expect(c.ConnectionId).To(Equal("id2"))
			Expect(c.ConnectionUrl).To(Equal("https://demo2.com"))

			// add already existing config
			errr = builder.AddConnectionConfig(vaultutils.ConnectionConfig{
				ConnectionId:  "id2",
				ConnectionUrl: "https://demo.com",
				Credentials: vaultutils.Credentials{
					Token: "token",
				},
			})
			Expect(errr).ToNot(BeNil())

			// update config that is not present
			err3 = builder.UpdateConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "id4", ConnectionUrl: "https://demo2.com", Credentials: vaultutils.Credentials{
				Token: "token",
			}})
			Expect(err3).ToNot(BeNil())
		})
	})
	Context("AddVaultConfig", func() {
		It("should add a new vault configuration", func() {
			vaultConfig := vaultutils.VaultConfig{VaultId: "newVault", Env: vaultutils.DEV, ClusterId: "id1", Credentials: vaultutils.Credentials{Token: "token1"}}
			builder, err := skyflowClient.Builder().Build()
			Expect(err).To(BeNil())
			err1 := builder.AddVaultConfig(vaultConfig)
			expectNoError(err1)

			config, err2 := builder.GetVaultConfig("newVault")
			expectNoError(err2)
			Expect(config).To(Equal(vaultConfig))
		})

		//It("should return an error when adding a duplicate vault configuration", func() {
		//	vaultConfig := vaultutils.VaultConfig{VaultId: "vault1"}
		//	builder, err := skyflowClient.Builder().WithVaultConfig(vaultConfig).Build()
		//	Expect(err).To(BeNil())
		//	err = builder.AddVaultConfig(vaultConfig)
		//	conf, errr := builder.GetVaultConfig(vaultConfig.VaultId)
		//	Expect(errr).To(BeNil())
		//	Expect(conf).To(Equal(vaultConfig))
		//})
	})
	Context("UpdateVaultConfig", func() {
		It("should update an existing vault configuration", func() {
			vaultConfig := vaultutils.VaultConfig{VaultId: "vault1", Env: vaultutils.DEV, ClusterId: "id", Credentials: vaultutils.Credentials{Token: "token1"}}
			builder, errr := skyflowClient.Builder().WithVaultConfig(vaultConfig).Build()
			Expect(errr).To(BeNil())
			updatedConfig := vaultutils.VaultConfig{VaultId: "vault1", Env: vaultutils.PROD, ClusterId: "id1", Credentials: vaultutils.Credentials{Token: "token1"}}
			err := builder.UpdateVaultConfig(updatedConfig)
			expectNoError(err)
			fmt.Println("errrtrtrtrtrt", err)
			config, err2 := builder.GetVaultConfig("vault1")
			expectNoError(err2)
			Expect(config).To(Equal(updatedConfig))
		})

		It("should return an error when updating a non-existing vault configuration", func() {
			updatedConfig := vaultutils.VaultConfig{VaultId: "nonExistentVault", Env: vaultutils.PROD}
			builder, errr := skyflowClient.Builder().Build()
			Expect(errr).To(BeNil())
			err := builder.UpdateVaultConfig(updatedConfig)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("Test Vault method", func() {
		It("should create a new Vault method", func() {
			vaultConfig := vaultutils.VaultConfig{VaultId: "vault1", Env: vaultutils.DEV, Credentials: vaultutils.Credentials{
				Token: "token1",
			}}
			builder, errr := skyflowClient.Builder().WithVaultConfig(vaultConfig).Build()
			Expect(errr).To(BeNil())
			service, err := builder.Vault("vault1")
			Expect(err).To(BeNil())
			Expect(service).NotTo(BeNil())

			service, err = builder.Vault("vault2")
			Expect(err).ToNot(BeNil())
			Expect(service).To(BeNil())

			er := builder.AddVaultConfig(vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.DEV, Credentials: vaultutils.Credentials{Token: "token"}, ClusterId: "id"})

			service, err = builder.Vault("vault3")
			Expect(er).To(BeNil())
			Expect(service).ToNot(Equal(vaultutils.VaultConfig{}))

			// remove all configs

			err = builder.RemoveVaultConfig(vaultConfig.VaultId)
			Expect(err).To(BeNil())
			err = builder.RemoveVaultConfig("vault3")
			Expect(err).To(BeNil())

			service, err = builder.Vault("vault3")
			Expect(er).To(BeNil())
			Expect(service).ToNot(Equal(vaultutils.VaultConfig{}))

			er = builder.AddVaultConfig(vaultutils.VaultConfig{VaultId: "vault3", Env: vaultutils.DEV, Credentials: vaultutils.Credentials{Token: "token"}, ClusterId: "id"})
			service, err = builder.Vault()
			Expect(er).To(BeNil())
			Expect(service).ToNot(Equal(vaultutils.VaultConfig{}))
		})
	})
	Context("Test Connection method", func() {
		It("should create vault service", func() {
			skyflowClient = Skyflow{}
			connectionConfig := vaultutils.ConnectionConfig{ConnectionId: "id", Credentials: vaultutils.Credentials{Token: "token"}}
			builder, err := skyflowClient.Builder().WithConnectionConfig(connectionConfig).WithConnectionConfig(vaultutils.ConnectionConfig{ConnectionId: "id4"}).Build()
			expectNoError(err)
			service, err := builder.Connection()
			Expect(err).To(BeNil())
			Expect(service).NotTo(BeNil())

			service, err = builder.Connection("id")
			Expect(err).To(BeNil())
			Expect(service).NotTo(BeNil())

			service, err = builder.Connection("id2")
			Expect(err).ToNot(BeNil())
			Expect(service).To(BeNil())

			service, err = builder.Connection("id4")
			Expect(err).ToNot(BeNil())
			Expect(service).To(BeNil())

			err = builder.RemoveConnectionConfig("id4")
			Expect(err).To(BeNil())
			err = builder.RemoveConnectionConfig("id")
			Expect(err).To(BeNil())

			service, err = builder.Connection()
			Expect(err).ToNot(BeNil())
			Expect(service).To(BeNil())

		})
	})
})

func expectNoError(err error) {
	Expect(err).NotTo(HaveOccurred())
}
