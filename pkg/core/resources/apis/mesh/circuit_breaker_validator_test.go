package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("CircuitBreaker", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(circuitBreakerYAML string) {
				// setup
				circuitBreaker := core_mesh.NewCircuitBreakerResource()

				// when
				err := util_proto.FromYAML([]byte(circuitBreakerYAML), circuitBreaker.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := circuitBreaker.Validate()
				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full example", `
                sources:
                - match:
                    kuma.io/service: frontend
                    region: us
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  interval: 1s
                  baseEjectionTime: 30s
                  maxEjectionPercent: 20
                  splitExternalAndLocalErrors: false 
                  detectors:
                    totalErrors: 
                      consecutive: 20
                    gatewayErrors: 
                      consecutive: 10
                    localErrors: 
                      consecutive: 7
                    standardDeviation:
                      requestVolume: 10
                      minimumHosts: 5
                      factor: 1.9
                    failure:
                      requestVolume: 10
                      minimumHosts: 5
                      threshold: 85
                  thresholds:
                    maxConnections: 2
                    maxPendingRequests: 2
                    maxRetries: 2
                    maxRequests: 2
`),
			Entry("one detector with default values", `
                sources:
                - match:
                    kuma.io/service: frontend
                    region: us
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    detectors:
                      totalErrors: {}
                      standardDeviation: {}`),
			Entry("only thresholds", `
                sources:
                - match:
                    kuma.io/service: frontend
                    region: us
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    thresholds:
                      maxConnections: 2`),
		)

		type testCase struct {
			circuitBreaker string
			expected       string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				circuitBreaker := core_mesh.NewCircuitBreakerResource()

				// when
				err := util_proto.FromYAML([]byte(given.circuitBreaker), circuitBreaker.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := circuitBreaker.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("spec: empty", testCase{
				circuitBreaker: ``,
				expected: `
               violations:
               - field: sources
                 message: must have at least one element
               - field: destinations
                 message: must have at least one element
               - field: conf
                 message: must have at least one of the detector or threshold configured`}),
			Entry("wrong format", testCase{
				circuitBreaker: `
                sources:
                - match:
                    kuma.io/service: frontend
                    region: us
                destinations:
                - match:
                    kuma.io/service: backend
                    region: eu
                conf:
                    maxEjectionPercent: 120
                    detectors:
                      failure:
                        threshold: 850`,
				expected: `
               violations:
               - field: destinations[0].match
                 message: must consist of exactly one tag "kuma.io/service"
               - field: destinations[0].match["region"]
                 message: tag "region" is not allowed
               - field: conf.maxEjectionPercent
                 message: has to be in [0.0 - 100.0] range
               - field: conf.detectors.failure.threshold
                 message: has to be in [0.0 - 100.0] range`}),
			Entry("empty thresholds section", testCase{
				circuitBreaker: `
                sources:
                - match:
                    kuma.io/service: frontend
                    region: us
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  detectors:
                    totalErrors: {}
                    standardDeviation: {}
                  thresholds: {}`,
				expected: `
               violations:
               - field: conf.thresholds
                 message: can't be empty`}),
			Entry("empty detectors section", testCase{
				circuitBreaker: `
                sources:
                - match:
                    kuma.io/service: frontend
                    region: us
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  detectors: {}
                  thresholds:
                    maxConnections: 2`,
				expected: `
               violations:
               - field: conf.detectors
                 message: can't be empty`}),
			Entry("empty detectors and thresholds section", testCase{
				circuitBreaker: `
                sources:
                - match:
                    kuma.io/service: frontend
                    region: us
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  detectors: {}
                  thresholds: {}`,
				expected: `
               violations:
               - field: conf
                 message: must have at least one of the detector or threshold configured`}),
		)
	})
})
