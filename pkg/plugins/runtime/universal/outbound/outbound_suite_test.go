package outbound_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOutbound(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Universal outbound test")
}
