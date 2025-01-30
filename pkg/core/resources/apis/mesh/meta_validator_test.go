package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("Meta", func() {
	type testCase struct {
		meta           core_model.ResourceMeta
		scope          core_model.ResourceScope
		expected       string
	}

	Describe("ValidateMeta", func() {
		DescribeTable("should pass validation",
			func(given testCase) {
				Expect(mesh.ValidateMeta(given.meta, given.scope).Violations).To(BeEmpty())
			},
			Entry("mesh-scoped valid name and mesh", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "mesh-1", Name: "name-1"},
				scope: core_model.ScopeMesh,
			}),
			Entry("global-scoped valid name", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "", Name: "name-1"},
				scope: core_model.ScopeGlobal,
			}),
			Entry("global-scoped valid name with dot", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "", Name: "default.name-1"},
				scope: core_model.ScopeGlobal,
			}),
			Entry("mesh-scoped ", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "", Name: "default.name-1"},
				scope: core_model.ScopeGlobal,
			}),
		)

		DescribeTable("should validate fields",
			func(given testCase) {
				verr := mesh.ValidateMeta(given.meta, given.scope)
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty name", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "mesh-1", Name: ""},
				scope: core_model.ScopeMesh,
				expected: `
violations:
 - field: name
   message: cannot be empty`,
			}),
			Entry("empty mesh", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "", Name: "name-1"},
				scope: core_model.ScopeMesh,
				expected: `
violations:
 - field: mesh
   message: cannot be empty`,
			}),
			Entry("name with 2 dot", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "mesh-1", Name: "two..dots"},
				scope: core_model.ScopeMesh,
				expected: `
violations:
 - field: name
   message: invalid characters. A lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character`,
			}),
			Entry("name with underscore", testCase{
				meta:  &test_model.ResourceMeta{Mesh: "mesh-1", Name: "under_score"},
				scope: core_model.ScopeMesh,
				expected: `
violations:
 - field: name
   message: invalid characters. A lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character`,
			}),
			Entry("name longer than 253 character", testCase{
				meta: &test_model.ResourceMeta{
					Mesh: "mesh-1",
					Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
						"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
						"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
				scope: core_model.ScopeMesh,
				expected: `
violations:
 - field: name
   message: value length must less or equal 253`,
			}),
		)
	})
})
