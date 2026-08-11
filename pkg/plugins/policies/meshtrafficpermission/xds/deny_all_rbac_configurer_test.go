package xds_test

import (
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
)

var _ = Describe("DenyAllRBACConfigurer", func() {
	It("should install an RBAC network filter that authorizes nothing", func() {
		// given
		configurer := &xds.DenyAllRBACConfigurer{StatsName: "deny_all_prefix"}
		res, err := listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).Build()
		Expect(err).ToNot(HaveOccurred())

		// when
		err = configurer.Configure(res.(*listenerv3.FilterChain))
		Expect(err).ToNot(HaveOccurred())

		// then
		actual, err := util_proto.ToYAML(res)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchYAML(`
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules: {}
      statPrefix: deny_all_prefix.`))
	})

	It("should install an RBAC http filter when the chain has an HCM", func() {
		// given
		configurer := &xds.DenyAllRBACConfigurer{StatsName: "deny_all_prefix"}
		res, err := listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).
			Configure(listeners.HttpConnectionManager("test", false, nil, true)).
			Build()
		Expect(err).ToNot(HaveOccurred())

		// when
		err = configurer.Configure(res.(*listenerv3.FilterChain))
		Expect(err).ToNot(HaveOccurred())

		// then
		actual, err := util_proto.ToYAML(res)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(ContainSubstring("envoy.filters.http.rbac"))
	})
})
