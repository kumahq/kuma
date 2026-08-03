package virtualhosts_test

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_virtual_hosts "github.com/kumahq/kuma/v3/pkg/xds/envoy/virtualhosts"
)

var _ = Describe("CatchAllRouteConfigurer", func() {
	It("should generate proper Envoy config", func() {
		// when
		virtualHost := &envoy_config_route_v3.VirtualHost{}
		err := envoy_virtual_hosts.CatchAllRouteConfigurer{
			Cluster: envoy_common.NewCluster(envoy_common.WithName("backend")),
		}.Configure(virtualHost)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		actual, err := util_proto.ToYAML(virtualHost)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(`
routes:
  - match:
      prefix: "/"
    route:
      timeout: "0s"
      cluster: backend`))
	})
})
