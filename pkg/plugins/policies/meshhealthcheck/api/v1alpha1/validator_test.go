package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshhealthcheck_proto "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
)

var _ = Describe("MeshHealthCheck", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mhcYAML string) {
				// setup
				meshHealthCheck := meshhealthcheck_proto.NewMeshHealthCheckResource()

				// when
				err := core_model.FromYAML([]byte(mhcYAML), &meshHealthCheck.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshHealthCheck.Validate()

				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full example", `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
      initialJitter: 5s # optional
      intervalJitter: 6s # optional
      intervalJitterPercent: 10 # optional
      healthyPanicThreshold: 60 # optional, by default 50
      failTrafficOnPanic: true # optional, by default false
      noTrafficInterval: 10s # optional, by default 60s
      eventLogPath: "/tmp/health-check.log" # optional
      alwaysLogHealthCheckFailures: true # optional, by default false
      reuseConnection: false # optional, by default true
      tcp: # it will pick the protocol as described in 'protocol selection' section
        disabled: true # new, default false, can be disabled for override
        send: Zm9v # optional, empty payloads imply a connect-only health check
        receive: # optional
        - YmFy
        - YmF6
      http:
        disabled: true # new, default false, can be disabled for override
        path: /health
        requestHeadersToAdd: # optional, empty by default
        set:
        - name: Content-Type
          value: application/json
        add:
        - name: Accept
          value: application/json
        expectedStatuses: [200, 201] # optional, by default [200]
      grpc: # new
        disabled: false # new, default false, can be disabled for override
        serviceName: "" # optional, service name parameter which will be sent to gRPC service
        authority: "" # optional, the value of the :authority header in the gRPC health check request, by default name of the cluster this health check is associated with
`),
			Entry("top level MeshGateway", `
targetRef:
  kind: MeshGateway
  name: edge
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      interval: 10s
      tcp: # it will pick the protocol as described in 'protocol selection' section
        disabled: true # new, default false, can be disabled for override
`),
			XEntry("to level MeshExternalService", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshExternalService
      name: mes
    default:
      interval: 10s
      tcp: # it will pick the protocol as described in 'protocol selection' section
        disabled: true # new, default false, can be disabled for override
`),
			Entry("to level MeshMultiZoneService", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshMultiZoneService
      name: web-backend
    default:
      interval: 10s
      tcp: # it will pick the protocol as described in 'protocol selection' section
        disabled: true # new, default false, can be disabled for override
`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshHealthCheck := meshhealthcheck_proto.NewMeshHealthCheckResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &meshHealthCheck.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshHealthCheck.Validate()
				actual, err := yaml.Marshal(verr)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("wrong top level target ref", testCase{
				inputYaml: `
targetRef:
  kind: MeshGatewayRoute
  name: some-mesh-gateway-route
`,
				expected: `
violations:
  - field: spec.targetRef.kind
    message: value is not supported`, // this could be more specific
			}),
			PEntry("to field is an empty array", testCase{ // this does not work, needs
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to: []
`,
				expected: `
violations:
  - field: spec.to
    message: must not be empty`,
			}),
			Entry("required fields are missing", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default: {}
`,
				expected: `
violations:
  - field: spec.to[0].default
    message: 'must have at least one defined: http, tcp, grpc'`,
			}),
			Entry("positive values are out of range", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: -3
      healthyThreshold: 0
      grpc: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.unhealthyThreshold
    message: must be greater than zero when defined
  - field: spec.to[0].default.healthyThreshold
    message: must be greater than zero when defined`,
			}),
			Entry("positive durations are out of range", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      interval: -10s
      timeout: -2s
      initialJitter: -5s
      intervalJitter: -6s
      noTrafficInterval: -10s
      unhealthyThreshold: 3
      healthyThreshold: 1
      grpc: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.interval
    message: must be greater than zero when defined
  - field: spec.to[0].default.timeout
    message: must be greater than zero when defined
  - field: spec.to[0].default.initialJitter
    message: must be greater than zero when defined
  - field: spec.to[0].default.intervalJitter
    message: must be greater than zero when defined
  - field: spec.to[0].default.noTrafficInterval
    message: must be greater than zero when defined`,
			}),
			Entry("all percentages are out of percentage range", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
      intervalJitterPercent: 110
      healthyPanicThreshold: -10
      grpc: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.intervalJitterPercent
    message: must be in inclusive range [0, 100]
  - field: spec.to[0].default.healthyPanicThreshold
    message: must be in inclusive range [0.0, 100.0]`,
			}),
			Entry("path is invalid", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
      eventLogPath: "#not_valid_path"
      grpc: {}
`,
				expected: `
violations:
  - field: spec.to[0].default.eventLogPath
    message: must be a valid path when defined`,
			}),
			Entry("status codes out of range in expectedStatuses", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
      http:
        path: /health
        expectedStatuses: [99, 600]
`,
				expected: `
violations:
  - field: spec.to[0].default.http.expectedStatuses[0]
    message: must be in inclusive range [100, 599]
  - field: spec.to[0].default.http.expectedStatuses[1]
    message: must be in inclusive range [100, 599]`,
			}),
			XEntry("cannot use MeshExternalService with other type than Mesh", testCase{
				inputYaml: `
targetRef:
  kind: MeshSubset
  tags:
    kuma.io/service: backend
to:
  - targetRef:
      kind: MeshExternalService
      name: web-backend
    default:
      interval: 10s
      timeout: 2s
      unhealthyThreshold: 3
      healthyThreshold: 1
      http:
        path: /health
        expectedStatuses: [200, 204]
`,
				expected: `
violations:
  - field: spec.to[0].targetRef.kind
    message: 'kind MeshExternalService is only allowed with targetRef.kind: Mesh as it is configured on the Zone Egress and shared by all clients in the mesh'`,
			}),
		)
	})
})
