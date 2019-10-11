package generate_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGenerateDpTokenCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generate Dp Token Suite")
}
