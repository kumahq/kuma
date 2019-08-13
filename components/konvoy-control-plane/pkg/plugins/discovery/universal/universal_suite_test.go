package universal

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUniversal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Universal Discovery Plugin test")
}
