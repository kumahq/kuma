package mesh

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

var _ = Describe("Mesh", func() {
	type testCase struct {
		mesh     string
		expected string
	}
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(given testCase) {
				// given
				mesh := NewMeshResource()

				// when
				err := util_proto.FromYAML([]byte(given.mesh), mesh.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				err = mesh.Validate()

				// then
				Expect(err).ToNot(HaveOccurred())
			},
			Entry("valid mtls dpCert rotation", testCase{
				mesh: `
            mtls:
              enabledBackend: builtin-1
              backends:
              - name: builtin-1
                type: builtin
                dpCert:
                  rotation:
                    expiration: 2y`,
				expected: "",
			}),
		)

		DescribeTable("should validate fields",
			func(given testCase) {
				// given
				mesh := NewMeshResource()

				// when
				err := util_proto.FromYAML([]byte(given.mesh), mesh.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := mesh.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("multiple ca backends of the same name", testCase{
				mesh: `
                mtls:
                  enabledBackend: backend-1
                  backends:
                  - name: backend-1
                    type: builtin
                  - name: backend-1
                    type: builtin`,
				expected: `
                violations:
                - field: mtls.backends
                  message: cannot have more than 1 backends
                - field: mtls.backends[1].name
                  message: '"backend-1" name is already used for another backend'`,
			}),
			Entry("enabledBackend of unknown name", testCase{
				mesh: `
                mtls:
                  enabledBackend: backend-2
                  backends:
                  - name: backend-1
                    type: builtin`,
				expected: `
                violations:
                - field: mtls.enabledBackend
                  message: has to be set to one of the backends in the mesh`,
			}),
			Entry("dpCert rotation invalid expiration time", testCase{
				mesh: `
                mtls:
                  enabledBackend: backend-3
                  backends:
                  - name: backend-3
                    type: builtin
                    dpCert:
                      rotation:
                        expiration: 2e`,
				expected: `
                violations:
                - field: mtls.dpcert.rotation.expiration
                  message: has to be a valid format`,
			}),
			Entry("multiple errors", testCase{
				mesh: `
                mtls:
                  enabledBackend: invalid-backend`,
				expected: `
                violations:
                - field: mtls.enabledBackend
                  message: has to be set to one of the backends in the mesh`,
			}),
		)
	})
})
