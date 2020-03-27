package routes_test

import (
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/envoy/routes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResetTagsConfigurer", func() {

	It("should generate proper Envoy config",
		func() {
			// when
			routeConfiguration, err := routes.NewRouteConfigurationBuilder().
				Configure(routes.ResetTags()).
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
