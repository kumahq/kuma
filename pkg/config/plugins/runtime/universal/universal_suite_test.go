package universal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUniversalConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Universal Config Suite")
}
