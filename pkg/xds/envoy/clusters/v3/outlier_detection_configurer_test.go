package clusters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("OutlierDetectionConfigurer", func() {
	type testCase struct {
		clusterName    string
		circuitBreaker *core_mesh.CircuitBreakerResource
		expected       string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, given.clusterName).
				Configure(clusters.EdsCluster()).
				Configure(clusters.OutlierDetection(given.circuitBreaker)).
				Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("CircuitBreaker with TotalError detector, default values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              enforcingConsecutive5xx: 100
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentage: 0
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with GatewayError detector, default values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							GatewayErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 100
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentage: 0
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with LocalError detector, default values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							LocalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 100
              enforcingFailurePercentage: 0
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with all error detectors, custom values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							TotalErrors:   &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{Consecutive: util_proto.UInt32(21)},
							GatewayErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{Consecutive: util_proto.UInt32(11)},
							LocalErrors:   &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{Consecutive: util_proto.UInt32(6)},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              consecutive5xx: 21
              consecutiveGatewayFailure: 11
              consecutiveLocalOriginFailure: 6
              enforcingConsecutive5xx: 100
              enforcingConsecutiveGatewayFailure: 100
              enforcingConsecutiveLocalOriginFailure: 100
              enforcingFailurePercentage: 0
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with StandardDeviation detector, default values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							StandardDeviation: &mesh_proto.CircuitBreaker_Conf_Detectors_StandardDeviation{},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentage: 0
              enforcingLocalOriginSuccessRate: 100
              enforcingSuccessRate: 100
            type: EDS`,
		}),
		Entry("CircuitBreaker with StandardDeviation detector, custom values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							StandardDeviation: &mesh_proto.CircuitBreaker_Conf_Detectors_StandardDeviation{
								RequestVolume: util_proto.UInt32(7),
								MinimumHosts:  util_proto.UInt32(8),
								Factor:        util_proto.Double(1.9),
							},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentage: 0
              enforcingLocalOriginSuccessRate: 100
              enforcingSuccessRate: 100
              successRateMinimumHosts: 8
              successRateRequestVolume: 7
              successRateStdevFactor: 1900
            type: EDS`,
		}),
		Entry("CircuitBreaker with Failure detector, default values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							Failure: &mesh_proto.CircuitBreaker_Conf_Detectors_Failure{},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentageLocalOrigin: 100
              enforcingFailurePercentage: 100
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with Failure detector, custom values", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							Failure: &mesh_proto.CircuitBreaker_Conf_Detectors_Failure{
								RequestVolume: util_proto.UInt32(7),
								MinimumHosts:  util_proto.UInt32(8),
								Threshold:     util_proto.UInt32(85),
							},
						},
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentageLocalOrigin: 100
              enforcingFailurePercentage: 100
              enforcingSuccessRate: 0
              failurePercentageMinimumHosts: 8
              failurePercentageRequestVolume: 7
              failurePercentageThreshold: 85
            type: EDS`,
		}),
	)
})
