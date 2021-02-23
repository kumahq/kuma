package clusters_test

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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
			cluster, err := clusters.NewClusterBuilder(envoy.APIV2).
				Configure(clusters.EdsCluster(given.clusterName)).
				Configure(clusters.LB(given.lb)).
				Configure(clusters.Timeout(mesh.ProtocolTCP, &mesh_proto.Timeout_Conf{ConnectTimeout: durationpb.New(5 * time.Second)})).
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
            lbPolicy: LEAST_REQUEST
            leastRequestLbConfig:
              choiceCount: 4
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
            lbPolicy: MAGLEV
            name: backend
            type: EDS`,
		}),
	)
})
