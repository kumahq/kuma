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
				Expect(verr).ToNot(HaveOccurred())
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
  kind: MeshService
  name: web-frontend
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
		)
	})
})
