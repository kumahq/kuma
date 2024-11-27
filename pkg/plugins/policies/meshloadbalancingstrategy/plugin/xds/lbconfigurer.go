package xds

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
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
		leastRequests := &envoy_cluster.Cluster_LeastRequestLbConfig{}
		if arb := c.LoadBalancer.LeastRequest.ActiveRequestBias; arb != nil {
			decimal, err := common_api.NewDecimalFromIntOrString(pointer.Deref(arb))
			if err != nil {
				return err
			}
			leastRequests.ActiveRequestBias = &corev3.RuntimeDouble{
				DefaultValue: decimal.InexactFloat64(),
			}
		}
		if cc := c.LoadBalancer.LeastRequest.ChoiceCount; cc != nil {
			leastRequests.ChoiceCount = util_proto.UInt32(*cc)
		}
		if leastRequests.ChoiceCount != nil || leastRequests.ActiveRequestBias != nil {
			cluster.LbConfig = &envoy_cluster.Cluster_LeastRequestLbConfig_{
				LeastRequestLbConfig: leastRequests,
			}
		}
	case api.RandomType:
		cluster.LbPolicy = envoy_cluster.Cluster_RANDOM
	case api.RingHashType:
		cluster.LbPolicy = envoy_cluster.Cluster_RING_HASH
		if c.LoadBalancer.RingHash == nil {
			return nil
		}
		var minimumRingSize *wrapperspb.UInt64Value
		if min := c.LoadBalancer.RingHash.MinRingSize; min != nil {
			minimumRingSize = util_proto.UInt64(uint64(*min))
		}
		var maximumRingSize *wrapperspb.UInt64Value
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
