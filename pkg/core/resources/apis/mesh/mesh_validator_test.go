package mesh

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("Mesh", func() {
	Describe("Validate()", func() {
		DescribeTable("should validate that name conforms to RFC 1035",
			func(name string, expected string) {
				// given
				mesh := NewMeshResource()
				mesh.SetMeta(&test_model.ResourceMeta{Name: name})

				// when
				verr := mesh.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(expected))
			},
			Entry("valid name", "mesh-1", "null"),
			Entry("name with a dot", "mesh.1", `
                violations:
                - field: name
                  message: a DNS-1035 label must consist of lower case alphanumeric characters
                    or '-', start with an alphabetic character, and end with an alphanumeric
                    character (e.g. 'my-name',  or 'abc-123', regex used for validation is
                    '[a-z]([-a-z0-9]*[a-z0-9])?')`),
			Entry("name starting with a digit", "1mesh", `
                violations:
                - field: name
                  message: a DNS-1035 label must consist of lower case alphanumeric characters
                    or '-', start with an alphabetic character, and end with an alphanumeric
                    character (e.g. 'my-name',  or 'abc-123', regex used for validation is
                    '[a-z]([-a-z0-9]*[a-z0-9])?')`),
			Entry("name with an underscore", "mesh_1", `
                violations:
                - field: name
                  message: a DNS-1035 label must consist of lower case alphanumeric characters
                    or '-', start with an alphabetic character, and end with an alphanumeric
                    character (e.g. 'my-name',  or 'abc-123', regex used for validation is
                    '[a-z]([-a-z0-9]*[a-z0-9])?')`),
			Entry("name of 63 characters", strings.Repeat("m", 63), "null"),
			Entry("name of 64 characters", strings.Repeat("m", 64), `
                violations:
                - field: name
                  message: must be no more than 63 characters`),
		)

		DescribeTable("should validate meshServices.mode",
			func(meshServices *mesh_proto.Mesh_MeshServices, expected string) {
				// given
				mesh := NewMeshResource()
				mesh.SetMeta(&test_model.ResourceMeta{Name: "mesh-1"})
				mesh.Spec.MeshServices = meshServices //nolint:staticcheck // deprecated on purpose

				// when
				verr := mesh.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(expected))
			},
			Entry("field not set", nil, "null"),
			Entry("Exclusive is accepted", &mesh_proto.Mesh_MeshServices{
				Mode: mesh_proto.Mesh_MeshServices_Exclusive,
			}, "null"),
			Entry("Disabled is rejected", &mesh_proto.Mesh_MeshServices{
				Mode: mesh_proto.Mesh_MeshServices_Disabled,
			}, `
                violations:
                - field: meshServices.mode
                  message: removed in 3.0 and every mesh behaves as Exclusive; remove the
                    field or set it to Exclusive`),
			Entry("ReachableBackends is rejected", &mesh_proto.Mesh_MeshServices{
				Mode: mesh_proto.Mesh_MeshServices_ReachableBackends,
			}, `
                violations:
                - field: meshServices.mode
                  message: removed in 3.0 and every mesh behaves as Exclusive; remove the
                    field or set it to Exclusive`),
		)
	})
})
