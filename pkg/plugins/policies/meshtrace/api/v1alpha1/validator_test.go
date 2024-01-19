package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshtrace_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
)

var _ = Describe("MeshTrace", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				meshTrace := meshtrace_proto.NewMeshTraceResource()

				// when
				err := core_model.FromYAML([]byte(mtpYAML), &meshTrace.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTrace.Validate()

				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full zipkin example", `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Zipkin
      zipkin:
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
    random: "60.1"
    client: 40
`),
			Entry("with opentelemetry backend", `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: OpenTelemetry
      openTelemetry:
        endpoint: otel-collector:4317
`),

			Entry("with empty backends", `
targetRef:
  kind: MeshService
  name: backend
default:
  backends: []
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
			Entry("with datadog backend", `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:8126
        splitService: true
`),
			Entry("top level MeshGateway", `
targetRef:
  kind: MeshGateway
  name: edge
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:8126
        splitService: true
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
				err := core_model.FromYAML([]byte(given.inputYaml), &meshTrace.Spec)
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
  - field: spec.default.backends
    message: must be defined`,
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
    message: 'must have zero or one backend defined'`,
			}),
			Entry("no url for zipkin backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Zipkin
      zipkin: {}
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
    - type: Zipkin
      zipkin:
        url: not_valid_url
`,
				expected: `
violations:
  - field: spec.default.backends[0].zipkin.url
    message: must be a valid url`,
			}),
			Entry("invalid url for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: not_valid_url
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.url
    message: must be a valid url`,
			}),
			Entry("no port for datadog backend url", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.url
    message: port must be defined`,
			}),
			Entry("invalid port for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:999999
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.url
    message: port must be a valid (1-65535)`,
			}),
			Entry("invalid scheme for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: sql://intake.datadoghq.eu:8126
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.url
    message: scheme must be http`,
			}),
			Entry("path provided for datadog backend", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:8126/some/path
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog.url
    message: path must not be defined`,
			}),
			Entry("tag missing name", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:443
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
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:443
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
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:443
  sampling:
    overall: 101
`,
				expected: `
violations:
  - field: spec.default.sampling.overall
    message: must be between 0 and 100
`,
			}),
			Entry("sampling invalid string", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:443
  sampling:
    overall: xyz
    client: xyz
    random: xyz
`,
				expected: `
violations:
  - field: spec.default.sampling.client
    message: string is not a number
  - field: spec.default.sampling.random
    message: string is not a number
  - field: spec.default.sampling.overall
    message: string is not a number
`,
			}),
			Entry("datadog backend must be defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Datadog
`,
				expected: `
violations:
  - field: spec.default.backends[0].datadog
    message: must be defined`,
			}),
			Entry("zipkin backend must be defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: Zipkin
`,
				expected: `
violations:
  - field: spec.default.backends[0].zipkin
    message: must be defined`,
			}),
			Entry("openTelemetry backend must be defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
default:
  backends:
    - type: OpenTelemetry
`,
				expected: `
violations:
  - field: spec.default.backends[0].openTelemetry
    message: must be defined`,
			}),
			Entry("gateway listener tags not allowed", testCase{
				inputYaml: `
targetRef:
  kind: MeshGateway
  name: edge
  tags:
    name: listener-1
default:
  backends:
    - type: Datadog
      datadog:
        url: http://intake.datadoghq.eu:8126
        splitService: true`,
				expected: `
violations:
  - field: spec.targetRef.tags
    message: must not be set with kind MeshGateway`,
			}),
		)
	})
})
