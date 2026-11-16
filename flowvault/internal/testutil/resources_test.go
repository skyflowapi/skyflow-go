package testutil

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTestutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowvault Testutil Suite")
}

var _ = Describe("Resource", func() {
	It("returns a fixture's contents without surrounding whitespace", func() {
		key := Resource("client_api_key.txt")
		Expect(key).ToNot(BeEmpty())
		Expect(key).To(Equal(strings.TrimSpace(key)))
	})

	It("panics when the fixture doesn't exist", func() {
		Expect(func() { Resource("missing.txt") }).To(Panic())
	})
})
