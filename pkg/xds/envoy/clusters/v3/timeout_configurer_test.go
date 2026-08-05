package clusters_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/defaults/mesh"
	policies_defaults "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/defaults"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
)

var _ = Describe("TimeoutConfigurer", func() {
	userTimeout := envoy.Timeouts{
		Connect:        100 * time.Second,
		TcpIdle:        101 * time.Second,
		HttpIdle:       103 * time.Second,
		HttpStreamIdle: 104 * time.Second,
	}

	timeoutConf := envoy.Timeouts{
		Connect:        policies_defaults.DefaultConnectTimeout,
		TcpIdle:        policies_defaults.DefaultIdleTimeout,
		HttpIdle:       policies_defaults.DefaultIdleTimeout,
		HttpStreamIdle: policies_defaults.DefaultStreamIdleTimeout,
	}

	type testCase struct {
		timeout  envoy.Timeouts
		expected string
	}

	DescribeTable("should set timeouts for outbound HTTP cluster",
		func(given testCase) {
			// given
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, "backend").
				Configure(clusters.EdsCluster()).
				Configure(clusters.Timeout(given.timeout, core_meta.ProtocolHTTP)).
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
      idleTimeout: 103s`,
		}),
		Entry("default timeout", testCase{
			timeout: timeoutConf,
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
	)

	DescribeTable("should set timeouts for outbound GRPC cluster",
		func(given testCase) {
			// given
			cluster, err := clusters.NewClusterBuilder(envoy.APIV3, "backend").
				Configure(clusters.EdsCluster()).
				Configure(clusters.Timeout(given.timeout, core_meta.ProtocolGRPC)).
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
      idleTimeout: 103s`,
		}),
		Entry("default timeout", testCase{
			timeout: timeoutConf,
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
	)

	It("should set timeouts for inbound HTTP cluster", func() {
		// given
		cluster, err := clusters.NewClusterBuilder(envoy.APIV3, "localhost:8080").
			Configure(clusters.ProvidedEndpointCluster(false, core_xds.Endpoint{Target: "192.168.0.1", Port: 8080})).
			Configure(clusters.Timeout(mesh.DefaultInboundTimeout(), core_meta.ProtocolHTTP)).
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
