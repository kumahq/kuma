package mesh

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMeshManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mesh Manager Suite")
}
