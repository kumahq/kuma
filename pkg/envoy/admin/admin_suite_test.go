package admin_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEnvoyAdmin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EnvoyAdmin Suite")
}
