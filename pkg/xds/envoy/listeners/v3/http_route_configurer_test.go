package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("HttpDynamicRouteConfigurer", func() {
	It("should generate proper Envoy config", func() {
		listener, err := NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 99, xds.SocketAddressProtocolTCP).
			WithOverwriteName("inbound").
			Configure(
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).Configure(
					HttpConnectionManager("inbound", false),
					HttpDynamicRoute("routes/inbound"),
				)),
			).Build()

		Expect(err).ToNot(HaveOccurred())

		config, err := util_proto.ToYAML(listener)
		Expect(err).ToNot(HaveOccurred())
		Expect(config).To(MatchYAML(`
      address:
        socketAddress:
          address: 127.0.0.1
          portValue: 99
      enableReusePort: false
      filterChains:
      - filters:
        - name: envoy.filters.network.http_connection_manager
          typedConfig:
            '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
            httpFilters:
            - name: envoy.filters.http.router
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
            rds:
              configSource:
                ads: {}
                resourceApiVersion: V3
              routeConfigName: routes/inbound
            statPrefix: inbound
      name: inbound
      trafficDirection: INBOUND
`))
	})
})

var _ = Describe("HttpScopedRouteConfigurer", func() {
	It("should fail", func() {
		_, err := NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 99, xds.SocketAddressProtocolTCP).
			WithOverwriteName("inbound").
			Configure(
				FilterChain(NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).Configure(
					HttpConnectionManager("inbound", false),
					AddFilterChainConfigurer(&HttpScopedRouteConfigurer{}),
				)),
			).Build()

		Expect(err).To(HaveOccurred())
	})
})
