package v2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

var _ = Describe("ResetTagsHeaderConfigurer", func() {

	It("should generate proper Envoy config", func() {
		// when
		routeConfiguration, err := routes.NewRouteConfigurationBuilder(envoy.APIV2).
			Configure(routes.ResetTagsHeader()).
			Build()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		actual, err := util_proto.ToYAML(routeConfiguration)
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(`
            requestHeadersToRemove:
              - x-kuma-tags`))
	})
})
