package mesh_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
)

// The OpenAPI specs advertise the name and mesh constraints the API enforces.
// Two of the three places that carry them are YAML, so nothing but this test
// stops them drifting from the validator.
var _ = Describe("name constraints in the OpenAPI specs", func() {
	repoRoot := filepath.Join("..", "..", "..", "..", "..")

	DescribeTable("should match the validator",
		func(path string) {
			raw, err := os.ReadFile(filepath.Join(repoRoot, path))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(raw)).To(ContainSubstring(core_mesh.NamePattern))
			Expect(string(raw)).To(ContainSubstring(core_mesh.MeshNamePattern))
		},
		Entry("policy schema template", filepath.Join("tools", "openapi", "templates", "schema.yaml")),
		Entry("shared metadata schema", filepath.Join("api", "openapi", "specs", "common", "resource.yaml")),
	)
})
