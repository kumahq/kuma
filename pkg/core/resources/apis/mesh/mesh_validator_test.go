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
			Entry("multiple ca backends of the same name", testCase{
				mesh: `
            mtls:
              enabledBackend: builtin-1
              backends:
              - name: builtin-1
                type: builtin
                dpCert:
                  rotation:
                    expiration: 2y
            logging:
              backends:
              - name: file-1
                type: file
                conf:
                  path: /path/to/file
              - name: file-2
                format: '%START_TIME% %KUMA_SOURCE_SERVICE%'
                type: file
                conf:
                  path: /path/to/file2
              - name: tcp-1
                type: tcp
                conf:
                  address: kibana:1234
              - name: tcp-2
                format: '%START_TIME% %KUMA_DESTINATION_SERVICE%'
                type: tcp
                conf:
                  address: kibana:1234
              defaultBackend: tcp-1
            constraints:
              dataplaneProxy:
                requirements:
                - tags:
                    k8s.kuma.io/namespace: ns-1
                    kuma.io/zone: east
                restrictions:
                - tags:
                    k8s.kuma.io/namespace: ns-1
                    kuma.io/zone: west
            routing:
              zoneEgress: true`,
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
			Entry("constraints dataplaneProxy with invalid tags", testCase{
				mesh: `
                constraints:
                  dataplaneProxy:
                    requirements:
                    - {}
                    - tags:
                        '': ''
                    - tags:
                        '!@#$': '!@#$'
                    restrictions:
                    - {}
                    - tags:
                        '': ''
                    - tags:
                        '!@#$': '!@#$'
`,
				expected: `
                violations:
                - field: constraints.dataplaneProxy.requirements[0].tags
                  message: must have at least one tag
                - field: constraints.dataplaneProxy.requirements[1].tags
                  message: tag name must be non-empty
                - field: constraints.dataplaneProxy.requirements[1].tags[""]
                  message: tag value must be non-empty
                - field: constraints.dataplaneProxy.requirements[2].tags["!@#$"]
                  message: tag name must consist of alphanumeric characters, dots, dashes, slashes
                    and underscores
                - field: constraints.dataplaneProxy.requirements[2].tags["!@#$"]
                  message: tag value must consist of alphanumeric characters, dots, dashes, slashes
                    and underscores or be "*"
                - field: constraints.dataplaneProxy.restrictions[0].tags
                  message: must have at least one tag
                - field: constraints.dataplaneProxy.restrictions[1].tags
                  message: tag name must be non-empty
                - field: constraints.dataplaneProxy.restrictions[1].tags[""]
                  message: tag value must be non-empty
                - field: constraints.dataplaneProxy.restrictions[2].tags["!@#$"]
                  message: tag name must consist of alphanumeric characters, dots, dashes, slashes
                    and underscores
                - field: constraints.dataplaneProxy.restrictions[2].tags["!@#$"]
                  message: tag value must consist of alphanumeric characters, dots, dashes, slashes
                    and underscores or be "*"`,
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
			Entry("zoneEgress enabled but mtls not defined", testCase{
				mesh: `
                routing:
                  zoneEgress: true`,
				expected: `
                violations:
                - field: mtls
                  message: has to be set when zoneEgress enabled`,
			}),
			Entry("zoneEgress enabled but multiple ca backends of the same name", testCase{
				mesh: `
                mtls:
                  enabledBackend: backend-1
                  backends:
                  - name: backend-1
                    type: builtin
                  - name: backend-1
                    type: builtin
                routing:
                  zoneEgress: true`,
				expected: `
                violations:
                - field: mtls.backends
                  message: cannot have more than 1 backends
                - field: mtls.backends[1].name
                  message: '"backend-1" name is already used for another backend'`,
			}),
			Entry("zoneEgress and mTLS enabled but no enabledBackend provided", testCase{
				mesh: `
                mtls:
                  backends:
                  - name: backend-1
                    type: builtin
                routing:
                  zoneEgress: true`,
				expected: `
                violations:
                - field: mtls
                  message: has to be set when zoneEgress enabled`,
			}),
		)
	})
})
