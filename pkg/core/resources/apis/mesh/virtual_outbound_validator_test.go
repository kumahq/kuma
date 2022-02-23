package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("VirtualOutbound_validator", func() {
	DescribeTable("should pass validation",
		func(in string) {
			// setup
			virtualOutbound := NewVirtualOutboundResource()

			// when
			err := util_proto.FromYAML([]byte(in), virtualOutbound.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := virtualOutbound.Validate()

			// then
			Expect(verr).ToNot(HaveOccurred())
		},
		Entry("full example", `
                selectors:
                - match:
                    region: eu
                    kuma.io/service: "*"
                conf:
                  host: "foo.bar"
                  port: "80"
                  parameters:
                    - name: "service"
                      tagKey: "kuma.io/service"`,
		),
		Entry("empty host", `
                selectors:
                - match:
                    region: eu
                    kuma.io/service: "*"
                conf:
                  host: "{{.service}}.mesh"
                  port: "{{.port}}"
                  parameters:
                    - name: "service"
                      tagKey: kuma.io/service
                    - name: "port"
                      tagKey: kuma.io/port
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
                  parameters:
                  - name: "service"
                    tagKey: "kuma.io/service"
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
                  parameters:
                  - name: "service"
                    tagKey: "kuma.io/service"
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
                  parameters:
                  - name: "service"
                    tagKey: "kuma.io/service"
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
                  parameters:
                  - name: "service"
                    tagKey: "kuma.io/service"
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
                  parameters:
                  - name: "service"
                    tagKey: "kuma.io/service"
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
                  - name: port
                    tagKey: "service"
                  - name: "service"
                    tagKey: "kuma.io/service"
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
                    - name: port
                      tagKey: "***"
                    - name: service
                      tagKey: "kuma.io/service"
`,
			expected: `
                violations:
                - field: 'conf.parameters[0].tagKey'
                  message: must be a valid tag name
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
                  - name: _
                    tagKey: "kuma.io/service"
`,
			expected: `
                violations:
                - field: 'conf.parameters[0].name'
                  message: must consist of alphanumeric characters to be used as a gotemplate key
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
                  - name: srv
                    tagKey: "kuma.io/service"
                  - name: "kuma.io/port"
`,
			expected: `
                violations:
                - field: 'conf.parameters[1].name'
                  message: must consist of alphanumeric characters to be used as a gotemplate key
`,
		}),
		Entry("duplicate name not allowed", testCase{
			input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.mesh"
                  port: "80"
                  parameters:
                  - name: srv
                    tagKey: "kuma.io/service"
                  - name: srv
`,
			expected: `
                violations:
                - field: 'conf.parameters[1].name'
                  message: name is already used
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
                    - name: ""
                      tagKey: "kuma.io/service"
`,
			expected: `
                violations:
                - field: 'conf.parameters[0].name'
                  message: must consist of alphanumeric characters to be used as a gotemplate key
`,
		}),
		Entry("missing service tag in parameters", testCase{
			input: `
                selectors:
                - match:
                    kuma.io/service: "*"
                conf:
                  host: "foo.mesh"
                  port: "80"
                  parameters:
                    - name: "foo"
                      tagKey: "kuma.io/port"
`,
			expected: `
                violations:
                - field: conf.parameters
                  message: must contain a parameter with kuma.io/service as a tagKey
`,
		}),
	)
})
