package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ = Describe("MeshCircuitBreaker", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(mtYAML string) {
				// setup
				meshCircuitBreaker := NewMeshCircuitBreakerResource()

				// when
				err := core_model.FromYAML([]byte(mtYAML), &meshCircuitBreaker.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshCircuitBreaker.Validate()

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
      connectionLimits:
        maxConnections: 12
        maxConnectionPools: 12
        maxPendingRequests: 92
        maxRetries: 8
        maxRequests: 128
      outlierDetection:
        disabled: false
        interval: 11s
        baseEjectionTime: 38s
        maxEjectionPercent: 22
        splitExternalAndLocalErrors: true
        detectors:
          totalFailures:
            consecutive: 10
          gatewayFailures:
            consecutive: 10
          localOriginFailures:
            consecutive: 10
          successRate:
            minimumHosts: 5
            requestVolume: 10
            standardDeviationFactor: "1.9"
          failurePercentage:
            requestVolume: 10
            minimumHosts: 5
            threshold: 31
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionLimits:
        maxConnections: 2
        maxConnectionPools: 2
        maxPendingRequests: 2
        maxRetries: 1
        maxRequests: 32
      outlierDetection:
        disabled: false
        interval: 5s
        baseEjectionTime: 30s
        maxEjectionPercent: 20
        splitExternalAndLocalErrors: true
        detectors:
          totalFailures:
            consecutive: 10
          gatewayFailures:
            consecutive: 10
          localOriginFailures:
            consecutive: 10
          successRate:
            minimumHosts: 5
            requestVolume: 10
            standardDeviationFactor: 1900
          failurePercentage:
            requestVolume: 10
            minimumHosts: 5
            threshold: 85
  - targetRef:
      kind: Mesh
    default:
      connectionLimits:
        maxConnections: 2
        maxConnectionPools: 2
        maxPendingRequests: 2
        maxRetries: 1
        maxRequests: 32
      outlierDetection:
        disabled: false
        interval: 5s
        baseEjectionTime: 30s
        maxEjectionPercent: 20
        splitExternalAndLocalErrors: true
        detectors:
          totalFailures:
            consecutive: 10
          gatewayFailures:
            consecutive: 10
          localOriginFailures:
            consecutive: 10
          successRate:
            minimumHosts: 5
            requestVolume: 10
            standardDeviationFactor: 1900
          failurePercentage:
            requestVolume: 10
            minimumHosts: 5
            threshold: 85
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionLimits:
        maxConnections: 22
        maxConnectionPools: 22
        maxPendingRequests: 32
        maxRetries: 2
        maxRequests: 64
      outlierDetection:
        disabled: false
        interval: 10s
        baseEjectionTime: 15s
        maxEjectionPercent: 43
        splitExternalAndLocalErrors: true
        detectors:
          totalFailures:
            consecutive: 20
          gatewayFailures:
            consecutive: 30
          localOriginFailures:
            consecutive: 40
          successRate:
            minimumHosts: 3
            requestVolume: 20
            standardDeviationFactor: 1300
          failurePercentage:
            requestVolume: 4
            minimumHosts: 2
            threshold: 75`),
			Entry("only to targetRef", `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: Mesh
    default:
      connectionLimits:
        maxConnections: 2
        maxConnectionPools: 2
        maxPendingRequests: 2
        maxRetries: 1
        maxRequests: 32
      outlierDetection:
        disabled: false
        interval: 5s
        baseEjectionTime: 30s
        maxEjectionPercent: 20
        splitExternalAndLocalErrors: true
        detectors:
          totalFailures:
            consecutive: 10
          gatewayFailures:
            consecutive: 10
          localOriginFailures:
            consecutive: 10
          successRate:
            minimumHosts: 5
            requestVolume: 10
            standardDeviationFactor: 1333
          failurePercentage:
            requestVolume: 10
            minimumHosts: 5
            threshold: 85
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionLimits:
        maxConnections: 22
        maxConnectionPools: 22
        maxPendingRequests: 32
        maxRetries: 2
        maxRequests: 64
      outlierDetection:
        disabled: false
        interval: 10s
        baseEjectionTime: 15s
        maxEjectionPercent: 43
        splitExternalAndLocalErrors: true
        detectors:
          totalFailures:
            consecutive: 20
          gatewayFailures:
            consecutive: 30
          localOriginFailures:
            consecutive: 40
          successRate:
            minimumHosts: 3
            requestVolume: 20
            standardDeviationFactor: 3311
          failurePercentage:
            requestVolume: 4
            minimumHosts: 2
            threshold: 75`),
			Entry("minimal example", `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionLimits: { }`),
			XEntry("with MeshExternalService example", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshExternalService
      name: web-backend
    default:
      connectionLimits: { }`),
			Entry("with MeshMultiZoneService", `
targetRef:
  kind: Mesh
to:
  - targetRef:
      kind: MeshMultiZoneService
      name: web-backend
    default:
      connectionLimits: { }`),
			Entry("gateway example", `
targetRef:
  kind: MeshGateway
  name: edge
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionLimits: { }`),
		)

		type testCase struct {
			inputYaml string
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				meshCircuitBreaker := NewMeshCircuitBreakerResource()

				// when
				err := core_model.FromYAML([]byte(given.inputYaml), &meshCircuitBreaker.Spec)
				Expect(err).ToNot(HaveOccurred())
				// and
				verr := meshCircuitBreaker.Validate()
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
			Entry("unsupported kind in from selector", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
from:
  - targetRef:
      kind: MeshGatewayRoute
    default:
      connectionLimits: { }`,
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
      connectionLimits: { }`,
				expected: `
violations:
  - field: spec.to[0].targetRef.kind
    message: value is not supported`,
			}),
			Entry("missing configuration", testCase{
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
    message: 'at least one of: ''connectionLimits'' or ''outlierDetection'' should be configured'`,
			}),
			Entry("limits cannot be be equal 0 when specified", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      connectionLimits:
        maxConnections: 0
        maxConnectionPools: 0
        maxPendingRequests: 0
        maxRetries: 0
        maxRequests: 0`,
				expected: `
violations:
  - field: spec.to[0].default.connectionLimits.maxConnections
    message: must be greater than 0
  - field: spec.to[0].default.connectionLimits.maxConnectionPools
    message: must be greater than 0
  - field: spec.to[0].default.connectionLimits.maxPendingRequests
    message: must be greater than 0
  - field: spec.to[0].default.connectionLimits.maxRetries
    message: must be greater than 0
  - field: spec.to[0].default.connectionLimits.maxRequests
    message: must be greater than 0`,
			}),
			Entry("any outlierDetection's numeric property cannot be be 0 when specified", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      outlierDetection:
        interval: 0s
        baseEjectionTime: 0s
        maxEjectionPercent: 0
        detectors:
          totalFailures:
            consecutive: 0
          gatewayFailures:
            consecutive: 0
          localOriginFailures:
            consecutive: 0
          successRate:
            minimumHosts: 0
            requestVolume: 0
            standardDeviationFactor: 0
          failurePercentage:
            requestVolume: 0
            minimumHosts: 0
            threshold: 0`,
				expected: `
violations:
  - field: spec.to[0].default.outlierDetection.interval
    message: must be greater than zero when defined
  - field: spec.to[0].default.outlierDetection.baseEjectionTime
    message: must be greater than zero when defined
  - field: spec.to[0].default.outlierDetection.detectors.totalFailures.consecutive
    message: must be greater than 0
  - field: spec.to[0].default.outlierDetection.detectors.gatewayFailures.consecutive
    message: must be greater than 0
  - field: spec.to[0].default.outlierDetection.detectors.localOriginFailures.consecutive
    message: must be greater than 0
  - field: spec.to[0].default.outlierDetection.detectors.successRate.minimumHosts
    message: must be greater than 0
  - field: spec.to[0].default.outlierDetection.detectors.successRate.requestVolume
    message: must be greater than 0
  - field: spec.to[0].default.outlierDetection.detectors.successRate.standardDeviationFactor
    message: must be greater than 0
  - field: spec.to[0].default.outlierDetection.detectors.failurePercentage.minimumHosts
    message: must be greater than 0
  - field: spec.to[0].default.outlierDetection.detectors.failurePercentage.requestVolume
    message: must be greater than 0`,
			}),
			Entry("any outlierDetection's percentage property cannot be be greater than 100 when specified", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      outlierDetection:
        maxEjectionPercent: 101
        detectors:
          failurePercentage:
            threshold: 101`,
				expected: `
violations:
  - field: spec.to[0].default.outlierDetection.maxEjectionPercent
    message: must be in inclusive range [0, 100]
  - field: spec.to[0].default.outlierDetection.detectors.failurePercentage.threshold
    message: must be in inclusive range [0, 100]`,
			}),
			Entry("detectors are not defined", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      outlierDetection:
        maxEjectionPercent: 100`,
				expected: `
violations:
  - field: spec.to[0].default.outlierDetection.detectors
    message: must be defined`,
			}),
			Entry("detector is empty", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      outlierDetection:
        maxEjectionPercent: 100
        detectors: {}`,
				expected: `
violations:
  - field: spec.to[0].default.outlierDetection.detectors
    message: 'must have at least one defined: totalFailures, gatewayFailures, localOriginFailures, successRate, failurePercentage'`,
			}),
			Entry("detector has incorrect values", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshService
      name: web-backend
    default:
      outlierDetection:
        detectors:
          successRate:
            standardDeviationFactor: "xyz"`,
				expected: `
violations:
  - field: spec.to[0].default.outlierDetection.detectors.successRate.standardDeviationFactor
    message: 'invalid number'`,
			}),
			XEntry("status codes out of range in expectedStatuses", testCase{
				inputYaml: `
targetRef:
  kind: MeshService
  name: web-frontend
to:
  - targetRef:
      kind: MeshExternalService
      name: web-backend
    default:
      outlierDetection:
        detectors:
          successRate:
            standardDeviationFactor: "1.9"`,
				expected: `
violations:
  - field: spec.to[0].targetRef.kind
    message: 'kind MeshExternalService is only allowed with targetRef.kind: Mesh as it is configured on the Zone Egress and shared by all clients in the mesh'`,
			}),
		)
	})
})
