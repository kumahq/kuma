package clusters

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"
)

type LbConfigurer struct {
	Lb *mesh_proto.TrafficRoute_LoadBalancer
}

var _ ClusterConfigurer = &LbConfigurer{}

func (e *LbConfigurer) Configure(c *envoy_api.Cluster) error {
	// default to Round Robin
	if e.Lb.GetLbType() == nil {
		c.LbPolicy = envoy_api.Cluster_ROUND_ROBIN
		return nil
	}

	switch e.Lb.GetLbType().(type) {
	case *mesh_proto.TrafficRoute_LoadBalancer_RoundRobin_:
		c.LbPolicy = envoy_api.Cluster_ROUND_ROBIN

	case *mesh_proto.TrafficRoute_LoadBalancer_LeastRequest_:
		c.LbPolicy = envoy_api.Cluster_LEAST_REQUEST

		lbConfig := e.Lb.GetLeastRequest()
		c.LbConfig = &envoy_api.Cluster_LeastRequestLbConfig_{
			LeastRequestLbConfig: &envoy_api.Cluster_LeastRequestLbConfig{
				ChoiceCount: &wrappers.UInt32Value{
					Value: lbConfig.ChoiceCount,
				},
			},
		}

	case *mesh_proto.TrafficRoute_LoadBalancer_RingHash_:
		c.LbPolicy = envoy_api.Cluster_RING_HASH

		lbConfig := e.Lb.GetRingHash()
		hashfn, ok := envoy_api.Cluster_RingHashLbConfig_HashFunction_value[lbConfig.HashFunction]
		if !ok {
			return errors.New(fmt.Sprintf("Invalid ring hash function %s", lbConfig.HashFunction))
		}

		c.LbConfig = &envoy_api.Cluster_RingHashLbConfig_{
			RingHashLbConfig: &envoy_api.Cluster_RingHashLbConfig{
				HashFunction: envoy_api.Cluster_RingHashLbConfig_HashFunction(hashfn),
				MinimumRingSize: &wrappers.UInt64Value{
					Value: lbConfig.MinRingSize,
				},
				MaximumRingSize: &wrappers.UInt64Value{
					Value: lbConfig.MaxRingSize,
				},
			},
		}

	case *mesh_proto.TrafficRoute_LoadBalancer_Random_:
		c.LbPolicy = envoy_api.Cluster_RANDOM

	case *mesh_proto.TrafficRoute_LoadBalancer_Maglev_:
		c.LbPolicy = envoy_api.Cluster_MAGLEV
	}

	return nil
}
