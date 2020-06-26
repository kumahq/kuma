package manager_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfigManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Manager Suite")
}
