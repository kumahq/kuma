package issuer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test"
)

func TestIssuer(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunSpecs(t, "Issuer Suite")
}
