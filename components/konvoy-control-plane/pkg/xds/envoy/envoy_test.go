package envoy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	envoy "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/envoy"
)

var _ = Describe("Envoy", func() {

	It("should generate 'static' Endpoints", func() {
		// given
		expected := `
        clusterName: localhost:8080
        endpoints:
        - lbEndpoints:
          - endpoint:
              address:
                socketAddress:
                  address: 127.0.0.1
                  portValue: 8080
`
		// when
		resource := envoy.CreateStaticEndpoint("localhost:8080", "127.0.0.1", 8080)

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate 'local' Cluster", func() {
		// given
		expected := `
        name: localhost:8080
        type: STATIC
        connectTimeout: 5s
        loadAssignment:
          clusterName: localhost:8080
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: 127.0.0.1
                    portValue: 8080
`
		// when
		resource := envoy.CreateLocalCluster("localhost:8080", "127.0.0.1", 8080)

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate 'pass-through' Cluster", func() {
		// given
		expected := `
        name: pass_through
        type: ORIGINAL_DST
        lbPolicy: ORIGINAL_DST_LB
        connectTimeout: 5s
`
		// when
		resource := envoy.CreatePassThroughCluster("pass_through")

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate 'inbound' Listener without transparent proxying", func() {
		// given
		expected := `
        name: inbound:192.168.0.1:8080
        address:
          socketAddress:
            address: 192.168.0.1
            portValue: 8080
        filterChains:
        - filters:
          - name: envoy.tcp_proxy
            typedConfig:
              '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
              cluster: localhost:8080
              statPrefix: localhost:8080
`
		ctx := envoy.Context{}

		// when
		resource := envoy.CreateInboundListener(ctx, "inbound:192.168.0.1:8080", "192.168.0.1", 8080, "localhost:8080", false)

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate 'inbound' Listener with transparent proxying", func() {
		// given
		expected := `
        name: inbound:192.168.0.1:8080
        address:
          socketAddress:
            address: 192.168.0.1
            portValue: 8080
        filterChains:
        - filters:
          - name: envoy.tcp_proxy
            typedConfig:
              '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
              cluster: localhost:8080
              statPrefix: localhost:8080
        deprecatedV1:
          bindToPort: false
`
		ctx := envoy.Context{}

		// when
		resource := envoy.CreateInboundListener(ctx, "inbound:192.168.0.1:8080", "192.168.0.1", 8080, "localhost:8080", true)

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})

	It("should generate 'catch all' Listener", func() {
		// given
		expected := `
        name: catch_all
        address:
          socketAddress:
            address: 0.0.0.0
            portValue: 15001
        filterChains:
        - filters:
          - name: envoy.tcp_proxy
            typedConfig:
              '@type': type.googleapis.com/envoy.config.filter.network.tcp_proxy.v2.TcpProxy
              cluster: pass_through
              statPrefix: pass_through
        useOriginalDst: true
`
		ctx := envoy.Context{}

		// when
		resource := envoy.CreateCatchAllListener(ctx, "catch_all", "0.0.0.0", 15001, "pass_through")

		// then
		actual, err := util_proto.ToYAML(resource)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(expected))
	})
})
