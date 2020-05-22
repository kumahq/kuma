package manager_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSecretManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Secret Manager Suite")
}
