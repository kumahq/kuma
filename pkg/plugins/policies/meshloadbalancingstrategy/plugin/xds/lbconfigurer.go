package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/golang/protobuf/ptypes/wrappers"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type LoadBalancerConfigurer struct {
	LoadBalancer api.LoadBalancer
}

func (c *LoadBalancerConfigurer) Configure(cluster *envoy_cluster.Cluster) error {
	switch c.LoadBalancer.Type {
	case api.RoundRobinType:
		cluster.LbPolicy = envoy_cluster.Cluster_ROUND_ROBIN
	case api.LeastRequestType:
		cluster.LbPolicy = envoy_cluster.Cluster_LEAST_REQUEST
		if c.LoadBalancer.LeastRequest == nil {
			return nil
		}
		if cc := c.LoadBalancer.LeastRequest.ChoiceCount; cc != nil {
			cluster.LbConfig = &envoy_cluster.Cluster_LeastRequestLbConfig_{
				LeastRequestLbConfig: &envoy_cluster.Cluster_LeastRequestLbConfig{
					ChoiceCount: util_proto.UInt32(*cc),
				},
			}
		}
	case api.RandomType:
		cluster.LbPolicy = envoy_cluster.Cluster_RANDOM
	case api.RingHashType:
		cluster.LbPolicy = envoy_cluster.Cluster_RING_HASH
		if c.LoadBalancer.RingHash == nil {
			return nil
		}
		var minimumRingSize *wrappers.UInt64Value
		if min := c.LoadBalancer.RingHash.MinRingSize; min != nil {
			minimumRingSize = util_proto.UInt64(uint64(*min))
		}
		var maximumRingSize *wrappers.UInt64Value
		if max := c.LoadBalancer.RingHash.MaxRingSize; max != nil {
			maximumRingSize = util_proto.UInt64(uint64(*max))
		}
		var hashFunction envoy_cluster.Cluster_RingHashLbConfig_HashFunction
		if hf := c.LoadBalancer.RingHash.HashFunction; hf != nil {
			switch *hf {
			case api.MurmurHash2Type:
				hashFunction = envoy_cluster.Cluster_RingHashLbConfig_MURMUR_HASH_2
			case api.XXHashType:
				hashFunction = envoy_cluster.Cluster_RingHashLbConfig_XX_HASH
			}
		}
		cluster.LbConfig = &envoy_cluster.Cluster_RingHashLbConfig_{
			RingHashLbConfig: &envoy_cluster.Cluster_RingHashLbConfig{
				MinimumRingSize: minimumRingSize,
				MaximumRingSize: maximumRingSize,
				HashFunction:    hashFunction,
			},
		}
	case api.MaglevType:
		cluster.LbPolicy = envoy_cluster.Cluster_MAGLEV
		if c.LoadBalancer.Maglev == nil {
			return nil
		}
		if tableSize := c.LoadBalancer.Maglev.TableSize; tableSize != nil {
			cluster.LbConfig = &envoy_cluster.Cluster_MaglevLbConfig_{
				MaglevLbConfig: &envoy_cluster.Cluster_MaglevLbConfig{
					TableSize: util_proto.UInt64(uint64(*tableSize)),
				},
			}
		}
	}
	return nil
}
