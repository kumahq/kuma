package v1alpha1_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("MeshTrace", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				meshTrace := meshtrace_proto.NewMeshTraceResource()

				// when
				err := util_proto.FromYAML([]byte(mtpYAML), meshTrace.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTrace.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("full example", `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - zipkin:
        url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
        apiVersion: httpJson
  tags:
    - name: team
      literal: core
    - name: env
      header:
        name: x-env
        default: prod
    - name: version
      header:
        name: x-version
  sampling:
    overall: 80
    random: 60
    client: 40
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshTrace := meshtrace_proto.NewMeshTraceResource()

				// when
				err := util_proto.FromYAML([]byte(given.inputYaml), meshTrace.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTrace.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("no default", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
`,
				expected: `
violations:
  - field: spec.default
    message: must be defined`,
			}),
			Entry("default empty", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default: {}
`,
				expected: `
violations:
  - field: spec.default.backends
    message: must have exactly one backend defined`,
			}),
			Entry("no backends", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends: []
`,
				expected: `
violations:
  - field: spec.default.backends
    message: must have exactly one backend defined`,
			}),
			Entry("no valid backends", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - unknown: {}
`,
				expected: `
violations:
  - field: spec.default.backends[0]
    message: 'backend must have only one type defined: datadog, zipkin'`,
			}),
			Entry("multiple backends", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - zipkin:
        url: http://jaeger-collector1.mesh-observability:9411/api/v2/spans
    - zipkin:
        url: http://jaeger-collector2.mesh-observability:9411/api/v2/spans
`,
				expected: `
violations:
  - field: spec.default.backends
    message: 'must have exactly one backend defined'`,
			}),
			Entry("no url for zipkin backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - zipkin: {}
`,
				expected: `
violations:
  - field: spec.default.backends[0].zipkin.url
    message: must not be empty`,
			}),
			Entry("invalid url for zipkin backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - zipkin:
        url: not_valid_url
`,
				expected: `
violations:
  - field: spec.default.backends[0].zipkin.url
    message: must be a valid url`,
			}),
			Entry("invalid apiVersion for zipkin backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - zipkin:
        url: http://jaeger-collector.mesh-observability:9411/api/v2/spans
        apiVersion: invalid_api_version
`,
				expected: `
violations:
  - field: spec.default.backends[0].zipkin.apiVersion
    message: must be one of httpJson, httpJsonV1, httpProto`,
			}),
			Entry("missing address for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - datadog:
        port: 443
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.address
    message: must not be empty`,
			}),
			Entry("invalid address for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - datadog:
        address: not_a_valid_address
        port: 443
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.address
    message: must be a valid address`,
			}),
			Entry("missing port for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - datadog:
        address: intake.logs.datadoghq.eu
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.port
    message: must be a valid port (0-65535)`,
			}),
			Entry("invalid port for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - datadog:
        address: intake.logs.datadoghq.eu
        port: 999999
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.port
    message: must be a valid port (0-65535)`,
			}),
			Entry("tag missing name", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - datadog:
        address: intake.logs.datadoghq.eu
        port: 443
  tags:
    - literal: example
    - header:
        default: example
`,
				expected: `
violations:
  - field: spec.default.tags[0].name
    message: must not be empty
  - field: spec.default.tags[1].name
    message: must not be empty`,
			}),
			Entry("tag missing type", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - datadog:
        address: intake.logs.datadoghq.eu
        port: 443
  tags:
    - name: example
`,
				expected: `
violations:
  - field: spec.default.tags[0]
    message: 'tag must have only one type defined: header, literal'
`,
			}),
			Entry("sampling out of range", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - datadog:
        address: intake.logs.datadoghq.eu
        port: 443
  sampling:
    overall: 101
`,
				expected: `
violations:
  - field: spec.default.sampling.overall
    message: must be between 0 and 100
`,
			}),
		)
	})
})
