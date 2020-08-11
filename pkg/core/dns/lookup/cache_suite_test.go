package lookup_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDNSCaching(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DNS with cache Suite")
}
