package generate_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGenerateDataplaneTokenCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generate Dataplane Token Suite")
}
