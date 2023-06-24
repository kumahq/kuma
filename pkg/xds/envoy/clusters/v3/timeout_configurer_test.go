package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/defaults/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

var _ = Describe("TimeoutConfigurer", func() {
	userTimeout := &mesh_proto.Timeout_Conf{
		ConnectTimeout: util_proto.Duration(100 * time.Second),
		Tcp: &mesh_proto.Timeout_Conf_Tcp{
			IdleTimeout: util_proto.Duration(101 * time.Second),
		},
		Http: &mesh_proto.Timeout_Conf_Http{
			RequestTimeout:    util_proto.Duration(102 * time.Second),
			IdleTimeout:       util_proto.Duration(103 * time.Second),
			StreamIdleTimeout: util_proto.Duration(104 * time.Second),
			MaxStreamDuration: util_proto.Duration(105 * time.Second),
		},
	}

	userTimeoutOldFormat := &mesh_proto.Timeout_Conf{
		ConnectTimeout: util_proto.Duration(100 * time.Second),
		Tcp: &mesh_proto.Timeout_Conf_Tcp{
			IdleTimeout: util_proto.Duration(101 * time.Second),
		},
		Http: &mesh_proto.Timeout_Conf_Http{
			RequestTimeout: util_proto.Duration(102 * time.Second),
			IdleTimeout:    util_proto.Duration(103 * time.Second),
		},
		Grpc: &mesh_proto.Timeout_Conf_Grpc{
			StreamIdleTimeout: util_proto.Duration(104 * time.Second),
			MaxStreamDuration: util_proto.Duration(105 * time.Second),
		},
	}

	type testCase struct {
		timeout  *mesh_proto.Timeout_Conf
		expected string
	}

	DescribeTable("should set timeouts for outbound HTTP cluster",
		func(given testCase) {
			// given
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, "backend").
				Configure(clusters.EdsCluster()).
				Configure(clusters.Timeout(given.timeout, core_mesh.ProtocolHTTP)).
				Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("user's timeout", testCase{
			timeout: userTimeout,
			expected: `
connectTimeout: 100s
edsClusterConfig:
  edsConfig:
    ads: {}
    resourceApiVersion: V3
name: backend
type: EDS
typedExtensionProtocolOptions:
  envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
    commonHttpProtocolOptions:
      idleTimeout: 103s
      maxStreamDuration: 105s`,
		}),
		Entry("default timeout", testCase{
			timeout: mesh.DefaultTimeoutResource().(*core_mesh.TimeoutResource).Spec.GetConf(),
			expected: `
connectTimeout: 5s
edsClusterConfig:
  edsConfig:
    ads: {}
    resourceApiVersion: V3
name: backend
type: EDS
typedExtensionProtocolOptions:
  envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
    commonHttpProtocolOptions:
      idleTimeout: 3600s`,
		}),
		Entry("user's timeout old format (with grpc)", testCase{
			timeout: userTimeoutOldFormat,
			expected: `
connectTimeout: 100s
edsClusterConfig:
  edsConfig:
    ads: {}
    resourceApiVersion: V3
name: backend
type: EDS
typedExtensionProtocolOptions:
  envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
    commonHttpProtocolOptions:
      idleTimeout: 103s`,
		}),
	)

	DescribeTable("should set timeouts for outbound GRPC cluster",
		func(given testCase) {
			// given
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, "backend").
				Configure(clusters.EdsCluster()).
				Configure(clusters.Timeout(given.timeout, core_mesh.ProtocolGRPC)).
				Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(cluster)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("user's timeout", testCase{
			timeout: userTimeout,
			expected: `
connectTimeout: 100s
edsClusterConfig:
  edsConfig:
    ads: {}
    resourceApiVersion: V3
name: backend
type: EDS
typedExtensionProtocolOptions:
  envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
    commonHttpProtocolOptions:
      idleTimeout: 103s
      maxStreamDuration: 105s`,
		}),
		Entry("default timeout", testCase{
			timeout: mesh.DefaultTimeoutResource().(*core_mesh.TimeoutResource).Spec.GetConf(),
			expected: `
connectTimeout: 5s
edsClusterConfig:
  edsConfig:
    ads: {}
    resourceApiVersion: V3
name: backend
type: EDS
typedExtensionProtocolOptions:
  envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
    commonHttpProtocolOptions:
      idleTimeout: 3600s`,
		}),
		Entry("user's timeout old format (with grpc)", testCase{
			timeout: userTimeoutOldFormat,
			expected: `
connectTimeout: 100s
edsClusterConfig:
  edsConfig:
    ads: {}
    resourceApiVersion: V3
name: backend
type: EDS
typedExtensionProtocolOptions:
  envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
    commonHttpProtocolOptions:
      idleTimeout: 103s
      maxStreamDuration: 105s`,
		}),
	)

	It("should set timeouts for inbound HTTP cluster", func() {
		// given
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3, "localhost:8080").
			Configure(clusters.ProvidedEndpointCluster(false, core_xds.Endpoint{Target: "192.168.0.1", Port: 8080})).
			Configure(clusters.Timeout(mesh.DefaultInboundTimeout(), core_mesh.ProtocolHTTP)).
			Build()
		Expect(err).ToNot(HaveOccurred())

		// when
		actual, err := util_proto.ToYAML(cluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		expected := `
altStatName: localhost_8080
connectTimeout: 10s
loadAssignment:
  clusterName: localhost:8080
  endpoints:
  - lbEndpoints:
    - endpoint:
        address:
          socketAddress:
            address: 192.168.0.1
            portValue: 8080
name: localhost:8080
type: STATIC
typedExtensionProtocolOptions:
  envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
    '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
    commonHttpProtocolOptions:
      idleTimeout: 7200s`
		Expect(actual).To(MatchYAML(expected))
	})
})
