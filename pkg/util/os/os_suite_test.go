package os

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OS Suite")
}
