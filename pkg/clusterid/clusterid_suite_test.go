package clusterid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterID(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster ID Suite")
}
