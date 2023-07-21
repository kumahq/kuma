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

var _ = Describe("CircuitBreakerConfigurer", func() {
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
				Configure(clusters.CircuitBreaker(given.circuitBreaker)).
				Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("CircuitBreaker with thresholds", testCase{
			clusterName: "backend",
			circuitBreaker: &core_mesh.CircuitBreakerResource{
				Spec: &mesh_proto.CircuitBreaker{
					Conf: &mesh_proto.CircuitBreaker_Conf{
						Thresholds: &mesh_proto.CircuitBreaker_Conf_Thresholds{
							MaxConnections:     util_proto.UInt32(2),
							MaxPendingRequests: util_proto.UInt32(3),
							MaxRequests:        util_proto.UInt32(4),
							MaxRetries:         util_proto.UInt32(5),
						},
					},
				},
			},
			expected: `
        circuitBreakers:
          thresholds:
          - maxConnections: 2
            maxPendingRequests: 3
            maxRequests: 4
            maxRetries: 5
        connectTimeout: 5s
        edsClusterConfig:
          edsConfig:
            ads: {}
            resourceApiVersion: V3
        name: backend
        type: EDS`,
		}),
	)
})
