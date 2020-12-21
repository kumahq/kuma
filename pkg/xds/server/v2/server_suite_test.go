package v2_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestXDSServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "XDS Server V2 Suite")
}
