package xds_test

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("LBConfigurer", func() {
	type testCase struct {
		conf     api.LoadBalancer
		expected string
	}

	DescribeTable("should generate proper envoy config",
		func(given testCase) {
			// given
			configurer := &xds.LoadBalancerConfigurer{
				LoadBalancer: given.conf,
			}
			cluster := clusters.NewClusterBuilder(envoy.APIV3, "test").
				MustBuild()

			// when
			err := configurer.Configure(cluster.(*envoy_cluster.Cluster))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("round robin lb", testCase{
			conf: api.LoadBalancer{
				Type: api.RoundRobinType,
			},
			expected: `
name: test
`,
		}),
		Entry("random lb", testCase{
			conf: api.LoadBalancer{
				Type: api.RandomType,
			},
			expected: `
lbPolicy: RANDOM
name: test
`,
		}),
		Entry("least request lb", testCase{
			conf: api.LoadBalancer{
				Type: api.LeastRequestType,
				LeastRequest: &api.LeastRequest{
					ChoiceCount: pointer.To[uint32](12),
				},
			},
			expected: `
lbPolicy: LEAST_REQUEST
leastRequestLbConfig:
  choiceCount: 12
name: test
`,
		}),
		Entry("least request lb, empty conf", testCase{
			conf: api.LoadBalancer{
				Type: api.LeastRequestType,
			},
			expected: `
lbPolicy: LEAST_REQUEST
name: test
`,
		}),
		Entry("ring hash lb", testCase{
			conf: api.LoadBalancer{
				Type: api.RingHashType,
				RingHash: &api.RingHash{
					MinRingSize:  pointer.To[uint32](9),
					MaxRingSize:  pointer.To[uint32](19),
					HashFunction: pointer.To(api.MurmurHash2Type),
				},
			},
			expected: `
lbPolicy: RING_HASH
ringHashLbConfig:
  hashFunction: MURMUR_HASH_2
  maximumRingSize: "19"
  minimumRingSize: "9"
name: test
`,
		}),
		Entry("ring hash lb, empty conf", testCase{
			conf: api.LoadBalancer{
				Type: api.RingHashType,
			},
			expected: `
lbPolicy: RING_HASH
name: test
`,
		}),
		Entry("ring hash lb, only hash func", testCase{
			conf: api.LoadBalancer{
				Type: api.RingHashType,
				RingHash: &api.RingHash{
					HashFunction: pointer.To(api.XXHashType),
				},
			},
			expected: `
lbPolicy: RING_HASH
ringHashLbConfig: {}
name: test
`,
		}),
		Entry("maglev lb", testCase{
			conf: api.LoadBalancer{
				Type: api.MaglevType,
				Maglev: &api.Maglev{
					TableSize: pointer.To[uint32](123),
				},
			},
			expected: `
lbPolicy: MAGLEV
maglevLbConfig: 
  tableSize: "123"
name: test
`,
		}),
	)
})
