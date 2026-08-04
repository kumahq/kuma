package v3_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/defaults/mesh"
	policies_defaults "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/defaults"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
)

var _ = Describe("TimeoutConfigurer", func() {
	userTimeout := envoy_common.Timeouts{
		Connect:        100 * time.Second,
		TcpIdle:        101 * time.Second,
		HttpIdle:       103 * time.Second,
		HttpStreamIdle: 104 * time.Second,
	}

	timeoutConf := envoy_common.Timeouts{
		Connect:        policies_defaults.DefaultConnectTimeout,
		TcpIdle:        policies_defaults.DefaultIdleTimeout,
		HttpIdle:       policies_defaults.DefaultIdleTimeout,
		HttpStreamIdle: policies_defaults.DefaultStreamIdleTimeout,
	}

	type testCase struct {
		timeout  envoy_common.Timeouts
		expected string
	}

	DescribeTable("should set timeouts for outbound TCP listener",
		func(given testCase) {
			// given
			listener, err := NewOutboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, xds.SocketAddressProtocolTCP).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(TcpProxyDeprecated("localhost:8080", plugins_xds.NewClusterBuilder().WithName("backend").Build())).
					Configure(Timeout(given.timeout, core_meta.ProtocolTCP)))).
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
			timeout: timeoutConf,
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
			listener, err := NewOutboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, xds.SocketAddressProtocolTCP).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(HttpConnectionManager("localhost:8080", false, nil, true)).
					Configure(Timeout(given.timeout, core_meta.ProtocolHTTP)))).
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
      internalAddressConfig:
        cidrRanges:
          - addressPrefix: 127.0.0.1
            prefixLen: 32
          - addressPrefix: ::1
            prefixLen: 128
      streamIdleTimeout: 104s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
		Entry("default timeout", testCase{
			timeout: timeoutConf,
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
      internalAddressConfig:
        cidrRanges:
          - addressPrefix: 127.0.0.1
            prefixLen: 32
          - addressPrefix: ::1
            prefixLen: 128
      streamIdleTimeout: 1800s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
	)

	DescribeTable("should set timeouts for outbound GRPC listener",
		func(given testCase) {
			// given
			listener, err := NewOutboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, xds.SocketAddressProtocolTCP).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(HttpConnectionManager("localhost:8080", false, nil, true)).
					Configure(Timeout(given.timeout, core_meta.ProtocolGRPC)))).
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
      internalAddressConfig:
        cidrRanges:
          - addressPrefix: 127.0.0.1
            prefixLen: 32
          - addressPrefix: ::1
            prefixLen: 128
      streamIdleTimeout: 104s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
		Entry("default timeout", testCase{
			timeout: timeoutConf,
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
      internalAddressConfig:
        cidrRanges:
          - addressPrefix: 127.0.0.1
            prefixLen: 32
          - addressPrefix: ::1
            prefixLen: 128
      streamIdleTimeout: 1800s
name: outbound:192.168.0.1:8080
trafficDirection: OUTBOUND`,
		}),
	)

	It("should set timeouts for inbound TCP listener", func() {
		// given
		listener, err := NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, xds.SocketAddressProtocolTCP, true).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(TcpProxyDeprecated("localhost:8080", plugins_xds.NewClusterBuilder().WithName("backend").Build())).
				Configure(Timeout(mesh.DefaultInboundTimeout(), core_meta.ProtocolTCP)))).
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
enableReusePort: true`
		Expect(actual).To(MatchYAML(expected))
	})

	It("should set timeouts for inbound HTTP listener", func() {
		// given
		listener, err := NewInboundListenerBuilder(envoy_common.APIV3, "192.168.0.1", 8080, xds.SocketAddressProtocolTCP, true).
			Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
				Configure(HttpConnectionManager("localhost:8080", false, nil, true)).
				Configure(Timeout(mesh.DefaultInboundTimeout(), core_meta.ProtocolHTTP)))).
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
      internalAddressConfig:
        cidrRanges:
          - addressPrefix: 127.0.0.1
            prefixLen: 32
          - addressPrefix: ::1
            prefixLen: 128
name: inbound:192.168.0.1:8080
trafficDirection: INBOUND
enableReusePort: true`
		Expect(actual).To(MatchYAML(expected))
	})
})
