package v1alpha1_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	meshaccesslog_proto "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
)

var _ = Describe("MeshAccessLog", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtpYAML string) {
				// setup
				meshAccessLog := meshaccesslog_proto.NewMeshAccessLogResource()

				// when
				err := core_model.FromYAML([]byte(mtpYAML), &meshAccessLog.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshAccessLog.Validate()

				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("mesh from/to example", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: Tcp
          tcp:
            format:
              type: Json
              json:
                - key: "start_time"
                  value: "%START_TIME%"
            address: 127.0.0.1:5000
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      backends:
        - type: File
          file:
           format:
             type: Plain
             plain: '{"start_time": "%START_TIME%"}'
           path: '/tmp/logs.txt'
`),
			Entry("empty format", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
`),
			Entry("empty backend list", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends: []
`),
			Entry("MeshMultiZoneService", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshMultiZoneService
      name: web-backend
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
`),
			Entry("openTelemetry with backendRef", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            backendRef:
              kind: MeshOpenTelemetryBackend
              labels:
                kuma.io/display-name: my-otel
`),
			Entry("openTelemetry with backendRef using labels", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            backendRef:
              kind: MeshOpenTelemetryBackend
              labels:
                app: otel-collector
`),
			Entry("openTelemetry with valid attributes", `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            endpoint: otel-collector:4317
            attributes:
              - key: "service.version"
                value: "%KUMA_MESH%"
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshAccessLog := meshaccesslog_proto.NewMeshAccessLogResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &meshAccessLog.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshAccessLog.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty 'from' and 'to' array", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
`,
				expected: `
violations:
  - field: spec
    message: at least one of 'from', 'to' or 'rules' has to be defined`,
			}),
			Entry("empty 'path'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            format:
              type: Plain
              plain: '{"start_time": "%START_TIME%"}'
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].file.path
    message: file backend requires a valid path`,
			}),
			Entry("invalid 'path'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
           format:
             type: Plain
             plain: '{"start_time": "%START_TIME%"}'
           path: '#not_valid'
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].file.path
    message: file backend requires a valid path`,
			}),
			Entry("empty 'key'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
            format:
              type: Json
              json:
                - value: "%START_TIME%"
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].file.format.json[0].key
    message: key cannot be empty`,
			}),
			Entry("empty 'value'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
            format:
              type: Json
              json:
                - key: "start_time"
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].file.format.json[0].value
    message: value cannot be empty`,
			}),
			Entry("invalid 'key'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
            format:
              type: Json
              json:
                - key: '"'
                  value: "%START_TIME%"
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].file.format.json[0]
    message: is not a valid JSON object`,
			}),
			Entry("'default' not defined in to", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: Mesh
`,
				expected: `
violations:
- field: spec.to[0].default.backends
  message: 'must be defined'`,
			}),
			Entry("'default' not defined in from", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
`,
				expected: `
violations:
- field: spec.from[0].default.backends
  message: 'must be defined'`,
			}),
			Entry("sectionName with outbound policies", testCase{
				inputYaml: `
targetRef:
  kind: Dataplane
  sectionName: test
to:
  - targetRef:
      kind: Mesh
`,
				expected: `
violations:
- field: spec.targetRef.sectionName
  message: can only be used with inbound policies
- field: spec.to[0].default.backends
  message: must be defined`,
			}),
			Entry("don't mix from with rules", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
rules:
  - default:
      backends:
        - type: Tcp
          tcp:
            format:
              type: Json
              json:
                - key: "start_time"
                  value: "%START_TIME%"
            address: google.com
`,
				expected: `
violations:
- field: spec
  message: fields 'to' and 'from' must be empty when 'rules' is defined
- field: spec.from[0].default.backends
  message: must be defined`,
			}),
			Entry("'address' not valid", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: Tcp
          tcp:
            format:
              type: Json
              json:
                - key: "start_time"
                  value: "%START_TIME%"
            address: not_valid_url
`,
				expected: `
violations:
- field: spec.from[0].default.backends[0].tcp.address
  message: 'tcp backend requires valid address'`,
			}),
			Entry("empty format json list", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
            format:
              type: Json
              json: []
        - type: Tcp
          tcp:
            address: http://logs.com
            format:
              type: Json
              json: []
`,
				expected: `
violations:
- field: spec.from[0].default.backends[0].file.format.json
  message: 'must not be empty'
- field: spec.from[0].default.backends[1].tcp.format.json
  message: 'must not be empty'`,
			}),
			Entry("empty format.plain", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
            format:
              type: Plain
              plain: ""
        - type: Tcp
          tcp:
            address: http://logs.com
            format:
              type: Plain
              plain: ""
`,
				expected: `
violations:
- field: spec.from[0].default.backends[0].file.format.plain
  message: 'must not be empty'
- field: spec.from[0].default.backends[1].tcp.format.plain
  message: 'must not be empty'`,
			}),
			Entry("backend must be defined", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
        - type: Tcp
        - type: OpenTelemetry
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].file
    message: must be defined
  - field: spec.from[0].default.backends[1].tcp
    message: must be defined
  - field: spec.from[0].default.backends[2].openTelemetry
    message: must be defined`,
			}),
			Entry("format must be defined", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: File
          file:
            path: '/tmp/logs.txt'
            format:
              type: Plain
        - type: Tcp
          tcp:
            address: http://logs.com
            format:
              type: Json
`,
				expected: `
violations:
- field: spec.from[0].default.backends[0].file.format.plain
  message: must be defined
- field: spec.from[0].default.backends[1].tcp.format.json
  message: must be defined`,
			}),
			Entry("openTelemetry neither endpoint nor backendRef", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            endpoint: ""
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].openTelemetry
    message: "openTelemetry must have exactly one defined: endpoint, backendRef"`,
			}),
			Entry("openTelemetry both endpoint and backendRef", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            endpoint: otel-collector:4317
            backendRef:
              kind: MeshOpenTelemetryBackend
              labels:
                kuma.io/display-name: my-otel
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].openTelemetry
    message: "openTelemetry must have only one type defined: endpoint, backendRef"`,
			}),
			Entry("openTelemetry backendRef no labels", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            backendRef:
              kind: MeshOpenTelemetryBackend
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].openTelemetry.backendRef
    message: "backendRef must have exactly one defined: labels"`,
			}),
			Entry("openTelemetry attribute key with spaces", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            endpoint: otel-collector:4317
            attributes:
              - key: "my custom attribute"
                value: "%KUMA_MESH%"
`,
				expected: fmt.Sprintf(`
violations:
  - field: spec.from[0].default.backends[0].openTelemetry.attributes[0].key
    message: %s`, validators.MustMatchOtelAttributeNameFormat),
			}),
			Entry("openTelemetry attribute key with placeholders", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            endpoint: otel-collector:4317
            attributes:
              - key: "%KUMA_ZONE%"
                value: "%KUMA_MESH%"
`,
				expected: fmt.Sprintf(`
violations:
  - field: spec.from[0].default.backends[0].openTelemetry.attributes[0].key
    message: "%s"`, validators.MustBeStaticOtelAttributeName),
			}),
			Entry("openTelemetry attribute key with reserved prefix", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            endpoint: otel-collector:4317
            attributes:
              - key: "otel.attribute"
                value: "%KUMA_MESH%"
`,
				expected: fmt.Sprintf(`
violations:
  - field: spec.from[0].default.backends[0].openTelemetry.attributes[0].key
    message: "%s"`, validators.MustNotUseReservedOtelPrefix),
			}),
			Entry("openTelemetry attribute keys accumulate violations", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - type: OpenTelemetry
          openTelemetry:
            endpoint: otel-collector:4317
            attributes:
              - key: "bad key"
                value: "%KUMA_MESH%"
              - key: "%KUMA_ZONE%"
                value: "%KUMA_ZONE%"
`,
				expected: fmt.Sprintf(`
violations:
  - field: spec.from[0].default.backends[0].openTelemetry.attributes[0].key
    message: %s
  - field: spec.from[0].default.backends[0].openTelemetry.attributes[1].key
    message: "%s"`, validators.MustMatchOtelAttributeNameFormat, validators.MustBeStaticOtelAttributeName),
			}),
			Entry("top-level MeshGateway is rejected", testCase{
				inputYaml: `
targetRef:
  kind: MeshGateway
  name: edge-gateway
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      backends:
        - type: File
          file:
           format:
             type: Plain
             plain: '{"start_time": "%START_TIME%"}'
           path: '/tmp/logs.txt'
`,
				expected: `
violations:
- field: spec.targetRef.kind
  message: value 'MeshGateway' is not supported
- field: spec.to[0].targetRef.kind
  message: value 'MeshService' is not supported`,
			}),
		)
	})
})
