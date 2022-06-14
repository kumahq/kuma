package v3_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/defaults/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
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

	DescribeTable("should set timeouts for outbound TCP listener",
		func(given testCase) {
			// given
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener("outbound:192.168.0.1:8080", "192.168.0.1", 8080, xds.SocketAddressProtocolTCP)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(TcpProxy("localhost:8080", envoy_common.NewCluster(envoy_common.WithName("backend")))).
					Configure(Timeout(given.timeout, core_mesh.ProtocolTCP)))).
				Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("user's timeout", testCase{
			timeout: userTimeout,
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.tcp_proxy
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
      cluster: backend
      idleTimeout: 101s
      statPrefix: localhost_8080
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND
`,
		}),
		Entry("default timeout", testCase{
			timeout: mesh.DefaultTimeoutResource().(*core_mesh.TimeoutResource).Spec.GetConf(),
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.tcp_proxy
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
      cluster: backend
      idleTimeout: 3600s
      statPrefix: localhost_8080
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
	)

	DescribeTable("should set timeouts for outbound HTTP listener",
		func(given testCase) {
			// given
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener("outbound:192.168.0.1:8080", "192.168.0.1", 8080, xds.SocketAddressProtocolTCP)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager("localhost:8080", false)).
					Configure(Timeout(given.timeout, core_mesh.ProtocolHTTP)))).
				Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("user's timeout", testCase{
			timeout: userTimeout,
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      commonHttpProtocolOptions:
        idleTimeout: 103s
      httpFilters:
      - name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      statPrefix: localhost_8080
      streamIdleTimeout: 104s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
		Entry("user's timeout old format (with grpc)", testCase{
			timeout: userTimeoutOldFormat,
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      commonHttpProtocolOptions:
        idleTimeout: 103s
      httpFilters:
      - name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      statPrefix: localhost_8080
      streamIdleTimeout: 0s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
		Entry("default timeout", testCase{
			timeout: mesh.DefaultTimeoutResource().(*core_mesh.TimeoutResource).Spec.GetConf(),
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      commonHttpProtocolOptions:
        idleTimeout: 3600s
      httpFilters:
      - name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      statPrefix: localhost_8080
      streamIdleTimeout: 1800s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
	)

	DescribeTable("should set timeouts for outbound GRPC listener",
		func(given testCase) {
			// given
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(OutboundListener("outbound:192.168.0.1:8080", "192.168.0.1", 8080, xds.SocketAddressProtocolTCP)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(HttpConnectionManager("localhost:8080", false)).
					Configure(Timeout(given.timeout, core_mesh.ProtocolGRPC)))).
				Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("user's timeout", testCase{
			timeout: userTimeout,
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      commonHttpProtocolOptions:
        idleTimeout: 103s
      httpFilters:
      - name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      statPrefix: localhost_8080
      streamIdleTimeout: 104s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
		Entry("user's timeout old format (with grpc)", testCase{
			timeout: userTimeoutOldFormat,
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      commonHttpProtocolOptions:
        idleTimeout: 103s
      httpFilters:
      - name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      statPrefix: localhost_8080
      streamIdleTimeout: 104s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
		Entry("default timeout", testCase{
			timeout: mesh.DefaultTimeoutResource().(*core_mesh.TimeoutResource).Spec.GetConf(),
			expected: `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
filterChains:
- filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      commonHttpProtocolOptions:
        idleTimeout: 3600s
      httpFilters:
      - name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      statPrefix: localhost_8080
      streamIdleTimeout: 1800s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
	)

	It("should set timeouts for inbound TCP listener", func() {
		// given
		listener, err := NewListenerBuilder(envoy_common.APIV3).
			Configure(InboundListener("inbound:192.168.0.1:8080", "192.168.0.1", 8080, xds.SocketAddressProtocolTCP)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
				Configure(TcpProxy("localhost:8080", envoy_common.NewCluster(envoy_common.WithName("backend")))).
				Configure(Timeout(mesh.DefaultInboundTimeout(), core_mesh.ProtocolTCP)))).
			Build()
		Expect(err).ToNot(HaveOccurred())

		// when
		actual, err := util_proto.ToYAML(listener)
		Expect(err).ToNot(HaveOccurred())

		// then
		expected := `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
enableReusePort: false
filterChains:
- filters:
  - name: envoy.filters.network.tcp_proxy
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy
      cluster: backend
      idleTimeout: 7200s
      statPrefix: localhost_8080
name: inbound:192.168.0.1:8080
trafficDirection: INBOUND
`
		Expect(actual).To(MatchYAML(expected))
	})

	It("should set timeouts for inbound HTTP listener", func() {
		// given
		listener, err := NewListenerBuilder(envoy_common.APIV3).
			Configure(InboundListener("inbound:192.168.0.1:8080", "192.168.0.1", 8080, xds.SocketAddressProtocolTCP)).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
				Configure(HttpConnectionManager("localhost:8080", false)).
				Configure(Timeout(mesh.DefaultInboundTimeout(), core_mesh.ProtocolHTTP)))).
			Build()
		Expect(err).ToNot(HaveOccurred())

		// when
		actual, err := util_proto.ToYAML(listener)
		Expect(err).ToNot(HaveOccurred())

		// then
		expected := `
address:
  socketAddress:
    address: 192.168.0.1
    portValue: 8080
enableReusePort: false
filterChains:
- filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      commonHttpProtocolOptions:
        idleTimeout: 7200s
      httpFilters:
      - name: envoy.filters.http.router
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      statPrefix: localhost_8080
      streamIdleTimeout: 3600s
name: inbound:192.168.0.1:8080
trafficDirection: INBOUND
`
		Expect(actual).To(MatchYAML(expected))
	})
})
