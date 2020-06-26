package dns_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDNS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DNS Suite")
}
