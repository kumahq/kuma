package proto_test

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/durationpb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

	Describe("MergeWithListReplacement", func() {
		It("should replace lists", func() {
			// given
			dst := &envoy_cluster.Cluster{
				Name: "dst",
				HealthChecks: []*envoy_config_core_v3.HealthCheck{
					{EventLogPath: "/path/for/dst/0"},
					{EventLogPath: "/path/for/dst/1"},
					{EventLogPath: "/path/for/dst/2"},
				},
			}
			src := &envoy_cluster.Cluster{
				Name: "src",
				HealthChecks: []*envoy_config_core_v3.HealthCheck{
					{EventLogPath: "/path/for/src/0"},
					{EventLogPath: "/path/for/src/1"},
					{EventLogPath: "/path/for/src/2"},
					{EventLogPath: "/path/for/src/3"},
					{EventLogPath: "/path/for/src/4"},
				},
			}

			// when
			util_proto.MergeWithListReplacement(dst, src)

			// then
			Expect(dst.Name).To(Equal("src"))
			Expect(dst.HealthChecks).To(HaveLen(5))
			Expect(dst.HealthChecks[0].EventLogPath).To(Equal("/path/for/src/0"))
			Expect(dst.HealthChecks[1].EventLogPath).To(Equal("/path/for/src/1"))
			Expect(dst.HealthChecks[2].EventLogPath).To(Equal("/path/for/src/2"))
			Expect(dst.HealthChecks[3].EventLogPath).To(Equal("/path/for/src/3"))
			Expect(dst.HealthChecks[4].EventLogPath).To(Equal("/path/for/src/4"))
		})

		It("should replace with empty list", func() {
			// given
			dst := &envoy_cluster.Cluster{
				Name: "dst",
				HealthChecks: []*envoy_config_core_v3.HealthCheck{
					{EventLogPath: "/path/for/dst/0"},
					{EventLogPath: "/path/for/dst/1"},
					{EventLogPath: "/path/for/dst/2"},
				},
			}
			src := &envoy_cluster.Cluster{
				Name:         "src",
				HealthChecks: []*envoy_config_core_v3.HealthCheck{},
			}

			// when
			util_proto.MergeWithListReplacement(dst, src)

			// then
			Expect(dst.Name).To(Equal("src"))
			Expect(dst.HealthChecks).To(HaveLen(0))
		})
	})
})
