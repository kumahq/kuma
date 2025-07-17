package cluster

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

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
		c.ActiveRequestBias = &corev3.RuntimeDouble{
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

func MinRingSize(minimum uint32) Configurer[envoy_cluster.Cluster_RingHashLbConfig] {
	return func(c *envoy_cluster.Cluster_RingHashLbConfig) error {
		c.MinimumRingSize = util_proto.UInt64(uint64(minimum))
		return nil
	}
}

func MaxRingSize(maximum uint32) Configurer[envoy_cluster.Cluster_RingHashLbConfig] {
	return func(c *envoy_cluster.Cluster_RingHashLbConfig) error {
		c.MaximumRingSize = util_proto.UInt64(uint64(maximum))
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
