package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ = Describe("MeshRateLimit", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtYAML string) {
				// setup
				meshRateLimit := NewMeshRateLimitResource()

				// when
				err := core_model.FromYAML([]byte(mtYAML), &meshRateLimit.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshRateLimit.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("full example", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          disabled: false
          requests: 100
          interval: 10s
          onRateLimit:
            status: 123
            headers:
            - key: "test"
              value: "123"
              append: true
        tcp:
          disabled: false
          connections: 100
          interval: 100ms`),
			Entry("full example, only http", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          requests: 100
          interval: 10s
          onRateLimit:
            status: 123
            headers:
            - key: "test"
              value: "123"
              append: true`),
			Entry("full example, only tcp", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      local:
        tcp:
          disabled: false
          connections: 100
          interval: 100ms`),
			Entry("minimal example", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          requests: 100
          interval: 10s
        tcp:
          connections: 100
          interval: 100ms`),
			Entry("disable rate limit", `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          disabled: true
        tcp:
          disabled: true`),
		)
		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshRateLimit := NewMeshRateLimitResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &meshRateLimit.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshRateLimit.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty 'from'", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
`,
				expected: `
violations:
  - field: spec.from
    message: needs at least one item`}),
			Entry("unsupported kind in from selector", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: MeshGatewayRoute
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`}),
			Entry("requests and interval needs to be defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: Mesh
  default:
    local:
      http:
        disabled: false`,
				expected: `
violations:
  - field: spec.from[0].default.local.http.requests
    message: must be greater than 0
  - field: spec.from[0].default.local.http.interval
    message: 'must be greater than: 50ms'`}),
			Entry("connections and interval needs to be defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: Mesh
  default:
    local:
      tcp:
        disabled: false`,
				expected: `
violations:
  - field: spec.from[0].default.local.tcp.connections
    message: must be greater than 0
  - field: spec.from[0].default.local.tcp.interval
    message: 'must be greater than: 50ms'`}),
			Entry("not allow invalid values", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: Mesh
  default:
    local:
      http:
        interval: 49ms
      tcp:
        interval: 49ms`,
				expected: `
violations:
  - field: spec.from[0].default.local.http.requests
    message: must be greater than 0
  - field: spec.from[0].default.local.http.interval
    message: 'must be greater than: 50ms'
  - field: spec.from[0].default.local.tcp.connections
    message: must be greater than 0
  - field: spec.from[0].default.local.tcp.interval
    message: 'must be greater than: 50ms'`}),
			Entry("not allow from to be MeshService for tcp", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: MeshService
    name: backend
  default:
    local:
      tcp:
        connections: 100
        interval: 500ms`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`}),
			Entry("not allow from to be MeshService when http and tcp set", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: MeshService
    name: backend
  default:
    local:
      http:
        requests: 100
        interval: 500ms
      tcp:
        connections: 100
        interval: 500ms`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`}),
			Entry("not allow targetRef to be MeshGatewayRoute when http and tcp set", testCase{
				inputYaml: `
targetRef:
  kind: MeshGatewayRoute
  name: web-frontend
from:
- targetRef:
    kind: MeshService
    name: backend
  default:
    local:
      http:
        requests: 100
        interval: 500ms
      tcp:
        connections: 100
        interval: 500ms`,
				expected: `
violations:
  - field: spec.targetRef.kind
    message: value is not supported
  - field: spec.from[0].targetRef.kind
    message: value is not supported`}),
			Entry("not allow from to be MeshService", testCase{
				inputYaml: `
targetRef:
  kind: MeshGatewayRoute
  name: web-frontend
from:
- targetRef:
    kind: MeshService
    name: backend
  default:
    local:
      http:
        requests: 100
        interval: 500ms`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`}),
		)
	})
})
