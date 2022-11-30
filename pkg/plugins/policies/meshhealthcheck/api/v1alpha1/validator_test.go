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
				Expect(verr).To(BeNil())
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
        - append: false
          header:
            key: Content-Type
            value: application/json
        - header:
            key: Accept
            value: application/json
        expectedStatuses: [200, 201] # optional, by default [200]
      grpc: # new
        disabled: false # new, default false, can be disabled for override
        serviceName: "" # optional, service name parameter which will be sent to gRPC service
        authority: "" # optional, the value of the :authority header in the gRPC health check request, by default name of the cluster this health check is associated with
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
    message: value is not supported`,
			}),
			Entry("to empty", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: backend
to: []
`,
				expected: `
violations:
  - field: spec.default
    message: must be defined`,
			}),
			Entry("interval, timeout, unhealthyThreshold and healthyThreshold are missing", testCase{
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
  - field: spec.default
    message: must be defined`,
			}),
		)
	})
})
