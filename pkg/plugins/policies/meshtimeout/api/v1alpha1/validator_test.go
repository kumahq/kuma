package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ = Describe("MeshTimeout", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtYAML string) {
				// setup
				meshTimeout := NewMeshTimeoutResource()

				// when
				err := core_model.FromYAML([]byte(mtYAML), &meshTimeout.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTimeout.Validate()

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
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 0s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 1s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h`),
			Entry("only to targetRef", `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 1s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h`),
			Entry("minimal example", `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http:
        requestTimeout: 1s`),
			Entry("top-level TargetRefKind is MeshHTTPRoute", `
targetRef:
  kind: MeshHTTPRoute
  name: route-1
to:
  - targetRef:
      kind: Mesh
    default:
      http:
        requestTimeout: 1s
        streamIdleTimeout: 2s
`),
			Entry("example MeshExternalService", `
targetRef:
  kind: MeshSubset
  tags:
    kuma.io/service: web-frontend
to:
  - targetRef:
      kind: MeshExternalService
      name: web-backend
    default:
      http:
        requestTimeout: 1s
  - targetRef:
      kind: MeshExternalService
      labels:
        kuma.io/display-name: web-backend
    default:
      http:
        requestTimeout: 1s`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshTimeout := NewMeshTimeoutResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &meshTimeout.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshTimeout.Validate()
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
    message: value is not supported`,
			}),
			Entry("unsupported kind in to selector", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshServiceSubset
    default:
      http:
        requestTimeout: 1s`,
				expected: `
violations:
  - field: spec.to[0].targetRef.kind
    message: value is not supported`,
			}),
			Entry("missing timeout configuration", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend`,
				expected: `
violations:
  - field: spec.to[0].default
    message: at least one timeout should be configured`,
			}),
			Entry("timeout cannot be negative", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      idleTimeout: -1s`,
				expected: `
violations:
  - field: spec.to[0].default.idleTimeout
    message: must not be negative when defined`,
			}),
			Entry("multiple timeout cannot be negative", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: -1s
      http:
        requestTimeout: -10s`,
				expected: `
violations:
  - field: spec.to[0].default.connectionTimeout
    message: must be greater than zero when defined
  - field: spec.to[0].default.http.requestTimeout
    message: must not be negative when defined`,
			}),
			Entry("multiple timeout cannot be negative", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      http: {}`,
				expected: `
violations:
  - field: spec.to[0].default.http
    message: at least one timeout in this section should be configured`,
			}),
			Entry("top-level targetRef is referencing MeshHTTPRoute", testCase{
				inputYaml: `
targetRef:
  kind: MeshHTTPRoute
  name: route-1
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 1s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h
from:
  - targetRef:
      kind: Mesh
    default:
      connectionTimeout: 11s`,
				expected: `
violations:
  - field: spec.from
    message: must not be defined
  - field: spec.to[0].targetRef.kind
    message: value is not supported
  - field: spec.to[0].default.connectionTimeout
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute
  - field: spec.to[0].default.idleTimeout
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute
  - field: spec.to[0].default.http.maxStreamDuration
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute
  - field: spec.to[0].default.http.maxConnectionDuration
    message: can't be specified when top-level TargetRef is referencing MeshHTTPRoute`,
			}),
			Entry("top-level targetRef is referencing MeshGateway", testCase{
				inputYaml: `
targetRef:
  kind: MeshGateway
  name: gateway-1
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h
      http:
        requestTimeout: 1s
        streamIdleTimeout: 1h
        maxStreamDuration: 1h
        maxConnectionDuration: 1h
from:
  - targetRef:
      kind: Mesh
    default:
      connectionTimeout: 11s`,
				expected: `
violations:
  - field: spec.from
    message: must not be defined
  - field: spec.to[0].targetRef.kind
    message: value is not supported`,
			}),
			Entry("to TargetRef using labels and name for MeshExternalService", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshExternalService
      name: web-backend
      labels:
        kuma.io/display-name: web-backend
    default:
      connectionTimeout: 10s
      idleTimeout: 1h`,
				expected: `
violations:
  - field: spec.to[0].targetRef.labels
    message: either labels or name must be specified`,
			}),
			Entry("when rules is defined, to cannot be defined", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - default:
      connectionTimeout: 10s
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionTimeout: 10s`,
				expected: `
violations:
  - field: spec
    message: fields 'to' and 'from' must be empty when 'rules' is defined`,
			}),
			Entry("when rules is defined, from cannot be defined", testCase{
				inputYaml: `
targetRef:
  kind: Mesh
rules:
  - default:
      connectionTimeout: 10s
from:
  - targetRef:
      kind: Mesh
    default:
      connectionTimeout: 10s`,
				expected: `
violations:
  - field: spec
    message: fields 'to' and 'from' must be empty when 'rules' is defined`,
			}),
		)
	})
})
