package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfigCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Cmd Suite")
}
