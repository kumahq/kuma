package mesh_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMesh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Defaults Mesh")
}
