package mesh_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMeshValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mesh Validators Suite")
}
