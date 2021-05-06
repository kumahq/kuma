package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("VirtualOutbound", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(virtualOutbound string) {
				// setup
				virtualOubound := NewVirtualOutboundResource()

				// when
				err := util_proto.FromYAML([]byte(virtualOutbound), virtualOubound.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := virtualOubound.Validate()

				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full example", `
                selectors:
                - match:
                    region: eu
                conf:
                  host: "foo.bar"
                  port: "80"`,
			),
			Entry("empty host", `
                selectors:
                - match:
                    region: eu
                conf:
                  host: "{{.service}}.mesh"
                  port: "{{.port}}"
                  parameters:
                    service: kuma.io/service
                    port: kuma.io/port
`,
			),
		)

		type testCase struct {
			input    string
			expected string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				virtualOutbound := NewVirtualOutboundResource()

				// when
				err := util_proto.FromYAML([]byte(given.input), virtualOutbound.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := virtualOutbound.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				input: ``,
				expected: `
                violations:
                - field: selectors
                  message: must have at least one element
                - field: conf
                  message: has to be defined 
`,
			}),
			Entry("selectors without tags", testCase{
				input: `
                selectors:
                - match: {}
                conf:
                  host: "foo.mesh"
                  port: "80"
`,
				expected: `
                violations:
                - field: selectors[0].match
                  message: must have at least one tag
`,
			}),
			Entry("selectors with empty tags values", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service:
                    region:
                conf:
                  host: "foo.mesh"
                  port: "80"
`,
				expected: `
                violations:
                - field: selectors[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: selectors[0].match["region"]
                  message: tag value must be non-empty
`,
			}),
			Entry("multiple selectors", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service:
                    region:
                - match: {}
                conf:
                  host: "foo.mesh"
                  port: "80"
`,
				expected: `
                violations:
                - field: selectors[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: selectors[0].match["region"]
                  message: tag value must be non-empty
                - field: selectors[1].match
                  message: must have at least one tag
`,
			}),
			Entry("bad host template", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.{{mesh"
                  port: "80"
`,
				expected: `
                violations:
                - field: conf.host 
                  message: 'template pre evaluation failed with error=''failed compiling gotemplate error=''template: :1: function "mesh" not defined'''''
`,
			}),
			Entry("bad port template", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.mesh"
                  port: "{{port"
`,
				expected: `
                violations:
                - field: conf.port 
                  message: 'template pre evaluation failed with error=''failed compiling gotemplate error=''template: :1: function "port" not defined'''''
`,
			}),
			Entry("port is not a number template", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.mesh"
                  port: "{{.port}}a"
                  parameters:
                    port: "service"
`,
				expected: `
                violations:
                - field: conf.port 
                  message: template pre evaluation failed with error='evaluation of template with parameters didn't evaluate to a parsable number result='1a''
`,
			}),
			Entry("parameter is not good tag", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.mesh"
                  port: "80"
                  parameters:
                    port: "***"
`,
				expected: `
                violations:
                - field: 'conf.parameters["port"]' 
                  message: value of parameters must be a valid tag name
`,
			}),
			Entry("parameter is not good template entry", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.mesh"
                  port: "80"
                  parameters:
                    _: "service"
`,
				expected: `
                violations:
                - field: 'conf.parameters["_"]' 
                  message: key of parameters must consist of alphanumeric characters
`,
			}),
			Entry("empty parameter is not good template entry", testCase{
				input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.mesh"
                  port: "80"
                  parameters:
                    "": "service"
`,
				expected: `
                violations:
                - field: 'conf.parameters[""]' 
                  message: key of parameters must consist of alphanumeric characters
`,
			}),
		)
	})
})
