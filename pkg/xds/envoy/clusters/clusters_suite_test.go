package clusters_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusters Suite")
}
