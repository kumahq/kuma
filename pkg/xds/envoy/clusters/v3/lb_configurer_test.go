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

var _ = Describe("Lb", func() {
	type testCase struct {
		clusterName string
		lb          *mesh_proto.TrafficRoute_LoadBalancer
		expected    string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, given.clusterName).
				Configure(clusters.EdsCluster()).
				Configure(clusters.LB(given.lb)).
				Configure(clusters.Timeout(DefaultTimeout(), core_mesh.ProtocolTCP)).
				Build()

			// then
			Expect(err).ToNot(HaveOccurred())

			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("round robin", testCase{
			clusterName: "backend",
			lb: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_{},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            name: backend
            type: EDS`,
		}),
		Entry("least request", testCase{
			clusterName: "backend",
			lb: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_LeastRequest_{
					LeastRequest: &mesh_proto.TrafficRoute_LoadBalancer_LeastRequest{
						ChoiceCount: 4,
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            lbPolicy: LEAST_REQUEST
            leastRequestLbConfig:
              choiceCount: 4
            name: backend
            type: EDS`,
		}),
		Entry("least request with default", testCase{
			clusterName: "backend",
			lb: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_LeastRequest_{},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            lbPolicy: LEAST_REQUEST
            leastRequestLbConfig:
              choiceCount: 2
            name: backend
            type: EDS`,
		}),
		Entry("ring hash", testCase{
			clusterName: "backend",
			lb: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_RingHash_{
					RingHash: &mesh_proto.TrafficRoute_LoadBalancer_RingHash{
						HashFunction: "MURMUR_HASH_2",
						MinRingSize:  64,
						MaxRingSize:  1024,
					},
				},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            lbPolicy: RING_HASH
            name: backend
            ringHashLbConfig:
              hashFunction: MURMUR_HASH_2
              maximumRingSize: "1024"
              minimumRingSize: "64"
            type: EDS`,
		}),
		Entry("random", testCase{
			clusterName: "backend",
			lb: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_Random_{},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            lbPolicy: RANDOM
            name: backend
            type: EDS`,
		}),
		Entry("random", testCase{
			clusterName: "backend",
			lb: &mesh_proto.TrafficRoute_LoadBalancer{
				LbType: &mesh_proto.TrafficRoute_LoadBalancer_Maglev_{},
			},
			expected: `
            connectTimeout: 5s
            edsClusterConfig:
              edsConfig:
                ads: {}
                resourceApiVersion: V3
            lbPolicy: MAGLEV
            name: backend
            type: EDS`,
		}),
	)
})
