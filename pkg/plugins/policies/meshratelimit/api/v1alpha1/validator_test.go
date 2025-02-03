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
				Expect(verr).ToNot(HaveOccurred())
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
          requestRate:
            num: 100
            interval: 10s
          onRateLimit:
            status: 123
            headers:
              add:
              - name: "test"
                value: "123"
        tcp:
          disabled: false
          connectionRate:
            num: 100
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
          requestRate:
            num: 100
            interval: 10s
          onRateLimit:
            status: 123
            headers:
              set:
              - name: "test"
                value: "123"`),
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
          connectionRate:
            num: 100
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
          requestRate:
            num: 100
            interval: 10s
        tcp:
          connectionRate:
            num: 100
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
			Entry("gateway example", `
targetRef:
  kind: MeshGateway
  name: edge
to:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          requestRate:
            num: 100
            interval: 10s
        tcp:
          connectionRate:
            num: 100
            interval: 100ms`),
			Entry("gateway example and targeting MeshHTTPRoute", `
targetRef:
  kind: MeshHTTPRoute
  name: http-route-1
to:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          requestRate:
            num: 100
            interval: 10s
        tcp:
          connectionRate:
            num: 100
            interval: 100ms`),
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
			Entry("unsupported kind in from selector", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: MeshGatewayRoute
    default:
      local:
        http:
          requestRate:
            num: 10
            interval: 10s`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`,
			}),
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
        requestRate:
          num: 0
          interval: 49ms
      tcp:
        connectionRate:
          num: 0
          interval: 49ms`,
				expected: `
violations:
  - field: spec.from[0].default.local.http.requestRate.num
    message: must be greater than 0
  - field: spec.from[0].default.local.http.requestRate.interval
    message: 'must be greater than: 50ms'
  - field: spec.from[0].default.local.tcp.connectionRate.num
    message: must be greater than 0
  - field: spec.from[0].default.local.tcp.connectionRate.interval
    message: 'must be greater than: 50ms'`,
			}),
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
        connectionRate:
          num: 100
          interval: 500ms`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`,
			}),
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
        requestRate:
          num: 100
          interval: 500ms
      tcp:
        connectionRate:
          num: 100
          interval: 500ms`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`,
			}),
			Entry("not allow from to be MeshService", testCase{
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
        requestRate:
          num: 100
          interval: 500ms`,
				expected: `
violations:
  - field: spec.from[0].targetRef.kind
    message: value is not supported`,
			}),
			Entry("empty default", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: Mesh
  default: {}`,
				expected: `
violations:
  - field: spec.from[0].default.local
    message: must be defined`,
			}),
			Entry("sectionName with outbound policy", testCase{
				inputYaml: `
targetRef:
  kind: Dataplane
  sectionName: test
to:
- targetRef:
    kind: Mesh
  default: {}`,
				expected: `
violations:
  - field: spec.targetRef.sectionName
    message: can only be used with inbound policies
  - field: spec.to
    message: must not be defined`,
			}),
			Entry("neither tcp or http defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: Mesh
  default: 
    local: {}`,
				expected: `
violations:
  - field: spec.from[0].default.local
    message: 'must have at least one defined: tcp, http'`,
			}),
			Entry("empty tcp", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: Mesh
  default: 
    local: 
      tcp: {}`,
				expected: `
violations:
  - field: spec.from[0].default.local.tcp
    message: 'must have at least one defined: disabled, connectionRate'`,
			}),
			Entry("empty http", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
- targetRef:
    kind: Mesh
  default: 
    local: 
      http: {}`,
				expected: `
violations:
  - field: spec.from[0].default.local.http
    message: 'must have at least one defined: disabled, requestRate, onRateLimit'`,
			}),
			Entry("invalid gateway example", testCase{
				inputYaml: `
targetRef:
  kind: MeshGateway
  name: edge
from:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          requestRate:
            num: 100
            interval: 10s
        tcp:
          connectionRate:
            num: 100
            interval: 100ms`,
				expected: `
violations:
  - field: spec.from
    message: 'must not be defined'`,
			}),
			Entry("invalid gateway example when targeting MeshHTTPRoute", testCase{
				inputYaml: `
targetRef:
  kind: MeshHTTPRoute
  name: http-route-1
from:
  - targetRef:
      kind: Mesh
    default:
      local:
        http:
          requestRate:
            num: 100
            interval: 10s
        tcp:
          connectionRate:
            num: 100
            interval: 100ms`,
				expected: `
violations:
  - field: spec.from
    message: 'must not be defined'`,
			}),
		)
	})
})
