package cluster

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func NewCluster() *Builder[envoy_cluster.Cluster] {
	return &Builder[envoy_cluster.Cluster]{}
}

func Name(name string) Configurer[envoy_cluster.Cluster] {
	return func(s *envoy_cluster.Cluster) error {
		s.Name = name
		return nil
	}
}

func ConnectTimeout(timeout time.Duration) Configurer[envoy_cluster.Cluster] {
	return func(s *envoy_cluster.Cluster) error {
		s.ConnectTimeout = durationpb.New(timeout)
		return nil
	}
}

func Endpoints(clusterName string, endpoints []*envoy_endpoint.LocalityLbEndpoints) Configurer[envoy_cluster.Cluster] {
	return func(s *envoy_cluster.Cluster) error {
		if s.LoadAssignment == nil {
			s.LoadAssignment = &envoy_endpoint.ClusterLoadAssignment{}
		}
		s.LoadAssignment.ClusterName = clusterName
		s.LoadAssignment.Endpoints = endpoints
		return nil
	}
}

func Http2() Configurer[envoy_cluster.Cluster] {
	return func(s *envoy_cluster.Cluster) error {
		if s.TypedExtensionProtocolOptions == nil {
			s.TypedExtensionProtocolOptions = map[string]*anypb.Any{}
		}
		options := &envoy_upstream_http.HttpProtocolOptions{}
		if any := s.TypedExtensionProtocolOptions["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"]; any != nil {
			if err := util_proto.UnmarshalAnyTo(any, options); err != nil {
				return err
			}
			options.UpstreamProtocolOptions = &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
						Http2ProtocolOptions: &envoy_core.Http2ProtocolOptions{},
					},
				},
			}
		} else {
			options = &envoy_upstream_http.HttpProtocolOptions{
				UpstreamProtocolOptions: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_{
					ExplicitHttpConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig{
						ProtocolConfig: &envoy_upstream_http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
							Http2ProtocolOptions: &envoy_core.Http2ProtocolOptions{},
						},
					},
				},
			}
		}

		pbst, err := util_proto.MarshalAnyDeterministic(options)
		if err != nil {
			return err
		}
		s.TypedExtensionProtocolOptions["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"] = pbst
		return nil
	}
}

func LocalityWeightedLbConfigurer() Configurer[envoy_cluster.Cluster] {
	return func(cluster *envoy_cluster.Cluster) error {
		if cluster.CommonLbConfig == nil {
			cluster.CommonLbConfig = &envoy_cluster.Cluster_CommonLbConfig{}
		}
		cluster.CommonLbConfig.LocalityConfigSpecifier = &envoy_cluster.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{}
		return nil
	}
}

func LbPolicy(p envoy_cluster.Cluster_LbPolicy) Configurer[envoy_cluster.Cluster] {
	return func(cluster *envoy_cluster.Cluster) error {
		cluster.LbPolicy = p
		return nil
	}
}

func LeastRequestLbConfig(configBuilder *Builder[envoy_cluster.Cluster_LeastRequestLbConfig]) Configurer[envoy_cluster.Cluster] {
	return func(c *envoy_cluster.Cluster) error {
		config, err := configBuilder.Build()
		if err != nil {
			return err
		}
		c.LbConfig = &envoy_cluster.Cluster_LeastRequestLbConfig_{
			LeastRequestLbConfig: config,
		}
		return nil
	}
}

func NewLeastRequestConfig() *Builder[envoy_cluster.Cluster_LeastRequestLbConfig] {
	return &Builder[envoy_cluster.Cluster_LeastRequestLbConfig]{}
}

func ChoiceCount(cc uint32) Configurer[envoy_cluster.Cluster_LeastRequestLbConfig] {
	return func(c *envoy_cluster.Cluster_LeastRequestLbConfig) error {
		c.ChoiceCount = util_proto.UInt32(cc)
		return nil
	}
}

func ActiveRequestBias(arb intstr.IntOrString) Configurer[envoy_cluster.Cluster_LeastRequestLbConfig] {
	return func(c *envoy_cluster.Cluster_LeastRequestLbConfig) error {
		decimal, err := common_api.NewDecimalFromIntOrString(arb)
		if err != nil {
			return err
		}
		c.ActiveRequestBias = &envoy_core.RuntimeDouble{
			DefaultValue: decimal.InexactFloat64(),
		}
		return nil
	}
}

func RingHashLbConfig(configBuilder *Builder[envoy_cluster.Cluster_RingHashLbConfig]) Configurer[envoy_cluster.Cluster] {
	return func(c *envoy_cluster.Cluster) error {
		config, err := configBuilder.Build()
		if err != nil {
			return err
		}
		c.LbConfig = &envoy_cluster.Cluster_RingHashLbConfig_{
			RingHashLbConfig: config,
		}
		return nil
	}
}

func NewRingHashConfig() *Builder[envoy_cluster.Cluster_RingHashLbConfig] {
	return &Builder[envoy_cluster.Cluster_RingHashLbConfig]{}
}

func MinRingSize(min uint32) Configurer[envoy_cluster.Cluster_RingHashLbConfig] {
	return func(c *envoy_cluster.Cluster_RingHashLbConfig) error {
		c.MinimumRingSize = util_proto.UInt64(uint64(min))
		return nil
	}
}

func MaxRingSize(max uint32) Configurer[envoy_cluster.Cluster_RingHashLbConfig] {
	return func(c *envoy_cluster.Cluster_RingHashLbConfig) error {
		c.MaximumRingSize = util_proto.UInt64(uint64(max))
		return nil
	}
}

func HashFunction(hf envoy_cluster.Cluster_RingHashLbConfig_HashFunction) Configurer[envoy_cluster.Cluster_RingHashLbConfig] {
	return func(c *envoy_cluster.Cluster_RingHashLbConfig) error {
		c.HashFunction = hf
		return nil
	}
}

func MaglevLbConfig(configBuilder *Builder[envoy_cluster.Cluster_MaglevLbConfig]) Configurer[envoy_cluster.Cluster] {
	return func(c *envoy_cluster.Cluster) error {
		config, err := configBuilder.Build()
		if err != nil {
			return err
		}
		c.LbConfig = &envoy_cluster.Cluster_MaglevLbConfig_{
			MaglevLbConfig: config,
		}
		return nil
	}
}

func NewMaglevConfig() *Builder[envoy_cluster.Cluster_MaglevLbConfig] {
	return &Builder[envoy_cluster.Cluster_MaglevLbConfig]{}
}

func TableSize(tableSize uint32) Configurer[envoy_cluster.Cluster_MaglevLbConfig] {
	return func(c *envoy_cluster.Cluster_MaglevLbConfig) error {
		c.TableSize = util_proto.UInt64(uint64(tableSize))
		return nil
	}
}
