package extensions

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Bootstrapped with plain ginkgo rather than pkg/test, which reaches the resource
// APIs that register extensions and would make this an import cycle.
func TestExtensions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Extensions Suite")
}
