package v3_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestListener(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Listener V3 Suite")
}
