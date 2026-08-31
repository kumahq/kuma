package proto_test

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"

	util_proto "github.com/kumahq/kuma/v2/pkg/util/proto"
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
		util_proto.Merge(dest, src, util_proto.ReplaceDurationOptionFn)
		Expect(dest.ConnectTimeout.AsDuration()).To(Equal(time.Millisecond * 500))
		Expect(dest.Name).To(Equal("new"))
		Expect(dest.EdsClusterConfig.ServiceName).To(Equal("srv"))
		Expect(dest.EdsClusterConfig.EdsConfig.InitialFetchTimeout.AsDuration()).To(Equal(time.Second))
		Expect(dest.EdsClusterConfig.EdsConfig.InitialFetchTimeout.AsDuration()).To(Equal(time.Second))
		Expect(dest.EdsClusterConfig.EdsConfig.ResourceApiVersion).To(Equal(envoy_config_core_v3.ApiVersion_V3))
	})

	It("should merge durations field by field without the option", func() {
		dest := &envoy_cluster.Cluster{
			ConnectTimeout: durationpb.New(time.Second * 10),
		}
		src := &envoy_cluster.Cluster{
			ConnectTimeout: &durationpb.Duration{Nanos: 5},
		}
		util_proto.Merge(dest, src)
		Expect(dest.ConnectTimeout.AsDuration()).To(Equal(time.Second*10 + 5))
	})

	Context("keyed lists", func() {
		thresholdsKeyedByPriority := util_proto.KeyedListOptionFn(
			"envoy.config.cluster.v3.CircuitBreakers.thresholds",
			"priority",
		)

		It("should merge into the dst entry with the same key", func() {
			dest := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_DEFAULT, TrackRemaining: true},
						{Priority: envoy_config_core_v3.RoutingPriority_HIGH, MaxRequests: util_proto.UInt32(5)},
					},
				},
			}
			src := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_HIGH, MaxConnections: util_proto.UInt32(77)},
					},
				},
			}

			util_proto.Merge(dest, src, thresholdsKeyedByPriority)

			thresholds := dest.CircuitBreakers.Thresholds
			Expect(thresholds).To(HaveLen(2))
			Expect(thresholds[0].Priority).To(Equal(envoy_config_core_v3.RoutingPriority_DEFAULT))
			Expect(thresholds[0].TrackRemaining).To(BeTrue())
			Expect(thresholds[1].Priority).To(Equal(envoy_config_core_v3.RoutingPriority_HIGH))
			Expect(thresholds[1].MaxConnections.GetValue()).To(Equal(uint32(77)))
			Expect(thresholds[1].MaxRequests.GetValue()).To(Equal(uint32(5)))
		})

		It("should append a src entry whose key is not in dst", func() {
			dest := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_DEFAULT, TrackRemaining: true},
					},
				},
			}
			src := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_HIGH, MaxConnections: util_proto.UInt32(2048)},
					},
				},
			}

			util_proto.Merge(dest, src, thresholdsKeyedByPriority)

			thresholds := dest.CircuitBreakers.Thresholds
			Expect(thresholds).To(HaveLen(2))
			Expect(thresholds[0].TrackRemaining).To(BeTrue())
			Expect(thresholds[1].Priority).To(Equal(envoy_config_core_v3.RoutingPriority_HIGH))
			Expect(thresholds[1].MaxConnections.GetValue()).To(Equal(uint32(2048)))
		})

		It("should skip nil dst entries instead of panicking", func() {
			dest := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						nil,
						{Priority: envoy_config_core_v3.RoutingPriority_DEFAULT, TrackRemaining: true},
					},
				},
			}
			src := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{MaxConnections: util_proto.UInt32(8192)},
					},
				},
			}

			Expect(func() { util_proto.Merge(dest, src, thresholdsKeyedByPriority) }).ToNot(Panic())

			thresholds := dest.CircuitBreakers.Thresholds
			Expect(thresholds).To(HaveLen(2))
			Expect(thresholds[0]).To(BeNil())
			Expect(thresholds[1].TrackRemaining).To(BeTrue())
			Expect(thresholds[1].MaxConnections.GetValue()).To(Equal(uint32(8192)))
		})

		It("should leave lists that were not registered as keyed appending", func() {
			dest := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					PerHostThresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_DEFAULT, TrackRemaining: true},
					},
				},
			}
			src := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					PerHostThresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{MaxConnections: util_proto.UInt32(512)},
					},
				},
			}

			util_proto.Merge(dest, src, thresholdsKeyedByPriority)

			Expect(dest.CircuitBreakers.PerHostThresholds).To(HaveLen(2))
		})

		It("should drop src entries that repeat a key", func() {
			dest := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_DEFAULT, TrackRemaining: true},
					},
				},
			}
			src := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{MaxConnections: util_proto.UInt32(1)},
						{MaxConnections: util_proto.UInt32(2)},
					},
				},
			}

			util_proto.Merge(dest, src, thresholdsKeyedByPriority)

			thresholds := dest.CircuitBreakers.Thresholds
			Expect(thresholds).To(HaveLen(1))
			Expect(thresholds[0].TrackRemaining).To(BeTrue())
			Expect(thresholds[0].MaxConnections.GetValue()).To(Equal(uint32(1)))
		})

		It("should fall back to appending when the key field is not a comparable scalar", func() {
			dest := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_DEFAULT, TrackRemaining: true},
					},
				},
			}
			src := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{MaxConnections: util_proto.UInt32(8192)},
					},
				},
			}

			util_proto.Merge(dest, src, util_proto.KeyedListOptionFn(
				"envoy.config.cluster.v3.CircuitBreakers.thresholds",
				"max_connections",
			))

			Expect(dest.CircuitBreakers.Thresholds).To(HaveLen(2))
		})

		It("should fall back to appending when the key field does not exist", func() {
			dest := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{Priority: envoy_config_core_v3.RoutingPriority_DEFAULT, TrackRemaining: true},
					},
				},
			}
			src := &envoy_cluster.Cluster{
				CircuitBreakers: &envoy_cluster.CircuitBreakers{
					Thresholds: []*envoy_cluster.CircuitBreakers_Thresholds{
						{MaxConnections: util_proto.UInt32(8192)},
					},
				},
			}

			util_proto.Merge(dest, src, util_proto.KeyedListOptionFn(
				"envoy.config.cluster.v3.CircuitBreakers.thresholds",
				"no_such_field",
			))

			Expect(dest.CircuitBreakers.Thresholds).To(HaveLen(2))
		})
	})
})
