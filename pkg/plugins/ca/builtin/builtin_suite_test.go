package builtin_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCaBuiltin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CA Builtin Suite")
}
