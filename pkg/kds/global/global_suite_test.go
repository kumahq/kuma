package global_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGlobalSync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Global Sync Suite")
}
