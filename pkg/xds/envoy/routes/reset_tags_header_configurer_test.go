package routes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/envoy/routes"
)

var _ = Describe("ResetTagsHeaderConfigurer", func() {

	It("should generate proper Envoy config", func() {
		// when
		routeConfiguration, err := routes.NewRouteConfigurationBuilder().
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
