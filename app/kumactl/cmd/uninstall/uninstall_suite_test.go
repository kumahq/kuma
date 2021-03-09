package uninstall_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInstallCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Uninstall Cmd Suite")
}
