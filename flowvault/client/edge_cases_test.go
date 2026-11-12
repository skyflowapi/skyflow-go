package client

import (
	"bytes"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fvlogger "github.com/skyflowapi/skyflow-go/flowvault/utils/logger"
)

var _ = Describe("Skyflow client, edge cases", func() {
	It("sets, updates and returns the log level", func() {
		DeferCleanup(fvlogger.SetLogLevel, fvlogger.ERROR)
		client := newTestClient(WithLogLevel(fvlogger.DEBUG))
		Expect(*client.GetLoglevel()).To(Equal(fvlogger.DEBUG))

		var logs bytes.Buffer
		fvlogger.SetOutput(&logs)
		DeferCleanup(fvlogger.SetOutput, os.Stderr)
		client.UpdateLogLevel(fvlogger.INFO)
		Expect(logs.String()).To(ContainSubstring("Current log level is %v.", fvlogger.INFO))

		client.UpdateLogLevel(fvlogger.WARN)
		level := client.GetLoglevel()
		Expect(*level).To(Equal(fvlogger.WARN))
		*level = fvlogger.OFF
		Expect(*client.GetLoglevel()).To(Equal(fvlogger.WARN))
	})

	It("sets and returns the log level with SetLogLevel and GetLogLevel, as in Java and Python", func() {
		DeferCleanup(fvlogger.SetLogLevel, fvlogger.ERROR)
		var logs bytes.Buffer
		fvlogger.SetOutput(&logs)
		DeferCleanup(fvlogger.SetOutput, os.Stderr)
		client := newTestClient(WithLogLevel(fvlogger.INFO))
		Expect(client.GetLogLevel()).To(Equal(fvlogger.INFO))
		Expect(logs.String()).To(ContainSubstring("Current log level is %v.", fvlogger.INFO))
		client.SetLogLevel(fvlogger.DEBUG)
		Expect(client.GetLogLevel()).To(Equal(fvlogger.DEBUG))
		Expect(*client.GetLoglevel()).To(Equal(fvlogger.DEBUG))
	})

	It("uses the first vault added as the default vault, as vault() does in Java and Python", func() {
		client := newTestClient(WithVaults(fvcommon.VaultConfig{VaultId: "second", ClusterId: "c2"}))
		Expect(client.AddVaultConfig(fvcommon.VaultConfig{VaultId: "third", ClusterId: "c3"})).To(BeNil())
		first, err := client.Vault(testVaultID)
		Expect(err).To(BeNil())
		defaultVault, err := client.Vault("")
		Expect(err).To(BeNil())
		Expect(defaultVault).To(BeIdenticalTo(first))
		config, err := client.GetVaultConfig("")
		Expect(err).To(BeNil())
		Expect(config.VaultId).To(Equal(testVaultID))

		// Removing the default makes the next vault added the default.
		Expect(client.RemoveVaultConfig(testVaultID)).To(BeNil())
		config, err = client.GetVaultConfig("")
		Expect(err).To(BeNil())
		Expect(config.VaultId).To(Equal("second"))
		Expect(client.RemoveVaultConfig("third")).To(BeNil())
		config, _ = client.GetVaultConfig("")
		Expect(config.VaultId).To(Equal("second"))

		Expect(client.RemoveVaultConfig("second")).To(BeNil())
		_, err = client.Vault("")
		Expect(err.GetMessage()).To(ContainSubstring("VaultId is missing"))
		_, err = client.GetVaultConfig("")
		Expect(err.GetMessage()).To(ContainSubstring("VaultId is missing"))
	})

	It("reports vault IDs that aren't registered", func() {
		client := newTestClient()
		_, err := client.Vault("unknown")
		Expect(err.GetMessage()).To(ContainSubstring("VaultId is missing"))
		_, err = client.GetVaultConfig("unknown")
		Expect(err.GetMessage()).To(ContainSubstring("VaultId is missing"))
		Expect(client.RemoveVaultConfig("unknown").GetMessage()).To(ContainSubstring("VaultId is missing"))
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: "unknown"}).GetMessage()).To(ContainSubstring("VaultId is missing"))
	})

	It("rejects an update without a vault ID or with negative HTTP settings", func() {
		client := newTestClient()
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{ClusterId: "c"})).ToNot(BeNil())
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, MaxRetryDelay: -time.Second}).GetMessage()).
			To(ContainSubstring("must not be negative"))
	})

	It("rejects invalid vault-level credentials on add and update", func() {
		client := newTestClient()
		twoSources := fvcommon.Credentials{ApiKey: testApiKey, Token: "token"}
		Expect(client.AddVaultConfig(fvcommon.VaultConfig{VaultId: "other", ClusterId: "c", Credentials: twoSources})).ToNot(BeNil())
		Expect(client.AddVaultConfig(fvcommon.VaultConfig{VaultId: "other", ClusterId: "c", Credentials: fvcommon.Credentials{ApiKey: "bad"}})).ToNot(BeNil())
		_, err := client.Vault("other")
		Expect(err).ToNot(BeNil())

		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, Credentials: twoSources})).ToNot(BeNil())
		config, _ := client.GetVaultConfig(testVaultID)
		Expect(config.Credentials.Token).To(BeEmpty())
	})

	It("rejects an invalid base vault URL on update", func() {
		client := newTestClient()
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, BaseVaultUrl: "not a url"})).ToNot(BeNil())
		config, _ := client.GetVaultConfig(testVaultID)
		Expect(config.BaseVaultUrl).To(BeEmpty())
	})

	It("applies the deprecated BaseVaultURL on update and keeps the URL when neither is set", func() {
		client := newTestClient()
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, BaseVaultUrl: "https://old.example.com"})).To(BeNil())
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID, BaseVaultURL: "https://new.example.com"})).To(BeNil())
		config, _ := client.GetVaultConfig(testVaultID)
		Expect(config.BaseVaultUrl).To(BeEmpty())
		Expect(config.BaseVaultURL).To(Equal("https://new.example.com"))

		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{VaultId: testVaultID})).To(BeNil())
		config, _ = client.GetVaultConfig(testVaultID)
		Expect(config.BaseVaultURL).To(Equal("https://new.example.com"))
	})

	It("updates credentials, base URL and retry delays without sharing the caller's values", func() {
		client := newTestClient()
		roles := []string{"r1"}
		Expect(client.UpdateVaultConfig(fvcommon.VaultConfig{
			VaultId:           testVaultID,
			Credentials:       fvcommon.Credentials{ApiKey: testApiKey, Roles: roles},
			BaseVaultUrl:      "https://vault.example.com",
			InitialRetryDelay: 100 * time.Millisecond,
			MaxRetryDelay:     time.Second,
			ConnectTimeout:    2 * time.Second,
			ReadTimeout:       3 * time.Second,
			WriteTimeout:      4 * time.Second,
		})).To(BeNil())
		roles[0] = "changed"

		config, err := client.GetVaultConfig(testVaultID)
		Expect(err).To(BeNil())
		Expect(config.Credentials.ApiKey).To(Equal(testApiKey))
		Expect(config.Credentials.Roles).To(Equal([]string{"r1"}))
		Expect(config.BaseVaultUrl).To(Equal("https://vault.example.com"))
		Expect(config.InitialRetryDelay).To(Equal(100 * time.Millisecond))
		Expect(config.MaxRetryDelay).To(Equal(time.Second))
		Expect([]time.Duration{config.ConnectTimeout, config.ReadTimeout, config.WriteTimeout}).
			To(Equal([]time.Duration{2 * time.Second, 3 * time.Second, 4 * time.Second}))
		Expect(config.ClusterId).To(Equal("cluster"))
	})

	It("rejects invalid client credentials and has none until they are set", func() {
		_, err := NewSkyflow(WithCredentials(fvcommon.Credentials{}))
		Expect(err).ToNot(BeNil())

		client, err := NewSkyflow(WithVaults(fvcommon.VaultConfig{VaultId: testVaultID, ClusterId: "cluster"}))
		Expect(err).To(BeNil())
		Expect(client.GetSkyflowCredentials()).To(BeNil())
		Expect(client.AddSkyflowCredentials(fvcommon.Credentials{ApiKey: testApiKey, Token: "both"})).ToNot(BeNil())
		Expect(client.GetSkyflowCredentials()).To(BeNil())
	})

	It("deep-copies list contexts", func() {
		original := []interface{}{map[string]interface{}{"id": "u1"}, []interface{}{"a"}, "plain"}
		copied := copyValue(original).([]interface{})
		Expect(copied).To(Equal(original))
		copied[0].(map[string]interface{})["id"] = "changed"
		copied[1].([]interface{})[0] = "changed"
		Expect(original[0]).To(Equal(map[string]interface{}{"id": "u1"}))
		Expect(original[1]).To(Equal([]interface{}{"a"}))
		Expect(copyValue(nil)).To(BeNil())
		Expect(copyValue(7)).To(Equal(7))
	})
})
