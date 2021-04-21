package v1_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestXds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prometheus SD V1 Suite")
}
