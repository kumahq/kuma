package mesh_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMeshSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MeshSnapshot Suite")
}
