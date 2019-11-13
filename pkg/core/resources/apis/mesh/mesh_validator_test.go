package mesh

import (
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mesh", func() {
	type testCase struct {
		mesh     string
		expected string
	}
	Describe("Validate()", func() {
		It("should pass validation", func() {
			// given
			spec := `
            mtls:
              enabled: true
              ca: {}
            logging:
              backends:
              - name: file-1
                file:
                  path: /path/to/file
              - name: tcp-1
                tcp:
                  address: kibana:1234
              defaultBackend: tcp-1
`
			mesh := MeshResource{}

			// when
			err := util_proto.FromYAML([]byte(spec), &mesh.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = mesh.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		DescribeTable("should validate fields",
			func(given testCase) {
				// given
				mesh := MeshResource{}

				// when
				err := util_proto.FromYAML([]byte(given.mesh), &mesh.Spec)
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
			Entry("nil ca when mtls is enabled", testCase{
				mesh: `
                mtls:
                  enabled: true`,
				expected: `
                violations:
                - field: mtls.ca
                  message: has to be set when mTLS is enabled`,
			}),
			Entry("logging backend with empty name", testCase{
				mesh: `
                logging:
                  backends:
                  - name:
                    tcp:
                      address: kibana:1234`,
				expected: `
                violations:
                - field: logging.backends[0].name
                  message: cannot be empty`,
			}),
			Entry("multiple logging backends of the same name", testCase{
				mesh: `
                logging:
                  backends:
                  - name: backend-1
                    tcp:
                      address: kibana:1234
                  - name: backend-1
                    file:
                      path: /path/to/file
                  defaultBackend: backend-1`,
				expected: `
                violations:
                - field: logging.backends[1].name
                  message: '"backend-1" name is already used for another backend'`,
			}),
			Entry("tcp logging address is empty", testCase{
				mesh: `
                logging:
                  backends:
                  - name: backend-1
                    tcp:
                      address:
                  defaultBackend: backend-1`,
				expected: `
                violations:
                - field: logging.backends[0].tcp.address
                  message: cannot be empty`,
			}),
			Entry("tcp logging address is invalid", testCase{
				mesh: `
                logging:
                  backends:
                  - name: backend-1
                    tcp:
                      address: wrong-format:234:234
                  defaultBackend: backend-1`,
				expected: `
                violations:
                - field: logging.backends[0].tcp.address
                  message: has to be in format of HOST:PORT`,
			}),
			Entry("file logging path is empty", testCase{
				mesh: `
                logging:
                  backends:
                  - name: backend-1
                    file:
                      path:
                  defaultBackend: backend-1`,
				expected: `
                violations:
                - field: logging.backends[0].file.path
                  message: cannot be empty`,
			}),
			Entry("default backend has to be set to one of the backends", testCase{
				mesh: `
                logging:
                  backends:
                  - name: backend-1
                    file:
                      path: /path
                  defaultBackend: non-existing-backend`,
				expected: `
                violations:
                - field: logging.defaultBackend
                  message: has to be set to one of the logging backend in mesh`,
			}),
			Entry("multiple errors", testCase{
				mesh: `
                mtls:
                  enabled: true
                logging:
                  backends:
                  - name:
                    file:
                      path: /path
                  - name: tcp-1
                    file:
                      tcp: invalid-address
                  - name: tcp-1
                    path:
                      address:
                  defaultBackend: invalid-backend`,
				expected: `
                violations:
                - field: mtls.ca
                  message: has to be set when mTLS is enabled
                - field: logging.backends[0].name
                  message: cannot be empty
                - field: logging.backends[1].file.path
                  message: cannot be empty
                - field: logging.backends[2].name
                  message: '"tcp-1" name is already used for another backend'
                - field: logging.defaultBackend
                  message: has to be set to one of the logging backend in mesh`,
			}),
		)
	})
})
