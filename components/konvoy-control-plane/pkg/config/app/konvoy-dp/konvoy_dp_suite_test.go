package konvoydp_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKonvoyDp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KonvoyDp Suite")
}
