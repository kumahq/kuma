package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("OutlierDetectionConfigurer", func() {

	type testCase struct {
		clusterName    string
		circuitBreaker *mesh_core.CircuitBreakerResource
		expected       string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.OutlierDetection(given.circuitBreaker)).
				Configure(clusters.Timeout(mesh_core.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: durationpb.New(5 * time.Second)})).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("CircuitBreaker with TotalError detector, default values", testCase{
			circuitBreaker: &mesh_core.CircuitBreakerResource{
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
            outlierDetection:
              enforcingConsecutive5xx: 100
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentage: 0
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with GatewayError detector, default values", testCase{
			circuitBreaker: &mesh_core.CircuitBreakerResource{
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
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 100
              enforcingConsecutiveLocalOriginFailure: 0
              enforcingFailurePercentage: 0
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with LocalError detector, default values", testCase{
			circuitBreaker: &mesh_core.CircuitBreakerResource{
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
            outlierDetection:
              enforcingConsecutive5xx: 0
              enforcingConsecutiveGatewayFailure: 0
              enforcingConsecutiveLocalOriginFailure: 100
              enforcingFailurePercentage: 0
              enforcingSuccessRate: 0
            type: EDS`,
		}),
		Entry("CircuitBreaker with all error detectors, custom values", testCase{
			circuitBreaker: &mesh_core.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							TotalErrors:   &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{Consecutive: &wrapperspb.UInt32Value{Value: 21}},
							GatewayErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{Consecutive: &wrapperspb.UInt32Value{Value: 11}},
							LocalErrors:   &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{Consecutive: &wrapperspb.UInt32Value{Value: 6}},
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
			circuitBreaker: &mesh_core.CircuitBreakerResource{
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
			circuitBreaker: &mesh_core.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							StandardDeviation: &mesh_proto.CircuitBreaker_Conf_Detectors_StandardDeviation{
								RequestVolume: &wrapperspb.UInt32Value{Value: 7},
								MinimumHosts:  &wrapperspb.UInt32Value{Value: 8},
								Factor:        &wrapperspb.DoubleValue{Value: 1.9},
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
			circuitBreaker: &mesh_core.CircuitBreakerResource{
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
			circuitBreaker: &mesh_core.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
							Failure: &mesh_proto.CircuitBreaker_Conf_Detectors_Failure{
								RequestVolume: &wrapperspb.UInt32Value{Value: 7},
								MinimumHosts:  &wrapperspb.UInt32Value{Value: 8},
								Threshold:     &wrapperspb.UInt32Value{Value: 85},
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
