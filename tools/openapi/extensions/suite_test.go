package extensions

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Bootstrapped with plain ginkgo rather than pkg/test, to keep the generator
// testable without pulling the control plane into a build tool's test binary.
func TestExtensions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenAPI Extensions Suite")
}
