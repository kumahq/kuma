package envoy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Envoy Suite")
}
