package blackbox_network_tests_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTransparentProxy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blackbox network tests Suite")
}
