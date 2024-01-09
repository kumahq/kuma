package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var _ = Describe("DirectResponseConfigurer", func() {
	It("", func() {
		// when
		filterChain, err := NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
			Configure(DirectResponse("test.host", []v3.DirectResponseEndpoints{{
				Path:       "/",
				StatusCode: 200,
				Response:   "test",
			}})).
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
})
