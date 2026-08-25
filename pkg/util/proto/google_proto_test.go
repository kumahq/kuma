package proto_test

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

var _ = Describe("MergeKuma", func() {
	It("should merge durations by replacing them", func() {
		dest := &envoy_cluster.Cluster{
			Name:           "old",
			ConnectTimeout: durationpb.New(time.Second * 10),
			EdsClusterConfig: &envoy_cluster.Cluster_EdsClusterConfig{
				ServiceName: "srv",
				EdsConfig: &envoy_config_core_v3.ConfigSource{
					InitialFetchTimeout: durationpb.New(time.Millisecond * 100),
				},
			},
		}
		src := &envoy_cluster.Cluster{
			Name:           "new",
			ConnectTimeout: durationpb.New(time.Millisecond * 500),
			EdsClusterConfig: &envoy_cluster.Cluster_EdsClusterConfig{
				EdsConfig: &envoy_config_core_v3.ConfigSource{
					InitialFetchTimeout: durationpb.New(time.Second),
					ResourceApiVersion:  envoy_config_core_v3.ApiVersion_V3,
				},
			},
		}
		util_proto.Merge(dest, src)
		Expect(dest.ConnectTimeout.AsDuration()).To(Equal(time.Millisecond * 500))
		Expect(dest.Name).To(Equal("new"))
		Expect(dest.EdsClusterConfig.ServiceName).To(Equal("srv"))
		Expect(dest.EdsClusterConfig.EdsConfig.InitialFetchTimeout.AsDuration()).To(Equal(time.Second))
		Expect(dest.EdsClusterConfig.EdsConfig.InitialFetchTimeout.AsDuration()).To(Equal(time.Second))
		Expect(dest.EdsClusterConfig.EdsConfig.ResourceApiVersion).To(Equal(envoy_config_core_v3.ApiVersion_V3))
	})

	It("should merge circuit breaker thresholds by priority instead of appending", func() {
		dest := &envoy_cluster.Cluster{
			Name: "cluster",
			CircuitBreakers: &envoy_cluster.CircuitBreakers{
				Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
					{
						Priority:       envoy_config_core_v3.RoutingPriority_DEFAULT,
						TrackRemaining: true,
					},
				},
			},
		}
		src := &envoy_cluster.Cluster{
			CircuitBreakers: &envoy_cluster.CircuitBreakers{
				Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
					{
						MaxConnections: util_proto.UInt32(8192),
					},
					{
						Priority:       envoy_config_core_v3.RoutingPriority_HIGH,
						MaxConnections: util_proto.UInt32(2048),
					},
				},
			},
		}
		util_proto.Merge(dest, src)
		Expect(dest.CircuitBreakers.Thresholds).To(HaveLen(2))
		Expect(dest.CircuitBreakers.Thresholds[0].Priority).To(Equal(envoy_config_core_v3.RoutingPriority_DEFAULT))
		Expect(dest.CircuitBreakers.Thresholds[0].TrackRemaining).To(BeTrue())
		Expect(dest.CircuitBreakers.Thresholds[0].MaxConnections.GetValue()).To(Equal(uint32(8192)))
		Expect(dest.CircuitBreakers.Thresholds[1].Priority).To(Equal(envoy_config_core_v3.RoutingPriority_HIGH))
		Expect(dest.CircuitBreakers.Thresholds[1].MaxConnections.GetValue()).To(Equal(uint32(2048)))
	})
})
