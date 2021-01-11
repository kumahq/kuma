package v3_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestXds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Xds V3 Suite")
}
