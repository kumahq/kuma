package provided_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCaProvided(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CA Provided Suite")
}
