package v3_test

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("DirectResponseConfigurer", func() {
	It("should correctly set up direct response", func() {
		// when
		filterChain, err := NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
			Configure(DirectResponse("test.host", []v3.DirectResponseEndpoints{{
				Path:       "/",
				StatusCode: 200,
				Response:   "test",
			}}, core_xds.LocalHostAddresses)).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		// then
		actual, err := util_proto.ToYAML(filterChain)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(`
filters:
  - name: envoy.filters.network.http_connection_manager
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
      httpFilters:
          - name: envoy.filters.http.router
            typedConfig:
              '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
      internalAddressConfig:
        cidrRanges:
        - addressPrefix: 127.0.0.1
          prefixLen: 32
        - addressPrefix: ::1
          prefixLen: 128
      routeConfig:
          virtualHosts:
              - domains:
                  - '*'
                name: test.host
                routes:
                  - directResponse:
                      body:
                          inlineString: test
                      status: 200
                    match:
                      prefix: /
      statPrefix: test_host`))
	})

	It("should correctly set up network direct response", func() {
		// when
		filterChain, err := NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
			Configure(NetworkDirectResponse("{\"json\":\"value\"}")).
			Build()

		// then
		Expect(err).ToNot(HaveOccurred())

		// then
		actual, err := util_proto.ToYAML(filterChain)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(`filters:
          - name: envoy.filters.network.direct_response
            typedConfig:
              '@type': type.googleapis.com/envoy.extensions.filters.network.direct_response.v3.Config
              response:
                  inlineBytes: eyJqc29uIjoidmFsdWUifQ==`))
	})
})
