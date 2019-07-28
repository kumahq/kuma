package konvoyinjector_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKonvoyInjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KonvoyInjector Suite")
}
