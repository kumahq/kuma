package persistence_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDNSPersistence(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DNS Persistence Suite")
}
