package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshaccesslog_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
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
				Expect(verr).To(BeNil())
			},
			Entry("mesh from/to example", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - tcp:
            format:
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
        - file:
           format:
             plain: '{"start_time": "%START_TIME%"}'
           path: '/tmp/logs.txt'
`),
			Entry("empty format", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
            path: '/tmp/logs.txt'
`),
			Entry("empty backend list", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends: []
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
    message: at least one of 'from', 'to' has to be defined`,
			}),
			Entry("empty 'path'", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
           format:
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
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
           format:
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
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
           path: '/tmp/logs.txt'
           format:
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
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
           path: '/tmp/logs.txt'
           format:
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
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
           path: '/tmp/logs.txt'
           format:
             json:
                - key: '"'
                  value: "%START_TIME%"
`,
				expected: `
violations:
  - field: spec.from[0].default.backends[0].file.format.json[0]
    message: is not a valid JSON object`,
			}),
			Entry("both 'plain' and 'json' defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - tcp:
            address: 127.0.0.1:5000
            format:
              plain: '{"start_time": "%START_TIME%"}'
              json:
                - key: "start_time"
                  value: "%START_TIME%"
`,
				expected: `
violations:
- field: spec.from[0].default.backends[0].tcp.format
  message: 'format must have only one type defined: plain, json'`,
			}),
			Entry("both 'tcp' and 'file' defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - tcp:
            address: 127.0.0.1:5000
            format:
              json:
                - key: "start_time"
                  value: "%START_TIME%"
          file:
           format:
             plain: '{"start_time": "%START_TIME%"}'
           path: '/tmp/logs.txt'
`,
				expected: `
violations:
- field: spec.from[0].default.backends[0]
  message: 'backend must have only one type defined: tcp, file, openTelemetry'`,
			}),

			Entry("'to' defined in MeshGatewayRoute", testCase{
				inputYaml: `
targetRef:
  kind: MeshGatewayRoute
  name: some-mesh-gateway-route
to:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
           format:
             plain: '{"start_time": "%START_TIME%"}'
           path: '/tmp/logs.txt'
`,
				expected: `
violations:
- field: spec.to
  message: 'cannot use "to" when "targetRef" is "MeshGatewayRoute" - there is no outbound'`,
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
			Entry("'address' not valid", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - tcp:
            format:
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
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
            path: '/tmp/logs.txt'
            format:
              json: []
        - tcp:
            address: http://logs.com
            format:
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
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      backends:
        - file:
            path: '/tmp/logs.txt'
            format:
              plain: ""
        - tcp:
            address: http://logs.com
            format:
              plain: ""
`,
				expected: `
violations:
- field: spec.from[0].default.backends[0].file.format.plain
  message: 'must not be empty'
- field: spec.from[0].default.backends[1].tcp.format.plain
  message: 'must not be empty'`,
			}),
		)
	})
})
