package issuer_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIssuer(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunSpecs(t, "Issuer Suite")
}
