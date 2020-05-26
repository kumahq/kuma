package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("CircuitBreaker", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(circuitBreakerYAML string) {
				// setup
				circuitBreaker := core_mesh.CircuitBreakerResource{}

				// when
				err := util_proto.FromYAML([]byte(circuitBreakerYAML), &circuitBreaker.Spec)
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
                    service: frontend
                    region: us
                destinations:
                - match:
                    service: backend
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
                      threshold: 85`),
			Entry("one detector with default values", `
                sources:
                - match:
                    service: frontend
                    region: us
                destinations:
                - match:
                    service: backend
                conf:
                    detectors:
                      totalErrors: {}
                      standardDeviation: {}`),
		)

		type testCase struct {
			circuitBreaker string
			expected       string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				circuitBreaker := core_mesh.CircuitBreakerResource{}

				// when
				err := util_proto.FromYAML([]byte(given.circuitBreaker), &circuitBreaker.Spec)
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
                 message: must have at least one of the detectors configured`}),
			Entry("wrong format", testCase{
				circuitBreaker: `
                sources:
                - match:
                    service: frontend
                    region: us
                destinations:
                - match:
                    service: backend
                    region: eu
                conf:
                    maxEjectionPercent: 120
                    detectors:
                      failure:
                        threshold: 850`,
				expected: `
               violations:
               - field: destinations[0].match
                 message: must consist of exactly one tag "service"
               - field: destinations[0].match["region"]
                 message: tag "region" is not allowed
               - field: conf.maxEjectionPercent
                 message: has to be in [0.0 - 100.0] range
               - field: conf.detectors.failure.threshold
                 message: has to be in [0.0 - 100.0] range`}),
		)
	})
})
