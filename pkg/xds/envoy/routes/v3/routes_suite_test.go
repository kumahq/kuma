package v3_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRoutes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Routes V3 Suite")
}
