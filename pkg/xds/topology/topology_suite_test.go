package topology_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTopology(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Topology Suite")
}
